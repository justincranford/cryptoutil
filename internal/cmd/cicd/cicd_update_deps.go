// Package cicd provides CI/CD quality control checks for the cryptoutil project.
//
// This file contains the go-update-direct-dependencies and go-update-all-dependencies command implementations.
package cicd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

type DepCheckMode int

const (
	DepCheckDirect DepCheckMode = iota // Check only direct dependencies
	DepCheckAll                        // Check all dependencies (direct + transitive)
)

type DepCache struct {
	LastCheck    time.Time `json:"last_check"`
	GoModModTime time.Time `json:"go_mod_mod_time"`
	GoSumModTime time.Time `json:"go_sum_mod_time"`
	OutdatedDeps []string  `json:"outdated_deps"`
	Mode         string    `json:"mode"`
}

// goUpdateDeps checks for outdated Go dependencies and fails if any are found.
// It supports two modes: direct dependencies only (go-update-direct-dependencies) or all dependencies (go-update-all-dependencies).
// This command uses caching to avoid repeated expensive checks and exits with code 1 if outdated dependencies are found.
func goUpdateDeps(mode DepCheckMode) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[PERF] goUpdateDeps started at %s (mode=%v)\n", start.Format(time.RFC3339Nano), mode)

	modeName := cryptoutilMagic.ModeNameDirect
	if mode == DepCheckAll {
		modeName = cryptoutilMagic.ModeNameAll
	}

	cacheFile := ".cicd-dep-cache.json"

	// Check cache first
	if cache, err := loadDepCache(cacheFile, modeName); err == nil {
		// Check if go.mod or go.sum have changed since cache was created
		goModStat, err := os.Stat("go.mod")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not stat go.mod: %v\n", err)

			goModStat = nil
		}

		goSumStat, err := os.Stat("go.sum")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not stat go.sum: %v\n", err)

			goSumStat = nil
		}

		cacheValid := true
		if goModStat != nil && !goModStat.ModTime().Equal(cache.GoModModTime) {
			cacheValid = false
		}

		if goSumStat != nil && !goSumStat.ModTime().Equal(cache.GoSumModTime) {
			cacheValid = false
		}

		if cacheValid && time.Since(cache.LastCheck) < time.Hour {
			fmt.Fprintf(os.Stderr, "Using cached dependency check results (age: %.1fs)\n", time.Since(cache.LastCheck).Seconds())

			if len(cache.OutdatedDeps) > 0 {
				fmt.Fprintf(os.Stderr, "Found outdated Go dependencies (cached, checking %s):\n", modeName)

				for _, dep := range cache.OutdatedDeps {
					fmt.Fprintln(os.Stderr, dep)
				}

				fmt.Fprintln(os.Stderr, "\nPlease run 'go get -u ./...' to update dependencies manually.")
				os.Exit(1) // Fail to block push
			}

			fmt.Fprintf(os.Stderr, "All %s Go dependencies are up to date (cached).\n", modeName)

			end := time.Now()
			fmt.Fprintf(os.Stderr, "[PERF] goUpdateDeps: duration=%v start=%s end=%s mode=%s outdated=%d (cached)\n",
				end.Sub(start), start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano), modeName, len(cache.OutdatedDeps))

			return
		}
	}

	// Cache miss or expired, perform actual check
	fmt.Fprintf(os.Stderr, "Performing fresh dependency check...\n")

	// Get file stats for the extracted function
	goModStat, err := os.Stat("go.mod")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading go.mod: %v\n", err)
		os.Exit(1)
	}

	goSumStat, err := os.Stat("go.sum")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading go.sum: %v\n", err)
		os.Exit(1)
	}

	// Run go list -u -m all to check for outdated dependencies
	cmd := exec.Command("go", "list", "-u", "-m", "all")

	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking dependencies: %v\n", err)
		os.Exit(1)
	}

	// Get direct dependencies for the check
	var directDeps map[string]bool
	if mode == DepCheckDirect {
		directDeps, err = getDirectDependencies()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading direct dependencies: %v\n", err)
			os.Exit(1)
		}
	}

	// Use the extracted function for the core logic
	outdated, err := checkDependencyUpdates(mode, goModStat, goSumStat, string(output), directDeps)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking dependency updates: %v\n", err)
		os.Exit(1)
	}

	// Save results to cache
	cache := DepCache{
		LastCheck:    time.Now(),
		GoModModTime: goModStat.ModTime(),
		GoSumModTime: goSumStat.ModTime(),
		OutdatedDeps: outdated,
		Mode:         modeName,
	}
	if err := saveDepCache(cacheFile, cache); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to save dependency cache: %v\n", err)
	}

	if len(outdated) > 0 {
		fmt.Fprintf(os.Stderr, "Found outdated Go dependencies (checking %s):\n", modeName)

		for _, dep := range outdated {
			fmt.Fprintln(os.Stderr, dep)
		}

		fmt.Fprintln(os.Stderr, "\nPlease run 'go get -u ./...' to update dependencies manually.")
		os.Exit(1) // Fail to block push
	}

	fmt.Fprintf(os.Stderr, "All %s Go dependencies are up to date.\n", modeName)

	end := time.Now()
	fmt.Fprintf(os.Stderr, "[PERF] goUpdateDeps: duration=%v start=%s end=%s mode=%s outdated=%d\n",
		end.Sub(start), start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano), modeName, len(outdated))
}

// checkDependencyUpdates analyzes dependency update information and returns outdated dependencies.
// It takes the mode, file stats, go list output, and direct dependencies map as inputs to enable testing with mock data.
// Returns a slice of outdated dependency strings and an error if the check fails.
func checkDependencyUpdates(mode DepCheckMode, goModStat, goSumStat os.FileInfo, goListOutput string, directDeps map[string]bool) ([]string, error) {
	lines := strings.Split(strings.TrimSpace(goListOutput), "\n")
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

	if err := os.WriteFile(cacheFile, content, cryptoutilMagic.CacheFilePermissions); err != nil {
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
