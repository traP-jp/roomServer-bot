package controller

import (
	"context"
	"errors"
	"log/slog"
)

// sendHelpMessage はヘルプメッセージを送信する
func (c *TraqController) sendHelpMessage(ctx context.Context, channelID string) error {
	helpMessage := "## 部室サーバー管理bot\n\n" +
		"- `/ls template` - 利用可能なテンプレートの一覧を表示\n" +
		"- `/ls vm` - 自分のVMの一覧を表示\n" +
		"- `/create <template_vmid>` - 指定したテンプレートから新しいVMを作成\n" +
		"- `/delete <vmid>` - 指定したVMを削除\n" +
		"- `/start <vmid>` - 指定したVMを起動\n" +
		"- `/stop <vmid>` - 指定したVMを停止\n"

	_, _ = c.chatSvc.SendMessage(ctx, channelID, helpMessage)
	return nil
}

// AddReaction はメッセージにリアクションを追加する
func (c *TraqController) AddReaction(ctx context.Context, messageID string, emoji string) error {
	err := c.chatSvc.AddReaction(ctx, messageID, emoji)
	if err != nil {
		slog.Error("Failed to add reaction", "error", err)
	}
	return err
}

// ExecuteWithReactions はハンドラを実行し、開始/成功/失敗のリアクションを統一的に付与する
// handler はエラーを返す設計とし、nil の場合は成功とみなす
func (c *TraqController) ExecuteWithReactions(ctx context.Context, messageID string, handler func() error) {
	// 開始リアクション
	_ = c.AddReaction(ctx, messageID, c.loadingStampID)

	// 実行
	err := handler()
	if err == nil {
		// 成功
		_ = c.AddReaction(ctx, messageID, c.completedStampID)
		return
	}

	// エラー時
	_ = c.AddReaction(ctx, messageID, c.errorStampID)

	// ログ出力
	if !errors.Is(err, context.Canceled) {
		slog.Error("handler failed", "error", err)
	}
}
