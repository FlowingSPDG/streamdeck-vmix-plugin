package logger

import (
	"context"
	"fmt"

	"github.com/FlowingSPDG/streamdeck"
)

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

func (l *streamDeckLogger) Info(ctx context.Context, format string, args ...any) error {
	return l.LogMessage(ctx, "[INFO] "+format, args...)
}

func (l *streamDeckLogger) Warn(ctx context.Context, format string, args ...any) error {
	return l.LogMessage(ctx, "[WARN] "+format, args...)
}

func (l *streamDeckLogger) Error(ctx context.Context, format string, args ...any) error {
	return l.LogMessage(ctx, "[ERROR] "+format, args...)
}

type Logger interface {
	LogMessage(ctx context.Context, format string, args ...any) error
	Info(ctx context.Context, format string, args ...any) error
	Warn(ctx context.Context, format string, args ...any) error
	Error(ctx context.Context, format string, args ...any) error
}
