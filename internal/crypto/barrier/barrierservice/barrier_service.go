package barrierservice

import (
	"context"
	"errors"
	"fmt"
	"sync"

	cryptoutilBarrierRepository "cryptoutil/internal/crypto/barrier/barrierrepository"
	cryptoutilUnsealService "cryptoutil/internal/crypto/barrier/unsealservice"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

type BarrierService struct {
	telemetryService          *cryptoutilTelemetry.TelemetryService
	ormRepository             *cryptoutilOrmRepository.OrmRepository
	aes256KeyGenPool          *cryptoutilKeygen.KeyGenPool
	rootKeyRepository         *cryptoutilBarrierRepository.BarrierRepository
	intermediateKeyRepository *cryptoutilBarrierRepository.BarrierRepository
	contentKeyRepository      *cryptoutilBarrierRepository.BarrierRepository
	closed                    bool
	shutdownOnce              sync.Once
}

func NewBarrierService(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, unsealService *cryptoutilUnsealService.UnsealService) (*BarrierService, error) {
	keyPoolConfig, err := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Crypto Service AES-256", 3, 6, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(256))
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-256 pool config: %w", err)
	}
	aes256KeyGenPool, err := cryptoutilKeygen.NewGenKeyPool(keyPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES-256 pool: %w", err)
	}

	rootKeyRepository, err := newRootKeyRepository(telemetryService, unsealService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize root JWK repository: %w", err)
	}

	intermediateKeyRepository, err := newIntermediateKeyRepository(telemetryService, ormRepository, rootKeyRepository, 2, aes256KeyGenPool)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize intermediate JWK repository: %w", err)
	}

	contentKeyRepository, err := newContentKeyRepository(telemetryService, ormRepository, intermediateKeyRepository, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize content JWK repository: %w", err)
	}

	return &BarrierService{
		telemetryService:          telemetryService,
		ormRepository:             ormRepository,
		aes256KeyGenPool:          aes256KeyGenPool,
		rootKeyRepository:         rootKeyRepository,
		intermediateKeyRepository: intermediateKeyRepository,
		contentKeyRepository:      contentKeyRepository,
		closed:                    false,
	}, nil
}

func (d *BarrierService) Shutdown() {
	d.shutdownOnce.Do(func() {
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
		if d.aes256KeyGenPool != nil {
			d.aes256KeyGenPool.Close()
			d.aes256KeyGenPool = nil
		}
		d.ormRepository = nil
		d.telemetryService = nil
		d.closed = true
	})
}

func (d *BarrierService) EncryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, clearBytes []byte) ([]byte, error) {
	if d.closed {
		return nil, fmt.Errorf("barrier service is closed")
	}
	rawKey, ok := d.aes256KeyGenPool.Get().Private.([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to cast AES-256 pool key to []byte")
	}
	cek, _, _, err := cryptoutilJose.CreateAesJWK(cryptoutilJose.AlgDIRECT, rawKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content JWK: %w", err)
	}
	err = d.contentKeyRepository.Put(sqlTransaction, cek)
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

func (d *BarrierService) DecryptContent(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, encodedJweMessage []byte) ([]byte, error) {
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
	kidUuid, err := googleUuid.Parse(kid)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kid as uuid: %w", err)
	}
	jwk, err := d.contentKeyRepository.Get(sqlTransaction, kidUuid)
	if err != nil {
		return nil, fmt.Errorf("failed to get key by kid as uuid: %w", err)
	}
	decryptedBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{jwk}, encodedJweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with JWK %s: %w", kid, err)
	}
	return decryptedBytes, nil
}

// Helpers

func encrypt(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, jwk joseJwk.Key, kekRepository *cryptoutilBarrierRepository.BarrierRepository, telemetryService *cryptoutilTelemetry.TelemetryService) (googleUuid.UUID, googleUuid.UUID, []byte, error) {
	jwkKidUuid, err := cryptoutilJose.ExtractKidUuid(jwk)
	if err != nil {
		return googleUuid.UUID{}, googleUuid.UUID{}, nil, fmt.Errorf("failed to get jwk kid uuid: %w", err)
	}

	kek, err := kekRepository.GetLatest(sqlTransaction)
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

	return *jwkKidUuid, *kekKidUuid, jweMessageBytes, nil
}

func decrypt(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, kekRepository *cryptoutilBarrierRepository.BarrierRepository, barrierKey cryptoutilOrmRepository.BarrierKey, err error) (joseJwk.Key, error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to load Key from database: %w", err)
	}
	kekJwk, err := kekRepository.Get(sqlTransaction, barrierKey.GetKEKUUID())
	if err != nil {
		return nil, fmt.Errorf("failed to parse kek kid from database: %w", err)
	}
	jwk, err := cryptoutilJose.DecryptKey([]joseJwk.Key{kekJwk}, []byte((barrierKey).GetEncrypted()))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt JWK from database: %w", err)
	}
	return jwk, nil
}

