// Copyright (c) 2025-2026 Justin Cranford.
// Package repositoryscan provides root-aware directory scan exclusions.
package repositoryscan

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ShouldSkipDir reports whether candidateDir should be skipped during a scan
// rooted at rootDir. The explicitly selected root is always retained.
func ShouldSkipDir(rootDir, candidateDir string, isExcludedDir func(string) bool) (bool, error) {
	return ShouldSkipDirWithLstat(rootDir, candidateDir, isExcludedDir, os.Lstat)
}

// ShouldSkipDirWithLstat applies the root-aware exclusion rules using lstatFn
// to inspect nested repository markers.
func ShouldSkipDirWithLstat(
	rootDir string,
	candidateDir string,
	isExcludedDir func(string) bool,
	lstatFn func(string) (os.FileInfo, error),
) (bool, error) {
	if filepath.Clean(rootDir) == filepath.Clean(candidateDir) {
		return false, nil
	}

	if isExcludedDir != nil && isExcludedDir(filepath.Base(candidateDir)) {
		return true, nil
	}

	return HasGitMarkerWithLstat(candidateDir, lstatFn)
}

// ShouldSkipDirFS applies the root-aware exclusion rules to slash-separated
// paths in fsys. Marker inspection uses fsys itself and does not follow links.
func ShouldSkipDirFS(fsys fs.FS, rootDir, candidateDir string, isExcludedDir func(string) bool) (bool, error) {
	rootDir = path.Clean(rootDir)
	candidateDir = path.Clean(candidateDir)

	if candidateDir == rootDir {
		return false, nil
	}

	if !fs.ValidPath(candidateDir) {
		return false, fmt.Errorf("invalid filesystem directory path %q", candidateDir)
	}

	if isExcludedDir != nil && isExcludedDir(path.Base(candidateDir)) {
		return true, nil
	}

	return HasGitMarkerFS(fsys, candidateDir)
}

// HasGitMarkerWithLstat reports whether candidateDir contains a .git entry.
// Lstat deliberately recognizes files, directories, symlinks, and dangling
// symlinks without following a marker into another checkout.
func HasGitMarkerWithLstat(candidateDir string, lstatFn func(string) (os.FileInfo, error)) (bool, error) {
	if lstatFn == nil {
		return false, errors.New("missing Lstat function")
	}

	gitMarkerPath := filepath.Join(candidateDir, cryptoutilSharedMagic.CICDExcludeDirGit)

	_, err := lstatFn(gitMarkerPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}

		return false, fmt.Errorf("failed to inspect repository marker %s: %w", gitMarkerPath, err)
	}

	return true, nil
}

// HasGitMarkerFS reports whether candidateDir contains a .git directory entry.
// ReadDir inspects the supplied filesystem without following symlinks, so
// dangling symlink markers are detected as well.
func HasGitMarkerFS(fsys fs.FS, candidateDir string) (bool, error) {
	if fsys == nil {
		return false, errors.New("missing filesystem")
	}

	if !fs.ValidPath(candidateDir) {
		return false, fmt.Errorf("invalid filesystem directory path %q", candidateDir)
	}

	entries, err := fs.ReadDir(fsys, candidateDir)
	if err != nil {
		return false, fmt.Errorf("failed to inspect repository marker under %s: %w", candidateDir, err)
	}

	for _, entry := range entries {
		if entry.Name() == cryptoutilSharedMagic.CICDExcludeDirGit {
			return true, nil
		}
	}

	return false, nil
}
