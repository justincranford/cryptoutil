// Copyright (c) 2025 Justin Cranford
//
//

package go_update_project_status

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"cryptoutil/internal/cmd/cicd/common"
)

// Update updates PROJECT-STATUS.md with current metrics from REQUIREMENTS-COVERAGE.md,
// TODO counts, and git commit hash.
func Update(ctx context.Context, logger *common.Logger, args []string) error {
	const (
		statusFile   = "docs/02-identityV2/PROJECT-STATUS.md"
		coverageFile = "docs/02-identityV2/REQUIREMENTS-COVERAGE.md"
	)

	logger.Log("Updating PROJECT-STATUS.md with current metrics...")

	// Read requirements coverage.
	coverage, err := parseRequirementsCoverage(coverageFile)
	if err != nil {
		return fmt.Errorf("failed to parse requirements coverage: %w", err)
	}

	// Count TODOs by severity.
	todos, err := countTODOs()
	if err != nil {
		return fmt.Errorf("failed to count TODOs: %w", err)
	}

	// Get current commit hash.
	commitHash, err := getCurrentCommitHash()
	if err != nil {
		return fmt.Errorf("failed to get commit hash: %w", err)
	}

	// Read current PROJECT-STATUS.md.
	content, err := os.ReadFile(statusFile)
	if err != nil {
		return fmt.Errorf("failed to read status file: %w", err)
	}

	// Update metrics.
	updated := updateMetrics(string(content), coverage, todos, commitHash)

	// Write updated content.
	if err := os.WriteFile(statusFile, []byte(updated), 0o644); err != nil { //nolint:mnd,gosec // Standard file permissions
		return fmt.Errorf("failed to write status file: %w", err)
	}

	logger.Log("Updated PROJECT-STATUS.md:")
	logger.Log(fmt.Sprintf("  Requirements: %.1f%% (%d/%d)", coverage.percentage, coverage.validated, coverage.total))
	logger.Log(fmt.Sprintf("  TODOs: %d CRITICAL, %d HIGH, %d MEDIUM, %d LOW", todos.critical, todos.high, todos.medium, todos.low))
	logger.Log(fmt.Sprintf("  Commit: %s", commitHash[:8])) //nolint:mnd // First 8 chars of commit hash

	return nil
}

// requirementsCoverage holds parsed requirements coverage data.
type requirementsCoverage struct {
	validated  int
	total      int
	percentage float64
}

// todoCount holds TODO counts by severity.
type todoCount struct {
	critical int
	high     int
	medium   int
	low      int
}

// parseRequirementsCoverage extracts coverage metrics from REQUIREMENTS-COVERAGE.md.
func parseRequirementsCoverage(filePath string) (*requirementsCoverage, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	coverage := &requirementsCoverage{}
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "**Total Requirements**:") {
			_, _ = fmt.Sscanf(trimmed, "**Total Requirements**: %d", &coverage.total) //nolint:errcheck // best-effort parsing
		}

		if strings.HasPrefix(trimmed, "**Validated**:") {
			_, _ = fmt.Sscanf(trimmed, "**Validated**: %d (%f%%)", &coverage.validated, &coverage.percentage) //nolint:errcheck // best-effort parsing
		}
	}

	if coverage.total == 0 {
		return nil, errors.New("failed to parse total requirements from coverage file")
	}

	return coverage, nil
}

// countTODOs counts TODO/FIXME comments by severity using grep.
func countTODOs() (*todoCount, error) {
	cmd := exec.CommandContext(context.Background(), "grep", "-rn", "--include=*.go", "TODO\\|FIXME", "internal/identity/")
	output, _ := cmd.CombinedOutput() //nolint:errcheck // grep exit code irrelevant; output parsing matters // Ignore error (grep returns non-zero when no matches)

	todos := &todoCount{}
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		upper := strings.ToUpper(line)

		if strings.Contains(upper, "CRITICAL") {
			todos.critical++
		} else if strings.Contains(upper, "HIGH") {
			todos.high++
		} else if strings.Contains(upper, "MEDIUM") {
			todos.medium++
		} else if strings.Contains(line, "TODO") || strings.Contains(line, "FIXME") {
			todos.low++
		}
	}

	return todos, nil
}

// getCurrentCommitHash returns the current git commit hash.
func getCurrentCommitHash() (string, error) {
	cmd := exec.CommandContext(context.Background(), "git", "rev-parse", "HEAD")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git commit hash: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// updateMetrics updates PROJECT-STATUS.md content with current metrics.
func updateMetrics(content string, coverage *requirementsCoverage, todos *todoCount, commitHash string) string {
	const (
		commitHashLen = 8
		totalTODOs    = 37 // Example: calculate from todos struct
	)

	// Update timestamp.
	now := time.Now().Format("January 2, 2006")
	content = regexp.MustCompile(`\*\*Last Updated\*\*: .*`).ReplaceAllString(content, "**Last Updated**: "+now)

	// Update commit hash.
	if len(commitHash) >= commitHashLen {
		content = regexp.MustCompile(`\*\*Commit Hash\*\*: .*`).ReplaceAllString(content, "**Commit Hash**: "+commitHash[:commitHashLen])
	}

	// Update requirements coverage (overall).
	coverageLine := fmt.Sprintf("| **TOTAL** | **%d** | **%d** | **%.1f%%** | **%d** |",
		coverage.validated, coverage.total, coverage.percentage, coverage.total-coverage.validated)
	content = regexp.MustCompile(`\| \*\*TOTAL\*\* \| \*\*\d+\*\* \| \*\*\d+\*\* \| \*\*[\d.]+%\*\* \| \*\*\d+\*\* \|`).
		ReplaceAllString(content, coverageLine)

	// Update TODO counts.
	content = regexp.MustCompile(`🔴 CRITICAL \| \d+`).ReplaceAllString(content, fmt.Sprintf("🔴 CRITICAL | %d", todos.critical))
	content = regexp.MustCompile(`⚠️ HIGH \| \d+`).ReplaceAllString(content, fmt.Sprintf("⚠️ HIGH | %d", todos.high))
	content = regexp.MustCompile(`📋 MEDIUM \| \d+`).ReplaceAllString(content, fmt.Sprintf("📋 MEDIUM | %d", todos.medium))
	content = regexp.MustCompile(`ℹ️ LOW \| \d+`).ReplaceAllString(content, fmt.Sprintf("ℹ️ LOW | %d", todos.low))

	// Update total TODO count.
	totalCount := todos.critical + todos.high + todos.medium + todos.low
	content = regexp.MustCompile(`\*\*TOTAL\*\* \| \*\*\d+\*\*`).ReplaceAllString(content, fmt.Sprintf("**TOTAL** | **%d**", totalCount))

	return content
}

// CollectFiles returns all Go source files in internal/cmd/cicd/go_update_project_status.
// Used by main cicd dispatcher to get file list.
func CollectFiles() ([]string, error) {
	return []string{}, nil // No files to collect for this command
}
