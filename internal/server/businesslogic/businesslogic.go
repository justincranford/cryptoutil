package businesslogic

import (
	"context"
	"fmt"
	"time"

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilBusinessLogicModel "cryptoutil/internal/openapi/model"
	cryptoutilBarrierService "cryptoutil/internal/server/barrier"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"

	googleUuid "github.com/google/uuid"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
)

// BusinessLogicService implements methods in StrictServerInterface
type BusinessLogicService struct {
	telemetryService *cryptoutilTelemetry.TelemetryService
	jwkGenService    *cryptoutilJose.JwkGenService
	ormRepository    *cryptoutilOrmRepository.OrmRepository
	serviceOrmMapper *serviceOrmMapper
	barrierService   *cryptoutilBarrierService.BarrierService
}

func NewBusinessLogicService(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, jwkGenService *cryptoutilJose.JwkGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, barrierService *cryptoutilBarrierService.BarrierService) (*BusinessLogicService, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx must be non-nil")
	} else if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	} else if ormRepository == nil {
		return nil, fmt.Errorf("ormRepository must be non-nil")
	} else if barrierService == nil {
		return nil, fmt.Errorf("ubarrierService must be non-nil")
	}

	return &BusinessLogicService{
		telemetryService: telemetryService,
		jwkGenService:    jwkGenService,
		ormRepository:    ormRepository,
		serviceOrmMapper: NewMapper(),
		barrierService:   barrierService,
	}, nil
}

func (s *BusinessLogicService) AddKeyPool(ctx context.Context, openapiKeyPoolCreate *cryptoutilBusinessLogicModel.KeyPoolCreate) (*cryptoutilBusinessLogicModel.KeyPool, error) {
	keyPoolID := s.jwkGenService.GenerateUUIDv7()
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

func (s *BusinessLogicService) PostEncryptByKeyPoolID(ctx context.Context, keyPoolID googleUuid.UUID, encryptParams *cryptoutilBusinessLogicModel.EncryptParams, clearPayloadBytes []byte) ([]byte, error) {
	keyPool, _, decryptedJweJwk, err := s.getAndDecryptKeyPoolJwk(ctx, &keyPoolID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt latest JWE JWK from Key Pool for KeyPoolID: %w", err)
	}
	_, _, err = cryptoutilJose.ExtractAlgEncFromJweJwk(&decryptedJweJwk, 0)
	if err != nil {
		return nil, fmt.Errorf("decrypted JWK doesn't have expected enc and alg headers: %w", err)
	}
	if keyPool.KeyPoolProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	// TODO Use encryptParams for encryption? IV, AAD
	_, jweMessageBytes, err := cryptoutilJose.EncryptBytes([]joseJwk.Key{decryptedJweJwk}, clearPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt bytes with latest Key for KeyPoolID: %w", err)
	}
	return jweMessageBytes, nil
}

func (s *BusinessLogicService) PostDecryptByKeyPoolID(ctx context.Context, keyPoolID googleUuid.UUID, jweMessageBytes []byte) ([]byte, error) {
	jweMessage, err := joseJwe.Parse(jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message bytes: %w", err)
	}
	kidUuid, _, _, err := cryptoutilJose.ExtractKidEncAlgFromJweMessage(jweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get kid, enc, and alg from JWE message: %w", err)
	}
	keyPool, _, decryptedJweJwk, err := s.getAndDecryptKeyPoolJwk(ctx, &keyPoolID, kidUuid)
	// TODO validate decrypted JWK is a JWE JWK
	if keyPool.KeyPoolProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	_, _, err = cryptoutilJose.ExtractAlgEncFromJweJwk(&decryptedJweJwk, 0)
	if err != nil {
		return nil, fmt.Errorf("decrypted JWK doesn't have expected enc and alg headers: %w", err)
	}
	decryptedJweMessageBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{decryptedJweJwk}, jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt bytes with Key for KeyPoolID from JWE kid UUID: %w", err)
	}
	return decryptedJweMessageBytes, nil
}

func (s *BusinessLogicService) PostSignByKeyPoolID(ctx context.Context, keyPoolID googleUuid.UUID, clearPayloadBytes []byte) ([]byte, error) {
	keyPool, _, decryptedJwsJwk, err := s.getAndDecryptKeyPoolJwk(ctx, &keyPoolID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt latest JWS JWK from Key Pool for KeyPoolID: %w", err)
	}
	_, err = cryptoutilJose.ExtractAlgFromJwsJwk(&decryptedJwsJwk, 0)
	if err != nil {
		return nil, fmt.Errorf("decrypted JWK doesn't have expected alg header: %w", err)
	}
	if keyPool.KeyPoolProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	_, jwsMessageBytes, err := cryptoutilJose.SignBytes([]joseJwk.Key{decryptedJwsJwk}, clearPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to sign bytes with latest Key for KeyPoolID: %w", err)
	}
	return jwsMessageBytes, nil
}

