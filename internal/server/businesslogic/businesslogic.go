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

func (s *BusinessLogicService) AddElasticKey(ctx context.Context, openapiElasticKeyCreate *cryptoutilBusinessLogicModel.ElasticKeyCreate) (*cryptoutilBusinessLogicModel.ElasticKey, error) {
	elasticKeyID := s.jwkGenService.GenerateUUIDv7()
	repositoryElasticKeyToInsert := s.serviceOrmMapper.toOrmAddElasticKey(*elasticKeyID, openapiElasticKeyCreate)

	if repositoryElasticKeyToInsert.ElasticKeyImportAllowed {
		return nil, fmt.Errorf("ElasticKeyImportAllowed=true not supported yet")
	}

	// generate first key automatically
	materialKeyID, _, _, encodedPrivateOrSecretJwk, _, err := s.generateJwk(&repositoryElasticKeyToInsert.ElasticKeyAlgorithm)
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

		err = TransitionState(cryptoutilBusinessLogicModel.Creating, cryptoutilBusinessLogicModel.ElasticKeyStatus(repositoryElasticKeyToInsert.ElasticKeyStatus))
		if err != nil {
			return fmt.Errorf("invalid ElasticKeyStatus transition: %w", err)
		}

		encryptedKeyBytes, err := s.barrierService.EncryptContent(sqlTransaction, encodedPrivateOrSecretJwk)
		if err != nil {
			return fmt.Errorf("failed to encrypt ElasticKey Key: %w", err)
		}

		repositoryKey := &cryptoutilOrmRepository.MaterialKey{
			ElasticKeyID:            *elasticKeyID,
			MaterialKeyID:           *materialKeyID,
			KeyMaterial:             encryptedKeyBytes,                  // nil if repositoryElasticKeyToInsert.ElasticKeyImportAllowed=true
			MaterialKeyGenerateDate: &repositoryMaterialKeyGenerateDate, // nil if repositoryElasticKeyToInsert.ElasticKeyImportAllowed=true
		}

		err = sqlTransaction.AddElasticKeyKey(repositoryKey)
		if err != nil {
			return fmt.Errorf("failed to add key: %w", err)
		}

		err = sqlTransaction.UpdateElasticKeyStatus(*elasticKeyID, cryptoutilOrmRepository.Active)
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
	var repositoryKey *cryptoutilOrmRepository.MaterialKey
	var repositoryKeyExportableMaterial *keyExportableMaterial
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryElasticKey, err = sqlTransaction.GetElasticKey(elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
		}

		if repositoryElasticKey.ElasticKeyStatus != cryptoutilOrmRepository.PendingGenerate && repositoryElasticKey.ElasticKeyStatus != cryptoutilOrmRepository.Active {
			return fmt.Errorf("invalid ElasticKeyStatus: %w", err)
		}

		materialKeyID, _, _, clearPrivateOrSecretJwkBytes, clearPublicJwkBytes, err := s.generateJwk(&repositoryElasticKey.ElasticKeyAlgorithm)
		if err != nil {
			return fmt.Errorf("failed to generate ElasticKey Key: %w", err)
		}
		repositoryMaterialKeyGenerateDate := time.Now().UTC()

		encryptedPrivateOrPublicJwkBytes, err := s.barrierService.EncryptContent(sqlTransaction, clearPrivateOrSecretJwkBytes)
		if err != nil {
			return fmt.Errorf("failed to encrypt ElasticKey Key: %w", err)
		}

		repositoryKey = &cryptoutilOrmRepository.MaterialKey{
			ElasticKeyID:            elasticKeyID,
			MaterialKeyID:           *materialKeyID,
			KeyMaterial:             encryptedPrivateOrPublicJwkBytes,
			MaterialKeyGenerateDate: &repositoryMaterialKeyGenerateDate,
		}

		// TODO test publicKey and export
		repositoryKeyExportableMaterial = s.prepareKeyExportableMaterial(clearPublicJwkBytes, clearPrivateOrSecretJwkBytes, repositoryElasticKey)

		err = sqlTransaction.AddElasticKeyKey(repositoryKey)
		if err != nil {
			return fmt.Errorf("failed to insert Key: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in ElasticKey: %w", err)
	}

	openapiPostElastickeyElasticKeyIDMaterialkeyResponseObject, err := s.serviceOrmMapper.toServiceKey(repositoryKey, repositoryKeyExportableMaterial)
	if err != nil {
		return nil, fmt.Errorf("failed to map key in ElasticKey: %w", err)
	}

	return openapiPostElastickeyElasticKeyIDMaterialkeyResponseObject, nil
}

