// Copyright (c) 2025 Justin Cranford
//
//

package go_update_direct_dependencies

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cryptoutil/internal/cmd/cicd/common"
	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilFiles "cryptoutil/internal/common/util/files"
)

const (
	cacheHitState = "cache_hit"
)

// Update checks for outdated Go dependencies and fails if any are found.
// It supports two modes: direct dependencies only (go-update-direct-dependencies) or all dependencies (go-update-all-dependencies).
// This command uses caching to avoid repeated expensive checks and returns an error if outdated dependencies are found.
func Update(logger *common.Logger, mode cryptoutilMagic.DepCheckMode) error {
	modeName := cryptoutilMagic.ModeNameDirect
	if mode == cryptoutilMagic.DepCheckAll {
		modeName = cryptoutilMagic.ModeNameAll
	}

	cacheFile := cryptoutilMagic.DepCacheFileName

	// Get go.mod and go.sum file stats - needed for cache timestamp comparison
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

	// Check cache first
	cacheUsed, cacheState, err := checkAndUseDepCache(cacheFile, modeName, goModStat, goSumStat, logger)
	if err != nil {
		return fmt.Errorf("cached dependency check failed: %w", err)
	}

	if cacheUsed {
		return nil
	}

	// Log the cache miss reason
	logger.Log(fmt.Sprintf("Cache miss: %s", cacheState))

	// Cache miss or expired, perform actual check
	logger.Log("Performing fresh dependency check")

	// Run go list -u -m all to check for outdated dependencies
	cmd := exec.CommandContext(context.Background(), "go", "list", "-u", "-m", "all")

	output, err := cmd.Output()
	if err != nil {
		logger.Log(fmt.Sprintf("Error checking dependencies: %v", err))

		return fmt.Errorf("failed to check dependencies: %w", err)
	}

	// Get direct dependencies for the check (only read go.mod once)
	var directDeps map[string]bool

	if mode == cryptoutilMagic.DepCheckDirect {
		// Read go.mod file to get direct dependencies (optimization: read once, use for both stat and parsing)
		goModContent, err := os.ReadFile("go.mod")
		if err != nil {
			logger.Log(fmt.Sprintf("Error reading go.mod: %v", err))

			return fmt.Errorf("failed to read go.mod for direct dependencies: %w", err)
		}

		directDeps, err = getDirectDependencies(goModContent)
		if err != nil {
			logger.Log(fmt.Sprintf("Error parsing direct dependencies: %v", err))

			return fmt.Errorf("failed to parse direct dependencies: %w", err)
		}
	}

	// Use the extracted function for the core logic
	outdated, err := checkDependencyUpdates(mode, string(output), directDeps)
	if err != nil {
		logger.Log(fmt.Sprintf("Error checking dependency updates: %v", err))

		return fmt.Errorf("failed to check dependency updates: %w", err)
	}

	// Save results to cache
	cache := cryptoutilMagic.DepCache{
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

	logger.Log("goUpdateDeps completed")

	return nil
}

// checkDependencyUpdates analyzes dependency update information and returns outdated dependencies.
// It takes the mode, file stats, go list output, and direct dependencies map as inputs to enable testing with mock data.
// Returns a slice of outdated dependency strings and an error if the check fails.
func checkDependencyUpdates(mode cryptoutilMagic.DepCheckMode, goListOutput string, directDeps map[string]bool) ([]string, error) {
	// Optimize: avoid trim and split overhead for empty output
	if goListOutput == "" {
		return []string{}, nil
	}

	lines := strings.Split(goListOutput, "\n")
	// Preallocate with reasonable capacity to avoid multiple allocations
	allOutdated := make([]string, 0, 16)

	// Check for lines containing [v...] indicating available updates
	// Optimize: use single pass with minimal string operations
	for _, line := range lines {
		// Skip empty lines early
		if len(line) == 0 {
			continue
		}

		// Check for update marker "[v"
		updateIdx := strings.Index(line, "[v")
		if updateIdx == -1 {
			continue
		}

		// Verify closing bracket exists
		if strings.Contains(line[updateIdx:], "]") {
			allOutdated = append(allOutdated, line)
		}
	}

	// If no outdated dependencies found, return early
	if len(allOutdated) == 0 {
		return []string{}, nil
	}

	var outdated []string

	if mode == cryptoutilMagic.DepCheckDirect {
		// For direct mode, only check dependencies that are explicitly listed in go.mod
		// Filter to only direct dependencies
		outdated = make([]string, 0, len(allOutdated)) // Preallocate with upper bound

		for _, dep := range allOutdated {
			// Extract module name from the line (format: "module/path v1.2.3 [v1.2.4]")
			// Optimize: use index-based parsing instead of Fields which allocates
			spaceIdx := strings.Index(dep, " ")
			if spaceIdx > 0 {
				moduleName := dep[:spaceIdx]
				if directDeps[moduleName] {
					outdated = append(outdated, dep)
				}
			}
		}
	} else {
		// For all mode, check all dependencies
		outdated = allOutdated
	}

	return outdated, nil
}

// loadDepCache loads dependency cache from the specified file.
// Returns the cache and any error encountered.
func loadDepCache(cacheFile, mode string) (*cryptoutilMagic.DepCache, error) {
	content, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache cryptoutilMagic.DepCache
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
func saveDepCache(cacheFile string, cache cryptoutilMagic.DepCache) error {
	content, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache JSON: %w", err)
	}

	// Ensure output directory exists.
	cacheDir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(cacheDir, cryptoutilMagic.CICDOutputDirPermissions); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	if err := cryptoutilFiles.WriteFile(cacheFile, content, cryptoutilMagic.CacheFilePermissions); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// getDirectDependencies parses go.mod content and returns a map of direct dependencies.
func getDirectDependencies(goModContent []byte) (map[string]bool, error) {
	// Preallocate map with reasonable capacity
	directDeps := make(map[string]bool, 32)

	// Convert to string once
	content := string(goModContent)
	lines := strings.Split(content, "\n")

	inRequireBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines early
		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true

			continue
		}

		if line == ")" {
			inRequireBlock = false

			continue
		}

		if inRequireBlock || strings.HasPrefix(line, "require ") {
			// Skip indirect dependencies early
			if strings.Contains(line, "// indirect") {
				continue
			}

			// Parse lines like "github.com/example/package v1.2.3"
			// Optimize: use index-based parsing instead of Fields
			spaceIdx := strings.Index(line, " ")
			if spaceIdx > 0 {
				moduleName := line[:spaceIdx]
				// Handle "require " prefix for single-line requires
				if strings.HasPrefix(moduleName, "require") {
					// Skip the "require " prefix
					moduleName = strings.TrimPrefix(line, "require ")
					spaceIdx = strings.Index(moduleName, " ")

					if spaceIdx > 0 {
						moduleName = moduleName[:spaceIdx]
					}
				}

				directDeps[moduleName] = true
			}
		}
	}

	return directDeps, nil
}

