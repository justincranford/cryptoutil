// Copyright (c) 2025 Justin Cranford

// Package validate_coverage validates that every required @propagate chunk from
// docs/required-propagations.yaml is covered by an @source block in at least one
// instruction or agent file.
package validate_coverage

import (
	"bytes"
	"fmt"
	"io"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilDocsValidation "cryptoutil/internal/apps/tools/cicd_lint/docs_validation"
)

// Check validates that required @propagate chunks are covered in instruction/agent files.
// Returns an error if any required chunk is missing or orphaned @propagate tags exist.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return check(logger, func(stdout, stderr io.Writer) int {
		return cryptoutilDocsValidation.ValidateCoverageCommand(stdout, stderr)
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
			return fmt.Errorf("validate-coverage failed: %s", stderr.String())
		}

		return fmt.Errorf("validate-coverage failed: required @propagate chunks are missing @source blocks")
	}

	return nil
}
