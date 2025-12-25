package domain

import "context"

type Instance struct {
	Vmid         uint32
	UserID       string
	TemplateVmid uint32
	IpAddress    string
}

type VmTemplate struct {
	Vmid uint32
	Name string
}

type VMRepository interface {
	SaveInstance(ctx context.Context, inst *Instance) error
	FindInstancesByUserID(ctx context.Context, userID string) ([]Instance, error)
	GetAllTemplates(ctx context.Context) ([]VmTemplate, error)
}
