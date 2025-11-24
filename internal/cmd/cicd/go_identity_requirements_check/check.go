// Copyright (c) 2025 Justin Cranford
//
//

package go_identity_requirements_check

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cryptoutil/internal/cmd/cicd/common"
)

// Enforce runs identity requirements coverage check with optional strict mode.
// In strict mode, fails fast if coverage thresholds are not met.
//
// Flags:
//
//	--strict: Enable fail-fast mode (exit code 1 if thresholds not met)
//	--task-threshold: Per-task coverage threshold percentage (default 90)
//	--overall-threshold: Overall coverage threshold percentage (default 85)
//	--skip-slow-checks: Disable expensive validation for faster feedback
//
// Returns error if coverage validation fails in strict mode.
func Enforce(ctx context.Context, logger *common.Logger, args []string) error {
	startTime := time.Now()

	defer func() {
		elapsed := time.Since(startTime)
		logger.Log(fmt.Sprintf("requirements check completed in %v", elapsed))
	}()

	// Default thresholds.
	const (
		defaultTaskThreshold    = 90.0 // Per-task coverage threshold (%)
		defaultOverallThreshold = 85.0 // Overall coverage threshold (%)
	)

	// Parse command flags.
	fs := flag.NewFlagSet("identity-requirements-check", flag.ContinueOnError)
	strict := fs.Bool("strict", false, "Enable fail-fast mode (exit code 1 if thresholds not met)")
	taskThreshold := fs.Float64("task-threshold", defaultTaskThreshold, "Per-task coverage threshold (0-100)")
	overallThreshold := fs.Float64("overall-threshold", defaultOverallThreshold, "Overall coverage threshold (0-100)")
	skipSlowChecks := fs.Bool("skip-slow-checks", false, "Disable expensive validation for faster feedback")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate threshold ranges.
	if *taskThreshold < 0 || *taskThreshold > 100 {
		return fmt.Errorf("task-threshold must be between 0-100, got %.2f", *taskThreshold)
	}

	if *overallThreshold < 0 || *overallThreshold > 100 {
		return fmt.Errorf("overall-threshold must be between 0-100, got %.2f", *overallThreshold)
	}

	logger.Log("Requirements Coverage Check Started")
	logger.Log(fmt.Sprintf("  Strict mode: %v", *strict))
	logger.Log(fmt.Sprintf("  Task threshold: %.1f%%", *taskThreshold))
	logger.Log(fmt.Sprintf("  Overall threshold: %.1f%%", *overallThreshold))
	logger.Log(fmt.Sprintf("  Skip slow checks: %v", *skipSlowChecks))

	// Locate requirements coverage file.
	requirementsCoverage := "docs/02-identityV2/REQUIREMENTS-COVERAGE.md"
	if _, err := os.Stat(requirementsCoverage); os.IsNotExist(err) {
		return fmt.Errorf("requirements coverage file not found: %s", requirementsCoverage)
	}

	// Parse requirements coverage file.
	coverage, err := parseRequirementsCoverage(requirementsCoverage)
	if err != nil {
		return fmt.Errorf("failed to parse requirements coverage: %w", err)
	}

	logger.Log("\nRequirements Coverage Summary:")
	logger.Log(fmt.Sprintf("  Total: %d/%d validated (%.1f%%)", coverage.validated, coverage.total, coverage.percentage))
	logger.Log(fmt.Sprintf("  CRITICAL: %d/%d validated (%.1f%%)", coverage.criticalValidated, coverage.criticalTotal, coverage.criticalPercentage))
	logger.Log(fmt.Sprintf("  HIGH: %d/%d validated (%.1f%%)", coverage.highValidated, coverage.highTotal, coverage.highPercentage))
	logger.Log(fmt.Sprintf("  MEDIUM: %d/%d validated (%.1f%%)", coverage.mediumValidated, coverage.mediumTotal, coverage.mediumPercentage))
	logger.Log(fmt.Sprintf("  LOW: %d/%d validated (%.1f%%)", coverage.lowValidated, coverage.lowTotal, coverage.lowPercentage))

	// Perform additional validation if not skipping slow checks.
	var taskCoverage float64

	if !*skipSlowChecks {
		logger.Log("\nValidating task-specific coverage...")

		taskCoverage, err = validateTaskCoverage(ctx, logger, requirementsCoverage)
		if err != nil {
			logger.Log(fmt.Sprintf("  WARNING: Task coverage validation failed: %v", err))

			taskCoverage = 0 // Treat as 0% if validation fails
		} else {
			logger.Log(fmt.Sprintf("  Task coverage: %.1f%%", taskCoverage))
		}
	} else {
		logger.Log("\nSkipping task-specific validation (--skip-slow-checks enabled)")

		taskCoverage = 100.0 // Assume passing when skipped
	}

	// Check thresholds in strict mode.
	if *strict {
		var failures []string

		if coverage.percentage < *overallThreshold {
			failures = append(failures, fmt.Sprintf("overall coverage %.1f%% below threshold %.1f%%", coverage.percentage, *overallThreshold))
		}

		if !*skipSlowChecks && taskCoverage < *taskThreshold {
			failures = append(failures, fmt.Sprintf("task coverage %.1f%% below threshold %.1f%%", taskCoverage, *taskThreshold))
		}

		if len(failures) > 0 {
			logger.Log("\n❌ STRICT MODE FAILURES:")

			for _, failure := range failures {
				logger.Log(fmt.Sprintf("  - %s", failure))
			}

			return errors.New("requirements coverage below threshold")
		}

		logger.Log("\n✅ STRICT MODE PASSED: All thresholds met")
	} else {
		logger.Log("\nReport-only mode: Not enforcing thresholds")
	}

	return nil
}

