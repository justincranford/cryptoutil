// Copyright (c) 2025 Justin Cranford
//
//

package go_update_project_status_v2

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// Options configures PROJECT-STATUS.md update behavior.
type Options struct {
	// StatusFile is path to PROJECT-STATUS.md (default: docs/02-identityV2/PROJECT-STATUS.md).
	StatusFile string
	// CoverageFile is path to REQUIREMENTS-COVERAGE.md (default: docs/02-identityV2/REQUIREMENTS-COVERAGE.md).
	CoverageFile string
	// TaskDocsDir is directory containing task documents (default: docs/02-identityV2/passthru5/).
	TaskDocsDir string
	// TestPackage is package for test execution (default: ./internal/identity/...).
	TestPackage string
}

// CoverageMetrics holds requirements coverage statistics.
type CoverageMetrics struct {
	TotalValidated    int
	TotalRequirements int
	CoveragePercent   float64
	Critical          int
	CriticalTotal     int
	High              int
	HighTotal         int
	Medium            int
	MediumTotal       int
	Low               int
	LowTotal          int
	UncoveredReqs     []string
	TaskCoverage      map[string]float64
}

// TestMetrics holds test execution statistics.
type TestMetrics struct {
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int
	Coverage     float64
}

// RecentActivity holds git activity statistics.
type RecentActivity struct {
	LastCommitHash string
	LastCommitDate time.Time
	CommitsLast7d  int
	ActiveTasks    []string
}

// Update updates PROJECT-STATUS.md with current metrics from REQUIREMENTS-COVERAGE.md,
// test results, coverage reports, and git commit history.
func Update(ctx context.Context, opts Options) error {
	// Apply defaults.
	if opts.StatusFile == "" {
		opts.StatusFile = cryptoutilMagic.CICDIdentityProjectStatusPath
	}

	if opts.CoverageFile == "" {
		opts.CoverageFile = cryptoutilMagic.CICDIdentityRequirementsCoveragePath
	}

	if opts.TaskDocsDir == "" {
		opts.TaskDocsDir = cryptoutilMagic.CICDIdentityTaskDocsDir
	}

	if opts.TestPackage == "" {
		opts.TestPackage = "./internal/identity/..."
	}

	// Parse requirements coverage metrics.
	coverage, err := parseRequirementsCoverage(opts.CoverageFile)
	if err != nil {
		return fmt.Errorf("failed to parse requirements coverage: %w", err)
	}

	// Run tests and collect metrics.
	testMetrics, err := runTestsWithCoverage(ctx, opts.TestPackage)
	if err != nil {
		return fmt.Errorf("failed to run tests: %w", err)
	}

	// Parse git commit history for recent activity.
	activity, err := parseRecentActivity(ctx)
	if err != nil {
		return fmt.Errorf("failed to parse recent activity: %w", err)
	}

	// Determine production readiness based on metrics.
	readinessStatus := determineProductionReadiness(coverage, testMetrics)

	// Read current PROJECT-STATUS.md content.
	content, err := os.ReadFile(opts.StatusFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", opts.StatusFile, err)
	}

	// Update sections in PROJECT-STATUS.md.
	updated := string(content)
	updated = updateRequirementsCoverageSection(updated, coverage)
	updated = updateTaskSpecificCoverageSection(updated, coverage)
	updated = updateProductionReadinessSection(updated, readinessStatus)
	updated = updateRecentActivitySection(updated, activity)
	updated = updateLastUpdatedSection(updated, activity.LastCommitHash)

	// Write updated content back to file.
	if err := os.WriteFile(opts.StatusFile, []byte(updated), cryptoutilMagic.CICDOutputFilePermissions); err != nil {
		return fmt.Errorf("failed to write %s: %w", opts.StatusFile, err)
	}

	return nil
}

