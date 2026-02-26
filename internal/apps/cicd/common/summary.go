// Copyright (c) 2025 Justin Cranford
//
//

package common

import (
	"fmt"
	"os"
	"strings"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// CommandResult tracks the execution result of a single command.
type CommandResult struct {
	Command  string
	Duration time.Duration
	Error    error
}

// PrintExecutionSummary outputs a formatted summary of command execution results.
// It displays individual command statuses, counts, and total execution time.
func PrintExecutionSummary(results []CommandResult, totalDuration time.Duration) {
	fmt.Fprintln(os.Stderr, "\n"+strings.Repeat("=", cryptoutilSharedMagic.SeparatorLength))
	fmt.Fprintln(os.Stderr, "EXECUTION SUMMARY")
	fmt.Fprintln(os.Stderr, strings.Repeat("=", cryptoutilSharedMagic.SeparatorLength))

	successCount := 0
	failureCount := 0

	for _, result := range results {
		status := cryptoutilSharedMagic.StatusSuccess
		if result.Error != nil {
			status = cryptoutilSharedMagic.StatusFailed
			failureCount++
		} else {
			successCount++
		}

		fmt.Fprintf(os.Stderr, "%s  %-45s  %8.2fs\n",
			status,
			result.Command,
			result.Duration.Seconds())
	}

	fmt.Fprintln(os.Stderr, strings.Repeat("-", cryptoutilSharedMagic.SeparatorLength))
	fmt.Fprintf(os.Stderr, "Total: %d commands  |  Passed: %d  |  Failed: %d  |  Time: %.2fs\n",
		len(results),
		successCount,
		failureCount,
		totalDuration.Seconds())
	fmt.Fprintln(os.Stderr, strings.Repeat("=", cryptoutilSharedMagic.SeparatorLength))
}

// PrintCommandSeparator outputs a visual separator between commands.
func PrintCommandSeparator() {
	fmt.Fprintln(os.Stderr, "\n"+strings.Repeat("=", cryptoutilSharedMagic.SeparatorLength)+"\n")
}

// SummaryStats calculates summary statistics from command results.
type SummaryStats struct {
	Total    int
	Passed   int
	Failed   int
	Duration time.Duration
}

// CalculateStats computes summary statistics from a slice of CommandResults.
func CalculateStats(results []CommandResult, totalDuration time.Duration) SummaryStats {
	stats := SummaryStats{
		Total:    len(results),
		Duration: totalDuration,
	}

	for _, result := range results {
		if result.Error != nil {
			stats.Failed++
		} else {
			stats.Passed++
		}
	}

	return stats
}

// HasFailures returns true if any command in the results failed.
func HasFailures(results []CommandResult) bool {
	for _, result := range results {
		if result.Error != nil {
			return true
		}
	}

	return false
}

// GetFailedCommands returns a list of command names that failed.
func GetFailedCommands(results []CommandResult) []string {
	var failed []string

	for _, result := range results {
		if result.Error != nil {
			failed = append(failed, result.Command)
		}
	}

	return failed
}
