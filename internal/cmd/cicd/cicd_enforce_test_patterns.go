// Copyright (c) 2025 Justin Cranford

package cicd

import (
	"cryptoutil/internal/cmd/cicd/common"
	"cryptoutil/internal/cmd/cicd/enforce/testpatterns"
)

// goEnforceTestPatterns enforces test patterns including UUIDv7 usage and testify assertions.
func goEnforceTestPatterns(logger *common.Logger, allFiles []string) error {
	return testpatterns.Enforce(logger, allFiles)
}
