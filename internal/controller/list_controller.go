package controller

import (
	"context"
	"log/slog"
)

// HandleList は /ls コマンドを処理する
func (c *TraqController) HandleList(ctx context.Context, parts []string, channelID string, messageID string, userID string) {
	_ = c.AddReaction(ctx, messageID, c.loadingStampID)

	if len(parts) < 2 {
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: 使い方: /ls <template|vm>")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}

	switch parts[1] {
	case "template":
		c.handleListTemplates(ctx, channelID, messageID)
	case "vm":
		c.handleListVMs(ctx, channelID, messageID, userID)
	default:
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: 不明なサブコマンドです。\n使い方: /ls <template|vm>")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
	}
}

// handleListTemplates はテンプレート一覧を取得して送信する
func (c *TraqController) handleListTemplates(ctx context.Context, channelID string, messageID string) {
	templates, err := c.vmUsecase.ListTemplates(ctx)
	if err != nil {
		slog.Error("Failed to list templates", "error", err)
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: テンプレート一覧の取得に失敗しました。")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}

	message := c.vmUsecase.FormatTemplateList(templates)
	_, _ = c.chatSvc.SendMessage(ctx, channelID, message)
	_ = c.AddReaction(ctx, messageID, c.completedStampID)
}

// handleListVMs はユーザーのVM一覧を取得して送信する
func (c *TraqController) handleListVMs(ctx context.Context, channelID string, messageID string, userID string) {
	instances, err := c.vmUsecase.ListVMsByUser(ctx, userID)
	if err != nil {
		slog.Error("Failed to list VMs", "error", err, "userID", userID)
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: VM一覧の取得に失敗しました。")
		_ = c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}

	message := c.vmUsecase.FormatVMList(ctx, instances)
	_, _ = c.chatSvc.SendMessage(ctx, channelID, message)
	_ = c.AddReaction(ctx, messageID, c.completedStampID)
}
