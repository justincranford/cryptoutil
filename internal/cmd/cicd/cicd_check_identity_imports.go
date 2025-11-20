// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cryptoutil/internal/cmd/cicd/common"
	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// IdentityImportsCache represents cached identity imports check results.
type IdentityImportsCache struct {
	LastCheck           time.Time `json:"last_check"`
	GoModModTime        time.Time `json:"go_mod_mod_time"`
	IdentityModTime     time.Time `json:"identity_mod_time"`
	HasForbiddenImports bool      `json:"has_forbidden_imports"`
	ForbiddenImports    []string  `json:"forbidden_imports"` // Format: "file:line:package"
}

// BlockedPackage represents a package that identity module cannot import.
type BlockedPackage struct {
	Path   string
	Reason string
}

// getBlockedPackages returns the list of packages that identity module should not import.
func getBlockedPackages() []BlockedPackage {
	return []BlockedPackage{
		{Path: "cryptoutil/internal/server", Reason: "KMS server domain"},
		{Path: "cryptoutil/internal/client", Reason: "KMS client"},
		{Path: "cryptoutil/api", Reason: "OpenAPI generated code"},
		{Path: "cryptoutil/cmd/cryptoutil", Reason: "CLI command"},
		{Path: "cryptoutil/internal/common/crypto", Reason: "use stdlib instead"},
		{Path: "cryptoutil/internal/common/pool", Reason: "KMS infrastructure"},
		{Path: "cryptoutil/internal/common/container", Reason: "KMS infrastructure"},
		{Path: "cryptoutil/internal/common/telemetry", Reason: "KMS infrastructure"},
		{Path: "cryptoutil/internal/common/util", Reason: "KMS infrastructure"},
	}
}

// goCheckIdentityImports checks that identity module doesn't import from forbidden packages.
// This enforces domain isolation between identity and KMS server domains.
func goCheckIdentityImports(logger *common.Logger) error {
	logger.Log("Checking identity module imports for domain isolation violations...")

	cacheFile := ".cicd/identity-imports-cache.json"

	// Check if we can use cached results
	goModStat, err := os.Stat("go.mod")
	if err != nil {
		logger.Log(fmt.Sprintf("Error reading go.mod: %v", err))

		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Get latest modification time of identity module files
	identityModTime, err := getLatestModTime("internal/identity")
	if err != nil {
		logger.Log(fmt.Sprintf("Error checking identity module files: %v", err))

		return fmt.Errorf("failed to check identity module: %w", err)
	}

	// Try to load cache
	if cache, err := loadIdentityImportsCache(cacheFile); err == nil {
		// Check if cache is still valid
		cacheAge := time.Since(cache.LastCheck)
		isExpired := cacheAge > cryptoutilMagic.CacheDuration
		goModChanged := cache.GoModModTime.Before(goModStat.ModTime())
		identityChanged := cache.IdentityModTime.Before(identityModTime)

		if !isExpired && !goModChanged && !identityChanged {
			logger.Log(fmt.Sprintf("Using cached identity imports check results (age: %.1fs)", cacheAge.Seconds()))

			if cache.HasForbiddenImports {
				errMsg := fmt.Sprintf("forbidden imports detected (cached): %s", strings.Join(cache.ForbiddenImports, ", "))
				logger.Log(fmt.Sprintf("❌ RESULT: %s", errMsg))

				return fmt.Errorf("%s", errMsg)
			}

			fmt.Fprintln(os.Stderr, "✅ RESULT: No forbidden imports found (cached)")
			fmt.Fprintln(os.Stderr, "Identity module maintains proper domain isolation.")
			logger.Log("goCheckIdentityImports completed (cached, no violations)")

			return nil
		}

		if isExpired {
			logger.Log(fmt.Sprintf("Cache expired (age: %.1fs > 300s)", cacheAge.Seconds()))
		}

		if goModChanged {
			logger.Log("Cache invalidated: go.mod was modified")
		}

		if identityChanged {
			logger.Log("Cache invalidated: identity module files were modified")
		}
	} else {
		logger.Log(fmt.Sprintf("Cache miss: %v", err))
	}

	// Cache miss or expired, perform actual check
	logger.Log("Scanning internal/identity/**/*.go files for forbidden imports")

	violations, err := checkIdentityImports(logger)
	if err != nil {
		logger.Log(fmt.Sprintf("Error checking imports: %v", err))

		return fmt.Errorf("failed to check imports: %w", err)
	}

	// Save results to cache
	cache := IdentityImportsCache{
		LastCheck:           time.Now().UTC(),
		GoModModTime:        goModStat.ModTime(),
		IdentityModTime:     identityModTime,
		HasForbiddenImports: len(violations) > 0,
		ForbiddenImports:    violations,
	}

	if err := saveIdentityImportsCache(cacheFile, cache); err != nil {
		logger.Log(fmt.Sprintf("Warning: failed to save cache: %v", err))
	}

	if len(violations) > 0 {
		errMsg := fmt.Sprintf("forbidden imports detected: %d violations", len(violations))
		logger.Log(fmt.Sprintf("❌ RESULT: %s", errMsg))

		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "  %s\n", v)
		}

		return fmt.Errorf("%s", errMsg)
	}

	fmt.Fprintln(os.Stderr, "✅ RESULT: No forbidden imports found")
	fmt.Fprintln(os.Stderr, "Identity module maintains proper domain isolation.")
	logger.Log("goCheckIdentityImports completed (no violations)")

	return nil
}

