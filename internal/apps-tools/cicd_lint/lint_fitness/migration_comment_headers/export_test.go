// Copyright (c) 2025-2026 Justin Cranford.
package migration_comment_headers

import (
	"os"
	"path/filepath"

	lintFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
)

func ExportedCheckPSWithWalk(rootDir string, ps lintFitnessRegistry.ProductService, walkFn func(string, filepath.WalkFunc) error) []string {
	return checkPSWithWalk(rootDir, ps, walkFn)
}

func ExportedCheckMigrationDirWithReadDir(rootDir string, ps lintFitnessRegistry.ProductService, migDir string, readDirFn func(string) ([]os.DirEntry, error)) []string {
	return checkMigrationDirWithReadDir(rootDir, ps, migDir, readDirFn)
}

func ExportedCheckMigrationFileWithOpen(rootDir string, ps lintFitnessRegistry.ProductService, filePath, direction string, openFn func(string) (*os.File, error)) []string {
	return checkMigrationFileWithOpen(rootDir, ps, filePath, direction, openFn)
}
