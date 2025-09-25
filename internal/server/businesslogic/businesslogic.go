package businesslogic

import (
	"context"
	"fmt"
	"time"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilBarrierService "cryptoutil/internal/server/barrier"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"

	googleUuid "github.com/google/uuid"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
)

const (
	providerInternal = "Internal"
)

// BusinessLogicService implements methods in StrictServerInterface
type BusinessLogicService struct {
	telemetryService *cryptoutilTelemetry.TelemetryService
	jwkGenService    *cryptoutilJose.JWKGenService
	ormRepository    *cryptoutilOrmRepository.OrmRepository
	oamOrmMapper     *oamOrmMapper
	barrierService   *cryptoutilBarrierService.BarrierService
}

func NewBusinessLogicService(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, jwkGenService *cryptoutilJose.JWKGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, barrierService *cryptoutilBarrierService.BarrierService) (*BusinessLogicService, error) {
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
	materialKeyID, _, _, materialKeyClearNonPublicJWKBytes, materialKeyClearPublicJWKBytes, err := s.generateJWK(&ormElasticKey.ElasticKeyAlgorithm)
	if err != nil {
		return nil, fmt.Errorf("failed to generate first MaterialKey for ElasticKey : %w", err)
	}
	materialKeyGenerateDate := time.Now().UTC()

	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		err := sqlTransaction.AddElasticKey(ormElasticKey)
		if err != nil {
			return fmt.Errorf("failed to add ElasticKey: %w", err)
		}

		err = TransitionElasticKeyStatus(cryptoutilOpenapiModel.Creating, ormElasticKey.ElasticKeyStatus)
		if err != nil {
			return fmt.Errorf("invalid ElasticKeyStatus transition: %w", err)
		}

		materialKeyEncryptedNonPublicJWKBytes, err := s.barrierService.EncryptContent(sqlTransaction, materialKeyClearNonPublicJWKBytes)
		if err != nil {
			return fmt.Errorf("failed to encrypt first MaterialKey for ElasticKey: %w", err)
		}

		ormMaterialKey := s.oamOrmMapper.toOrmAddMaterialKey(elasticKeyID, materialKeyID, materialKeyClearPublicJWKBytes, materialKeyEncryptedNonPublicJWKBytes, materialKeyGenerateDate)

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

		materialKeyID, _, _, clearMaterialKeyNonPublicJWKBytes, clearPublicJWKBytes, err := s.generateJWK(&ormElasticKey.ElasticKeyAlgorithm)
		if err != nil {
			return fmt.Errorf("failed to generate new MaterialKey for ElasticKey: %w", err)
		}
		materialKeyGenerateDate := time.Now().UTC()

		encryptedMaterialKeyPrivateOrPublicJWKBytes, err := s.barrierService.EncryptContent(sqlTransaction, clearMaterialKeyNonPublicJWKBytes)
		if err != nil {
			return fmt.Errorf("failed to encrypt new MaterialKey for ElasticKey: %w", err)
		}

		ormMaterialKey = &cryptoutilOrmRepository.MaterialKey{
			ElasticKeyID:                  *elasticKeyID,
			MaterialKeyID:                 *materialKeyID,
			MaterialKeyClearPublic:        clearPublicJWKBytes,
			MaterialKeyEncryptedNonPublic: encryptedMaterialKeyPrivateOrPublicJWKBytes,
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

func (s *BusinessLogicService) PostGenerateByElasticKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID, generateParams *cryptoutilOpenapiModel.GenerateParams) ([]byte, []byte, []byte, error) {
	alg, err := cryptoutilJose.ToGenerateAlgorithm((*string)(generateParams.Alg))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to map generate algorithm: %w", err)
	}

	_, _, _, clearNonPublicJWKBytes, clearPublicJWKBytes, err := cryptoutilJose.GenerateJWKForAlg(alg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate key from algorithm: %w", err)
	}

	encryptedNonPublicJWKBytes, err := s.PostEncryptByElasticKeyID(ctx, elasticKeyID, &cryptoutilOpenapiModel.EncryptParams{Context: generateParams.Context}, clearNonPublicJWKBytes)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to encrypt generated key: %w", err)
	}

	return encryptedNonPublicJWKBytes, clearNonPublicJWKBytes, clearPublicJWKBytes, nil
}

