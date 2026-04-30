// Copyright (c) 2025-2026 Justin Cranford.
package apps_product_no_service_dirs

import cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"

// ExportedCheckInDirWithExclusions exposes checkInDirWithExclusions for white-box testing.
func ExportedCheckInDirWithExclusions(logger *cryptoutilCmdCicdCommon.Logger, rootDir string, exclusions map[string]bool) error {
	return checkInDirWithExclusions(logger, rootDir, exclusions)
}

// ExportedBuildProductServiceMap exposes buildProductServiceMap for white-box testing.
var ExportedBuildProductServiceMap = buildProductServiceMap

// ExportedCheckProductDir exposes checkProductDir for white-box testing.
var ExportedCheckProductDir = checkProductDir