func (s *BusinessLogicService) PostVerifyByKeyPoolID(ctx context.Context, keyPoolID googleUuid.UUID, jwsMessageBytes []byte) ([]byte, error) {
	jwsMessage, err := joseJws.Parse(jwsMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWS message bytes: %w", err)
	}
	kidUuid, _, err := cryptoutilJose.ExtractKidAlgFromJwsMessage(jwsMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get kid, enc, and alg from JWE message: %w", err)
	}
	keyPool, _, decryptedJwsJwk, err := s.getAndDecryptKeyPoolJwk(ctx, &keyPoolID, kidUuid)
	// TODO validate decrypted JWK is a JWS JWK
	if keyPool.KeyPoolProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	_, err = cryptoutilJose.ExtractAlgFromJwsJwk(&decryptedJwsJwk, 0)
	if err != nil {
		return nil, fmt.Errorf("decrypted JWK doesn't have expected alg header: %w", err)
	}
	verifiedJwsMessageBytes, err := cryptoutilJose.VerifyBytes([]joseJwk.Key{decryptedJwsJwk}, jwsMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt bytes with Key for KeyPoolID from JWE kid UUID: %w", err)
	}
	return verifiedJwsMessageBytes, nil
}

func (s *BusinessLogicService) generateKeyPoolKeyForInsert(sqlTransaction *cryptoutilOrmRepository.OrmTransaction, keyPoolID googleUuid.UUID, keyPoolAlgorithm cryptoutilOrmRepository.KeyPoolAlgorithm) (*cryptoutilOrmRepository.Key, error) {
	var keyID *googleUuid.UUID
	var clearKeyMaterial []byte

	if s.serviceOrmMapper.isJwe(&keyPoolAlgorithm) {
		enc, alg, err := s.serviceOrmMapper.toJweEncAndAlg(&keyPoolAlgorithm)
		if err != nil {
			return nil, fmt.Errorf("failed to map JWE Key Pool Algorithm: %w", err)
		}
		keyID, _, clearKeyMaterial, err = s.jwkGenService.GenerateJweJwk(enc, alg)
		if err != nil {
			return nil, fmt.Errorf("failed to generate JWE: %w", err)
		}
	} else if s.serviceOrmMapper.isJws(&keyPoolAlgorithm) {
		alg, err := s.serviceOrmMapper.toJwsAlg(&keyPoolAlgorithm)
		if err != nil {
			return nil, fmt.Errorf("failed to map JWS Key Pool Algorithm: %w", err)
		}
		keyID, _, clearKeyMaterial, err = s.jwkGenService.GenerateJwsJwk(alg)
		if err != nil {
			return nil, fmt.Errorf("failed to generate JWS: %w", err)
		}
	} else {
		return nil, fmt.Errorf("unsupported KeyPoolAlgorithm %s", keyPoolAlgorithm)
	}

	repositoryKeyGenerateDate := time.Now().UTC()
	encryptedKeyMaterial, err := s.barrierService.EncryptContent(sqlTransaction, clearKeyMaterial)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt JWK: %w", err)
	}

	return &cryptoutilOrmRepository.Key{
		KeyPoolID:       keyPoolID,
		KeyID:           *keyID,
		KeyMaterial:     encryptedKeyMaterial,
		KeyGenerateDate: &repositoryKeyGenerateDate,
	}, nil
}

func (s *BusinessLogicService) getAndDecryptKeyPoolJwk(ctx context.Context, keyPoolID *googleUuid.UUID, kidUuid *googleUuid.UUID) (*cryptoutilOrmRepository.KeyPool, *cryptoutilOrmRepository.Key, joseJwk.Key, error) {
	var repositoryKeyPool *cryptoutilOrmRepository.KeyPool
	var repositoryKeyPoolKey *cryptoutilOrmRepository.Key
	var decryptedJwkBytes []byte
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
		decryptedJwkBytes, err = s.barrierService.DecryptContent(sqlTransaction, repositoryKeyPoolKey.KeyMaterial)
		if err != nil {
			return fmt.Errorf("failed to decrypt Key material from KeyPool: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get and decrypt Key material from KeyPool: %w", err)
	}
	decryptedJwk, err := joseJwk.ParseKey(decryptedJwkBytes)

	return repositoryKeyPool, repositoryKeyPoolKey, decryptedJwk, nil
}
