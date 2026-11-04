// Copyright (c) 2025-2026 Justin Cranford.
package common

import (
	"fmt"
	"os"
	"path/filepath"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedRepositoryScan "cryptoutil/internal/shared/util/repositoryscan"
)

// IsRepositoryScanExcludedDir reports whether name is generated output or dependency metadata.
func IsRepositoryScanExcludedDir(name string) bool {
	for _, exclusion := range cryptoutilSharedMagic.DirectoryNameExclusions {
		if filepath.Base(exclusion) == exclusion && name == exclusion {
			return true
		}
	}

	return false
}

// ShouldSkipRepositoryScanDir reports whether candidateDir should be skipped
// while scanning rootDir. The explicitly selected root is always retained,
// and unexpected repository-marker inspection errors are returned.
func ShouldSkipRepositoryScanDir(rootDir, candidateDir string) (bool, error) {
	return shouldSkipRepositoryScanDir(rootDir, candidateDir, os.Lstat)
}

func shouldSkipRepositoryScanDir(
	rootDir string,
	candidateDir string,
	lstatFn func(string) (os.FileInfo, error),
) (bool, error) {
	skip, err := cryptoutilSharedRepositoryScan.ShouldSkipDirWithLstat(
		rootDir,
		candidateDir,
		IsRepositoryScanExcludedDir,
		lstatFn,
	)
	if err != nil {
		return false, fmt.Errorf("inspect repository scan directory %s: %w", candidateDir, err)
	}

	return skip, nil
}
