package loggers

import "context"

type Logger interface {
	LogMessage(ctx context.Context, format string, args ...any) error
}
