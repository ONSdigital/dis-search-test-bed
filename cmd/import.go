package cmd

import (
	"github.com/ONSdigital/dis-search-test-bed/app"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const importCommandName = "import"

func importCommand() (*cobra.Command, error) {
	var inputPath string

	importCmd := &cobra.Command{
		Use:   importCommandName,
		Short: "Import re-scored relevance judgements from an export-format CSV",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return app.New().Import(cmd.Context(), inputPath)
		},
	}
	importCmd.Flags().StringVarP(&inputPath, "input", "i", "", "path to the CSV input file")
	if err := importCmd.MarkFlagRequired("input"); err != nil {
		return nil, errors.Wrap(err, "failed to require input flag")
	}

	return importCmd, nil
}
