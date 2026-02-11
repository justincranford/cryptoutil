// Copyright (c) 2025 Justin Cranford
//
//

package unsealkeysservice

import (
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedUtilSysinfo "cryptoutil/internal/shared/util/sysinfo"
)

// RequireNewFromSysInfoForTest creates an UnsealKeysService using MockSysInfoProvider.
// Using the mock provider avoids slow sysinfo collection (CPU info can take 4+ seconds on Windows).
func RequireNewFromSysInfoForTest() UnsealKeysService {
	mockProvider := &cryptoutilSharedUtilSysinfo.MockSysInfoProvider{}
	unsealKeysService, err := NewUnsealKeysServiceFromSysInfo(mockProvider)
	cryptoutilSharedApperr.RequireNoError(err, "failed to create unseal repository")

	return unsealKeysService
}
