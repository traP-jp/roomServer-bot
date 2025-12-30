package proxmox

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/luthermonson/go-proxmox"
	"github.com/trap-jp/roomserver-bot/internal/domain"
)

type ProxmoxService struct {
	client *proxmox.Client
}

func NewProxmoxService(endpoint, tokenID, secret string, insecure bool) domain.ProxmoxService {
	insecureHTTPClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		},
	}
	c := proxmox.NewClient(endpoint, proxmox.WithHTTPClient(&insecureHTTPClient), proxmox.WithAPIToken(tokenID, secret))

	return &ProxmoxService{
		client: c,
	}
}

func (p *ProxmoxService) CloneVM(ctx context.Context, nodeName string, newVmID uint32, newVMName string, templateID uint32) error {
	// ノード取得
	node, err := p.client.Node(ctx, nodeName)
	if err != nil {
		return err
	}

	// テンプレート取得
	template, err := node.VirtualMachine(ctx, int(templateID))
	if err != nil {
		return err
	}

	// クローン設定
	cloneOptions := &proxmox.VirtualMachineCloneOptions{
		NewID: int(newVmID),
		Name:  newVMName,
		Full:  1, // 1: フルクローン, 0: リンククローン
	}

	// クローン実行
	_, task, err := template.Clone(ctx, cloneOptions)
	if err != nil {
		return fmt.Errorf("failed to initiate VM clone: %w", err)
	}

	if err := task.Wait(ctx, 5*time.Second, 300*time.Second); err != nil {
		return err
	}

	return nil
}

func (p *ProxmoxService) StartVM(ctx context.Context, nodeName string, vmid uint32) error {
	// ノード取得
	node, err := p.client.Node(ctx, nodeName)
	if err != nil {
		return err
	}

	// 仮想マシン取得
	vm, err := node.VirtualMachine(ctx, int(vmid))
	if err != nil {
		return err
	}

	// 仮想マシン起動
	task, err := vm.Start(ctx)
	if err != nil {
		return err
	}

	if err := task.Wait(ctx, 2*time.Second, 60*time.Second); err != nil {
		return err
	}

	return nil
}

func (p *ProxmoxService) StopVM(ctx context.Context, nodeName string, vmid uint32) error {
	// ノード取得
	node, err := p.client.Node(ctx, nodeName)
	if err != nil {
		return err
	}

	// 仮想マシン取得
	vm, err := node.VirtualMachine(ctx, int(vmid))
	if err != nil {
		return err
	}

	// 仮想マシン停止
	task, err := vm.Stop(ctx)
	if err != nil {
		return err
	}

	if err := task.Wait(ctx, 2*time.Second, 60*time.Second); err != nil {
		return err
	}

	return nil
}

func (p *ProxmoxService) GetIPAddress(ctx context.Context, nodeName string, vmid uint32) (string, error) {
	// ノード取得
	node, err := p.client.Node(ctx, nodeName)
	if err != nil {
		return "", err
	}

	// 仮想マシン取得
	vm, err := node.VirtualMachine(ctx, int(vmid))
	if err != nil {
		return "", err
	}

	// エージェント情報取得
	interfaces, err := vm.AgentGetNetworkIFaces(ctx)
	if err != nil {
		return "", err
	}

	// IPアドレスを取得（最初に見つかったIPv4アドレスを返す）
	for _, iface := range interfaces {
		if iface.Name == "lo" {
			continue // loopbackインターフェースをスキップ
		}
		for _, addr := range iface.IPAddresses {
			if addr.IPAddressType == "ipv4" {
				return addr.IPAddress, nil
			}
		}
	}

	return "", nil
}
