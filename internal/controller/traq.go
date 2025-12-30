package controller

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
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

// executeCommand はコマンドを実行する
func (c *TraqController) executeCommand(ctx context.Context, command string, messageID string, channelID string, userID string) {
	cmd := strings.TrimSpace(command)
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return
	}

	switch parts[0] {
	case "/create": // VM作成コマンド
		c.AddReaction(ctx, messageID, c.loadingStampID)

		// バリデーション
		if len(parts) < 2 {
			_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: 使い方: /create <template_vmid>")
			c.AddReaction(ctx, messageID, c.errorStampID)
			return
		}
		id, err := strconv.Atoi(parts[1])
		if err != nil {
			_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: テンプレートIDは数値で指定してください。")
			c.AddReaction(ctx, messageID, c.errorStampID)
			return
		}

		// VM作成処理
		botMessageID, _ := c.chatSvc.SendMessage(ctx, channelID, ":arrows_counterclockwise: VMを作成しています...")
		inst, ipAddress, err := c.vmUsecase.CreateVM(ctx, userID, uint32(id))
		if err != nil {
			slog.Error("Failed to create VM", "error", err)
			_ = c.chatSvc.EditMessage(ctx, messageID, ":exclamation: VM作成に失敗しました。")
			c.AddReaction(ctx, messageID, c.errorStampID)
			return
		}

		// 作成完了メッセージ送信
		_ = c.chatSvc.EditMessage(ctx, botMessageID, fmt.Sprintf(":white_check_mark: VM作成完了\n\n- VMID: %d\n- IP: `%s`", inst.Vmid, ipAddress))
		c.AddReaction(ctx, messageID, c.completedStampID)

	case "/ls": // 一覧表示コマンド
		c.AddReaction(ctx, messageID, c.loadingStampID)

		// サブコマンドによって処理を分岐
		if len(parts) < 2 {
			_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: 使い方: /ls <template|vm>")
			c.AddReaction(ctx, messageID, c.errorStampID)
			return
		}

		switch parts[1] {
		case "template":
			// テンプレート一覧表示
			c.handleListTemplates(ctx, channelID, messageID)
		case "vm":
			// VM一覧表示
			c.handleListVMs(ctx, channelID, messageID, userID)
		default:
			// 不明なサブコマンド
			_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: 不明なサブコマンドです。\n使い方: /ls <template|vm>")
			c.AddReaction(ctx, messageID, c.errorStampID)
		}

	case "/start": // VM起動コマンド
		c.AddReaction(ctx, messageID, c.loadingStampID)

		// バリデーション
		if len(parts) < 2 {
			_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: 使い方: /start <vmid>")
			c.AddReaction(ctx, messageID, c.errorStampID)
			return
		}
		vmid, err := strconv.Atoi(parts[1])
		if err != nil {
			_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: VMIDは数値で指定してください。")
			c.AddReaction(ctx, messageID, c.errorStampID)
			return
		}

		// VM起動処理
		ip, err := c.vmUsecase.StartVM(ctx, userID, uint32(vmid))
		if err != nil {
			slog.Error("Failed to start VM", "error", err, "vmid", vmid)
			_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: VM起動に失敗しました。\nVMIDが正しいか確認してください。")
			c.AddReaction(ctx, messageID, c.errorStampID)
			return
		}

		_, _ = c.chatSvc.SendMessage(ctx, channelID, fmt.Sprintf(":white_check_mark: VM %d を起動しました。\nIP: `%s`", vmid, ip))
		c.AddReaction(ctx, messageID, c.completedStampID)

	case "/stop": // VM停止コマンド
		c.AddReaction(ctx, messageID, c.loadingStampID)

		// バリデーション
		if len(parts) < 2 {
			_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: 使い方: /stop <vmid>")
			c.AddReaction(ctx, messageID, c.errorStampID)
			return
		}
		vmid, err := strconv.Atoi(parts[1])
		if err != nil {
			_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: VMIDは数値で指定してください。")
			c.AddReaction(ctx, messageID, c.errorStampID)
			return
		}

		// VM停止処理
		err = c.vmUsecase.StopVM(ctx, userID, uint32(vmid))
		if err != nil {
			slog.Error("Failed to stop VM", "error", err, "vmid", vmid)
			_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: VM停止に失敗しました。\nVMIDが正しいか確認してください。")
			c.AddReaction(ctx, messageID, c.errorStampID)
			return
		}

		_, _ = c.chatSvc.SendMessage(ctx, channelID, fmt.Sprintf(":white_check_mark: VM %d を停止しました。", vmid))
		c.AddReaction(ctx, messageID, c.completedStampID)

	case "/help": // ヘルプコマンド
		c.AddReaction(ctx, messageID, c.loadingStampID)
		c.sendHelpMessage(ctx, channelID)
		c.AddReaction(ctx, messageID, c.completedStampID)
	}
}

// handleListTemplates はテンプレート一覧を取得して送信する
func (c *TraqController) handleListTemplates(ctx context.Context, channelID string, messageID string) {
	templates, err := c.vmUsecase.ListTemplates(ctx)
	if err != nil {
		slog.Error("Failed to list templates", "error", err)
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: テンプレート一覧の取得に失敗しました。")
		c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}

	message := c.vmUsecase.FormatTemplateList(templates)
	_, _ = c.chatSvc.SendMessage(ctx, channelID, message)
	c.AddReaction(ctx, messageID, c.completedStampID)
}

// handleListVMs はユーザーのVM一覧を取得して送信する
func (c *TraqController) handleListVMs(ctx context.Context, channelID string, messageID string, userID string) {
	instances, err := c.vmUsecase.ListVMsByUser(ctx, userID)
	if err != nil {
		slog.Error("Failed to list VMs", "error", err, "userID", userID)
		_, _ = c.chatSvc.SendMessage(ctx, channelID, ":exclamation: VM一覧の取得に失敗しました。")
		c.AddReaction(ctx, messageID, c.errorStampID)
		return
	}

	message := c.vmUsecase.FormatVMList(ctx, instances)
	_, _ = c.chatSvc.SendMessage(ctx, channelID, message)
	c.AddReaction(ctx, messageID, c.completedStampID)
}

// sendHelpMessage はヘルプメッセージを送信する
func (c *TraqController) sendHelpMessage(ctx context.Context, channelID string) {
	helpMessage := "## 部室サーバー管理bot\n\n" +
		"- `/ls template` - 利用可能なテンプレートの一覧を表示\n" +
		"- `/ls vm` - 自分のVMの一覧を表示\n" +
		"- `/create <template_vmid>` - 指定したテンプレートから新しいVMを作成\n" +
		"- `/start <vmid>` - 指定したVMを起動\n" +
		"- `/stop <vmid>` - 指定したVMを停止\n"

	_, _ = c.chatSvc.SendMessage(ctx, channelID, helpMessage)
}

func (c *TraqController) AddReaction(ctx context.Context, messageID string, emoji string) error {
	err := c.chatSvc.AddReaction(ctx, messageID, emoji)
	if err != nil {
		slog.Error("Failed to add reaction", "error", err)
	}
	return err
}
