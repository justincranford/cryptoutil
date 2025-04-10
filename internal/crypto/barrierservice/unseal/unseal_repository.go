package unseal

import (
	cryptoutilSysinfo "cryptoutil/internal/sysinfo"
	"fmt"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type UnsealKeyRepository struct {
	unsealJwks []joseJwk.Key
}

// TODO Support other sources of unseal JWKs (e.g. HSM, 3rd-party KMS, secret key sharing, etc)

func NewUnsealKeyRepositoryFromSysInfo(sysInfoProvider *cryptoutilSysinfo.DefaultSysInfoProvider) (*UnsealKeyRepository, error) {
	unsealJwks, unsealJwksErr := sysFingerprintUnsealJwks(sysInfoProvider)
	if unsealJwksErr != nil {
		return nil, fmt.Errorf("failed to get unseal JWKs: %w", unsealJwksErr)
	}
	return &UnsealKeyRepository{unsealJwks: unsealJwks}, nil
}

func (u *UnsealKeyRepository) UnsealJwks() []joseJwk.Key {
	return u.unsealJwks
}
