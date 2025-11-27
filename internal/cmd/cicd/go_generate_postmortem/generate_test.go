package go_generate_postmortem_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	googleUuid "github.com/google/uuid"

	cryptoutilGoGeneratePostmortem "cryptoutil/internal/cmd/cicd/go_generate_postmortem"
)

const (
	sampleTaskDoc = `# P5.01: Automated Quality Gates

## Overview
**Objective**: Implement automated quality gates in pre-commit and pre-push hooks.

## Implementation Plan

### Phase 1: Hook Setup
Setup pre-commit hooks.

## Achievements

Successfully implemented:
- Pre-commit hooks with quality gates
- Pre-push hooks with test validation
- golangci-lint integration

## Challenges

Main challenges:
- Hook ordering complexity
- Cross-platform compatibility
- Performance optimization

## Lessons Learned

Key lessons:
1. Hook ordering matters for performance
2. Use native tools when possible
3. Document hook dependencies clearly

## Evidence Checklist

- [x] Pre-commit hooks configured
- [x] Pre-push hooks configured
- [x] All quality gates passing
`

	sampleRequirementsCoverage = `# Requirements Coverage Report

## Summary

### Overall Coverage
- **Total**: 65/65 (100.0%)
- **Browser**: 30/30 (100.0%)
- **Service**: 35/35 (100.0%)

### Before This Task
- **Total**: 58/65 (89.2%)

### After This Task
- **Total**: 65/65 (100.0%)
`

	sampleProjectStatus = `# PROJECT STATUS

## Current Status
Status: ❌ NOT READY

## Completion Metrics
Tasks Complete: 5/10 tasks (50%)

## Current Status (Before)
Status: ⚠️ CONDITIONAL

## Current Status (After)
Status: ✅ PRODUCTION READY
`
)

func TestGenerate_EndToEnd(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temp directory with unique subdirectory using UUIDv7.
	tempDir := t.TempDir()
	testID := googleUuid.NewString()
	taskDocsDir := filepath.Join(tempDir, "tasks", testID)
	outputPath := filepath.Join(tempDir, "output", testID, "POSTMORTEM.md")

	// Create task docs directory.
	err := os.MkdirAll(taskDocsDir, 0o755)
	require.NoError(t, err)

	// Create sample task documents.
	taskFiles := map[string]string{
		"P5.01-automated-quality-gates.md":   sampleTaskDoc,
		"P5.02-single-source-enforcement.md": sampleTaskDoc,
		"P5.03-progressive-validation.md":    sampleTaskDoc,
		"P5.04-client-secret-rotation.md":    sampleTaskDoc,
		"P5.05-requirements-validation.md":   sampleTaskDoc,
	}

	for filename, content := range taskFiles {
		taskPath := filepath.Join(taskDocsDir, filename)
		err = os.WriteFile(taskPath, []byte(content), 0o600)
		require.NoError(t, err)
	}

	// Create requirements coverage file.
	reqCoveragePath := filepath.Join(tempDir, "REQUIREMENTS-COVERAGE.md")
	err = os.WriteFile(reqCoveragePath, []byte(sampleRequirementsCoverage), 0o600)
	require.NoError(t, err)

	// Create project status file.
	projectStatusPath := filepath.Join(tempDir, "PROJECT-STATUS.md")
	err = os.WriteFile(projectStatusPath, []byte(sampleProjectStatus), 0o600)
	require.NoError(t, err)

	// Generate post-mortem.
	opts := cryptoutilGoGeneratePostmortem.Options{
		StartTask:                "P5.01",
		EndTask:                  "P5.05",
		OutputPath:               outputPath,
		TaskDocsDir:              taskDocsDir,
		RequirementsCoveragePath: reqCoveragePath,
		ProjectStatusPath:        projectStatusPath,
	}

	err = cryptoutilGoGeneratePostmortem.Generate(ctx, opts)
	require.NoError(t, err)

	// Verify output file created.
	_, err = os.Stat(outputPath)
	require.NoError(t, err)

	// Verify output content.
	outputBytes, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	output := string(outputBytes)

	// Verify structure - all 8 sections present.
	require.Contains(t, output, "# Post-Mortem: P5.01 - P5.05")
	require.Contains(t, output, "## 1. Executive Summary")
	require.Contains(t, output, "## 2. Task-by-Task Analysis")
	require.Contains(t, output, "## 3. Pattern Validations")
	require.Contains(t, output, "## 4. Process Improvements")
	require.Contains(t, output, "## 5. Gap Analysis")
	require.Contains(t, output, "## 6. Template Improvements")
	require.Contains(t, output, "## 7. Automation Opportunities")
	require.Contains(t, output, "## 8. Evidence Quality Assessment")

	// Verify task data extracted.
	require.Contains(t, output, "P5.01: Automated Quality Gates")
	require.Contains(t, output, "Successfully implemented")
	require.Contains(t, output, "Main challenges")
	require.Contains(t, output, "Key lessons")

	// Verify metrics extracted.
	require.Contains(t, output, "65/65 (100.0%)")
	require.Contains(t, output, "89.2%")

	// Verify status transitions extracted.
	require.Contains(t, output, "CONDITIONAL")
	require.Contains(t, output, "PRODUCTION READY")
}

