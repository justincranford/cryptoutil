package businesslogic

import (
	"context"
	"fmt"
	"time"

	cryptoutilBusinessModel "cryptoutil/internal/common/businessmodel"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"
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
	oamOrmMapper     *oamOrmMapper
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
		oamOrmMapper:     NewOamOrmMapper(),
		barrierService:   barrierService,
	}, nil
}

func (s *BusinessLogicService) AddElasticKey(ctx context.Context, openapiElasticKeyCreate *cryptoutilOpenapiModel.ElasticKeyCreate) (*cryptoutilOpenapiModel.ElasticKey, error) {
	elasticKeyID := s.jwkGenService.GenerateUUIDv7()
	repositoryElasticKeyToInsert := s.oamOrmMapper.toOrmAddElasticKey(*elasticKeyID, openapiElasticKeyCreate)

	if repositoryElasticKeyToInsert.ElasticKeyImportAllowed {
		return nil, fmt.Errorf("ElasticKeyImportAllowed=true not supported yet")
	}

	// generate first Material Key automatically
	materialKeyID, _, _, clearMaterialKeyNonPublicJwkBytes, materialKeyClearPublicJwkBytes, err := s.generateJwk(&repositoryElasticKeyToInsert.ElasticKeyAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ElasticKey Key: %w", err)
	}
	repositoryMaterialKeyGenerateDate := time.Now().UTC()

	var insertedElasticKey *cryptoutilOrmRepository.ElasticKey
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		err := sqlTransaction.AddElasticKey(repositoryElasticKeyToInsert)
		if err != nil {
			return fmt.Errorf("failed to add ElasticKey: %w", err)
		}

		err = TransitionElasticKeyStatus(cryptoutilOpenapiModel.Creating, cryptoutilOpenapiModel.ElasticKeyStatus(repositoryElasticKeyToInsert.ElasticKeyStatus))
		if err != nil {
			return fmt.Errorf("invalid ElasticKeyStatus transition: %w", err)
		}

		materialKeyEncryptedNonPublicJwkBytes, err := s.barrierService.EncryptContent(sqlTransaction, clearMaterialKeyNonPublicJwkBytes)
		if err != nil {
			return fmt.Errorf("failed to encrypt ElasticKey Key: %w", err)
		}

		repositoryMaterialKey := &cryptoutilOrmRepository.MaterialKey{
			ElasticKeyID:                  *elasticKeyID,
			MaterialKeyID:                 *materialKeyID,
			MaterialKeyClearPublic:        materialKeyClearPublicJwkBytes,        // nil if repositoryElasticKeyToInsert.ElasticKeyAlgorithm is Symmetric
			MaterialKeyEncryptedNonPublic: materialKeyEncryptedNonPublicJwkBytes, // nil if repositoryElasticKeyToInsert.ElasticKeyImportAllowed=true
			MaterialKeyGenerateDate:       &repositoryMaterialKeyGenerateDate,    // nil if repositoryElasticKeyToInsert.ElasticKeyImportAllowed=true
		}

		err = sqlTransaction.AddElasticKeyKey(repositoryMaterialKey)
		if err != nil {
			return fmt.Errorf("failed to add key: %w", err)
		}

		err = sqlTransaction.UpdateElasticKeyStatus(*elasticKeyID, cryptoutilBusinessModel.Active)
		if err != nil {
			return fmt.Errorf("failed to update ElasticKeyStatus to active: %w", err)
		}

		insertedElasticKey, err = sqlTransaction.GetElasticKey(*elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get updated ElasticKey from DB: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add elastic Key: %w", err)
	}

	return s.oamOrmMapper.toOamElasticKey(insertedElasticKey), nil
}

func (s *BusinessLogicService) GetElasticKeyByElasticKeyID(ctx context.Context, elasticKeyID googleUuid.UUID) (*cryptoutilOpenapiModel.ElasticKey, error) {
	var repositoryElasticKey *cryptoutilOrmRepository.ElasticKey
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryElasticKey, err = sqlTransaction.GetElasticKey(elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ElasticKey: %w", err)
	}
	return s.oamOrmMapper.toOamElasticKey(repositoryElasticKey), nil
}

func (s *BusinessLogicService) GetElasticKeys(ctx context.Context, elasticKeyQueryParams *cryptoutilOpenapiModel.ElasticKeysQueryParams) ([]cryptoutilOpenapiModel.ElasticKey, error) {
	ormElasticKeysQueryParams, err := s.oamOrmMapper.toOrmGetElasticKeysQueryParams(elasticKeyQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid Get Elastic Keys parameters: %w", err)
	}
	var repositoryElasticKeys []cryptoutilOrmRepository.ElasticKey
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryElasticKeys, err = sqlTransaction.GetElasticKeys(ormElasticKeysQueryParams)
		if err != nil {
			return fmt.Errorf("failed to list ElasticKeys: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list ElasticKeys: %w", err)
	}
	return s.oamOrmMapper.toOamElasticKeys(repositoryElasticKeys), nil
}

