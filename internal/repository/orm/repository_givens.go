package orm

import (
	"context"
	cryptoutilKeyGen "cryptoutil/internal/crypto/keygen"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	"errors"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
)

type Givens struct {
	telemetryService *cryptoutilTelemetry.Service
	aes256Pool       *cryptoutilKeyGen.KeyPool
	uuidV7Pool       *cryptoutilKeyGen.KeyPool
}

func NewGivens(ctx context.Context, telemetryService *cryptoutilTelemetry.Service) (*Givens, error) {
	aes256PoolConfig, err1 := cryptoutilKeyGen.NewKeyPoolConfig(ctx, telemetryService, "Orm Givens AES256", 3, 3, cryptoutilKeyGen.MaxLifetimeKeys, cryptoutilKeyGen.MaxLifetimeDuration, cryptoutilKeyGen.GenerateAESKeyFunction(256))
	uuidV7PoolConfig, err2 := cryptoutilKeyGen.NewKeyPoolConfig(ctx, telemetryService, "Orm Givens UUIDv7", 3, 3, cryptoutilKeyGen.MaxLifetimeKeys, cryptoutilKeyGen.MaxLifetimeDuration, cryptoutilKeyGen.GenerateUUIDv7Function())
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("failed to create pool configs: %w", errors.Join(err1, err2))
	}

	aes256Pool, err1 := cryptoutilKeyGen.NewKeyPool(aes256PoolConfig)
	uuidV7Pool, err2 := cryptoutilKeyGen.NewKeyPool(uuidV7PoolConfig)
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("failed to create pools: %w", errors.Join(err1, err2))
	}
	return &Givens{telemetryService: telemetryService, aes256Pool: aes256Pool, uuidV7Pool: uuidV7Pool}, nil
}

func (g *Givens) Shutdown() {
	g.uuidV7Pool.Close()
	g.aes256Pool.Close()
}

func (g *Givens) UUIDv7() googleUuid.UUID {
	return g.uuidV7Pool.Get().Private.(googleUuid.UUID)
}

func (g *Givens) AES256() []byte {
	return g.aes256Pool.Get().Private.([]byte)
}

func (g *Givens) KeyPool(versioningAllowed, importAllowed, exportAllowed bool) (*KeyPool, error) {
	uuid := g.UUIDv7()
	return BuildKeyPool(uuid, string("Key Pool Name "+uuid.String()), string("Key Pool Description "+uuid.String()), string(Internal), string(AES256), versioningAllowed, importAllowed, exportAllowed, string(Creating))
}

func (g *Givens) Key(keyPoolID googleUuid.UUID, generateDate, importDate, expirationDate, revocationDate *time.Time) *Key {
	uuid := g.UUIDv7()
	keyMaterial := g.AES256()
	return BuildKey(keyPoolID, uuid, keyMaterial, generateDate, importDate, expirationDate, revocationDate)
}