func (*BusinessLogicService) prepareKeyExportableMaterial(clearPublicBytes []byte, clearPrivateOrSecretBytes []byte, repositoryElasticKey *cryptoutilOrmRepository.ElasticKey) *keyExportableMaterial {
	var public *string
	if cryptoutilOrmRepository.IsAsymmetric(&repositoryElasticKey.ElasticKeyAlgorithm) && len(clearPublicBytes) > 0 {
		newVar := string(clearPublicBytes)
		public = &newVar
	}
	var decrypted *string
	if repositoryElasticKey.ElasticKeyExportAllowed && len(clearPrivateOrSecretBytes) > 0 {
		newVar := string(clearPrivateOrSecretBytes)
		decrypted = &newVar
	}
	return &keyExportableMaterial{
		public:    public,
		decrypted: decrypted,
	}
}

func (s *BusinessLogicService) GetMaterialKeysByElasticKey(ctx context.Context, elasticKeyID googleUuid.UUID, elasticKeyMaterialKeysQueryParams *cryptoutilBusinessLogicModel.ElasticKeyMaterialKeysQueryParams) ([]cryptoutilBusinessLogicModel.MaterialKey, error) {
	ormElasticKeyMaterialKeysQueryParams, err := s.serviceOrmMapper.toOrmGetElasticKeyMaterialKeysQueryParams(elasticKeyMaterialKeysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid Get Elastic Key Keys parameters: %w", err)
	}
	var repositoryElasticKey *cryptoutilOrmRepository.ElasticKey
	var repositoryKeys []cryptoutilOrmRepository.MaterialKey
	var repositoryKeyExportableMaterials []*keyExportableMaterial
	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryElasticKey, err = sqlTransaction.GetElasticKey(elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
		}

		repositoryKeys, err = sqlTransaction.GetElasticKeyKeys(elasticKeyID, ormElasticKeyMaterialKeysQueryParams)
		if err != nil {
			return fmt.Errorf("failed to list Keys by ElasticKeyID: %w", err)
		}

		// TODO test publicKey and export
		if cryptoutilOrmRepository.IsAsymmetric(&repositoryElasticKey.ElasticKeyAlgorithm) || repositoryElasticKey.ElasticKeyExportAllowed {
			// asymmetric => optionally export clear private key, and extract public key from it
			// symmetric => optionally export clear secret key
			for _, repositoryKey := range repositoryKeys {
				clearPrivateOrSecretJwkBytes, err := s.barrierService.DecryptContent(sqlTransaction, repositoryKey.KeyMaterial)
				if err != nil {
					return fmt.Errorf("failed to decrypt ElasticKey Key: %w", err)
				}
				privateOrSecretJwk, err := joseJwk.ParseKey(clearPrivateOrSecretJwkBytes)
				if err != nil {
					return fmt.Errorf("failed to parse ElasticKey Key: %w", err)
				}
				publicJwk, err := privateOrSecretJwk.PublicKey()
				if err != nil {
					return fmt.Errorf("failed to extract ElasticKey Key public: %w", err)
				}
				clearPublicJwkBytes, err := json.Marshal(publicJwk)
				if err != nil {
					return fmt.Errorf("failed to encode ElasticKey Key public: %w", err)
				}
				repositoryKeyExportableMaterial := s.prepareKeyExportableMaterial(clearPublicJwkBytes, clearPrivateOrSecretJwkBytes, repositoryElasticKey)
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
		return nil, fmt.Errorf("failed to generate key in ElasticKey: %w", err)
	}

	openapiPostElastickeyElasticKeyIDMaterialkeyResponseObjects, err := s.serviceOrmMapper.toServiceKeys(repositoryKeys, repositoryKeyExportableMaterials)
	if err != nil {
		return nil, fmt.Errorf("failed to map keys in ElasticKey: %w", err)
	}

	return openapiPostElastickeyElasticKeyIDMaterialkeyResponseObjects, err
}

func (s *BusinessLogicService) GetMaterialKeys(ctx context.Context, keysQueryParams *cryptoutilBusinessLogicModel.MaterialKeysQueryParams) ([]cryptoutilBusinessLogicModel.MaterialKey, error) {
	ormMaterialKeysQueryParams, err := s.serviceOrmMapper.toOrmGetMaterialKeysQueryParams(keysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("invalid Get Keys parameters: %w", err)
	}
	var repositoryElasticKey *cryptoutilOrmRepository.ElasticKey
	var repositoryKeys []cryptoutilOrmRepository.MaterialKey
	var repositoryKeyExportableMaterials []*keyExportableMaterial
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
			// TODO test publicKey and export
			if cryptoutilOrmRepository.IsAsymmetric(&repositoryElasticKey.ElasticKeyAlgorithm) || repositoryElasticKey.ElasticKeyExportAllowed {
				// asymmetric => optionally export clear private key, and extract public key from it
				// symmetric => optionally export clear secret key
				clearPrivateOrSecretJwkBytes, err := s.barrierService.DecryptContent(sqlTransaction, repositoryKey.KeyMaterial)
				if err != nil {
					return fmt.Errorf("failed to decrypt ElasticKey Key: %w", err)
				}
				privateOrSecretJwk, err := joseJwk.ParseKey(clearPrivateOrSecretJwkBytes)
				if err != nil {
					return fmt.Errorf("failed to parse ElasticKey Key: %w", err)
				}
				publicJwk, err := privateOrSecretJwk.PublicKey()
				if err != nil {
					return fmt.Errorf("failed to extract ElasticKey Key public: %w", err)
				}
				clearPublicJwkBytes, err := json.Marshal(publicJwk)
				if err != nil {
					return fmt.Errorf("failed to encode ElasticKey Key public: %w", err)
				}
				repositoryKeyExportableMaterial := s.prepareKeyExportableMaterial(clearPublicJwkBytes, clearPrivateOrSecretJwkBytes, repositoryElasticKey)
				repositoryKeyExportableMaterials = append(repositoryKeyExportableMaterials, repositoryKeyExportableMaterial)
			} else {
				repositoryKeyExportableMaterials = append(repositoryKeyExportableMaterials, emptyKeyExportableMaterial)
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
	var repositoryKey *cryptoutilOrmRepository.MaterialKey
	var repositoryKeyExportableMaterial *keyExportableMaterial
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryElasticKey, err = sqlTransaction.GetElasticKey(repositoryKey.ElasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
		}

		repositoryKey, err = sqlTransaction.GetElasticKeyKey(elasticKeyID, materialKeyID)
		if err != nil {
			return fmt.Errorf("failed to get Key by ElasticKeyID and MaterialKeyID: %w", err)
		}

		// TODO test publicKey and export
		if cryptoutilOrmRepository.IsAsymmetric(&repositoryElasticKey.ElasticKeyAlgorithm) || repositoryElasticKey.ElasticKeyExportAllowed {
			// asymmetric => optionally export clear private key, and extract public key from it
			// symmetric => optionally export clear secret key
			clearPrivateOrSecretJwkBytes, err := s.barrierService.DecryptContent(sqlTransaction, repositoryKey.KeyMaterial)
			if err != nil {
				return fmt.Errorf("failed to decrypt ElasticKey Key: %w", err)
			}
			privateOrSecretJwk, err := joseJwk.ParseKey(clearPrivateOrSecretJwkBytes)
			if err != nil {
				return fmt.Errorf("failed to parse ElasticKey Key: %w", err)
			}
			publicJwk, err := privateOrSecretJwk.PublicKey()
			if err != nil {
				return fmt.Errorf("failed to extract ElasticKey Key public: %w", err)
			}
			clearPublicJwkBytes, err := json.Marshal(publicJwk)
			if err != nil {
				return fmt.Errorf("failed to encode ElasticKey Key public: %w", err)
			}
			repositoryKeyExportableMaterial = s.prepareKeyExportableMaterial(clearPublicJwkBytes, clearPrivateOrSecretJwkBytes, repositoryElasticKey)
		} else {
			repositoryKeyExportableMaterial = emptyKeyExportableMaterial
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate key in ElasticKey: %w", err)
	}

	openapiPostElastickeyElasticKeyIDMaterialkeyResponseObject, err := s.serviceOrmMapper.toServiceKey(repositoryKey, repositoryKeyExportableMaterial)
	if err != nil {
		return nil, fmt.Errorf("failed to map keys in ElasticKey: %w", err)
	}

	return openapiPostElastickeyElasticKeyIDMaterialkeyResponseObject, nil
}

func (s *BusinessLogicService) PostEncryptByElasticKeyID(ctx context.Context, elasticKeyID googleUuid.UUID, encryptParams *cryptoutilBusinessLogicModel.EncryptParams, clearPayloadBytes []byte) ([]byte, error) {
	elasticKey, _, decryptedJweJwk, err := s.getAndDecryptElasticKeyJwk(ctx, &elasticKeyID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt latest JWE JWK for Elastic Key: %w", err)
	}
	if elasticKey.ElasticKeyProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	// TODO Use encryptParams.Context for encryption
	_, jweMessageBytes, err := cryptoutilJose.EncryptBytes([]joseJwk.Key{decryptedJweJwk}, clearPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt bytes with latest Key for ElasticKeyID: %w", err)
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
		return nil, fmt.Errorf("failed to get kid, enc, and alg from JWE message: %w", err)
	}
	elasticKey, _, decryptedJweJwk, err := s.getAndDecryptElasticKeyJwk(ctx, &elasticKeyID, kidUuid)
	if elasticKey.ElasticKeyProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	decryptedJweMessageBytes, err := cryptoutilJose.DecryptBytes([]joseJwk.Key{decryptedJweJwk}, jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt bytes with JWE kid UUID Key for ElasticKeyID : %w", err)
	}
	return decryptedJweMessageBytes, nil
}

func (s *BusinessLogicService) PostSignByElasticKeyID(ctx context.Context, elasticKeyID googleUuid.UUID, clearPayloadBytes []byte) ([]byte, error) {
	elasticKey, _, decryptedJwsJwk, err := s.getAndDecryptElasticKeyJwk(ctx, &elasticKeyID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt latest JWS JWK from Elastic Key for ElasticKeyID: %w", err)
	}
	if elasticKey.ElasticKeyProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	_, jwsMessageBytes, err := cryptoutilJose.SignBytes([]joseJwk.Key{decryptedJwsJwk}, clearPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to sign bytes with latest Key for ElasticKeyID: %w", err)
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
		return nil, fmt.Errorf("failed to get kid and alg from JWS message: %w", err)
	}
	elasticKey, _, decryptedJwsJwk, err := s.getAndDecryptElasticKeyJwk(ctx, &elasticKeyID, kidUuid)
	// TODO validate decrypted JWK is a JWS JWK
	if elasticKey.ElasticKeyProvider != "Internal" {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	verifiedJwsMessageBytes, err := cryptoutilJose.VerifyBytes([]joseJwk.Key{decryptedJwsJwk}, jwsMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to verify bytes with JWS kid UUID Key for ElasticKeyID: %w", err)
	}
	return verifiedJwsMessageBytes, nil
}

func (s *BusinessLogicService) generateJwk(elasticKeyAlgorithm *cryptoutilOrmRepository.ElasticKeyAlgorithm) (*googleUuid.UUID, joseJwk.Key, joseJwk.Key, []byte, []byte, error) {
	var materialKeyID *googleUuid.UUID
	var privateOrSecretJwk joseJwk.Key
	var publicJwk joseJwk.Key
	var encodedPrivateOrSecretJwk []byte
	var encodedPublicJwk []byte

	if s.serviceOrmMapper.isJwe(elasticKeyAlgorithm) {
		enc, alg, err := s.serviceOrmMapper.toJweEncAndAlg(elasticKeyAlgorithm)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to map JWE Elastic Key Algorithm: %w", err)
		}
		materialKeyID, privateOrSecretJwk, publicJwk, encodedPrivateOrSecretJwk, encodedPublicJwk, err = s.jwkGenService.GenerateJweJwk(enc, alg)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to generate JWE: %w", err)
		}
	} else if s.serviceOrmMapper.isJws(elasticKeyAlgorithm) {
		alg, err := s.serviceOrmMapper.toJwsAlg(elasticKeyAlgorithm)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to map JWS Elastic Key Algorithm: %w", err)
		}
		materialKeyID, privateOrSecretJwk, publicJwk, encodedPrivateOrSecretJwk, encodedPublicJwk, err = s.jwkGenService.GenerateJwsJwk(alg)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to generate JWS: %w", err)
		}
	} else {
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported ElasticKeyAlgorithm %v", elasticKeyAlgorithm)
	}

	return materialKeyID, privateOrSecretJwk, publicJwk, encodedPrivateOrSecretJwk, encodedPublicJwk, nil
}

