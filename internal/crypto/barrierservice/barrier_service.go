package barrierservice

import (
	"context"
	cryptoutilBarrierRepository "cryptoutil/internal/crypto/barrierrepository"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"
	"errors"
	"fmt"
	"sync"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

var shutdownOnce sync.Once

type BarrierService struct {
	telemetryService          *cryptoutilTelemetry.Service
	ormRepository             *cryptoutilOrmRepository.RepositoryProvider
	aes256Pool                *cryptoutilKeygen.KeyPool
	rootKeyRepository         *cryptoutilBarrierRepository.Repository
	intermediateKeyRepository *cryptoutilBarrierRepository.Repository
	contentKeyRepository      *cryptoutilBarrierRepository.Repository
	closed                    bool
}

func NewBarrierService(ctx context.Context, telemetryService *cryptoutilTelemetry.Service, ormRepository *cryptoutilOrmRepository.RepositoryProvider) (*BarrierService, error) {
	keyPoolConfig, err := cryptoutilKeygen.NewKeyPoolConfig(ctx, telemetryService, "Crypto Service AES-256", 3, 6, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(256))
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-256 pool config: %w", err)
	}
	aes256Pool, err := cryptoutilKeygen.NewKeyPool(keyPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-256 pool: %w", err)
	}

	rootKeyRepository, err := newRootKeyRepository(ormRepository, telemetryService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize root JWK repository: %w", err)
	}

	intermediateKeyRepository, err := newIntermediateKeyRepository(rootKeyRepository, 2, ormRepository, telemetryService, aes256Pool)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize intermediate JWK repository: %w", err)
	}

	contentKeyRepository, err := newContentKeyRepository(intermediateKeyRepository, 10, ormRepository, telemetryService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize content JWK repository: %w", err)
	}

	return &BarrierService{
		telemetryService:          telemetryService,
		ormRepository:             ormRepository,
		aes256Pool:                aes256Pool,
		rootKeyRepository:         rootKeyRepository,
		intermediateKeyRepository: intermediateKeyRepository,
		contentKeyRepository:      contentKeyRepository,
		closed:                    false,
	}, nil
}

func (d *BarrierService) Shutdown() {
	shutdownOnce.Do(func() {
		if d.contentKeyRepository != nil {
			err := d.contentKeyRepository.Shutdown()
			if err != nil {
				d.telemetryService.Slogger.Error("failed to shutdown content key cache", "error", err)
			}
			d.contentKeyRepository = nil
		}
		if d.intermediateKeyRepository != nil {
			err := d.intermediateKeyRepository.Shutdown()
			if err != nil {
				d.telemetryService.Slogger.Error("failed to shutdown intermediate key cache", "error", err)
			}
			d.intermediateKeyRepository = nil
		}
		if d.rootKeyRepository != nil {
			err := d.rootKeyRepository.Shutdown()
			if err != nil {
				d.telemetryService.Slogger.Error("failed to shutdown root key cache", "error", err)
			}
			d.rootKeyRepository = nil
		}
		if d.aes256Pool != nil {
			d.aes256Pool.Close()
			d.aes256Pool = nil
		}
		d.ormRepository = nil
		d.telemetryService = nil
		d.closed = true
	})
}

func (d *BarrierService) EncryptContent(clearBytes []byte) ([]byte, error) {
	if d.closed {
		return nil, fmt.Errorf("barrier service is closed")
	}
	rawKey, ok := d.aes256Pool.Get().Private.([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to cast AES-256 pool key to []byte")
	}
	cek, _, err := cryptoutilJose.CreateAesJWK(cryptoutilJose.AlgDIRECT, rawKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content JWK: %w", err)
	}
	err = d.contentKeyRepository.Put(cek)
	if err != nil {
		return nil, fmt.Errorf("failed to put content JWK in cache: %w", err)
	}
	jweMessage, encodedJweMessage, err := cryptoutilJose.EncryptBytes([]joseJwk.Key{cek}, clearBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt clear bytes: %w", err)
	}
	jweHeaders, err := cryptoutilJose.JSONHeadersString(jweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWE message headers: %w", err)
	}
	d.telemetryService.Slogger.Info("Encrypted Bytes", "JWE Headers", jweHeaders)

	return encodedJweMessage, nil
}

func (d *BarrierService) DecryptContent(encodedJweMessage []byte) ([]byte, error) {
	if d.closed {
		return nil, fmt.Errorf("barrier service is closed")
	}
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
	jwk, err := d.contentKeyRepository.Get(uuid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid as uuid: %w", err)
	}
	decryptedBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{jwk}, encodedJweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with JWK %s: %w", kid, err)
	}
	return decryptedBytes, nil
}

