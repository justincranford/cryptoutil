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

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

const (
	// UI constants.
	separatorLength = cryptoutilMagic.UIConsoleSeparatorLength

	// GitHub API constants.
	githubAPIDelay   = cryptoutilMagic.TimeoutGitHubAPIDelay
	githubAPITimeout = cryptoutilMagic.TimeoutGitHubAPITimeout

	// Progress reporting.
	progressInterval = cryptoutilMagic.CountUIProgressInterval

	// File permissions.
	filePermissions = cryptoutilMagic.FilePermissionsDefault // Permissions for created files.

	// UUID regex pattern for validation.
	uuidRegexPattern = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`

	// Minimum number of regex match groups for action parsing.
	minActionMatchGroups = cryptoutilMagic.CountMinActionMatchGroups
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

type DepCheckMode int

const (
	DepCheckDirect DepCheckMode = iota // Check only direct dependencies
	DepCheckAll                        // Check all dependencies (direct + transitive)
	modeNameDirect = "direct"
	modeNameAll    = "all"
)

type ActionInfo struct {
	Name           string
	CurrentVersion string
	LatestVersion  string
	WorkflowFile   string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: go run scripts/cicd_utils.go <command> [command...]\n\nCommands:\n  go-update-direct-dependencies    - Check direct Go dependencies only\n  go-update-all-dependencies       - Check all Go dependencies (direct + transitive)\n  go-check-circular-package-dependencies          - Check for circular dependencies in Go packages\n  github-action-versions           - Check GitHub Actions versions\n  gofumpter                        - Custom Go source code fixes (interface{} -> any, etc.)\n  enforce-test-patterns            - Enforce test patterns (UUIDv7 usage, testify assertions)\n\nExamples:\n  go run scripts/cicd_utils.go go-update-direct-dependencies\n  go run scripts/cicd_utils.go go-update-all-dependencies\n  go run scripts/cicd_utils.go go-check-circular-package-dependencies\n  go run scripts/cicd_utils.go github-action-versions\n  go run scripts/cicd_utils.go gofumpter\n  go run scripts/cicd_utils.go enforce-test-patterns\n  go run scripts/cicd_utils.go go-update-direct-dependencies github-action-versions\n")
		os.Exit(1)
	}

	// Process all commands provided as arguments
	for i := 1; i < len(os.Args); i++ {
		command := os.Args[i]
		fmt.Fprintf(os.Stderr, "Executing command: %s\n", command)

		switch command {
		case "go-update-direct-dependencies":
			checkDeps(DepCheckDirect)
		case "go-update-all-dependencies":
			checkDeps(DepCheckAll)
		case "go-check-circular-package-dependencies":
			checkCircularDeps()
		case "github-action-versions":
			checkActions()
		case "gofumpter":
			runGofumpter()
		case "enforce-test-patterns":
			enforceTestPatterns()
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\nCommands:\n  go-update-direct-dependencies    - Check direct Go dependencies only\n  go-update-all-dependencies       - Check all Go dependencies (direct + transitive)\n  go-check-circular-package-dependencies          - Check for circular dependencies in Go packages\n  github-action-versions           - Check GitHub Actions versions\n  gofumpter                        - Custom Go source code fixes (interface{} -> any, etc.)\n  enforce-test-patterns            - Enforce test patterns (UUIDv7 usage, testify assertions)\n", command)
			os.Exit(1)
		}

		// Add a separator between multiple commands
		if i < len(os.Args)-1 {
			fmt.Fprintln(os.Stderr, "\n"+strings.Repeat("=", separatorLength)+"\n")
		}
	}
}

func checkDeps(mode DepCheckMode) {
	// Run go list -u -m all to check for outdated dependencies
	cmd := exec.Command("go", "list", "-u", "-m", "all")

	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking dependencies: %v\n", err)
		os.Exit(1)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	allOutdated := []string{}

	// Check for lines containing [v...] indicating available updates
	for _, line := range lines {
		if strings.Contains(line, "[v") && strings.Contains(line, "]") {
			allOutdated = append(allOutdated, line)
		}
	}

	var outdated []string

	if mode == DepCheckDirect {
		// For direct mode, only check dependencies that are explicitly listed in go.mod
		directDeps, err := getDirectDependencies()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading direct dependencies: %v\n", err)
			os.Exit(1)
		}

		// Filter to only direct dependencies
		for _, dep := range allOutdated {
			// Extract module name from the line (format: "module/path v1.2.3 [v1.2.4]")
			parts := strings.Fields(dep)
			if len(parts) > 0 {
				moduleName := parts[0]
				if directDeps[moduleName] {
					outdated = append(outdated, dep)
				}
			}
		}
	} else {
		// For all mode, check all dependencies
		outdated = allOutdated
	}

	if len(outdated) > 0 {
		modeName := modeNameDirect
		if mode == DepCheckAll {
			modeName = modeNameAll
		}

		fmt.Fprintf(os.Stderr, "Found outdated Go dependencies (checking %s):\n", modeName)

		for _, dep := range outdated {
			fmt.Fprintln(os.Stderr, dep)
		}

		fmt.Fprintln(os.Stderr, "\nPlease run 'go get -u ./...' to update dependencies manually.")
		os.Exit(1) // Fail to block push
	}

	modeName := "direct"
	if mode == DepCheckAll {
		modeName = "all"
	}

	fmt.Fprintf(os.Stderr, "All %s Go dependencies are up to date.\n", modeName)
}

func getDirectDependencies() (map[string]bool, error) {
	// Read go.mod file to get direct dependencies
	goModContent, err := os.ReadFile("go.mod")
	if err != nil {
		return nil, fmt.Errorf("failed to read go.mod: %w", err)
	}

	directDeps := make(map[string]bool)
	lines := strings.Split(string(goModContent), "\n")

	inRequireBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true

			continue
		}

		if line == ")" {
			inRequireBlock = false

			continue
		}

		if inRequireBlock || strings.HasPrefix(line, "require ") {
			// Parse lines like "github.com/example/package v1.2.3"
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				// Skip indirect dependencies
				if len(parts) >= 3 && parts[2] == "indirect" {
					continue
				}

				directDeps[parts[0]] = true
			}
		}
	}

	return directDeps, nil
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

func getLatestVersion(actionName string) (string, error) {
	// GitHub API has rate limits, so add a delay
	time.Sleep(githubAPIDelay)

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", actionName)

	ctx, cancel := context.WithTimeout(context.Background(), githubAPITimeout)
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

	if resp.StatusCode == http.StatusNotFound {
		// Some actions might not have releases, try tags
		return getLatestTag(actionName)
	}

	if resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("GitHub API rate limit exceeded (403). Set GITHUB_TOKEN environment variable to increase limit")
	}

	if resp.StatusCode != http.StatusOK {
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

	ctx, cancel := context.WithTimeout(context.Background(), githubAPITimeout)
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

	if resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("GitHub API rate limit exceeded (403). Set GITHUB_TOKEN environment variable to increase limit")
	}

	if resp.StatusCode != http.StatusOK {
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

func checkCircularDeps() {
	startTime := time.Now()

	fmt.Fprintln(os.Stderr, "Checking for circular dependencies in Go packages...")

	// Get all packages in the project
	fmt.Fprintln(os.Stderr, "Running: go list ./...")

	cmd := exec.Command("go", "list", "./...")

	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running go list: %v\n", err)
		os.Exit(1)
	}

	packages := strings.Split(strings.TrimSpace(string(output)), "\n")
	fmt.Fprintf(os.Stderr, "Found %d packages:\n", len(packages))

	for i, pkg := range packages {
		if strings.TrimSpace(pkg) != "" {
			fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, pkg)
		}
	}

	fmt.Fprintln(os.Stderr, "")

	// Filter out empty packages
	var validPackages []string

	for _, pkg := range packages {
		if strings.TrimSpace(pkg) != "" {
			validPackages = append(validPackages, pkg)
		}
	}

	packages = validPackages

	if len(packages) == 0 {
		fmt.Fprintln(os.Stderr, "No packages found")

		return
	}

	// Build dependency graph
	fmt.Fprintln(os.Stderr, "Building dependency graph...")

	graphStart := time.Now()
	processed := 0
	dependencyGraph := make(map[string][]string)

	for _, pkg := range packages {
		// Get imports for this package
		importCmd := exec.Command("go", "list", "-f", "{{.Imports}}", pkg)

		importOutput, err := importCmd.Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not get imports for %s: %v\n", pkg, err)

			continue
		}

		importStr := strings.TrimSpace(string(importOutput))
		if len(importStr) <= 2 { // Empty array "[]"
			dependencyGraph[pkg] = []string{} // Empty slice instead of nil
			processed++

			continue
		}

		// Parse imports (remove [ and ] and split by space)
		importStr = strings.Trim(importStr, "[]")
		if importStr == "" {
			dependencyGraph[pkg] = []string{}
			processed++

			continue
		}

		imports := strings.Split(importStr, " ")
		dependencyGraph[pkg] = imports

		processed++
		if processed%progressInterval == 0 {
			elapsed := time.Since(graphStart)
			fmt.Fprintf(os.Stderr, "Processed %d/%d packages... (%.2fs)\n", processed, len(packages), elapsed.Seconds())
		}
	}

	graphElapsed := time.Since(graphStart)
	fmt.Fprintf(os.Stderr, "Built dependency graph with %d packages (%.2fs)\n", len(dependencyGraph), graphElapsed.Seconds())

	// Show why we have fewer packages in graph
	if len(dependencyGraph) < len(packages) {
		fmt.Fprintf(os.Stderr, "Note: %d packages were excluded from graph (failed to get imports)\n", len(packages)-len(dependencyGraph))
	}

	fmt.Fprintln(os.Stderr, "")

	// Find circular dependencies using DFS
	fmt.Fprintln(os.Stderr, "Starting DFS cycle detection...")

	dfsStart := time.Now()
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)
	circularDeps := [][]string{}

	var dfs func(string, []string)
	dfs = func(pkg string, path []string) {
		// Check if package exists in dependency graph
		if _, exists := dependencyGraph[pkg]; !exists {
			return
		}

		visited[pkg] = true
		recursionStack[pkg] = true

		for _, dep := range dependencyGraph[pkg] {
			// Only check internal packages (those starting with our module name)
			if !strings.HasPrefix(dep, "cryptoutil/") {
				continue
			}

			if !visited[dep] {
				newPath := append(path, dep)
				dfs(dep, newPath)
			} else if recursionStack[dep] {
				// Found a cycle
				cycleStart := -1

				for i, p := range path {
					if p == dep {
						cycleStart = i

						break
					}
				}

				if cycleStart >= 0 {
					cycle := append(path[cycleStart:], dep)
					circularDeps = append(circularDeps, cycle)
				}
			}
		}

		recursionStack[pkg] = false
	}

	// Check each package for circular dependencies
	dfsCount := 0

	for pkg := range dependencyGraph {
		if !visited[pkg] {
			dfs(pkg, []string{pkg})

			dfsCount++
			if dfsCount%5 == 0 {
				elapsed := time.Since(dfsStart)
				fmt.Fprintf(os.Stderr, "DFS processed %d/%d packages... (%.2fs)\n", dfsCount, len(dependencyGraph), elapsed.Seconds())
			}
		}
	}

	dfsElapsed := time.Since(dfsStart)
	fmt.Fprintf(os.Stderr, "DFS completed for %d packages (%.2fs)\n", dfsCount, dfsElapsed.Seconds())
	fmt.Fprintln(os.Stderr, "")

	// Summary report
	totalElapsed := time.Since(startTime)

	fmt.Fprintf(os.Stderr, "=== CIRCULAR DEPENDENCY ANALYSIS SUMMARY ===\n")
	fmt.Fprintf(os.Stderr, "Total execution time: %.2fs\n", totalElapsed.Seconds())
	fmt.Fprintf(os.Stderr, "Packages analyzed: %d\n", len(dependencyGraph))
	fmt.Fprintf(os.Stderr, "Internal dependencies checked: %d\n", func() int {
		count := 0

		for _, deps := range dependencyGraph {
			for _, dep := range deps {
				if strings.HasPrefix(dep, "cryptoutil/") {
					count++
				}
			}
		}

		return count
	}())

	if len(circularDeps) == 0 {
		fmt.Fprintln(os.Stderr, "✅ RESULT: No circular dependencies found")
		fmt.Fprintln(os.Stderr, "All internal package dependencies are acyclic.")

		return
	}

	fmt.Fprintf(os.Stderr, "❌ RESULT: Found %d circular dependency chain(s):\n\n", len(circularDeps))

	for i, cycle := range circularDeps {
		fmt.Fprintf(os.Stderr, "Chain %d (%d packages):\n", i+1, len(cycle))

		for j, pkg := range cycle {
			prefix := "  "
			if j > 0 {
				prefix = "  → "
			}

			fmt.Fprintf(os.Stderr, "%s%s\n", prefix, pkg)
		}

		fmt.Fprintln(os.Stderr, "")
	}

	fmt.Fprintln(os.Stderr, "Circular dependencies can prevent enabling advanced linters like gomnd.")
	fmt.Fprintln(os.Stderr, "Consider refactoring to break these cycles.")
	os.Exit(1) // Fail the build
}

func enforceTestPatterns() {
	fmt.Fprintln(os.Stderr, "Enforcing test patterns (UUIDv7 usage, testify assertions)...")

	// Find all test files
	var testFiles []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, "_test.go") {
			testFiles = append(testFiles, path)
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}

	if len(testFiles) == 0 {
		fmt.Fprintln(os.Stderr, "No test files found")

		return
	}

	fmt.Fprintf(os.Stderr, "Found %d test files to check\n", len(testFiles))

	// Check each test file
	totalIssues := 0

	for _, filePath := range testFiles {
		issues := checkTestFile(filePath)

		if len(issues) > 0 {
			fmt.Fprintf(os.Stderr, "%s:\n", filePath)

			for _, issue := range issues {
				fmt.Fprintf(os.Stderr, "  - %s\n", issue)
			}

			totalIssues += len(issues)
		}
	}

	if totalIssues > 0 {
		fmt.Fprintf(os.Stderr, "\n❌ Found %d test pattern violations\n", totalIssues)
		fmt.Fprintln(os.Stderr, "Please fix the issues above to follow established test patterns.")
		os.Exit(1) // Fail the build
	} else {
		fmt.Fprintln(os.Stderr, "\n✅ All test files follow established patterns")
	}
}

func checkTestFile(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return []string{fmt.Sprintf("Error reading file: %v", err)}
	}

	var issues []string

	contentStr := string(content)

	// Pattern 1: Check for UUIDv7 usage
	// Look for uuid.New() instead of uuid.NewV7()
	if strings.Contains(contentStr, "uuid.New()") {
		issues = append(issues, "Found uuid.New() - should use uuid.NewV7() for concurrency safety")
	}

	// Pattern 2: Check for hardcoded UUIDs (basic pattern)
	uuidPattern := regexp.MustCompile(uuidRegexPattern)
	if uuidPattern.MatchString(contentStr) {
		issues = append(issues, "Found hardcoded UUID - consider using uuid.NewV7() for test data")
	}

	// Pattern 3: Check for testify usage patterns
	// Look for t.Errorf/t.Fatalf that should use require/assert
	errorfPattern := regexp.MustCompile(`t\.Errorf\([^)]+\)`)
	if errorfPattern.MatchString(contentStr) {
		matches := errorfPattern.FindAllString(contentStr, -1)
		issues = append(issues, fmt.Sprintf("Found %d instances of t.Errorf() - should use require.Errorf() or assert.Errorf()", len(matches)))
	}

	fatalfPattern := regexp.MustCompile(`t\.Fatalf\([^)]+\)`)
	if fatalfPattern.MatchString(contentStr) {
		matches := fatalfPattern.FindAllString(contentStr, -1)
		issues = append(issues, fmt.Sprintf("Found %d instances of t.Fatalf() - should use require.Fatalf() or assert.Fatalf()", len(matches)))
	}

	// Pattern 4: Check for testify imports if testify assertions are used
	hasTestifyUsage := strings.Contains(contentStr, "require.") || strings.Contains(contentStr, "assert.")
	hasTestifyImport := strings.Contains(contentStr, "github.com/stretchr/testify")

	if hasTestifyUsage && !hasTestifyImport {
		issues = append(issues, "Test file uses testify assertions but doesn't import testify")
	}

	return issues
}

func runGofumpter() {
	fmt.Fprintln(os.Stderr, "Running gofumpter - Custom Go source code fixes...")

	// Define exclusion patterns (same as pre-commit-config.yaml)
	excludedPatterns := []string{
		`_gen\.go$`,               // Generated files
		`\.pb\.go$`,               // Protocol buffer files
		`vendor/`,                 // Vendored dependencies
		`api/client`,              // Generated API client
		`api/model`,               // Generated API models
		`api/server`,              // Generated API server
		`scripts/cicd_utils\.go$`, // Exclude this file itself to avoid replacing the regex pattern
	}

	// Find all .go files
	var goFiles []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			// Check if file should be excluded
			excluded := false

			for _, pattern := range excludedPatterns {
				matched, err := regexp.MatchString(pattern, path)
				if err != nil {
					return fmt.Errorf("invalid regex pattern %s: %w", pattern, err)
				}

				if matched {
					excluded = true

					break
				}
			}

			if !excluded {
				goFiles = append(goFiles, path)
			}
		}

		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}

	if len(goFiles) == 0 {
		fmt.Fprintln(os.Stderr, "No Go files found to process")

		return
	}

	fmt.Fprintf(os.Stderr, "Found %d Go files to process\n", len(goFiles))

	// Process each file
	filesModified := 0
	totalReplacements := 0

	for _, filePath := range goFiles {
		replacements, err := processGoFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", filePath, err)

			continue
		}

		if replacements > 0 {
			filesModified++
			totalReplacements += replacements
			fmt.Fprintf(os.Stderr, "Modified %s: %d replacements\n", filePath, replacements)
		}
	}

	// Summary
	fmt.Fprintf(os.Stderr, "\n=== GOFUMPTER SUMMARY ===\n")
	fmt.Fprintf(os.Stderr, "Files processed: %d\n", len(goFiles))
	fmt.Fprintf(os.Stderr, "Files modified: %d\n", filesModified)
	fmt.Fprintf(os.Stderr, "Total replacements: %d\n", totalReplacements)

	if filesModified > 0 {
		fmt.Fprintln(os.Stderr, "\n✅ Successfully applied custom Go source code fixes")
		fmt.Fprintln(os.Stderr, "Please review and commit the changes")
		os.Exit(1) // Exit with error to indicate files were modified
	} else {
		fmt.Fprintln(os.Stderr, "\n✅ All Go files are already properly formatted")
	}
}

func processGoFile(filePath string) (int, error) {
	// Read the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	originalContent := string(content)

	// Replace interface{} with any
	// Use a regex to match interface{} as a whole word, not part of other identifiers
	// Construct the pattern to avoid self-replacement in this source file
	interfacePattern := `interface\{\}`
	re := regexp.MustCompile(interfacePattern)
	modifiedContent := re.ReplaceAllString(originalContent, "any")

	// Count actual replacements by comparing interface{} counts
	originalInterfaceCount := strings.Count(originalContent, "interface{}")
	modifiedInterfaceCount := strings.Count(modifiedContent, "interface{}")
	replacements := originalInterfaceCount - modifiedInterfaceCount

	// Only write if there were changes
	if replacements > 0 {
		err = os.WriteFile(filePath, []byte(modifiedContent), filePermissions)
		if err != nil {
			return 0, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return replacements, nil
}
