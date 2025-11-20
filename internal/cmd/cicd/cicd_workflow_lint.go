// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"
	"cryptoutil/internal/cmd/cicd/common"
)

// checkWorkflowLintWithError is a wrapper that returns an error instead of calling os.Exit.
func checkWorkflowLintWithError(logger *common.Logger, allFiles []string) error {
	workflowActionExceptions, err := loadWorkflowActionExceptions()
	if err != nil {
		logger.Log(fmt.Sprintf("Warning: Failed to load action exceptions: %v", err))

		workflowActionExceptions = &WorkflowActionExceptions{Exceptions: make(map[string]WorkflowActionException)}
	}

	workflowsActionDetails, err := validateAndGetWorkflowActionsDetails(logger, allFiles)
	if err != nil {
		return fmt.Errorf("workflow validation failed: %w", err)
	}

	// If no actions found, nothing to check
	if len(workflowsActionDetails) == 0 {
		return nil
	}

	// Check versions concurrently for better performance
	logger.Log(fmt.Sprintf("Checking %d unique actions for updates", len(workflowsActionDetails)))

	versionCheckStart := time.Now().UTC()
	outdated, exempted, errors := checkActionVersionsConcurrently(logger, workflowsActionDetails, workflowActionExceptions)
	versionCheckEnd := time.Now().UTC()

	logger.Log(fmt.Sprintf("Version checks completed in %.2fs", versionCheckEnd.Sub(versionCheckStart).Seconds()))

	// Report results
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

	logger.Log("checkWorkflowLint completed")

	return nil
}

