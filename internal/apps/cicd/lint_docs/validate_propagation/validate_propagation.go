// Copyright (c) 2025 Justin Cranford

// Package validate_propagation validates that @source references in instruction
// files correspond to valid @propagate marker blocks.
package validate_propagation

import (
	"bytes"
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilDocsValidation "cryptoutil/internal/apps/cicd/docs_validation"
)

// Check validates that all @source references have corresponding @propagate blocks.
// Returns an error if any broken references are found.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilDocsValidation.ValidatePropagationCommand(&stdout, &stderr)

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
