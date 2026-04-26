// Copyright (c) 2025 Justin Cranford
//
//

package unsealkeysservice

import (
	"fmt"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// UnsealKeysServiceSimple implements UnsealKeysService using pre-generated JWKs.
type UnsealKeysServiceSimple struct {
	unsealJWKs []joseJwk.Key
}

// EncryptKey encrypts a JWK with the unseal keys.
func (u *UnsealKeysServiceSimple) EncryptKey(clearJWK joseJwk.Key) ([]byte, error) {
	return encryptKey(u.unsealJWKs, clearJWK)
}

// DecryptKey decrypts a JWK encrypted with the unseal keys.
func (u *UnsealKeysServiceSimple) DecryptKey(encryptedJWKBytes []byte) (joseJwk.Key, error) {
	return decryptKey(u.unsealJWKs, encryptedJWKBytes)
}

// EncryptData encrypts data bytes with the unseal keys.
func (u *UnsealKeysServiceSimple) EncryptData(clearData []byte) ([]byte, error) {
	return encryptData(u.unsealJWKs, clearData)
}

// DecryptData decrypts data bytes encrypted with the unseal keys.
func (u *UnsealKeysServiceSimple) DecryptData(encryptedDataBytes []byte) ([]byte, error) {
	return decryptData(u.unsealJWKs, encryptedDataBytes)
}

// Shutdown releases all resources held by the UnsealKeysServiceSimple.
func (u *UnsealKeysServiceSimple) Shutdown() {
	u.unsealJWKs = nil
}

// NewUnsealKeysServiceSimple creates a new UnsealKeysService using pre-generated JWKs.
func NewUnsealKeysServiceSimple(unsealJWKs []joseJwk.Key) (UnsealKeysService, error) {
	if unsealJWKs == nil {
		return nil, fmt.Errorf("unsealJWKs can't be nil")
	} else if len(unsealJWKs) == 0 {
		return nil, fmt.Errorf("unsealJWKs can't be empty")
	}

	return &UnsealKeysServiceSimple{unsealJWKs: unsealJWKs}, nil
}
