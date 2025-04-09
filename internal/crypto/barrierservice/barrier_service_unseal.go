package barrierservice

import (
	cryptoutilDigests "cryptoutil/internal/crypto/digests"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilSysinfo "cryptoutil/internal/sysinfo"
	cryptoutilUtil "cryptoutil/internal/util"
	"fmt"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

func UnsealJwkSet() (joseJwk.Set, error) {
	sysinfos, err := cryptoutilSysinfo.GetAllInfo(&cryptoutilSysinfo.DefaultSysInfoProvider{})
	if err != nil {
		return nil, fmt.Errorf("failed to get sysinfo: %w", err)
	}

	sysinfosBytes := cryptoutilUtil.ConcatBytes(sysinfos)
	derivedKeyBytes := cryptoutilDigests.Sha256(sysinfosBytes)
	rootJwk, _, err := cryptoutilJose.CreateAesJWK(cryptoutilJose.AlgA256GCMKW, derivedKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWK: %w", err)
	}

	set := joseJwk.NewSet()
	err = set.AddKey(rootJwk)
	if err != nil {
		return nil, fmt.Errorf("failed to create root JWK: %w", err)
	}
	return set, nil
}
