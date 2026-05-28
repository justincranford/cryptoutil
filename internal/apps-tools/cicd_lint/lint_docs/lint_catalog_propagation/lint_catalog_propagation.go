// Copyright (c) 2025-2026 Justin Cranford.
// Package lint_catalog_propagation verifies that every @to-appendix chunk in
// ENG-HANDBOOK.md that targets a catalogued file (one with a @file-catalog or
// @file-catalog-pair entry in Appendix D) also has a matching @from-eng-handbook
// block with identical content inside that catalog entry's body.
package lint_catalog_propagation

import (
	"bytes"
	"fmt"
	"io"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilDocsValidation "cryptoutil/internal/apps-tools/cicd_lint/docs_validation"
)

// Check verifies that @to-appendix chunk content is reflected verbatim inside the
// @file-catalog / @file-catalog-pair body for each target file that is catalogued.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return check(logger, func(stdout, stderr io.Writer) int {
		return cryptoutilDocsValidation.CatalogPropagationCommand(stdout, stderr)
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
			return fmt.Errorf("lint-catalog-propagation failed: %s", stderr.String())
		}

		return fmt.Errorf("lint-catalog-propagation failed: catalog propagation violations found")
	}

	return nil
}