func TestGenerate_InvalidTaskRange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tempDir := t.TempDir()

	opts := cryptoutilGoGeneratePostmortem.Options{
		StartTask:   "P5.99",
		EndTask:     "P5.01",
		OutputPath:  filepath.Join(tempDir, "POSTMORTEM.md"),
		TaskDocsDir: tempDir,
	}

	err := cryptoutilGoGeneratePostmortem.Generate(ctx, opts)
	require.Error(t, err)
	require.Contains(t, err.Error(), "start task number")
}

func TestGenerate_MissingTaskDocsDir(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tempDir := t.TempDir()

	opts := cryptoutilGoGeneratePostmortem.Options{
		StartTask:   "P5.01",
		EndTask:     "P5.05",
		OutputPath:  filepath.Join(tempDir, "POSTMORTEM.md"),
		TaskDocsDir: filepath.Join(tempDir, "nonexistent"),
	}

	err := cryptoutilGoGeneratePostmortem.Generate(ctx, opts)
	require.Error(t, err)
	require.Contains(t, err.Error(), "task documents directory")
}

func TestGenerate_OutputDirCreation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temp directory with unique subdirectory using UUIDv7.
	tempDir := t.TempDir()
	testID := googleUuid.NewString()
	taskDocsDir := filepath.Join(tempDir, "tasks", testID)
	outputPath := filepath.Join(tempDir, "deep", "nested", "path", testID, "POSTMORTEM.md")

	// Create task docs directory.
	err := os.MkdirAll(taskDocsDir, 0o755)
	require.NoError(t, err)

	// Create single task doc.
	taskPath := filepath.Join(taskDocsDir, "P5.01-test.md")
	err = os.WriteFile(taskPath, []byte(sampleTaskDoc), 0o600)
	require.NoError(t, err)

	// Generate post-mortem (should create output directory).
	opts := cryptoutilGoGeneratePostmortem.Options{
		StartTask:   "P5.01",
		EndTask:     "P5.01",
		OutputPath:  outputPath,
		TaskDocsDir: taskDocsDir,
	}

	err = cryptoutilGoGeneratePostmortem.Generate(ctx, opts)
	require.NoError(t, err)

	// Verify output file created in deep path.
	_, err = os.Stat(outputPath)
	require.NoError(t, err)
}

func TestGenerate_NoTaskDocsInRange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temp directory with unique subdirectory using UUIDv7.
	tempDir := t.TempDir()
	testID := googleUuid.NewString()
	taskDocsDir := filepath.Join(tempDir, "tasks", testID)

	// Create task docs directory.
	err := os.MkdirAll(taskDocsDir, 0o755)
	require.NoError(t, err)

	// Create task outside range.
	taskPath := filepath.Join(taskDocsDir, "P5.99-outside-range.md")
	err = os.WriteFile(taskPath, []byte(sampleTaskDoc), 0o600)
	require.NoError(t, err)

	// Generate post-mortem with range P5.01-P5.05 (no matching tasks).
	opts := cryptoutilGoGeneratePostmortem.Options{
		StartTask:   "P5.01",
		EndTask:     "P5.05",
		OutputPath:  filepath.Join(tempDir, "POSTMORTEM.md"),
		TaskDocsDir: taskDocsDir,
	}

	err = cryptoutilGoGeneratePostmortem.Generate(ctx, opts)
	require.NoError(t, err)

	// Verify output file created (with empty task list).
	outputBytes, err := os.ReadFile(opts.OutputPath)
	require.NoError(t, err)

	output := string(outputBytes)
	require.Contains(t, output, "No tasks found in range")
}

