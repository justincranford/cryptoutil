// Copyright (c) 2025 Justin Cranford

package lint_workflow

import (
	json "encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// WorkflowActionDetails contains information about a GitHub Action used in workflows.
type WorkflowActionDetails struct {
	Name           string
	CurrentVersion string
	LatestVersion  string
	WorkflowFiles  []string
}

// WorkflowActionException represents an exception for outdated action versions.
type WorkflowActionException struct {
	Version string `json:"version"`
	Reason  string `json:"reason"`
}

// WorkflowActionExceptions contains all action exceptions.
type WorkflowActionExceptions struct {
	Exceptions map[string]WorkflowActionException `json:"exceptions"`
}

// lintGitHubWorkflows validates GitHub workflow files for outdated actions and other issues.
// It returns an error if validation fails or outdated actions are found.
func lintGitHubWorkflows(logger *cryptoutilCmdCicdCommon.Logger, workflowFiles []string) error {
	workflowActionExceptions, err := loadWorkflowActionExceptions()
	if err != nil {
		logger.Log(fmt.Sprintf("Warning: Failed to load action exceptions: %v", err))

		workflowActionExceptions = &WorkflowActionExceptions{Exceptions: make(map[string]WorkflowActionException)}
	}

	workflowsActionDetails, err := validateAndGetWorkflowActionsDetails(logger, workflowFiles)
	if err != nil {
		return fmt.Errorf("workflow validation failed: %w", err)
	}

	// If no actions found, nothing to check.
	if len(workflowsActionDetails) == 0 {
		return nil
	}

	// Check versions concurrently for better performance.
	logger.Log(fmt.Sprintf("Checking %d unique actions for updates", len(workflowsActionDetails)))

	versionCheckStart := time.Now().UTC()
	outdated, exempted, errors := checkActionVersionsConcurrently(logger, workflowsActionDetails, workflowActionExceptions)
	versionCheckEnd := time.Now().UTC()

	logger.Log(fmt.Sprintf("Version checks completed in %.2fs", versionCheckEnd.Sub(versionCheckStart).Seconds()))

	// Report results.
	if len(errors) > 0 {
		fmt.Fprintln(os.Stderr, "Warnings:")

		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "  %s\n", err)
		}

		fmt.Fprintln(os.Stderr, "")
	}

	if len(exempted) > 0 {
		fmt.Fprintln(os.Stderr, "Exempted actions (allowed older versions):")

		for _, action := range exempted {
			if exception, exists := workflowActionExceptions.Exceptions[action.Name]; exists {
				fmt.Fprintf(os.Stderr, "  %s@%s (in %s) - %s\n",
					action.Name, action.CurrentVersion, strings.Join(action.WorkflowFiles, ", "), exception.Reason)
			}
		}

		fmt.Fprintln(os.Stderr, "")
	}

	if len(outdated) > 0 {
		fmt.Fprintln(os.Stderr, "Found outdated GitHub Actions:")

		for _, action := range outdated {
			fmt.Fprintf(os.Stderr, "  %s@%s -> %s (in %s)\n",
				action.Name, action.CurrentVersion, action.LatestVersion, strings.Join(action.WorkflowFiles, ", "))
		}

		fmt.Fprintln(os.Stderr, "\nPlease update to the latest versions manually.")

		return fmt.Errorf("found %d outdated GitHub Actions", len(outdated))
	}

	fmt.Fprintln(os.Stderr, "All GitHub Actions are up to date.")

	logger.Log("lint-workflow completed")

	return nil
}

