package traq

import (
	"context"

	"github.com/traPtitech/go-traq"
	traqwsbot "github.com/traPtitech/traq-ws-bot"
	"github.com/trap-jp/roomserver-bot/internal/domain"
)

type TraqService struct {
	bot *traqwsbot.Bot
}

func NewTraqService(accessToken string, origin string) (domain.ChatService, error) {
	bot, err := traqwsbot.NewBot(&traqwsbot.Options{
		AccessToken: accessToken,
		Origin:      origin,
	})
	if err != nil {
		return nil, err
	}
	return &TraqService{
		bot: bot,
	}, nil
}

func (t *TraqService) SendMessage(ctx context.Context, channelID string, message string) error {
	_, _, err := t.bot.API().MessageApi.PostMessage(ctx, channelID).
		PostMessageRequest(traq.PostMessageRequest{
			Content: message,
		}).
		Execute()
	return err
}

func (t *TraqService) EditMessage(ctx context.Context, messageID string, newContent string) error {
	_, err := t.bot.API().MessageApi.EditMessage(ctx, messageID).
		PostMessageRequest(traq.PostMessageRequest{
			Content: newContent,
		}).
		Execute()
	return err
}

func (t *TraqService) AddReaction(ctx context.Context, messageID string, emoji string) error {
	_, err := t.bot.API().MessageApi.AddMessageStamp(ctx, messageID, emoji).Execute()
	return err
}
