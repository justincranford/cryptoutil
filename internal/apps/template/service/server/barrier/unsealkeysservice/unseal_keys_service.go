// Copyright (c) 2025 Justin Cranford
//
//

package unsealkeysservice

import (
	"fmt"

	cryptoutilSharedCryptoDigests "cryptoutil/internal/shared/crypto/digests"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilCombinations "cryptoutil/internal/shared/util/combinations"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// UnsealKeysService defines the interface for unsealing root keys in the barrier hierarchy.
type UnsealKeysService interface {
	EncryptKey(clearRootKey joseJwk.Key) ([]byte, error)
	DecryptKey(encryptedRootKeyBytes []byte) (joseJwk.Key, error)
	EncryptData(clearData []byte) ([]byte, error)
	DecryptData(encryptedDataBytes []byte) ([]byte, error)
	Shutdown()
}

func deriveJWKsFromMChooseNCombinations(m [][]byte, chooseN int) ([]joseJwk.Key, error) {
	combinations, err := cryptoutilSharedUtilCombinations.ComputeCombinations(m, chooseN)
	if err != nil {
		return nil, fmt.Errorf("failed to compute %d of %d combinations of shared secrets: %w", len(m), chooseN, err)
	} else if len(combinations) == 0 {
		return nil, fmt.Errorf("no combinations")
	}

	derivedUnsealJWKs := make([]joseJwk.Key, 0, len(combinations))

	// CRITICAL: All cryptoutil instances using the same set of shared, unseal secrets MUST derive the same unseal JWKs,
	// including the KIDs and key materials, for cryptographic interoperability between instances.
	// CRITICAL: If only key materials were derived deterministically, using different KIDs would break interoperability.
	for _, combination := range combinations {
		var currentCombinationBytesConcat []byte
		for _, combinationElement := range combination {
			currentCombinationBytesConcat = append(currentCombinationBytesConcat, combinationElement...)
		}

		// Derive deterministic KID UUID from current combination of shared secrets

		ikmForDerivedKid := append(append([]byte{}, cryptoutilSharedMagic.FixedIKMForDerivedKid...), currentCombinationBytesConcat...)
		saltForDerivedKid := append(append([]byte{}, cryptoutilSharedMagic.FixedSaltForDerivedKid...), currentCombinationBytesConcat...)

		derivedKidBytes, err := cryptoutilSharedCryptoDigests.HKDFwithSHA256(ikmForDerivedKid, saltForDerivedKid, cryptoutilSharedMagic.FixedContextForDerivedKid, cryptoutilSharedMagic.DerivedKeySizeBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to derive unseal JWK kid bytes: %w", err)
		}

		derivedKidUUID, err := googleUuid.FromBytes(derivedKidBytes[:cryptoutilSharedMagic.UUIDBytesLength])
		if err != nil {
			return nil, fmt.Errorf("failed to create unseal JWK kid UUID from derived kid bytes: %w", err)
		}

		// Derive deterministic key material from current combination of shared secrets

		ikmForDerivedSecret := append(append([]byte{}, cryptoutilSharedMagic.FixedIKMForDerivedSecret...), currentCombinationBytesConcat...)
		saltForDerivedSecret := append(append([]byte{}, cryptoutilSharedMagic.FixedSaltForDerivedSecret...), currentCombinationBytesConcat...)

		derivedSecretBytes, err := cryptoutilSharedCryptoDigests.HKDFwithSHA256(ikmForDerivedSecret, saltForDerivedSecret, cryptoutilSharedMagic.FixedContextForDerivedSecret, cryptoutilSharedMagic.DerivedKeySizeBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to derive unseal JWK secret bytes: %w", err)
		}

		// Create JWK from derived KID and derived key material

		// CRITICAL: Use JWK for envelope encryption (i.e. alg=A256GCMKW), not DIRECT encryption (i.e. alg=dir)
		unsealJWKEncHeader := cryptoutilSharedCryptoJose.EncA256GCM
		unsealJWKAlgHeader := cryptoutilSharedCryptoJose.AlgA256KW

		_, derivedJWK, _, _, _, err := cryptoutilSharedCryptoJose.CreateJWEJWKFromKey(&derivedKidUUID, &unsealJWKEncHeader, &unsealJWKAlgHeader, cryptoutilSharedCryptoKeygen.SecretKey(derivedSecretBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create unseal JWK from derived kid bytes and secret bytes: %w", err)
		}

		derivedUnsealJWKs = append(derivedUnsealJWKs, derivedJWK)
	}

	return derivedUnsealJWKs, nil
}

func encryptKey(unsealJWKs []joseJwk.Key, clearRootKey joseJwk.Key) ([]byte, error) {
	_, encryptedRootKeyBytes, err := cryptoutilSharedCryptoJose.EncryptKey(unsealJWKs, clearRootKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt root JWK with unseal JWK: %w", err)
	}

	return encryptedRootKeyBytes, nil
}

func decryptKey(unsealJWKs []joseJwk.Key, encryptedRootKeyBytes []byte) (joseJwk.Key, error) {
	decryptedRootKey, err := cryptoutilSharedCryptoJose.DecryptKey(unsealJWKs, encryptedRootKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt root JWK with unseal JWK: %w", err)
	}

	return decryptedRootKey, nil
}

func encryptData(unsealJWKs []joseJwk.Key, clearData []byte) ([]byte, error) {
	_, encryptedDataBytes, err := cryptoutilSharedCryptoJose.EncryptBytes(unsealJWKs, clearData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data with unseal JWK: %w", err)
	}

	return encryptedDataBytes, nil
}

func decryptData(unsealJWKs []joseJwk.Key, encryptedDataBytes []byte) ([]byte, error) {
	decryptedData, err := cryptoutilSharedCryptoJose.DecryptBytes(unsealJWKs, encryptedDataBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data with unseal JWK: %w", err)
	}

	return decryptedData, nil
}
