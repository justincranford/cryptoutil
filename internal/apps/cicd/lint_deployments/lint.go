// Copyright (c) 2025 Justin Cranford

package lint_deployments

import (
	"fmt"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
)

// Lint runs all 8 deployment validators sequentially with aggregated error reporting.
// Uses the default deployments/ and configs/ directories.
// Returns an error if any validator fails.
func Lint(logger *cryptoutilCmdCicdCommon.Logger) error {
	logger.Log(fmt.Sprintf("Validating deployments dir=%s configs dir=%s", defaultDeploymentsDir, defaultConfigsDir))

	result := ValidateAll(defaultDeploymentsDir, defaultConfigsDir)

	logger.Log(FormatAllValidationResult(result))

	if !result.AllPassed() {
		return fmt.Errorf("lint-deployments failed: %d of %d validators failed (duration: %s)",
			countFailed(result), len(result.Results), result.TotalDuration)
	}

	logger.Log(fmt.Sprintf("  \u2705 All %d validators passed (duration: %s)", len(result.Results), result.TotalDuration))

	return nil
}

// countFailed returns the number of validators that did not pass.
func countFailed(result *AllValidationResult) int {
	count := 0

	for i := range result.Results {
		if !result.Results[i].Passed {
			count++
		}
	}

	return count
}
