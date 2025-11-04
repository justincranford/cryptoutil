// Package cicd provides CI/CD quality control checks for the cryptoutil project.
//
// This file contains the go-update-direct-dependencies and go-update-all-dependencies command implementations.
package cicd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// goUpdateDeps checks for outdated Go dependencies and fails if any are found.
// It supports two modes: direct dependencies only (go-update-direct-dependencies) or all dependencies (go-update-all-dependencies).
// This command uses caching to avoid repeated expensive checks and exits with code 1 if outdated dependencies are found.
func goUpdateDeps(mode DepCheckMode) {
	start := time.Now()
	fmt.Fprintf(os.Stderr, "[PERF] goUpdateDeps started at %s (mode=%v)\n", start.Format(time.RFC3339Nano), mode)

	modeName := modeNameDirect
	if mode == DepCheckAll {
		modeName = modeNameAll
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

	// Save results to cache
	goModModTime := time.Time{}
	goSumModTime := time.Time{}

	if goModStat, err := os.Stat("go.mod"); err == nil {
		goModModTime = goModStat.ModTime()
	}

	if goSumStat, err := os.Stat("go.sum"); err == nil {
		goSumModTime = goSumStat.ModTime()
	}

	cache := DepCache{
		LastCheck:    time.Now(),
		GoModModTime: goModModTime,
		GoSumModTime: goSumModTime,
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
