package orm

import (
	"context"
	"time"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilKeyGenPoolTest "cryptoutil/internal/common/crypto/keygenpooltest"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"

	googleUuid "github.com/google/uuid"
)

type Givens struct {
	telemetryService *cryptoutilTelemetry.TelemetryService
	aes256KeyGenPool *cryptoutilPool.ValueGenPool[cryptoutilKeygen.SecretKey]
	uuidV7KeyGenPool *cryptoutilPool.ValueGenPool[*googleUuid.UUID]
}

func RequireNewGivensForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService) *Givens {
	aes256KeyGenPool := cryptoutilKeyGenPoolTest.RequireNewAes256GcmGenKeyPoolForTest(telemetryService)
	uuidV7KeyGenPool := cryptoutilKeyGenPoolTest.RequireNewUuidV7GenKeyPoolForTest(telemetryService)
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

func (g *Givens) Aes256KeyPool(versioningAllowed, importAllowed, exportAllowed bool) *KeyPool {
	uuidV7 := g.UUIDv7()
	keyPool, err := BuildKeyPool(uuidV7, string("Key Pool Name "+uuidV7.String()), string("Key Pool Description "+uuidV7.String()), Internal, A256GCM_dir, versioningAllowed, importAllowed, exportAllowed, string(Creating))
	cryptoutilAppErr.RequireNoError(err, "failed to create AES 256 key pool")
	return keyPool
}

func (g *Givens) Aes256Key(keyPoolID googleUuid.UUID, generateDate, importDate, expirationDate, revocationDate *time.Time) *Key {
	return BuildKey(keyPoolID, g.UUIDv7(), g.A256(), generateDate, importDate, expirationDate, revocationDate)
}
