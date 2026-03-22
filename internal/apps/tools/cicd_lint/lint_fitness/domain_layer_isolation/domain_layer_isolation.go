// Copyright (c) 2025 Justin Cranford

// Package domain_layer_isolation verifies that domain/ packages do not import
// server/, client/, or api/ packages. Domain packages contain business logic
// and must not depend on delivery-layer concerns.
package domain_layer_isolation

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Test seams: replaceable in tests to exercise unreachable OS-level error paths.
// See ARCHITECTURE.md Section 10.2.4 (Test Seam Injection Pattern).
var domainIsolationWalkFn = filepath.Walk

// Check verifies domain layer isolation from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir verifies domain layer isolation under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking domain layer isolation...")

	projectRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("failed to resolve root dir: %w", err)
	}

	var violations []string

	walkErr := domainIsolationWalkFn(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if name == cryptoutilSharedMagic.CICDExcludeDirGit || name == cryptoutilSharedMagic.CICDExcludeDirVendor || name == cryptoutilSharedMagic.CICDExcludeDirTestOutput {
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files - they have different import rules.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Only scan files inside a domain/ directory.
		if !isDomainFile(path) {
			return nil
		}

		fileViolations, scanErr := scanDomainFile(path, projectRoot)
		if scanErr != nil {
			return scanErr
		}

		violations = append(violations, fileViolations...)

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("filesystem walk failed: %w", walkErr)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintln(os.Stderr, v)
		}

		return fmt.Errorf("found %d domain layer isolation violations", len(violations))
	}

	logger.Log("Domain layer isolation check passed")

	return nil
}

// isDomainFile returns true if the file is inside a domain/ directory.
// Uses parent == dir termination to handle both Unix (root="/") and
// Windows (root="C:\") without an infinite loop.
func isDomainFile(path string) bool {
	dir := filepath.Dir(path)

	for {
		if filepath.Base(dir) == "domain" {
			return true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}

		dir = parent
	}
}

// forbiddenSuffixes are import path suffixes forbidden in domain packages.
var forbiddenSuffixes = []string{"/server", "/client", "/api"}

// scanDomainFile checks a single domain file for forbidden imports.
func scanDomainFile(filePath, projectRoot string) ([]string, error) {
	f, err := os.Open(filePath) //nolint:gosec // filePath from filepath.Walk, controlled
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filePath, err)
	}

	defer func() { _ = f.Close() }()

	rel, _ := filepath.Rel(projectRoot, filePath)

	var violations []string

	inImport := false

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "import (" {
			inImport = true

			continue
		}

		if inImport && trimmed == ")" {
			inImport = false

			continue
		}

		if !inImport && !strings.HasPrefix(trimmed, `import "`) {
			continue
		}

		importPath := extractImportPath(line)
		if importPath == "" {
			continue
		}

		for _, suffix := range forbiddenSuffixes {
			if strings.HasSuffix(importPath, suffix) || strings.Contains(importPath, suffix+"/") {
				violations = append(violations, fmt.Sprintf(
					"%s: domain package imports %s (forbidden: %s)", rel, importPath, suffix))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file %s: %w", filePath, err)
	}

	return violations, nil
}

// extractImportPath extracts the import path string from a line like `\t"path/to/pkg"` or `\talias "path/to/pkg"`.
func extractImportPath(line string) string {
	trimmed := strings.TrimSpace(line)
	// Handle alias imports: `alias "path"` or just `"path"`.
	start := strings.Index(trimmed, `"`)
	end := strings.LastIndex(trimmed, `"`)

	if start < 0 || start == end {
		return ""
	}

	return trimmed[start+1 : end]
}
