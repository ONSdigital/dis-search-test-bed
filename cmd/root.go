package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ONSdigital/dis-search-test-bed/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile     string
	verbose     bool
	versionInfo struct {
		version string
		commit  string
		date    string
	}
)

var rootCmd = &cobra.Command{
	Use:   "search-testbed",
	Short: "Search relevance testing tool",
	Long: `A comprehensive tool for testing and comparing search algorithm 
relevance across different configurations and datasets.`,
	SilenceUsage: true,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// SetVersionInfo sets version information for the binary
func SetVersionInfo(version, commit, date string) {
	versionInfo.version = version
	versionInfo.commit = commit
	versionInfo.date = date
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default: $HOME/.search-testbed/config.yaml or ./config/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"verbose output")

	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("search-testbed %s\n", versionInfo.version)
		if verbose {
			fmt.Printf("  commit: %s\n", versionInfo.commit)
			fmt.Printf("  built:  %s\n", versionInfo.date)
		}
	},
}

func initConfig() {
	if cfgFile == "" {
		// Try home directory first
		home, err := os.UserHomeDir()
		if err == nil {
			homeConfig := filepath.Join(home, ".search-testbed", "config.yaml")
			if _, err := os.Stat(homeConfig); err == nil {
				cfgFile = homeConfig
				return
			}
		}

		// Fall back to local config
		localConfig := filepath.Join("config", "config.yaml")
		if _, err := os.Stat(localConfig); err == nil {
			cfgFile = localConfig
			return
		}

		// Use default location
		cfgFile = localConfig
	}
}

func loadConfig() (*config.Config, error) {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", cfgFile, err)
	}
	return cfg, nil
}
