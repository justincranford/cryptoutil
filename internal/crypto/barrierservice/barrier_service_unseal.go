package barrierservice

import (
	"cryptoutil/internal/combinations"
	cryptoutilDigests "cryptoutil/internal/crypto/digests"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilSysinfo "cryptoutil/internal/sysinfo"
	"fmt"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

const fingerprintLeeway = 1

func UnsealJwks() ([]joseJwk.Key, error) {
	sysinfos, err := cryptoutilSysinfo.GetAllInfo(&cryptoutilSysinfo.DefaultSysInfoProvider{})
	if err != nil {
		return nil, fmt.Errorf("failed to get sysinfo: %w", err)
	}

	numSysinfos := len(sysinfos)
	if numSysinfos == 0 {
		return nil, fmt.Errorf("empty sysinfos not supported")
	}
	unsealJwks := make([]joseJwk.Key, 0, numSysinfos) // could be more if leeway is more than 1

	var chooseN int
	if numSysinfos == 1 {
		chooseN = numSysinfos // use it as-is
	} else {
		chooseN = numSysinfos - fingerprintLeeway // use combinations of M choose M-1
	}

	combinations, err := combinations.ComputeCombinations(sysinfos, chooseN) // M choose N combinationss
	if err != nil {
		return nil, fmt.Errorf("failed to compute %d of %d combinations of sysinfo: %w", numSysinfos, numSysinfos-1, err)
	}
	for _, combination := range combinations {
		var sysinfoCombinationBytes []byte
		for _, value := range combination {
			sysinfoCombinationBytes = append(sysinfoCombinationBytes, value...)
		}

		derivedSecretBytes := cryptoutilDigests.SHA512(fmt.Append(sysinfoCombinationBytes, []byte("secret")))
		derivedSaltBytes := cryptoutilDigests.SHA512(fmt.Append(sysinfoCombinationBytes, []byte("salt")))
		derivedUnsealKeyBytes, err := cryptoutilDigests.HKDFwithSHA256(derivedSecretBytes, derivedSaltBytes, []byte("derive unsealed JWKs algorithm v1"), 32)
		if err != nil {
			return nil, fmt.Errorf("failed to create JWK: %w", err)
		}

		unsealJwk, _, err := cryptoutilJose.CreateAesJWK(cryptoutilJose.AlgA256GCMKW, derivedUnsealKeyBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to create JWK: %w", err)
		}
		unsealJwks = append(unsealJwks, unsealJwk)
	}
	return unsealJwks, nil
}
