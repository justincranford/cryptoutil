package businesslogic

import (
	"context"
	"crypto/ecdh"
	"crypto/elliptic"
	"errors"
	"fmt"
	"time"

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/common/crypto/keygen"
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilUtil "cryptoutil/internal/common/util"
	cryptoutilBusinessLogicModel "cryptoutil/internal/openapi/model"
	cryptoutilBarrierService "cryptoutil/internal/server/barrier"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"

	googleUuid "github.com/google/uuid"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// BusinessLogicService implements methods in StrictServerInterface
type BusinessLogicService struct {
	ormRepository         *cryptoutilOrmRepository.OrmRepository
	serviceOrmMapper      *serviceOrmMapper
	barrierService        *cryptoutilBarrierService.BarrierService
	rsa4096KeyGenPool     *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 512-bytes
	rsa3072KeyGenPool     *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 384-bytes
	rsa2048KeyGenPool     *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 256-bytes
	ecdsaP521KeyGenPool   *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 65.125-bytes
	ecdsaP384KeyGenPool   *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 48-bytes
	ecdsaP256KeyGenPool   *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 32-bytes
	ecdhP521KeyGenPool    *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 65.125-bytes
	ecdhP384KeyGenPool    *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 48-bytes
	ecdhP256KeyGenPool    *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 32-bytes
	ed25519KeyGenPool     *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 32-bytes
	aes256KeyGenPool      *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 32-bytes A256GCM, A256KW, A256GCMKW
	aes192KeyGenPool      *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 24-bytes A192GCM, A192KW, A192GCMKW
	aes128KeyGenPool      *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 16-bytes A128GCM, A128KW, A128GCMKW
	aes256HS512KeyGenPool *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 32-bytes A256CBC + 32-bytes HS512 (half of 64-bytes)
	aes192HS384KeyGenPool *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 24-bytes A192CBC + 24-bytes HS384 (half of 48-bytes)
	aes128HS256KeyGenPool *cryptoutilPool.ValueGenPool[cryptoutilKeygen.Key] // 16-bytes A128CBC + 16-bytes HS256 (half of 32-bytes)
	uuidV7KeyGenPool      *cryptoutilPool.ValueGenPool[*googleUuid.UUID]
}

func NewBusinessLogicService(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, barrierService *cryptoutilBarrierService.BarrierService) (*BusinessLogicService, error) {
	rsa4096KeyGenPoolConfig, err1 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service RSA-4096", 1, 1, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateRSAKeyPairFunction(4096))
	rsa3072KeyGenPoolConfig, err2 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service RSA-3072", 1, 1, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateRSAKeyPairFunction(3072))
	rsa2048KeyGenPoolConfig, err3 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service RSA-2048", 1, 1, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateRSAKeyPairFunction(2048))
	ecdsaP521KeyGenPoolConfig, err4 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service ECDSA-P521", 1, 1, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDSAKeyPairFunction(elliptic.P521()))
	ecdsaP384KeyGenPoolConfig, err5 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service ECDSA-P384", 1, 1, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDSAKeyPairFunction(elliptic.P384()))
	ecdsaP256KeyGenPoolConfig, err6 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service ECDSA-P256", 1, 1, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDSAKeyPairFunction(elliptic.P256()))
	ecdhP521KeyGenPoolConfig, err7 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service ECDH-P521", 1, 1, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDHKeyPairFunction(ecdh.P521()))
	ecdhP384KeyGenPoolConfig, err8 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service ECSH-P384", 1, 1, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDHKeyPairFunction(ecdh.P384()))
	ecdhP256KeyGenPoolConfig, err9 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service ECDH-P256", 1, 1, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateECDHKeyPairFunction(ecdh.P256()))
	ed25519KeyGenPoolConfig, err10 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service Ed25519", 1, 1, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateEDDSAKeyPairFunction("Ed25519"))
	aes256KeyGenPoolConfig, err11 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service AES-256-GCM", 2, 6, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(256))
	aes192KeyGenPoolConfig, err12 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service AES-192-GCM", 1, 4, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(192))
	aes128KeyGenPoolConfig, err13 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service AES-128-GCM", 1, 2, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(128))
	aes256HS512KeyGenPoolConfig, err14 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service AES-256-CBC HS-512", 1, 6, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESHSKeyFunction(512))
	aes192HS384KeyGenPoolConfig, err15 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service AES-192-CBC HS-384", 1, 4, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESHSKeyFunction(384))
	aes128HS256KeyGenPoolConfig, err16 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service AES-128-CBC HS-256", 1, 2, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESHSKeyFunction(256))
	uuidV7KeyGenPoolConfig, err17 := cryptoutilPool.NewValueGenPoolConfig(ctx, telemetryService, "Service UUIDv7", 2, 2, cryptoutilPool.MaxLifetimeValues, cryptoutilPool.MaxLifetimeDuration, cryptoutilUtil.GenerateUUIDv7Function())
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil || err8 != nil || err9 != nil || err10 != nil || err11 != nil || err12 != nil || err13 != nil || err14 != nil || err15 != nil || err16 != nil || err17 != nil {
		return nil, fmt.Errorf("failed to create pool configs: %w", errors.Join(err1, err2, err3, err4, err5, err6, err7, err8, err9, err10, err11, err12, err13, err14, err15, err16, err17))
	}

	rsa4096KeyGenPool, err1 := cryptoutilPool.NewValueGenPool(rsa4096KeyGenPoolConfig)
	rsa3072KeyGenPool, err2 := cryptoutilPool.NewValueGenPool(rsa3072KeyGenPoolConfig)
	rsa2048KeyGenPool, err3 := cryptoutilPool.NewValueGenPool(rsa2048KeyGenPoolConfig)
	ecdsaP521KeyGenPool, err4 := cryptoutilPool.NewValueGenPool(ecdsaP521KeyGenPoolConfig)
	ecdsaP384KeyGenPool, err5 := cryptoutilPool.NewValueGenPool(ecdsaP384KeyGenPoolConfig)
	ecdsaP256KeyGenPool, err6 := cryptoutilPool.NewValueGenPool(ecdsaP256KeyGenPoolConfig)
	ecdhP521KeyGenPool, err7 := cryptoutilPool.NewValueGenPool(ecdhP521KeyGenPoolConfig)
	ecdhP384KeyGenPool, err8 := cryptoutilPool.NewValueGenPool(ecdhP384KeyGenPoolConfig)
	ecdhP256KeyGenPool, err9 := cryptoutilPool.NewValueGenPool(ecdhP256KeyGenPoolConfig)
	ed25519KeyGenPool, err10 := cryptoutilPool.NewValueGenPool(ed25519KeyGenPoolConfig)
	aes256KeyGenPool, err11 := cryptoutilPool.NewValueGenPool(aes256KeyGenPoolConfig)
	aes192KeyGenPool, err12 := cryptoutilPool.NewValueGenPool(aes192KeyGenPoolConfig)
	aes128KeyGenPool, err13 := cryptoutilPool.NewValueGenPool(aes128KeyGenPoolConfig)
	aes256HS512KeyGenPool, err14 := cryptoutilPool.NewValueGenPool(aes256HS512KeyGenPoolConfig)
	aes192HS384KeyGenPool, err15 := cryptoutilPool.NewValueGenPool(aes192HS384KeyGenPoolConfig)
	aes128HS256KeyGenPool, err16 := cryptoutilPool.NewValueGenPool(aes128HS256KeyGenPoolConfig)
	uuidV7KeyGenPool, err17 := cryptoutilPool.NewValueGenPool(uuidV7KeyGenPoolConfig)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil || err8 != nil || err9 != nil || err10 != nil || err11 != nil || err12 != nil || err13 != nil || err14 != nil || err15 != nil || err16 != nil || err17 != nil {
		return nil, fmt.Errorf("failed to create pools: %w", errors.Join(err1, err2, err3, err4, err5, err6, err7, err8, err9, err10, err11, err12, err13, err14, err15, err16, err17))
	}

	return &BusinessLogicService{
		ormRepository:         ormRepository,
		serviceOrmMapper:      NewMapper(),
		barrierService:        barrierService,
		rsa4096KeyGenPool:     rsa4096KeyGenPool,
		rsa3072KeyGenPool:     rsa3072KeyGenPool,
		rsa2048KeyGenPool:     rsa2048KeyGenPool,
		ecdsaP521KeyGenPool:   ecdsaP521KeyGenPool,
		ecdsaP384KeyGenPool:   ecdsaP384KeyGenPool,
		ecdsaP256KeyGenPool:   ecdsaP256KeyGenPool,
		ecdhP521KeyGenPool:    ecdhP521KeyGenPool,
		ecdhP384KeyGenPool:    ecdhP384KeyGenPool,
		ecdhP256KeyGenPool:    ecdhP256KeyGenPool,
		ed25519KeyGenPool:     ed25519KeyGenPool,
		aes256KeyGenPool:      aes256KeyGenPool,
		aes192KeyGenPool:      aes192KeyGenPool,
		aes128KeyGenPool:      aes128KeyGenPool,
		aes256HS512KeyGenPool: aes256HS512KeyGenPool,
		aes192HS384KeyGenPool: aes192HS384KeyGenPool,
		aes128HS256KeyGenPool: aes128HS256KeyGenPool,
		uuidV7KeyGenPool:      uuidV7KeyGenPool,
	}, nil
}

func (s *BusinessLogicService) AddKeyPool(ctx context.Context, openapiKeyPoolCreate *cryptoutilBusinessLogicModel.KeyPoolCreate) (*cryptoutilBusinessLogicModel.KeyPool, error) {
	keyPoolID := s.uuidV7KeyGenPool.Get()
	repositoryKeyPoolToInsert := s.serviceOrmMapper.toOrmAddKeyPool(*keyPoolID, openapiKeyPoolCreate)

	var insertedKeyPool *cryptoutilOrmRepository.KeyPool
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		err := sqlTransaction.AddKeyPool(repositoryKeyPoolToInsert)
		if err != nil {
			return fmt.Errorf("failed to add KeyPool: %w", err)
		}

		err = TransitionState(cryptoutilBusinessLogicModel.Creating, cryptoutilBusinessLogicModel.KeyPoolStatus(repositoryKeyPoolToInsert.KeyPoolStatus))
		if repositoryKeyPoolToInsert.KeyPoolStatus != cryptoutilOrmRepository.PendingGenerate {
			return fmt.Errorf("invalid KeyPoolStatus transition detected: %w", err)
		}

		if repositoryKeyPoolToInsert.KeyPoolStatus != cryptoutilOrmRepository.PendingGenerate {
			return nil // import first key manually later
		}

		// generate first key automatically now
		repositoryKey, err := s.generateKeyPoolKeyForInsert(sqlTransaction, *keyPoolID, repositoryKeyPoolToInsert.KeyPoolAlgorithm)
		if err != nil {
			return fmt.Errorf("failed to generate key material: %w", err)
		}

		err = sqlTransaction.AddKeyPoolKey(repositoryKey)
		if err != nil {
			return fmt.Errorf("failed to add key: %w", err)
		}

		err = sqlTransaction.UpdateKeyPoolStatus(*keyPoolID, cryptoutilOrmRepository.Active)
		if err != nil {
			return fmt.Errorf("failed to update KeyPoolStatus to active: %w", err)
		}

		insertedKeyPool, err = sqlTransaction.GetKeyPool(*keyPoolID)
		if err != nil {
			return fmt.Errorf("failed to get updated KeyPool from DB: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add key pool: %w", err)
	}

	return s.serviceOrmMapper.toServiceKeyPool(insertedKeyPool), nil
}

func (s *BusinessLogicService) GetKeyPoolByKeyPoolID(ctx context.Context, keyPoolID googleUuid.UUID) (*cryptoutilBusinessLogicModel.KeyPool, error) {
	var repositoryKeyPool *cryptoutilOrmRepository.KeyPool
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeyPool, err = sqlTransaction.GetKeyPool(keyPoolID)
		if err != nil {
			return fmt.Errorf("failed to get KeyPool: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get KeyPool: %w", err)
	}
	return s.serviceOrmMapper.toServiceKeyPool(repositoryKeyPool), nil
}

func (s *BusinessLogicService) GetKeyPools(ctx context.Context, keyPoolQueryParams *cryptoutilBusinessLogicModel.KeyPoolsQueryParams) ([]cryptoutilBusinessLogicModel.KeyPool, error) {
	ormKeyPoolsQueryParams, err := s.serviceOrmMapper.toOrmGetKeyPoolsQueryParams(keyPoolQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid Get Key Pools parameters: %w", err)
	}
	var repositoryKeyPools []cryptoutilOrmRepository.KeyPool
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeyPools, err = sqlTransaction.GetKeyPools(ormKeyPoolsQueryParams)
		if err != nil {
			return fmt.Errorf("failed to list KeyPools: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list KeyPools: %w", err)
	}
	return s.serviceOrmMapper.toServiceKeyPools(repositoryKeyPools), nil
}

func (s *BusinessLogicService) GenerateKeyInPoolKey(ctx context.Context, keyPoolID googleUuid.UUID, _ *cryptoutilBusinessLogicModel.KeyGenerate) (*cryptoutilBusinessLogicModel.Key, error) {
	var repositoryKey *cryptoutilOrmRepository.Key
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeyPool, err := sqlTransaction.GetKeyPool(keyPoolID)
		if err != nil {
			return fmt.Errorf("failed to get KeyPool by KeyPoolID: %w", err)
		}

		if repositoryKeyPool.KeyPoolStatus != cryptoutilOrmRepository.PendingGenerate && repositoryKeyPool.KeyPoolStatus != cryptoutilOrmRepository.Active {
			return fmt.Errorf("invalid KeyPoolStatus detected for generate Key: %w", err)
		}

		repositoryKey, err = s.generateKeyPoolKeyForInsert(sqlTransaction, repositoryKeyPool.KeyPoolID, repositoryKeyPool.KeyPoolAlgorithm)
		if err != nil {
			return fmt.Errorf("failed to generate key material: %w", err)
		}

		err = sqlTransaction.AddKeyPoolKey(repositoryKey)
		if err != nil {
			return fmt.Errorf("failed to insert Key: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in KeyPool: %w", err)
	}

	openapiPostKeypoolKeyPoolIDKeyResponseObject := *s.serviceOrmMapper.toServiceKey(repositoryKey)
	return &openapiPostKeypoolKeyPoolIDKeyResponseObject, nil
}

func (s *BusinessLogicService) GetKeysByKeyPool(ctx context.Context, keyPoolID googleUuid.UUID, keyPoolKeysQueryParams *cryptoutilBusinessLogicModel.KeyPoolKeysQueryParams) ([]cryptoutilBusinessLogicModel.Key, error) {
	ormKeyPoolKeysQueryParams, err := s.serviceOrmMapper.toOrmGetKeyPoolKeysQueryParams(keyPoolKeysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid Get Key Pool Keys parameters: %w", err)
	}
	var repositoryKeys []cryptoutilOrmRepository.Key
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeys, err = sqlTransaction.GetKeyPoolKeys(keyPoolID, ormKeyPoolKeysQueryParams)
		if err != nil {
			return fmt.Errorf("failed to list Keys by KeyPoolID: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in KeyPool: %w", err)
	}

	return s.serviceOrmMapper.toServiceKeys(repositoryKeys), nil
}

func (s *BusinessLogicService) GetKeys(ctx context.Context, keysQueryParams *cryptoutilBusinessLogicModel.KeysQueryParams) ([]cryptoutilBusinessLogicModel.Key, error) {
	ormKeysQueryParams, err := s.serviceOrmMapper.toOrmGetKeysQueryParams(keysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid Get Keys parameters: %w", err)
	}
	var repositoryKeys []cryptoutilOrmRepository.Key
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeys, err = sqlTransaction.GetKeys(ormKeysQueryParams)
		if err != nil {
			return fmt.Errorf("failed to list Keys by KeyPoolID: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in KeyPool: %w", err)
	}

	return s.serviceOrmMapper.toServiceKeys(repositoryKeys), nil
}

func (s *BusinessLogicService) GetKeyByKeyPoolAndKeyID(ctx context.Context, keyPoolID googleUuid.UUID, keyID googleUuid.UUID) (*cryptoutilBusinessLogicModel.Key, error) {
	var repositoryKey *cryptoutilOrmRepository.Key
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKey, err = sqlTransaction.GetKeyPoolKey(keyPoolID, keyID)
		if err != nil {
			return fmt.Errorf("failed to get Key by KeyPoolID and KeyID: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in KeyPool: %w", err)
	}

	return s.serviceOrmMapper.toServiceKey(repositoryKey), nil
}

func (s *BusinessLogicService) PostEncryptByKeyPoolID(ctx context.Context, keyPoolID googleUuid.UUID, encryptParams *cryptoutilBusinessLogicModel.SymmetricEncryptParams, clearPayloadBytes []byte) ([]byte, error) {
	keyPool, keyPoolKey, decryptedKeyPoolKeyMaterialBytes, err := s.getAndDecryptKeyPoolKeyMaterial(ctx, &keyPoolID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt latest Key from Key Pool for KeyPoolID: %w", err)
	}
	if keyPool.KeyPoolProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	enc, alg, err := s.toEncAndAlg(&keyPool.KeyPoolAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to map enc and alg from Key Pool Algorithm: %w", err)
	}
	_, jweJwk, _, err := cryptoutilJose.CreateJweJwkFromKey(&keyPoolKey.KeyID, enc, alg, &cryptoutilKeygen.Key{Secret: decryptedKeyPoolKeyMaterialBytes})
	if err != nil {
		return nil, fmt.Errorf("failed to create Key from latest Key material for KeyPoolID: %w", err)
	}
	// TODO Use encryptParams for encryption? IV, AAD
	_, jweMessageBytes, err := cryptoutilJose.EncryptBytes([]joseJwk.Key{jweJwk}, clearPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt bytes with latest Key for KeyPoolID: %w", err)
	}
	return jweMessageBytes, nil
}

func (s *BusinessLogicService) PostDecryptByKeyPoolID(ctx context.Context, keyPoolID googleUuid.UUID, encryptedPayloadBytes []byte) ([]byte, error) {
	jweMessage, err := joseJwe.Parse(encryptedPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message bytes: %w", err)
	}
	kidUuid, enc, alg, err := cryptoutilJose.ExtractKidEncAlgFromJweMessage(jweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get kid, enc, and alg from JWE message: %w", err)
	}
	keyPool, keyPoolKey, decryptedKeyPoolKeyMaterialBytes, err := s.getAndDecryptKeyPoolKeyMaterial(ctx, &keyPoolID, kidUuid)
	if keyPool.KeyPoolProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	_, jweJwk, _, err := cryptoutilJose.CreateJweJwkFromKey(&keyPoolKey.KeyID, enc, alg, &cryptoutilKeygen.Key{Secret: decryptedKeyPoolKeyMaterialBytes})
	if err != nil {
		return nil, fmt.Errorf("failed to create Key from latest Key material for KeyPoolID from JWE kid UUID: %w", err)
	}
	decryptedJweMessageBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{jweJwk}, encryptedPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt bytes with Key for KeyPoolID from JWE kid UUID: %w", err)
	}
	return decryptedJweMessageBytes, nil
}

