// Package cicd provides CI/CD quality control checks for the cryptoutil project.
//
// This file contains the go-update-direct-dependencies and go-update-all-dependencies command implementations.
package cicd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// goUpdateDeps checks for outdated Go dependencies and fails if any are found.
// It supports two modes: direct dependencies only (go-update-direct-dependencies) or all dependencies (go-update-all-dependencies).
// This command uses caching to avoid repeated expensive checks and exits with code 1 if outdated dependencies are found.
func goUpdateDeps(logger *LogUtil, mode cryptoutilMagic.DepCheckMode) {
	modeName := cryptoutilMagic.ModeNameDirect
	if mode == cryptoutilMagic.DepCheckAll {
		modeName = cryptoutilMagic.ModeNameAll
	}

	cacheFile := cryptoutilMagic.DepCacheFileName

	// Get go.mod and go.sum file stats - needed for cache timestamp comparison
	goModStat, err := os.Stat("go.mod")
	if err != nil {
		logger.Log(fmt.Sprintf("Error reading go.mod: %v", err))
		os.Exit(1)
	}

	goSumStat, err := os.Stat("go.sum")
	if err != nil {
		logger.Log(fmt.Sprintf("Error reading go.sum: %v", err))
		os.Exit(1)
	}

	// Check cache first
	cacheUsed, cacheState := checkAndUseDepCache(cacheFile, modeName, goModStat, goSumStat, logger)
	if cacheUsed {
		return
	}

	// Log the cache miss reason
	logger.Log(fmt.Sprintf("Cache miss: %s", cacheState))

	// Cache miss or expired, perform actual check
	logger.Log("Performing fresh dependency check")

	// Run go list -u -m all to check for outdated dependencies
	cmd := exec.Command("go", "list", "-u", "-m", "all")

	output, err := cmd.Output()
	if err != nil {
		logger.Log(fmt.Sprintf("Error checking dependencies: %v", err))
		os.Exit(1)
	}

	// Get direct dependencies for the check
	var directDeps map[string]bool

	if mode == cryptoutilMagic.DepCheckDirect {
		// Read go.mod file to get direct dependencies (this serves as exists check)
		goModContent, err := os.ReadFile("go.mod")
		if err != nil {
			logger.Log(fmt.Sprintf("Error reading go.mod: %v", err))
			os.Exit(1)
		}

		directDeps, err = getDirectDependencies(goModContent)
		if err != nil {
			logger.Log(fmt.Sprintf("Error parsing direct dependencies: %v", err))
			os.Exit(1)
		}
	}

	// Use the extracted function for the core logic
	outdated, err := checkDependencyUpdates(mode, string(output), directDeps)
	if err != nil {
		logger.Log(fmt.Sprintf("Error checking dependency updates: %v", err))
		os.Exit(1)
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
		os.Exit(1) // Fail to block push
	}

	fmt.Fprintf(os.Stderr, "All %s Go dependencies are up to date.\n", modeName)

	logger.Log("goUpdateDeps completed")
}

// checkDependencyUpdates analyzes dependency update information and returns outdated dependencies.
// It takes the mode, file stats, go list output, and direct dependencies map as inputs to enable testing with mock data.
// Returns a slice of outdated dependency strings and an error if the check fails.
func checkDependencyUpdates(mode cryptoutilMagic.DepCheckMode, goListOutput string, directDeps map[string]bool) ([]string, error) {
	lines := strings.Split(strings.TrimSpace(goListOutput), "\n")
	allOutdated := []string{}

	// Check for lines containing [v...] indicating available updates
	for _, line := range lines {
		if strings.Contains(line, "[v") && strings.Contains(line, "]") {
			allOutdated = append(allOutdated, line)
		}
	}

	var outdated []string

	if mode == cryptoutilMagic.DepCheckDirect {
		// For direct mode, only check dependencies that are explicitly listed in go.mod
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

	if err := os.WriteFile(cacheFile, content, cryptoutilMagic.CacheFilePermissions); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// getDirectDependencies parses go.mod content and returns a map of direct dependencies.
func getDirectDependencies(goModContent []byte) (map[string]bool, error) {
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

// checkAndUseDepCache checks if valid cached dependency results exist and uses them if available.
// Returns true if cache was used (and function should return), false if cache miss occurred.
// Also returns a descriptive string about the cache state for logging.
func checkAndUseDepCache(cacheFile, modeName string, goModStat, goSumStat os.FileInfo, logger *LogUtil) (bool, string) {
	cache, err := loadDepCache(cacheFile, modeName)
	if err != nil {
		// Determine the specific reason for cache miss
		if errors.Is(err, os.ErrNotExist) {
			return false, "cache_not_exists"
		}
		// Check if it's a mode mismatch (our custom error)
		if strings.Contains(err.Error(), "cache mode mismatch") {
			return false, "cache_mode_mismatch"
		}
		// Other errors (JSON parsing, etc.)
		return false, "cache_invalid"
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
		return false, fmt.Sprintf("cache_expired_files (%s)", invalidReason)
	}

	if time.Since(cache.LastCheck) >= time.Hour {
		return false, fmt.Sprintf("cache_expired_time (age: %.1fs)", time.Since(cache.LastCheck).Seconds())
	}

	// Cache is valid and fresh - use it
	logger.Log(fmt.Sprintf("Using cached dependency check results (age: %.1fs)", time.Since(cache.LastCheck).Seconds()))

	if len(cache.OutdatedDeps) > 0 {
		logger.Log(fmt.Sprintf("Found outdated Go dependencies (cached, checking %s)", modeName))

		for _, dep := range cache.OutdatedDeps {
			fmt.Fprintln(os.Stderr, dep)
		}

		fmt.Fprintln(os.Stderr, "\nPlease run 'go get -u ./...' to update dependencies manually.")
		os.Exit(1) // Fail to block push
	}

	logger.Log(fmt.Sprintf("All %s Go dependencies are up to date (cached)", modeName))

	logger.Log("goUpdateDeps completed (cached)")

	return true, "cache_hit"
}
