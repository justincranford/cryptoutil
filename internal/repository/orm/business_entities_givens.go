package orm

import (
	"context"
	"time"

	cryptoutilAppErr "cryptoutil/internal/apperr"
	cryptoutilKeyGen "cryptoutil/internal/crypto/keygen"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"
)

type Givens struct {
	telemetryService *cryptoutilTelemetry.TelemetryService
	aes256KeyGenPool *cryptoutilKeyGen.KeyGenPool
	uuidV7KeyGenPool *cryptoutilKeyGen.KeyGenPool
}

func RequireNewGivensForTest(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService) *Givens {
	aes256KeyGenPoolConfig, err := cryptoutilKeyGen.NewKeyGenPoolConfig(ctx, telemetryService, "Orm Givens AES256", 3, 3, cryptoutilKeyGen.MaxLifetimeKeys, cryptoutilKeyGen.MaxLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(256))
	cryptoutilAppErr.RequireNoError(err, "failed to create AES 256 pool config")

	aes256KeyGenPool, err := cryptoutilKeyGen.NewGenKeyPool(aes256KeyGenPoolConfig)
	cryptoutilAppErr.RequireNoError(err, "failed to create AES 256 pool")

	uuidV7KeyGenPoolConfig, err := cryptoutilKeyGen.NewKeyGenPoolConfig(ctx, telemetryService, "Orm Givens UUIDv7", 3, 3, cryptoutilKeyGen.MaxLifetimeKeys, cryptoutilKeyGen.MaxLifetimeDuration, cryptoutilKeyGen.GenerateUUIDv7Function())
	cryptoutilAppErr.RequireNoError(err, "failed to create UUID v7 pool config")

	uuidV7KeyGenPool, err := cryptoutilKeyGen.NewGenKeyPool(uuidV7KeyGenPoolConfig)
	cryptoutilAppErr.RequireNoError(err, "failed to create UUID v7 pool")
	return &Givens{telemetryService: telemetryService, aes256KeyGenPool: aes256KeyGenPool, uuidV7KeyGenPool: uuidV7KeyGenPool}
}

func (g *Givens) Shutdown() {
	g.uuidV7KeyGenPool.Close()
	g.aes256KeyGenPool.Close()
}

func (g *Givens) UUIDv7() googleUuid.UUID {
	return g.uuidV7KeyGenPool.Get().Private.(googleUuid.UUID)
}

func (g *Givens) AES256() []byte {
	return g.aes256KeyGenPool.Get().Private.([]byte)
}

func (g *Givens) Aes256KeyPool(versioningAllowed, importAllowed, exportAllowed bool) *KeyPool {
	uuid := g.UUIDv7()
	keyPool, err := BuildKeyPool(uuid, string("Key Pool Name "+uuid.String()), string("Key Pool Description "+uuid.String()), string(Internal), string(AES256), versioningAllowed, importAllowed, exportAllowed, string(Creating))
	cryptoutilAppErr.RequireNoError(err, "failed to create AES 256 key pool")
	return keyPool
}

func (g *Givens) Aes256Key(keyPoolID googleUuid.UUID, generateDate, importDate, expirationDate, revocationDate *time.Time) *Key {
	return BuildKey(keyPoolID, g.UUIDv7(), g.AES256(), generateDate, importDate, expirationDate, revocationDate)
}
