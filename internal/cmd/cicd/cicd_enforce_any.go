// Copyright (c) 2025 Justin Cranford

package cicd

import (
	"cryptoutil/internal/cmd/cicd/common"
	enforceAny "cryptoutil/internal/cmd/cicd/enforce/any"
)

// goEnforceAny enforces custom Go source code fixes across all Go files.
// It applies automated fixes like replacing any with any.
func goEnforceAny(logger *common.Logger, allFiles []string) error {
	return enforceAny.Enforce(logger, allFiles)
}
