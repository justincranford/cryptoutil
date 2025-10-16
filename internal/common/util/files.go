package util

import (
	"fmt"
	"os"
	"strings"
)

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

func ReadFileBytes(filePath string) ([]byte, error) {
	fileData, err := os.ReadFile(filePath) // #nosec G304 -- General purpose file utility function
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	return fileData, nil
}

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
