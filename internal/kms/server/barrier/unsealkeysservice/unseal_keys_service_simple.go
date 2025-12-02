// Copyright (c) 2025 Justin Cranford
//
//

package unsealkeysservice

import (
	"fmt"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type UnsealKeysServiceSimple struct {
	unsealJWKs []joseJwk.Key
}

func (u *UnsealKeysServiceSimple) EncryptKey(clearJWK joseJwk.Key) ([]byte, error) {
	return encryptKey(u.unsealJWKs, clearJWK)
}

func (u *UnsealKeysServiceSimple) DecryptKey(encryptedJWKBytes []byte) (joseJwk.Key, error) {
	return decryptKey(u.unsealJWKs, encryptedJWKBytes)
}

func (u *UnsealKeysServiceSimple) EncryptData(clearData []byte) ([]byte, error) {
	return encryptData(u.unsealJWKs, clearData)
}

func (u *UnsealKeysServiceSimple) DecryptData(encryptedDataBytes []byte) ([]byte, error) {
	return decryptData(u.unsealJWKs, encryptedDataBytes)
}

func (u *UnsealKeysServiceSimple) Shutdown() {
	u.unsealJWKs = nil
}

func NewUnsealKeysServiceSimple(unsealJWKs []joseJwk.Key) (UnsealKeysService, error) {
	if unsealJWKs == nil {
		return nil, fmt.Errorf("unsealJWKs can't be nil")
	} else if len(unsealJWKs) == 0 {
		return nil, fmt.Errorf("unsealJWKs can't be empty")
	}

	return &UnsealKeysServiceSimple{unsealJWKs: unsealJWKs}, nil
}
