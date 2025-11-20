// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"cryptoutil/internal/cmd/cicd/check/circulardeps"
	"cryptoutil/internal/cmd/cicd/common"
)

// goCheckCircularPackageDeps checks for circular package dependencies.
// Thin wrapper that delegates to circulardeps.Check().
func goCheckCircularPackageDeps(logger *common.Logger) error {
	return circulardeps.Check(logger) //nolint:wrapcheck // Thin wrapper - error already contains full context
}
