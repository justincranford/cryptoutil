package unsealrepository

import (
	"fmt"

	cryptoutilDigests "cryptoutil/internal/crypto/digests"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilCombinations "cryptoutil/internal/util/combinations"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type UnsealRepository interface {
	UnsealJwks() []joseJwk.Key
}

func deriveJwksFromMChooseNCombinations(m [][]byte, chooseN int) ([]joseJwk.Key, error) {
	combinations, err := cryptoutilCombinations.ComputeCombinations(m, chooseN)
	if err != nil {
		return nil, fmt.Errorf("failed to compute %d of %d combinations of shared secrets: %w", len(m), chooseN, err)
	} else if len(combinations) == 0 {
		return nil, fmt.Errorf("no combinations")
	}

	fixedContextBytes := []byte("derive unseal JWKs v1")
	unsealJwks := make([]joseJwk.Key, 0, len(combinations))
	for _, combination := range combinations {
		var concatenatedCombinationBytes []byte
		for _, combinationElement := range combination {
			concatenatedCombinationBytes = append(concatenatedCombinationBytes, combinationElement...)
		}

		derivedSecretKeyBytes := cryptoutilDigests.SHA512(append(concatenatedCombinationBytes, []byte("secret")...))
		derivedSaltBytes := cryptoutilDigests.SHA512(append(concatenatedCombinationBytes, []byte("salt")...))
		derivedKeyBytes, err := cryptoutilDigests.HKDFwithSHA256(derivedSecretKeyBytes, derivedSaltBytes, fixedContextBytes, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to derive key: %w", err)
		}

		jwk, _, err := cryptoutilJose.CreateAesJWK(cryptoutilJose.AlgA256GCMKW, derivedKeyBytes) // use derived JWK for envelope encryption (i.e. AES256GCM Key Wrap), not DIRECT encryption
		if err != nil {
			return nil, fmt.Errorf("failed to create JWK: %w", err)
		}

		unsealJwks = append(unsealJwks, jwk)
	}

	return unsealJwks, nil
}
