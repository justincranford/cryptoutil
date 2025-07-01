package businesslogic

import (
	"context"
	"fmt"
	"time"

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
	ormElasticKey := s.oamOrmMapper.toOrmAddElasticKey(elasticKeyID, openapiElasticKeyCreate)

	if ormElasticKey.ElasticKeyImportAllowed {
		return nil, fmt.Errorf("ElasticKeyImportAllowed=true not supported yet")
	}

	// generate first MaterialKey automatically
	materialKeyID, _, _, materialKeyClearNonPublicJwkBytes, materialKeyClearPublicJwkBytes, err := s.generateJwk(&ormElasticKey.ElasticKeyAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to generate first MaterialKey for ElasticKey : %w", err)
	}
	materialKeyGenerateDate := time.Now().UTC()

	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		err := sqlTransaction.AddElasticKey(ormElasticKey)
		if err != nil {
			return fmt.Errorf("failed to add ElasticKey: %w", err)
		}

		err = TransitionElasticKeyStatus(cryptoutilOpenapiModel.Creating, cryptoutilOpenapiModel.ElasticKeyStatus(ormElasticKey.ElasticKeyStatus))
		if err != nil {
			return fmt.Errorf("invalid ElasticKeyStatus transition: %w", err)
		}

		materialKeyEncryptedNonPublicJwkBytes, err := s.barrierService.EncryptContent(sqlTransaction, materialKeyClearNonPublicJwkBytes)
		if err != nil {
			return fmt.Errorf("failed to encrypt first MaterialKey for ElasticKey: %w", err)
		}

		ormMaterialKey := s.oamOrmMapper.toOrmAddMaterialKey(elasticKeyID, materialKeyID, materialKeyClearPublicJwkBytes, materialKeyEncryptedNonPublicJwkBytes, materialKeyGenerateDate)

		err = sqlTransaction.AddElasticKeyMaterialKey(ormMaterialKey)
		if err != nil {
			return fmt.Errorf("failed to add first MaterialKey for ElasticKey: %w", err)
		}

		err = sqlTransaction.UpdateElasticKeyStatus(*elasticKeyID, cryptoutilOpenapiModel.Active)
		if err != nil {
			return fmt.Errorf("failed to update ElasticKeyStatus to active: %w", err)
		}

		ormElasticKey, err = sqlTransaction.GetElasticKey(elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get updated ElasticKey: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add ElasticKey: %w", err)
	}

	return s.oamOrmMapper.toOamElasticKey(ormElasticKey), nil
}

func (s *BusinessLogicService) GetElasticKeyByElasticKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID) (*cryptoutilOpenapiModel.ElasticKey, error) {
	var ormElasticKey *cryptoutilOrmRepository.ElasticKey
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		ormElasticKey, err = sqlTransaction.GetElasticKey(elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ElasticKey: %w", err)
	}
	return s.oamOrmMapper.toOamElasticKey(ormElasticKey), nil
}

func (s *BusinessLogicService) GetElasticKeys(ctx context.Context, elasticKeyQueryParams *cryptoutilOpenapiModel.ElasticKeysQueryParams) ([]cryptoutilOpenapiModel.ElasticKey, error) {
	ormElasticKeysQueryParams, err := s.oamOrmMapper.toOrmGetElasticKeysQueryParams(elasticKeyQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid ElasticKeysQueryParams: %w", err)
	}
	var ormElasticKeys []cryptoutilOrmRepository.ElasticKey
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		ormElasticKeys, err = sqlTransaction.GetElasticKeys(ormElasticKeysQueryParams)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKeys: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ElasticKeys: %w", err)
	}
	return s.oamOrmMapper.toOamElasticKeys(ormElasticKeys), nil
}

