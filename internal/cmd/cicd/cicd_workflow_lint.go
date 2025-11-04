// Package cicd provides CI/CD quality control checks for the cryptoutil project.
//
// This file contains the github-workflow-lint command implementation.
package cicd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type ActionException struct {
	AllowedVersions []string `json:"allowed_versions"`
	Reason          string   `json:"reason"`
}

type ActionExceptions struct {
	Exceptions map[string]ActionException `json:"exceptions"`
}

type ActionInfo struct {
	Name           string
	CurrentVersion string
	LatestVersion  string
	WorkflowFile   string
}

// checkWorkflowLint validates GitHub Actions workflow files in two ways:
//  1. Enforces repository-level workflow conventions (filename prefix "ci-", presence of a top-level
//     "name:" field, and a logging step that prints the workflow name/filename before executing jobs).
//  2. Performs the existing GitHub Actions version checks (detects outdated "uses: owner/repo@version").
//
// The function walks ".github/workflows", validates each YAML file using a lightweight text-based
// validation (regex/search) to avoid adding a YAML dependency, and then reuses the existing
// action-version logic to check for outdated actions. Any violations cause the function to print
// human-friendly messages and exit with a non-zero status to block pushes.
func checkWorkflowLint(allFiles []string) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[PERF] checkWorkflowLint started at %s\n", start.Format(time.RFC3339Nano))

	// Load action exceptions (same behavior as prior implementation)
	exceptions, err := loadActionExceptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load action exceptions: %v\n", err)

		exceptions = &ActionExceptions{Exceptions: make(map[string]ActionException)}
	}

	var (
		actions          []ActionInfo
		validationErrors []string
	)

	// Filter workflow files from allFiles
	workflowsDir := ".github/workflows"

	var workflowFiles []string

	for _, path := range allFiles {
		// Normalize path separators for cross-platform compatibility
		normalizedPath := filepath.ToSlash(path)
		if strings.HasPrefix(normalizedPath, workflowsDir) && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
			workflowFiles = append(workflowFiles, path)
		}
	}

	// Process workflow files
	for _, path := range workflowFiles {
		// Run combined validation and action parsing (single file read)
		issues, fileActions, vErr := validateAndParseWorkflowFile(path)
		if vErr != nil {
			// Non-fatal: report and continue
			validationErrors = append(validationErrors, fmt.Sprintf("Failed to validate %s: %v", path, vErr))
		}

		for _, issue := range issues {
			validationErrors = append(validationErrors, fmt.Sprintf("%s: %s", filepath.Base(path), issue))
		}

		// Add actions for version checks
		actions = append(actions, fileActions...)
	}

	// If validation errors were found, report and fail fast
	if len(validationErrors) > 0 {
		fmt.Fprintln(os.Stderr, "Workflow validation errors:")

		for _, e := range validationErrors {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}

		fmt.Fprintln(os.Stderr, "\nPlease fix the workflow files to match naming and logging conventions.")
		os.Exit(1)
	}

	// If no actions were found, report and exit (no further checks necessary)
	if len(actions) == 0 {
		fmt.Fprintln(os.Stderr, "No actions found in workflow files")

		end := time.Now()
		fmt.Fprintf(os.Stderr, "[PERF] checkWorkflowLint: duration=%v start=%s end=%s workflows=%d actions=%d\n",
			end.Sub(start), start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano), len(workflowFiles), len(actions))
		os.Exit(0)
	}

	// Remove duplicates and check versions (reuse prior logic)
	actionMap := make(map[string]ActionInfo)

	for _, action := range actions {
		key := action.Name + "@" + action.CurrentVersion
		actionMap[key] = action
	}

	// Check versions concurrently for better performance
	fmt.Fprintf(os.Stderr, "Checking %d unique actions for updates...\n", len(actionMap))

	versionCheckStart := time.Now()
	outdated, exempted, errors := checkActionVersionsConcurrently(actionMap, exceptions)
	versionCheckEnd := time.Now()

	fmt.Fprintf(os.Stderr, "Version checks completed in %.2fs\n", versionCheckEnd.Sub(versionCheckStart).Seconds())

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
			if exception, exists := exceptions.Exceptions[action.Name]; exists {
				fmt.Fprintf(os.Stderr, "  %s@%s (in %s) - %s\n",
					action.Name, action.CurrentVersion, action.WorkflowFile, exception.Reason)
			}
		}

		fmt.Fprintln(os.Stderr, "")
	}

	if len(outdated) > 0 {
		fmt.Fprintln(os.Stderr, "Found outdated GitHub Actions:")

		for _, action := range outdated {
			fmt.Fprintf(os.Stderr, "  %s@%s -> %s (in %s)\n",
				action.Name, action.CurrentVersion, action.LatestVersion, action.WorkflowFile)
		}

		fmt.Fprintln(os.Stderr, "\nPlease update to the latest versions manually.")
		os.Exit(1) // Fail to block push
	}

	fmt.Fprintln(os.Stderr, "All GitHub Actions are up to date.")

	end := time.Now()
	fmt.Fprintf(os.Stderr, "[PERF] checkWorkflowLint: duration=%v start=%s end=%s workflows=%d actions=%d outdated=%d exempted=%d\n",
		end.Sub(start), start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano), len(workflowFiles), len(actions), len(outdated), len(exempted))
}

