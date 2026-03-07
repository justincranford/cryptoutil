// Copyright (c) 2025 Justin Cranford

// Package migration_range_compliance verifies migration file version numbers are
// within assigned ranges: template migrations 1001-1999, domain migrations 2001+.
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

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	templateMigrationMin = 1001
	templateMigrationMax = 1999
	domainMigrationMin   = 2001
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

	// Check template migration directory.
	templateDir := filepath.Join(rootDir,
		"internal", "apps", cryptoutilSharedMagic.SkeletonTemplateServiceName,
		"service", "server", "repository", "migrations")

	templateViolations, err := checkDir(templateDir, templateMigrationMin, templateMigrationMax, true)
	if err != nil {
		return fmt.Errorf("checking template migrations: %w", err)
	}

	violations = append(violations, templateViolations...)

	// Check all domain migration directories.
	appsDir := filepath.Join(rootDir, "internal", "apps")

	domainDirs, err := findDomainMigrationDirs(appsDir, templateDir)
	if err != nil {
		return fmt.Errorf("finding domain migration dirs: %w", err)
	}

	for _, dir := range domainDirs {
		dirViolations, dirErr := checkDir(dir, domainMigrationMin, 0, false)
		if dirErr != nil {
			return fmt.Errorf("checking domain migrations in %s: %w", dir, dirErr)
		}

		violations = append(violations, dirViolations...)
	}

	if len(violations) > 0 {
		return fmt.Errorf("migration range compliance violations:\n%s", strings.Join(violations, "\n"))
	}

	logger.Log("Migration range compliance check passed")

	return nil
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
		label = cryptoutilSharedMagic.SkeletonTemplateServiceName
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
var nonFrameworkProducts = map[string]bool{
	cryptoutilSharedMagic.IdentityProductName: true,
}

// findDomainMigrationDirs finds all migrations/ directories under appsDir, excluding
// the template service and non-framework products. Skips _-prefixed directories.
func findDomainMigrationDirs(appsDir, templateDir string) ([]string, error) {
	if _, statErr := os.Stat(appsDir); os.IsNotExist(statErr) {
		return nil, nil
	}

	absTemplateDir, err := filepath.Abs(templateDir)
	if err != nil {
		return nil, fmt.Errorf("abs template dir: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("abs template dir: %w", err)
	}

	absAppsDir, err := filepath.Abs(appsDir)
	if err != nil {
		return nil, fmt.Errorf("abs apps dir: %w", err)
	}

	var dirs []string

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
		if len(parts) >= 1 && nonFrameworkProducts[parts[0]] {
			return filepath.SkipDir
		}

		if info.Name() != "migrations" {
			return nil
		}

		if absPath == absTemplateDir {
			return filepath.SkipDir
		}

		dirs = append(dirs, path)

		return filepath.SkipDir
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walk %s: %w", appsDir, walkErr)
	}

	return dirs, nil
}