func (s *BusinessLogicService) PostSignByKeyPoolID(ctx context.Context, keyPoolID googleUuid.UUID, clearPayloadBytes []byte) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *BusinessLogicService) PostVerifyByKeyPoolID(ctx context.Context, keyPoolID googleUuid.UUID, signedPayloadBytes []byte) error {
	return fmt.Errorf("not implemented")
}

func (s *BusinessLogicService) generateKeyPoolKeyForInsert(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, keyPoolID googleUuid.UUID, keyPoolAlgorithm cryptoutilOrmRepository.KeyPoolAlgorithm) (*cryptoutilOrmRepository.Key, error) {
	keyID := s.uuidV7KeyGenPool.Get()

	// TODO Generate JWK instead of []byte
	clearKeyMaterial, err := s.GenerateKeyMaterial(keyPoolAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Key material: %w", err)
	}
	repositoryKeyGenerateDate := time.Now().UTC()

	encryptedKeyMaterial, err := s.barrierService.EncryptContent(sqlTransaction, clearKeyMaterial)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt Key material: %w", err)
	}

	return &cryptoutilOrmRepository.Key{
		KeyPoolID:       keyPoolID,
		KeyID:           *keyID,
		KeyMaterial:     encryptedKeyMaterial,
		KeyGenerateDate: &repositoryKeyGenerateDate,
	}, nil
}

