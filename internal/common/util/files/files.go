// Package files provides utilities for file operations.
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
func ListAllFiles(startDirectory string) ([]string, error) {
	var allFiles []string

	err := filepath.Walk(startDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			allFiles = append(allFiles, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return allFiles, nil
}
