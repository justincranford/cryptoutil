// Copyright (c) 2025 Justin Cranford
//
//

package unsealkeysservice

import (
	cryptoutilAppErr "cryptoutil/internal/shared/apperr"
	cryptoutilSysinfo "cryptoutil/internal/shared/util/sysinfo"
)

// RequireNewFromSysInfoForTest creates an UnsealKeysService using MockSysInfoProvider.
// Using the mock provider avoids slow sysinfo collection (CPU info can take 4+ seconds on Windows).
func RequireNewFromSysInfoForTest() UnsealKeysService {
	mockProvider := &cryptoutilSysinfo.MockSysInfoProvider{}
	unsealKeysService, err := NewUnsealKeysServiceFromSysInfo(mockProvider)
	cryptoutilAppErr.RequireNoError(err, "failed to create unseal repository")

	return unsealKeysService
}
