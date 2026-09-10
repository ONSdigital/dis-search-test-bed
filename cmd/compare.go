package cmd

import (
	"github.com/ONSdigital/dis-search-test-bed/app"
	"github.com/spf13/cobra"
)

func compareCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "compare",
		Short: "Compare algorithms across terms",
		Long:  `Compare algorithms across different terms requests with the same index.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.New().Compare(cmd.Context())
		},
	}
}
