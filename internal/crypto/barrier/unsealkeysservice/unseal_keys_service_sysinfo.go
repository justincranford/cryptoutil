package unsealkeysservice

import (
	"fmt"

	cryptoutilSysinfo "cryptoutil/internal/common/util/sysinfo"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

const fingerprintLeeway = 1

type UnsealKeysServiceFromSysInfo struct {
	unsealJwks []joseJwk.Key
}

func (u *UnsealKeysServiceFromSysInfo) EncryptKey(clearRootKey joseJwk.Key) ([]byte, error) {
	return encryptKey(u.unsealJwks, clearRootKey)
}

func (u *UnsealKeysServiceFromSysInfo) DecryptKey(encryptedRootKeyBytes []byte) (joseJwk.Key, error) {
	return decryptKey(u.unsealJwks, encryptedRootKeyBytes)
}

func (u *UnsealKeysServiceFromSysInfo) Shutdown() {
	u.unsealJwks = nil
}

func NewUnsealKeysServiceFromSysInfo(sysInfoProvider cryptoutilSysinfo.SysInfoProvider) (UnsealKeysService, error) {
	sysinfos, err := cryptoutilSysinfo.GetAllInfo(sysInfoProvider)
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
