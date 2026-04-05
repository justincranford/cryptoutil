// Copyright (c) 2025 Justin Cranford

// Package configs_empty_dir enforces that every directory inside configs/
// contains at least one file. Empty directories must have a .gitkeep marker
// to be tracked by git and to make their presence intentional.
package configs_empty_dir

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Check runs the configs-empty-dir check from the current working directory.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return CheckInDir(logger, ".", os.Stat, filepath.WalkDir, os.ReadDir)
}

// CheckInDir checks rootDir/configs/ for empty directories without .gitkeep.
func CheckInDir(
	logger *cryptoutilCmdCicdCommon.Logger,
	rootDir string,
	statFn func(string) (os.FileInfo, error),
	walkFn func(string, fs.WalkDirFunc) error,
	readDirFn func(string) ([]os.DirEntry, error),
) error {
	logger.Log("Checking configs/ for empty directories without .gitkeep...")

	violations, err := FindViolationsInDir(rootDir, statFn, walkFn, readDirFn)
	if err != nil {
		return fmt.Errorf("failed to check configs/ for empty directories: %w", err)
	}

	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Fprintf(os.Stderr, "  empty directory without .gitkeep: %s\n", v)
		}

		return fmt.Errorf("found %d empty directories in configs/ without .gitkeep", len(violations))
	}

	logger.Log("configs-empty-dir: no empty directories found")

	return nil
}

// FindViolationsInDir walks rootDir/configs/ and returns paths of empty directories
// that do not contain a .gitkeep file.
func FindViolationsInDir(
	rootDir string,
	statFn func(string) (os.FileInfo, error),
	walkFn func(string, fs.WalkDirFunc) error,
	readDirFn func(string) ([]os.DirEntry, error),
) ([]string, error) {
	configsDir := filepath.Join(rootDir, cryptoutilSharedMagic.CICDConfigsDir)

	if _, err := statFn(configsDir); err != nil {
		return nil, fmt.Errorf("failed to access configs directory %s: %w", configsDir, err)
	}

	var violations []string

	err := walkFn(configsDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("failed to walk directory %s: %w", path, walkErr)
		}

		if !d.IsDir() {
			return nil
		}

		children, readErr := readDirFn(path)
		if readErr != nil {
			return fmt.Errorf("failed to read directory %s: %w", path, readErr)
		}

		if len(children) > 0 {
			return nil
		}

		violations = append(violations, path)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk configs directory: %w", err)
	}

	return violations, nil
}
