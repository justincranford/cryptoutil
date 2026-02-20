// Copyright (c) 2025 Justin Cranford

// Package lint_workflow provides linting utilities for GitHub workflow files.
package lint_workflow

import (
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintWorkflowGitHubActions "cryptoutil/internal/apps/cicd/lint_workflow/github_actions"
)

// LinterFunc is a function type for individual workflow linters.
// Each linter receives a logger and a list of workflow files, returning an error if issues are found.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger, workflowFiles []string) error

// registeredLinters holds all linters to run as part of lint-workflow.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"github-actions", lintWorkflowGitHubActions.Check},
}

// Lint runs all registered workflow linters.
// It filters the provided files to only include workflow files.
// Returns an error if any linter finds issues.
func Lint(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	logger.Log("Running workflow linters...")

	// Filter to workflow files only from yml and yaml extensions.
	workflowFiles := lintWorkflowGitHubActions.FilterWorkflowFiles(filesByExtension)

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