func (s *BusinessLogicService) GenerateKeyInPoolKey(ctx context.Context, elasticKeyID googleUuid.UUID, _ *cryptoutilOpenapiModel.MaterialKeyGenerate) (*cryptoutilOpenapiModel.MaterialKey, error) {
	var repositoryElasticKey *cryptoutilOrmRepository.ElasticKey
	var repositoryMaterialKey *cryptoutilOrmRepository.MaterialKey
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryElasticKey, err = sqlTransaction.GetElasticKey(elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get Elastic Key by ElasticKeyID: %w", err)
		}

		if repositoryElasticKey.ElasticKeyStatus != cryptoutilBusinessModel.PendingGenerate && repositoryElasticKey.ElasticKeyStatus != cryptoutilBusinessModel.Active {
			return fmt.Errorf("invalid Elastic Key Status: %w", err)
		}

		materialKeyKidUuid, _, _, clearMaterialKeyNonPublicJwkBytes, clearPublicJwkBytes, err := s.generateJwk(&repositoryElasticKey.ElasticKeyAlgorithm)
		if err != nil {
			return fmt.Errorf("failed to generate new Material Key for Elastic Key: %w", err)
		}
		materialKeyGenerateDate := time.Now().UTC()

		encryptedMaterialKeyPrivateOrPublicJwkBytes, err := s.barrierService.EncryptContent(sqlTransaction, clearMaterialKeyNonPublicJwkBytes)
		if err != nil {
			return fmt.Errorf("failed to encrypt new Material Key for Elastic Key: %w", err)
		}

		repositoryMaterialKey = &cryptoutilOrmRepository.MaterialKey{
			ElasticKeyID:                  elasticKeyID,
			MaterialKeyID:                 *materialKeyKidUuid,
			MaterialKeyClearPublic:        clearPublicJwkBytes,
			MaterialKeyEncryptedNonPublic: encryptedMaterialKeyPrivateOrPublicJwkBytes,
			MaterialKeyGenerateDate:       &materialKeyGenerateDate,
		}

		err = sqlTransaction.AddElasticKeyKey(repositoryMaterialKey)
		if err != nil {
			return fmt.Errorf("failed to insert new Material Key for Elastic Key: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate new Material Key for Elastic Key: %w", err)
	}

	openapiPostElastickeyElasticKeyIDMaterialkeyResponseObject, err := s.oamOrmMapper.toOamKey(repositoryMaterialKey)
	if err != nil {
		return nil, fmt.Errorf("failed to map new Material Key for ElasticKey: %w", err)
	}

	return openapiPostElastickeyElasticKeyIDMaterialkeyResponseObject, nil
}

func (s *BusinessLogicService) GetMaterialKeysForElasticKey(ctx context.Context, elasticKeyID googleUuid.UUID, elasticKeyMaterialKeysQueryParams *cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams) ([]cryptoutilOpenapiModel.MaterialKey, error) {
	ormElasticKeyMaterialKeysQueryParams, err := s.oamOrmMapper.toOrmGetMaterialKeysForElasticKeyQueryParams(elasticKeyMaterialKeysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to map Material Keys for Elastic Key query parameters: %w", err)
	}
	var repositoryElasticKey *cryptoutilOrmRepository.ElasticKey
	var repositoryMaterialKeys []cryptoutilOrmRepository.MaterialKey
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryElasticKey, err = sqlTransaction.GetElasticKey(elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get Elastic Key by ElasticKeyID: %w", err)
		} else if repositoryElasticKey == nil {
			return fmt.Errorf("got nil Elastic Key by ElasticKeyID: %w", err)
		}
		repositoryMaterialKeys, err = sqlTransaction.GetMaterialKeysForElasticKey(elasticKeyID, ormElasticKeyMaterialKeysQueryParams)
		if err != nil {
			return fmt.Errorf("failed to list Material Keys by ElasticKeyID: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Material Key for Elastic Key: %w", err)
	}

	openapiPostElastickeyElasticKeyIDMaterialkeyResponseObjects, err := s.oamOrmMapper.toOamKeys(repositoryMaterialKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to map Material Key for Elastic Key: %w", err)
	}

	return openapiPostElastickeyElasticKeyIDMaterialkeyResponseObjects, err
}