func (s *BusinessLogicService) GenerateMaterialKeyInElasticKey(ctx context.Context, elasticKeyID *googleUuid.UUID, _ *cryptoutilOpenapiModel.MaterialKeyGenerate) (*cryptoutilOpenapiModel.MaterialKey, error) {
	var ormElasticKey *cryptoutilOrmRepository.ElasticKey
	var ormMaterialKey *cryptoutilOrmRepository.MaterialKey
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		ormElasticKey, err = sqlTransaction.GetElasticKey(elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
		}

		if ormElasticKey.ElasticKeyStatus != cryptoutilOpenapiModel.PendingGenerate && ormElasticKey.ElasticKeyStatus != cryptoutilOpenapiModel.Active {
			return fmt.Errorf("invalid ElasticKey Status: %w", err)
		}

		materialKeyID, _, _, clearMaterialKeyNonPublicJwkBytes, clearPublicJwkBytes, err := s.generateJwk(&ormElasticKey.ElasticKeyAlgorithm)
		if err != nil {
			return fmt.Errorf("failed to generate new MaterialKey for ElasticKey: %w", err)
		}
		materialKeyGenerateDate := time.Now().UTC()

		encryptedMaterialKeyPrivateOrPublicJwkBytes, err := s.barrierService.EncryptContent(sqlTransaction, clearMaterialKeyNonPublicJwkBytes)
		if err != nil {
			return fmt.Errorf("failed to encrypt new MaterialKey for ElasticKey: %w", err)
		}

		ormMaterialKey = &cryptoutilOrmRepository.MaterialKey{
			ElasticKeyID:                  *elasticKeyID,
			MaterialKeyID:                 *materialKeyID,
			MaterialKeyClearPublic:        clearPublicJwkBytes,
			MaterialKeyEncryptedNonPublic: encryptedMaterialKeyPrivateOrPublicJwkBytes,
			MaterialKeyGenerateDate:       &materialKeyGenerateDate,
		}

		err = sqlTransaction.AddElasticKeyMaterialKey(ormMaterialKey)
		if err != nil {
			return fmt.Errorf("failed to insert new MaterialKey for ElasticKey: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate new MaterialKey for ElasticKey: %w", err)
	}

	oamMaterialKey, err := s.oamOrmMapper.toOamMaterialKey(ormMaterialKey)
	if err != nil {
		return nil, fmt.Errorf("failed to map new MaterialKey for ElasticKey: %w", err)
	}

	return oamMaterialKey, nil
}

func (s *BusinessLogicService) GetMaterialKeysForElasticKey(ctx context.Context, elasticKeyID *googleUuid.UUID, elasticKeyMaterialKeysQueryParams *cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams) ([]cryptoutilOpenapiModel.MaterialKey, error) {
	ormElasticKeyMaterialKeysQueryParams, err := s.oamOrmMapper.toOrmGetMaterialKeysForElasticKeyQueryParams(elasticKeyMaterialKeysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to map MaterialKeys for ElasticKey query parameters: %w", err)
	}
	var ormElasticKey *cryptoutilOrmRepository.ElasticKey
	var ormMaterialKeys []cryptoutilOrmRepository.MaterialKey
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		ormElasticKey, err = sqlTransaction.GetElasticKey(elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
		} else if ormElasticKey == nil {
			return fmt.Errorf("got nil ElasticKey by ElasticKeyID: %w", err)
		}
		ormMaterialKeys, err = sqlTransaction.GetMaterialKeysForElasticKey(elasticKeyID, ormElasticKeyMaterialKeysQueryParams)
		if err != nil {
			return fmt.Errorf("failed to get MaterialKeys by ElasticKeyID: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get MaterialKey for ElasticKey: %w", err)
	}

	oamMaterialKeys, err := s.oamOrmMapper.toOamMaterialKeys(ormMaterialKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to map MaterialKey for ElasticKey: %w", err)
	}

	return oamMaterialKeys, err
}

func (s *BusinessLogicService) GetMaterialKeys(ctx context.Context, materialKeysQueryParams *cryptoutilOpenapiModel.MaterialKeysQueryParams) ([]cryptoutilOpenapiModel.MaterialKey, error) {
	ormMaterialKeysQueryParams, err := s.oamOrmMapper.toOrmGetMaterialKeysQueryParams(materialKeysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid MaterialKeysQueryParams: %w", err)
	}
	var ormElasticKey *cryptoutilOrmRepository.ElasticKey
	var ormMaterialKeys []cryptoutilOrmRepository.MaterialKey
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		ormMaterialKeys, err = sqlTransaction.GetMaterialKeys(ormMaterialKeysQueryParams)
		if err != nil {
			return fmt.Errorf("failed to get MaterialKeys by ElasticKeyID: %w", err)
		}

		for _, ormMaterialKey := range ormMaterialKeys {
			// TODO cache GetElasticKey
			ormElasticKey, err = sqlTransaction.GetElasticKey(&ormMaterialKey.ElasticKeyID)
			if err != nil {
				return fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
			} else if ormElasticKey == nil {
				return fmt.Errorf("got nil ElasticKey by ElasticKeyID: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get MaterialKeys in ElasticKey: %w", err)
	}

	oamMaterialKeys, err := s.oamOrmMapper.toOamMaterialKeys(ormMaterialKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to map MaterialKeys in ElasticKey: %w", err)
	}

	return oamMaterialKeys, err
}

func (s *BusinessLogicService) GetMaterialKeyByElasticKeyAndMaterialKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID, materialKeyID *googleUuid.UUID) (*cryptoutilOpenapiModel.MaterialKey, error) {
	var ormMaterialKey *cryptoutilOrmRepository.MaterialKey
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		_, err = sqlTransaction.GetElasticKey(&ormMaterialKey.ElasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
		}

		ormMaterialKey, err = sqlTransaction.GetElasticKeyMaterialKeyVersion(elasticKeyID, materialKeyID)
		if err != nil {
			return fmt.Errorf("failed to get MaterialKeys by ElasticKeyID and MaterialKeyID: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get MaterialKey for ElasticKey: %w", err)
	}

	openapiPostElastickeyElasticKeyIDMaterialkeyResponseObject, err := s.oamOrmMapper.toOamMaterialKey(ormMaterialKey)
	if err != nil {
		return nil, fmt.Errorf("failed to map MaterialKey for ElasticKey: %w", err)
	}

	return openapiPostElastickeyElasticKeyIDMaterialkeyResponseObject, nil
}

