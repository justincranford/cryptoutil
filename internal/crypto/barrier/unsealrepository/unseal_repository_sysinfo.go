package unsealrepository

import (
	"fmt"

	cryptoutilSysinfo "cryptoutil/internal/util/sysinfo"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

const fingerprintLeeway = 1

type UnsealRepositoryFromSysInfo struct {
	unsealJwks []joseJwk.Key
}

func (u *UnsealRepositoryFromSysInfo) EncryptKey(clearRootKey joseJwk.Key) ([]byte, error) {
	return encryptKey(u.unsealJwks, clearRootKey)
}

func (u *UnsealRepositoryFromSysInfo) DecryptKey(encryptedRootKeyBytes []byte) (joseJwk.Key, error) {
	return decryptKey(u.unsealJwks, encryptedRootKeyBytes)
}

func (u *UnsealRepositoryFromSysInfo) Shutdown() {
	u.unsealJwks = nil
}

func NewUnsealRepositoryFromSysInfo(sysInfoProvider cryptoutilSysinfo.SysInfoProvider) (UnsealRepository, error) {
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
	return &UnsealRepositoryFromSysInfo{unsealJwks: unsealJwks}, nil
}
