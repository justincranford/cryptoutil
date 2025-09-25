package unsealkeysservice

import (
	"fmt"

	cryptoutilDigests "cryptoutil/internal/common/crypto/digests"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilCombinations "cryptoutil/internal/common/util/combinations"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type UnsealKeysService interface {
	EncryptKey(clearRootKey joseJwk.Key) ([]byte, error)
	DecryptKey(encryptedRootKeyBytes []byte) (joseJwk.Key, error)
	EncryptData(clearData []byte) ([]byte, error)
	DecryptData(encryptedDataBytes []byte) ([]byte, error)
	Shutdown()
}

func deriveJWKsFromMChooseNCombinations(m [][]byte, chooseN int) ([]joseJwk.Key, error) {
	combinations, err := cryptoutilCombinations.ComputeCombinations(m, chooseN)
	if err != nil {
		return nil, fmt.Errorf("failed to compute %d of %d combinations of shared secrets: %w", len(m), chooseN, err)
	} else if len(combinations) == 0 {
		return nil, fmt.Errorf("no combinations")
	}

	fixedContextBytes := []byte("derive unseal JWKs v1")
	unsealJWKs := make([]joseJwk.Key, 0, len(combinations))
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

		kekKidUUID, err := googleUuid.NewV7()
		if err != nil {
			return nil, fmt.Errorf("failed to create UUIDv7: %w", err)
		}
		_, jwk, _, _, _, err := cryptoutilJose.CreateJweJWKFromKey(&kekKidUUID, &cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW, cryptoutilKeyGen.SecretKey(derivedKeyBytes)) // use derived JWK for envelope encryption (i.e. A256GCM Key Wrap), not DIRECT encryption
		if err != nil {
			return nil, fmt.Errorf("failed to create JWK: %w", err)
		}

		unsealJWKs = append(unsealJWKs, jwk)
	}

	return unsealJWKs, nil
}

func encryptKey(unsealJWKs []joseJwk.Key, clearRootKey joseJwk.Key) ([]byte, error) {
	_, encryptedRootKeyBytes, err := cryptoutilJose.EncryptKey(unsealJWKs, clearRootKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt root JWK with unseal JWK: %w", err)
	}
	return encryptedRootKeyBytes, nil
}

func decryptKey(unsealJWKs []joseJwk.Key, encryptedRootKeyBytes []byte) (joseJwk.Key, error) {
	decryptedRootKey, err := cryptoutilJose.DecryptKey(unsealJWKs, encryptedRootKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt root JWK with unseal JWK: %w", err)
	}
	return decryptedRootKey, nil
}

func encryptData(unsealJWKs []joseJwk.Key, clearData []byte) ([]byte, error) {
	_, encryptedDataBytes, err := cryptoutilJose.EncryptBytes(unsealJWKs, clearData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data with unseal JWK: %w", err)
	}
	return encryptedDataBytes, nil
}

func decryptData(unsealJWKs []joseJwk.Key, encryptedDataBytes []byte) ([]byte, error) {
	decryptedData, err := cryptoutilJose.DecryptBytes(unsealJWKs, encryptedDataBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data with unseal JWK: %w", err)
	}
	return decryptedData, nil
}