// parseRequirementsCoverage extracts metrics from REQUIREMENTS-COVERAGE.md.
func parseRequirementsCoverage(path string) (*CoverageMetrics, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	text := string(content)
	metrics := &CoverageMetrics{
		TaskCoverage: make(map[string]float64),
	}

	// Extract total coverage: "Total: 65/65 (100.0%)".
	totalPattern := regexp.MustCompile(`Total:\s+(\d+)/(\d+)\s+\((\d+(?:\.\d+)?)%\)`)
	if matches := totalPattern.FindStringSubmatch(text); len(matches) == cryptoutilMagic.RequirementsTotalPatternGroups {
		metrics.TotalValidated, _ = strconv.Atoi(matches[1])           //nolint:errcheck // Regex validated format
		metrics.TotalRequirements, _ = strconv.Atoi(matches[2])        //nolint:errcheck // Regex validated format
		metrics.CoveragePercent, _ = strconv.ParseFloat(matches[3], 64) //nolint:errcheck // Regex validated format
	}

	// Extract priority breakdowns: "- 🔴 CRITICAL: 10/10 (100.0%)".
	priorityPattern := regexp.MustCompile(`-\s+🔴\s+CRITICAL:\s+(\d+)/(\d+)\s+\((\d+(?:\.\d+)?)%\)`)
	if matches := priorityPattern.FindStringSubmatch(text); len(matches) == cryptoutilMagic.RequirementsPriorityPatternGroups {
		metrics.Critical, _ = strconv.Atoi(matches[1])      //nolint:errcheck // Regex validated format
		metrics.CriticalTotal, _ = strconv.Atoi(matches[2]) //nolint:errcheck // Regex validated format
	}

	priorityPattern = regexp.MustCompile(`-\s+🟠\s+HIGH:\s+(\d+)/(\d+)\s+\((\d+(?:\.\d+)?)%\)`)
	if matches := priorityPattern.FindStringSubmatch(text); len(matches) == cryptoutilMagic.RequirementsPriorityPatternGroups {
		metrics.High, _ = strconv.Atoi(matches[1])      //nolint:errcheck // Regex validated format
		metrics.HighTotal, _ = strconv.Atoi(matches[2]) //nolint:errcheck // Regex validated format
	}

	priorityPattern = regexp.MustCompile(`-\s+🟡\s+MEDIUM:\s+(\d+)/(\d+)\s+\((\d+(?:\.\d+)?)%\)`)
	if matches := priorityPattern.FindStringSubmatch(text); len(matches) == cryptoutilMagic.RequirementsPriorityPatternGroups {
		metrics.Medium, _ = strconv.Atoi(matches[1])      //nolint:errcheck // Regex validated format
		metrics.MediumTotal, _ = strconv.Atoi(matches[2]) //nolint:errcheck // Regex validated format
	}

	priorityPattern = regexp.MustCompile(`-\s+🟡\s+LOW:\s+(\d+)/(\d+)\s+\((\d+(?:\.\d+)?)%\)`)
	if matches := priorityPattern.FindStringSubmatch(text); len(matches) == cryptoutilMagic.RequirementsPriorityPatternGroups {
		metrics.Low, _ = strconv.Atoi(matches[1])      //nolint:errcheck // Regex validated format
		metrics.LowTotal, _ = strconv.Atoi(matches[2]) //nolint:errcheck // Regex validated format
	}

	// Extract task-specific coverage: "### Task P5.01 Coverage: 10/12 (83.3%)".
	taskPattern := regexp.MustCompile(`###\s+Task\s+(P\d+\.\d+)\s+Coverage:\s+\d+/\d+\s+\((\d+(?:\.\d+)?)%\)`)

	for _, match := range taskPattern.FindAllStringSubmatch(text, -1) {
		if len(match) == cryptoutilMagic.RequirementsTaskCoveragePatternGroups {
			taskID := match[1]

			coverage, _ := strconv.ParseFloat(match[2], 64) //nolint:errcheck // Regex validated format
			metrics.TaskCoverage[taskID] = coverage
		}
	}

	// Extract uncovered requirements: "- [ ] R01.01: Description".
	uncoveredPattern := regexp.MustCompile(`-\s+\[\s+\]\s+(R\d+\.\d+):\s+[^\n]+`)

	for _, match := range uncoveredPattern.FindAllStringSubmatch(text, -1) {
		if len(match) == cryptoutilMagic.RequirementsUncoveredPatternGroups {
			metrics.UncoveredReqs = append(metrics.UncoveredReqs, match[1])
		}
	}

	return metrics, nil
}

