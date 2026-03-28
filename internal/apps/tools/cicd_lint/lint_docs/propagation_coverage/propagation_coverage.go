// Copyright (c) 2025 Justin Cranford

// Package propagation_coverage reports @source block line and file coverage
// across instruction and agent files.
package propagation_coverage

import (
	"bytes"
	"fmt"
	"io"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/tools/cicd_lint/common"
	cryptoutilDocsValidation "cryptoutil/internal/apps/tools/cicd_lint/docs_validation"
)

// propagationCoverageFn is the seam for testing, replacing PropagationCoverageCommand.
var propagationCoverageFn = func(stdout, stderr io.Writer) int {
	return cryptoutilDocsValidation.PropagationCoverageCommand(stdout, stderr)
}

// Check reports @source block coverage metrics for instruction and agent files.
func Check(logger *cryptoutilCmdCicdCommon.Logger) error {
	var stdout, stderr bytes.Buffer

	exitCode := propagationCoverageFn(&stdout, &stderr)

	if stdout.Len() > 0 {
		logger.Log(stdout.String())
	}

	if exitCode != 0 {
		if stderr.Len() > 0 {
			return fmt.Errorf("propagation-coverage failed: %s", stderr.String())
		}

		return fmt.Errorf("propagation-coverage failed: coverage computation error")
	}

	return nil
}
