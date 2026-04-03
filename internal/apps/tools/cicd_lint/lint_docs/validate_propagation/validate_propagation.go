// Copyright (c) 2025 Justin Cranford

// Package validate_propagation validates that @source references in instruction
// files correspond to valid @propagate marker blocks.
package validate_propagation

import (
	"bytes"
	"fmt"
	"io"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilDocsValidation "cryptoutil/internal/apps/tools/cicd_lint/docs_validation"
)

// Check validates that all @source references have corresponding @propagate blocks.
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

		return fmt.Errorf("validate-propagation failed: broken @source references found")
	}

	return nil
}
