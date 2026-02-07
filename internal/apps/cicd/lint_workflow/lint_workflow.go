// Copyright (c) 2025 Justin Cranford

// Package lint_workflow provides linting utilities for GitHub workflow files.
package lint_workflow

import (
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// LinterFunc is a function type for individual workflow linters.
// Each linter receives a logger and a list of workflow files, returning an error if issues are found.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger, workflowFiles []string) error

// registeredLinters holds all linters to run as part of lint-workflow.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"github-actions", lintGitHubWorkflows},
}

// Lint runs all registered workflow linters.
// It filters the provided files to only include workflow files.
// Returns an error if any linter finds issues.
func Lint(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running workflow linters...")

	// Filter to workflow files only from yml and yaml extensions.
	workflowFiles := filterWorkflowFiles(filesByExtension)

	if len(workflowFiles) == 0 {
		logger.Log("lint-workflow completed (no workflow files)")

		return nil
	}

	logger.Log(fmt.Sprintf("Found %d workflow files to lint", len(workflowFiles)))

	var errors []error

	for _, l := range registeredLinters {
		logger.Log(fmt.Sprintf("Running linter: %s", l.name))

		if err := l.linter(logger, workflowFiles); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", l.name, err))
		}
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("lint-workflow completed with %d errors", len(errors)))

		return fmt.Errorf("lint-workflow failed with %d errors", len(errors))
	}

	logger.Log("lint-workflow completed successfully")

	return nil
}

// filterWorkflowFiles returns only GitHub workflow files from the yml/yaml files in the map.
func filterWorkflowFiles(filesByExtension map[string][]string) []string {
	var workflowFiles []string

	// Check yml files.
	for _, f := range filesByExtension["yml"] {
		if isWorkflowFile(f) {
			workflowFiles = append(workflowFiles, f)
		}
	}

	// Check yaml files.
	for _, f := range filesByExtension["yaml"] {
		if isWorkflowFile(f) {
			workflowFiles = append(workflowFiles, f)
		}
	}

	return workflowFiles
}

// isWorkflowFile checks if a file path is a GitHub workflow file.
func isWorkflowFile(path string) bool {
	// Check for .github/workflows/ in the path.
	// Support both forward and backward slashes for cross-platform compatibility.
	return (len(path) > 4 && (path[len(path)-4:] == ".yml" || path[len(path)-5:] == ".yaml")) &&
		(contains(path, ".github/workflows/") || contains(path, ".github\\workflows\\"))
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

// findSubstring returns the index of substr in s, or -1 if not found.
func findSubstring(s, substr string) int {
	if len(substr) == 0 {
		return 0
	}

	if len(substr) > len(s) {
		return -1
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}

	return -1
}
