package main

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"time"

	"github.com/FlowingSPDG/streamdeck"

	stdvmix "github.com/FlowingSPDG/streamdeck-vmix-plugin/Source/code"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	now := time.Now()
	// filename-safe timestamp
	timestamp := now.Format("2021-01-02T15-04-05")
	fileName := fmt.Sprintf("logs/streamdeck-vmix-plugin-%s.log", timestamp)
	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		panic("cannot create log directory:" + err.Error())
	}
	logfile, err := os.Create(fileName)
	if err != nil {
		panic("cannnot open log:" + err.Error())
	}
	defer logfile.Close()

	params, err := streamdeck.ParseRegistrationParams(os.Args)
	if err != nil {
		return err
	}

	client := stdvmix.NewStdVmix(ctx, params, logfile)

	return client.Run(ctx)
}