func (s *BusinessLogicService) PostEncryptByElasticKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID, encryptParams *cryptoutilOpenapiModel.EncryptParams, clearPayloadBytes []byte) ([]byte, error) {
	elasticKey, _, decryptedMaterialKeyNonPublicJweJWK, clearMaterialKeyPublicJweJWK, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, elasticKeyID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt latest MaterialKey for ElasticKey: %w", err)
	}
	if elasticKey.ElasticKeyProvider != providerInternal {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	// TODO Use encryptParams.Context for encryption
	var jweMessageBytes []byte
	if clearMaterialKeyPublicJweJWK != nil {
		_, jweMessageBytes, err = cryptoutilJose.EncryptBytes([]joseJwk.Key{clearMaterialKeyPublicJweJWK}, clearPayloadBytes) // asymmetric
	} else {
		_, jweMessageBytes, err = cryptoutilJose.EncryptBytes([]joseJwk.Key{decryptedMaterialKeyNonPublicJweJWK}, clearPayloadBytes) // symmetric
	}
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
	elasticKey, _, decryptedMaterialKeyNonPublicJweJWK, _, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, elasticKeyID, materialKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt material key: %w", err)
	}
	if elasticKey.ElasticKeyProvider != providerInternal {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	} else if !cryptoutilJose.IsJwe(&elasticKey.ElasticKeyAlgorithm) {
		return nil, fmt.Errorf("decrypt not supported by KeyMaterial with ElasticKeyAlgorithm %v", elasticKey.ElasticKeyAlgorithm)
	}
	decryptedJweMessageBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{decryptedMaterialKeyNonPublicJweJWK}, jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt bytes with MaterialKey for ElasticKeyID : %w", err)
	}
	return decryptedJweMessageBytes, nil
}

func (s *BusinessLogicService) PostSignByElasticKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID, clearPayloadBytes []byte) ([]byte, error) {
	elasticKey, _, decryptedMaterialKeyNonPublicJWSJWK, _, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, elasticKeyID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt latest MaterialKey JWS JWK from ElasticKey for ElasticKeyID: %w", err)
	}
	if elasticKey.ElasticKeyProvider != providerInternal {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	_, jwsMessageBytes, err := cryptoutilJose.SignBytes([]joseJwk.Key{decryptedMaterialKeyNonPublicJWSJWK}, clearPayloadBytes)
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
	kidUUID, _, err := cryptoutilJose.ExtractKidAlgFromJWSMessage(jwsMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWS message headers kid and alg: %w", err)
	}
	elasticKey, _, decryptedMaterialKeyNonPublicJweJWK, clearMaterialKeyPublicJweJWK, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, elasticKeyID, kidUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt material key: %w", err)
	}
	if elasticKey.ElasticKeyProvider != providerInternal {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	} else if !cryptoutilJose.IsJWS(&elasticKey.ElasticKeyAlgorithm) {
		return nil, fmt.Errorf("verify not supported by KeyMaterial with ElasticKeyAlgorithm %v", elasticKey.ElasticKeyAlgorithm)
	}
	var verifiedJWSMessageBytes []byte
	if clearMaterialKeyPublicJweJWK != nil {
		verifiedJWSMessageBytes, err = cryptoutilJose.VerifyBytes([]joseJwk.Key{clearMaterialKeyPublicJweJWK}, jwsMessageBytes) // asymmetric
	} else {
		verifiedJWSMessageBytes, err = cryptoutilJose.VerifyBytes([]joseJwk.Key{decryptedMaterialKeyNonPublicJweJWK}, jwsMessageBytes) // symmetric
	}

	if err != nil {
		return nil, fmt.Errorf("failed to verify bytes with MaterialKey for ElasticKeyID: %w", err)
	}
	return verifiedJWSMessageBytes, nil
}

