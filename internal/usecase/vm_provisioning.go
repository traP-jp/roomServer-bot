package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/trap-jp/roomserver-bot/internal/domain"
)

type VMProvisioningUsecase struct {
	vmRepo   domain.VMRepository
	proxmox  domain.ProxmoxService
	chatSvc  domain.ChatService
	nodeName string
}

func NewVMProvisioningUsecase(
	vmRepo domain.VMRepository,
	proxmox domain.ProxmoxService,
	chatSvc domain.ChatService,
	nodeName string,
) *VMProvisioningUsecase {
	return &VMProvisioningUsecase{
		vmRepo:   vmRepo,
		proxmox:  proxmox,
		chatSvc:  chatSvc,
		nodeName: nodeName,
	}
}

// ListTemplates は利用可能なテンプレートの一覧を取得して返す
func (u *VMProvisioningUsecase) ListTemplates(ctx context.Context) ([]domain.VmTemplate, error) {
	templates, err := u.vmRepo.GetAllTemplates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates: %w", err)
	}
	return templates, nil
}

// FormatTemplateList はテンプレート一覧を整形されたメッセージに変換する
func (u *VMProvisioningUsecase) FormatTemplateList(templates []domain.VmTemplate) string {
	if len(templates) == 0 {
		return "利用可能なテンプレートはありません。"
	}

	var builder strings.Builder
	builder.WriteString("利用可能なテンプレート\n\n")

	for _, template := range templates {
		builder.WriteString(fmt.Sprintf("- %d: `%s`\n", template.Vmid, template.Name))
	}

	return builder.String()
}

// ListVMsByUser は指定されたユーザーのVM一覧を取得して返す
func (u *VMProvisioningUsecase) ListVMsByUser(ctx context.Context, userID string) ([]domain.Instance, error) {
	instances, err := u.vmRepo.FindInstancesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instances: %w", err)
	}
	return instances, nil
}

// FormatVMList はVM一覧を整形されたメッセージに変換する
func (u *VMProvisioningUsecase) FormatVMList(ctx context.Context, instances []domain.Instance) string {
	if len(instances) == 0 {
		return "VMを所有していません。"
	}

	var builder strings.Builder
	builder.WriteString("所有VM一覧\n\n")

	for _, inst := range instances {
		// テンプレート名を取得
		templateName := fmt.Sprintf("%d", inst.TemplateVmid)
		if tpl, err := u.vmRepo.GetVMTemplateByVMID(ctx, inst.TemplateVmid); err == nil {
			templateName = tpl.Name
		}

		builder.WriteString(fmt.Sprintf("- VMID: %d (テンプレート: `%s`)\n", inst.Vmid, templateName))
	}

	return builder.String()
}

// CreateVM は指定されたテンプレートIDからVMをクローンし、DBに保存する
func (u *VMProvisioningUsecase) CreateVM(ctx context.Context, userID string, templateVmid uint32) (domain.Instance, string, error) {
	// テンプレートをDBから直接取得
	tpl, err := u.vmRepo.GetVMTemplateByVMID(ctx, templateVmid)
	if err != nil {
		return domain.Instance{}, "", fmt.Errorf("template %d not found: %w", templateVmid, err)
	}

	// 新しいVMIDを2026-01-01 00:00:00からの経過秒で生成
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	newVmID := uint32(time.Now().Unix() - base.Unix())

	// テンプレート名からVM名を生成
	name := strings.ToLower(strings.TrimSpace(tpl.Name))
	if name == "" {
		name = "vm"
	}
	newName := fmt.Sprintf("%s-%d", name, newVmID)

	// クローン実行
	if err := u.proxmox.CloneVM(ctx, u.nodeName, newVmID, newName, templateVmid); err != nil {
		return domain.Instance{}, "", fmt.Errorf("failed to clone vm: %w", err)
	}

	// 起動
	if err := u.proxmox.StartVM(ctx, u.nodeName, newVmID); err != nil {
		return domain.Instance{}, "", fmt.Errorf("failed to start vm: %w", err)
	}

	// IPアドレス取得
	// QEMU Guest Agentが起動するまで待機
	var ip string
	const maxRetries = 36
	const retryInterval = 5 * time.Second

	for range maxRetries {
		ip, err = u.proxmox.GetIPAddress(ctx, u.nodeName, newVmID)
		if err == nil && ip != "" {
			break
		}
		time.Sleep(retryInterval)
	}
	if ip == "" {
		return domain.Instance{}, "", fmt.Errorf("failed to get ip address after %d retries: %w", maxRetries, err)
	}

	inst := domain.Instance{
		Vmid:         newVmID,
		UserID:       userID,
		TemplateVmid: templateVmid,
	}

	if err := u.vmRepo.SaveInstance(ctx, &inst); err != nil {
		return domain.Instance{}, "", fmt.Errorf("failed to save instance: %w", err)
	}

	return inst, ip, nil
}

