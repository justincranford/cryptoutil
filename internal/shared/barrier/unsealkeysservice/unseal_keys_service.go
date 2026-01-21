// Copyright (c) 2025 Justin Cranford
//
//

package unsealkeysservice

import (
	"fmt"

	cryptoutilDigests "cryptoutil/internal/shared/crypto/digests"
	cryptoutilKeyGen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilCombinations "cryptoutil/internal/shared/util/combinations"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"

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
	combinations, err := cryptoutilCombinations.ComputeCombinations(m, chooseN)
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

		ikmForDerivedKid := append(append([]byte{}, cryptoutilMagic.FixedIKMForDerivedKid...), currentCombinationBytesConcat...)
		saltForDerivedKid := append(append([]byte{}, cryptoutilMagic.FixedSaltForDerivedKid...), currentCombinationBytesConcat...)

		derivedKidBytes, err := cryptoutilDigests.HKDFwithSHA256(ikmForDerivedKid, saltForDerivedKid, cryptoutilMagic.FixedContextForDerivedKid, cryptoutilMagic.DerivedKeySizeBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to derive unseal JWK kid bytes: %w", err)
		}

		derivedKidUUID, err := googleUuid.FromBytes(derivedKidBytes[:cryptoutilMagic.UUIDBytesLength])
		if err != nil {
			return nil, fmt.Errorf("failed to create unseal JWK kid UUID from derived kid bytes: %w", err)
		}

		// Derive deterministic key material from current combination of shared secrets

		ikmForDerivedSecret := append(append([]byte{}, cryptoutilMagic.FixedIKMForDerivedSecret...), currentCombinationBytesConcat...)
		saltForDerivedSecret := append(append([]byte{}, cryptoutilMagic.FixedSaltForDerivedSecret...), currentCombinationBytesConcat...)

		derivedSecretBytes, err := cryptoutilDigests.HKDFwithSHA256(ikmForDerivedSecret, saltForDerivedSecret, cryptoutilMagic.FixedContextForDerivedSecret, cryptoutilMagic.DerivedKeySizeBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to derive unseal JWK secret bytes: %w", err)
		}

		// Create JWK from derived KID and derived key material

		// CRITICAL: Use JWK for envelope encryption (i.e. alg=A256GCMKW), not DIRECT encryption (i.e. alg=dir)
		unsealJWKEncHeader := cryptoutilJose.EncA256GCM
		unsealJWKAlgHeader := cryptoutilJose.AlgA256KW

		_, derivedJWK, _, _, _, err := cryptoutilJose.CreateJWEJWKFromKey(&derivedKidUUID, &unsealJWKEncHeader, &unsealJWKAlgHeader, cryptoutilKeyGen.SecretKey(derivedSecretBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create unseal JWK from derived kid bytes and secret bytes: %w", err)
		}

		derivedUnsealJWKs = append(derivedUnsealJWKs, derivedJWK)
	}

	return derivedUnsealJWKs, nil
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
