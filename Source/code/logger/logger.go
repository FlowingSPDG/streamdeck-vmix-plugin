package logger

import (
	"context"
	"fmt"

	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger/loggers"

	"github.com/FlowingSPDG/streamdeck"
)

type streamDeckLogger struct {
	client *streamdeck.Client
}

func NewStreamDeckLogger(client *streamdeck.Client) loggers.Logger {
	return &streamDeckLogger{
		client: client,
	}
}

func (l *streamDeckLogger) LogMessage(ctx context.Context, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return l.client.LogMessage(ctx, msg)
}
