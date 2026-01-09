package controller

import (
	"context"
	"log/slog"
)

// sendHelpMessage はヘルプメッセージを送信する
func (c *TraqController) sendHelpMessage(ctx context.Context, channelID string) {
	helpMessage := "## 部室サーバー管理bot\n\n" +
		"- `/ls template` - 利用可能なテンプレートの一覧を表示\n" +
		"- `/ls vm` - 自分のVMの一覧を表示\n" +
		"- `/create <template_vmid>` - 指定したテンプレートから新しいVMを作成\n" +
		"- `/delete <vmid>` - 指定したVMを削除\n" +
		"- `/start <vmid>` - 指定したVMを起動\n" +
		"- `/stop <vmid>` - 指定したVMを停止\n"

	_, _ = c.chatSvc.SendMessage(ctx, channelID, helpMessage)
}

// AddReaction はメッセージにリアクションを追加する
func (c *TraqController) AddReaction(ctx context.Context, messageID string, emoji string) error {
	err := c.chatSvc.AddReaction(ctx, messageID, emoji)
	if err != nil {
		slog.Error("Failed to add reaction", "error", err)
	}
	return err
}
