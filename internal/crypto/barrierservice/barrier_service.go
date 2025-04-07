package barrier

import (
	"context"
	cryptoutilBarrierCache "cryptoutil/internal/crypto/barriercache"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	"encoding/json"
	"errors"
	"fmt"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

var (
	rootJwk, _, _ = cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgDIRECT)
	// rootKeyCacheSize         = 1000
	intermediateKeyCacheSize = 1000
	leafKeyCacheSize         = 1000
)

type BarrierService struct {
	telemetryService *cryptoutilTelemetry.Service
	ormRepository    *cryptoutilOrmRepository.RepositoryProvider
	aes256Pool       *cryptoutilKeygen.KeyPool
	// rootKeyCache         *cryptoutilBarrierCache.Cache
	intermediateKeyCache *cryptoutilBarrierCache.Cache
	leafKeyCache         *cryptoutilBarrierCache.Cache
}

func NewBarrierService(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, ormRepository *cryptoutilOrmRepository.RepositoryProvider) (*BarrierService, error) {
	aes256Pool := cryptoutilKeygen.NewKeyPool(ctx, telemetryService, "Crypto Service AES-256", 3, 1, cryptoutilKeygen.MaxKeys, cryptoutilKeygen.MaxTime, cryptoutilKeygen.GenerateAESKeyFunction(256))

	intermediateKeyCache, err1 := newIntermediateKeyCache(ormRepository, telemetryService)
	leafKeyCache, err2 := newLeafKeyCache(ormRepository, telemetryService)
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("failed to initialize JWK caches: %w", errors.Join(err1, err2))
	}

	return &BarrierService{
		telemetryService:     telemetryService,
		ormRepository:        ormRepository,
		aes256Pool:           aes256Pool,
		intermediateKeyCache: intermediateKeyCache,
		leafKeyCache:         leafKeyCache,
	}, nil
}

func (d *BarrierService) Shutdown() {
	if d.aes256Pool != nil {
		d.aes256Pool.Close()
	}
	if d.leafKeyCache != nil {
		err := d.leafKeyCache.Shutdown()
		if err != nil {
			d.telemetryService.Slogger.Error("failed to shutdown leaf key cache", "error", err)
		}
	}
	if d.intermediateKeyCache != nil {
		err := d.intermediateKeyCache.Shutdown()
		if err != nil {
			d.telemetryService.Slogger.Error("failed to shutdown intermediate key cache", "error", err)
		}
	}
}

func (d *BarrierService) Encrypt(clearBytes []byte) ([]byte, error) {
	intermediateJwk, err := d.intermediateKeyCache.GetLatest()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest intermediate JWK from cache: %w", err)
	}

	leafJwk, _, err := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgDIRECT)
	if err != nil {
		return nil, fmt.Errorf("failed to generate leaf JWK: %w", err)
	}
	err = d.leafKeyCache.Put(leafJwk, intermediateJwk)
	if err != nil {
		return nil, fmt.Errorf("failed to put leaf JWK in cache: %w", err)
	}
	jweMessage, encodedJweMessage, err := cryptoutilJose.EncryptBytes(leafJwk, clearBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt clear bytes: %w", err)
	}
	jweHeaders, err := cryptoutilJose.JSONHeadersString(jweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWE message headers: %w", err)
	}
	d.telemetryService.Slogger.Info("Encrypted", "jwe", jweHeaders)

	return encodedJweMessage, nil
}

func (d *BarrierService) Decrypt(encodedJweMessage []byte) ([]byte, error) {
	jweMessage, err := joseJwe.Parse(encodedJweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message: %w", err)
	}
	var kid string
	err = jweMessage.ProtectedHeaders().Get(joseJwk.KeyIDKey, &kid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message kid: %w", err)
	}
	uuid, err := googleUuid.Parse(kid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid as uuid: %w", err)
	}
	jwk, err := d.leafKeyCache.Get(uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid as uuid: %w", err)
	}
	decryptedBytes, err := cryptoutilJose.DecryptBytes(jwk, encodedJweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with JWK %s: %w", kid, err)
	}
	return decryptedBytes, nil
}

// Helpers

func deserilalizeLatest(barrierKey cryptoutilOrmRepository.BarrierKey, err error) (joseJwk.Key, error) {
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load latest Key from database: %w", err)
	}
	var jwk joseJwk.Key
	err = json.Unmarshal([]byte((barrierKey).GetSerialized()), &jwk)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal latest Key from database: %w", err)
	}
	return jwk, nil
}

func deserilalize(barrierKey cryptoutilOrmRepository.BarrierKey, err error) (joseJwk.Key, error) {
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load Key from database: %w", err)
	}
	var jwk joseJwk.Key
	err = json.Unmarshal([]byte((barrierKey).GetSerialized()), &jwk)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal Key from database: %w", err)
	}
	return jwk, nil
}