// checkAndUseDepCache checks if valid cached dependency results exist and uses them if available.
// Returns true if cache was used (and function should return), false if cache miss occurred.
// Also returns a descriptive string about the cache state for logging and an error if outdated deps were found in cache.
func checkAndUseDepCache(cacheFile, modeName string, goModStat, goSumStat os.FileInfo, logger *common.Logger) (bool, string, error) {
	cache, err := loadDepCache(cacheFile, modeName)
	if err != nil {
		// Determine the specific reason for cache miss
		if errors.Is(err, os.ErrNotExist) {
			return false, "cache_not_exists", nil
		}
		// Check if it's a mode mismatch (our custom error)
		if strings.Contains(err.Error(), "cache mode mismatch") {
			return false, "cache_mode_mismatch", nil
		}
		// Other errors (JSON parsing, etc.)
		return false, "cache_invalid", nil
	}

	// Cache loaded successfully, check if it's still valid
	cacheValid := true

	var invalidReason string

	if goModStat != nil && !goModStat.ModTime().Equal(cache.GoModModTime) {
		cacheValid = false
		invalidReason = "go.mod modified"
	}

	if goSumStat != nil && !goSumStat.ModTime().Equal(cache.GoSumModTime) {
		cacheValid = false

		if invalidReason == "" {
			invalidReason = "go.sum modified"
		} else {
			invalidReason = "go.mod and go.sum modified"
		}
	}

	if !cacheValid {
		return false, fmt.Sprintf("cache_expired_files (%s)", invalidReason), nil
	}

	if time.Since(cache.LastCheck) >= time.Hour {
		return false, fmt.Sprintf("cache_expired_time (age: %.1fs)", time.Since(cache.LastCheck).Seconds()), nil
	}

	// Cache is valid and fresh - use it
	logger.Log(fmt.Sprintf("Using cached dependency check results (age: %.1fs)", time.Since(cache.LastCheck).Seconds()))

	if len(cache.OutdatedDeps) > 0 {
		logger.Log(fmt.Sprintf("Found outdated Go dependencies (cached, checking %s)", modeName))

		for _, dep := range cache.OutdatedDeps {
			fmt.Fprintln(os.Stderr, dep)
		}

		fmt.Fprintln(os.Stderr, "\nPlease run 'go get -u ./...' to update dependencies manually.")

		return true, cacheHitState, fmt.Errorf("outdated dependencies found in cache")
	}

	logger.Log(fmt.Sprintf("All %s Go dependencies are up to date (cached)", modeName))

	logger.Log("goUpdateDeps completed (cached)")

	return true, cacheHitState, nil
}
