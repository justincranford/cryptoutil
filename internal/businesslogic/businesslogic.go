package businesslogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	cryptoutilBarrierService "cryptoutil/internal/crypto/barrier"
	cryptoutilJose "cryptoutil/internal/crypto/jose"
	cryptoutilKeygen "cryptoutil/internal/crypto/keygen"
	cryptoutilBusinessLogicModel "cryptoutil/internal/openapi/model"
	cryptoutilOrmRepository "cryptoutil/internal/repository/orm"
	cryptoutilTelemetry "cryptoutil/internal/telemetry"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// BusinessLogicService implements methods in StrictServerInterface
type BusinessLogicService struct {
	ormRepository         *cryptoutilOrmRepository.OrmRepository
	serviceOrmMapper      *serviceOrmMapper
	aes256KeyGenPool      *cryptoutilKeygen.KeyGenPool // 32-bytes A256GCM, A256KW, A256GCMKW
	aes192KeyGenPool      *cryptoutilKeygen.KeyGenPool // 24-bytes A192GCM, A192KW, A192GCMKW
	aes128KeyGenPool      *cryptoutilKeygen.KeyGenPool // 16-bytes A128GCM, A128KW, A128GCMKW
	aes256HS512KeyGenPool *cryptoutilKeygen.KeyGenPool // 32-bytes A256CBC + 32-bytes HS512 (half of 64-bytes)
	aes192HS384KeyGenPool *cryptoutilKeygen.KeyGenPool // 24-bytes A192CBC + 24-bytes HS384 (half of 48-bytes)
	aes128HS256KeyGenPool *cryptoutilKeygen.KeyGenPool // 16-bytes A128CBC + 16-bytes HS256 (half of 32-bytes)
	uuidV7KeyGenPool      *cryptoutilKeygen.KeyGenPool
	barrierService        *cryptoutilBarrierService.BarrierService
}

func NewBusinessLogicService(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, ormRepository *cryptoutilOrmRepository.OrmRepository, barrierService *cryptoutilBarrierService.BarrierService) (*BusinessLogicService, error) {
	aes256KeyGenPoolConfig, err1 := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Service AES-256-GCM", 2, 6, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(256))
	aes192KeyGenPoolConfig, err2 := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Service AES-192-GCM", 1, 4, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(192))
	aes128KeyGenPoolConfig, err3 := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Service AES-128-GCM", 1, 2, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESKeyFunction(128))
	aes256HS512KeyGenPoolConfig, err4 := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Service AES-256-CBC HS-512", 1, 6, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESHSKeyFunction(512))
	aes192HS384KeyGenPoolConfig, err5 := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Service AES-192-CBC HS-384", 1, 4, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESHSKeyFunction(384))
	aes128HS256KeyGenPoolConfig, err6 := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Service AES-128-CBC HS-256", 1, 2, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateAESHSKeyFunction(256))
	uuidV7KeyGenPoolConfig, err7 := cryptoutilKeygen.NewKeyGenPoolConfig(ctx, telemetryService, "Service UUIDv7", 2, 2, cryptoutilKeygen.MaxLifetimeKeys, cryptoutilKeygen.MaxLifetimeDuration, cryptoutilKeygen.GenerateUUIDv7Function())
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil {
		return nil, fmt.Errorf("failed to create pool configs: %w", errors.Join(err1, err2, err3, err4, err5, err6, err7))
	}

	aes256KeyGenPool, err1 := cryptoutilKeygen.NewGenKeyPool(aes256KeyGenPoolConfig)
	aes192KeyGenPool, err2 := cryptoutilKeygen.NewGenKeyPool(aes192KeyGenPoolConfig)
	aes128KeyGenPool, err3 := cryptoutilKeygen.NewGenKeyPool(aes128KeyGenPoolConfig)
	aes256HS512KeyGenPool, err4 := cryptoutilKeygen.NewGenKeyPool(aes256HS512KeyGenPoolConfig)
	aes192HS384KeyGenPool, err5 := cryptoutilKeygen.NewGenKeyPool(aes192HS384KeyGenPoolConfig)
	aes128HS256KeyGenPool, err6 := cryptoutilKeygen.NewGenKeyPool(aes128HS256KeyGenPoolConfig)
	uuidV7KeyGenPool, err7 := cryptoutilKeygen.NewGenKeyPool(uuidV7KeyGenPoolConfig)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil || err7 != nil {
		return nil, fmt.Errorf("failed to create pools: %w", errors.Join(err1, err2, err3, err4, err5, err6, err7))
	}

	return &BusinessLogicService{ormRepository: ormRepository, serviceOrmMapper: NewMapper(), aes256KeyGenPool: aes256KeyGenPool, aes192KeyGenPool: aes192KeyGenPool, aes128KeyGenPool: aes128KeyGenPool, aes256HS512KeyGenPool: aes256HS512KeyGenPool, aes192HS384KeyGenPool: aes192HS384KeyGenPool, aes128HS256KeyGenPool: aes128HS256KeyGenPool, uuidV7KeyGenPool: uuidV7KeyGenPool, barrierService: barrierService}, nil
}

