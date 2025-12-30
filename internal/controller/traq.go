package controller

import (
	"context"
	"log/slog"

	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/traPtitech/traq-ws-bot/payload"
	"github.com/trap-jp/roomserver-bot/internal/domain"
	"github.com/trap-jp/roomserver-bot/internal/usecase"
)

type TraqController struct {
	bot       *traqwsbot.Bot
	botUserID string
	chatSvc   domain.ChatService
	vmUsecase *usecase.VMProvisioningUsecase
}

func NewTraqController(
	bot *traqwsbot.Bot,
	botUserID string,
	chatSvc domain.ChatService,
	vmUsecase *usecase.VMProvisioningUsecase,
) *TraqController {
	return &TraqController{
		bot:       bot,
		botUserID: botUserID,
		chatSvc:   chatSvc,
		vmUsecase: vmUsecase,
	}
}

// Start はボットを起動し、メッセージイベントをリッスンする
func (c *TraqController) Start() error {
	// メッセージ受信時のハンドラを登録
	c.bot.OnMessageCreated(func(p *payload.MessageCreated) {
		ctx := context.Background()
		c.handleMessage(ctx, p)
	})

	// ボットを起動
	return c.bot.Start()
}

// handleMessage はメッセージを処理し、適切なコマンドを実行する
func (c *TraqController) handleMessage(ctx context.Context, p *payload.MessageCreated) {
	// ボット自身のメッセージは無視
	if p.Message.User.Bot {
		return
	}

	slog.Info("Received message", "userID", p.Message.User.ID, "content", p.Message.PlainText)

	// メッセージからコマンドを抽出
	text := p.Message.PlainText
	// コマンドに応じて処理を実行
	c.executeCommand(ctx, text, p.Message.ChannelID)
}

// executeCommand はコマンドを実行する
func (c *TraqController) executeCommand(ctx context.Context, command string, channelID string) {
	switch command {
	case "/ls template":
		c.handleListTemplates(ctx, channelID)
	case "/help":
		c.sendHelpMessage(ctx, channelID)
	}
}

// handleListTemplates はテンプレート一覧を取得して送信する
func (c *TraqController) handleListTemplates(ctx context.Context, channelID string) {
	templates, err := c.vmUsecase.ListTemplates(ctx)
	if err != nil {
		_ = c.chatSvc.SendMessage(ctx, channelID, "エラー: テンプレート一覧の取得に失敗しました。")
		return
	}

	message := c.vmUsecase.FormatTemplateList(templates)
	_ = c.chatSvc.SendMessage(ctx, channelID, message)
}

// sendHelpMessage はヘルプメッセージを送信する
func (c *TraqController) sendHelpMessage(ctx context.Context, channelID string) {
	helpMessage := "## 部室サーバー管理bot\n\n" +
		"- `/ls template` - 利用可能なテンプレートの一覧を表示\n" +
		"- `/ls vm` - 自分のVMの一覧を表示\n" +
		"- `/create <template_vmid>` - 指定したテンプレートから新しいVMを作成\n" +
		"- `/start <vmid>` - 指定したVMを起動\n" +
		"- `/stop <vmid>` - 指定したVMを停止\n"

	_ = c.chatSvc.SendMessage(ctx, channelID, helpMessage)
}
