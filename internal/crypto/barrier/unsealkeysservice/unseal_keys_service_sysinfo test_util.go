package unsealkeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/apperr"
	"cryptoutil/internal/util/sysinfo"
)

func RequireNewFromSysInfoForTest() UnsealKeysService {
	unsealKeysService, err := NewUnsealKeysServiceFromSysInfo(&sysinfo.DefaultSysInfoProvider{})
	cryptoutilAppErr.RequireNoError(err, "failed to create unseal repository")

	return unsealKeysService
}