func (s *BusinessLogicService) AddKeyPool(ctx context.Context, openapiKeyPoolCreate *cryptoutilBusinessLogicModel.KeyPoolCreate) (*cryptoutilBusinessLogicModel.KeyPool, error) {
	keyPoolID := s.uuidV7KeyGenPool.Get().Private.(googleUuid.UUID)
	repositoryKeyPoolToInsert := s.serviceOrmMapper.toOrmAddKeyPool(keyPoolID, openapiKeyPoolCreate)

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
		repositoryKey, err := s.generateKeyPoolKeyForInsert(sqlTransaction, keyPoolID, repositoryKeyPoolToInsert.KeyPoolAlgorithm)
		if err != nil {
			return fmt.Errorf("failed to generate key material: %w", err)
		}

		err = sqlTransaction.AddKeyPoolKey(repositoryKey)
		if err != nil {
			return fmt.Errorf("failed to add key: %w", err)
		}

		err = sqlTransaction.UpdateKeyPoolStatus(keyPoolID, cryptoutilOrmRepository.Active)
		if err != nil {
			return fmt.Errorf("failed to update KeyPoolStatus to active: %w", err)
		}

		insertedKeyPool, err = sqlTransaction.GetKeyPool(keyPoolID)
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

func (s *BusinessLogicService) PostEncryptByKeyPoolIDAndKeyID(ctx context.Context, keyPoolID googleUuid.UUID, encryptParams *cryptoutilBusinessLogicModel.SymmetricEncryptParams, clearPayloadBytes []byte) ([]byte, error) {
	var repositoryKeyPool *cryptoutilOrmRepository.KeyPool
	var repositoryKeyPoolLatestKey *cryptoutilOrmRepository.Key
	var decryptedKeyPoolLatestKeyMaterialBytes []byte
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeyPool, err = sqlTransaction.GetKeyPool(keyPoolID)
		if err != nil {
			return fmt.Errorf("failed to get KeyPool for KeyPoolID: %w", err)
		}
		repositoryKeyPoolLatestKey, err = sqlTransaction.GetKeyPoolLatestKey(keyPoolID)
		if err != nil {
			return fmt.Errorf("failed to latest Key material for KeyPoolID: %w", err)
		}
		decryptedKeyPoolLatestKeyMaterialBytes, err = s.barrierService.DecryptContent(sqlTransaction, repositoryKeyPoolLatestKey.KeyMaterial)
		if err != nil {
			return fmt.Errorf("failed to decrypt latest Key material for KeyPoolID: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get latest Key material for KeyPoolID: %w", err)
	}

	if repositoryKeyPool.KeyPoolProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}

	// TODO Use encryptParams for encryption? IV, AAD (N.B. Already using ALG below)

	var kekAlg *joseJwa.KeyEncryptionAlgorithm
	var cekAlg *joseJwa.ContentEncryptionAlgorithm
	switch repositoryKeyPool.KeyPoolAlgorithm {
	case cryptoutilOrmRepository.A256GCM_A256KW:
		cekAlg = &cryptoutilJose.EncA256GCM
		kekAlg = &cryptoutilJose.AlgA256KW
	case cryptoutilOrmRepository.A192GCM_A256KW:
		cekAlg = &cryptoutilJose.EncA192GCM
		kekAlg = &cryptoutilJose.AlgA256KW
	case cryptoutilOrmRepository.A128GCM_A256KW:
		cekAlg = &cryptoutilJose.EncA128GCM
		kekAlg = &cryptoutilJose.AlgA256KW
	case cryptoutilOrmRepository.A192GCM_A192KW:
		cekAlg = &cryptoutilJose.EncA192GCM
		kekAlg = &cryptoutilJose.AlgA192KW
	case cryptoutilOrmRepository.A128GCM_A192KW:
		cekAlg = &cryptoutilJose.EncA128GCM
		kekAlg = &cryptoutilJose.AlgA192KW
	case cryptoutilOrmRepository.A128GCM_A128KW:
		cekAlg = &cryptoutilJose.EncA128GCM
		kekAlg = &cryptoutilJose.AlgA128KW
	case cryptoutilOrmRepository.A256GCM_A256GCMKW:
		cekAlg = &cryptoutilJose.EncA256GCM
		kekAlg = &cryptoutilJose.AlgA256GCMKW
	case cryptoutilOrmRepository.A192GCM_A256GCMKW:
		cekAlg = &cryptoutilJose.EncA192GCM
		kekAlg = &cryptoutilJose.AlgA256GCMKW
	case cryptoutilOrmRepository.A128GCM_A256GCMKW:
		cekAlg = &cryptoutilJose.EncA128GCM
		kekAlg = &cryptoutilJose.AlgA256GCMKW
	case cryptoutilOrmRepository.A192GCM_A192GCMKW:
		cekAlg = &cryptoutilJose.EncA192GCM
		kekAlg = &cryptoutilJose.AlgA192GCMKW
	case cryptoutilOrmRepository.A128GCM_A192GCMKW:
		cekAlg = &cryptoutilJose.EncA128GCM
		kekAlg = &cryptoutilJose.AlgA192GCMKW
	case cryptoutilOrmRepository.A128GCM_A128GCMKW:
		cekAlg = &cryptoutilJose.EncA128GCM
		kekAlg = &cryptoutilJose.AlgA128GCMKW
	case cryptoutilOrmRepository.A256GCM_Dir:
		cekAlg = &cryptoutilJose.EncA256GCM
		kekAlg = &cryptoutilJose.AlgDIRECT
	case cryptoutilOrmRepository.A192GCM_Dir:
		cekAlg = &cryptoutilJose.EncA192GCM
		kekAlg = &cryptoutilJose.AlgDIRECT
	case cryptoutilOrmRepository.A128GCM_Dir:
		cekAlg = &cryptoutilJose.EncA128GCM
		kekAlg = &cryptoutilJose.AlgDIRECT
	case cryptoutilOrmRepository.A256CBCHS512_A256KW:
		cekAlg = &cryptoutilJose.EncA256CBC_HS512
		kekAlg = &cryptoutilJose.AlgA256KW
	case cryptoutilOrmRepository.A192CBCHS384_A256KW:
		cekAlg = &cryptoutilJose.EncA192CBC_HS384
		kekAlg = &cryptoutilJose.AlgA256KW
	case cryptoutilOrmRepository.A128CBCHS256_A256KW:
		cekAlg = &cryptoutilJose.EncA128CBC_HS256
		kekAlg = &cryptoutilJose.AlgA256KW
	case cryptoutilOrmRepository.A192CBCHS384_A192KW:
		cekAlg = &cryptoutilJose.EncA192CBC_HS384
		kekAlg = &cryptoutilJose.AlgA192KW
	case cryptoutilOrmRepository.A128CBCHS256_A192KW:
		cekAlg = &cryptoutilJose.EncA128CBC_HS256
		kekAlg = &cryptoutilJose.AlgA192KW
	case cryptoutilOrmRepository.A128CBCHS256_A128KW:
		cekAlg = &cryptoutilJose.EncA128CBC_HS256
		kekAlg = &cryptoutilJose.AlgA128KW
	case cryptoutilOrmRepository.A256CBCHS512_A256GCMKW:
		cekAlg = &cryptoutilJose.EncA256CBC_HS512
		kekAlg = &cryptoutilJose.AlgA256GCMKW
	case cryptoutilOrmRepository.A192CBCHS384_A256GCMKW:
		cekAlg = &cryptoutilJose.EncA192CBC_HS384
		kekAlg = &cryptoutilJose.AlgA256GCMKW
	case cryptoutilOrmRepository.A128CBCHS256_A256GCMKW:
		cekAlg = &cryptoutilJose.EncA128CBC_HS256
		kekAlg = &cryptoutilJose.AlgA256GCMKW
	case cryptoutilOrmRepository.A192CBCHS384_A192GCMKW:
		cekAlg = &cryptoutilJose.EncA192CBC_HS384
		kekAlg = &cryptoutilJose.AlgA192GCMKW
	case cryptoutilOrmRepository.A128CBCHS256_A192GCMKW:
		cekAlg = &cryptoutilJose.EncA128CBC_HS256
		kekAlg = &cryptoutilJose.AlgA192GCMKW
	case cryptoutilOrmRepository.A128CBCHS256_A128GCMKW:
		cekAlg = &cryptoutilJose.EncA128CBC_HS256
		kekAlg = &cryptoutilJose.AlgA128GCMKW
	case cryptoutilOrmRepository.A256CBCHS512_Dir:
		cekAlg = &cryptoutilJose.EncA256CBC_HS512
		kekAlg = &cryptoutilJose.AlgDIRECT
	case cryptoutilOrmRepository.A192CBCHS384_Dir:
		cekAlg = &cryptoutilJose.EncA192CBC_HS384
		kekAlg = &cryptoutilJose.AlgDIRECT
	case cryptoutilOrmRepository.A128CBCHS256_Dir:
		cekAlg = &cryptoutilJose.EncA128CBC_HS256
		kekAlg = &cryptoutilJose.AlgDIRECT
	default:
		return nil, fmt.Errorf("keyPool key type algorithm '%s' not supported", repositoryKeyPool.KeyPoolAlgorithm)
	}

	// envelope encrypt => latestKeyInKeyPool( randomA256GCM(clearBytes) )
	_, latestKeyInKeyPool, _, err := cryptoutilJose.CreateAesJWKFromBytes(&repositoryKeyPoolLatestKey.KeyID, kekAlg, cekAlg, decryptedKeyPoolLatestKeyMaterialBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create Key from latest Key material for KeyPoolID: %w", err)
	}

	// TODO Debug
	// failed to encrypt: failed to encrypt bytes with latest Key for KeyPoolID: failed to encrypt clearBytes: jwe.Encrypt: failed to encrypt payload: failed to crypt content:
	// failed to fetch AEAD: cipher: failed to create AES cipher for CBC: failed to execute block cipher function: crypto/aes: invalid key size 8

	// JWE Headers: alg=A256GCMKW, enc=A256GCM, iv=Uy6bFPp_mflirpPN (base64url-encoded 12-byte nonce), tag=c8f7buGvHOV9FK0ls3cSug (base64url-encoded 16-byte tag), kid=019656e9-6ee4-729f-abfb-6c6986eaa3f4 (uuid v7)
	_, encryptedJweMessageBytes, err := cryptoutilJose.EncryptBytes([]joseJwk.Key{latestKeyInKeyPool}, clearPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt bytes with latest Key for KeyPoolID: %w", err)
	}
	return encryptedJweMessageBytes, nil
}

