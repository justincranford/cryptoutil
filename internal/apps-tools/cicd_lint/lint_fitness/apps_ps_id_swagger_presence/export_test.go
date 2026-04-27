// Copyright (c) 2025 Justin Cranford

package apps_ps_id_swagger_presence

import cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"

// ExportedCheckPSIDSwaggerFiles exposes checkPSIDSwaggerFiles for white-box testing.
var ExportedCheckPSIDSwaggerFiles = checkPSIDSwaggerFiles

// ExportedCheckInDirNoExclusions calls checkInDirWithExclusions with an empty exclusion set
// so tests can exercise the error-return paths that are unreachable when all PS-IDs are excluded.
func ExportedCheckInDirNoExclusions(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDirWithExclusions(logger, rootDir, map[string]bool{})
}