//nolint:unparam // Some callers ignore certain return values by design
func (s *BusinessLogicService) generateJWK(elasticKeyAlgorithm *cryptoutilOpenapiModel.ElasticKeyAlgorithm) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	var materialKeyID *googleUuid.UUID
	var materialKeyNonPublicJWK joseJwk.Key
	var materialKeyPublicJWK joseJwk.Key
	var materialKeyNonPublicJWKBytes []byte
	var materialKeyPublicJWKBytes []byte

	if cryptoutilJose.IsJwe(elasticKeyAlgorithm) {
		enc, alg, err := cryptoutilJose.ToJweEncAndAlg(elasticKeyAlgorithm)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to map ElasticKeyAlgorithm: %w", err)
		}
		materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, err = s.jwkGenService.GenerateJweJWK(enc, alg)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to generate MaterialKey JWE JWK: %w", err)
		}
	} else if cryptoutilJose.IsJWS(elasticKeyAlgorithm) {
		alg, err := cryptoutilJose.ToJWSAlg(elasticKeyAlgorithm)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to map JWS ElasticKey Algorithm: %w", err)
		}
		materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, err = s.jwkGenService.GenerateJWSJWK(alg)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to generate MaterialKey JWS JWK: %w", err)
		}
	} else {
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported ElasticKeyAlgorithm %v", elasticKeyAlgorithm)
	}

	return materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, nil
}

//nolint:unparam // Some callers ignore certain return values by design
func (s *BusinessLogicService) getAndDecryptMaterialKeyInElasticKey(ctx context.Context, elasticKeyID *googleUuid.UUID, materialKeyKidUUID *googleUuid.UUID) (*cryptoutilOrmRepository.ElasticKey, *cryptoutilOrmRepository.MaterialKey, joseJwk.Key, joseJwk.Key, error) {
	var ormElasticKey *cryptoutilOrmRepository.ElasticKey
	var ormMaterialKey *cryptoutilOrmRepository.MaterialKey
	var materialKeyDecryptedNonPublicJWKBytes []byte
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		ormElasticKey, err = sqlTransaction.GetElasticKey(elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
		}
		if materialKeyKidUUID == nil {
			ormMaterialKey, err = sqlTransaction.GetElasticKeyMaterialKeyLatest(*elasticKeyID)
		} else {
			ormMaterialKey, err = sqlTransaction.GetElasticKeyMaterialKeyVersion(elasticKeyID, materialKeyKidUUID)
		}
		if err != nil {
			return fmt.Errorf("failed to get MaterialKey in ElasticKey: %w", err)
		}
		materialKeyDecryptedNonPublicJWKBytes, err = s.barrierService.DecryptContent(sqlTransaction, ormMaterialKey.MaterialKeyEncryptedNonPublic)
		if err != nil {
			return fmt.Errorf("failed to decrypt MaterialKeyEncryptedNonPublic in ElasticKey: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get and decrypt MaterialKeyEncryptedNonPublic in ElasticKey: %w", err)
	}
	decryptedMaterialKeyNonPublicJWK, err := joseJwk.ParseKey(materialKeyDecryptedNonPublicJWKBytes)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to parse decrypted MaterialKeyEncryptedNonPublic: %w", err)
	}

	var clearMaterialKeyPublicJWK joseJwk.Key
	if ormMaterialKey.MaterialKeyClearPublic != nil {
		clearMaterialKeyPublicJWK, err = joseJwk.ParseKey(ormMaterialKey.MaterialKeyClearPublic)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to parse MaterialKeyClearPublic: %w", err)
		}
	}

	return ormElasticKey, ormMaterialKey, decryptedMaterialKeyNonPublicJWK, clearMaterialKeyPublicJWK, nil
}