// requirementsCoverage holds parsed requirements coverage data.
type requirementsCoverage struct {
	validated          int
	total              int
	percentage         float64
	criticalValidated  int
	criticalTotal      int
	criticalPercentage float64
	highValidated      int
	highTotal          int
	highPercentage     float64
	mediumValidated    int
	mediumTotal        int
	mediumPercentage   float64
	lowValidated       int
	lowTotal           int
	lowPercentage      float64
}

// parseRequirementsCoverage extracts coverage metrics from REQUIREMENTS-COVERAGE.md.
func parseRequirementsCoverage(filePath string) (*requirementsCoverage, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	coverage := &requirementsCoverage{}
	lines := strings.Split(string(content), "\n")

	// Parse summary lines.
	// Expected format:
	// **Total Requirements**: 65
	// **Validated**: 38 (58.5%)
	// **Uncovered CRITICAL**: 7
	// **Uncovered HIGH**: 13
	// **Uncovered MEDIUM**: 6
	//
	// ### CRITICAL: 15/22 (68.2%) ⚠️
	// ### HIGH: 13/26 (50.0%) ⚠️
	// ### MEDIUM: 10/16 (62.5%) ⚠️
	// ### LOW: 0/1 (0.0%) ❌
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Parse top-level summary.
		if strings.HasPrefix(trimmed, "**Total Requirements**:") {
			_, _ = fmt.Sscanf(trimmed, "**Total Requirements**: %d", &coverage.total) //nolint:errcheck // best-effort parsing

			continue
		}

		if strings.HasPrefix(trimmed, "**Validated**:") {
			_, _ = fmt.Sscanf(trimmed, "**Validated**: %d (%f%%)", &coverage.validated, &coverage.percentage) //nolint:errcheck // best-effort parsing

			continue
		}

		// Parse priority sections.
		if strings.HasPrefix(trimmed, "### CRITICAL:") {
			_, _ = fmt.Sscanf(trimmed, "### CRITICAL: %d/%d (%f%%)", &coverage.criticalValidated, &coverage.criticalTotal, &coverage.criticalPercentage) //nolint:errcheck // best-effort parsing

			continue
		}

		if strings.HasPrefix(trimmed, "### HIGH:") {
			_, _ = fmt.Sscanf(trimmed, "### HIGH: %d/%d (%f%%)", &coverage.highValidated, &coverage.highTotal, &coverage.highPercentage) //nolint:errcheck // best-effort parsing

			continue
		}

		if strings.HasPrefix(trimmed, "### MEDIUM:") {
			_, _ = fmt.Sscanf(trimmed, "### MEDIUM: %d/%d (%f%%)", &coverage.mediumValidated, &coverage.mediumTotal, &coverage.mediumPercentage) //nolint:errcheck // best-effort parsing

			continue
		}

		if strings.HasPrefix(trimmed, "### LOW:") {
			_, _ = fmt.Sscanf(trimmed, "### LOW: %d/%d (%f%%)", &coverage.lowValidated, &coverage.lowTotal, &coverage.lowPercentage) //nolint:errcheck // best-effort parsing

			continue
		}
	}

	// Validate we got total metrics.
	if coverage.total == 0 {
		return nil, errors.New("failed to parse total requirements from coverage file")
	}

	return coverage, nil
}

// validateTaskCoverage performs task-specific validation (expensive operation).
// Returns average task coverage percentage.
func validateTaskCoverage(_ context.Context, logger *common.Logger, requirementsCoverageFile string) (float64, error) {
	// Read coverage file.
	content, err := os.ReadFile(requirementsCoverageFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read coverage file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	var taskCoverages []float64

	// Parse task-specific coverage sections.
	// Expected format:
	// ### Task P4.01 Coverage
	// Validated: 8/10 (80%)
	inTaskSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect task section start.
		if strings.HasPrefix(trimmed, "### Task") && strings.Contains(trimmed, "Coverage") {
			inTaskSection = true

			continue
		}

		// Parse validated metrics in task section.
		if inTaskSection && strings.HasPrefix(trimmed, "Validated:") {
			var (
				validated, total int
				percentage       float64
			)

			if n, err := fmt.Sscanf(trimmed, "Validated: %d/%d (%f%%)", &validated, &total, &percentage); err == nil && n == 3 {
				taskCoverages = append(taskCoverages, percentage)
				logger.Log(fmt.Sprintf("  Found task coverage: %.1f%%", percentage))

				inTaskSection = false
			}
		}

		// Exit task section if we hit next section.
		if inTaskSection && strings.HasPrefix(trimmed, "##") && !strings.HasPrefix(trimmed, "###") {
			inTaskSection = false
		}
	}

	// Calculate average task coverage.
	if len(taskCoverages) == 0 {
		logger.Log("  No task-specific coverage sections found")

		return 0, errors.New("no task-specific coverage sections found")
	}

	var sum float64
	for _, coverage := range taskCoverages {
		sum += coverage
	}

	average := sum / float64(len(taskCoverages))

	logger.Log(fmt.Sprintf("  Found %d tasks, average coverage: %.1f%%", len(taskCoverages), average))

	return average, nil
}

// CollectFiles returns all Go source files in internal/identity.
// Used by main cicd dispatcher to get file list.
func CollectFiles() ([]string, error) {
	identityDir := filepath.Join("internal", "identity")

	var files []string

	err := filepath.Walk(identityDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			// Convert to forward slashes for consistency.
			files = append(files, filepath.ToSlash(path))
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}
