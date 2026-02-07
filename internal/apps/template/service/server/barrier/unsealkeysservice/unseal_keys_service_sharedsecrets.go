// Copyright (c) 2025 Justin Cranford
//
//

package unsealkeysservice

import (
	"fmt"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// UnsealKeysServiceSharedSecrets implements UnsealKeysService using M-of-N shared secrets.
type UnsealKeysServiceSharedSecrets struct {
	unsealJWKs []joseJwk.Key
}

// EncryptKey encrypts a JWK with the unseal keys.
func (u *UnsealKeysServiceSharedSecrets) EncryptKey(clearJWK joseJwk.Key) ([]byte, error) {
	return encryptKey(u.unsealJWKs, clearJWK)
}

// DecryptKey decrypts a JWK encrypted with the unseal keys.
func (u *UnsealKeysServiceSharedSecrets) DecryptKey(encryptedJWKBytes []byte) (joseJwk.Key, error) {
	return decryptKey(u.unsealJWKs, encryptedJWKBytes)
}

// EncryptData encrypts data bytes with the unseal keys.
func (u *UnsealKeysServiceSharedSecrets) EncryptData(clearData []byte) ([]byte, error) {
	return encryptData(u.unsealJWKs, clearData)
}

// DecryptData decrypts data bytes encrypted with the unseal keys.
func (u *UnsealKeysServiceSharedSecrets) DecryptData(encryptedDataBytes []byte) ([]byte, error) {
	return decryptData(u.unsealJWKs, encryptedDataBytes)
}

// Shutdown releases all resources held by the UnsealKeysServiceSharedSecrets.
func (u *UnsealKeysServiceSharedSecrets) Shutdown() {
	u.unsealJWKs = nil
}

// NewUnsealKeysServiceSharedSecrets creates a new UnsealKeysService using M-of-N shared secrets.
func NewUnsealKeysServiceSharedSecrets(sharedSecretsM [][]byte, chooseN int) (UnsealKeysService, error) { // pragma: allowlist secret
	if sharedSecretsM == nil { // pragma: allowlist secret
		return nil, fmt.Errorf("shared secrets can't be nil") // pragma: allowlist secret
	}

	countM := len(sharedSecretsM) // pragma: allowlist secret
	if countM == 0 {
		return nil, fmt.Errorf("shared secrets can't be zero") // pragma: allowlist secret
	} else if countM >= cryptoutilSharedMagic.MaxUnsealSharedSecrets { // pragma: allowlist secret
		return nil, fmt.Errorf("shared secrets can't be greater than %d", cryptoutilSharedMagic.MaxUnsealSharedSecrets) // pragma: allowlist secret
	} else if chooseN == 0 {
		return nil, fmt.Errorf("n can't be zero")
	} else if chooseN < 0 {
		return nil, fmt.Errorf("n can't be negative")
	} else if chooseN > countM {
		return nil, fmt.Errorf("n can't be greater than shared secrets count") // pragma: allowlist secret
	}

	for i, sharedSecret := range sharedSecretsM { // pragma: allowlist secret
		if sharedSecret == nil { // pragma: allowlist secret
			return nil, fmt.Errorf("shared secret %d can't be nil", i) // pragma: allowlist secret
		} else if len(sharedSecret) < cryptoutilSharedMagic.MinSharedSecretLength { // pragma: allowlist secret
			return nil, fmt.Errorf("shared secret %d length can't be less than %d", i, cryptoutilSharedMagic.MinSharedSecretLength) // pragma: allowlist secret
		} else if len(sharedSecret) > cryptoutilSharedMagic.MaxSharedSecretLength { // pragma: allowlist secret
			return nil, fmt.Errorf("shared secret %d length can't be greater than %d", i, cryptoutilSharedMagic.MaxSharedSecretLength) // pragma: allowlist secret
		}
	}

	unsealJWKs, err := deriveJWKsFromMChooseNCombinations(sharedSecretsM, chooseN) // pragma: allowlist secret
	if err != nil {
		return nil, fmt.Errorf("failed to create unseal JWK combinations: %w", err)
	}

	return &UnsealKeysServiceSharedSecrets{unsealJWKs: unsealJWKs}, nil // pragma: allowlist secret
}
