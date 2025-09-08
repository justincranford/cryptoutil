package unsealkeysservice

import (
	"fmt"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type UnsealKeysServiceSimple struct {
	unsealJwks []joseJwk.Key
}

func (u *UnsealKeysServiceSimple) EncryptKey(clearJwk joseJwk.Key) ([]byte, error) {
	return encryptKey(u.unsealJwks, clearJwk)
}

func (u *UnsealKeysServiceSimple) DecryptKey(encryptedJwkBytes []byte) (joseJwk.Key, error) {
	return decryptKey(u.unsealJwks, encryptedJwkBytes)
}

func (u *UnsealKeysServiceSimple) EncryptData(clearData []byte) ([]byte, error) {
	return encryptData(u.unsealJwks, clearData)
}

func (u *UnsealKeysServiceSimple) DecryptData(encryptedDataBytes []byte) ([]byte, error) {
	return decryptData(u.unsealJwks, encryptedDataBytes)
}

func (u *UnsealKeysServiceSimple) Shutdown() {
	u.unsealJwks = nil
}

func NewUnsealKeysServiceSimple(unsealJwks []joseJwk.Key) (UnsealKeysService, error) {
	if unsealJwks == nil {
		return nil, fmt.Errorf("unsealJwks can't be nil")
	} else if len(unsealJwks) == 0 {
		return nil, fmt.Errorf("unsealJwks can't be empty")
	}
	return &UnsealKeysServiceSimple{unsealJwks: unsealJwks}, nil
}