func TestGenerate_MissingOptionalFiles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temp directory with unique subdirectory using UUIDv7.
	tempDir := t.TempDir()
	testID := googleUuid.NewString()
	taskDocsDir := filepath.Join(tempDir, "tasks", testID)
	outputPath := filepath.Join(tempDir, "output", testID, "POSTMORTEM.md")

	// Create task docs directory.
	err := os.MkdirAll(taskDocsDir, 0o755)
	require.NoError(t, err)

	// Create single task doc.
	taskPath := filepath.Join(taskDocsDir, "P5.01-test.md")
	err = os.WriteFile(taskPath, []byte(sampleTaskDoc), 0o600)
	require.NoError(t, err)

	// Generate post-mortem WITHOUT requirements coverage or project status files.
	opts := cryptoutilGoGeneratePostmortem.Options{
		StartTask:                "P5.01",
		EndTask:                  "P5.01",
		OutputPath:               outputPath,
		TaskDocsDir:              taskDocsDir,
		RequirementsCoveragePath: filepath.Join(tempDir, "nonexistent-requirements.md"),
		ProjectStatusPath:        filepath.Join(tempDir, "nonexistent-status.md"),
	}

	err = cryptoutilGoGeneratePostmortem.Generate(ctx, opts)
	require.NoError(t, err)

	// Verify output file created.
	outputBytes, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	output := string(outputBytes)

	// Verify structure still present.
	require.Contains(t, output, "# Post-Mortem: P5.01 - P5.01")
	require.Contains(t, output, "## 1. Executive Summary")

	// Verify task data still extracted.
	require.Contains(t, output, "P5.01: Test")
}

func TestGenerate_DefaultPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temp directory with unique subdirectory using UUIDv7.
	tempDir := t.TempDir()
	testID := googleUuid.NewString()
	taskDocsDir := filepath.Join(tempDir, "tasks", testID)
	outputPath := filepath.Join(tempDir, "output", testID, "POSTMORTEM.md")

	// Create task docs directory.
	err := os.MkdirAll(taskDocsDir, 0o755)
	require.NoError(t, err)

	// Create single task doc.
	taskPath := filepath.Join(taskDocsDir, "P5.01-test.md")
	err = os.WriteFile(taskPath, []byte(sampleTaskDoc), 0o600)
	require.NoError(t, err)

	// Generate post-mortem with default paths (empty strings = use defaults).
	opts := cryptoutilGoGeneratePostmortem.Options{
		StartTask:   "P5.01",
		EndTask:     "P5.01",
		OutputPath:  outputPath,
		TaskDocsDir: taskDocsDir,
		// RequirementsCoveragePath and ProjectStatusPath use defaults.
	}

	err = cryptoutilGoGeneratePostmortem.Generate(ctx, opts)
	require.NoError(t, err)

	// Verify output file created.
	_, err = os.Stat(outputPath)
	require.NoError(t, err)
}