func (s *BusinessLogicService) GetMaterialKeys(ctx context.Context, keysQueryParams *cryptoutilOpenapiModel.MaterialKeysQueryParams) ([]cryptoutilOpenapiModel.MaterialKey, error) {
	ormMaterialKeysQueryParams, err := s.oamOrmMapper.toOrmGetMaterialKeysQueryParams(keysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid map Material Keys query parameters: %w", err)
	}
	var repositoryElasticKey *cryptoutilOrmRepository.ElasticKey
	var repositoryKeys []cryptoutilOrmRepository.MaterialKey
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		repositoryKeys, err = sqlTransaction.GetMaterialKeys(ormMaterialKeysQueryParams)
		if err != nil {
			return fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
		}

		for _, repositoryKey := range repositoryKeys {
			// TODO cache GetElasticKey
			repositoryElasticKey, err = sqlTransaction.GetElasticKey(repositoryKey.ElasticKeyID)
			if err != nil {
				return fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
			} else if repositoryElasticKey == nil {
				return fmt.Errorf("got nil Elastic Key by ElasticKeyID: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list keys in ElasticKey: %w", err)
	}

	openapiPostElastickeyElasticKeyIDMaterialkeyResponseObjects, err := s.oamOrmMapper.toOamKeys(repositoryKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to map keys in ElasticKey: %w", err)
	}

	return openapiPostElastickeyElasticKeyIDMaterialkeyResponseObjects, err
}

func (s *BusinessLogicService) GetMaterialKeyByElasticKeyAndMaterialKeyID(ctx context.Context, elasticKeyID googleUuid.UUID, materialKeyID googleUuid.UUID) (*cryptoutilOpenapiModel.MaterialKey, error) {
	var repositoryElasticKey *cryptoutilOrmRepository.ElasticKey
	var repositoryMaterialKey *cryptoutilOrmRepository.MaterialKey
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryElasticKey, err = sqlTransaction.GetElasticKey(repositoryMaterialKey.ElasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
		} else if repositoryElasticKey == nil {
			return fmt.Errorf("got nil Elastic Key by ElasticKeyID: %w", err)
		}

		repositoryMaterialKey, err = sqlTransaction.GetElasticKeyMaterialKeyVersion(elasticKeyID, materialKeyID)
		if err != nil {
			return fmt.Errorf("failed to get Key by ElasticKeyID and MaterialKeyID: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Material Key for Elastic Key: %w", err)
	}

	openapiPostElastickeyElasticKeyIDMaterialkeyResponseObject, err := s.oamOrmMapper.toOamKey(repositoryMaterialKey)
	if err != nil {
		return nil, fmt.Errorf("failed to map Material Key for Elastic Key: %w", err)
	}

	return openapiPostElastickeyElasticKeyIDMaterialkeyResponseObject, nil
}

func (s *BusinessLogicService) PostEncryptByElasticKeyID(ctx context.Context, elasticKeyID googleUuid.UUID, encryptParams *cryptoutilOpenapiModel.EncryptParams, clearPayloadBytes []byte) ([]byte, error) {
	elasticKey, _, decryptedJweJwk, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, &elasticKeyID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt latest Material Key for Elastic Key: %w", err)
	}
	if elasticKey.ElasticKeyProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	// TODO Use encryptParams.Context for encryption
	_, jweMessageBytes, err := cryptoutilJose.EncryptBytes([]joseJwk.Key{decryptedJweJwk}, clearPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt bytes with latest Material Key for ElasticKeyID: %w", err)
	}
	return jweMessageBytes, nil
}

func (s *BusinessLogicService) PostDecryptByElasticKeyID(ctx context.Context, elasticKeyID googleUuid.UUID, jweMessageBytes []byte) ([]byte, error) {
	jweMessage, err := joseJwe.Parse(jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message bytes: %w", err)
	}
	kidUuid, _, _, err := cryptoutilJose.ExtractKidEncAlgFromJweMessage(jweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWE message headers kid, enc, and alg: %w", err)
	}
	elasticKey, _, decryptedMaterialKeyJwk, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, &elasticKeyID, kidUuid)
	if elasticKey.ElasticKeyProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	decryptedJweMessageBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{decryptedMaterialKeyJwk}, jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt bytes with Material Key for ElasticKeyID : %w", err)
	}
	return decryptedJweMessageBytes, nil
}

func (s *BusinessLogicService) PostSignByElasticKeyID(ctx context.Context, elasticKeyID googleUuid.UUID, clearPayloadBytes []byte) ([]byte, error) {
	elasticKey, _, decryptedJwsJwk, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, &elasticKeyID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt latest Material Key JWS JWK from Elastic Key for ElasticKeyID: %w", err)
	}
	if elasticKey.ElasticKeyProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	_, jwsMessageBytes, err := cryptoutilJose.SignBytes([]joseJwk.Key{decryptedJwsJwk}, clearPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to sign bytes with latest Material Key for ElasticKeyID: %w", err)
	}
	return jwsMessageBytes, nil
}

