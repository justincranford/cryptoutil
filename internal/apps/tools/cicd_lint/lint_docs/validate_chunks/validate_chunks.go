// Copyright (c) 2025 Justin Cranford

// Package validate_chunks validates that @propagate blocks in ENG-HANDBOOK.md
// match their @source counterparts in instruction files.
package validate_chunks

import (
	"bytes"
	"fmt"
	"io"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilDocsValidation "cryptoutil/internal/apps/tools/cicd_lint/docs_validation"
)

// Check validates that all @propagate blocks match their @source counterparts.
// Returns an error if any chunks are mismatched, missing, or have file errors.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return check(logger, func(stdout, stderr io.Writer) int {
		return cryptoutilDocsValidation.ValidateChunksCommand(stdout, stderr)
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
			return fmt.Errorf("validate-chunks failed: %s", stderr.String())
		}

		return fmt.Errorf("validate-chunks failed: propagated chunks are out of sync")
	}

	return nil
}