func (s *BusinessLogicService) PostGenerateByElasticKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID, generateParams *cryptoutilOpenapiModel.GenerateParams) ([]byte, error) {
	_, err := cryptoutilJose.ToGenerateAlgorithm((*string)(generateParams.Alg))
	if err != nil {
		return nil, fmt.Errorf("failed to map generate algorithm: %w", err)
	}

	// TODO generate JWK
	clearPayloadBytes := []byte{}

	return s.PostEncryptByElasticKeyID(ctx, elasticKeyID, &cryptoutilOpenapiModel.EncryptParams{Context: generateParams.Context}, clearPayloadBytes)
}

func (s *BusinessLogicService) PostEncryptByElasticKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID, encryptParams *cryptoutilOpenapiModel.EncryptParams, clearPayloadBytes []byte) ([]byte, error) {
	elasticKey, _, decryptedMaterialKeyNonPublicJwsJwk, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, elasticKeyID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt latest MaterialKey for ElasticKey: %w", err)
	}
	if elasticKey.ElasticKeyProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	// TODO Use encryptParams.Context for encryption
	_, jweMessageBytes, err := cryptoutilJose.EncryptBytes([]joseJwk.Key{decryptedMaterialKeyNonPublicJwsJwk}, clearPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt bytes with latest MaterialKey for ElasticKeyID: %w", err)
	}
	return jweMessageBytes, nil
}

func (s *BusinessLogicService) PostDecryptByElasticKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID, jweMessageBytes []byte) ([]byte, error) {
	jweMessage, err := joseJwe.Parse(jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message bytes: %w", err)
	}
	materialKeyID, err := cryptoutilJose.ExtractKidFromJweMessage(jweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWE message header kid: %w", err)
	}
	elasticKey, _, decryptedMaterialKeyNonPublicJwsJwk, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, elasticKeyID, materialKeyID)
	if elasticKey.ElasticKeyProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	} else if !cryptoutilJose.IsJwe(&elasticKey.ElasticKeyAlgorithm) {
		return nil, fmt.Errorf("decrypt not supported by KeyMaterial with ElasticKeyAlgorithm %v", elasticKey.ElasticKeyAlgorithm)
	}
	decryptedJweMessageBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{decryptedMaterialKeyNonPublicJwsJwk}, jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt bytes with MaterialKey for ElasticKeyID : %w", err)
	}
	return decryptedJweMessageBytes, nil
}

func (s *BusinessLogicService) PostSignByElasticKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID, clearPayloadBytes []byte) ([]byte, error) {
	elasticKey, _, decryptedMaterialKeyNonPublicJwsJwk, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, elasticKeyID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt latest MaterialKey JWS JWK from ElasticKey for ElasticKeyID: %w", err)
	}
	if elasticKey.ElasticKeyProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	_, jwsMessageBytes, err := cryptoutilJose.SignBytes([]joseJwk.Key{decryptedMaterialKeyNonPublicJwsJwk}, clearPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to sign bytes with latest MaterialKey for ElasticKeyID: %w", err)
	}
	return jwsMessageBytes, nil
}

