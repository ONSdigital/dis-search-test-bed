package main

import (
	"context"
	"os"

	"github.com/ONSdigital/dis-search-test-bed/cmd"
	"github.com/ONSdigital/dis-search-test-bed/ui"
)

func main() {
	if err := run(); err != nil {
		ui.Error("%s", err.Error())
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	rootCommand, err := cmd.Load(ctx)
	if err != nil {
		return err
	}

	return rootCommand.Execute()
}
