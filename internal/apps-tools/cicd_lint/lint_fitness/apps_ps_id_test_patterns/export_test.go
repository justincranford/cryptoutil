// Copyright (c) 2025-2026 Justin Cranford.
package apps_ps_id_test_patterns

import cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"

// ExportedCheckInDirWithExclusions exposes checkInDirWithExclusions for white-box testing.
// Tests can inject custom exclusion sets to exercise paths that are unreachable when all PS-IDs
// are in the production exclusion maps.
func ExportedCheckInDirWithExclusions(
	logger *cryptoutilCmdCicdCommon.Logger,
	rootDir string,
	exclTestMain, exclLifecycle, exclPortConflict map[string]bool,
) error {
	return checkInDirWithExclusions(logger, rootDir, exclTestMain, exclLifecycle, exclPortConflict)
}

// ExportedCheckPSIDTestPatterns exposes checkPSIDTestPatterns for white-box testing.
var ExportedCheckPSIDTestPatterns = checkPSIDTestPatterns

// ExportedFileExists exposes fileExists for white-box testing.
var ExportedFileExists = fileExists

// ExportedGlobExists exposes globExists for white-box testing.
var ExportedGlobExists = globExists
