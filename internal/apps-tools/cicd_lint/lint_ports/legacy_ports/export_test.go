// Copyright (c) 2025-2026 Justin Cranford.
package legacy_ports

import cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"

// CheckWithAllFn exposes checkWithAllFn for white-box testing.
func CheckWithAllFn(logger *cryptoutilCmdCicdCommon.Logger, filesByExtension map[string][]string, allFn func() []uint16) error {
	return checkWithAllFn(logger, filesByExtension, allFn)
}