func newIntermediateKeyCache(ormRepository *cryptoutilOrmRepository.RepositoryProvider, telemetryService *cryptoutilTelemetry.Service) (*cryptoutilBarrierCache.Cache, error) {
	loadLatestIntermediateKey := func() (joseJwk.Key, error) {
		return deserilalizeLatest(ormRepository.GetIntermediateKeyLatest())
	}
	loadIntermediateKey := func(uuid googleUuid.UUID) (joseJwk.Key, error) {
		return deserilalize(ormRepository.GetIntermediateKey(uuid))
	}
	storeIntermediateKey := func(jwk joseJwk.Key, kek joseJwk.Key) error {
		jwkKidUuid, err := cryptoutilJose.ExtractKidUuid(jwk)
		if err != nil {
			return fmt.Errorf("failed to get jwk kid uuid: %w", err)
		}
		kekKidUuid, err := cryptoutilJose.ExtractKidUuid(kek)
		if err != nil {
			return fmt.Errorf("failed to get kek kid uuid: %w", err)
		}
		jweMessage, jweMessageBytes, err := cryptoutilJose.EncryptKey(kek, jwk)
		if err != nil {
			return fmt.Errorf("failed to serialize jwk: %w", err)
		}
		jweHeaders, err := cryptoutilJose.JSONHeadersString(jweMessage)
		if err != nil {
			return fmt.Errorf("failed to get jwe message headers: %w", err)
		}
		telemetryService.Slogger.Info("Encrypted", "JWE Headers", jweHeaders)

		return ormRepository.AddIntermediateKey(&cryptoutilOrmRepository.IntermediateKey{UUID: jwkKidUuid, Serialized: string(jweMessageBytes), KEKUUID: kekKidUuid})
	}
	deleteKey := func(uuid googleUuid.UUID) (joseJwk.Key, error) {
		return deserilalize(ormRepository.DeleteIntermediateKey(uuid))
	}

	intermediateKeyCache, err := cryptoutilBarrierCache.NewJWKCache("Intermediate", telemetryService, intermediateKeyCacheSize, loadLatestIntermediateKey, loadIntermediateKey, storeIntermediateKey, deleteKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create intermediate Key cache: %w", err)
	}

	latestJwk, err := intermediateKeyCache.GetLatest()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest intermediate Key: %w", err)
	}
	if latestJwk == nil {
		intermediateJwk, _, err := cryptoutilJose.GenerateAesJWK(cryptoutilJose.AlgDIRECT)
		if err != nil {
			return nil, fmt.Errorf("failed to generate DEK JWK: %w", err)
		}
		err = intermediateKeyCache.Put(intermediateJwk, rootJwk)
		if err != nil {
			return nil, fmt.Errorf("failed to store first intermediate Key: %w", err)
		}
	}

	return intermediateKeyCache, nil
}

func newLeafKeyCache(ormRepository *cryptoutilOrmRepository.RepositoryProvider, telemetryService *cryptoutilTelemetry.Service) (*cryptoutilBarrierCache.Cache, error) {
	loadLatestLeafKey := func() (joseJwk.Key, error) {
		return deserilalizeLatest(ormRepository.GetLeafKeyLatest())
	}
	loadLeafKey := func(uuid googleUuid.UUID) (joseJwk.Key, error) {
		return deserilalize(ormRepository.GetLeafKey(uuid))
	}
	storeLeafKey := func(jwk joseJwk.Key, kek joseJwk.Key) error {
		jwkKidUuid, err := cryptoutilJose.ExtractKidUuid(jwk)
		if err != nil {
			return fmt.Errorf("failed to get jwk kid uuid: %w", err)
		}
		kekKidUuid, err := cryptoutilJose.ExtractKidUuid(kek)
		if err != nil {
			return fmt.Errorf("failed to get kek kid uuid: %w", err)
		}
		jweMessage, jweMessageBytes, err := cryptoutilJose.EncryptKey(kek, jwk)
		if err != nil {
			return fmt.Errorf("failed to serialize jwk: %w", err)
		}
		jweHeaders, err := cryptoutilJose.JSONHeadersString(jweMessage)
		if err != nil {
			return fmt.Errorf("failed to get jwe message headers: %w", err)
		}
		telemetryService.Slogger.Info("Encrypted", "JWE Headers", jweHeaders)

		return ormRepository.AddLeafKey(&cryptoutilOrmRepository.LeafKey{UUID: jwkKidUuid, Serialized: string(jweMessageBytes), KEKUUID: kekKidUuid})
	}
	deleteKey := func(uuid googleUuid.UUID) (joseJwk.Key, error) {
		return deserilalize(ormRepository.DeleteLeafKey(uuid))
	}

	leafKeyCache, err := cryptoutilBarrierCache.NewJWKCache("Leaf", telemetryService, leafKeyCacheSize, loadLatestLeafKey, loadLeafKey, storeLeafKey, deleteKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create leaf Key cache: %w", err)
	}
	return leafKeyCache, nil
}
