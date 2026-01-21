// Copyright (c) 2025 Justin Cranford
//
//

package unsealkeysservice

import (
	"fmt"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilSysinfo "cryptoutil/internal/shared/util/sysinfo"

	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

const fingerprintLeeway = 1

// UnsealKeysServiceFromSysInfo implements UnsealKeysService using system information fingerprinting.
type UnsealKeysServiceFromSysInfo struct {
	unsealJWKs []joseJwk.Key
}

// EncryptKey encrypts a JWK with the unseal keys derived from system information.
func (u *UnsealKeysServiceFromSysInfo) EncryptKey(clearJWK joseJwk.Key) ([]byte, error) {
	return encryptKey(u.unsealJWKs, clearJWK)
}

// DecryptKey decrypts a JWK encrypted with the unseal keys derived from system information.
func (u *UnsealKeysServiceFromSysInfo) DecryptKey(encryptedJWKBytes []byte) (joseJwk.Key, error) {
	return decryptKey(u.unsealJWKs, encryptedJWKBytes)
}

// EncryptData encrypts data bytes with the unseal keys derived from system information.
func (u *UnsealKeysServiceFromSysInfo) EncryptData(clearData []byte) ([]byte, error) {
	return encryptData(u.unsealJWKs, clearData)
}

// DecryptData decrypts data bytes encrypted with the unseal keys derived from system information.
func (u *UnsealKeysServiceFromSysInfo) DecryptData(encryptedDataBytes []byte) ([]byte, error) {
	return decryptData(u.unsealJWKs, encryptedDataBytes)
}

// Shutdown releases all resources held by the UnsealKeysServiceFromSysInfo.
func (u *UnsealKeysServiceFromSysInfo) Shutdown() {
	u.unsealJWKs = nil
}

// NewUnsealKeysServiceFromSysInfo creates a new UnsealKeysService using system information fingerprinting.
func NewUnsealKeysServiceFromSysInfo(sysInfoProvider cryptoutilSysinfo.SysInfoProvider) (UnsealKeysService, error) {
	sysinfos, err := cryptoutilSysinfo.GetAllInfoWithTimeout(sysInfoProvider, cryptoutilMagic.DefaultSysInfoAllTimeout)
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

	unsealJWKs, err := deriveJWKsFromMChooseNCombinations(sysinfos, chooseN)
	if err != nil {
		return nil, fmt.Errorf("failed to create unseal JWKs: %w", err)
	}

	return &UnsealKeysServiceFromSysInfo{unsealJWKs: unsealJWKs}, nil
}
