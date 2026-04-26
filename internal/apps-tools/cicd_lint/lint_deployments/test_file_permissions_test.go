package lint_deployments

import (
	"os"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	dirPermissions  = os.FileMode(cryptoutilSharedMagic.CICDTempDirPermissions)
	filePermissions = os.FileMode(cryptoutilSharedMagic.FilePermissions)
)
