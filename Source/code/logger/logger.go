package logger

import (
	"context"
	"fmt"

	"github.com/FlowingSPDG/streamdeck"
)

type Logger interface {
	Log(ctx context.Context, format string, a ...any)
}

type logger struct {
	c *streamdeck.Client
}

func NewLogger(c *streamdeck.Client) Logger {
	return &logger{
		c: c,
	}
}

func (l *logger) Log(ctx context.Context, format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	l.c.LogMessage(ctx, message)
}
