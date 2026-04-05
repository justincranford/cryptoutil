// Copyright (c) 2025 Justin Cranford

// Package check_chunk_verification verifies that ENG-HANDBOOK.md section chunks
// are present in their target instruction files.
package check_chunk_verification

import (
	"bytes"
	"fmt"
	"io"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilDocsValidation "cryptoutil/internal/apps/tools/cicd_lint/docs_validation"
)

// Check verifies that all architecture chunk references exist in instruction files.
// Returns an error if any chunk references are missing.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	return check(logger, func(stdout, stderr io.Writer) int {
		return cryptoutilDocsValidation.CheckChunkVerification(stdout, stderr)
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
			return fmt.Errorf("check-chunk-verification failed: %s", stderr.String())
		}

		return fmt.Errorf("check-chunk-verification failed: missing chunk references in instruction files")
	}

	return nil
}
