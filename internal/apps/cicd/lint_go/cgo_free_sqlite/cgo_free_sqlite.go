// Copyright (c) 2025 Justin Cranford

// Package cgo_free_sqlite verifies that source code uses the CGO-free SQLite driver.
package cgo_free_sqlite

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

const (
	bannedCGOModule          = "github.com/mattn/go-sqlite3"
	bannedCGOMigrateModule   = "github.com/golang-migrate/migrate/v4/database/sqlite3"
	requiredCGOModule        = "modernc.org/sqlite"
	requiredCGOMigrateModule = "github.com/golang-migrate/migrate/v4/database/sqlite"
	excludeDirVendor         = "vendor"
	excludeDirGit            = ".git"
)

// Check verifies the project uses CGO-free SQLite (modernc.org/sqlite).
// Returns error if banned CGO SQLite module found or required module missing.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	logger.Log("Checking CGO-free SQLite compliance")

	// Check go.mod for banned CGO sqlite module (direct dependencies only).
	goModViolations, err := CheckGoModForCGO("go.mod")
	if err != nil {
		return fmt.Errorf("failed to check go.mod: %w", err)
	}

	// Check *.go files for banned CGO sqlite import.
	importViolations, err := CheckGoFilesForCGO()
	if err != nil {
		return fmt.Errorf("failed to check Go files: %w", err)
	}

	// Check that required CGO-free module exists.
	hasRequired, err := CheckRequiredCGOModule("go.mod")
	if err != nil {
		return fmt.Errorf("failed to check required module: %w", err)
	}

	if len(goModViolations) > 0 || len(importViolations) > 0 || !hasRequired {
		PrintCGOViolations(goModViolations, importViolations, hasRequired)

		return fmt.Errorf("CGO validation failed")
	}

	logger.Log("✅ CGO-free validation passed")

	return nil
}

func CheckGoModForCGO(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open go.mod: %w", err)
	}

	defer func() { _ = file.Close() }()

	var violations []string

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		// Only flag DIRECT dependencies (lines without "// indirect").
		// Indirect dependencies are acceptable since we're not importing them.
		if strings.Contains(line, bannedCGOModule) && !strings.Contains(line, "// indirect") {
			violations = append(violations, fmt.Sprintf("go.mod:%d: banned CGO module '%s' (direct dependency)", lineNum, bannedCGOModule))
		}

		if strings.Contains(line, bannedCGOMigrateModule) && !strings.Contains(line, "// indirect") {
			violations = append(violations, fmt.Sprintf("go.mod:%d: banned CGO migrate module '%s' (direct dependency)", lineNum, bannedCGOMigrateModule))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading go.mod: %w", err)
	}

	return violations, nil
}

func CheckRequiredCGOModule(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open go.mod: %w", err)
	}

	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, requiredCGOModule) {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading go.mod: %w", err)
	}

	return false, nil
}

func CheckGoFilesForCGO() ([]string, error) {
	var violations []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor directories and linting package itself.
		if info.IsDir() && (info.Name() == excludeDirVendor || info.Name() == excludeDirGit) {
			return filepath.SkipDir
		}

		// Only check .go files.
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			fileViolations, err := CheckGoFileForCGO(path)
			if err != nil {
				return fmt.Errorf("error checking %s: %w", path, err)
			}

			violations = append(violations, fileViolations...)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking directory tree: %w", err)
	}

	return violations, nil
}

func CheckGoFileForCGO(path string) ([]string, error) {
	// Skip checking the linting package itself (would flag its own string literals).
	if strings.Contains(path, "lint_go") || strings.Contains(path, "check_no_cgo_sqlite") {
		return nil, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}

	defer func() { _ = file.Close() }()

	var violations []string

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		// Check for actual import statements, not just string literals.
		if (strings.Contains(line, "import") || strings.HasPrefix(line, "_")) &&
			strings.Contains(line, `"github.com/mattn/go-sqlite3"`) {
			violations = append(violations, fmt.Sprintf("%s:%d: banned CGO import detected", path, lineNum))
		}

		if (strings.Contains(line, "import") || strings.HasPrefix(line, "_")) &&
			strings.Contains(line, `"github.com/golang-migrate/migrate/v4/database/sqlite3"`) {
			violations = append(violations, fmt.Sprintf("%s:%d: banned CGO migrate import detected", path, lineNum))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", path, err)
	}

	return violations, nil
}

func PrintCGOViolations(goModViolations, importViolations []string, hasRequired bool) {
	fmt.Fprintln(os.Stderr, "❌ CGO validation failed:")
	fmt.Fprintln(os.Stderr)

	if len(goModViolations) > 0 {
		fmt.Fprintln(os.Stderr, "go.mod violations:")

		for _, v := range goModViolations {
			fmt.Fprintf(os.Stderr, "  - %s\n", v)
		}

		fmt.Fprintln(os.Stderr)
	}

	if len(importViolations) > 0 {
		fmt.Fprintln(os.Stderr, "Import violations:")

		for _, v := range importViolations {
			fmt.Fprintf(os.Stderr, "  - %s\n", v)
		}

		fmt.Fprintln(os.Stderr)
	}

	if !hasRequired {
		fmt.Fprintf(os.Stderr, "Required module missing:\n")
		fmt.Fprintf(os.Stderr, "  - '%s' not found in go.mod\n", requiredCGOModule)
		fmt.Fprintf(os.Stderr, "  - Run: go get %s\n", requiredCGOModule)
		fmt.Fprintln(os.Stderr)
	}

	fmt.Fprintln(os.Stderr, "Fix:")
	fmt.Fprintf(os.Stderr, "  1. Remove %s from go.mod and code\n", bannedCGOModule)
	fmt.Fprintf(os.Stderr, "  2. Remove %s from go.mod and code\n", bannedCGOMigrateModule)
	fmt.Fprintf(os.Stderr, "  3. Use %s (CGO-free) instead\n", requiredCGOModule)
	fmt.Fprintf(os.Stderr, "  4. Use %s (CGO-free) instead\n", requiredCGOMigrateModule)
	fmt.Fprintln(os.Stderr, "  5. Run: go mod tidy")
}
