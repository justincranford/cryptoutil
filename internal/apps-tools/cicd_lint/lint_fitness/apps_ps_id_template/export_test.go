// Copyright (c) 2025 Justin Cranford

package apps_ps_id_template

import cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"

// ExportedCheckInDirNoExclusions calls checkInDirWithExclusions with empty exclusion sets
// so tests can exercise error-return paths that are unreachable when exclusions are active.
func ExportedCheckInDirNoExclusions(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDirWithExclusions(logger, rootDir, map[string]map[string]bool{}, map[string]map[string]bool{})
}
