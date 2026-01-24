// Copyright (c) 2025 Justin Cranford

package lint_go

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
)

// Directory exclusions for lint checks.
const (
	excludeDirVendor = "vendor"
	excludeDirGit    = ".git"
)

// LinterFunc is a function type for individual Go linters.
// Each linter receives a logger, returning an error if issues are found.
type LinterFunc func(logger *cryptoutilCmdCicdCommon.Logger) error

// registeredLinters holds all linters to run as part of lint-go.
var registeredLinters = []struct {
	name   string
	linter LinterFunc
}{
	{"circular-deps", checkCircularDeps},
	{"cgo-free-sqlite", checkCGOFreeSQLite},
	{"non-fips-algorithms", checkNonFIPS},
	{"no-unaliased-cryptoutil-imports", checkNoUnaliasedCryptoutilImports},
}

// Lint runs all registered Go linters.
// Returns an error if any linter finds issues.
func Lint(logger *cryptoutilCmdCicdCommon.Logger) error {
	logger.Log("Running Go linters...")

	var errors []error

	for _, l := range registeredLinters {
		logger.Log(fmt.Sprintf("Running linter: %s", l.name))

		if err := l.linter(logger); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", l.name, err))
		}
	}

	if len(errors) > 0 {
		logger.Log(fmt.Sprintf("lint-go completed with %d errors", len(errors)))

		return fmt.Errorf("lint-go failed with %d errors", len(errors))
	}

	logger.Log("lint-go completed successfully")

	return nil
}

const (
	bannedCGOModule          = "github.com/mattn/go-sqlite3"
	bannedCGOMigrateModule   = "github.com/golang-migrate/migrate/v4/database/sqlite3"
	requiredCGOModule        = "modernc.org/sqlite"
	requiredCGOMigrateModule = "github.com/golang-migrate/migrate/v4/database/sqlite"
)

// checkCGOFreeSQLite verifies the project uses CGO-free SQLite (modernc.org/sqlite).
// Returns error if banned CGO SQLite module found or required module missing.
func checkCGOFreeSQLite(logger *cryptoutilCmdCicdCommon.Logger) error {
	logger.Log("Checking CGO-free SQLite compliance")

	// Check go.mod for banned CGO sqlite module (direct dependencies only).
	goModViolations, err := checkGoModForCGO("go.mod")
	if err != nil {
		return fmt.Errorf("failed to check go.mod: %w", err)
	}

	// Check *.go files for banned CGO sqlite import.
	importViolations, err := checkGoFilesForCGO()
	if err != nil {
		return fmt.Errorf("failed to check Go files: %w", err)
	}

	// Check that required CGO-free module exists.
	hasRequired, err := checkRequiredCGOModule("go.mod")
	if err != nil {
		return fmt.Errorf("failed to check required module: %w", err)
	}

	if len(goModViolations) > 0 || len(importViolations) > 0 || !hasRequired {
		printCGOViolations(goModViolations, importViolations, hasRequired)

		return fmt.Errorf("CGO validation failed")
	}

	logger.Log("✅ CGO-free validation passed")

	return nil
}

func checkGoModForCGO(path string) ([]string, error) {
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

func checkRequiredCGOModule(path string) (bool, error) {
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

func checkGoFilesForCGO() ([]string, error) {
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
			fileViolations, err := checkGoFileForCGO(path)
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

func checkGoFileForCGO(path string) ([]string, error) {
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

func printCGOViolations(goModViolations, importViolations []string, hasRequired bool) {
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

// checkNoUnaliasedCryptoutilImports validates all cryptoutil imports use aliases from .golangci.yml.
// Returns error if any unaliased cryptoutil imports are found.
func checkNoUnaliasedCryptoutilImports(logger *cryptoutilCmdCicdCommon.Logger) error {
	logger.Log("Checking for unaliased cryptoutil imports")

	violations, err := findUnaliasedCryptoutilImports()
	if err != nil {
		return fmt.Errorf("failed to check cryptoutil imports: %w", err)
	}

	if len(violations) > 0 {
		printCryptoutilImportViolations(violations)

		return fmt.Errorf("found %d unaliased cryptoutil imports", len(violations))
	}

	logger.Log("✅ All cryptoutil imports use aliases")

	return nil
}

func findUnaliasedCryptoutilImports() ([]string, error) {
	var violations []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor directories.
		if info.IsDir() && (info.Name() == excludeDirVendor || info.Name() == excludeDirGit) {
			return filepath.SkipDir
		}

		// Only check .go files.
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			fileViolations, err := checkGoFileForUnaliasedCryptoutilImports(path)
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

func checkGoFileForUnaliasedCryptoutilImports(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %w", path, err)
	}

	defer func() { _ = file.Close() }()

	var violations []string

	scanner := bufio.NewScanner(file)
	lineNum := 0
	inImportBlock := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Track import block.
		if trimmed == "import (" {
			inImportBlock = true

			continue
		}

		if trimmed == ")" && inImportBlock {
			inImportBlock = false

			continue
		}

		// Check for unaliased cryptoutil imports.
		// Pattern: starts with optional whitespace, then "cryptoutil/".
		// If it has an alias, it would be: alias "cryptoutil/..."
		// If it doesn't have an alias, it would be: "cryptoutil/..."
		if inImportBlock || strings.HasPrefix(trimmed, "import ") {
			// Extract the import line.
			importLine := trimmed

			if strings.HasPrefix(trimmed, "import ") {
				importLine = strings.TrimPrefix(trimmed, "import ")
			}

			importLine = strings.TrimSpace(importLine)

			// Check if it starts with "cryptoutil/" (unaliased).
			if strings.HasPrefix(importLine, `"cryptoutil/`) {
				violations = append(violations, fmt.Sprintf("%s:%d: unaliased cryptoutil import detected (must use importas alias)", path, lineNum))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", path, err)
	}

	return violations, nil
}

func printCryptoutilImportViolations(violations []string) {
	fmt.Fprintln(os.Stderr, "❌ Unaliased cryptoutil imports found:")
	fmt.Fprintln(os.Stderr)

	for _, v := range violations {
		fmt.Fprintf(os.Stderr, "  - %s\n", v)
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Fix:")
	fmt.Fprintln(os.Stderr, "  1. All cryptoutil imports MUST use aliases defined in .golangci.yml")
	fmt.Fprintln(os.Stderr, "  2. Run: golangci-lint run --fix")
	fmt.Fprintln(os.Stderr, "  3. If alias is missing, add it to .golangci.yml importas section")
}
