package controller

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
)

// HandleCreate は /create コマンドを処理する
func (c *TraqController) HandleCreate(ctx context.Context, parts []string, messageID string, channelID string, userID string) {
	_ = c.AddReaction(ctx, messageID, c.loadingStampID)

	// バリデーション
	if len(parts) < 2 {
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: 使い方: /create <template_vmid>")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil {
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: テンプレートIDは数値で指定してください。")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}

	// VM作成処理
	botMessageID, _ := c.chatSvc.SendMessage(ctx, channelID, ":arrows_counterclockwise: VMを作成しています...")
	inst, ipAddress, err := c.vmUsecase.CreateVM(ctx, userID, uint32(id))
	if err != nil {
		slog.Error("Failed to create VM", "error", err)
		_ = c.chatSvc.EditMessage(ctx, botMessageID, ":exclamation: VM作成に失敗しました。")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}

	// 作成完了メッセージ送信
	_ = c.chatSvc.EditMessage(ctx, botMessageID, fmt.Sprintf(":white_check_mark: VM作成完了\n\n- VMID: %d\n- IP: `%s`", inst.Vmid, ipAddress))
	_ = c.AddReaction(ctx, messageID, c.completedStampID)
}

// HandleStart は /start コマンドを処理する
func (c *TraqController) HandleStart(ctx context.Context, parts []string, messageID string, channelID string, userID string) {
	_ = c.AddReaction(ctx, messageID, c.loadingStampID)

	// バリデーション
	if len(parts) < 2 {
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: 使い方: /start <vmid>")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}
	vmid, err := strconv.Atoi(parts[1])
	if err != nil {
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: VMIDは数値で指定してください。")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}

	// VM起動処理
	ip, err := c.vmUsecase.StartVM(ctx, userID, uint32(vmid))
	if err != nil {
		slog.Error("Failed to start VM", "error", err, "vmid", vmid)
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: VM起動に失敗しました。\nVMIDが正しいか確認してください。")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}

	_, _ = c.chatSvc.SendMessage(ctx, channelID, fmt.Sprintf(":white_check_mark: VM %d を起動しました。\nIP: `%s`", vmid, ip))
	_ = c.AddReaction(ctx, messageID, c.completedStampID)
}

// HandleStop は /stop コマンドを処理する
func (c *TraqController) HandleStop(ctx context.Context, parts []string, messageID string, channelID string, userID string) {
	_ = c.AddReaction(ctx, messageID, c.loadingStampID)

	// バリデーション
	if len(parts) < 2 {
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: 使い方: /stop <vmid>")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}
	vmid, err := strconv.Atoi(parts[1])
	if err != nil {
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: VMIDは数値で指定してください。")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}

	// VM停止処理
	err = c.vmUsecase.StopVM(ctx, userID, uint32(vmid))
	if err != nil {
		slog.Error("Failed to stop VM", "error", err, "vmid", vmid)
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: VM停止に失敗しました。\nVMIDが正しいか確認してください。")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}

	_, _ = c.chatSvc.SendMessage(ctx, channelID, fmt.Sprintf(":white_check_mark: VM %d を停止しました。", vmid))
	_ = c.AddReaction(ctx, messageID, c.completedStampID)
}

// HandleDelete は /delete コマンドを処理する
func (c *TraqController) HandleDelete(ctx context.Context, parts []string, messageID string, channelID string, userID string) {
	_ = c.AddReaction(ctx, messageID, c.loadingStampID)

	// バリデーション
	if len(parts) < 2 {
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: 使い方: /delete <vmid>")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}
	vmid, err := strconv.Atoi(parts[1])
	if err != nil {
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: VMIDは数値で指定してください。")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}

	// VM削除処理
	err = c.vmUsecase.DeleteVM(ctx, userID, uint32(vmid))
	if err != nil {
		slog.Error("Failed to delete VM", "error", err, "vmid", vmid)
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: VM削除に失敗しました。\nVMIDが正しいか、またVMが停止しているか確認してください。")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}

	_, _ = c.chatSvc.SendMessage(ctx, channelID, fmt.Sprintf(":white_check_mark: VM %d を削除しました。", vmid))
	_ = c.AddReaction(ctx, messageID, c.completedStampID)
}
