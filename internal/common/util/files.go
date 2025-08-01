package util

import (
	"fmt"
	"os"
	"strings"
)

func ReadFilesBytes(filePaths *string) ([][]byte, error) {
	fileList := strings.Split(*filePaths, ",")
	if len(fileList) == 0 {
		return nil, fmt.Errorf("no files specified")
	}

	for i, filePath := range fileList {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			return nil, fmt.Errorf("empty file path %d of %d in list", i+1, len(fileList))
		}
	}

	filesContents := make([][]byte, 0, len(fileList))
	for i, filePath := range fileList {
		filePath = strings.TrimSpace(filePath)
		fileContents, err := ReadFileBytes(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %d of %d (%s): %w", i+1, len(fileList), filePath, err)
		}
		filesContents = append(filesContents, fileContents)
	}

	return filesContents, nil
}

func ReadFileBytes(filePath string) ([]byte, error) {
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return fileData, nil
}
