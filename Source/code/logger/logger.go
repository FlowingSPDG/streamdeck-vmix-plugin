package logger

import (
	"context"
	"fmt"

	"github.com/FlowingSPDG/streamdeck"
)

type Logger interface {
	LogMessage(ctx context.Context, format string, args ...any) error
}

type streamDeckLogger struct {
	client *streamdeck.Client
}

func NewStreamDeckLogger(client *streamdeck.Client) Logger {
	return &streamDeckLogger{
		client: client,
	}
}

func (l *streamDeckLogger) LogMessage(ctx context.Context, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return l.client.LogMessage(ctx, msg)
}
