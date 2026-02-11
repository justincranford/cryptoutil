// Copyright (c) 2025 Justin Cranford

// Package lint_go_mod provides go.mod linting utilities for dependency management.
package lint_go_mod

import (
	"context"
	json "encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilFiles "cryptoutil/internal/shared/util/files"
)

// checkOutdatedDeps checks for outdated Go direct dependencies and fails if any are found.
// This command uses caching to avoid repeated expensive checks and returns an error if outdated dependencies are found.
func checkOutdatedDeps(logger *cryptoutilCmdCicdCommon.Logger) error {
	modeName := cryptoutilSharedMagic.ModeNameDirect

	cacheFile := cryptoutilSharedMagic.DepCacheFileName

	// Get go.mod and go.sum file stats - needed for cache timestamp comparison.
	goModStat, err := os.Stat("go.mod")
	if err != nil {
		logger.Log(fmt.Sprintf("Error reading go.mod: %v", err))

		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	goSumStat, err := os.Stat("go.sum")
	if err != nil {
		logger.Log(fmt.Sprintf("Error reading go.sum: %v", err))

		return fmt.Errorf("failed to read go.sum: %w", err)
	}

	// Check cache first.
	cacheUsed, cacheState, cacheErr := checkAndUseDepCache(cacheFile, modeName, goModStat, goSumStat, logger)
	if cacheErr != nil {
		return fmt.Errorf("cached dependency check failed: %w", cacheErr)
	}

	if cacheUsed {
		return nil
	}

	// Log the cache miss reason.
	logger.Log(fmt.Sprintf("Cache miss: %s", cacheState))

	// Cache miss or expired, perform actual check.
	logger.Log("Performing fresh dependency check")

	// Run go list -u -m all to check for outdated dependencies.
	cmd := exec.CommandContext(context.Background(), "go", "list", "-u", "-m", "all")

	output, err := cmd.Output()
	if err != nil {
		logger.Log(fmt.Sprintf("Error checking dependencies: %v", err))

		return fmt.Errorf("failed to check dependencies: %w", err)
	}

	// Get direct dependencies for the check (only read go.mod once).
	goModContent, err := os.ReadFile("go.mod")
	if err != nil {
		logger.Log(fmt.Sprintf("Error reading go.mod: %v", err))

		return fmt.Errorf("failed to read go.mod for direct dependencies: %w", err)
	}

	directDeps, err := getDirectDependencies(goModContent)
	if err != nil {
		logger.Log(fmt.Sprintf("Error parsing direct dependencies: %v", err))

		return fmt.Errorf("failed to parse direct dependencies: %w", err)
	}

	// Use the extracted function for the core logic.
	outdated, err := checkDependencyUpdates(string(output), directDeps)
	if err != nil {
		logger.Log(fmt.Sprintf("Error checking dependency updates: %v", err))

		return fmt.Errorf("failed to check dependency updates: %w", err)
	}

	// Save results to cache.
	cache := cryptoutilSharedMagic.DepCache{
		LastCheck:    time.Now().UTC(),
		GoModModTime: goModStat.ModTime(),
		GoSumModTime: goSumStat.ModTime(),
		OutdatedDeps: outdated,
		Mode:         modeName,
	}
	if err := saveDepCache(cacheFile, cache); err != nil {
		logger.Log(fmt.Sprintf("Warning: Failed to save dependency cache: %v", err))
	}

	if len(outdated) > 0 {
		logger.Log(fmt.Sprintf("Found outdated Go dependencies (checking %s)", modeName))

		for _, dep := range outdated {
			fmt.Fprintln(os.Stderr, dep)
		}

		fmt.Fprintln(os.Stderr, "\nPlease run 'go get -u ./...' to update dependencies manually.")

		return fmt.Errorf("outdated dependencies found")
	}

	fmt.Fprintf(os.Stderr, "All %s Go dependencies are up to date.\n", modeName)

	logger.Log("lint-go-mod completed")

	return nil
}

// checkDependencyUpdates analyzes dependency update information and returns outdated dependencies.
// It takes the go list output and direct dependencies map as inputs to enable testing with mock data.
// Returns a slice of outdated dependency strings and an error if the check fails.
func checkDependencyUpdates(goListOutput string, directDeps map[string]bool) ([]string, error) {
	// Optimize: avoid trim and split overhead for empty output.
	if goListOutput == "" {
		return []string{}, nil
	}

	lines := strings.Split(goListOutput, "\n")
	// Preallocate with reasonable capacity to avoid multiple allocations.
	allOutdated := make([]string, 0, 16)

	// Check for lines containing [v...] indicating available updates.
	// Optimize: use single pass with minimal string operations.
	for _, line := range lines {
		// Skip empty lines early.
		if len(line) == 0 {
			continue
		}

		// Check for update marker "[v".
		updateIdx := strings.Index(line, "[v")
		if updateIdx == -1 {
			continue
		}

		// Verify closing bracket exists.
		if strings.Contains(line[updateIdx:], "]") {
			allOutdated = append(allOutdated, line)
		}
	}

	// If no outdated dependencies found, return early.
	if len(allOutdated) == 0 {
		return []string{}, nil
	}

	// For direct mode, only check dependencies that are explicitly listed in go.mod.
	outdated := make([]string, 0, len(allOutdated)) // Preallocate with upper bound.

	for _, dep := range allOutdated {
		// Extract module name from the line (format: "module/path v1.2.3 [v1.2.4]").
		// Optimize: use index-based parsing instead of Fields which allocates.
		spaceIdx := strings.Index(dep, " ")
		if spaceIdx > 0 {
			moduleName := dep[:spaceIdx]
			if directDeps[moduleName] {
				outdated = append(outdated, dep)
			}
		}
	}

	return outdated, nil
}

// getDirectDependencies parses go.mod content and returns a map of direct dependencies.
func getDirectDependencies(goModContent []byte) (map[string]bool, error) {
	directDeps := make(map[string]bool)

	lines := strings.Split(string(goModContent), "\n")
	inRequireBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments.
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Check for require block start/end.
		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true

			continue
		}

		if inRequireBlock && line == ")" {
			inRequireBlock = false

			continue
		}

		// Parse require statements.
		if inRequireBlock || strings.HasPrefix(line, "require ") {
			// Extract module path.
			var moduleLine string
			if strings.HasPrefix(line, "require ") {
				moduleLine = strings.TrimPrefix(line, "require ")
			} else {
				moduleLine = line
			}

			// Split by space to get module path.
			parts := strings.Fields(moduleLine)
			if len(parts) >= 1 {
				modulePath := parts[0]
				// Skip indirect dependencies.
				if !strings.Contains(line, "// indirect") {
					directDeps[modulePath] = true
				}
			}
		}
	}

	return directDeps, nil
}