func validateAndGetWorkflowActionsDetails(logger *common.Logger, allFiles []string) (map[string]WorkflowActionDetails, error) {
	workflowsActionDetails := make(map[string]WorkflowActionDetails)

	var allValidationErrors []string

	for _, workflowFile := range filterWorkflowFiles(allFiles) {
		workflowActionDetails, workflowValidationErrors, vErr := validateAndParseWorkflowFile(workflowFile)
		if vErr != nil {
			allValidationErrors = append(allValidationErrors, fmt.Sprintf("Failed to validate %s: %v", workflowFile, vErr))
		}

		for _, issue := range workflowValidationErrors {
			allValidationErrors = append(allValidationErrors, fmt.Sprintf("%s: %s", filepath.Base(workflowFile), issue))
		}

		// Merge workflow action details, combining workflow files for duplicate actions
		for key, newWorkflowActionDetails := range workflowActionDetails {
			if existingWorkflowActionDetails, exists := workflowsActionDetails[key]; exists {
				// Merge workflow files lists
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

	// If no actions were found, return empty map (not an error)
	if len(workflowsActionDetails) == 0 {
		fmt.Fprintln(os.Stderr, "No actions found in workflow files")

		logger.Log("checkWorkflowLint completed (no actions)")

		return workflowsActionDetails, nil
	}

	return workflowsActionDetails, nil
}

func filterWorkflowFiles(allFiles []string) []string {
	workflowFiles := make([]string, 0)

	for _, workflowFile := range allFiles {
		normalizedFilePath := filepath.ToSlash(workflowFile)
		// Check if path contains .github/workflows/ (not just prefix) to support test temp directories.
		if strings.Contains(normalizedFilePath, cryptoutilMagic.WorkflowsDir+"/") && (strings.HasSuffix(normalizedFilePath, ".yml") || strings.HasSuffix(normalizedFilePath, ".yaml")) {
			workflowFiles = append(workflowFiles, normalizedFilePath)
		}
	}

	return workflowFiles
}

func validateAndParseWorkflowFile(workflowFile string) (map[string]WorkflowActionDetails, []string, error) {
	var validationErrors []string

	workflowActions := make(map[string]WorkflowActionDetails)

	workflowFileBytes, err := os.ReadFile(workflowFile)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	workflowFileContents := string(workflowFileBytes)

	// 1) Filename prefix check
	base := filepath.Base(workflowFile)
	if !strings.HasPrefix(base, "ci-") {
		validationErrors = append(validationErrors, "workflow filename must be prefixed with 'ci-'")
	}

	if !strings.HasSuffix(base, ".yml") && !strings.HasSuffix(base, ".yaml") {
		validationErrors = append(validationErrors, "workflow filename must have .yml or .yaml suffix")
	}

	// 2) Top-level name key: look for a 'name:' at the start of a line
	if !cryptoutilMagic.RegexWorkflowName.MatchString(workflowFileContents) {
		validationErrors = append(validationErrors, "missing top-level 'name:' field (for consistency across workflows)")
	}

	// 3) Logging requirement: ensure the workflow references the workflow name/filename so that jobs can log it
	// We require at least one of these tokens to be present in the file, OR the workflow-job-begin action which handles logging.
	hasLoggingReference := strings.Contains(workflowFileContents, "${{ github.workflow }}") ||
		strings.Contains(workflowFileContents, "github.workflow") ||
		strings.Contains(workflowFileContents, "GITHUB_WORKFLOW") ||
		strings.Contains(workflowFileContents, "$GITHUB_WORKFLOW") ||
		strings.Contains(workflowFileContents, "./.github/actions/workflow-job-begin")
	if !hasLoggingReference {
		validationErrors = append(validationErrors, "missing logging of workflow name/filename - include '${{ github.workflow }}' or reference 'GITHUB_WORKFLOW' in an early step, or use the ./.github/actions/workflow-job-begin action")
	}

	// 4) Extract actions for version checks (same logic as parseWorkflowFile)
	// Regex to match "uses: owner/repo@version" patterns
	matches := cryptoutilMagic.RegexWorkflowActionUses.FindAllStringSubmatch(workflowFileContents, -1)

	for _, match := range matches {
		if len(match) >= cryptoutilMagic.MinActionMatchGroups {
			key := match[1] + "@" + match[2]

			if existingAction, exists := workflowActions[key]; exists {
				// Append to existing workflow files list
				existingAction.WorkflowFiles = append(existingAction.WorkflowFiles, filepath.Base(workflowFile))
				workflowActions[key] = existingAction
			} else {
				// Create new entry
				workflowActions[key] = WorkflowActionDetails{
					Name:           match[1],
					CurrentVersion: match[2],
					WorkflowFiles:  []string{filepath.Base(workflowFile)},
				}
			}
		}
	}

	return workflowActions, validationErrors, nil
}

// checkActionVersionsConcurrently checks multiple GitHub actions for updates concurrently.
// It uses goroutines to make parallel API calls, significantly reducing total execution time.
// Returns slices of outdated actions, exempted actions, and any errors encountered.
func checkActionVersionsConcurrently(logger *common.Logger, actionMap map[string]WorkflowActionDetails, exceptions *WorkflowActionExceptions) ([]WorkflowActionDetails, []WorkflowActionDetails, []string) {
	type result struct {
		action   WorkflowActionDetails
		latest   string
		err      error
		exempted bool
	}

	results := make(chan result, len(actionMap))

	// Start goroutines for each action check
	for _, action := range actionMap {
		go func(act WorkflowActionDetails) {
			// Check if this action is exempted
			isExempted := false

			if exception, exists := exceptions.Exceptions[act.Name]; exists {
				for _, allowedVersion := range exception.AllowedVersions {
					if act.CurrentVersion == allowedVersion {
						results <- result{action: act, exempted: true}

						return
					}
				}
			}

			latest, err := getLatestVersion(logger, act.Name)
			results <- result{action: act, latest: latest, err: err, exempted: isExempted}
		}(action)
	}

	// Collect results
	outdated := make([]WorkflowActionDetails, 0)
	exempted := make([]WorkflowActionDetails, 0)
	errors := make([]string, 0)

	for i := 0; i < len(actionMap); i++ {
		res := <-results

		if res.exempted {
			exempted = append(exempted, res.action)

			continue
		}

		if res.err != nil {
			errors = append(errors, fmt.Sprintf("Failed to check %s: %v", res.action.Name, res.err))

			continue
		}

		if isOutdated(res.action.CurrentVersion, res.latest) {
			res.action.LatestVersion = res.latest
			outdated = append(outdated, res.action)
		}
	}

	return outdated, exempted, errors
}

// parseWorkflowFile extracts GitHub Actions from a workflow YAML file.
// Returns a slice of ActionInfo structs containing action names, versions, and workflow file names.
func parseWorkflowFile(path string) ([]WorkflowActionDetails, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	var actions []WorkflowActionDetails

	// Regex to match "uses: owner/repo@version" patterns
	matches := cryptoutilMagic.RegexWorkflowActionUses.FindAllStringSubmatch(string(content), -1)

	for _, match := range matches {
		if len(match) >= cryptoutilMagic.MinActionMatchGroups {
			workflowActionDetails := WorkflowActionDetails{
				Name:           match[1],
				CurrentVersion: match[2],
				WorkflowFiles:  []string{filepath.Base(path)},
			}
			actions = append(actions, workflowActionDetails)
		}
	}

	return actions, nil
}

// isOutdated determines if a GitHub Action version is outdated compared to the latest version.
// It handles special cases like @main, @master, and variable references.
// For major version pins (e.g., v4), it checks if the major version has increased.
// For specific versions, it performs direct string comparison.
func isOutdated(current, latest string) bool {
	// Skip checking for @main, @master, etc.
	if current == "main" || current == "master" || strings.HasPrefix(current, "$") {
		return false
	}

	// For major version pins (e.g., v4), check if latest major version is higher
	if matched, err := regexp.MatchString(`^v(\d+)$`, current); err == nil && matched {
		currentMajor := strings.TrimPrefix(current, "v")

		latestMajor := strings.TrimPrefix(latest, "v")
		if strings.Contains(latestMajor, ".") {
			// Extract major version from latest (e.g., "5.0.0" -> "5")
			parts := strings.Split(latestMajor, ".")
			latestMajor = parts[0]
		}

		return currentMajor != latestMajor
	}

	// For specific versions, simple comparison
	return current != latest
}

// loadWorkflowActionExceptions loads action exceptions from the JSON file.
// Returns an empty exceptions struct if the file doesn't exist.
func loadWorkflowActionExceptions() (*WorkflowActionExceptions, error) {
	const exceptionsFile = ".github/workflows-outdated-action-exemptions.json"

	if _, err := os.Stat(exceptionsFile); os.IsNotExist(err) {
		// No exceptions file, return empty exceptions
		return &WorkflowActionExceptions{Exceptions: make(map[string]WorkflowActionException)}, nil
	}

	content, err := os.ReadFile(exceptionsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read exceptions file: %w", err)
	}

	var exceptions WorkflowActionExceptions
	if err := json.Unmarshal(content, &exceptions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal exceptions JSON: %w", err)
	}

	return &exceptions, nil
}

type WorkflowActionException struct {
	AllowedVersions []string `json:"allowed_versions"`
	Reason          string   `json:"reason"`
}

type WorkflowActionExceptions struct {
	Exceptions map[string]WorkflowActionException `json:"exceptions"`
}

type WorkflowActionDetails struct {
	Name           string
	CurrentVersion string
	WorkflowFiles  []string
	LatestVersion  string
}
