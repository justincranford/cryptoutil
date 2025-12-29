// Copyright (c) 2025 Justin Cranford
//
//

package check_no_cgo_sqlite

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/cmd/cicd/common"
)

const (
	bannedModule   = "github.com/mattn/go-sqlite3"
	requiredModule = "modernc.org/sqlite"
)

// Check verifies the project uses CGO-free SQLite (modernc.org/sqlite).
// Returns error if banned CGO SQLite module found or required module missing.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	logger.Log("Starting CGO-free SQLite validation")

	// Check go.mod for banned CGO sqlite module (direct dependencies only).
	goModViolations, err := checkGoMod("go.mod")
	if err != nil {
		return fmt.Errorf("failed to check go.mod: %w", err)
	}

	// Check *.go files for banned CGO sqlite import.
	importViolations, err := checkGoFiles()
	if err != nil {
		return fmt.Errorf("failed to check Go files: %w", err)
	}

	// Check that required CGO-free module exists.
	hasRequired, err := checkRequiredModule("go.mod")
	if err != nil {
		return fmt.Errorf("failed to check required module: %w", err)
	}

	if len(goModViolations) > 0 || len(importViolations) > 0 || !hasRequired {
		printViolations(goModViolations, importViolations, hasRequired)

		return fmt.Errorf("CGO validation failed")
	}

	logger.Log("✅ CGO-free validation passed")

	return nil
}

func checkGoMod(path string) ([]string, error) {
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
		if strings.Contains(line, bannedModule) && !strings.Contains(line, "// indirect") {
			violations = append(violations, fmt.Sprintf("go.mod:%d: banned CGO module '%s' (direct dependency)", lineNum, bannedModule))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading go.mod: %w", err)
	}

	return violations, nil
}

func checkRequiredModule(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open go.mod: %w", err)
	}

	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, requiredModule) {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("error reading go.mod: %w", err)
	}

	return false, nil
}

func checkGoFiles() ([]string, error) {
	var violations []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor directories and the check_no_cgo_sqlite package itself.
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == ".git") {
			return filepath.SkipDir
		}

		// Only check .go files.
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			fileViolations, err := checkGoFile(path)
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

func checkGoFile(path string) ([]string, error) {
	// Skip checking the check_no_cgo_sqlite checker itself (would flag its own string literals).
	if strings.Contains(path, "check_no_cgo_sqlite") {
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
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", path, err)
	}

	return violations, nil
}

func printViolations(goModViolations, importViolations []string, hasRequired bool) {
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
		fmt.Fprintf(os.Stderr, "  - '%s' not found in go.mod\n", requiredModule)
		fmt.Fprintf(os.Stderr, "  - Run: go get %s\n", requiredModule)
		fmt.Fprintln(os.Stderr)
	}

	fmt.Fprintln(os.Stderr, "Fix:")
	fmt.Fprintf(os.Stderr, "  1. Remove %s from go.mod and code\n", bannedModule)
	fmt.Fprintf(os.Stderr, "  2. Use %s (CGO-free) instead\n", requiredModule)
	fmt.Fprintln(os.Stderr, "  3. Run: go mod tidy")
}