func TestGenerate_AchievementsParsing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temp directory with unique subdirectory using UUIDv7.
	tempDir := t.TempDir()
	testID := googleUuid.NewString()
	taskDocsDir := filepath.Join(tempDir, "tasks", testID)
	outputPath := filepath.Join(tempDir, "output", testID, "POSTMORTEM.md")

	// Create task docs directory.
	err := os.MkdirAll(taskDocsDir, 0o755)
	require.NoError(t, err)

	// Create task doc with specific achievements.
	taskContent := `# P5.01: Test Task

## Achievements

Successfully implemented:
- Feature A with tests
- Feature B with documentation
- Feature C with integration

## Challenges

No major challenges.

## Lessons Learned

No specific lessons.
`

	taskPath := filepath.Join(taskDocsDir, "P5.01-test.md")
	err = os.WriteFile(taskPath, []byte(taskContent), 0o600)
	require.NoError(t, err)

	// Generate post-mortem.
	opts := cryptoutilGoGeneratePostmortem.Options{
		StartTask:   "P5.01",
		EndTask:     "P5.01",
		OutputPath:  outputPath,
		TaskDocsDir: taskDocsDir,
	}

	err = cryptoutilGoGeneratePostmortem.Generate(ctx, opts)
	require.NoError(t, err)

	// Verify achievements extracted.
	outputBytes, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	output := string(outputBytes)

	require.Contains(t, output, "Feature A with tests")
	require.Contains(t, output, "Feature B with documentation")
	require.Contains(t, output, "Feature C with integration")
}

func TestGenerate_ChallengesParsing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temp directory with unique subdirectory using UUIDv7.
	tempDir := t.TempDir()
	testID := googleUuid.NewString()
	taskDocsDir := filepath.Join(tempDir, "tasks", testID)
	outputPath := filepath.Join(tempDir, "output", testID, "POSTMORTEM.md")

	// Create task docs directory.
	err := os.MkdirAll(taskDocsDir, 0o755)
	require.NoError(t, err)

	// Create task doc with specific challenges.
	taskContent := `# P5.01: Test Task

## Achievements

No major achievements.

## Challenges

Main challenges:
1. Performance bottleneck in database queries
2. Cross-platform compatibility issues
3. Complex error handling requirements

## Lessons Learned

No specific lessons.
`

	taskPath := filepath.Join(taskDocsDir, "P5.01-test.md")
	err = os.WriteFile(taskPath, []byte(taskContent), 0o600)
	require.NoError(t, err)

	// Generate post-mortem.
	opts := cryptoutilGoGeneratePostmortem.Options{
		StartTask:   "P5.01",
		EndTask:     "P5.01",
		OutputPath:  outputPath,
		TaskDocsDir: taskDocsDir,
	}

	err = cryptoutilGoGeneratePostmortem.Generate(ctx, opts)
	require.NoError(t, err)

	// Verify challenges extracted.
	outputBytes, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	output := string(outputBytes)

	require.Contains(t, output, "Performance bottleneck")
	require.Contains(t, output, "Cross-platform compatibility")
	require.Contains(t, output, "Complex error handling")
}

func TestGenerate_LessonsLearnedParsing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temp directory with unique subdirectory using UUIDv7.
	tempDir := t.TempDir()
	testID := googleUuid.NewString()
	taskDocsDir := filepath.Join(tempDir, "tasks", testID)
	outputPath := filepath.Join(tempDir, "output", testID, "POSTMORTEM.md")

	// Create task docs directory.
	err := os.MkdirAll(taskDocsDir, 0o755)
	require.NoError(t, err)

	// Create task doc with specific lessons.
	taskContent := `# P5.01: Test Task

## Achievements

No major achievements.

## Challenges

No major challenges.

## Lessons Learned

Key lessons:
- Use parallel testing for better coverage
- Pre-commit hooks improve code quality
- Documentation should be written alongside code
`

	taskPath := filepath.Join(taskDocsDir, "P5.01-test.md")
	err = os.WriteFile(taskPath, []byte(taskContent), 0o600)
	require.NoError(t, err)

	// Generate post-mortem.
	opts := cryptoutilGoGeneratePostmortem.Options{
		StartTask:   "P5.01",
		EndTask:     "P5.01",
		OutputPath:  outputPath,
		TaskDocsDir: taskDocsDir,
	}

	err = cryptoutilGoGeneratePostmortem.Generate(ctx, opts)
	require.NoError(t, err)

	// Verify lessons learned extracted.
	outputBytes, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	output := string(outputBytes)

	require.Contains(t, output, "parallel testing")
	require.Contains(t, output, "Pre-commit hooks")
	require.Contains(t, output, "Documentation should be written")
}