func (s *BusinessLogicService) GenerateKeyMaterial(keyPoolAlgorithm cryptoutilOrmRepository.KeyPoolAlgorithm) ([]byte, error) {
	switch keyPoolAlgorithm {
	case cryptoutilOrmRepository.A256GCM_A256KW, cryptoutilOrmRepository.A192GCM_A256KW, cryptoutilOrmRepository.A128GCM_A256KW,
		cryptoutilOrmRepository.A256GCM_A256GCMKW, cryptoutilOrmRepository.A192GCM_A256GCMKW, cryptoutilOrmRepository.A128GCM_A256GCMKW,
		cryptoutilOrmRepository.A256CBCHS512_A256KW, cryptoutilOrmRepository.A192CBCHS384_A256KW, cryptoutilOrmRepository.A128CBCHS256_A256KW,
		cryptoutilOrmRepository.A256CBCHS512_A256GCMKW, cryptoutilOrmRepository.A192CBCHS384_A256GCMKW, cryptoutilOrmRepository.A128CBCHS256_A256GCMKW,
		cryptoutilOrmRepository.A256GCM_dir:
		return s.aes256KeyGenPool.Get().Secret.([]byte), nil
	case cryptoutilOrmRepository.A192GCM_A192KW, cryptoutilOrmRepository.A128GCM_A192KW,
		cryptoutilOrmRepository.A192GCM_A192GCMKW, cryptoutilOrmRepository.A128GCM_A192GCMKW,
		cryptoutilOrmRepository.A192CBCHS384_A192KW, cryptoutilOrmRepository.A128CBCHS256_A192KW,
		cryptoutilOrmRepository.A192CBCHS384_A192GCMKW, cryptoutilOrmRepository.A128CBCHS256_A192GCMKW,
		cryptoutilOrmRepository.A192GCM_dir:
		return s.aes192KeyGenPool.Get().Secret.([]byte), nil
	case cryptoutilOrmRepository.A128GCM_A128KW,
		cryptoutilOrmRepository.A128GCM_A128GCMKW,
		cryptoutilOrmRepository.A128CBCHS256_A128KW,
		cryptoutilOrmRepository.A128CBCHS256_A128GCMKW,
		cryptoutilOrmRepository.A128GCM_dir:
		return s.aes128KeyGenPool.Get().Secret.([]byte), nil
	case cryptoutilOrmRepository.A256CBCHS512_dir:
		return s.aes256HS512KeyGenPool.Get().Secret.([]byte), nil
	case cryptoutilOrmRepository.A192CBCHS384_dir:
		return s.aes192HS384KeyGenPool.Get().Secret.([]byte), nil
	case cryptoutilOrmRepository.A128CBCHS256_dir:
		return s.aes128HS256KeyGenPool.Get().Secret.([]byte), nil
	default:
		return nil, fmt.Errorf("unsuppported keyPoolAlgorithm: %s", keyPoolAlgorithm)
	}
}

