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
	files := cryptoutilCmdCicdCommon.FilterFilesForCommand(cryptoutilCmdCicdCommon.FlattenFileMap(filesByExtension), "format-text")
	sort.Strings(files)
	logger.Log(fmt.Sprintf("Normalizing line endings in %d text files", len(files)))

	var failures []error

	modified := 0

	for _, filePath := range files {
		content, err := os.ReadFile(filePath)
		if err != nil {
			failures = append(failures, fmt.Errorf("failed to read %s: %w", filePath, err))

			continue
		}

		normalized := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
		if bytes.Equal(content, normalized) {
			continue
		}

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			failures = append(failures, fmt.Errorf("failed to stat %s: %w", filePath, err))

			continue
		}

		if err := os.WriteFile(filePath, normalized, fileInfo.Mode().Perm()); err != nil {
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