// runTestsWithCoverage executes tests and collects coverage metrics.
func runTestsWithCoverage(ctx context.Context, packagePattern string) (*TestMetrics, error) {
	// Create temp file for coverage output.
	coverageFile, err := os.CreateTemp("", "coverage-*.out")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp coverage file: %w", err)
	}

	defer func() { _ = os.Remove(coverageFile.Name()) }() //nolint:errcheck // Temp file cleanup

	// Run tests with coverage.
	cmd := exec.CommandContext(ctx, "go", "test", packagePattern, "-cover", "-coverprofile="+coverageFile.Name())

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("test execution failed: %w\nOutput: %s", err, string(output))
	}

	// Parse test output: "PASS    cryptoutil/internal/identity/domain    0.123s    coverage: 85.3% of statements".
	metrics := &TestMetrics{}
	testOutputPattern := regexp.MustCompile(`coverage:\s+(\d+(?:\.\d+)?)%\s+of\s+statements`)

	if matches := testOutputPattern.FindStringSubmatch(string(output)); len(matches) == cryptoutilMagic.TestCoveragePatternGroups {
		metrics.Coverage, _ = strconv.ParseFloat(matches[1], 64) //nolint:errcheck // Regex validated format
	}

	// Count passed/failed tests from output.
	passedPattern := regexp.MustCompile(`(?m)^ok\s+`)
	metrics.PassedTests = len(passedPattern.FindAllString(string(output), -1))

	failedPattern := regexp.MustCompile(`(?m)^FAIL\s+`)
	metrics.FailedTests = len(failedPattern.FindAllString(string(output), -1))

	metrics.TotalTests = metrics.PassedTests + metrics.FailedTests

	return metrics, nil
}

// parseRecentActivity extracts git commit history and active tasks.
func parseRecentActivity(ctx context.Context) (*RecentActivity, error) {
	activity := &RecentActivity{}

	// Get last commit hash.
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "HEAD")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit hash: %w", err)
	}

	activity.LastCommitHash = strings.TrimSpace(string(output))[:cryptoutilMagic.GitShortHashLength]

	// Get last commit date.
	cmd = exec.CommandContext(ctx, "git", "log", "-1", "--format=%ct")

	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit date: %w", err)
	}

	timestamp, _ := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64) //nolint:errcheck // Git output validated
	activity.LastCommitDate = time.Unix(timestamp, 0)

	// Count commits in last 7 days.
	since := time.Now().AddDate(0, 0, -cryptoutilMagic.GitRecentActivityDays)
	cmd = exec.CommandContext(ctx, "git", "rev-list", "--count", "--since="+since.Format(time.RFC3339), "HEAD")

	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to count recent commits: %w", err)
	}

	activity.CommitsLast7d, _ = strconv.Atoi(strings.TrimSpace(string(output))) //nolint:errcheck // Git output validated

	// Extract active tasks from recent commit messages: "feat(P5.07): ...".
	cmd = exec.CommandContext(ctx, "git", "log", "--since="+since.Format(time.RFC3339), "--format=%s")

	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit messages: %w", err)
	}

	taskPattern := regexp.MustCompile(`\(P\d+\.\d+\)`)
	tasksSeen := make(map[string]bool)

	for _, match := range taskPattern.FindAllString(string(output), -1) {
		taskID := strings.Trim(match, "()")

		if !tasksSeen[taskID] {
			tasksSeen[taskID] = true
			activity.ActiveTasks = append(activity.ActiveTasks, taskID)
		}
	}

	return activity, nil
}

// determineProductionReadiness calculates readiness status based on metrics.
func determineProductionReadiness(coverage *CoverageMetrics, tests *TestMetrics) string {
	// ✅ PRODUCTION READY: ≥85% requirements coverage AND ≥85% test coverage AND 0 failed tests.
	if coverage.CoveragePercent >= cryptoutilMagic.RequirementsProductionReadyThreshold &&
		tests.Coverage >= cryptoutilMagic.TestCoverageProductionReadyThreshold &&
		tests.FailedTests == 0 {
		return "✅ PRODUCTION READY"
	}

	// ⚠️ CONDITIONAL: Coverage ≥80% OR passing tests but <85% coverage.
	if coverage.CoveragePercent >= cryptoutilMagic.RequirementsConditionalThreshold ||
		tests.Coverage >= cryptoutilMagic.TestCoverageConditionalThreshold {
		return "⚠️ CONDITIONAL"
	}

	// ❌ NOT READY: Coverage <80% OR failing tests.
	return "❌ NOT READY"
}

