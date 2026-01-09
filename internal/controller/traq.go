package controller

import (
	"context"
	"log/slog"
	"strings"

	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/traPtitech/traq-ws-bot/payload"

	"github.com/trap-jp/roomserver-bot/internal/domain"
	"github.com/trap-jp/roomserver-bot/internal/usecase"
)

type TraqController struct {
	bot              *traqwsbot.Bot
	botUserID        string
	loadingStampID   string
	completedStampID string
	errorStampID     string
	chatSvc          domain.ChatService
	vmUsecase        *usecase.VMProvisioningUsecase
}

func NewTraqController(
	bot *traqwsbot.Bot,
	botUserID string,
	loadingStampID string,
	completedStampID string,
	errorStampID string,
	chatSvc domain.ChatService,
	vmUsecase *usecase.VMProvisioningUsecase,
) *TraqController {
	return &TraqController{
		bot:              bot,
		botUserID:        botUserID,
		loadingStampID:   loadingStampID,
		completedStampID: completedStampID,
		errorStampID:     errorStampID,
		chatSvc:          chatSvc,
		vmUsecase:        vmUsecase,
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
	c.executeCommand(ctx, text, p.Message.ID, p.Message.ChannelID, p.Message.User.ID)
}

// executeCommand はコマンドを実行する（各コマンドは別ファイルのハンドラへ委譲）
func (c *TraqController) executeCommand(ctx context.Context, command string, messageID string, channelID string, userID string) {
	cmd := strings.TrimSpace(command)
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "/create": // VM作成コマンド
		c.HandleCreate(ctx, parts, messageID, channelID, userID)

	case "/ls": // 一覧表示コマンド
		c.HandleList(ctx, parts, channelID, messageID, userID)

	case "/start": // VM起動コマンド
		c.HandleStart(ctx, parts, messageID, channelID, userID)

	case "/stop": // VM停止コマンド
		c.HandleStop(ctx, parts, messageID, channelID, userID)

	case "/delete": // VM削除コマンド
		c.HandleDelete(ctx, parts, messageID, channelID, userID)

	case "/help": // ヘルプコマンド
		_ = c.AddReaction(ctx, messageID, c.loadingStampID)
		c.sendHelpMessage(ctx, channelID)
		_ = c.AddReaction(ctx, messageID, c.completedStampID)
	}
}
