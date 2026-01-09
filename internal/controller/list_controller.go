package controller

import (
	"context"
	"fmt"
	"log/slog"
)

// HandleList は /ls コマンドを処理する
func (c *TraqController) HandleList(ctx context.Context, parts []string, channelID string, messageID string, userID string) error {
	if len(parts) < 2 {
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: 使い方: /ls <template|vm>")
		return fmt.Errorf("invalid args")
	}

	switch parts[1] {
	case "template":
		return c.handleListTemplates(ctx, channelID)
	case "vm":
		return c.handleListVMs(ctx, channelID, userID)
	default:
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: 不明なサブコマンドです。\n使い方: /ls <template|vm>")
		return fmt.Errorf("invalid subcommand")
	}
}

// handleListTemplates はテンプレート一覧を取得して送信する
func (c *TraqController) handleListTemplates(ctx context.Context, channelID string) error {
	templates, err := c.vmUsecase.ListTemplates(ctx)
	if err != nil {
		slog.Error("Failed to list templates", "error", err)
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: テンプレート一覧の取得に失敗しました。")
		return err
	}

	message := c.vmUsecase.FormatTemplateList(templates)
	_, _ = c.chatSvc.SendMessage(ctx, channelID, message)
	return nil
}

// handleListVMs はユーザーのVM一覧を取得して送信する
func (c *TraqController) handleListVMs(ctx context.Context, channelID string, userID string) error {
	instances, err := c.vmUsecase.ListVMsByUser(ctx, userID)
	if err != nil {
		slog.Error("Failed to list VMs", "error", err, "userID", userID)
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: VM一覧の取得に失敗しました。")
		return err
	}

	message := c.vmUsecase.FormatVMList(ctx, instances)
	_, _ = c.chatSvc.SendMessage(ctx, channelID, message)
	return nil
}
