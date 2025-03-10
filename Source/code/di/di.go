package di

import (
	"context"
	"os"

	"github.com/FlowingSPDG/streamdeck"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/action"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/connection"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/logger"
	"github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code/setting"
)

func InitializeStreamDeckClient(ctx context.Context) *streamdeck.Client {
	params, err := streamdeck.ParseRegistrationParams(os.Args)
	if err != nil {
		panic(err)
	}
	streamDeckClient := streamdeck.NewClient(ctx, params)
	return streamDeckClient
}

func InitializeLogger(client *streamdeck.Client, logLevel logger.LogLevel) logger.Logger {
	return logger.NewStreamDeckLogger(client, logLevel)
}

func InitializeFileLogger(ctx context.Context, logLevel logger.LogLevel) logger.Logger {
	return logger.NewFileLogger(ctx, logLevel)
}

func InitializeMultiLogger(logLevel logger.LogLevel, loggers ...logger.Logger) logger.Logger {
	return logger.NewMultiLogger(logLevel, loggers...)
}

func InitializeConnectionManager(logger logger.Logger) *connection.ConnectionManager {
	return connection.NewConnectionManager(logger)
}

func InitializePreviewAction(
	logger logger.Logger,
	connectionManager *connection.ConnectionManager,
	client *streamdeck.Client,
	inputCache setting.SettingStore[[]*action.Input],
) action.PreviewAction {
	store := setting.NewSettingStore[*setting.PreviewSetting]()
	return action.NewPreviewAction(logger, connectionManager, store, client, inputCache)
}

func InitializeSettingStore[T any]() setting.SettingStore[T] {
	return setting.NewSettingStore[T]()
}