func (s *BusinessLogicService) getAndDecryptElasticKeyJwk(ctx context.Context, elasticKeyID *googleUuid.UUID, kidUuid *googleUuid.UUID) (*cryptoutilOrmRepository.ElasticKey, *cryptoutilOrmRepository.MaterialKey, joseJwk.Key, error) {
	var repositoryElasticKey *cryptoutilOrmRepository.ElasticKey
	var repositoryElasticKeyKey *cryptoutilOrmRepository.MaterialKey
	var decryptedJwkBytes []byte
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error
		repositoryElasticKey, err = sqlTransaction.GetElasticKey(*elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey from ElasticKey: %w", err)
		}
		if kidUuid == nil {
			repositoryElasticKeyKey, err = sqlTransaction.GetElasticKeyLatestKey(*elasticKeyID)
			if err != nil {
				return fmt.Errorf("failed to latest Key from ElasticKey: %w", err)
			}
		} else {
			repositoryElasticKeyKey, err = sqlTransaction.GetElasticKeyKey(*elasticKeyID, *kidUuid)
			if err != nil {
				return fmt.Errorf("failed to specified Key from ElasticKey: %w", err)
			}
		}
		decryptedJwkBytes, err = s.barrierService.DecryptContent(sqlTransaction, repositoryElasticKeyKey.KeyMaterial)
		if err != nil {
			return fmt.Errorf("failed to decrypt Key from ElasticKey: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get and decrypt Key from ElasticKey: %w", err)
	}
	decryptedJwk, err := joseJwk.ParseKey(decryptedJwkBytes)

	return repositoryElasticKey, repositoryElasticKeyKey, decryptedJwk, nil
}
