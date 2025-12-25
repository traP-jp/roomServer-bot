package domain

import "context"

type ProxmoxService interface {
	ProvisionVM(ctx context.Context, templateID uint32, newVMName string) (uint32, error)
	StartVM(ctx context.Context, vmid uint32) error
	GetIPAddress(ctx context.Context, vmid uint32) (string, error)
}
