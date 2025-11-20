// Copyright (c) 2025 Justin Cranford

package cicd

import (
	"cryptoutil/internal/cmd/cicd/common"
	"cryptoutil/internal/cmd/cicd/enforce/utf8"
)

// allEnforceUtf8 enforces UTF-8 encoding without BOM for all text files.
func allEnforceUtf8(logger *common.Logger, allFiles []string) error {
	return utf8.Enforce(logger, allFiles)
}