// checkIdentityImports scans identity module files for forbidden imports.
func checkIdentityImports(logger *common.Logger) ([]string, error) {
	violations := []string{}
	blockedPkgs := getBlockedPackages()

	// Build map for fast lookup
	blocked := make(map[string]string)
	for _, pkg := range blockedPkgs {
		blocked[pkg.Path] = pkg.Reason
	}

	// Walk identity directory
	err := filepath.Walk("internal/identity", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err //nolint:wrapcheck // Cache helper - wrapping adds no value
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Parse the file
		fset := token.NewFileSet()

		node, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			logger.Log(fmt.Sprintf("Warning: failed to parse %s: %v", path, parseErr))

			return nil // Continue checking other files
		}

		// Check each import
		for _, imp := range node.Imports {
			// Remove quotes from import path
			importPath := strings.Trim(imp.Path.Value, `"`)

			// Check if this import is forbidden
			if reason, isForbidden := blocked[importPath]; isForbidden {
				pos := fset.Position(imp.Pos())
				violation := fmt.Sprintf("%s:%d: forbidden import %q (%s)", path, pos.Line, importPath, reason)
				violations = append(violations, violation)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err //nolint:wrapcheck // Cache helper - wrapping adds no value
	}

	return violations, nil
}

// getLatestModTime returns the latest modification time of files in a directory.
func getLatestModTime(dir string) (time.Time, error) {
	var latest time.Time

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			if info.ModTime().After(latest) {
				latest = info.ModTime()
			}
		}

		return nil
	})

	return latest, err //nolint:wrapcheck // Cache helper - wrapping adds no value
}

// loadIdentityImportsCache loads cached results from file.
func loadIdentityImportsCache(filename string) (IdentityImportsCache, error) {
	var cache IdentityImportsCache

	data, err := os.ReadFile(filename)
	if err != nil {
		return cache, err //nolint:wrapcheck // Cache helper - wrapping adds no value
	}

	err = json.Unmarshal(data, &cache)

	return cache, err //nolint:wrapcheck // Cache helper - wrapping adds no value
}

// saveIdentityImportsCache saves check results to cache file.
func saveIdentityImportsCache(filename string, cache IdentityImportsCache) error {
	// Ensure cache directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, cryptoutilMagic.CICDOutputDirPermissions); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err) //nolint:wrapcheck // Cache helper - wrapping adds no value
	}

	if err := os.WriteFile(filename, data, cryptoutilMagic.CacheFilePermissions); err != nil { //nolint:wrapcheck // Cache helper - wrapping adds no value
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}
