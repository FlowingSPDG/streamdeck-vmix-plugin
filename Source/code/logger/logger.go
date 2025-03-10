package logger

import (
	"context"
	"fmt"
	"os"

	"github.com/FlowingSPDG/streamdeck"
)

type Logger interface {
	LogMessage(ctx context.Context, format string, args ...any) error
	Info(ctx context.Context, format string, args ...any) error
	Warn(ctx context.Context, format string, args ...any) error
	Error(ctx context.Context, format string, args ...any) error
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

func (l *streamDeckLogger) Info(ctx context.Context, format string, args ...any) error {
	return l.LogMessage(ctx, "[INFO] "+format, args...)
}

func (l *streamDeckLogger) Warn(ctx context.Context, format string, args ...any) error {
	return l.LogMessage(ctx, "[WARN] "+format, args...)
}

func (l *streamDeckLogger) Error(ctx context.Context, format string, args ...any) error {
	return l.LogMessage(ctx, "[ERROR] "+format, args...)
}

type fileLogger struct {
	file *os.File
}

func NewFileLogger(ctx context.Context) Logger {
	file, err := os.OpenFile("./log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return &fileLogger{
		file: file,
	}
}

func (l *fileLogger) LogMessage(ctx context.Context, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	_, err := l.file.WriteString(msg + "\n")
	if err != nil {
		return err
	}
	return nil
}

func (l *fileLogger) Info(ctx context.Context, format string, args ...any) error {
	return l.LogMessage(ctx, "[INFO] "+format, args...)
}

func (l *fileLogger) Warn(ctx context.Context, format string, args ...any) error {
	return l.LogMessage(ctx, "[WARN] "+format, args...)
}

func (l *fileLogger) Error(ctx context.Context, format string, args ...any) error {
	return l.LogMessage(ctx, "[ERROR] "+format, args...)
}

type multiLogger struct {
	loggers []Logger
}

func NewMultiLogger(loggers ...Logger) Logger {
	return &multiLogger{loggers: loggers}
}

func (l *multiLogger) LogMessage(ctx context.Context, format string, args ...any) error {
	for _, logger := range l.loggers {
		logger.LogMessage(ctx, format, args...)
	}
	return nil
}

func (l *multiLogger) Info(ctx context.Context, format string, args ...any) error {
	for _, logger := range l.loggers {
		logger.Info(ctx, format, args...)
	}
	return nil
}

func (l *multiLogger) Warn(ctx context.Context, format string, args ...any) error {
	for _, logger := range l.loggers {
		logger.Warn(ctx, format, args...)
	}
	return nil
}

func (l *multiLogger) Error(ctx context.Context, format string, args ...any) error {
	for _, logger := range l.loggers {
		logger.Error(ctx, format, args...)
	}
	return nil
}
