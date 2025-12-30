package domain

import "context"

type ProxmoxService interface {
	CloneVM(ctx context.Context, nodeName string, newVmID uint32, newVMName string, templateID uint32) error
	StartVM(ctx context.Context, nodeName string, vmid uint32) error
	StopVM(ctx context.Context, nodeName string, vmid uint32) error
	DeleteVM(ctx context.Context, nodeName string, vmid uint32) error
	GetIPAddress(ctx context.Context, nodeName string, vmid uint32) (string, error)
}
