// Copyright (c) 2025 Justin Cranford

// Package archive_detector enforces that archived or orphaned directories
// do not exist in the repository. Directories named "_archived/", "archived/",
// or "orphaned/" represent dead code and must be removed per the Archive and
// Dead Code Policy (ARCHITECTURE.md Section 13.9).
package archive_detector

import (
	"fmt"
	"os"
	"path/filepath"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
)

// bannedDirNames is the set of directory names that must NOT exist anywhere in
// the repository tree.
var bannedDirNames = []string{
	"_archived",
	"archived",
	"orphaned",
}

// Check runs the archive-detector check from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".")
}

// CheckInDir checks rootDir for archived or orphaned directory presence.
func CheckInDir(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	logger.Log("Checking for archived/orphaned directory presence...")

	violations, err := FindViolationsInDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to check for archived directories: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "  archived/orphaned directory found: %s\n", v)
		}

		return fmt.Errorf("found %d archived/orphaned directories that must be removed", len(violations))
	}

	logger.Log("archive-detector: no archived/orphaned directories found")

	return nil
}

// FindViolationsInDir walks rootDir and returns paths of any banned directories.
func FindViolationsInDir(rootDir string) ([]string, error) {
	var violations []string

	bannedSet := make(map[string]struct{}, len(bannedDirNames))
	for _, name := range bannedDirNames {
		bannedSet[name] = struct{}{}
	}

	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk directory %s: %w", path, err)
		}

		if d.IsDir() {
			if _, banned := bannedSet[d.Name()]; banned {
				violations = append(violations, path)

				return filepath.SkipDir
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk repository: %w", err)
	}

	return violations, nil
}
