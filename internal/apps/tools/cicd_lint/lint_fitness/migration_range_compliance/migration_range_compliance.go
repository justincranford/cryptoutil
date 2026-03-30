// Copyright (c) 2025 Justin Cranford

// Package migration_range_compliance verifies migration file version numbers are
// within assigned ranges: template migrations 1001-1999, domain migrations per-PS-ID
// ranges declared in the entity registry YAML. Also detects cross-service range overlaps.
// This is complementary to migration_numbering (which checks naming patterns and
// up/down pairing); this check focuses exclusively on the numeric range constraint.
package migration_range_compliance

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	lintFitnessRegistry "cryptoutil/internal/apps/tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	templateMigrationMin = 1001
	templateMigrationMax = 1999
)

// migrationVersionPattern extracts the leading version number from a migration filename.
var migrationVersionPattern = regexp.MustCompile(`^(\d+)_`)

// Check verifies migration range compliance from the workspace root.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir verifies migration range compliance under rootDir.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking migration range compliance...")

	var violations []string

	// Step 1: Check registry-declared ranges for cross-service collisions.
	collisions := checkRegistryRangeCollisions()
	violations = append(violations, collisions...)

	// Step 2: Check template migration directory.
	templateDir := filepath.Join(rootDir,
		"internal", "apps", cryptoutilSharedMagic.FrameworkProductName,
		"service", "server", "repository", "migrations")

	templateViolations, err := checkDir(templateDir, templateMigrationMin, templateMigrationMax, true)
	if err != nil {
		return fmt.Errorf("checking template migrations: %w", err)
	}

	violations = append(violations, templateViolations...)

	// Step 3: Check all domain migration directories with per-PS-ID ranges.
	appsDir := filepath.Join(rootDir, "internal", "apps")

	domainDirs, err := findDomainMigrationDirsWithPSID(appsDir, templateDir)
	if err != nil {
		return fmt.Errorf("finding domain migration dirs: %w", err)
	}

	// Build per-PS-ID range map from registry.
	rangeMap := buildPSIDRangeMap()

	for _, entry := range domainDirs {
		rangeInfo, ok := rangeMap[entry.psID]
		if !ok {
			// PS-ID not in registry — use loose lower bound only.
			dirViolations, dirErr := checkDir(entry.dir, templateMigrationMax+1, 0, false)
			if dirErr != nil {
				return fmt.Errorf("checking migrations in %s: %w", entry.dir, dirErr)
			}

			violations = append(violations, dirViolations...)

			continue
		}

		dirViolations, dirErr := checkDir(entry.dir, rangeInfo.Start, rangeInfo.End, false)
		if dirErr != nil {
			return fmt.Errorf("checking migrations in %s: %w", entry.dir, dirErr)
		}

		violations = append(violations, dirViolations...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("migration range compliance violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("Migration range compliance check passed")

	return nil
}

// migrationDirEntry holds a migrations directory path and its associated PS-ID.
type migrationDirEntry struct {
	dir  string
	psID string
}

// buildPSIDRangeMap returns a map of PS-ID → MigrationRangeInfo from the entity registry.
func buildPSIDRangeMap() map[string]lintFitnessRegistry.MigrationRangeInfo {
	ranges := lintFitnessRegistry.AllMigrationRanges()
	m := make(map[string]lintFitnessRegistry.MigrationRangeInfo, len(ranges))

	for _, r := range ranges {
		m[r.PSID] = r
	}

	return m
}

// checkRegistryRangeCollisions detects cross-service migration range overlaps declared in
// the entity registry. Two PS-IDs overlap if their [start, end] intervals intersect.
func checkRegistryRangeCollisions() []string {
	return checkRangeCollisions(lintFitnessRegistry.AllMigrationRanges())
}

// checkRangeCollisions detects overlapping intervals in the provided ranges slice.
func checkRangeCollisions(ranges []lintFitnessRegistry.MigrationRangeInfo) []string {
	var violations []string

	for i := 0; i < len(ranges); i++ {
		for j := i + 1; j < len(ranges); j++ {
			a, b := ranges[i], ranges[j]
			// Intervals [a.Start, a.End] and [b.Start, b.End] overlap iff a.Start <= b.End && b.Start <= a.End.
			if a.Start <= b.End && b.Start <= a.End {
				violations = append(violations, fmt.Sprintf(
					"cross-service range collision: %s [%d-%d] overlaps %s [%d-%d]",
					a.PSID, a.Start, a.End, b.PSID, b.Start, b.End))
			}
		}
	}

	return violations
}

// checkDir validates that all migration SQL files in dir have version numbers within [min, max].
// If max is 0, there is no upper bound.
func checkDir(dir string, min, max int, isTemplate bool) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("read dir %s: %w", dir, err)
	}

	label := "domain"
	if isTemplate {
		label = cryptoutilSharedMagic.FrameworkProductName
	}

	var violations []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		matches := migrationVersionPattern.FindStringSubmatch(name)
		if matches == nil {
			continue // naming errors caught by migration_numbering
		}

		version, parseErr := strconv.Atoi(matches[1])
		if parseErr != nil {
			violations = append(violations, fmt.Sprintf("%s/%s: failed to parse version: %v", dir, name, parseErr))

			continue
		}

		if version < min {
			violations = append(violations, fmt.Sprintf(
				"%s/%s: %s migration version %d is below minimum %d",
				dir, name, label, version, min))
		}

		if max > 0 && version > max {
			violations = append(violations, fmt.Sprintf(
				"%s/%s: %s migration version %d exceeds maximum %d",
				dir, name, label, version, max))
		}
	}

	return violations, nil
}