// checkAndUseDepCache checks if cached results can be used.
func checkAndUseDepCache(cacheFile, modeName string, goModStat, goSumStat os.FileInfo, logger *cryptoutilCmdCicdCommon.Logger) (bool, string, error) {
	cache, err := loadDepCache(cacheFile)
	if cache == nil || err != nil {
		return false, "cache not found", nil //nolint:nilerr // Intentional: cache miss is not an error for caller.
	}

	// Check if cache is for the same mode.
	if cache.Mode != modeName {
		return false, "cache mode mismatch", nil
	}

	// Check if cache is still valid.
	cacheAge := time.Since(cache.LastCheck)
	isExpired := cacheAge > cryptoutilSharedMagic.DepCacheValidDuration

	if isExpired {
		return false, "cache expired", nil
	}

	// Check if go.mod or go.sum changed.
	goModChanged := cache.GoModModTime.Before(goModStat.ModTime())
	goSumChanged := cache.GoSumModTime.Before(goSumStat.ModTime())

	if goModChanged || goSumChanged {
		return false, "go.mod or go.sum changed", nil
	}

	// Cache is valid.
	logger.Log(fmt.Sprintf("Using cached dependency check results (age: %.1fs)", cacheAge.Seconds()))

	if len(cache.OutdatedDeps) > 0 {
		for _, dep := range cache.OutdatedDeps {
			fmt.Fprintln(os.Stderr, dep)
		}

		return true, "", fmt.Errorf("outdated dependencies found (cached)")
	}

	fmt.Fprintf(os.Stderr, "All %s Go dependencies are up to date (cached).\n", modeName)

	return true, "", nil
}

// loadDepCache loads dependency cache from the specified file.
func loadDepCache(cacheFile string) (*cryptoutilSharedMagic.DepCache, error) {
	content, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache cryptoutilSharedMagic.DepCache
	if err := json.Unmarshal(content, &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache JSON: %w", err)
	}

	return &cache, nil
}

// saveDepCache saves dependency cache to the specified file.
func saveDepCache(cacheFile string, cache cryptoutilSharedMagic.DepCache) error {
	content, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache JSON: %w", err)
	}

	// Ensure output directory exists.
	cacheDir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(cacheDir, cryptoutilSharedMagic.CICDOutputDirPermissions); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := cryptoutilSharedUtilFiles.WriteFile(cacheFile, string(content), cryptoutilSharedMagic.CacheFilePermissions); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}