// validateAndParseWorkflowFile performs lightweight validation and action parsing on a workflow YAML file.
// It reads the file only once and performs both validation and action extraction in a single pass.
//
// Returns validation issues (empty if file is valid), extracted actions, and any error encountered reading the file.
func validateAndParseWorkflowFile(path string) ([]string, []ActionInfo, error) {
	var issues []string

	var actions []ActionInfo

	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	content := string(contentBytes)

	// 1) Filename prefix check
	base := filepath.Base(path)
	if !strings.HasPrefix(base, "ci-") {
		issues = append(issues, "workflow filename must be prefixed with 'ci-'")
	}

	// 2) Top-level name key: look for a 'name:' at the start of a line
	nameRe := regexp.MustCompile(`(?m)^\s*name:\s*.+`)
	if !nameRe.MatchString(content) {
		issues = append(issues, "missing top-level 'name:' field (required and should be consistent across workflows)")
	}

	// 3) Logging requirement: ensure the workflow references the workflow name/filename so that jobs can log it
	// We require at least one of these tokens to be present in the file, OR the workflow-job-begin action which handles logging.
	hasLoggingReference := strings.Contains(content, "${{ github.workflow }}") ||
		strings.Contains(content, "github.workflow") ||
		strings.Contains(content, "GITHUB_WORKFLOW") ||
		strings.Contains(content, "$GITHUB_WORKFLOW") ||
		strings.Contains(content, "./.github/actions/workflow-job-begin")
	if !hasLoggingReference {
		issues = append(issues, "missing logging of workflow name/filename - include '${{ github.workflow }}' or reference 'GITHUB_WORKFLOW' in an early step, or use the ./.github/actions/workflow-job-begin action")
	}

	// 4) Extract actions for version checks (same logic as parseWorkflowFile)
	// Regex to match "uses: owner/repo@version" patterns
	re := regexp.MustCompile(`uses:\s*([^\s@]+)@([^\s]+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) >= minActionMatchGroups {
			action := ActionInfo{
				Name:           match[1],
				CurrentVersion: match[2],
				WorkflowFile:   filepath.Base(path),
			}
			actions = append(actions, action)
		}
	}

	return issues, actions, nil
}

// checkActionVersionsConcurrently checks multiple GitHub actions for updates concurrently.
// It uses goroutines to make parallel API calls, significantly reducing total execution time.
// Returns slices of outdated actions, exempted actions, and any errors encountered.
func checkActionVersionsConcurrently(actionMap map[string]ActionInfo, exceptions *ActionExceptions) ([]ActionInfo, []ActionInfo, []string) {
	type result struct {
		action   ActionInfo
		latest   string
		err      error
		exempted bool
	}

	results := make(chan result, len(actionMap))

	// Start goroutines for each action check
	for _, action := range actionMap {
		go func(act ActionInfo) {
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

			latest, err := getLatestVersion(act.Name)
			results <- result{action: act, latest: latest, err: err, exempted: isExempted}
		}(action)
	}

	// Collect results
	var outdated []ActionInfo

	var exempted []ActionInfo

	var errors []string

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
func parseWorkflowFile(path string) ([]ActionInfo, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	var actions []ActionInfo

	// Regex to match "uses: owner/repo@version" patterns
	re := regexp.MustCompile(`uses:\s*([^\s@]+)@([^\s]+)`)
	matches := re.FindAllStringSubmatch(string(content), -1)

	for _, match := range matches {
		if len(match) >= minActionMatchGroups {
			action := ActionInfo{
				Name:           match[1],
				CurrentVersion: match[2],
				WorkflowFile:   filepath.Base(path),
			}
			actions = append(actions, action)
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

// loadActionExceptions loads action exceptions from the JSON file.
// Returns an empty exceptions struct if the file doesn't exist.
func loadActionExceptions() (*ActionExceptions, error) {
	exceptionsFile := ".github/workflows-outdated-action-exemptions.json"
	if _, err := os.Stat(exceptionsFile); os.IsNotExist(err) {
		// No exceptions file, return empty exceptions
		return &ActionExceptions{Exceptions: make(map[string]ActionException)}, nil
	}

	content, err := os.ReadFile(exceptionsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read exceptions file: %w", err)
	}

	var exceptions ActionExceptions
	if err := json.Unmarshal(content, &exceptions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal exceptions JSON: %w", err)
	}

	return &exceptions, nil
}
