package unsealkeysservice

import (
	"fmt"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

const (
	// Maximum number of shared secrets allowed.
	maxSharedSecrets = 256

	// Minimum shared secret length in bytes.
	minSharedSecretLength = 32

	// Maximum shared secret length in bytes.
	maxSharedSecretLength = 64
)

type UnsealKeysServiceSharedSecrets struct {
	unsealJWKs []joseJwk.Key
}

func (u *UnsealKeysServiceSharedSecrets) EncryptKey(clearJWK joseJwk.Key) ([]byte, error) {
	return encryptKey(u.unsealJWKs, clearJWK)
}

func (u *UnsealKeysServiceSharedSecrets) DecryptKey(encryptedJWKBytes []byte) (joseJwk.Key, error) {
	return decryptKey(u.unsealJWKs, encryptedJWKBytes)
}

func (u *UnsealKeysServiceSharedSecrets) EncryptData(clearData []byte) ([]byte, error) {
	return encryptData(u.unsealJWKs, clearData)
}

func (u *UnsealKeysServiceSharedSecrets) DecryptData(encryptedDataBytes []byte) ([]byte, error) {
	return decryptData(u.unsealJWKs, encryptedDataBytes)
}

func (u *UnsealKeysServiceSharedSecrets) Shutdown() {
	u.unsealJWKs = nil
}

func NewUnsealKeysServiceSharedSecrets(sharedSecretsM [][]byte, chooseN int) (UnsealKeysService, error) {
	if sharedSecretsM == nil {
		return nil, fmt.Errorf("shared secrets can't be nil")
	}

	countM := len(sharedSecretsM)
	if countM == 0 {
		return nil, fmt.Errorf("shared secrets can't be zero")
	} else if countM >= maxSharedSecrets {
		return nil, fmt.Errorf("shared secrets can't be greater than %d", maxSharedSecrets)
	} else if chooseN == 0 {
		return nil, fmt.Errorf("n can't be zero")
	} else if chooseN < 0 {
		return nil, fmt.Errorf("n can't be negative")
	} else if chooseN > countM {
		return nil, fmt.Errorf("n can't be greater than shared secrets count")
	}

	for i, sharedSecret := range sharedSecretsM {
		if sharedSecret == nil {
			return nil, fmt.Errorf("shared secret %d can't be nil", i)
		} else if len(sharedSecret) < minSharedSecretLength {
			return nil, fmt.Errorf("shared secret %d length can't be less than %d", i, minSharedSecretLength)
		} else if len(sharedSecret) > maxSharedSecretLength {
			return nil, fmt.Errorf("shared secret %d length can't be greater than %d", i, maxSharedSecretLength)
		}
	}

	unsealJWKs, err := deriveJWKsFromMChooseNCombinations(sharedSecretsM, chooseN)
	if err != nil {
		return nil, fmt.Errorf("failed to create unseal JWK combinations: %w", err)
	}

	return &UnsealKeysServiceSharedSecrets{unsealJWKs: unsealJWKs}, nil
}
