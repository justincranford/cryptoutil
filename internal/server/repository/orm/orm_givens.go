package orm

import (
	"context"
	"time"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilKeyGenPoolTest "cryptoutil/internal/common/crypto/keygenpooltest"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"

	googleUuid "github.com/google/uuid"
)

type Givens struct {
	telemetryService *cryptoutilTelemetry.TelemetryService
	aes256KeyGenPool *cryptoutilPool.ValueGenPool[cryptoutilKeyGen.SecretKey]
	uuidV7KeyGenPool *cryptoutilPool.ValueGenPool[*googleUuid.UUID]
}

func RequireNewGivensForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService) *Givens {
	aes256KeyGenPool := cryptoutilKeyGenPoolTest.RequireNewAes256GcmGenElasticKeyForTest(telemetryService)
	uuidV7KeyGenPool := cryptoutilKeyGenPoolTest.RequireNewUuidV7GenElasticKeyForTest(telemetryService)
	return &Givens{telemetryService: telemetryService, aes256KeyGenPool: aes256KeyGenPool, uuidV7KeyGenPool: uuidV7KeyGenPool}
}

func (g *Givens) Shutdown() {
	g.uuidV7KeyGenPool.Cancel()
	g.aes256KeyGenPool.Cancel()
}

func (g *Givens) UUIDv7() googleUuid.UUID {
	return *g.uuidV7KeyGenPool.Get()
}

func (g *Givens) A256() []byte {
	return g.aes256KeyGenPool.Get()
}

func (g *Givens) Aes256ElasticKey(versioningAllowed, importAllowed, exportAllowed bool) *ElasticKey {
	uuidV7 := g.UUIDv7()
	elasticKey, err := BuildElasticKey(uuidV7, string("Elastic Key Name "+uuidV7.String()), string("Elastic Key Description "+uuidV7.String()), Internal, A256GCM_dir, versioningAllowed, importAllowed, exportAllowed, string(Creating))
	cryptoutilAppErr.RequireNoError(err, "failed to create AES 256 elastic Key")
	return elasticKey
}

func (g *Givens) Aes256Key(elasticKeyID googleUuid.UUID, generateDate, importDate, expirationDate, revocationDate *time.Time) *MaterialKey {
	key := MaterialKey{
		ElasticKeyID:                  elasticKeyID,
		MaterialKeyID:                 g.UUIDv7(),
		ClearPublicKeyMaterial:        nil,
		EncryptedNonPublicKeyMaterial: g.A256(),
		MaterialKeyGenerateDate:       generateDate,
		MaterialKeyImportDate:         importDate,
		MaterialKeyExpirationDate:     expirationDate,
		MaterialKeyRevocationDate:     revocationDate,
	}
	return &key
}