// Helpers

func encrypt(jwk joseJwk.Key, kekRepository *cryptoutilBarrierRepository.Repository, telemetryService *cryptoutilTelemetry.Service) (googleUuid.UUID, googleUuid.UUID, []byte, error) {
	jwkKidUuid, err := cryptoutilJose.ExtractKidUuid(jwk)
	if err != nil {
		return googleUuid.UUID{}, googleUuid.UUID{}, nil, fmt.Errorf("failed to get jwk kid uuid: %w", err)
	}

	kek, err := kekRepository.GetLatest()
	if err != nil {
		return googleUuid.UUID{}, googleUuid.UUID{}, nil, fmt.Errorf("failed to get latest kek jwk kid uuid: %w", err)
	}
	kekKidUuid, err := cryptoutilJose.ExtractKidUuid(kek)
	if err != nil {
		return googleUuid.UUID{}, googleUuid.UUID{}, nil, fmt.Errorf("failed to get latest kek kid uuid: %w", err)
	}

	jweMessage, jweMessageBytes, err := cryptoutilJose.EncryptKey([]joseJwk.Key{kek}, jwk)
	if err != nil {
		return googleUuid.UUID{}, googleUuid.UUID{}, nil, fmt.Errorf("failed to serialize jwk: %w", err)
	}
	jweHeaders, err := cryptoutilJose.JSONHeadersString(jweMessage)
	if err != nil {
		return googleUuid.UUID{}, googleUuid.UUID{}, nil, fmt.Errorf("failed to get jwe message headers: %w", err)
	}
	telemetryService.Slogger.Info("Encrypted Intermediate JWK", "JWE Headers", jweHeaders)

	return jwkKidUuid, kekKidUuid, jweMessageBytes, nil
}

func decrypt(kekRepository *cryptoutilBarrierRepository.Repository, barrierKey cryptoutilOrmRepository.BarrierKey, err error) (joseJwk.Key, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to load Key from database: %w", err)
	}
	kekJwk, err := kekRepository.Get(barrierKey.GetKEKUUID())
	if err != nil {
		return nil, fmt.Errorf("failed to parse kek kid from database: %w", err)
	}
	jwk, err := cryptoutilJose.DecryptKey([]joseJwk.Key{kekJwk}, []byte((barrierKey).GetSerialized()))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt JWK from database: %w", err)
	}
	return jwk, nil
}

func newRootKeyRepository(ormRepository *cryptoutilOrmRepository.RepositoryProvider, telemetryService *cryptoutilTelemetry.Service) (*cryptoutilBarrierRepository.Repository, error) {
	unsealService, err := newUnsealService(telemetryService, ormRepository)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize unseal service")
	}

	loadLatestRootKey := func() (joseJwk.Key, error) {
		return *unsealService.getLatest(), nil
	}
	loadRootKey := func(uuid googleUuid.UUID) (joseJwk.Key, error) {
		return *unsealService.get(uuid), nil
	}
	storeRootKey := func(jwk joseJwk.Key) error {
		return nil
	}
	deleteKey := func(uuid googleUuid.UUID) (joseJwk.Key, error) {
		return nil, nil
	}

	rootKeyRepository, err := cryptoutilBarrierRepository.New("Root", telemetryService, 1, loadLatestRootKey, loadRootKey, storeRootKey, deleteKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create root Key cache: %w", err)
	}

	return rootKeyRepository, nil
}

