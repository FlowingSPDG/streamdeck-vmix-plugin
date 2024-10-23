package logger

import (
	"fmt"

	"github.com/FlowingSPDG/streamdeck"
)

type Logger interface {
	Log(format string, a ...any)
}

type logger struct {
	c *streamdeck.Client
}

func NewLogger(c *streamdeck.Client) Logger {
	return &logger{
		c: c,
	}
}

func (l *logger) Log(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	l.c.LogMessage(message)
}
