package logger

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/FlowingSPDG/streamdeck"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DebugLevel    LogLevel = 1 << 0 // 1
	InfoLevel     LogLevel = 1 << 1 // 2
	WarnLevel     LogLevel = 1 << 2 // 4
	ErrorLevel    LogLevel = 1 << 3 // 8
	CriticalLevel LogLevel = 1 << 4 // 16
)

func checkDebugLevel(level LogLevel) bool {
	return level&DebugLevel != 0
}

func checkInfoLevel(level LogLevel) bool {
	return level&InfoLevel != 0
}

func checkWarnLevel(level LogLevel) bool {
	return level&WarnLevel != 0
}

func checkErrorLevel(level LogLevel) bool {
	return level&ErrorLevel != 0
}

func checkCriticalLevel(level LogLevel) bool {
	return level&CriticalLevel != 0
}

type Logger interface {
	LogMessage(ctx context.Context, format string, args ...any) error
	Debug(ctx context.Context, format string, args ...any) error
	Info(ctx context.Context, format string, args ...any) error
	Warn(ctx context.Context, format string, args ...any) error
	Error(ctx context.Context, format string, args ...any) error
}

type streamDeckLogger struct {
	client *streamdeck.Client
	level  LogLevel
}

func NewStreamDeckLogger(client *streamdeck.Client, level LogLevel) Logger {
	return &streamDeckLogger{
		client: client,
		level:  level,
	}
}

func (l *streamDeckLogger) LogMessage(ctx context.Context, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	if l.client == nil {
		return nil
	}
	if !l.client.IsConnected() {
		return nil
	}
	return l.client.LogMessage(ctx, msg)
}

func (l *streamDeckLogger) Debug(ctx context.Context, format string, args ...any) error {
	if !checkDebugLevel(l.level) {
		return nil
	}
	return l.LogMessage(ctx, "[DEBUG] "+format, args...)
}

func (l *streamDeckLogger) Info(ctx context.Context, format string, args ...any) error {
	if !checkInfoLevel(l.level) {
		return nil
	}
	return l.LogMessage(ctx, "[INFO] "+format, args...)
}

func (l *streamDeckLogger) Warn(ctx context.Context, format string, args ...any) error {
	if !checkWarnLevel(l.level) {
		return nil
	}
	return l.LogMessage(ctx, "[WARN] "+format, args...)
}

func (l *streamDeckLogger) Error(ctx context.Context, format string, args ...any) error {
	if !checkErrorLevel(l.level) {
		return nil
	}
	return l.LogMessage(ctx, "[ERROR] "+format, args...)
}

type fileLogger struct {
	file  *os.File
	level LogLevel
}

func NewFileLogger(ctx context.Context, level LogLevel) Logger {
	file, err := os.OpenFile("./log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	return &fileLogger{
		file:  file,
		level: level,
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

func (l *fileLogger) Debug(ctx context.Context, format string, args ...any) error {
	if !checkDebugLevel(l.level) {
		return nil
	}
	return l.LogMessage(ctx, "[DEBUG] "+format, args...)
}

func (l *fileLogger) Info(ctx context.Context, format string, args ...any) error {
	if !checkInfoLevel(l.level) {
		return nil
	}
	return l.LogMessage(ctx, "[INFO] "+format, args...)
}

func (l *fileLogger) Warn(ctx context.Context, format string, args ...any) error {
	if !checkWarnLevel(l.level) {
		return nil
	}
	return l.LogMessage(ctx, "[WARN] "+format, args...)
}

func (l *fileLogger) Error(ctx context.Context, format string, args ...any) error {
	if !checkErrorLevel(l.level) {
		return nil
	}
	return l.LogMessage(ctx, "[ERROR] "+format, args...)
}

type multiLogger struct {
	loggers []Logger
	level   LogLevel
}

func NewMultiLogger(level LogLevel, loggers ...Logger) Logger {
	return &multiLogger{
		loggers: loggers,
		level:   level,
	}
}

func (l *multiLogger) LogMessage(ctx context.Context, format string, args ...any) error {
	for _, logger := range l.loggers {
		logger.LogMessage(ctx, format, args...)
	}
	return nil
}

func (l *multiLogger) Debug(ctx context.Context, format string, args ...any) error {
	if !checkDebugLevel(l.level) {
		return nil
	}
	for _, logger := range l.loggers {
		logger.Debug(ctx, format, args...)
	}
	return nil
}

func (l *multiLogger) Info(ctx context.Context, format string, args ...any) error {
	if !checkInfoLevel(l.level) {
		return nil
	}
	for _, logger := range l.loggers {
		logger.Info(ctx, format, args...)
	}
	return nil
}

func (l *multiLogger) Warn(ctx context.Context, format string, args ...any) error {
	if !checkWarnLevel(l.level) {
		return nil
	}
	for _, logger := range l.loggers {
		logger.Warn(ctx, format, args...)
	}
	return nil
}

func (l *multiLogger) Error(ctx context.Context, format string, args ...any) error {
	if !checkErrorLevel(l.level) {
		return nil
	}
	for _, logger := range l.loggers {
		logger.Error(ctx, format, args...)
	}
	return nil
}

type testLogger struct {
	t *testing.T
}

func NewTestLogger(t *testing.T) Logger {
	return &testLogger{t: t}
}

func (l *testLogger) LogMessage(ctx context.Context, format string, args ...any) error {
	l.t.Logf(format, args...)
	return nil
}

func (l *testLogger) Debug(ctx context.Context, format string, args ...any) error {
	l.t.Logf(format, args...)
	return nil
}

func (l *testLogger) Info(ctx context.Context, format string, args ...any) error {
	l.t.Logf(format, args...)
	return nil
}

func (l *testLogger) Warn(ctx context.Context, format string, args ...any) error {
	l.t.Logf(format, args...)
	return nil
}

func (l *testLogger) Error(ctx context.Context, format string, args ...any) error {
	l.t.Logf(format, args...)
	return nil
}
