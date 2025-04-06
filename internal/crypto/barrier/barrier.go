package barrier

import (
	"context"
	"cryptoutil/internal/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	"cryptoutil/internal/pointer"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	"encoding/json"
	"fmt"

	googleUuid "github.com/google/uuid"

	"github.com/lestrrat-go/jwx/v3/jwe"
	"github.com/lestrrat-go/jwx/v3/jwk"
)

var (
	rootKeyCacheSize         = 1000
	intermediateKeyCacheSize = 1000
	leafKeyCacheSize         = 1000
	intermediateKeyUUID      = googleUuid.Must(googleUuid.NewV7())
)

type BarrierService struct {
	telemetryService     *cryptoutilTelemetry.Service
	ormRepository        *cryptoutilOrmRepository.RepositoryProvider
	aes256Pool           *cryptoutilKeygen.KeyPool
	rootKeyCache         *JWKCache
	intermediateKeyCache *JWKCache
	leafKeyCache         *JWKCache
}

func NewBarrierService(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, ormRepository *cryptoutilOrmRepository.RepositoryProvider) (*BarrierService, error) {
	aes256Pool := cryptoutilKeygen.NewKeyPool(ctx, telemetryService, "Crypto Service AES-256", 3, 1, cryptoutilKeygen.MaxKeys, cryptoutilKeygen.MaxTime, cryptoutilKeygen.GenerateAESKeyFunction(256))

	loadLatestLeafKey := func() (*JWKCacheEntry, error) {
		return deserilalizeLatest(ormRepository.GetRootKeyLatest())
	}
	loadLeafKey := func(uuid googleUuid.UUID) (jwk.Key, error) {
		return deserilalize(ormRepository.GetRootKeyVersioned(uuid))
	}
	storeLeafKey := func(uuid googleUuid.UUID, jwk jwk.Key, parentUuid googleUuid.UUID) error {
		serialized, err := serialize(jwk, parentUuid)
		if err != nil {
			return fmt.Errorf("failed to serialize leaf Key: %w", err)
		}
		return ormRepository.AddRootKey(&cryptoutilOrmRepository.RootKey{UUID: uuid, Serialized: *serialized, UnsealKeyUUID: parentUuid})
	}

	leafKeyCache, err := NewJWKCache(rootKeyCacheSize, loadLatestLeafKey, loadLeafKey, storeLeafKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create leaf Key cache: %w", err)
	}

	return &BarrierService{
		telemetryService: telemetryService,
		ormRepository:    ormRepository,
		aes256Pool:       aes256Pool,
		leafKeyCache:     leafKeyCache,
	}, nil
}

func (d *BarrierService) Shutdown() {
	if d.aes256Pool != nil {
		d.aes256Pool.Close()
	}
	if d.leafKeyCache != nil {
		d.leafKeyCache.Shutdown()
	}
}

func (d *BarrierService) Encrypt(clearBytes []byte) ([]byte, error) {
	leafJwk, _, err := jose.GenerateAesJWK(jose.AlgDIRECT)
	if err != nil {
		return nil, fmt.Errorf("failed to generate DEK JWK: %w", err)
	}
	kid, ok := leafJwk.KeyID()
	if !ok {
		return nil, fmt.Errorf("failed to get JWK kid")
	}
	uuid, err := googleUuid.Parse(kid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid as uuid: %w", err)
	}
	err = d.leafKeyCache.Put(uuid, leafJwk, intermediateKeyUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to put leaf JWK in cache: %w", err)
	}
	jweMessage, encodedJweMessage, err := jose.EncryptBytes(leafJwk, clearBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt clear bytes: %w", err)
	}
	jweHeaders, err := jose.JSONHeadersString(jweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWE message headers: %w", err)
	}
	d.telemetryService.Slogger.Info("Encrypted", "jwe", jweHeaders)

	return encodedJweMessage, nil
}

func (d *BarrierService) Decrypt(encodedJweMessage []byte) ([]byte, error) {
	jweMessage, err := jwe.Parse(encodedJweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message: %w", err)
	}
	var kid string
	err = jweMessage.ProtectedHeaders().Get(jwk.KeyIDKey, &kid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message kid: %w", err)
	}
	uuid, err := googleUuid.Parse(kid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid as uuid: %w", err)
	}
	aesJwk, err := d.leafKeyCache.Get(uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid as uuid: %w", err)
	}
	decryptedBytes, err := jose.DecryptBytes(aesJwk, encodedJweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with JWK %s: %w", kid, err)
	}
	return decryptedBytes, nil
}

func serialize(jwk jwk.Key, parentUuid googleUuid.UUID) (*string, error) {
	bytes, err := json.Marshal(&jwk)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal latest Key from database: %w", err)
	}
	return pointer.StringPtr(string(bytes)), nil
}

func deserilalizeLatest(barrierKey cryptoutilOrmRepository.BarrierKey, err error) (*JWKCacheEntry, error) {
	if err != nil {
		return nil, fmt.Errorf("failed to load latest Key from database: %w", err)
	}
	var jwk jwk.Key
	err = json.Unmarshal([]byte((barrierKey).GetSerialized()), &jwk)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal latest Key from database: %w", err)
	}
	return &JWKCacheEntry{key: (barrierKey).GetUUID(), value: jwk}, nil
}

func deserilalize(barrierKey cryptoutilOrmRepository.BarrierKey, err error) (jwk.Key, error) {
	if err != nil {
		return nil, fmt.Errorf("failed to load latest Key from database: %w", err)
	}
	var jwk jwk.Key
	err = json.Unmarshal([]byte((barrierKey).GetSerialized()), &jwk)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal latest Key from database: %w", err)
	}
	return jwk, nil
}
