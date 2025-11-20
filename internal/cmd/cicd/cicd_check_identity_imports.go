// Copyright (c) 2025 Justin Cranford
//
//

package cicd

import (
	"cryptoutil/internal/cmd/cicd/check/identityimports"
	"cryptoutil/internal/cmd/cicd/common"
)

// goCheckIdentityImports checks identity module imports for domain isolation violations.
// Thin wrapper that delegates to identityimports.Check().
func goCheckIdentityImports(logger *common.Logger) error {
	return identityimports.Check(logger) //nolint:wrapcheck // Thin wrapper - error already contains full context
}
