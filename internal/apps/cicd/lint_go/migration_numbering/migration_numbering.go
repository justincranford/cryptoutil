// Copyright (c) 2025 Justin Cranford

// Package migration_numbering validates SQL migration file naming conventions.
package migration_numbering

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

const (
	templateMigrationMin = 1001
	templateMigrationMax = 1999
	domainMigrationMin   = 2001
)

// migrationFilePattern matches migration SQL files like "2001_name.up.sql" or "2001_name.down.sql".
var migrationFilePattern = regexp.MustCompile(`^(\d+)_[a-z][a-z0-9_]*\.(up|down)\.sql$`)

// Test seams: replaceable in tests for error path coverage.
var (
	pathAbsFunc = filepath.Abs
	atoiFunc    = strconv.Atoi
)

// legacyMigrationPaths contains migration directories that use legacy numbering (pre-2001 convention).
// These are excluded from domain migration validation until they are migrated to the 2001+ scheme.
var legacyMigrationPaths = []string{
	filepath.Join("internal", "apps", "identity", "repository", "migrations"),
	filepath.Join("internal", "apps", "identity", "repository", "orm", "migrations"),
}

// Check validates migration numbering from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates migration numbering under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	var errors []string

	templateDir := filepath.Join(rootDir, "internal", "apps", "template", "service", "server", "repository", "migrations")
	if errs := checkMigrationDir(templateDir, templateMigrationMin, templateMigrationMax, true); len(errs) > 0 {
		errors = append(errors, errs...)
	}

	domainDirs, findErr := findDomainMigrationDirs(rootDir, templateDir)
	if findErr != nil {
		return fmt.Errorf("failed to find domain migration directories: %w", findErr)
	}

	for _, dir := range domainDirs {
		if errs := checkMigrationDir(dir, domainMigrationMin, 0, false); len(errs) > 0 {
			errors = append(errors, errs...)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("migration numbering violations:\n%s", strings.Join(errors, "\n"))
	}

	logger.Log("migration-numbering: all migration files pass naming validation")

	return nil
}

// findDomainMigrationDirs finds all migrations/ directories under internal/apps/ excluding the template service dir.
func findDomainMigrationDirs(rootDir, templateDir string) ([]string, error) {
	absTemplateDir, err := pathAbsFunc(templateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for template dir: %w", err)
	}

	absLegacyPaths := make(map[string]bool, len(legacyMigrationPaths))

	for _, lp := range legacyMigrationPaths {
		absLP, lpErr := pathAbsFunc(filepath.Join(rootDir, lp))
		if lpErr == nil {
			absLegacyPaths[absLP] = true
		}
	}

	var dirs []string

	appsDir := filepath.Join(rootDir, "internal", "apps")
	if _, statErr := os.Stat(appsDir); os.IsNotExist(statErr) {
		return nil, nil
	}

	walkErr := filepath.Walk(appsDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !info.IsDir() {
			return nil
		}

		// Skip _-prefixed directories (archived).
		if strings.HasPrefix(info.Name(), "_") {
			return filepath.SkipDir
		}

		if info.Name() != "migrations" {
			return nil
		}

		absPath, absErr := pathAbsFunc(path)
		if absErr != nil {
			return fmt.Errorf("failed to get absolute path: %w", absErr)
		}

		switch {
		case absPath == absTemplateDir:
			return nil
		case absLegacyPaths[absPath]:
			return nil
		default:
			dirs = append(dirs, path)
		}

		return filepath.SkipDir
	})

	sort.Strings(dirs)

	if walkErr != nil {
		return dirs, fmt.Errorf("failed to walk apps directory: %w", walkErr)
	}

	return dirs, nil
}

// checkMigrationDir validates migration files in a single directory.
// If maxVersion is 0, there is no upper bound.
func checkMigrationDir(dir string, minVersion, maxVersion int, isTemplate bool) []string {
	var errors []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return []string{fmt.Sprintf("%s: failed to read directory: %v", dir, err)}
	}

	versionPairs := make(map[int]map[string]bool) // version -> {"up": true, "down": true}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		matches := migrationFilePattern.FindStringSubmatch(name)

		if matches == nil {
			errors = append(errors, fmt.Sprintf("%s/%s: does not match migration naming pattern (expected NNNN_name.up.sql or NNNN_name.down.sql)", dir, name))

			continue
		}

		version, parseErr := atoiFunc(matches[1])
		if parseErr != nil {
			errors = append(errors, fmt.Sprintf("%s/%s: failed to parse version number: %v", dir, name, parseErr))

			continue
		}

		direction := matches[2]

		label := "domain"
		if isTemplate {
			label = "template"
		}

		if version < minVersion {
			errors = append(errors, fmt.Sprintf("%s/%s: %s migration version %d is below minimum %d", dir, name, label, version, minVersion))
		}

		if maxVersion > 0 && version > maxVersion {
			errors = append(errors, fmt.Sprintf("%s/%s: %s migration version %d exceeds maximum %d", dir, name, label, version, maxVersion))
		}

		if _, ok := versionPairs[version]; !ok {
			versionPairs[version] = make(map[string]bool)
		}

		versionPairs[version][direction] = true
	}

	// Check for missing up/down pairs.
	versions := make([]int, 0, len(versionPairs))
	for v := range versionPairs {
		versions = append(versions, v)
	}

	sort.Ints(versions)

	for _, v := range versions {
		dirs := versionPairs[v]
		if !dirs["up"] {
			errors = append(errors, fmt.Sprintf("%s: migration version %d is missing .up.sql file", dir, v))
		}

		if !dirs["down"] {
			errors = append(errors, fmt.Sprintf("%s: migration version %d is missing .down.sql file", dir, v))
		}
	}

	return errors
}