func (s *BusinessLogicService) getAndDecryptKeyPoolKeyMaterial(ctx context.Context, keyPoolID *googleUuid.UUID, kidUuid *googleUuid.UUID) (*cryptoutilOrmRepository.KeyPool, *cryptoutilOrmRepository.Key, []byte, error) {
	var repositoryKeyPool *cryptoutilOrmRepository.KeyPool
	var repositoryKeyPoolKey *cryptoutilOrmRepository.Key
	var decryptedKeyPoolKeyMaterialBytes []byte
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeyPool, err = sqlTransaction.GetKeyPool(*keyPoolID)
		if err != nil {
			return fmt.Errorf("failed to get KeyPool from KeyPool: %w", err)
		}
		if kidUuid == nil {
			repositoryKeyPoolKey, err = sqlTransaction.GetKeyPoolLatestKey(*keyPoolID)
			if err != nil {
				return fmt.Errorf("failed to latest Key from KeyPool: %w", err)
			}
		} else {
			repositoryKeyPoolKey, err = sqlTransaction.GetKeyPoolKey(*keyPoolID, *kidUuid)
			if err != nil {
				return fmt.Errorf("failed to specified Key from KeyPool: %w", err)
			}
		}
		decryptedKeyPoolKeyMaterialBytes, err = s.barrierService.DecryptContent(sqlTransaction, repositoryKeyPoolKey.KeyMaterial)
		if err != nil {
			return fmt.Errorf("failed to decrypt Key material from KeyPool: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get and decrypt Key material from KeyPool: %w", err)
	}
	return repositoryKeyPool, repositoryKeyPoolKey, decryptedKeyPoolKeyMaterialBytes, nil
}
