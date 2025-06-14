package businesslogic

import (
	"context"
	"encoding/json"
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

type keyExportableMaterial struct {
	public    *string
	decrypted *string
}

var (
	emptyKeyExportableMaterial = &keyExportableMaterial{}
)

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

	if repositoryKeyPoolToInsert.KeyPoolImportAllowed {
		return nil, fmt.Errorf("KeyPoolImportAllowed=true not supported yet")
	}

	// generate first key automatically
	keyID, _, _, encodedPrivateOrSecretJwk, _, err := s.generateJwk(&repositoryKeyPoolToInsert.KeyPoolAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to generate KeyPool Key: %w", err)
	}
	repositoryKeyGenerateDate := time.Now().UTC()

	var insertedKeyPool *cryptoutilOrmRepository.KeyPool
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		err := sqlTransaction.AddKeyPool(repositoryKeyPoolToInsert)
		if err != nil {
			return fmt.Errorf("failed to add KeyPool: %w", err)
		}

		err = TransitionState(cryptoutilBusinessLogicModel.Creating, cryptoutilBusinessLogicModel.KeyPoolStatus(repositoryKeyPoolToInsert.KeyPoolStatus))
		if err != nil {
			return fmt.Errorf("invalid KeyPoolStatus transition: %w", err)
		}

		encryptedKeyBytes, err := s.barrierService.EncryptContent(sqlTransaction, encodedPrivateOrSecretJwk)
		if err != nil {
			return fmt.Errorf("failed to encrypt KeyPool Key: %w", err)
		}

		repositoryKey := &cryptoutilOrmRepository.Key{
			KeyPoolID:       *keyPoolID,
			KeyID:           *keyID,
			KeyMaterial:     encryptedKeyBytes,          // nil if repositoryKeyPoolToInsert.KeyPoolImportAllowed=true
			KeyGenerateDate: &repositoryKeyGenerateDate, // nil if repositoryKeyPoolToInsert.KeyPoolImportAllowed=true
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
	var repositoryKeyPool *cryptoutilOrmRepository.KeyPool
	var repositoryKey *cryptoutilOrmRepository.Key
	var repositoryKeyExportableMaterial *keyExportableMaterial
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeyPool, err = sqlTransaction.GetKeyPool(keyPoolID)
		if err != nil {
			return fmt.Errorf("failed to get KeyPool by KeyPoolID: %w", err)
		}

		if repositoryKeyPool.KeyPoolStatus != cryptoutilOrmRepository.PendingGenerate && repositoryKeyPool.KeyPoolStatus != cryptoutilOrmRepository.Active {
			return fmt.Errorf("invalid KeyPoolStatus: %w", err)
		}

		keyID, _, _, clearPrivateOrSecretJwkBytes, clearPublicJwkBytes, err := s.generateJwk(&repositoryKeyPool.KeyPoolAlgorithm)
		if err != nil {
			return fmt.Errorf("failed to generate KeyPool Key: %w", err)
		}
		repositoryKeyGenerateDate := time.Now().UTC()

		encryptedPrivateOrPublicJwkBytes, err := s.barrierService.EncryptContent(sqlTransaction, clearPrivateOrSecretJwkBytes)
		if err != nil {
			return fmt.Errorf("failed to encrypt KeyPool Key: %w", err)
		}

		repositoryKey = &cryptoutilOrmRepository.Key{
			KeyPoolID:       keyPoolID,
			KeyID:           *keyID,
			KeyMaterial:     encryptedPrivateOrPublicJwkBytes,
			KeyGenerateDate: &repositoryKeyGenerateDate,
		}

		// TODO test publicKey and export
		repositoryKeyExportableMaterial = s.prepareKeyExportableMaterial(clearPublicJwkBytes, clearPrivateOrSecretJwkBytes, repositoryKeyPool)

		err = sqlTransaction.AddKeyPoolKey(repositoryKey)
		if err != nil {
			return fmt.Errorf("failed to insert Key: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in KeyPool: %w", err)
	}

	openapiPostKeypoolKeyPoolIDKeyResponseObject, err := s.serviceOrmMapper.toServiceKey(repositoryKey, repositoryKeyExportableMaterial)
	if err != nil {
		return nil, fmt.Errorf("failed to map key in KeyPool: %w", err)
	}

	return openapiPostKeypoolKeyPoolIDKeyResponseObject, nil
}

func (*BusinessLogicService) prepareKeyExportableMaterial(clearPublicBytes []byte, clearPrivateOrSecretBytes []byte, repositoryKeyPool *cryptoutilOrmRepository.KeyPool) *keyExportableMaterial {
	var public *string
	if cryptoutilOrmRepository.IsAsymmetric(&repositoryKeyPool.KeyPoolAlgorithm) && len(clearPublicBytes) > 0 {
		newVar := string(clearPublicBytes)
		public = &newVar
	}
	var decrypted *string
	if repositoryKeyPool.KeyPoolExportAllowed && len(clearPrivateOrSecretBytes) > 0 {
		newVar := string(clearPrivateOrSecretBytes)
		decrypted = &newVar
	}
	return &keyExportableMaterial{
		public:    public,
		decrypted: decrypted,
	}
}

func (s *BusinessLogicService) GetKeysByKeyPool(ctx context.Context, keyPoolID googleUuid.UUID, keyPoolKeysQueryParams *cryptoutilBusinessLogicModel.KeyPoolKeysQueryParams) ([]cryptoutilBusinessLogicModel.Key, error) {
	ormKeyPoolKeysQueryParams, err := s.serviceOrmMapper.toOrmGetKeyPoolKeysQueryParams(keyPoolKeysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid Get Key Pool Keys parameters: %w", err)
	}
	var repositoryKeyPool *cryptoutilOrmRepository.KeyPool
	var repositoryKeys []cryptoutilOrmRepository.Key
	var repositoryKeyExportableMaterials []*keyExportableMaterial
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeyPool, err = sqlTransaction.GetKeyPool(keyPoolID)
		if err != nil {
			return fmt.Errorf("failed to get KeyPool by KeyPoolID: %w", err)
		}

		repositoryKeys, err = sqlTransaction.GetKeyPoolKeys(keyPoolID, ormKeyPoolKeysQueryParams)
		if err != nil {
			return fmt.Errorf("failed to list Keys by KeyPoolID: %w", err)
		}

		// TODO test publicKey and export
		if cryptoutilOrmRepository.IsAsymmetric(&repositoryKeyPool.KeyPoolAlgorithm) || repositoryKeyPool.KeyPoolExportAllowed {
			// asymmetric => optionally export clear private key, and extract public key from it
			// symmetric => optionally export clear secret key
			for _, repositoryKey := range repositoryKeys {
				clearPrivateOrSecretJwkBytes, err := s.barrierService.DecryptContent(sqlTransaction, repositoryKey.KeyMaterial)
				if err != nil {
					return fmt.Errorf("failed to decrypt KeyPool Key: %w", err)
				}
				privateOrSecretJwk, err := joseJwk.ParseKey(clearPrivateOrSecretJwkBytes)
				if err != nil {
					return fmt.Errorf("failed to parse KeyPool Key: %w", err)
				}
				publicJwk, err := privateOrSecretJwk.PublicKey()
				if err != nil {
					return fmt.Errorf("failed to extract KeyPool Key public: %w", err)
				}
				clearPublicJwkBytes, err := json.Marshal(publicJwk)
				if err != nil {
					return fmt.Errorf("failed to encode KeyPool Key public: %w", err)
				}
				repositoryKeyExportableMaterial := s.prepareKeyExportableMaterial(clearPublicJwkBytes, clearPrivateOrSecretJwkBytes, repositoryKeyPool)
				repositoryKeyExportableMaterials = append(repositoryKeyExportableMaterials, repositoryKeyExportableMaterial)
			}
		} else {
			for range repositoryKeys {
				repositoryKeyExportableMaterials = append(repositoryKeyExportableMaterials, emptyKeyExportableMaterial)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in KeyPool: %w", err)
	}

	openapiPostKeypoolKeyPoolIDKeyResponseObjects, err := s.serviceOrmMapper.toServiceKeys(repositoryKeys, repositoryKeyExportableMaterials)
	if err != nil {
		return nil, fmt.Errorf("failed to map keys in KeyPool: %w", err)
	}

	return openapiPostKeypoolKeyPoolIDKeyResponseObjects, err
}

func (s *BusinessLogicService) GetKeys(ctx context.Context, keysQueryParams *cryptoutilBusinessLogicModel.KeysQueryParams) ([]cryptoutilBusinessLogicModel.Key, error) {
	ormKeysQueryParams, err := s.serviceOrmMapper.toOrmGetKeysQueryParams(keysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid Get Keys parameters: %w", err)
	}
	var repositoryKeyPool *cryptoutilOrmRepository.KeyPool
	var repositoryKeys []cryptoutilOrmRepository.Key
	var repositoryKeyExportableMaterials []*keyExportableMaterial
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		repositoryKeys, err = sqlTransaction.GetKeys(ormKeysQueryParams)
		if err != nil {
			return fmt.Errorf("failed to list Keys by KeyPoolID: %w", err)
		}

		for _, repositoryKey := range repositoryKeys {
			// TODO cache GetKeyPool
			repositoryKeyPool, err = sqlTransaction.GetKeyPool(repositoryKey.KeyPoolID)
			if err != nil {
				return fmt.Errorf("failed to get KeyPool by KeyPoolID: %w", err)
			}
			// TODO test publicKey and export
			if cryptoutilOrmRepository.IsAsymmetric(&repositoryKeyPool.KeyPoolAlgorithm) || repositoryKeyPool.KeyPoolExportAllowed {
				// asymmetric => optionally export clear private key, and extract public key from it
				// symmetric => optionally export clear secret key
				clearPrivateOrSecretJwkBytes, err := s.barrierService.DecryptContent(sqlTransaction, repositoryKey.KeyMaterial)
				if err != nil {
					return fmt.Errorf("failed to decrypt KeyPool Key: %w", err)
				}
				privateOrSecretJwk, err := joseJwk.ParseKey(clearPrivateOrSecretJwkBytes)
				if err != nil {
					return fmt.Errorf("failed to parse KeyPool Key: %w", err)
				}
				publicJwk, err := privateOrSecretJwk.PublicKey()
				if err != nil {
					return fmt.Errorf("failed to extract KeyPool Key public: %w", err)
				}
				clearPublicJwkBytes, err := json.Marshal(publicJwk)
				if err != nil {
					return fmt.Errorf("failed to encode KeyPool Key public: %w", err)
				}
				repositoryKeyExportableMaterial := s.prepareKeyExportableMaterial(clearPublicJwkBytes, clearPrivateOrSecretJwkBytes, repositoryKeyPool)
				repositoryKeyExportableMaterials = append(repositoryKeyExportableMaterials, repositoryKeyExportableMaterial)
			} else {
				repositoryKeyExportableMaterials = append(repositoryKeyExportableMaterials, emptyKeyExportableMaterial)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list keys in KeyPool: %w", err)
	}

	openapiPostKeypoolKeyPoolIDKeyResponseObjects, err := s.serviceOrmMapper.toServiceKeys(repositoryKeys, repositoryKeyExportableMaterials)
	if err != nil {
		return nil, fmt.Errorf("failed to map keys in KeyPool: %w", err)
	}

	return openapiPostKeypoolKeyPoolIDKeyResponseObjects, err
}

func (s *BusinessLogicService) GetKeyByKeyPoolAndKeyID(ctx context.Context, keyPoolID googleUuid.UUID, keyID googleUuid.UUID) (*cryptoutilBusinessLogicModel.Key, error) {
	var repositoryKeyPool *cryptoutilOrmRepository.KeyPool
	var repositoryKey *cryptoutilOrmRepository.Key
	var repositoryKeyExportableMaterial *keyExportableMaterial
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryKeyPool, err = sqlTransaction.GetKeyPool(repositoryKey.KeyPoolID)
		if err != nil {
			return fmt.Errorf("failed to get KeyPool by KeyPoolID: %w", err)
		}

		repositoryKey, err = sqlTransaction.GetKeyPoolKey(keyPoolID, keyID)
		if err != nil {
			return fmt.Errorf("failed to get Key by KeyPoolID and KeyID: %w", err)
		}

		// TODO test publicKey and export
		if cryptoutilOrmRepository.IsAsymmetric(&repositoryKeyPool.KeyPoolAlgorithm) || repositoryKeyPool.KeyPoolExportAllowed {
			// asymmetric => optionally export clear private key, and extract public key from it
			// symmetric => optionally export clear secret key
			clearPrivateOrSecretJwkBytes, err := s.barrierService.DecryptContent(sqlTransaction, repositoryKey.KeyMaterial)
			if err != nil {
				return fmt.Errorf("failed to decrypt KeyPool Key: %w", err)
			}
			privateOrSecretJwk, err := joseJwk.ParseKey(clearPrivateOrSecretJwkBytes)
			if err != nil {
				return fmt.Errorf("failed to parse KeyPool Key: %w", err)
			}
			publicJwk, err := privateOrSecretJwk.PublicKey()
			if err != nil {
				return fmt.Errorf("failed to extract KeyPool Key public: %w", err)
			}
			clearPublicJwkBytes, err := json.Marshal(publicJwk)
			if err != nil {
				return fmt.Errorf("failed to encode KeyPool Key public: %w", err)
			}
			repositoryKeyExportableMaterial = s.prepareKeyExportableMaterial(clearPublicJwkBytes, clearPrivateOrSecretJwkBytes, repositoryKeyPool)
		} else {
			repositoryKeyExportableMaterial = emptyKeyExportableMaterial
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in KeyPool: %w", err)
	}

	openapiPostKeypoolKeyPoolIDKeyResponseObject, err := s.serviceOrmMapper.toServiceKey(repositoryKey, repositoryKeyExportableMaterial)
	if err != nil {
		return nil, fmt.Errorf("failed to map keys in KeyPool: %w", err)
	}

	return openapiPostKeypoolKeyPoolIDKeyResponseObject, nil
}

func (s *BusinessLogicService) PostEncryptByKeyPoolID(ctx context.Context, keyPoolID googleUuid.UUID, encryptParams *cryptoutilBusinessLogicModel.EncryptParams, clearPayloadBytes []byte) ([]byte, error) {
	keyPool, _, decryptedJweJwk, err := s.getAndDecryptKeyPoolJwk(ctx, &keyPoolID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt latest JWE JWK for Key Pool: %w", err)
	}
	if keyPool.KeyPoolProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	// TODO Use encryptParams.Context for encryption
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
	if keyPool.KeyPoolProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	decryptedJweMessageBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{decryptedJweJwk}, jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt bytes with JWE kid UUID Key for KeyPoolID : %w", err)
	}
	return decryptedJweMessageBytes, nil
}

func (s *BusinessLogicService) PostSignByKeyPoolID(ctx context.Context, keyPoolID googleUuid.UUID, clearPayloadBytes []byte) ([]byte, error) {
	keyPool, _, decryptedJwsJwk, err := s.getAndDecryptKeyPoolJwk(ctx, &keyPoolID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt latest JWS JWK from Key Pool for KeyPoolID: %w", err)
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
		return nil, fmt.Errorf("failed to get kid and alg from JWS message: %w", err)
	}
	keyPool, _, decryptedJwsJwk, err := s.getAndDecryptKeyPoolJwk(ctx, &keyPoolID, kidUuid)
	// TODO validate decrypted JWK is a JWS JWK
	if keyPool.KeyPoolProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	verifiedJwsMessageBytes, err := cryptoutilJose.VerifyBytes([]joseJwk.Key{decryptedJwsJwk}, jwsMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to verify bytes with JWS kid UUID Key for KeyPoolID: %w", err)
	}
	return verifiedJwsMessageBytes, nil
}

func (s *BusinessLogicService) generateJwk(keyPoolAlgorithm *cryptoutilOrmRepository.KeyPoolAlgorithm) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	var keyID *googleUuid.UUID
	var privateOrSecretJwk joseJwk.Key
	var publicJwk joseJwk.Key
	var encodedPrivateOrSecretJwk []byte
	var encodedPublicJwk []byte

	if s.serviceOrmMapper.isJwe(keyPoolAlgorithm) {
		enc, alg, err := s.serviceOrmMapper.toJweEncAndAlg(keyPoolAlgorithm)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to map JWE Key Pool Algorithm: %w", err)
		}
		keyID, privateOrSecretJwk, publicJwk, encodedPrivateOrSecretJwk, encodedPublicJwk, err = s.jwkGenService.GenerateJweJwk(enc, alg)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to generate JWE: %w", err)
		}
	} else if s.serviceOrmMapper.isJws(keyPoolAlgorithm) {
		alg, err := s.serviceOrmMapper.toJwsAlg(keyPoolAlgorithm)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to map JWS Key Pool Algorithm: %w", err)
		}
		keyID, privateOrSecretJwk, publicJwk, encodedPrivateOrSecretJwk, encodedPublicJwk, err = s.jwkGenService.GenerateJwsJwk(alg)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to generate JWS: %w", err)
		}
	} else {
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported KeyPoolAlgorithm %v", keyPoolAlgorithm)
	}

	return keyID, privateOrSecretJwk, publicJwk, encodedPrivateOrSecretJwk, encodedPublicJwk, nil
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
			return fmt.Errorf("failed to decrypt Key from KeyPool: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get and decrypt Key from KeyPool: %w", err)
	}
	decryptedJwk, err := joseJwk.ParseKey(decryptedJwkBytes)

	return repositoryKeyPool, repositoryKeyPoolKey, decryptedJwk, nil
}
