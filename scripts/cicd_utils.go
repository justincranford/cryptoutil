package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

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

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: go run scripts/cicd_utils.go <command> [command...]\n\nCommands:\n  go-dependency-versions    - Check Go dependencies\n  github-action-versions     - Check GitHub Actions versions\n\nExamples:\n  go run scripts/cicd_utils.go go-dependency-versions\n  go run scripts/cicd_utils.go github-action-versions\n  go run scripts/cicd_utils.go go-dependency-versions github-action-versions\n")
		os.Exit(1)
	}

	// Process all commands provided as arguments
	for i := 1; i < len(os.Args); i++ {
		command := os.Args[i]
		fmt.Fprintf(os.Stderr, "Executing command: %s\n", command)

		switch command {
		case "go-dependency-versions":
			checkDeps()
		case "github-action-versions":
			checkActions()
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\nCommands:\n  go-dependency-versions    - Check Go dependencies\n  github-action-versions     - Check GitHub Actions versions\n", command)
			os.Exit(1)
		}

		// Add a separator between multiple commands
		if i < len(os.Args)-1 {
			fmt.Fprintln(os.Stderr, "\n"+strings.Repeat("=", 50)+"\n")
		}
	}
}

func checkDeps() {
	// Run go list -u -m all to check for outdated dependencies
	cmd := exec.Command("go", "list", "-u", "-m", "all")
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking dependencies: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	outdated := []string{}

	// Check for lines containing [v...] indicating available updates
	for _, line := range lines {
		if strings.Contains(line, "[v") && strings.Contains(line, "]") {
			outdated = append(outdated, line)
		}
	}

	if len(outdated) > 0 {
		fmt.Fprintln(os.Stderr, "Found outdated Go dependencies:")
		for _, dep := range outdated {
			fmt.Fprintln(os.Stderr, dep)
		}
		fmt.Fprintln(os.Stderr, "\nPlease run 'go get -u ./...' to update dependencies manually.")
		os.Exit(1) // Fail to block push
	}

	fmt.Fprintln(os.Stderr, "All Go dependencies are up to date.")
}

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

func checkActions() {
	// Load action exceptions
	exceptions, err := loadActionExceptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load action exceptions: %v\n", err)
		exceptions = &ActionExceptions{Exceptions: make(map[string]ActionException)}
	}

	// Find all workflow files
	workflowsDir := ".github/workflows"
	if _, err := os.Stat(workflowsDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "No .github/workflows directory found\n")
		os.Exit(0)
	}

	var actions []ActionInfo

	// Walk through workflow files
	err = filepath.Walk(workflowsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
			fileActions, err := parseWorkflowFile(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to parse %s: %v\n", path, err)
				return nil
			}
			actions = append(actions, fileActions...)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking workflows directory: %v\n", err)
		os.Exit(1)
	}

	if len(actions) == 0 {
		fmt.Fprintln(os.Stderr, "No actions found in workflow files")
		os.Exit(0)
	}

	// Remove duplicates and check versions
	actionMap := make(map[string]ActionInfo)
	for _, action := range actions {
		key := action.Name + "@" + action.CurrentVersion
		actionMap[key] = action
	}

	var outdated []ActionInfo
	var errors []string
	var exempted []ActionInfo

	for _, action := range actionMap {
		// Check if this action is exempted
		isExempted := false
		if exception, exists := exceptions.Exceptions[action.Name]; exists {
			for _, allowedVersion := range exception.AllowedVersions {
				if action.CurrentVersion == allowedVersion {
					exempted = append(exempted, action)
					isExempted = true
					break
				}
			}
		}

		if isExempted {
			continue
		}

		latest, err := getLatestVersion(action.Name)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to check %s: %v", action.Name, err))
			continue
		}

		if isOutdated(action.CurrentVersion, latest) {
			action.LatestVersion = latest
			outdated = append(outdated, action)
		}
	}

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
			fmt.Fprintf(os.Stderr, "  %s@%s â†’ %s (in %s)\n",
				action.Name, action.CurrentVersion, action.LatestVersion, action.WorkflowFile)
		}
		fmt.Fprintln(os.Stderr, "\nPlease update to the latest versions manually.")
		os.Exit(1) // Fail to block push
	}

	fmt.Fprintln(os.Stderr, "All GitHub Actions are up to date.")
}

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
		if len(match) >= 3 {
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

func getLatestVersion(actionName string) (string, error) {
	// GitHub API has rate limits, so add a delay
	time.Sleep(200 * time.Millisecond)

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", actionName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Use GitHub token if available to increase rate limit
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Set User-Agent as recommended by GitHub API
	req.Header.Set("User-Agent", "check-script")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close HTTP response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode == 404 {
		// Some actions might not have releases, try tags
		return getLatestTag(actionName)
	}

	if resp.StatusCode == 403 {
		return "", fmt.Errorf("GitHub API rate limit exceeded (403). Set GITHUB_TOKEN environment variable to increase limit")
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", fmt.Errorf("failed to unmarshal GitHub release JSON: %w", err)
	}

	return release.TagName, nil
}

func getLatestTag(actionName string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/tags", actionName)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request for tags: %w", err)
	}

	// Use GitHub token if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("User-Agent", "check-script")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make HTTP request for tags: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close HTTP response body: %v\n", closeErr)
		}
	}()

	if resp.StatusCode == 403 {
		return "", fmt.Errorf("GitHub API rate limit exceeded (403). Set GITHUB_TOKEN environment variable to increase limit")
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read tags response body: %w", err)
	}

	var tags []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &tags); err != nil {
		return "", fmt.Errorf("failed to unmarshal GitHub tags JSON: %w", err)
	}

	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found")
	}

	// Return the first tag (should be the latest)
	return tags[0].Name, nil
}

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
