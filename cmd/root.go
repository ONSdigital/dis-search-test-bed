package cmd

import (
	"context"

	"github.com/ONSdigital/dis-search-test-bed/ui"
	"github.com/spf13/cobra"
)

// Load initializes the root command and its subcommands
func Load(ctx context.Context) (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "search-testbed",
		Short: "Search relevance testing tool",
		Long: `A comprehensive tool for testing and comparing search algorithm 
relevance across different configurations and datasets.`,
		SilenceUsage: true,
	}

	// Add persistent flag for verbose
	root.PersistentFlags().BoolVarP(&ui.Verbose, "verbose", "v", false, "Enable verbose output")

	subCommands, err := getSubCommands(ctx)
	if err != nil {
		return nil, err
	}

	root.AddCommand(subCommands...)
	return root, nil
}

func getSubCommands(ctx context.Context) ([]*cobra.Command, error) {
	compareCmd, err := compareCommand(ctx)
	if err != nil {
		return nil, err
	}

	return []*cobra.Command{
		compareCmd,
	}, nil
}
