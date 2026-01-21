// Copyright (c) 2025 Justin Cranford

// Package lint_go provides Go linting utilities for the CICD pipeline.
package lint_go

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilFiles "cryptoutil/internal/shared/util/files"
)

// PackageInfo represents Go package metadata from 'go list -json'.
type PackageInfo struct {
	ImportPath string   `json:"ImportPath"`
	Imports    []string `json:"Imports"`
}

// checkCircularDeps checks for circular dependencies in Go packages.
// It analyzes the dependency graph of all packages in the project and returns an error if circular dependencies are found.
// Uses caching to avoid expensive go list calls when nothing has changed.
func checkCircularDeps(logger *cryptoutilCmdCicdCommon.Logger) error {
	logger.Log("Checking for circular package dependencies...")

	cacheFile := cryptoutilMagic.CircularDepCacheFileName

	// Check if we can use cached results.
	goModStat, err := os.Stat("go.mod")
	if err != nil {
		logger.Log(fmt.Sprintf("Error reading go.mod: %v", err))

		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Try to load cache.
	if cache, err := loadCircularDepCache(cacheFile); err == nil {
		// Check if cache is still valid.
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
			logger.Log("Circular dependency check completed (cached, no circular dependencies)")

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

	// Cache miss or expired, perform actual check.
	logger.Log("Running: go list -json ./...")

	cmd := exec.CommandContext(context.Background(), "go", "list", "-json", "./...")

	output, err := cmd.Output()
	if err != nil {
		logger.Log(fmt.Sprintf("Error running go list: %v", err))

		return fmt.Errorf("failed to run go list: %w", err)
	}

	// Use the extracted function for the core logic.
	circularDepsError := CheckDependencies(string(output))

	// Save results to cache.
	cache := cryptoutilMagic.CircularDepCache{
		LastCheck:       time.Now().UTC(),
		GoModModTime:    goModStat.ModTime(),
		HasCircularDeps: circularDepsError != nil,
		CircularDeps:    []string{},
	}

	if circularDepsError != nil {
		// Store simplified error message in cache.
		cache.CircularDeps = []string{circularDepsError.Error()}
	}

	if err := saveCircularDepCache(cacheFile, cache); err != nil {
		logger.Log(fmt.Sprintf("Warning: Failed to save circular dependency cache: %v", err))
	}

	if circularDepsError != nil {
		logger.Log(fmt.Sprintf("❌ RESULT: %v", circularDepsError))
		logger.Log("Circular dependency check completed (circular dependencies found)")

		return circularDepsError
	}

	fmt.Fprintln(os.Stderr, "✅ RESULT: No circular dependencies found")
	fmt.Fprintln(os.Stderr, "All internal package dependencies are acyclic.")

	logger.Log("Circular dependency check completed (no circular dependencies)")

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

	if err := cryptoutilFiles.WriteFile(cacheFile, string(content), cryptoutilMagic.CacheFilePermissions); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// CheckDependencies analyzes go list JSON output for circular dependencies.
// This function is exported for testing purposes.
func CheckDependencies(goListOutput string) error {
	// Parse JSON stream.
	packages := make(map[string][]string)

	decoder := json.NewDecoder(strings.NewReader(goListOutput))

	for decoder.More() {
		var pkg PackageInfo
		if err := decoder.Decode(&pkg); err != nil {
			return fmt.Errorf("failed to decode package info: %w", err)
		}

		packages[pkg.ImportPath] = pkg.Imports
	}

	// Build adjacency list for internal packages only.
	internalPrefix := getModulePath(packages)

	graph := make(map[string][]string)

	for pkgPath, imports := range packages {
		if !strings.HasPrefix(pkgPath, internalPrefix) {
			continue
		}

		var internalImports []string

		for _, imp := range imports {
			if strings.HasPrefix(imp, internalPrefix) {
				internalImports = append(internalImports, imp)
			}
		}

		graph[pkgPath] = internalImports
	}

	// Detect cycles using DFS.
	visited := make(map[string]int) // 0=unvisited, 1=visiting, 2=visited

	var cycle []string

	var dfs func(node string) bool

	dfs = func(node string) bool {
		if visited[node] == 2 {
			return false
		}

		if visited[node] == 1 {
			// Found cycle.
			cycle = append(cycle, node)

			return true
		}

		visited[node] = 1

		for _, neighbor := range graph[node] {
			if dfs(neighbor) {
				if len(cycle) == 0 || cycle[0] != cycle[len(cycle)-1] {
					cycle = append(cycle, node)
				}

				return true
			}
		}

		visited[node] = 2

		return false
	}

	for node := range graph {
		if visited[node] == 0 {
			if dfs(node) {
				return fmt.Errorf("circular dependency detected: %s", strings.Join(cycle, " -> "))
			}
		}
	}

	return nil
}

// getModulePath extracts the module path from the packages.
func getModulePath(packages map[string][]string) string {
	for pkgPath := range packages {
		parts := strings.Split(pkgPath, "/")
		if len(parts) > 0 {
			// Return the root module path.
			return parts[0]
		}
	}

	return ""
}
