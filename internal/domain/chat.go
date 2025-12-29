package domain

import "context"

type ChatService interface {
	SendMessage(ctx context.Context, channelID string, message string) error
	EditMessage(ctx context.Context, messageID string, newContent string) error
}
