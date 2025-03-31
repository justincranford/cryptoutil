package orm

import (
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
)

func GivenKeyPoolCreate(versioningAllowed, importAllowed, exportAllowed bool) (*KeyPoolCreate, error) {
	uuid, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}
	keyPool := &KeyPoolCreate{
		Name:                KeyPoolName("Key Pool Name " + uuid.String()),
		Description:         KeyPoolDescription("Key Pool Description " + uuid.String()),
		Provider:            Internal,
		Algorithm:           AES256,
		IsVersioningAllowed: KeyPoolIsVersioningAllowed(versioningAllowed),
		IsImportAllowed:     KeyPoolIsImportAllowed(importAllowed),
		IsExportAllowed:     KeyPoolIsExportAllowed(exportAllowed),
	}
	return keyPool, nil
}

func GivenKeyPoolForAdd(versioningAllowed, importAllowed, exportAllowed bool) (*KeyPool, error) {
	uuid, err := googleUuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate UUID: %w", err)
	}
	keyPool := &KeyPool{
		KeyPoolID:                  uuidZero,
		KeyPoolName:                string("Key Pool Name " + uuid.String()),
		KeyPoolDescription:         string("Key Pool Description " + uuid.String()),
		KeyPoolProvider:            Internal,
		KeyPoolAlgorithm:           AES256,
		KeyPoolIsVersioningAllowed: versioningAllowed,
		KeyPoolIsImportAllowed:     importAllowed,
		KeyPoolIsExportAllowed:     exportAllowed,
		KeyPoolStatus:              Creating,
	}
	return keyPool, nil
}

func GivenKeyForAdd(keyPoolID googleUuid.UUID, keyID int, keyMaterial []byte, generateDate, importDate, expirationDate, revocationDate *time.Time) *Key {
	key := &Key{
		KeyPoolID:         keyPoolID,
		KeyID:             keyID,
		KeyMaterial:       keyMaterial,
		KeyGenerateDate:   generateDate,
		KeyImportDate:     importDate,
		KeyExpirationDate: expirationDate,
		KeyRevocationDate: revocationDate,
	}
	return key
}