func newIntermediateKeyRepository(rootKeyRepository *cryptoutilBarrierRepository.Repository, cacheSize int, ormRepository *cryptoutilOrmRepository.RepositoryProvider, telemetryService *cryptoutilTelemetry.Service, aes256Pool *cryptoutilKeygen.KeyPool) (*cryptoutilBarrierRepository.Repository, error) {
	loadLatestIntermediateKey := func() (joseJwk.Key, error) {
		jwk, err := ormRepository.GetIntermediateKeyLatest()
		return decrypt(rootKeyRepository, jwk, err)
	}
	loadIntermediateKey := func(uuid googleUuid.UUID) (joseJwk.Key, error) {
		jwk, err := ormRepository.GetIntermediateKey(uuid)
		return decrypt(rootKeyRepository, jwk, err)
	}
	storeIntermediateKey := func(jwk joseJwk.Key) error {
		kekRepository := rootKeyRepository

		jwkKidUuid, kekKidUuid, jweMessageBytes, err := encrypt(jwk, kekRepository, telemetryService)
		if err != nil {
			return fmt.Errorf("failed to encrypt intermediate Key cache: %w", err)
		}

		return ormRepository.AddIntermediateKey(&cryptoutilOrmRepository.IntermediateKey{UUID: jwkKidUuid, KEKUUID: kekKidUuid, Serialized: string(jweMessageBytes)})
	}
	deleteKey := func(uuid googleUuid.UUID) (joseJwk.Key, error) {
		jwk, err := ormRepository.DeleteIntermediateKey(uuid)
		return decrypt(rootKeyRepository, jwk, err)
	}

	intermediateKeyRepository, err := cryptoutilBarrierRepository.New("Intermediate", telemetryService, cacheSize, loadLatestIntermediateKey, loadIntermediateKey, storeIntermediateKey, deleteKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create intermediate Key cache: %w", err)
	}

	latestJwk, err := intermediateKeyRepository.GetLatest()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest intermediate Key: %w", err)
	}
	if latestJwk == nil {
		intermediateJwk, _, err := cryptoutilJose.CreateAesJWK(cryptoutilJose.AlgDIRECT, aes256Pool.Get().Private.([]byte))
		if err != nil {
			return nil, fmt.Errorf("failed to generate DEK JWK: %w", err)
		}
		err = intermediateKeyRepository.Put(intermediateJwk) // calls injected storeIntermediateKey, which calls encrypt() using rootKeyRepository.GetLatest()
		if err != nil {
			return nil, fmt.Errorf("failed to store first intermediate Key: %w", err)
		}
	}

	return intermediateKeyRepository, nil
}

func newContentKeyRepository(intermediateKeyRepository *cryptoutilBarrierRepository.Repository, cacheSize int, ormRepository *cryptoutilOrmRepository.RepositoryProvider, telemetryService *cryptoutilTelemetry.Service) (*cryptoutilBarrierRepository.Repository, error) {
	loadLatestContentKey := func() (joseJwk.Key, error) {
		jwk, err := ormRepository.GetContentKeyLatest()
		return decrypt(intermediateKeyRepository, jwk, err)
	}
	loadContentKey := func(uuid googleUuid.UUID) (joseJwk.Key, error) {
		jwk, err := ormRepository.GetContentKey(uuid)
		return decrypt(intermediateKeyRepository, jwk, err)
	}
	storeContentKey := func(jwk joseJwk.Key) error {
		jwkKidUuid, err := cryptoutilJose.ExtractKidUuid(jwk)
		if err != nil {
			return fmt.Errorf("failed to get content JWK kid uuid: %w", err)
		}
		kek, err := intermediateKeyRepository.GetLatest()
		if err != nil {
			return fmt.Errorf("failed to get latest intermediate jwk kid uuid: %w", err)
		}
		kekKidUuid, err := cryptoutilJose.ExtractKidUuid(kek)
		if err != nil {
			return fmt.Errorf("failed to get kek kid uuid: %w", err)
		}
		jweMessage, jweMessageBytes, err := cryptoutilJose.EncryptKey([]joseJwk.Key{kek}, jwk)
		if err != nil {
			return fmt.Errorf("failed to serialize jwk: %w", err)
		}
		jweHeaders, err := cryptoutilJose.JSONHeadersString(jweMessage)
		if err != nil {
			return fmt.Errorf("failed to get jwe message headers: %w", err)
		}
		telemetryService.Slogger.Info("Encrypted Leaf JWK", "JWE Headers", jweHeaders)

		return ormRepository.AddContentKey(&cryptoutilOrmRepository.ContentKey{UUID: jwkKidUuid, Serialized: string(jweMessageBytes), KEKUUID: kekKidUuid})
	}
	deleteKey := func(uuid googleUuid.UUID) (joseJwk.Key, error) {
		jwk, err := ormRepository.DeleteContentKey(uuid)
		return decrypt(intermediateKeyRepository, jwk, err)
	}

	contentKeyRepository, err := cryptoutilBarrierRepository.New("Leaf", telemetryService, cacheSize, loadLatestContentKey, loadContentKey, storeContentKey, deleteKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create content JWK repository: %w", err)
	}
	return contentKeyRepository, nil
}
