package unsealrepository

import (
	cryptoutilAppErr "cryptoutil/internal/apperr"
	"cryptoutil/internal/util/sysinfo"
)

func RequireNewForTest() UnsealRepository {
	unsealRepository, err := NewUnsealRepositoryFromSysInfo(&sysinfo.DefaultSysInfoProvider{})
	cryptoutilAppErr.RequireNoError(err, "failed to create unseal repository")

	return unsealRepository
}
