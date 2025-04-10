package barrierservice

import (
	cryptoutilDigests "cryptoutil/internal/crypto/digests"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilSysinfo "cryptoutil/internal/sysinfo"
	cryptoutilUtil "cryptoutil/internal/util"
	"fmt"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func UnsealJwks() ([]joseJwk.Key, error) {
	sysinfos, err := cryptoutilSysinfo.GetAllInfo(&cryptoutilSysinfo.DefaultSysInfoProvider{})
	if err != nil {
		return nil, fmt.Errorf("failed to get sysinfo: %w", err)
	}

	sysinfosBytes := cryptoutilUtil.ConcatBytes(sysinfos)

	derivedSecretBytes := cryptoutilDigests.SHA512(fmt.Append(sysinfosBytes, []byte("secret")))
	derivedSaltBytes := cryptoutilDigests.SHA512(fmt.Append(sysinfosBytes, []byte("salt")))
	derivedUnsealKeyBytes, err := cryptoutilDigests.HKDFwithSHA256(derivedSecretBytes, derivedSaltBytes, []byte("derive unsealed JWKs algorithm v1"), 32)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWK: %w", err)
	}

	unsealJwk, _, err := cryptoutilJose.CreateAesJWK(cryptoutilJose.AlgA256GCMKW, derivedUnsealKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWK: %w", err)
	}

	unsealJwks := make([]joseJwk.Key, 0, 1)
	unsealJwks = append(unsealJwks, unsealJwk)

	return unsealJwks, nil
}
