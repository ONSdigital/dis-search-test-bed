package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "search-testbed",
	Short: "Search Test Bed - Elasticsearch query testing and comparison tool",
	Long: `Search Test Bed helps you test and compare Elasticsearch search algorithms
by managing test indexes, running queries, and generating diffs between results.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config",
		"config/config.yaml", "config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v",
		false, "verbose output")
}
