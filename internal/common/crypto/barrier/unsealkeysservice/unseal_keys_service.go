package unsealkeysservice

import (
	"fmt"

	cryptoutilDigests "cryptoutil/internal/common/crypto/digests"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilCombinations "cryptoutil/internal/common/util/combinations"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type UnsealKeysService interface {
	EncryptKey(clearRootKey joseJwk.Key) ([]byte, error)
	DecryptKey(encryptedRootKeyBytes []byte) (joseJwk.Key, error)
	Shutdown()
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

		kekKidUuid := googleUuid.Must(googleUuid.NewV7())
		_, jwk, _, err := cryptoutilJose.CreateAesJWKFromBytes(&kekKidUuid, &cryptoutilJose.AlgA256KW, &cryptoutilJose.EncA256GCM, derivedKeyBytes) // use derived JWK for envelope encryption (i.e. A256GCM Key Wrap), not DIRECT encryption
		if err != nil {
			return nil, fmt.Errorf("failed to create JWK: %w", err)
		}

		unsealJwks = append(unsealJwks, jwk)
	}

	return unsealJwks, nil
}

func encryptKey(unsealJwks []joseJwk.Key, clearRootKey joseJwk.Key) ([]byte, error) {
	_, encryptedRootKeyBytes, err := cryptoutilJose.EncryptKey(unsealJwks, clearRootKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt root JWK with unseal JWK: %w", err)
	}
	return encryptedRootKeyBytes, nil
}

func decryptKey(unsealJwks []joseJwk.Key, encryptedRootKeyBytes []byte) (joseJwk.Key, error) {
	decryptedRootKey, err := cryptoutilJose.DecryptKey(unsealJwks, []byte(encryptedRootKeyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt root JWK with unseal JWK: %w", err)
	}
	return decryptedRootKey, nil
}
