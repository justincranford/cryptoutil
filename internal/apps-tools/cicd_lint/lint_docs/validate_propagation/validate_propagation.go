// Copyright (c) 2025-2026 Justin Cranford.
// Package validate_propagation validates that ENG-HANDBOOK cross-references in
// instruction and agent files resolve to valid handbook anchors.
package validate_propagation

import (
	"bytes"
	"fmt"
	"io"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilDocsValidation "cryptoutil/internal/apps-tools/cicd_lint/docs_validation"
)

// Check validates that handbook cross-references resolve to existing anchors.
// Returns an error if any broken references are found.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return check(logger, func(stdout, stderr io.Writer) int {
		return cryptoutilDocsValidation.ValidatePropagationCommand(stdout, stderr)
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
			return fmt.Errorf("validate-propagation failed: %s", stderr.String())
		}

		return fmt.Errorf("validate-propagation failed: broken ENG-HANDBOOK cross-references found")
	}

	return nil
}
