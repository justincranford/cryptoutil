package unsealkeysservice

import (
	"fmt"
	"time"

	cryptoutilSysinfo "cryptoutil/internal/common/util/sysinfo"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

const fingerprintLeeway = 1

type UnsealKeysServiceFromSysInfo struct {
	unsealJwks []joseJwk.Key
}

func (u *UnsealKeysServiceFromSysInfo) EncryptKey(clearJwk joseJwk.Key) ([]byte, error) {
	return encryptKey(u.unsealJwks, clearJwk)
}

func (u *UnsealKeysServiceFromSysInfo) DecryptKey(encryptedJwkBytes []byte) (joseJwk.Key, error) {
	return decryptKey(u.unsealJwks, encryptedJwkBytes)
}

func (u *UnsealKeysServiceFromSysInfo) EncryptData(clearData []byte) ([]byte, error) {
	return encryptData(u.unsealJwks, clearData)
}

func (u *UnsealKeysServiceFromSysInfo) DecryptData(encryptedDataBytes []byte) ([]byte, error) {
	return decryptData(u.unsealJwks, encryptedDataBytes)
}

func (u *UnsealKeysServiceFromSysInfo) Shutdown() {
	u.unsealJwks = nil
}

func NewUnsealKeysServiceFromSysInfo(sysInfoProvider cryptoutilSysinfo.SysInfoProvider) (UnsealKeysService, error) {
	sysinfos, err := cryptoutilSysinfo.GetAllInfoWithTimeout(sysInfoProvider, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to get sysinfo: %w", err)
	}

	numSysinfos := len(sysinfos)
	if numSysinfos == 0 {
		return nil, fmt.Errorf("empty sysinfos not supported")
	}

	var chooseN int
	if numSysinfos == 1 {
		chooseN = 1 // use it as-is
	} else {
		chooseN = numSysinfos - fingerprintLeeway // use combinations of M choose M-1
	}

	unsealJwks, err := deriveJwksFromMChooseNCombinations(sysinfos, chooseN)
	if err != nil {
		return nil, fmt.Errorf("failed to create unseal JWKs: %w", err)
	}
	return &UnsealKeysServiceFromSysInfo{unsealJwks: unsealJwks}, nil
}
