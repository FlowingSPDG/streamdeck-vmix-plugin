package main

import (
	"context"
	_ "embed"
	"os"

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
	params, err := streamdeck.ParseRegistrationParams(os.Args)
	if err != nil {
		return err
	}

	client := stdvmix.NewStdVmix(ctx, params)

	return client.Run(ctx)
}