func (s *BusinessLogicService) PostDecryptByKeyPoolIDAndKeyID(ctx context.Context, keyPoolID googleUuid.UUID, encryptedPayload []byte) ([]byte, error) {
	jweMessage, err := joseJwe.Parse(encryptedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted JWE message bytes: %w", err)
	}
	var jweMessageKidUuidString string
	err = jweMessage.ProtectedHeaders().Get(joseJwk.KeyIDKey, &jweMessageKidUuidString)
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted JWE kid UUID: %w", err)
	}
	jweMessageKidUuid, err := googleUuid.Parse(jweMessageKidUuidString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse encrypted JWE kid UUID: %w", err)
	}
	var kekAlg joseJwa.KeyEncryptionAlgorithm
	err = jweMessage.ProtectedHeaders().Get(joseJwk.AlgorithmKey, &kekAlg)
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted JWE kid UUID: %w", err)
	}
	var cekAlg joseJwa.ContentEncryptionAlgorithm
	err = jweMessage.ProtectedHeaders().Get("enc", &cekAlg)
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted JWE kid UUID: %w", err)
	}

	var repositoryKeyPool *cryptoutilOrmRepository.KeyPool
	var repositoryKeyPoolKey *cryptoutilOrmRepository.Key
	var decryptedKeyPoolKeyMaterialBytes []byte
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeyPool, err = sqlTransaction.GetKeyPool(keyPoolID)
		if err != nil {
			return fmt.Errorf("failed to get KeyPool for KeyPoolID: %w", err)
		}
		repositoryKeyPoolKey, err = sqlTransaction.GetKeyPoolKey(keyPoolID, jweMessageKidUuid)
		if err != nil {
			return fmt.Errorf("failed to Key material for KeyPoolID from JWE kid UUID: %w", err)
		}
		decryptedKeyPoolKeyMaterialBytes, err = s.barrierService.DecryptContent(sqlTransaction, repositoryKeyPoolKey.KeyMaterial)
		if err != nil {
			return fmt.Errorf("failed to decrypt Key material for KeyPoolID from JWE kid UUID: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get latest Key material for KeyPoolID from JWE kid UUID: %w", err)
	}
	keyPoolProvider := repositoryKeyPool.KeyPoolProvider
	repositoryKeyPoolLatestKeyKidUuid := &repositoryKeyPoolKey.KeyID

	if keyPoolProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}

	// envelope encrypt => keyInKeyPool( randomA256GCM(clearBytes) )
	_, keyInKeyPool, _, err := cryptoutilJose.CreateAesJWKFromBytes(repositoryKeyPoolLatestKeyKidUuid, &kekAlg, &cekAlg, decryptedKeyPoolKeyMaterialBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to create Key from latest Key material for KeyPoolID from JWE kid UUID: %w", err)
	}

	// JWE Headers: alg=A256GCMKW, enc=A256GCM, iv=Uy6bFPp_mflirpPN (base64url-encoded 12-byte nonce), tag=c8f7buGvHOV9FK0ls3cSug (base64url-encoded 16-byte tag), kid=019656e9-6ee4-729f-abfb-6c6986eaa3f4 (uuid v7)
	decryptedJweMessageBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{keyInKeyPool}, encryptedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt bytes with Key for KeyPoolID from JWE kid UUID: %w", err)
	}
	return decryptedJweMessageBytes, nil
}

