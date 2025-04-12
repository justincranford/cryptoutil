package unsealrepository

import (
	"cryptoutil/internal/crypto/digests"
	"cryptoutil/internal/crypto/jose"
	"cryptoutil/internal/util/combinations"
	"fmt"

	"github.com/lestrrat-go/jwx/v3/jwk"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type UnsealKeyRepository interface {
	UnsealJwks() []joseJwk.Key
}

func computeCombinationsAndDeriveJwks(m [][]byte, chooseN int) ([]jwk.Key, error) {
	combinations, err := combinations.ComputeCombinations(m, chooseN)
	if err != nil {
		return nil, fmt.Errorf("failed to compute %d of %d combinations of shared secrets: %w", len(m), chooseN, err)
	} else if len(combinations) == 0 {
		return nil, fmt.Errorf("no combinations")
	}

	unsealJwks := make([]jwk.Key, 0, len(combinations))
	for _, combination := range combinations {
		var concatenatedCombinationBytes []byte
		for _, combinationElement := range combination {
			concatenatedCombinationBytes = append(concatenatedCombinationBytes, combinationElement...)
		}

		secret := digests.SHA512(append(concatenatedCombinationBytes, []byte("secret")...))
		salt := digests.SHA512(append(concatenatedCombinationBytes, []byte("salt")...))
		derivedKey, err := digests.HKDFwithSHA256(secret, salt, []byte("derive unseal JWKs v1"), 32)
		if err != nil {
			return nil, fmt.Errorf("failed to derive key: %w", err)
		}

		jwk, _, err := jose.CreateAesJWK(jose.AlgA256GCMKW, derivedKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create JWK: %w", err)
		}

		unsealJwks = append(unsealJwks, jwk)
	}

	return unsealJwks, nil
}