func (s *BusinessLogicService) PostVerifyByElasticKeyID(ctx context.Context, elasticKeyID googleUuid.UUID, jwsMessageBytes []byte) ([]byte, error) {
	jwsMessage, err := joseJws.Parse(jwsMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWS message bytes: %w", err)
	}
	kidUuid, _, err := cryptoutilJose.ExtractKidAlgFromJwsMessage(jwsMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWS message headers kid and alg: %w", err)
	}
	elasticKey, _, decryptedJwsJwk, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, &elasticKeyID, kidUuid)
	// TODO validate decrypted JWK is a JWS JWK
	if elasticKey.ElasticKeyProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	verifiedJwsMessageBytes, err := cryptoutilJose.VerifyBytes([]joseJwk.Key{decryptedJwsJwk}, jwsMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to verify bytes with Mateiral Key for ElasticKeyID: %w", err)
	}
	return verifiedJwsMessageBytes, nil
}

func (s *BusinessLogicService) generateJwk(elasticKeyAlgorithm *cryptoutilBusinessModel.ElasticKeyAlgorithm) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	var materialKeyID *googleUuid.UUID
	var materialKeyNonPublicJwk joseJwk.Key
	var materialKeyPublicJwk joseJwk.Key
	var materialKeyNonPublicJwkBytes []byte
	var materialKeyPublicJwkBytes []byte

	if cryptoutilJose.IsJwe(elasticKeyAlgorithm) {
		enc, alg, err := cryptoutilJose.ToJweEncAndAlg(elasticKeyAlgorithm)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to map Elastic Key Algorithm: %w", err)
		}
		materialKeyID, materialKeyNonPublicJwk, materialKeyPublicJwk, materialKeyNonPublicJwkBytes, materialKeyPublicJwkBytes, err = s.jwkGenService.GenerateJweJwk(enc, alg)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to generate Material Key JWE JWK: %w", err)
		}
	} else if cryptoutilJose.IsJws(elasticKeyAlgorithm) {
		alg, err := cryptoutilJose.ToJwsAlg(elasticKeyAlgorithm)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to map JWS Elastic Key Algorithm: %w", err)
		}
		materialKeyID, materialKeyNonPublicJwk, materialKeyPublicJwk, materialKeyNonPublicJwkBytes, materialKeyPublicJwkBytes, err = s.jwkGenService.GenerateJwsJwk(alg)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to generate Material Key JWS JWK: %w", err)
		}
	} else {
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported ElasticKeyAlgorithm %v", elasticKeyAlgorithm)
	}

	return materialKeyID, materialKeyNonPublicJwk, materialKeyPublicJwk, materialKeyNonPublicJwkBytes, materialKeyPublicJwkBytes, nil
}

func (s *BusinessLogicService) getAndDecryptMaterialKeyInElasticKey(ctx context.Context, elasticKeyID *googleUuid.UUID, materialKeyKidUuid *googleUuid.UUID) (*cryptoutilOrmRepository.ElasticKey, *cryptoutilOrmRepository.MaterialKey, joseJwk.Key, error) {
	var repositoryElasticKey *cryptoutilOrmRepository.ElasticKey
	var repositoryMaterialKey *cryptoutilOrmRepository.MaterialKey
	var materialKeyJwkBytes []byte
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryElasticKey, err = sqlTransaction.GetElasticKey(*elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey by Elastic Key ID: %w", err)
		}
		if materialKeyKidUuid == nil {
			repositoryMaterialKey, err = sqlTransaction.GetElasticKeyMaterialKeyLatest(*elasticKeyID)
			if err != nil {
				return fmt.Errorf("failed to get latest Material Key in ElasticKey: %w", err)
			}
		} else {
			repositoryMaterialKey, err = sqlTransaction.GetElasticKeyMaterialKeyVersion(*elasticKeyID, *materialKeyKidUuid)
			if err != nil {
				return fmt.Errorf("failed to get versioned Material Key in ElasticKey: %w", err)
			}
		}
		materialKeyJwkBytes, err = s.barrierService.DecryptContent(sqlTransaction, repositoryMaterialKey.MaterialKeyEncryptedNonPublic)
		if err != nil {
			return fmt.Errorf("failed to decrypt Material Key in ElasticKey: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get and decrypt Material Key in ElasticKey: %w", err)
	}
	decryptedMaterialKeyJwk, err := joseJwk.ParseKey(materialKeyJwkBytes)

	return repositoryElasticKey, repositoryMaterialKey, decryptedMaterialKeyJwk, nil
}
