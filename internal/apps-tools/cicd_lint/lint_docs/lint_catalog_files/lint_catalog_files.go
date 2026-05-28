// Copyright (c) 2025-2026 Justin Cranford.
// Package lint_catalog_files verifies that every @file-catalog and @file-catalog-pair
// entry in ENG-HANDBOOK.md Appendix D matches the corresponding file(s) on disk exactly.
// Frontmatter + body concatenation must produce the verbatim file content.
package lint_catalog_files

import (
	"bytes"
	"fmt"
	"io"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilDocsValidation "cryptoutil/internal/apps-tools/cicd_lint/docs_validation"
)

// Check verifies that all catalog entries in ENG-HANDBOOK.md Appendix D match the actual
// files on disk. Each @file-catalog entry must contain the complete verbatim file content;
// each @file-catalog-pair entry's reconstructed Copilot and Claude files must match disk.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return check(logger, func(stdout, stderr io.Writer) int {
		return cryptoutilDocsValidation.CatalogFilesCommand(stdout, stderr)
	})
}

func check(logger *cryptoutilCmdCicdCommon.Logger, fn func(io.Writer, io.Writer) int) error {
	var stdout, stderr bytes.Buffer

	exitCode := fn(&stdout, &stderr)

	if stdout.Len() > 0 {
		logger.Log(stdout.String())
	}

	if exitCode != 0 {
		if stderr.Len() > 0 {
			return fmt.Errorf("lint-catalog-files failed: %s", stderr.String())
		}

		return fmt.Errorf("lint-catalog-files failed: catalog file violations found")
	}

	return nil
}
