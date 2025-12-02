// Copyright (c) 2025 Justin Cranford
//
//

package unsealkeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilSysinfo "cryptoutil/internal/common/util/sysinfo"
)

func RequireNewFromSysInfoForTest() UnsealKeysService {
	unsealKeysService, err := NewUnsealKeysServiceFromSysInfo(&cryptoutilSysinfo.DefaultSysInfoProvider{})
	cryptoutilAppErr.RequireNoError(err, "failed to create unseal repository")

	return unsealKeysService
}