func TestGenerate_EvidenceChecklistParsing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temp directory with unique subdirectory using UUIDv7.
	tempDir := t.TempDir()
	testID := googleUuid.NewString()
	taskDocsDir := filepath.Join(tempDir, "tasks", testID)
	outputPath := filepath.Join(tempDir, "output", testID, "POSTMORTEM.md")

	// Create task docs directory.
	err := os.MkdirAll(taskDocsDir, 0o755)
	require.NoError(t, err)

	// Create task doc with evidence checklist.
	taskContent := `# P5.01: Test Task

## Achievements

No major achievements.

## Challenges

No major challenges.

## Lessons Learned

No specific lessons.

## Evidence Checklist

- [x] Pre-commit hooks configured
- [x] Pre-push hooks configured
- [x] All quality gates passing
- [x] Documentation updated
- [x] Tests passing with ≥85% coverage
`

	taskPath := filepath.Join(taskDocsDir, "P5.01-test.md")
	err = os.WriteFile(taskPath, []byte(taskContent), 0o600)
	require.NoError(t, err)

	// Generate post-mortem.
	opts := cryptoutilGoGeneratePostmortem.Options{
		StartTask:   "P5.01",
		EndTask:     "P5.01",
		OutputPath:  outputPath,
		TaskDocsDir: taskDocsDir,
	}

	err = cryptoutilGoGeneratePostmortem.Generate(ctx, opts)
	require.NoError(t, err)

	// Verify evidence checklist extracted.
	outputBytes, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	output := string(outputBytes)

	require.Contains(t, output, "Pre-commit hooks configured")
	require.Contains(t, output, "Pre-push hooks configured")
	require.Contains(t, output, "All quality gates passing")
}

func TestGenerate_MultipleTasksSorted(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temp directory with unique subdirectory using UUIDv7.
	tempDir := t.TempDir()
	testID := googleUuid.NewString()
	taskDocsDir := filepath.Join(tempDir, "tasks", testID)
	outputPath := filepath.Join(tempDir, "output", testID, "POSTMORTEM.md")

	// Create task docs directory.
	err := os.MkdirAll(taskDocsDir, 0o755)
	require.NoError(t, err)

	// Create tasks in random order (should be sorted in output).
	taskFiles := map[string]string{
		"P5.05-task-five.md":  "# P5.05: Task Five\n\n## Achievements\nFive",
		"P5.01-task-one.md":   "# P5.01: Task One\n\n## Achievements\nOne",
		"P5.03-task-three.md": "# P5.03: Task Three\n\n## Achievements\nThree",
		"P5.02-task-two.md":   "# P5.02: Task Two\n\n## Achievements\nTwo",
		"P5.04-task-four.md":  "# P5.04: Task Four\n\n## Achievements\nFour",
	}

	for filename, content := range taskFiles {
		taskPath := filepath.Join(taskDocsDir, filename)
		err = os.WriteFile(taskPath, []byte(content), 0o600)
		require.NoError(t, err)
	}

	// Generate post-mortem.
	opts := cryptoutilGoGeneratePostmortem.Options{
		StartTask:   "P5.01",
		EndTask:     "P5.05",
		OutputPath:  outputPath,
		TaskDocsDir: taskDocsDir,
	}

	err = cryptoutilGoGeneratePostmortem.Generate(ctx, opts)
	require.NoError(t, err)

	// Verify tasks appear in sorted order.
	outputBytes, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	output := string(outputBytes)

	// Find positions of task IDs in output.
	posOne := strings.Index(output, "P5.01")
	posTwo := strings.Index(output, "P5.02")
	posThree := strings.Index(output, "P5.03")
	posFour := strings.Index(output, "P5.04")
	posFive := strings.Index(output, "P5.05")

	// Verify sorted order (earlier tasks appear before later tasks).
	require.Less(t, posOne, posTwo, "P5.01 should appear before P5.02")
	require.Less(t, posTwo, posThree, "P5.02 should appear before P5.03")
	require.Less(t, posThree, posFour, "P5.03 should appear before P5.04")
	require.Less(t, posFour, posFive, "P5.04 should appear before P5.05")
}

