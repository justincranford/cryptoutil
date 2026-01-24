// Copyright (c) 2025 Justin Cranford
//
//

// Package files provides file system utility functions.
package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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

// ListAllFiles walks the directory tree starting from startDirectory and returns files grouped by extension.
// It uses the extension inclusions and directory exclusions from magic constants.
// Returns a map where keys are file extensions (without dot) and values are slices of relative file paths.
// Paths are normalized to use forward slashes for cross-platform compatibility.
// Excluded directories are skipped entirely.
func ListAllFiles(startDirectory string) (map[string][]string, error) {
	return ListAllFilesWithOptions(startDirectory, cryptoutilSharedMagic.TextFilenameExtensionInclusions, cryptoutilSharedMagic.DirectoryNameExclusions)
}

// ListAllFilesWithOptions walks the directory tree starting from startDirectory and returns files grouped by extension.
// It uses the provided extension inclusions and directory exclusions.
// Returns a map where keys are file extensions (without dot) and values are slices of relative file paths.
// Paths are normalized to use forward slashes for cross-platform compatibility.
// Excluded directories are skipped entirely.
func ListAllFilesWithOptions(startDirectory string, inclusions []string, exclusions []string) (map[string][]string, error) {
	matches := make(map[string][]string)

	// Build a set of included extensions for fast lookup.
	includedExtensions := make(map[string]bool)
	for _, ext := range inclusions {
		includedExtensions[ext] = true
	}

	err := filepath.Walk(startDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Normalize path to forward slashes for consistency.
		normalizedPath := filepath.ToSlash(path)

		if info.IsDir() {
			for _, excl := range exclusions {
				// Check if the path matches the exclusion (exact match or prefix).
				if normalizedPath == excl || strings.HasPrefix(normalizedPath, excl+"/") {
					return filepath.SkipDir
				}
			}

			return nil
		}

		// Extract the extension (without the dot).
		ext := strings.TrimPrefix(filepath.Ext(path), ".")

		// Handle files without extension (like .gitignore, .dockerignore).
		if ext == "" {
			// Check if the base name itself is in inclusions (e.g., "gitignore" for ".gitignore").
			baseName := filepath.Base(path)
			if strings.HasPrefix(baseName, ".") {
				ext = strings.TrimPrefix(baseName, ".")
			}
		}

		// Only include files with matching extensions.
		if includedExtensions[ext] {
			matches[ext] = append(matches[ext], normalizedPath)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return matches, nil
}

// ReadFilesBytes reads the contents of multiple files and returns them as byte slices.
func ReadFilesBytes(filePaths []string) ([][]byte, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("no files specified")
	}

	for i, filePath := range filePaths {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			return nil, fmt.Errorf("empty file path %d of %d in list", i+1, len(filePaths))
		}
	}

	filesContents := make([][]byte, 0, len(filePaths))

	for i, filePath := range filePaths {
		filePath = strings.TrimSpace(filePath)

		fileContents, err := ReadFileBytes(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %d of %d (%s): %w", i+1, len(filePaths), filePath, err)
		}

		filesContents = append(filesContents, fileContents)
	}

	return filesContents, nil
}

// ReadFileBytes reads the contents of a single file and returns it as a byte slice.
func ReadFileBytes(filePath string) ([]byte, error) {
	fileData, err := os.ReadFile(filePath) // #nosec G304 -- General purpose file utility function
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return fileData, nil
}

// ReadFilesBytesLimit reads multiple files with size and count limits.
func ReadFilesBytesLimit(filePaths []string, maxFiles, maxBytesPerFile int64) ([][]byte, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("no files specified")
	} else if len(filePaths) > int(maxFiles) {
		return nil, fmt.Errorf("too many files specified: maximum is %d, got %d", maxFiles, len(filePaths))
	}

	for i, filePath := range filePaths {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			return nil, fmt.Errorf("empty file path %d of %d in list", i+1, len(filePaths))
		}
	}

	filesContents := make([][]byte, 0, len(filePaths))

	for i, filePath := range filePaths {
		filePath = strings.TrimSpace(filePath)

		fileContents, err := ReadFileBytesLimit(filePath, maxBytesPerFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %d of %d (%s): %w", i+1, len(filePaths), filePath, err)
		}

		filesContents = append(filesContents, fileContents)
	}

	return filesContents, nil
}

// ReadFileBytesLimit reads a single file with a size limit.
func ReadFileBytesLimit(filePath string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		// If no limit or negative value, read the entire file
		return ReadFileBytes(filePath)
	}

	// Open file
	file, err := os.Open(filePath) // #nosec G304 -- General purpose file utility function
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log error but don't override the main return error
			fmt.Printf("Warning: failed to close file %s: %v\n", filePath, closeErr)
		}
	}()

	// Get file info to determine file size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats for %s: %w", filePath, err)
	}

	// Determine how many bytes to read
	bytesToRead := maxBytes
	if fileInfo.Size() < maxBytes {
		bytesToRead = fileInfo.Size()
	}

	// Read the limited bytes
	buffer := make([]byte, bytesToRead)

	n, err := file.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("failed to read bytes from file %s: %w", filePath, err)
	}

	return buffer[:n], nil
}