// updateRequirementsCoverageSection replaces Requirements Coverage table in PROJECT-STATUS.md.
func updateRequirementsCoverageSection(content string, metrics *CoverageMetrics) string {
	replacement := fmt.Sprintf(`### Requirements Coverage

| Priority   | Validated | Total | Coverage |
|------------|-----------|-------|----------|
| 🔴 CRITICAL | %d        | %d    | %.1f%%   |
| 🟠 HIGH     | %d        | %d    | %.1f%%   |
| 🟡 MEDIUM   | %d        | %d    | %.1f%%   |
| 🟢 LOW      | %d        | %d    | %.1f%%   |
| **TOTAL**   | **%d**    | **%d**| **%.1f%%** |`,
		metrics.Critical, metrics.CriticalTotal, float64(metrics.Critical)/float64(metrics.CriticalTotal)*cryptoutilMagic.PercentMultiplier,
		metrics.High, metrics.HighTotal, float64(metrics.High)/float64(metrics.HighTotal)*cryptoutilMagic.PercentMultiplier,
		metrics.Medium, metrics.MediumTotal, float64(metrics.Medium)/float64(metrics.MediumTotal)*cryptoutilMagic.PercentMultiplier,
		metrics.Low, metrics.LowTotal, float64(metrics.Low)/float64(metrics.LowTotal)*cryptoutilMagic.PercentMultiplier,
		metrics.TotalValidated, metrics.TotalRequirements, metrics.CoveragePercent)

	// Replace section from "### Requirements Coverage" to next "###" heading.
	pattern := regexp.MustCompile(`(?s)###\s+Requirements Coverage.*?(###|$)`)

	return pattern.ReplaceAllString(content, replacement+"\n\n$1")
}

// updateTaskSpecificCoverageSection replaces Task-Specific Coverage list in PROJECT-STATUS.md.
func updateTaskSpecificCoverageSection(content string, metrics *CoverageMetrics) string {
	lines := make([]string, 0, len(metrics.TaskCoverage))

	for taskID, coverage := range metrics.TaskCoverage {
		emoji := "✅"

		if coverage < cryptoutilMagic.RequirementsTaskMinimumThreshold {
			emoji = "⚠️"
		}

		lines = append(lines, fmt.Sprintf("- %s **%s**: %.1f%% coverage", emoji, taskID, coverage))
	}

	replacement := "### Task-Specific Coverage\n\n" + strings.Join(lines, "\n")

	// Replace section from "### Task-Specific Coverage" to next "###" heading.
	pattern := regexp.MustCompile(`(?s)###\s+Task-Specific Coverage.*?(###|$)`)

	return pattern.ReplaceAllString(content, replacement+"\n\n$1")
}

// updateProductionReadinessSection replaces Production Readiness indicator in PROJECT-STATUS.md.
func updateProductionReadinessSection(content string, readinessStatus string) string {
	replacement := "## Current Status\n\n**Production Readiness**: " + readinessStatus

	// Replace section from "## Current Status" to next "##" heading.
	pattern := regexp.MustCompile(`(?s)##\s+Current Status.*?(##|$)`)

	return pattern.ReplaceAllString(content, replacement+"\n\n$1")
}

// updateRecentActivitySection replaces Recent Activity section in PROJECT-STATUS.md.
func updateRecentActivitySection(content string, activity *RecentActivity) string {
	replacement := fmt.Sprintf(`### Recent Activity

- Last updated: %s
- Commits in last 7 days: %d
- Active tasks: %s`,
		activity.LastCommitDate.Format(cryptoutilMagic.DateFormatYYYYMMDD),
		activity.CommitsLast7d,
		strings.Join(activity.ActiveTasks, ", "))

	// Replace section from "### Recent Activity" to next "###" heading.
	pattern := regexp.MustCompile(`(?s)###\s+Recent Activity.*?(###|$)`)

	return pattern.ReplaceAllString(content, replacement+"\n\n$1")
}

// updateLastUpdatedSection replaces Last Updated timestamp and commit hash.
func updateLastUpdatedSection(content string, commitHash string) string {
	now := time.Now().Format(cryptoutilMagic.DateFormatYYYYMMDD)
	replacement := fmt.Sprintf("**Last Updated**: %s (Commit: %s)", now, commitHash)

	// Replace line starting with "**Last Updated**:".
	pattern := regexp.MustCompile(`\*\*Last Updated\*\*:.*`)

	return pattern.ReplaceAllString(content, replacement)
}
