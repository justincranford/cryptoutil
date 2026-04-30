// Copyright (c) 2025-2026 Justin Cranford.
package apps_ps_id_template

import cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"

// ExportedCheckInDirNoExclusions calls checkInDirWithExclusions with empty exclusion sets
// so tests can exercise error-return paths that are unreachable when exclusions are active.
func ExportedCheckInDirNoExclusions(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) error {
	return checkInDirWithExclusions(logger, rootDir, psIDExclusions{
		rootFiles:   map[string]map[string]bool{},
		serverFiles: map[string]map[string]bool{},
		serverDirs:  map[string]map[string]bool{},
		configFiles: map[string]map[string]bool{},
		repoFiles:   map[string]map[string]bool{},
		repoDirs:    map[string]map[string]bool{},
		e2eFiles:    map[string]map[string]bool{},
	})
}
