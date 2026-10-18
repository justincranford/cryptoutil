package format_text

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
)

func Format(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string) error {
	return formatWithFileOperations(logger, filesByExtension, formatFileOperations{
		readFile:  os.ReadFile,
		stat:      os.Stat,
		writeFile: os.WriteFile,
	})
}

type formatFileOperations struct {
	readFile  func(string) ([]byte, error)
	stat      func(string) (os.FileInfo, error)
	writeFile func(string, []byte, os.FileMode) error
}

func formatWithFileOperations(
	logger *cryptoutilCmdCicdCommon.Logger,
	filesByExtension map[string][]string,
	fileOperations formatFileOperations,
) error {
	files := cryptoutilCmdCicdCommon.FilterFilesForCommand(cryptoutilCmdCicdCommon.FlattenFileMap(filesByExtension), "format-text")
	sort.Strings(files)
	logger.Log(fmt.Sprintf("Normalizing line endings in %d text files", len(files)))

	var failures []error

	modified := 0

	for _, filePath := range files {
		content, err := fileOperations.readFile(filePath)
		if err != nil {
			failures = append(failures, fmt.Errorf("failed to read %s: %w", filePath, err))

			continue
		}

		normalized := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
		if bytes.Equal(content, normalized) {
			continue
		}

		fileInfo, err := fileOperations.stat(filePath)
		if err != nil {
			failures = append(failures, fmt.Errorf("failed to stat %s: %w", filePath, err))

			continue
		}

		if err := fileOperations.writeFile(filePath, normalized, fileInfo.Mode().Perm()); err != nil {
			failures = append(failures, fmt.Errorf("failed to write %s: %w", filePath, err))

			continue
		}

		modified++
	}

	logger.Log(fmt.Sprintf("format-text completed: processed=%d modified=%d", len(files), modified))

	if len(failures) > 0 {
		return fmt.Errorf("format-text failed for %d files: %w", len(failures), errors.Join(failures...))
	}

	return nil
}
