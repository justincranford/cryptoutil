// Copyright (c) 2025 Justin Cranford

// Package check_chunk_verification verifies that ARCHITECTURE.md section chunks
// are present in their target instruction files.
package check_chunk_verification

import (
	"bytes"
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilDocsValidation "cryptoutil/internal/apps/cicd/docs_validation"
)

// Check verifies that all architecture chunk references exist in instruction files.
// Returns an error if any chunk references are missing.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	var stdout, stderr bytes.Buffer

	exitCode := cryptoutilDocsValidation.CheckChunkVerification(&stdout, &stderr)

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
