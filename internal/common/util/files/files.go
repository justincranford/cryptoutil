// Copyright (c) 2025 Justin Cranford
//
//

package files

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile writes content to a file with the specified permissions.
// If permissions is 0, it defaults to CacheFilePermissions.
// The content can be a string or []byte.
func WriteFile(filePath string, bytesOrString any, permissions os.FileMode) error {
	if permissions == 0 {
		return fmt.Errorf("missing file permissions")
	}

	var data []byte

	switch v := bytesOrString.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("content must be string or []byte, got %T", bytesOrString)
	}

	err := os.WriteFile(filePath, data, permissions)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// ListAllFiles walks the directory tree starting from startDirectory and returns all file paths.
// It excludes directories and only includes regular files.
// Paths are normalized to use forward slashes for cross-platform compatibility.
// Excluded directories are skipped entirely.
func ListAllFiles(startDirectory string, exclusions ...string) ([]string, error) {
	var allFiles []string

	err := filepath.Walk(startDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			for _, excl := range exclusions {
				if path == excl || (len(path) > len(excl) && path[:len(excl)+1] == excl+"/") {
					return filepath.SkipDir
				}
			}

			return nil
		}

		// Normalize path to forward slashes
		normalizedPath := filepath.ToSlash(path)
		allFiles = append(allFiles, normalizedPath)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return allFiles, nil
}
