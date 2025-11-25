package paths

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// CreateRunFolder creates a timestamped run folder
func CreateRunFolder(baseDir string) (string, error) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	runFolder := filepath.Join(baseDir, "run_"+timestamp)

	if err := os.MkdirAll(runFolder, 0755); err != nil {
		return "", fmt.Errorf("create run folder: %w", err)
	}

	return runFolder, nil
}

// FindLatestIndex finds the most recent index.json file
func FindLatestIndex(baseDir string) (string, error) {
	pattern := filepath.Join(baseDir, "run_*", "index.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("glob pattern: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no index files found in %s", baseDir)
	}

	// Sort by modification time
	sort.Slice(matches, func(i, j int) bool {
		infoI, _ := os.Stat(matches[i])
		infoJ, _ := os.Stat(matches[j])
		return infoI.ModTime().After(infoJ.ModTime())
	})

	return matches[0], nil
}

// FindLatestResults finds the most recent results.json file
func FindLatestResults(baseDir string) (string, error) {
	pattern := filepath.Join(baseDir, "run_*", "results.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("glob pattern: %w", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no results files found in %s", baseDir)
	}

	// Sort by path (which includes timestamp)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i] > matches[j]
	})

	return matches[0], nil
}

// FindPreviousResults finds the previous results.json file
func FindPreviousResults(baseDir, currentPath string) (string, error) {
	pattern := filepath.Join(baseDir, "run_*", "results.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("glob pattern: %w", err)
	}

	if len(matches) < 2 {
		return "", fmt.Errorf("no previous results found")
	}

	// Sort by path (which includes timestamp) in descending order
	sort.Slice(matches, func(i, j int) bool {
		return matches[i] > matches[j]
	})

	// Find the previous one (not the current)
	for _, match := range matches {
		if match != currentPath {
			return match, nil
		}
	}

	return "", fmt.Errorf("no previous results found")
}

// ListRunFolders lists all run folders in the base directory
func ListRunFolders(baseDir string) ([]string, error) {
	pattern := filepath.Join(baseDir, "run_*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob pattern: %w", err)
	}

	// Filter for directories only
	var folders []string
	for _, match := range matches {
		info, err := os.Stat(match)
		if err == nil && info.IsDir() {
			folders = append(folders, match)
		}
	}

	// Sort by name (which includes timestamp)
	sort.Slice(folders, func(i, j int) bool {
		return folders[i] > folders[j]
	})

	return folders, nil
}

// ExtractTimestamp extracts timestamp from run folder name
func ExtractTimestamp(runFolder string) (time.Time, error) {
	base := filepath.Base(runFolder)
	if !strings.HasPrefix(base, "run_") {
		return time.Time{}, fmt.Errorf("invalid run folder name: %s", base)
	}

	timestampStr := strings.TrimPrefix(base, "run_")
	t, err := time.Parse("2006-01-02_15-04-05", timestampStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse timestamp: %w", err)
	}

	return t, nil
}
