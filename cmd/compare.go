package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/ONSdigital/dis-search-test-bed/models"
	"github.com/ONSdigital/dis-search-test-bed/shared/comparison"
	"github.com/ONSdigital/dis-search-test-bed/shared/output"
	"github.com/ONSdigital/dis-search-test-bed/shared/paths"
	"github.com/ONSdigital/dis-search-test-bed/ui"
	"github.com/spf13/cobra"
)

var (
	compareWith string
	compareMode string
)

var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare query results",
	Long: `Compare query results between different runs or between queries 
within the same run.`,
	RunE: runCompare,
}

func init() {
	rootCmd.AddCommand(compareCmd)

	compareCmd.Flags().StringVar(&compareWith, "with", "",
		"Previous results file to compare against (defaults to previous run)")
	compareCmd.Flags().StringVar(&compareMode, "mode", "both",
		"Comparison mode: historical, cross-query, or both")
}

func runCompare(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	printer := ui.NewPrinter(verbose)

	// Load current results
	currentPath, err := paths.FindLatestResults(cfg.Output.BaseDir)
	if err != nil {
		return fmt.Errorf("failed to find current results: %w", err)
	}

	printer.Info("Current results: %s", currentPath)

	current, err := output.LoadResults(currentPath)
	if err != nil {
		return fmt.Errorf("failed to load current results: %w", err)
	}

	var previous []models.QueryResults
	mode := parseComparisonMode(compareMode)

	// Load previous results if needed
	if mode == comparison.ModeHistorical || mode == comparison.ModeBoth {
		if compareWith == "" {
			prevPath, err := paths.FindPreviousResults(cfg.Output.BaseDir, currentPath)
			if err != nil {
				printer.Warning("No previous results found, skipping historical comparison")
				mode = comparison.ModeCrossQuery
			} else {
				compareWith = prevPath
			}
		}

		if compareWith != "" {
			printer.Info("Comparing with: %s", compareWith)
			previous, err = output.LoadResults(compareWith)
			if err != nil {
				return fmt.Errorf("failed to load previous results: %w", err)
			}
		}
	}

	// Create comparison
	opts := comparison.Options{
		ShowUnchanged:  cfg.Comparison.ShowUnchanged,
		HighlightNew:   cfg.Comparison.HighlightNew,
		ShowScores:     cfg.Comparison.ShowScores,
		MaxRankDisplay: cfg.Comparison.MaxRankDisplay,
	}

	comp := comparison.NewComparison(current, previous, opts, mode)

	spinner := ui.NewSpinner("Generating comparison report...")
	spinner.Start()

	report, err := comp.Generate()
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to generate comparison: %w", err)
	}

	spinner.Stop()

	// Save report
	runFolder := filepath.Dir(currentPath)
	diffPath := filepath.Join(runFolder, "comparison.txt")

	if err := output.WriteText(diffPath, report); err != nil {
		return fmt.Errorf("failed to write comparison: %w", err)
	}

	printer.Success("Comparison saved to: %s", diffPath)

	// Print summary
	printer.Section("Comparison Summary")
	summary := comp.GetSummary()
	printer.Info("Mode: %s", summary.Mode)
	if mode == comparison.ModeHistorical || mode == comparison.ModeBoth {
		printer.Info("New results: %d", summary.NewResults)
		printer.Info("Removed results: %d", summary.RemovedResults)
		printer.Info("Improved rankings: %d", summary.ImprovedRankings)
		printer.Info("Worsened rankings: %d", summary.WorsenedRankings)
	}

	printer.Celebrate("Comparison complete!")
	return nil
}

func parseComparisonMode(mode string) comparison.Mode {
	switch mode {
	case "historical":
		return comparison.ModeHistorical
	case "cross-query":
		return comparison.ModeCrossQuery
	case "both":
		return comparison.ModeBoth
	default:
		return comparison.ModeBoth
	}
}
