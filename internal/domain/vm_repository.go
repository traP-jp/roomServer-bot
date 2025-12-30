package domain

import "context"

type VmTemplate struct {
	Vmid uint32
	Name string
}

type VMRepository interface {
	SaveInstance(ctx context.Context, inst *Instance) error
	FindInstancesByUserID(ctx context.Context, userID string) ([]Instance, error)
	GetAllTemplates(ctx context.Context) ([]VmTemplate, error)
}
