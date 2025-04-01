package orm

import (
	"context"
	"cryptoutil/internal/crypto/keygen"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	"time"

	googleUuid "github.com/google/uuid"
)

type Givens struct {
	telemetryService *cryptoutilTelemetry.Service
	aes256Pool       *keygen.KeyPool
	uuidV7Pool       *keygen.KeyPool
}

func NewGivens(ctx context.Context, telemetryService *cryptoutilTelemetry.Service) *Givens {
	aes256Pool := keygen.NewKeyPool(ctx, telemetryService, "AES", 3, 1, keygen.MaxKeys, keygen.MaxTime, keygen.GenerateAESKeyFunction(256))
	uuidV7Pool := keygen.NewKeyPool(ctx, telemetryService, "UUIDv7", 3, 1, keygen.MaxKeys, keygen.MaxTime, keygen.GenerateUUIDv7Function())
	return &Givens{telemetryService: telemetryService, aes256Pool: aes256Pool, uuidV7Pool: uuidV7Pool}
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

func (g *Givens) KeyPoolForAdd(versioningAllowed, importAllowed, exportAllowed bool) (*KeyPool, error) {
	uuid := g.UUIDv7()
	return BuildKeyPool(uuidZero, string("Key Pool Name "+uuid.String()), string("Key Pool Description "+uuid.String()), string(Internal), string(AES256), versioningAllowed, importAllowed, exportAllowed, string(Creating))
}

func (g *Givens) Key(keyPoolID googleUuid.UUID, keyID int, generateDate, importDate, expirationDate, revocationDate *time.Time) *Key {
	keyMaterial := g.AES256()
	return BuildKey(keyPoolID, keyID, keyMaterial, generateDate, importDate, expirationDate, revocationDate)
}

func (g *Givens) KeyPoolCreate(versioningAllowed, importAllowed, exportAllowed bool) (*KeyPoolCreate, error) {
	uuid := g.UUIDv7()
	return BuildKeyPoolCreate(string("Key Pool Name "+uuid.String()), string("Key Pool Description "+uuid.String()), string(Internal), string(AES256), versioningAllowed, importAllowed, exportAllowed)
}