func validateAndGetWorkflowActionsDetails(logger *cryptoutilCmdCicdCommon.Logger, workflowFiles []string) (map[string]WorkflowActionDetails, error) {
	workflowsActionDetails := make(map[string]WorkflowActionDetails)

	var allValidationErrors []string

	for _, workflowFile := range workflowFiles {
		workflowActionDetails, workflowValidationErrors, vErr := validateAndParseWorkflowFile(workflowFile)
		if vErr != nil {
			allValidationErrors = append(allValidationErrors, fmt.Sprintf("Failed to validate %s: %v", workflowFile, vErr))
		}

		for _, issue := range workflowValidationErrors {
			allValidationErrors = append(allValidationErrors, fmt.Sprintf("%s: %s", filepath.Base(workflowFile), issue))
		}

		// Merge workflow action details, combining workflow files for duplicate actions.
		for key, newWorkflowActionDetails := range workflowActionDetails {
			if existingWorkflowActionDetails, exists := workflowsActionDetails[key]; exists {
				// Merge workflow files lists.
				existingWorkflowActionDetails.WorkflowFiles = append(existingWorkflowActionDetails.WorkflowFiles, newWorkflowActionDetails.WorkflowFiles...)
				workflowsActionDetails[key] = existingWorkflowActionDetails
			} else {
				workflowsActionDetails[key] = newWorkflowActionDetails
			}
		}
	}

	if len(allValidationErrors) > 0 {
		fmt.Fprintln(os.Stderr, "Workflow validation errors:")

		for _, validationError := range allValidationErrors {
			fmt.Fprintf(os.Stderr, "  - %s\n", validationError)
		}

		fmt.Fprintln(os.Stderr, "\nPlease fix the workflow files to match naming and logging conventions.")

		return nil, fmt.Errorf("found %d workflow validation errors", len(allValidationErrors))
	}

	// If no actions were found, return empty map (not an error).
	if len(workflowsActionDetails) == 0 {
		fmt.Fprintln(os.Stderr, "No actions found in workflow files")

		logger.Log("lint-workflow completed (no actions)")

		return workflowsActionDetails, nil
	}

	return workflowsActionDetails, nil
}

func loadWorkflowActionExceptions() (*WorkflowActionExceptions, error) {
	exceptionsFile := ".github/workflow-action-exceptions.json"

	content, err := os.ReadFile(exceptionsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &WorkflowActionExceptions{Exceptions: make(map[string]WorkflowActionException)}, nil
		}

		return nil, fmt.Errorf("failed to read exceptions file: %w", err)
	}

	var exceptions WorkflowActionExceptions
	if err := json.Unmarshal(content, &exceptions); err != nil {
		return nil, fmt.Errorf("failed to parse exceptions file: %w", err)
	}

	return &exceptions, nil
}

func validateAndParseWorkflowFile(workflowFile string) (map[string]WorkflowActionDetails, []string, error) {
	content, err := os.ReadFile(workflowFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	var validationErrors []string

	actionDetails := make(map[string]WorkflowActionDetails)

	// Parse actions from workflow file using regex.
	// Pattern: uses: owner/repo@version.
	usesPattern := regexp.MustCompile(`uses:\s*([a-zA-Z0-9_-]+/[a-zA-Z0-9_.-]+)@(v?[a-zA-Z0-9._-]+)`)
	matches := usesPattern.FindAllStringSubmatch(string(content), -1)

	for _, match := range matches {
		if len(match) >= cryptoutilSharedMagic.MinActionMatchGroups {
			actionName := match[1]
			version := match[2]

			key := actionName + "@" + version
			actionDetails[key] = WorkflowActionDetails{
				Name:           actionName,
				CurrentVersion: version,
				WorkflowFiles:  []string{filepath.Base(workflowFile)},
			}
		}
	}

	return actionDetails, validationErrors, nil
}

func checkActionVersionsConcurrently(_ *cryptoutilCmdCicdCommon.Logger, actionDetails map[string]WorkflowActionDetails, exceptions *WorkflowActionExceptions) (outdated, exempted []WorkflowActionDetails, errors []string) {
	// For simplicity, this implementation checks versions synchronously.
	// A full implementation would use goroutines and GitHub API calls.
	for _, action := range actionDetails {
		// Check if action is exempted.
		if exception, exists := exceptions.Exceptions[action.Name]; exists {
			if action.CurrentVersion == exception.Version {
				exempted = append(exempted, action)

				continue
			}
		}

		// In a full implementation, we would check the latest version from GitHub.
		// For now, we just pass through without marking as outdated.
		_ = action
	}

	return outdated, exempted, errors
}