func (s *BusinessLogicService) PostVerifyByElasticKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID, jwsMessageBytes []byte) ([]byte, error) {
	jwsMessage, err := joseJws.Parse(jwsMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWS message bytes: %w", err)
	}
	kidUuid, _, err := cryptoutilJose.ExtractKidAlgFromJwsMessage(jwsMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWS message headers kid and alg: %w", err)
	}
	elasticKey, _, decryptedMaterialKeyNonPublicJwsJwk, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, elasticKeyID, kidUuid)
	if elasticKey.ElasticKeyProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	} else if !cryptoutilJose.IsJws(&elasticKey.ElasticKeyAlgorithm) {
		return nil, fmt.Errorf("verify not supported by KeyMaterial with ElasticKeyAlgorithm %v", elasticKey.ElasticKeyAlgorithm)
	}
	verifiedJwsMessageBytes, err := cryptoutilJose.VerifyBytes([]joseJwk.Key{decryptedMaterialKeyNonPublicJwsJwk}, jwsMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to verify bytes with MateiralKey for ElasticKeyID: %w", err)
	}
	return verifiedJwsMessageBytes, nil
}

func (s *BusinessLogicService) generateJwk(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	var materialKeyID *googleUuid.UUID
	var materialKeyNonPublicJwk joseJwk.Key
	var materialKeyPublicJwk joseJwk.Key
	var materialKeyNonPublicJwkBytes []byte
	var materialKeyPublicJwkBytes []byte

	if cryptoutilJose.IsJwe(elasticKeyAlgorithm) {
		enc, alg, err := cryptoutilJose.ToJweEncAndAlg(elasticKeyAlgorithm)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to map ElasticKeyAlgorithm: %w", err)
		}
		materialKeyID, materialKeyNonPublicJwk, materialKeyPublicJwk, materialKeyNonPublicJwkBytes, materialKeyPublicJwkBytes, err = s.jwkGenService.GenerateJweJwk(enc, alg)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to generate MaterialKey JWE JWK: %w", err)
		}
	} else if cryptoutilJose.IsJws(elasticKeyAlgorithm) {
		alg, err := cryptoutilJose.ToJwsAlg(elasticKeyAlgorithm)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to map JWS ElasticKey Algorithm: %w", err)
		}
		materialKeyID, materialKeyNonPublicJwk, materialKeyPublicJwk, materialKeyNonPublicJwkBytes, materialKeyPublicJwkBytes, err = s.jwkGenService.GenerateJwsJwk(alg)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to generate MaterialKey JWS JWK: %w", err)
		}
	} else {
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported ElasticKeyAlgorithm %v", elasticKeyAlgorithm)
	}

	return materialKeyID, materialKeyNonPublicJwk, materialKeyPublicJwk, materialKeyNonPublicJwkBytes, materialKeyPublicJwkBytes, nil
}

func (s *BusinessLogicService) getAndDecryptMaterialKeyInElasticKey(ctx context.Context, elasticKeyID *googleUuid.UUID, materialKeyKidUuid *googleUuid.UUID) (*cryptoutilOrmRepository.ElasticKey, *cryptoutilOrmRepository.MaterialKey, joseJwk.Key, error) {
	var ormElasticKey *cryptoutilOrmRepository.ElasticKey
	var ormMaterialKey *cryptoutilOrmRepository.MaterialKey
	var materialKeyDecryptedNonPublicJwkBytes []byte
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		ormElasticKey, err = sqlTransaction.GetElasticKey(elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
		}
		if materialKeyKidUuid == nil {
			ormMaterialKey, err = sqlTransaction.GetElasticKeyMaterialKeyLatest(*elasticKeyID)
		} else {
			ormMaterialKey, err = sqlTransaction.GetElasticKeyMaterialKeyVersion(elasticKeyID, materialKeyKidUuid)
		}
		if err != nil {
			return fmt.Errorf("failed to get MaterialKey in ElasticKey: %w", err)
		}
		materialKeyDecryptedNonPublicJwkBytes, err = s.barrierService.DecryptContent(sqlTransaction, ormMaterialKey.MaterialKeyEncryptedNonPublic)
		if err != nil {
			return fmt.Errorf("failed to decrypt MaterialKey in ElasticKey: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get and decrypt MaterialKey in ElasticKey: %w", err)
	}
	decryptedMaterialKeyNonPublicJwk, err := joseJwk.ParseKey(materialKeyDecryptedNonPublicJwkBytes)

	return ormElasticKey, ormMaterialKey, decryptedMaterialKeyNonPublicJwk, nil
}
