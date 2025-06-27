package businesslogic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cryptoutil/internal/common/businessmodel"
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

// TODO Remove clearPublic
type materialKeyExport struct {
	clearPublic    *string
	clearNonPublic *string
}

var (
	emptyMaterialKeyExport = &materialKeyExport{}
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

func (s *BusinessLogicService) AddElasticKey(ctx context.Context, openapiElasticKeyCreate *cryptoutilBusinessLogicModel.ElasticKeyCreate) (*cryptoutilBusinessLogicModel.ElasticKey, error) {
	elasticKeyID := s.jwkGenService.GenerateUUIDv7()
	repositoryElasticKeyToInsert := s.serviceOrmMapper.toOrmAddElasticKey(*elasticKeyID, openapiElasticKeyCreate)

	if repositoryElasticKeyToInsert.ElasticKeyImportAllowed {
		return nil, fmt.Errorf("ElasticKeyImportAllowed=true not supported yet")
	}

	// generate first Material Key automatically
	materialKeyID, _, _, clearMaterialKeyNonPublicJwkBytes, clearMaterialKeyPublicJwkBytes, err := s.generateJwk(&repositoryElasticKeyToInsert.ElasticKeyAlgorithm)
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

		err = businessmodel.EkasticKeyStatusTransition(cryptoutilBusinessLogicModel.Creating, cryptoutilBusinessLogicModel.ElasticKeyStatus(repositoryElasticKeyToInsert.ElasticKeyStatus))
		if err != nil {
			return fmt.Errorf("invalid ElasticKeyStatus transition: %w", err)
		}

		encryptedKeyBytes, err := s.barrierService.EncryptContent(sqlTransaction, clearMaterialKeyNonPublicJwkBytes)
		if err != nil {
			return fmt.Errorf("failed to encrypt ElasticKey Key: %w", err)
		}

		repositoryKey := &cryptoutilOrmRepository.MaterialKey{
			ElasticKeyID:                  *elasticKeyID,
			MaterialKeyID:                 *materialKeyID,
			ClearPublicKeyMaterial:        clearMaterialKeyPublicJwkBytes,     // nil if repositoryElasticKeyToInsert.ElasticKeyAlgorithm is Symmetric
			EncryptedNonPublicKeyMaterial: encryptedKeyBytes,                  // nil if repositoryElasticKeyToInsert.ElasticKeyImportAllowed=true
			MaterialKeyGenerateDate:       &repositoryMaterialKeyGenerateDate, // nil if repositoryElasticKeyToInsert.ElasticKeyImportAllowed=true
		}

		err = sqlTransaction.AddElasticKeyKey(repositoryKey)
		if err != nil {
			return fmt.Errorf("failed to add key: %w", err)
		}

		err = sqlTransaction.UpdateElasticKeyStatus(*elasticKeyID, businessmodel.Active)
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

	return s.serviceOrmMapper.toServiceElasticKey(insertedElasticKey), nil
}

func (s *BusinessLogicService) GetElasticKeyByElasticKeyID(ctx context.Context, elasticKeyID googleUuid.UUID) (*cryptoutilBusinessLogicModel.ElasticKey, error) {
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
	return s.serviceOrmMapper.toServiceElasticKey(repositoryElasticKey), nil
}

func (s *BusinessLogicService) GetElasticKeys(ctx context.Context, elasticKeyQueryParams *cryptoutilBusinessLogicModel.ElasticKeysQueryParams) ([]cryptoutilBusinessLogicModel.ElasticKey, error) {
	ormElasticKeysQueryParams, err := s.serviceOrmMapper.toOrmGetElasticKeysQueryParams(elasticKeyQueryParams)
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
	return s.serviceOrmMapper.toServiceElasticKeys(repositoryElasticKeys), nil
}

func (s *BusinessLogicService) GenerateKeyInPoolKey(ctx context.Context, elasticKeyID googleUuid.UUID, _ *cryptoutilBusinessLogicModel.MaterialKeyGenerate) (*cryptoutilBusinessLogicModel.MaterialKey, error) {
	var repositoryElasticKey *cryptoutilOrmRepository.ElasticKey
	var repositoryMaterialKey *cryptoutilOrmRepository.MaterialKey
	var repositoryExportableMaterialKeyDetails *materialKeyExport
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryElasticKey, err = sqlTransaction.GetElasticKey(elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get Elastic Key by ElasticKeyID: %w", err)
		}

		if repositoryElasticKey.ElasticKeyStatus != businessmodel.PendingGenerate && repositoryElasticKey.ElasticKeyStatus != businessmodel.Active {
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
			ClearPublicKeyMaterial:        clearPublicJwkBytes,
			EncryptedNonPublicKeyMaterial: encryptedMaterialKeyPrivateOrPublicJwkBytes,
			MaterialKeyGenerateDate:       &materialKeyGenerateDate,
		}

		repositoryExportableMaterialKeyDetails = s.prepareMaterialKeyExportableDetails(clearPublicJwkBytes, clearMaterialKeyNonPublicJwkBytes, repositoryElasticKey)

		err = sqlTransaction.AddElasticKeyKey(repositoryMaterialKey)
		if err != nil {
			return fmt.Errorf("failed to insert new Material Key for Elastic Key: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate new Material Key for Elastic Key: %w", err)
	}

	openapiPostElastickeyElasticKeyIDMaterialkeyResponseObject, err := s.serviceOrmMapper.toServiceKey(repositoryMaterialKey, repositoryExportableMaterialKeyDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to map new Material Key for ElasticKey: %w", err)
	}

	return openapiPostElastickeyElasticKeyIDMaterialkeyResponseObject, nil
}

func (*BusinessLogicService) prepareMaterialKeyExportableDetails(clearPublicBytes []byte, clearNonPublicBytes []byte, repositoryElasticKey *cryptoutilOrmRepository.ElasticKey) *materialKeyExport {
	var clearPublicStringPointer *string
	if businessmodel.IsAsymmetric(&repositoryElasticKey.ElasticKeyAlgorithm) && len(clearPublicBytes) > 0 {
		clearPublicString := string(clearPublicBytes)
		clearPublicStringPointer = &clearPublicString
	}
	// TODO Only allow this option in dev mode, not prod mode
	var clearNonPublicStringPointer *string
	if repositoryElasticKey.ElasticKeyExportAllowed && len(clearNonPublicBytes) > 0 {
		clearNonPublicString := string(clearNonPublicBytes)
		clearNonPublicStringPointer = &clearNonPublicString
	}
	return &materialKeyExport{clearPublic: clearPublicStringPointer, clearNonPublic: clearNonPublicStringPointer}
}

func (s *BusinessLogicService) GetMaterialKeysForElasticKey(ctx context.Context, elasticKeyID googleUuid.UUID, elasticKeyMaterialKeysQueryParams *cryptoutilBusinessLogicModel.ElasticKeyMaterialKeysQueryParams) ([]cryptoutilBusinessLogicModel.MaterialKey, error) {
	ormElasticKeyMaterialKeysQueryParams, err := s.serviceOrmMapper.toOrmGetMaterialKeysForElasticKeyQueryParams(elasticKeyMaterialKeysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to map Material Keys for Elastic Key query parameters: %w", err)
	}
	var repositoryElasticKey *cryptoutilOrmRepository.ElasticKey
	var repositoryMaterialKeys []cryptoutilOrmRepository.MaterialKey
	var repositoryMaterialKeyExportDetails []*materialKeyExport
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryElasticKey, err = sqlTransaction.GetElasticKey(elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get Elastic Key by ElasticKeyID: %w", err)
		}

		repositoryMaterialKeys, err = sqlTransaction.GetMaterialKeysForElasticKey(elasticKeyID, ormElasticKeyMaterialKeysQueryParams)
		if err != nil {
			return fmt.Errorf("failed to list Material Keys by ElasticKeyID: %w", err)
		}

		if businessmodel.IsAsymmetric(&repositoryElasticKey.ElasticKeyAlgorithm) || repositoryElasticKey.ElasticKeyExportAllowed {
			// asymmetric => always export clear public key, optionally export clear private key
			// symmetric => optionally export clear secret key
			for _, repositoryKey := range repositoryMaterialKeys {
				clearMaterialKeyNonPublicJwkBytes, err := s.barrierService.DecryptContent(sqlTransaction, repositoryKey.EncryptedNonPublicKeyMaterial)
				if err != nil {
					return fmt.Errorf("failed to decrypt Material Key for Elastic Key: %w", err)
				}
				clearMaterialKeyNonPublicJwk, err := joseJwk.ParseKey(clearMaterialKeyNonPublicJwkBytes)
				if err != nil {
					return fmt.Errorf("failed to parse Material Key for Elastic Key: %w", err)
				}
				clearMaterialKeyPublicJwk, err := clearMaterialKeyNonPublicJwk.PublicKey()
				if err != nil {
					return fmt.Errorf("failed to extract Material Key for Elastic Key public: %w", err)
				}
				clearMaterialKeyPublicJwkBytes, err := json.Marshal(clearMaterialKeyPublicJwk)
				if err != nil {
					return fmt.Errorf("failed to encode Material Key for Elastic Key public: %w", err)
				}
				repositoryMaterialKeyExportDetail := s.prepareMaterialKeyExportableDetails(clearMaterialKeyPublicJwkBytes, clearMaterialKeyNonPublicJwkBytes, repositoryElasticKey)
				repositoryMaterialKeyExportDetails = append(repositoryMaterialKeyExportDetails, repositoryMaterialKeyExportDetail)
			}
		} else {
			for range repositoryMaterialKeys {
				repositoryMaterialKeyExportDetails = append(repositoryMaterialKeyExportDetails, emptyMaterialKeyExport)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Material Key for Elastic Key: %w", err)
	}

	openapiPostElastickeyElasticKeyIDMaterialkeyResponseObjects, err := s.serviceOrmMapper.toServiceKeys(repositoryMaterialKeys, repositoryMaterialKeyExportDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to map Material Key for Elastic Key: %w", err)
	}

	return openapiPostElastickeyElasticKeyIDMaterialkeyResponseObjects, err
}

func (s *BusinessLogicService) GetMaterialKeys(ctx context.Context, keysQueryParams *cryptoutilBusinessLogicModel.MaterialKeysQueryParams) ([]cryptoutilBusinessLogicModel.MaterialKey, error) {
	ormMaterialKeysQueryParams, err := s.serviceOrmMapper.toOrmGetMaterialKeysQueryParams(keysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid map Material Keys query parameters: %w", err)
	}
	var repositoryElasticKey *cryptoutilOrmRepository.ElasticKey
	var repositoryKeys []cryptoutilOrmRepository.MaterialKey
	var repositoryKeyExportableMaterials []*materialKeyExport
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
			}
			if businessmodel.IsAsymmetric(&repositoryElasticKey.ElasticKeyAlgorithm) || repositoryElasticKey.ElasticKeyExportAllowed {
				// asymmetric => optionally export clear private key, and extract public key from it
				// symmetric => optionally export clear secret key
				clearNonPublicJwkBytes, err := s.barrierService.DecryptContent(sqlTransaction, repositoryKey.EncryptedNonPublicKeyMaterial)
				if err != nil {
					return fmt.Errorf("failed to decrypt ElasticKey Key: %w", err)
				}
				nonPublicJwk, err := joseJwk.ParseKey(clearNonPublicJwkBytes)
				if err != nil {
					return fmt.Errorf("failed to parse ElasticKey Key: %w", err)
				}
				publicJwk, err := nonPublicJwk.PublicKey()
				if err != nil {
					return fmt.Errorf("failed to extract ElasticKey Key public: %w", err)
				}
				clearPublicJwkBytes, err := json.Marshal(publicJwk)
				if err != nil {
					return fmt.Errorf("failed to encode ElasticKey Key public: %w", err)
				}
				repositoryKeyExportableMaterial := s.prepareMaterialKeyExportableDetails(clearPublicJwkBytes, clearNonPublicJwkBytes, repositoryElasticKey)
				repositoryKeyExportableMaterials = append(repositoryKeyExportableMaterials, repositoryKeyExportableMaterial)
			} else {
				repositoryKeyExportableMaterials = append(repositoryKeyExportableMaterials, emptyMaterialKeyExport)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list keys in ElasticKey: %w", err)
	}

	openapiPostElastickeyElasticKeyIDMaterialkeyResponseObjects, err := s.serviceOrmMapper.toServiceKeys(repositoryKeys, repositoryKeyExportableMaterials)
	if err != nil {
		return nil, fmt.Errorf("failed to map keys in ElasticKey: %w", err)
	}

	return openapiPostElastickeyElasticKeyIDMaterialkeyResponseObjects, err
}

func (s *BusinessLogicService) GetMaterialKeyByElasticKeyAndMaterialKeyID(ctx context.Context, elasticKeyID googleUuid.UUID, materialKeyID googleUuid.UUID) (*cryptoutilBusinessLogicModel.MaterialKey, error) {
	var repositoryElasticKey *cryptoutilOrmRepository.ElasticKey
	var repositoryMaterialKey *cryptoutilOrmRepository.MaterialKey
	var repositoryMaterialKeyExportMaterial *materialKeyExport
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryElasticKey, err = sqlTransaction.GetElasticKey(repositoryMaterialKey.ElasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
		}

		repositoryMaterialKey, err = sqlTransaction.GetElasticKeyMaterialKeyVersion(elasticKeyID, materialKeyID)
		if err != nil {
			return fmt.Errorf("failed to get Key by ElasticKeyID and MaterialKeyID: %w", err)
		}

		if businessmodel.IsAsymmetric(&repositoryElasticKey.ElasticKeyAlgorithm) || repositoryElasticKey.ElasticKeyExportAllowed {
			// asymmetric => always export clear public key, optionally export clear private key
			// symmetric => optionally export clear secret key
			clearNonPublicJwkBytes, err := s.barrierService.DecryptContent(sqlTransaction, repositoryMaterialKey.EncryptedNonPublicKeyMaterial)
			if err != nil {
				return fmt.Errorf("failed to decrypt Material Key for Elastic Key: %w", err)
			}
			nonPublicJwk, err := joseJwk.ParseKey(clearNonPublicJwkBytes)
			if err != nil {
				return fmt.Errorf("failed to parse Material Key for Elastic Key: %w", err)
			}
			publicJwk, err := nonPublicJwk.PublicKey()
			if err != nil {
				return fmt.Errorf("failed to extract Material Key for Elastic Key public: %w", err)
			}
			clearPublicJwkBytes, err := json.Marshal(publicJwk)
			if err != nil {
				return fmt.Errorf("failed to encode Material Key for Elastic Key public: %w", err)
			}
			repositoryMaterialKeyExportMaterial = s.prepareMaterialKeyExportableDetails(clearPublicJwkBytes, clearNonPublicJwkBytes, repositoryElasticKey)
		} else {
			repositoryMaterialKeyExportMaterial = emptyMaterialKeyExport
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get Material Key for Elastic Key: %w", err)
	}

	openapiPostElastickeyElasticKeyIDMaterialkeyResponseObject, err := s.serviceOrmMapper.toServiceKey(repositoryMaterialKey, repositoryMaterialKeyExportMaterial)
	if err != nil {
		return nil, fmt.Errorf("failed to map Material Key for Elastic Key: %w", err)
	}

	return openapiPostElastickeyElasticKeyIDMaterialkeyResponseObject, nil
}

func (s *BusinessLogicService) PostEncryptByElasticKeyID(ctx context.Context, elasticKeyID googleUuid.UUID, encryptParams *cryptoutilBusinessLogicModel.EncryptParams, clearPayloadBytes []byte) ([]byte, error) {
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

func (s *BusinessLogicService) generateJwk(elasticKeyAlgorithm *businessmodel.ElasticKeyAlgorithm) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
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
		materialKeyJwkBytes, err = s.barrierService.DecryptContent(sqlTransaction, repositoryMaterialKey.EncryptedNonPublicKeyMaterial)
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