func newRootKeyRepository(telemetryService *cryptoutilTelemetry.TelemetryService, unsealService *cryptoutilUnsealService.UnsealService) (*cryptoutilBarrierRepository.BarrierRepository, error) {
	loadLatestRootKey := func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) (joseJwk.Key, error) {
		return unsealService.GetLatest(), nil
	}
	loadRootKey := func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, kidUuid googleUuid.UUID) (joseJwk.Key, error) {
		return unsealService.Get(kidUuid), nil
	}
	storeRootKey := func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, jwk joseJwk.Key) error {
		return nil
	}
	deleteKey := func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, kidUuid googleUuid.UUID) (joseJwk.Key, error) {
		return nil, nil
	}

	rootKeyRepository, err := cryptoutilBarrierRepository.NewBarrierRepository("Root", telemetryService, 1, loadLatestRootKey, loadRootKey, storeRootKey, deleteKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create root Key cache: %w", err)
	}

	return rootKeyRepository, nil
}

func newIntermediateKeyRepository(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, rootKeyRepository *cryptoutilBarrierRepository.BarrierRepository, cacheSize int, aes256KeyGenPool *cryptoutilKeygen.KeyGenPool) (*cryptoutilBarrierRepository.BarrierRepository, error) {
	loadLatestIntermediateKey := func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) (joseJwk.Key, error) {
		jwk, err := sqlTransaction.GetIntermediateKeyLatest()
		return decrypt(sqlTransaction, rootKeyRepository, jwk, err)
	}
	loadIntermediateKey := func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, uuid googleUuid.UUID) (joseJwk.Key, error) {
		jwk, err := sqlTransaction.GetIntermediateKey(uuid)
		return decrypt(sqlTransaction, rootKeyRepository, jwk, err)
	}
	storeIntermediateKey := func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, jwk joseJwk.Key) error {
		kekRepository := rootKeyRepository

		jwkKidUuid, kekKidUuid, jweMessageBytes, err := encrypt(sqlTransaction, jwk, kekRepository, telemetryService)
		if err != nil {
			return fmt.Errorf("failed to encrypt intermediate Key cache: %w", err)
		}

		return sqlTransaction.AddIntermediateKey(&cryptoutilOrmRepository.BarrierIntermediateKey{UUID: jwkKidUuid, KEKUUID: kekKidUuid, Encrypted: string(jweMessageBytes)})
	}
	deleteKey := func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, uuid googleUuid.UUID) (joseJwk.Key, error) {
		jwk, err := sqlTransaction.DeleteIntermediateKey(uuid)
		return decrypt(sqlTransaction, rootKeyRepository, jwk, err)
	}

	intermediateKeyRepository, err := cryptoutilBarrierRepository.NewBarrierRepository("Intermediate", telemetryService, cacheSize, loadLatestIntermediateKey, loadIntermediateKey, storeIntermediateKey, deleteKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create intermediate Key cache: %w", err)
	}

	err = ormRepository.WithTransaction(context.Background(), cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		latestJwk, err := intermediateKeyRepository.GetLatest(sqlTransaction)
		if err != nil {
			return fmt.Errorf("failed to get latest intermediate Key: %w", err)
		}
		if latestJwk == nil {
			intermediateJwk, _, _, err := cryptoutilJose.CreateAesJWK(cryptoutilJose.AlgDIRECT, aes256KeyGenPool.Get().Private.([]byte))
			if err != nil {
				return fmt.Errorf("failed to generate DEK JWK: %w", err)
			}
			err = intermediateKeyRepository.Put(sqlTransaction, intermediateJwk) // calls injected storeIntermediateKey, which calls encrypt() using rootKeyRepository.GetLatest()
			if err != nil {
				return fmt.Errorf("failed to store first intermediate Key: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get updated KeyPool from DB: %w", err)
	}

	return intermediateKeyRepository, nil
}

func newContentKeyRepository(telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, intermediateKeyRepository *cryptoutilBarrierRepository.BarrierRepository, cacheSize int) (*cryptoutilBarrierRepository.BarrierRepository, error) {
	loadLatestContentKey := func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) (joseJwk.Key, error) {
		jwk, err := sqlTransaction.GetContentKeyLatest()
		return decrypt(sqlTransaction, intermediateKeyRepository, jwk, err)
	}
	loadContentKey := func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, uuid googleUuid.UUID) (joseJwk.Key, error) {
		jwk, err := sqlTransaction.GetContentKey(uuid)
		return decrypt(sqlTransaction, intermediateKeyRepository, jwk, err)
	}
	storeContentKey := func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, jwk joseJwk.Key) error {
		jwkKidUuid, err := cryptoutilJose.ExtractKidUuid(jwk)
		if err != nil {
			return fmt.Errorf("failed to get content JWK kid uuid: %w", err)
		}
		kek, err := intermediateKeyRepository.GetLatest(sqlTransaction)
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

		return sqlTransaction.AddContentKey(&cryptoutilOrmRepository.BarrierContentKey{UUID: *jwkKidUuid, Encrypted: string(jweMessageBytes), KEKUUID: *kekKidUuid})
	}
	deleteKey := func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, uuid googleUuid.UUID) (joseJwk.Key, error) {
		jwk, err := sqlTransaction.DeleteContentKey(uuid)
		return decrypt(sqlTransaction, intermediateKeyRepository, jwk, err)
	}

	contentKeyRepository, err := cryptoutilBarrierRepository.NewBarrierRepository("Leaf", telemetryService, cacheSize, loadLatestContentKey, loadContentKey, storeContentKey, deleteKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create content JWK repository: %w", err)
	}
	return contentKeyRepository, nil
}
