// Copyright (c) 2025-2026 Justin Cranford.
package template_consistency

import "os"

func ExportedFindViolationsInDirWithReadDir(
	rootDir string,
	readDirFn func(string) ([]os.DirEntry, error),
) ([]string, error) {
	return findViolationsInDir(rootDir, readDirFn)
}
