package unsealkeysservice

import (
	"fmt"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type UnsealKeysServiceSharedSecrets struct {
	unsealJwks []joseJwk.Key
}

func (u *UnsealKeysServiceSharedSecrets) EncryptKey(clearJwk joseJwk.Key) ([]byte, error) {
	return encryptKey(u.unsealJwks, clearJwk)
}

func (u *UnsealKeysServiceSharedSecrets) DecryptKey(encryptedJwkBytes []byte) (joseJwk.Key, error) {
	return decryptKey(u.unsealJwks, encryptedJwkBytes)
}

func (u *UnsealKeysServiceSharedSecrets) EncryptData(clearData []byte) ([]byte, error) {
	return encryptData(u.unsealJwks, clearData)
}

func (u *UnsealKeysServiceSharedSecrets) DecryptData(encryptedDataBytes []byte) ([]byte, error) {
	return decryptData(u.unsealJwks, encryptedDataBytes)
}

func (u *UnsealKeysServiceSharedSecrets) Shutdown() {
	u.unsealJwks = nil
}

func NewUnsealKeysServiceSharedSecrets(sharedSecretsM [][]byte, chooseN int) (UnsealKeysService, error) {
	if sharedSecretsM == nil {
		return nil, fmt.Errorf("shared secrets can't be nil")
	}
	countM := len(sharedSecretsM)
	if countM == 0 {
		return nil, fmt.Errorf("shared secrets can't be zero")
	} else if countM >= 256 {
		return nil, fmt.Errorf("shared secrets can't be greater than 256")
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
		} else if len(sharedSecret) < 32 {
			return nil, fmt.Errorf("shared secret %d length can't be greater than 32", i)
		} else if len(sharedSecret) > 64 {
			return nil, fmt.Errorf("shared secret %d length can't be greater than 64", i)
		}
	}

	unsealJwks, err := deriveJwksFromMChooseNCombinations(sharedSecretsM, chooseN)
	if err != nil {
		return nil, fmt.Errorf("failed to create unseal JWK combinations: %w", err)
	}
	return &UnsealKeysServiceSharedSecrets{unsealJwks: unsealJwks}, nil
}
