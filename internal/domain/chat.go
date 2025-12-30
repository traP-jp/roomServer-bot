package domain

import "context"

type ChatService interface {
	SendMessage(ctx context.Context, channelID string, message string) (messageID string, err error)
	EditMessage(ctx context.Context, messageID string, newContent string) error
	AddReaction(ctx context.Context, messageID string, emoji string) error
}