// StartVM は指定されたVMIDのVMを起動する（所有者チェック付き）
func (u *VMProvisioningUsecase) StartVM(ctx context.Context, userID string, vmid uint32) (string, error) {
	// VMの所有者を確認
	instances, err := u.vmRepo.FindInstancesByUserID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get instances: %w", err)
	}

	owned := false
	for _, inst := range instances {
		if inst.Vmid == vmid {
			owned = true
			break
		}
	}

	if !owned {
		return "", fmt.Errorf("VM %d is not owned by user %s", vmid, userID)
	}

	// VM起動
	if err := u.proxmox.StartVM(ctx, u.nodeName, vmid); err != nil {
		return "", fmt.Errorf("failed to start vm: %w", err)
	}

	// IPアドレス取得
	// QEMU Guest Agentが起動するまで待機
	var ip string
	const maxRetries = 36
	const retryInterval = 5 * time.Second

	for range maxRetries {
		ip, err = u.proxmox.GetIPAddress(ctx, u.nodeName, vmid)
		if err == nil && ip != "" {
			break
		}
		time.Sleep(retryInterval)
	}
	if ip == "" {
		return "", fmt.Errorf("failed to get ip address after %d retries: %w", maxRetries, err)
	}

	return ip, nil
}

// StopVM は指定されたVMIDのVMを停止する（所有者チェック付き）
func (u *VMProvisioningUsecase) StopVM(ctx context.Context, userID string, vmid uint32) error {
	// VMの所有者を確認
	instances, err := u.vmRepo.FindInstancesByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get instances: %w", err)
	}

	owned := false
	for _, inst := range instances {
		if inst.Vmid == vmid {
			owned = true
			break
		}
	}

	if !owned {
		return fmt.Errorf("VM %d is not owned by user %s", vmid, userID)
	}

	// VM停止
	if err := u.proxmox.StopVM(ctx, u.nodeName, vmid); err != nil {
		return fmt.Errorf("failed to stop vm: %w", err)
	}

	return nil
}

// DeleteVM は指定されたVMIDのVMを削除する（所有者チェック付き）
func (u *VMProvisioningUsecase) DeleteVM(ctx context.Context, userID string, vmid uint32) error {
	// VMの所有者を確認
	instances, err := u.vmRepo.FindInstancesByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get instances: %w", err)
	}

	owned := false
	for _, inst := range instances {
		if inst.Vmid == vmid {
			owned = true
			break
		}
	}

	if !owned {
		return fmt.Errorf("VM %d is not owned by user %s", vmid, userID)
	}

	// Proxmox上でVM削除
	if err := u.proxmox.DeleteVM(ctx, u.nodeName, vmid); err != nil {
		return fmt.Errorf("failed to delete vm: %w", err)
	}

	// DB上のインスタンス情報を削除
	if err := u.vmRepo.DeleteInstance(ctx, vmid); err != nil {
		return fmt.Errorf("failed to delete instance from db: %w", err)
	}

	return nil
}
