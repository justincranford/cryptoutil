// Package cicd provides CI/CD quality control checks for the cryptoutil project.
//
// This file contains the go-check-circular-package-dependencies command implementation.
package cicd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	cryptoutilMagic "cryptoutil/internal/common/magic"
	cryptoutilFiles "cryptoutil/internal/common/util/files"
)

type PackageInfo struct {
	ImportPath string   `json:"ImportPath"`
	Imports    []string `json:"Imports"`
}

// goCheckCircularPackageDeps checks for circular dependencies in Go packages.
// It analyzes the dependency graph of all packages in the project and returns an error if circular dependencies are found.
// Uses caching to avoid expensive go list calls when nothing has changed.
func goCheckCircularPackageDeps(logger *LogUtil) error {
	logger.Log("Checking for circular dependencies in Go packages")

	cacheFile := cryptoutilMagic.CircularDepCacheFileName

	// Check if we can use cached results
	goModStat, err := os.Stat("go.mod")
	if err != nil {
		logger.Log(fmt.Sprintf("Error reading go.mod: %v", err))

		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Try to load cache
	if cache, err := loadCircularDepCache(cacheFile); err == nil {
		// Check if cache is still valid
		cacheAge := time.Since(cache.LastCheck)
		isExpired := cacheAge > cryptoutilMagic.CircularDepCacheValidDuration
		goModChanged := cache.GoModModTime.Before(goModStat.ModTime())

		if !isExpired && !goModChanged {
			logger.Log(fmt.Sprintf("Using cached circular dependency check results (age: %.1fs)", cacheAge.Seconds()))

			if cache.HasCircularDeps {
				errMsg := fmt.Sprintf("circular dependencies detected (cached): %s", strings.Join(cache.CircularDeps, ", "))
				logger.Log(fmt.Sprintf("❌ RESULT: %s", errMsg))

				return fmt.Errorf("%s", errMsg)
			}

			fmt.Fprintln(os.Stderr, "✅ RESULT: No circular dependencies found (cached)")
			fmt.Fprintln(os.Stderr, "All internal package dependencies are acyclic.")
			logger.Log("goCheckCircularPackageDeps completed (cached, no circular dependencies)")

			return nil
		}

		if isExpired {
			logger.Log(fmt.Sprintf("Cache expired (age: %.1fs > %.0fs)", cacheAge.Seconds(), cryptoutilMagic.CircularDepCacheValidDuration.Seconds()))
		}

		if goModChanged {
			logger.Log("Cache invalidated: go.mod was modified")
		}
	} else {
		logger.Log(fmt.Sprintf("Cache miss: %v", err))
	}

	// Cache miss or expired, perform actual check
	logger.Log("Running: go list -json ./...")

	cmd := exec.Command("go", "list", "-json", "./...")

	output, err := cmd.Output()
	if err != nil {
		logger.Log(fmt.Sprintf("Error running go list: %v", err))

		return fmt.Errorf("failed to run go list: %w", err)
	}

	// Use the extracted function for the core logic
	circularDepsError := checkCircularDependencies(string(output))

	// Save results to cache
	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    goModStat.ModTime(),
		HasCircularDeps: circularDepsError != nil,
		CircularDeps:    []string{},
	}

	if circularDepsError != nil {
		// Store simplified error message in cache
		cache.CircularDeps = []string{circularDepsError.Error()}
	}

	if err := saveCircularDepCache(cacheFile, cache); err != nil {
		logger.Log(fmt.Sprintf("Warning: Failed to save circular dependency cache: %v", err))
	}

	if circularDepsError != nil {
		logger.Log(fmt.Sprintf("❌ RESULT: %v", circularDepsError))
		logger.Log("goCheckCircularPackageDeps completed (circular dependencies found)")

		return circularDepsError
	}

	fmt.Fprintln(os.Stderr, "✅ RESULT: No circular dependencies found")
	fmt.Fprintln(os.Stderr, "All internal package dependencies are acyclic.")

	logger.Log("goCheckCircularPackageDeps completed (no circular dependencies)")

	return nil
}

// loadCircularDepCache loads circular dependency cache from the specified file.
func loadCircularDepCache(cacheFile string) (*cryptoutilMagic.CircularDepCache, error) {
	content, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache cryptoutilMagic.CircularDepCache
	if err := json.Unmarshal(content, &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache JSON: %w", err)
	}

	return &cache, nil
}

// saveCircularDepCache saves circular dependency cache to the specified file.
func saveCircularDepCache(cacheFile string, cache cryptoutilMagic.CircularDepCache) error {
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

// checkCircularDependencies analyzes the JSON output from 'go list -json ./...' for circular dependencies.
// Returns an error if circular dependencies are found, nil if no circular dependencies exist.
func checkCircularDependencies(jsonOutput string) error {
	// Parse JSON output (multiple JSON objects in stream)
	decoder := json.NewDecoder(strings.NewReader(jsonOutput))
	packages := make([]PackageInfo, 0)

	for {
		var pkg PackageInfo
		if err := decoder.Decode(&pkg); err != nil {
			if err == io.EOF {
				break
			}

			return fmt.Errorf("failed to parse package info: %w", err)
		}

		packages = append(packages, pkg)
	}

	if len(packages) == 0 {
		return fmt.Errorf("no packages found")
	}

	// Build dependency graph
	dependencyGraph := make(map[string][]string)

	for _, pkg := range packages {
		dependencyGraph[pkg.ImportPath] = pkg.Imports
	}

	// Find circular dependencies using DFS
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
	for pkg := range dependencyGraph {
		if !visited[pkg] {
			dfs(pkg, []string{pkg})
		}
	}

	if len(circularDeps) == 0 {
		return nil // No circular dependencies found
	}

	// Build error message with details about circular dependencies
	var errorMsg strings.Builder

	errorMsg.WriteString(fmt.Sprintf("Found %d circular dependency chain(s):\n\n", len(circularDeps)))

	for i, cycle := range circularDeps {
		errorMsg.WriteString(fmt.Sprintf("Chain %d (%d packages):\n", i+1, len(cycle)))

		for j, pkg := range cycle {
			prefix := "  "
			if j > 0 {
				prefix = "  → "
			}

			errorMsg.WriteString(fmt.Sprintf("%s%s\n", prefix, pkg))
		}

		errorMsg.WriteString("\n")
	}

	errorMsg.WriteString("Circular dependencies can prevent enabling advanced linters like gomnd.\n")
	errorMsg.WriteString("Consider refactoring to break these cycles.")

	return fmt.Errorf("circular dependencies detected: %s", errorMsg.String())
}
