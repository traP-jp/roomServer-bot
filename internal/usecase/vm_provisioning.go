package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/trap-jp/roomserver-bot/internal/domain"
)

type VMProvisioningUsecase struct {
	vmRepo  domain.VMRepository
	proxmox domain.ProxmoxService
	chatSvc domain.ChatService
}

func NewVMProvisioningUsecase(
	vmRepo domain.VMRepository,
	proxmox domain.ProxmoxService,
	chatSvc domain.ChatService,
) *VMProvisioningUsecase {
	return &VMProvisioningUsecase{
		vmRepo:  vmRepo,
		proxmox: proxmox,
		chatSvc: chatSvc,
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
