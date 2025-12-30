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

// CreateVM は指定されたテンプレートIDからVMをクローンし、DBに保存する
func (u *VMProvisioningUsecase) CreateVM(ctx context.Context, userID string, templateVmid uint32) (domain.Instance, error) {
	// テンプレートをDBから直接取得
	tpl, err := u.vmRepo.GetVMTemplateByVMID(ctx, templateVmid)
	if err != nil {
		return domain.Instance{}, fmt.Errorf("template %d not found: %w", templateVmid, err)
	}

	// 新しいVMIDを時刻から生成
	newVmID := uint32((uint64(time.Now().Unix()) % 900000) + 100000)

	// テンプレート名からVM名を生成
	name := strings.ToLower(strings.TrimSpace(tpl.Name))
	if name == "" {
		name = "vm"
	}
	newName := fmt.Sprintf("%s-%d", name, newVmID)

	// クローン実行
	if err := u.proxmox.CloneVM(ctx, u.nodeName, newVmID, newName, templateVmid); err != nil {
		return domain.Instance{}, fmt.Errorf("failed to clone vm: %w", err)
	}

	// 起動
	if err := u.proxmox.StartVM(ctx, u.nodeName, newVmID); err != nil {
		return domain.Instance{}, fmt.Errorf("failed to start vm: %w", err)
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
		return domain.Instance{}, fmt.Errorf("failed to get ip address after %d retries: %w", maxRetries, err)
	}

	inst := domain.Instance{
		Vmid:         newVmID,
		UserID:       userID,
		TemplateVmid: templateVmid,
		IpAddress:    ip,
	}

	if err := u.vmRepo.SaveInstance(ctx, &inst); err != nil {
		return domain.Instance{}, fmt.Errorf("failed to save instance: %w", err)
	}

	return inst, nil
}