func (s *BusinessLogicService) generateKeyPoolKeyForInsert(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, keyPoolID googleUuid.UUID, keyPoolAlgorithm cryptoutilOrmRepository.KeyPoolAlgorithm) (*cryptoutilOrmRepository.Key, error) {
	keyID := s.uuidV7KeyGenPool.Get().Private.(googleUuid.UUID)

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
		KeyID:           keyID,
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
		cryptoutilOrmRepository.A256GCM_Dir:
		return s.aes256KeyGenPool.Get().Private.([]byte), nil
	case cryptoutilOrmRepository.A192GCM_A192KW, cryptoutilOrmRepository.A128GCM_A192KW,
		cryptoutilOrmRepository.A192GCM_A192GCMKW, cryptoutilOrmRepository.A128GCM_A192GCMKW,
		cryptoutilOrmRepository.A192CBCHS384_A192KW, cryptoutilOrmRepository.A128CBCHS256_A192KW,
		cryptoutilOrmRepository.A192CBCHS384_A192GCMKW, cryptoutilOrmRepository.A128CBCHS256_A192GCMKW,
		cryptoutilOrmRepository.A192GCM_Dir:
		return s.aes192KeyGenPool.Get().Private.([]byte), nil
	case cryptoutilOrmRepository.A128GCM_A128KW,
		cryptoutilOrmRepository.A128GCM_A128GCMKW,
		cryptoutilOrmRepository.A128CBCHS256_A128KW,
		cryptoutilOrmRepository.A128CBCHS256_A128GCMKW,
		cryptoutilOrmRepository.A128GCM_Dir:
		return s.aes128KeyGenPool.Get().Private.([]byte), nil
	case cryptoutilOrmRepository.A256CBCHS512_Dir:
		return s.aes256HS512KeyGenPool.Get().Private.([]byte), nil
	case cryptoutilOrmRepository.A192CBCHS384_Dir:
		return s.aes192HS384KeyGenPool.Get().Private.([]byte), nil
	case cryptoutilOrmRepository.A128CBCHS256_Dir:
		return s.aes128HS256KeyGenPool.Get().Private.([]byte), nil
	default:
		return nil, fmt.Errorf("unsuppported keyPoolAlgorithm: %s", keyPoolAlgorithm)
	}
}
