package orm

import (
	"time"

	googleUuid "github.com/google/uuid"
)

func BuildKeyPool(keyPoolID googleUuid.UUID, name, description, provider, algorithm string, versioningAllowed, importAllowed, exportAllowed bool, status string) (*KeyPool, error) {
	keyPool := KeyPool{
		KeyPoolID:                  keyPoolID,
		KeyPoolName:                name,
		KeyPoolDescription:         description,
		KeyPoolProvider:            KeyPoolProvider(provider),
		KeyPoolAlgorithm:           KeyPoolAlgorithm(algorithm),
		KeyPoolIsVersioningAllowed: versioningAllowed,
		KeyPoolIsImportAllowed:     importAllowed,
		KeyPoolIsExportAllowed:     exportAllowed,
		KeyPoolStatus:              KeyPoolStatus(status),
	}
	return &keyPool, nil
}

func BuildKey(keyPoolID googleUuid.UUID, keyID int, keyMaterial []byte, generateDate, importDate, expirationDate, revocationDate *time.Time) *Key {
	key := Key{
		KeyPoolID:         keyPoolID,
		KeyID:             keyID,
		KeyMaterial:       keyMaterial,
		KeyGenerateDate:   generateDate,
		KeyImportDate:     importDate,
		KeyExpirationDate: expirationDate,
		KeyRevocationDate: revocationDate,
	}
	return &key
}

func BuildKeyPoolCreate(name, description, provider, algorithm string, versioningAllowed, importAllowed, exportAllowed bool) (*KeyPoolCreate, error) {
	keyPoolCreate := KeyPoolCreate{
		Name:                KeyPoolName(name),
		Description:         KeyPoolDescription(description),
		Provider:            KeyPoolProvider(provider),
		Algorithm:           KeyPoolAlgorithm(algorithm),
		IsVersioningAllowed: KeyPoolIsVersioningAllowed(versioningAllowed),
		IsImportAllowed:     KeyPoolIsImportAllowed(importAllowed),
		IsExportAllowed:     KeyPoolIsExportAllowed(exportAllowed),
	}
	return &keyPoolCreate, nil
}

func KeyPoolStatusInitial(importAllowed bool) string {
	if importAllowed {
		return string(PendingImport)
	}
	return string(PendingGenerate)
}