func TestGenerate_RequirementsMetricsExtraction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temp directory with unique subdirectory using UUIDv7.
	tempDir := t.TempDir()
	testID := googleUuid.NewString()
	taskDocsDir := filepath.Join(tempDir, "tasks", testID)
	outputPath := filepath.Join(tempDir, "output", testID, "POSTMORTEM.md")

	// Create task docs directory.
	err := os.MkdirAll(taskDocsDir, 0o755)
	require.NoError(t, err)

	// Create single task doc.
	taskPath := filepath.Join(taskDocsDir, "P5.01-test.md")
	err = os.WriteFile(taskPath, []byte(sampleTaskDoc), 0o600)
	require.NoError(t, err)

	// Create requirements coverage file with specific metrics.
	reqCoverageContent := `# Requirements Coverage

## Summary

### Overall Coverage
- **Total**: 100/120 (83.3%)

### Before This Task
- **Total**: 85/120 (70.8%)

### After This Task
- **Total**: 100/120 (83.3%)
`

	reqCoveragePath := filepath.Join(tempDir, "REQUIREMENTS-COVERAGE.md")
	err = os.WriteFile(reqCoveragePath, []byte(reqCoverageContent), 0o600)
	require.NoError(t, err)

	// Generate post-mortem.
	opts := cryptoutilGoGeneratePostmortem.Options{
		StartTask:                "P5.01",
		EndTask:                  "P5.01",
		OutputPath:               outputPath,
		TaskDocsDir:              taskDocsDir,
		RequirementsCoveragePath: reqCoveragePath,
	}

	err = cryptoutilGoGeneratePostmortem.Generate(ctx, opts)
	require.NoError(t, err)

	// Verify metrics extracted.
	outputBytes, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	output := string(outputBytes)

	require.Contains(t, output, "100/120 (83.3%)")
	require.Contains(t, output, "70.8%")
	require.Contains(t, output, "83.3%")
}

func TestGenerate_ProjectStatusTransitionsExtraction(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temp directory with unique subdirectory using UUIDv7.
	tempDir := t.TempDir()
	testID := googleUuid.NewString()
	taskDocsDir := filepath.Join(tempDir, "tasks", testID)
	outputPath := filepath.Join(tempDir, "output", testID, "POSTMORTEM.md")

	// Create task docs directory.
	err := os.MkdirAll(taskDocsDir, 0o755)
	require.NoError(t, err)

	// Create single task doc.
	taskPath := filepath.Join(taskDocsDir, "P5.01-test.md")
	err = os.WriteFile(taskPath, []byte(sampleTaskDoc), 0o600)
	require.NoError(t, err)

	// Create project status file with specific transitions.
	projectStatusContent := `# PROJECT STATUS

## Current Status
Status: ✅ PRODUCTION READY

## Completion Metrics
Tasks Complete: 10/10 tasks (100%)

## Status Transitions
- Database Schema: ⚠️ CONDITIONAL → ✅ PRODUCTION READY
- Test Coverage: ❌ NOT READY → ✅ PRODUCTION READY
- Documentation: ⚠️ CONDITIONAL → ✅ PRODUCTION READY
`

	projectStatusPath := filepath.Join(tempDir, "PROJECT-STATUS.md")
	err = os.WriteFile(projectStatusPath, []byte(projectStatusContent), 0o600)
	require.NoError(t, err)

	// Generate post-mortem.
	opts := cryptoutilGoGeneratePostmortem.Options{
		StartTask:         "P5.01",
		EndTask:           "P5.01",
		OutputPath:        outputPath,
		TaskDocsDir:       taskDocsDir,
		ProjectStatusPath: projectStatusPath,
	}

	err = cryptoutilGoGeneratePostmortem.Generate(ctx, opts)
	require.NoError(t, err)

	// Verify transitions extracted.
	outputBytes, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	output := string(outputBytes)

	require.Contains(t, output, "Database Schema")
	require.Contains(t, output, "Test Coverage")
	require.Contains(t, output, "Documentation")
	require.Contains(t, output, "CONDITIONAL")
	require.Contains(t, output, "NOT READY")
}
