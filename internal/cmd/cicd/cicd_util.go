// Package cicd provides common utilities for CI/CD quality control checks.
//
// This file contains shared types, constants, and utility functions used across
// different CI/CD commands. It provides common functionality for performance timing,
// file operations, command validation, and caching.
package cicd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// UI constants.
	separatorLength = 50

	// Minimum number of regex match groups for action parsing.
	minActionMatchGroups = 3

	// Cache file permissions (owner read/write only).
	cacheFilePermissions = 0o600
)

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

type DepCache struct {
	LastCheck    time.Time `json:"last_check"`
	GoModModTime time.Time `json:"go_mod_mod_time"`
	GoSumModTime time.Time `json:"go_sum_mod_time"`
	OutdatedDeps []string  `json:"outdated_deps"`
	Mode         string    `json:"mode"`
}

type PackageInfo struct {
	ImportPath string   `json:"ImportPath"`
	Imports    []string `json:"Imports"`
}

// getUsageMessage returns the usage message for the cicd command.
func getUsageMessage() string {
	return `Usage: cicd <command> [command...]

Commands:
  all-enforce-utf8                       - Enforce UTF-8 encoding without BOM
  go-enforce-test-patterns               - Enforce test patterns (UUIDv7 usage, testify assertions)
  go-enforce-any                         - Custom Go source code fixes (any -> any, etc.)
  go-check-circular-package-dependencies - Check for circular dependencies in Go packages
  go-update-direct-dependencies          - Check direct Go dependencies only
  go-update-all-dependencies             - Check all Go dependencies (direct + transitive)
  github-workflow-lint                   - Validate GitHub Actions workflow naming and structure, and check for outdated actions`
}

// validateCommands validates the provided commands for duplicates, mutually exclusive combinations,
// and empty command lists. Returns doFindAllFiles flag and any validation error.
func validateCommands(commands []string) (bool, error) {
	// Start performance timing
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[PERF] validateCommands started at %s\n", start.Format(time.RFC3339Nano))

	// Check for empty commands first (also handles nil slices since len(nil) == 0)
	if len(commands) == 0 {
		end := time.Now()
		fmt.Fprintf(os.Stderr, "[PERF] validateCommands: start=%s end=%s duration=%v (empty commands)\n",
			start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano), end.Sub(start))

		return false, fmt.Errorf("%s", getUsageMessage())
	}

	doFindAllFiles := false

	var errs []error

	commandCounts := make(map[string]int)

	// Count command occurrences and determine if file walk is needed
	for _, command := range commands {
		switch command {
		case "all-enforce-utf8":
			commandCounts[command]++
			doFindAllFiles = true
		case "go-enforce-test-patterns":
			commandCounts[command]++
			doFindAllFiles = true
		case "go-enforce-any":
			commandCounts[command]++
			doFindAllFiles = true
		case "go-check-circular-package-dependencies":
			commandCounts[command]++
		case "go-update-direct-dependencies":
			commandCounts[command]++
		case "go-update-all-dependencies":
			commandCounts[command]++
		case "github-workflow-lint":
			commandCounts[command]++
			doFindAllFiles = true
		default:
			errs = append(errs, fmt.Errorf("unknown command: %s\n\n%s", command, getUsageMessage()))
		}
	}

	// Check for duplicate commands
	for command, count := range commandCounts {
		if count > 1 {
			errs = append(errs, fmt.Errorf("command '%s' specified %d times - each command can only be used once", command, count))
		}
	}

	// Check for mutually exclusive commands
	if commandCounts["go-update-direct-dependencies"] > 0 && commandCounts["go-update-all-dependencies"] > 0 {
		errs = append(errs, fmt.Errorf("commands 'go-update-direct-dependencies' and 'go-update-all-dependencies' cannot be used together - choose one dependency update mode"))
	}

	if len(errs) > 0 {
		end := time.Now()
		fmt.Fprintf(os.Stderr, "[PERF] validateCommands: duration=%v start=%s end=%s (validation errors)\n",
			end.Sub(start), start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano))

		return false, fmt.Errorf("command validation failed: %w", errors.Join(errs...))
	}

	end := time.Now()
	fmt.Fprintf(os.Stderr, "[PERF] validateCommands: duration=%v start=%s end=%s (success)\n",
		end.Sub(start), start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano))

	return doFindAllFiles, nil
}

// collectAllFiles walks the current directory and collects all file paths.
// Returns a slice of all file paths found.
func collectAllFiles() ([]string, error) {
	var allFiles []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			allFiles = append(allFiles, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return allFiles, nil
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

// loadDepCache loads dependency cache from the specified file.
// Returns the cache and any error encountered.
func loadDepCache(cacheFile, mode string) (*DepCache, error) {
	content, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache DepCache
	if err := json.Unmarshal(content, &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache JSON: %w", err)
	}

	// Verify cache is for the same mode
	if cache.Mode != mode {
		return nil, fmt.Errorf("cache mode mismatch")
	}

	return &cache, nil
}

// saveDepCache saves dependency cache to the specified file.
func saveDepCache(cacheFile string, cache DepCache) error {
	content, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache JSON: %w", err)
	}

	if err := os.WriteFile(cacheFile, content, cacheFilePermissions); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// getDirectDependencies reads go.mod and returns a map of direct dependencies.
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
				// Skip indirect dependencies (marked with // indirect comment)
				if strings.Contains(line, "// indirect") {
					continue
				}

				directDeps[parts[0]] = true
			}
		}
	}

	return directDeps, nil
}
