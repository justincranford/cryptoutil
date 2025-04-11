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

func computeCombinationsAndDeriveJwks(m [][]byte, n int) ([]jwk.Key, error) {
	combinations, err := combinations.ComputeCombinations(m, n)
	if err != nil {
		return nil, fmt.Errorf("failed to compute %d of %d combinations of shared secrets: %w", len(m), n, err)
	} else if len(combinations) == 0 {
		return nil, fmt.Errorf("no combinations")
	}

	unsealJwks := make([]jwk.Key, 0, len(combinations))
	for _, combo := range combinations {
		var comboBytes []byte
		for _, key := range combo {
			comboBytes = append(comboBytes, key...)
		}

		secret := digests.SHA512(append(comboBytes, []byte("secret")...))
		salt := digests.SHA512(append(comboBytes, []byte("salt")...))
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
