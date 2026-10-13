// Copyright (c) 2025-2026 Justin Cranford.
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

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	templateMigrationMin                           = 1001
	templateMigrationMax                           = 1999
	domainMigrationMin                             = 2001
	migrationNumberingWalkAppsDirectoryErrorPrefix = "failed to walk apps directory"
)

// migrationFilePattern matches migration SQL files like "2001_name.up.sql" or "2001_name.down.sql".
var migrationFilePattern = regexp.MustCompile(`^(\d+)_[a-z][a-z0-9_]*\.(up|down)\.sql$`)

type migrationNumberingOperations struct {
	pathAbsFn func(string) (string, error)
	atoiFn    func(string) (int, error)
	statFn    func(string) (os.FileInfo, error)
	walkFn    func(string, filepath.WalkFunc) error
	readDirFn func(string) ([]os.DirEntry, error)
}

func productionMigrationNumberingOperations() migrationNumberingOperations {
	return migrationNumberingOperations{
		pathAbsFn: filepath.Abs,
		atoiFn:    strconv.Atoi,
		statFn:    os.Stat,
		walkFn:    filepath.Walk,
		readDirFn: os.ReadDir,
	}
}

// legacyMigrationPaths contains migration directories that use legacy numbering (pre-2001 convention).
// These are excluded from domain migration validation until they are migrated to the 2001+ scheme.
var legacyMigrationPaths = []string{
	filepath.Join("internal", "apps", cryptoutilSharedMagic.IdentityProductName, "repository", "migrations"),
	filepath.Join("internal", "apps", cryptoutilSharedMagic.IdentityProductName, "repository", "orm", "migrations"),
}

// Check validates migration numbering from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir validates migration numbering under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDirWithOperations(logger, rootDir, productionMigrationNumberingOperations())
}

func checkInDirWithOperations(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, operations migrationNumberingOperations) error {
	var errors []string

	templateDir := filepath.Join(rootDir, cryptoutilSharedMagic.FrameworkInternalDir, "service", cryptoutilSharedMagic.CMD_SERVER, "repository", "migrations")
	errors = append(errors, checkMigrationDirWithOperations(templateDir, templateMigrationMin, templateMigrationMax, true, operations)...)

	domainDirs, findErr := findDomainMigrationDirsWithOperations(rootDir, templateDir, operations)
	if findErr != nil {
		return fmt.Errorf("failed to find domain migration directories: %w", findErr)
	}

	for _, dir := range domainDirs {
		errors = append(errors, checkMigrationDirWithOperations(dir, domainMigrationMin, 0, false, operations)...)
	}

	if len(errors) > 0 {
		return fmt.Errorf("migration numbering violations:\n%s", strings.Join(errors, "\n"))
	}

	logger.Log("migration-numbering: all migration files pass naming validation")

	return nil
}

// findDomainMigrationDirs finds all migrations/ directories under internal/apps/ excluding the template service dir.
func findDomainMigrationDirs(rootDir, templateDir string) ([]string, error) {
	return findDomainMigrationDirsWithOperations(rootDir, templateDir, productionMigrationNumberingOperations())
}

func findDomainMigrationDirsWithOperations(rootDir, templateDir string, operations migrationNumberingOperations) ([]string, error) {
	absTemplateDir, err := operations.pathAbsFn(templateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for template dir: %w", err)
	}

	absLegacyPaths := make(map[string]bool, len(legacyMigrationPaths))

	for _, lp := range legacyMigrationPaths {
		absLP, lpErr := operations.pathAbsFn(filepath.Join(rootDir, lp))
		if lpErr != nil {
			return nil, fmt.Errorf("failed to get absolute path for legacy migration dir %s: %w", lp, lpErr)
		}

		absLegacyPaths[absLP] = true
	}

	var dirs []string

	appsDir := filepath.Join(rootDir, "internal", "apps")
	if _, statErr := operations.statFn(appsDir); statErr != nil {
		if os.IsNotExist(statErr) {
			return nil, fmt.Errorf("internal/apps directory not found at %s", appsDir)
		}

		return nil, fmt.Errorf("failed to stat internal/apps directory %s: %w", appsDir, statErr)
	}

	walkErr := operations.walkFn(appsDir, func(path string, info os.FileInfo, walkErr error) error {
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

		absPath, absErr := operations.pathAbsFn(path)
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
		return dirs, fmt.Errorf("%s: %w", migrationNumberingWalkAppsDirectoryErrorPrefix, walkErr)
	}

	return dirs, nil
}

// checkMigrationDir validates migration files in a single directory.
// If maxVersion is 0, there is no upper bound.
func checkMigrationDir(dir string, minVersion, maxVersion int, isTemplate bool) []string {
	return checkMigrationDirWithOperations(dir, minVersion, maxVersion, isTemplate, productionMigrationNumberingOperations())
}

func checkMigrationDirWithOperations(dir string, minVersion, maxVersion int, isTemplate bool, operations migrationNumberingOperations) []string {
	var errors []string

	entries, err := operations.readDirFn(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{fmt.Sprintf("%s: directory does not exist", dir)}
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

		version, parseErr := operations.atoiFn(matches[1])
		if parseErr != nil {
			errors = append(errors, fmt.Sprintf("%s/%s: failed to parse version number: %v", dir, name, parseErr))

			continue
		}

		direction := matches[2]

		label := "domain"
		if isTemplate {
			label = cryptoutilSharedMagic.FrameworkProductName
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
