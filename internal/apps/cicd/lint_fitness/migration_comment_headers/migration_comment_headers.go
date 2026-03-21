// Copyright (c) 2025 Justin Cranford

// Package migration_comment_headers validates that every domain migration file
// (number 2001+) has the canonical comment header on its first non-blank line.
//
// Required format:
//   - *.up.sql   first non-blank comment line: "-- {Display Name} database schema"
//   - *.down.sql first non-blank comment line: "-- {Display Name} database schema rollback"
//
// Framework migration files in the 1001-1999 range are excluded.
// The "first non-blank comment line" is defined as the first line starting with
// "-- " (double-dash space) where the line contains more than just "--".
//
// This check would have caught regressions where the service's display name
// in migration headers doesn't match the current canonical name in the registry.
package migration_comment_headers

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	lintFitnessRegistry "cryptoutil/internal/apps/cicd/lint_fitness/registry"
)

const (
	domainMigrationMin = 2001
)

// migrationFileRe matches migration SQL files: "2001_name.up.sql" or "2001_name.down.sql".
var migrationFileRe = regexp.MustCompile(`^(\d+)_[a-z][a-z0-9_]*\.(up|down)\.sql$`)

// Check validates migration comment headers from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates migration comment headers under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking migration comment headers...")

	var violations []string

	for _, ps := range lintFitnessRegistry.AllProductServices() {
		v := checkPS(rootDir, ps)
		violations = append(violations, v...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("migration comment header violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("migration-comment-headers: all domain migration files have correct comment headers")

	return nil
}

// checkPS finds all migrations/ directories under the PS's InternalAppsDir and checks them.
func checkPS(rootDir string, ps lintFitnessRegistry.ProductService) []string {
	psDir := filepath.Join(rootDir, "internal", "apps", filepath.FromSlash(ps.InternalAppsDir))

	if _, err := os.Stat(psDir); os.IsNotExist(err) {
		return nil
	}

	var migrationDirs []string

	walkErr := filepath.Walk(psDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		// Skip archived directories (underscore prefix).
		if strings.HasPrefix(info.Name(), "_") {
			return filepath.SkipDir
		}

		if info.Name() == "migrations" {
			migrationDirs = append(migrationDirs, path)

			return filepath.SkipDir
		}

		return nil
	})
	if walkErr != nil {
		return []string{fmt.Sprintf("%s: error walking directory: %v", ps.PSID, walkErr)}
	}

	var violations []string

	for _, migDir := range migrationDirs {
		v := checkMigrationDir(rootDir, ps, migDir)
		violations = append(violations, v...)
	}

	return violations
}

// checkMigrationDir scans all domain migration SQL files (2001+) in a directory.
func checkMigrationDir(rootDir string, ps lintFitnessRegistry.ProductService, migDir string) []string {
	entries, err := os.ReadDir(migDir)
	if err != nil {
		return []string{fmt.Sprintf("%s: cannot read migrations dir %s: %v", ps.PSID, migDir, err)}
	}

	var violations []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		m := migrationFileRe.FindStringSubmatch(entry.Name())
		if m == nil {
			continue
		}

		num, _ := strconv.Atoi(m[1])
		if num < domainMigrationMin {
			// Framework migrations (1001-1999) are excluded.
			continue
		}

		direction := m[2] // "up" or "down"
		filePath := filepath.Join(migDir, entry.Name())

		v := checkMigrationFile(rootDir, ps, filePath, direction)
		violations = append(violations, v...)
	}

	return violations
}

// checkMigrationFile verifies the first non-blank comment line of a migration file.
func checkMigrationFile(rootDir string, ps lintFitnessRegistry.ProductService, filePath, direction string) []string {
	f, err := os.Open(filePath) //nolint:gosec // filePath from controlled directory walk
	if err != nil {
		rel, _ := filepath.Rel(rootDir, filePath)

		return []string{fmt.Sprintf("%s: cannot open %s: %v", ps.PSID, rel, err)}
	}

	defer func() { _ = f.Close() }()

	var expectedContains string
	if direction == "up" {
		expectedContains = ps.DisplayName + " database schema"
	} else {
		expectedContains = ps.DisplayName + " database schema rollback"
	}

	firstContentLine := ""

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		// A content comment line starts with "-- " and has more than just "--".
		if strings.HasPrefix(line, "-- ") {
			firstContentLine = line

			break
		}
	}

	if scanner.Err() != nil {
		rel, _ := filepath.Rel(rootDir, filePath)

		return []string{fmt.Sprintf("%s: %s: read error: %v", ps.PSID, rel, scanner.Err())}
	}

	if firstContentLine == "" {
		rel, _ := filepath.Rel(rootDir, filePath)

		return []string{fmt.Sprintf("%s: %s: no comment header found (expected %q)", ps.PSID, rel, "-- "+expectedContains)}
	}

	if !strings.Contains(firstContentLine, expectedContains) {
		rel, _ := filepath.Rel(rootDir, filePath)

		return []string{fmt.Sprintf("%s: %s: first comment %q does not contain %q", ps.PSID, rel, firstContentLine, "-- "+expectedContains)}
	}

	return nil
}