// nonFrameworkProducts are products not yet migrated to the template service framework.
// Their migrations use legacy numbering and are excluded from domain range compliance.
// With flat PS-ID structure (e.g. "identity-idp/"), matching checks for prefix "identity-".
var nonFrameworkProducts = []string{
	cryptoutilSharedMagic.IdentityProductName,
}

// isNonFrameworkProduct checks if a directory name belongs to a non-framework product.
// Matches both product-level dirs (e.g. "identity") and flat PS-ID dirs (e.g. "identity-idp").
func isNonFrameworkProduct(dirName string) bool {
	for _, product := range nonFrameworkProducts {
		if dirName == product || strings.HasPrefix(dirName, product+"-") {
			return true
		}
	}

	return false
}

// findDomainMigrationDirsWithPSID finds all migrations/ directories under appsDir, excluding
// the template service and non-framework products. Returns (dir, psID) pairs where psID is
// derived from the first path component under appsDir.
func findDomainMigrationDirsWithPSID(appsDir, templateDir string) ([]migrationDirEntry, error) {
	if _, statErr := os.Stat(appsDir); os.IsNotExist(statErr) {
		return nil, nil
	}

	absTemplateDir, err := filepath.Abs(templateDir)
	if err != nil {
		return nil, fmt.Errorf("abs template dir: %w", err)
	}

	absAppsDir, err := filepath.Abs(appsDir)
	if err != nil {
		return nil, fmt.Errorf("abs apps dir: %w", err)
	}

	var entries []migrationDirEntry

	walkErr := filepath.Walk(appsDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !info.IsDir() {
			return nil
		}

		if strings.HasPrefix(info.Name(), "_") {
			return filepath.SkipDir
		}

		// Skip non-framework products (using legacy migration numbering).
		absPath, absErr := filepath.Abs(path)
		if absErr != nil {
			return fmt.Errorf("abs path %s: %w", path, absErr)
		}

		rel, _ := filepath.Rel(absAppsDir, absPath)

		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) >= 1 && isNonFrameworkProduct(parts[0]) {
			return filepath.SkipDir
		}

		if info.Name() != "migrations" {
			return nil
		}

		if absPath == absTemplateDir {
			return filepath.SkipDir
		}

		// Derive PS-ID from the first component under appsDir.
		psID := ""
		if len(parts) >= 1 {
			psID = parts[0]
		}

		entries = append(entries, migrationDirEntry{dir: path, psID: psID})

		return filepath.SkipDir
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walk %s: %w", appsDir, walkErr)
	}

	return entries, nil
}

// findDomainMigrationDirs is a backward-compatible wrapper for tests that only need directories.
func findDomainMigrationDirs(appsDir, templateDir string) ([]string, error) {
	entries, err := findDomainMigrationDirsWithPSID(appsDir, templateDir)
	if err != nil {
		return nil, err
	}

	dirs := make([]string, len(entries))

	for i, e := range entries {
		dirs[i] = e.dir
	}

	return dirs, nil
}
