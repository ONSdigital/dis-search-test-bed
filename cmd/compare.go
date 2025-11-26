package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

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
	runFolder := filepath.Dir(currentPath)

	// Load previous results if needed
	if mode == comparison.ModeHistorical || mode == comparison.ModeBoth {
		if compareWith == "" {
			prevPath, err := paths.FindPreviousResults(cfg.Output.BaseDir, currentPath)
			if err != nil {
				printer.Warning("No previous results found, skipping historical comparison")
				if mode == comparison.ModeHistorical {
					return fmt.Errorf("historical comparison requested but no previous results found")
				}
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

	// Create comparison and generate reports
	switch mode {
	case comparison.ModeHistorical:
		return generateHistoricalComparison(current, previous, runFolder, printer)
	case comparison.ModeCrossQuery:
		return generateCrossQueryComparison(current, runFolder, printer)
	case comparison.ModeBoth:
		if err := generateHistoricalComparison(current, previous, runFolder, printer); err != nil {
			return err
		}
		return generateCrossQueryComparison(current, runFolder, printer)
	default:
		return fmt.Errorf("unknown comparison mode: %s", compareMode)
	}
}

func generateHistoricalComparison(current, previous []models.QueryResults, runFolder string, printer *ui.Printer) error {
	if len(previous) == 0 {
		printer.Warning("No previous results to compare against")
		return nil
	}

	printer.Info("Generating historical comparison...")

	opts := comparison.Options{
		ShowUnchanged:  true,
		HighlightNew:   true,
		ShowScores:     true,
		MaxRankDisplay: 20,
	}

	comp := comparison.NewComparison(current, previous, opts, comparison.ModeHistorical)

	spinner := ui.NewSpinner("Generating historical comparison report...")
	spinner.Start()

	report, err := comp.Generate()
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to generate historical comparison: %w", err)
	}

	spinner.Stop()

	// Save historical comparison
	historicalPath := filepath.Join(runFolder, "comparison_historical.txt")
	if err := output.WriteText(historicalPath, report); err != nil {
		return fmt.Errorf("failed to write historical comparison: %w", err)
	}

	printer.Success("Historical comparison saved to: %s", historicalPath)

	// Print summary
	summary := comp.GetSummary()
	printer.Section("Historical Comparison Summary")
	printer.Info("New results: %d", summary.NewResults)
	printer.Info("Removed results: %d", summary.RemovedResults)
	printer.Info("Improved rankings: %d", summary.ImprovedRankings)
	printer.Info("Worsened rankings: %d", summary.WorsenedRankings)

	return nil
}

func generateCrossQueryComparison(current []models.QueryResults, runFolder string, printer *ui.Printer) error {
	if len(current) < 2 {
		printer.Warning("Need at least 2 queries to perform cross-query comparison")
		return nil
	}

	printer.Info("Generating cross-query comparison...")

	opts := comparison.Options{
		ShowUnchanged:  false,
		HighlightNew:   true,
		ShowScores:     true,
		MaxRankDisplay: 20,
	}

	comp := comparison.NewComparison(current, nil, opts, comparison.ModeCrossQuery)

	spinner := ui.NewSpinner("Generating cross-query comparison report...")
	spinner.Start()

	report, err := comp.Generate()
	if err != nil {
		spinner.Stop()
		return fmt.Errorf("failed to generate cross-query comparison: %w", err)
	}

	spinner.Stop()

	// Save cross-query comparison
	crossQueryPath := filepath.Join(runFolder, "comparison_cross_query.txt")
	if err := output.WriteText(crossQueryPath, report); err != nil {
		return fmt.Errorf("failed to write cross-query comparison: %w", err)
	}

	printer.Success("Cross-query comparison saved to: %s", crossQueryPath)

	printer.Section("Cross-Query Comparison Summary")
	printer.Info("Total queries analyzed: %d", len(current))
	printer.Info("Comparison pairs: %d", (len(current)*(len(current)-1))/2)

	return nil
}

func parseComparisonMode(mode string) comparison.Mode {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "historical":
		return comparison.ModeHistorical
	case "cross-query", "crossquery":
		return comparison.ModeCrossQuery
	case "both":
		return comparison.ModeBoth
	default:
		return comparison.ModeBoth
	}
}
