// Copyright (c) 2025 Justin Cranford
//
//

// Package businesslogic implements the KMS business logic layer for key management operations.
package businesslogic

import (
	"context"
	"fmt"
	"time"

	cryptoutilKmsServer "cryptoutil/api/kms/server"
	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilKmsMiddleware "cryptoutil/internal/apps/sm/kms/server/middleware"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm/kms/server/repository/orm"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedTelemetry "cryptoutil/internal/apps/template/service/telemetry"

	googleUuid "github.com/google/uuid"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
)

const (
	providerInternal = "Internal"
)

// BusinessLogicService implements methods in StrictServerInterface.
type BusinessLogicService struct {
	telemetryService *cryptoutilSharedTelemetry.TelemetryService
	jwkGenService    *cryptoutilSharedCryptoJose.JWKGenService
	ormRepository    *cryptoutilOrmRepository.OrmRepository
	oamOrmMapper     *OamOrmMapper
	barrierService   *cryptoutilAppsTemplateServiceServerBarrier.Service
}

// NewBusinessLogicService creates a new BusinessLogicService with injected dependencies.
func NewBusinessLogicService(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService, jwkGenService *cryptoutilSharedCryptoJose.JWKGenService, ormRepository *cryptoutilOrmRepository.OrmRepository, barrierService *cryptoutilAppsTemplateServiceServerBarrier.Service) (*BusinessLogicService, error) {
	if ctx == nil {
		return nil, fmt.Errorf("ctx must be non-nil")
	} else if telemetryService == nil {
		return nil, fmt.Errorf("telemetryService must be non-nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("jwkGenService must be non-nil")
	} else if ormRepository == nil {
		return nil, fmt.Errorf("ormRepository must be non-nil")
	} else if barrierService == nil {
		return nil, fmt.Errorf("barrierService must be non-nil")
	}

	return &BusinessLogicService{
		telemetryService: telemetryService,
		jwkGenService:    jwkGenService,
		ormRepository:    ormRepository,
		oamOrmMapper:     NewOamOrmMapper(),
		barrierService:   barrierService,
	}, nil
}

// getTenantID extracts the tenant ID from context. Returns error if not set.
func getTenantID(ctx context.Context) (googleUuid.UUID, error) {
	realmCtx := cryptoutilKmsMiddleware.GetRealmContext(ctx)
	if realmCtx == nil || realmCtx.TenantID == googleUuid.Nil {
		return googleUuid.Nil, fmt.Errorf("tenant context required")
	}

	return realmCtx.TenantID, nil
}

// AddElasticKey creates a new ElasticKey with an initial MaterialKey.
func (s *BusinessLogicService) AddElasticKey(ctx context.Context, openapiElasticKeyCreate *cryptoutilKmsServer.ElasticKeyCreate) (*cryptoutilKmsServer.ElasticKey, error) {
	// Extract tenant from context (set by middleware)
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	elasticKeyID := s.jwkGenService.GenerateUUIDv7()
	ormElasticKey := s.oamOrmMapper.toOrmAddElasticKey(elasticKeyID, tenantID, openapiElasticKeyCreate)

	if ormElasticKey.ElasticKeyImportAllowed {
		return nil, fmt.Errorf("elasticKeyImportAllowed=true not supported yet")
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

		err = TransitionElasticKeyStatus(cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.Creating), ormElasticKey.ElasticKeyStatus)
		if err != nil {
			return fmt.Errorf("invalid ElasticKeyStatus transition: %w", err)
		}

		materialKeyEncryptedNonPublicJWKBytes, err := s.barrierService.EncryptContentWithContext(ctx, materialKeyClearNonPublicJWKBytes)
		if err != nil {
			return fmt.Errorf("failed to encrypt first MaterialKey for ElasticKey: %w", err)
		}

		ormMaterialKey := s.oamOrmMapper.toOrmAddMaterialKey(elasticKeyID, materialKeyID, materialKeyClearPublicJWKBytes, materialKeyEncryptedNonPublicJWKBytes, materialKeyGenerateDate)

		err = sqlTransaction.AddElasticKeyMaterialKey(ormMaterialKey)
		if err != nil {
			return fmt.Errorf("failed to add first MaterialKey for ElasticKey: %w", err)
		}

		err = sqlTransaction.UpdateElasticKeyStatus(*elasticKeyID, cryptoutilKmsServer.Active)
		if err != nil {
			return fmt.Errorf("failed to update ElasticKeyStatus to active: %w", err)
		}

		ormElasticKey, err = sqlTransaction.GetElasticKey(tenantID, elasticKeyID)
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

// GetElasticKeyByElasticKeyID retrieves an ElasticKey by its ID.
func (s *BusinessLogicService) GetElasticKeyByElasticKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID) (*cryptoutilKmsServer.ElasticKey, error) {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	var ormElasticKey *cryptoutilOrmRepository.ElasticKey

	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error

		ormElasticKey, err = sqlTransaction.GetElasticKey(tenantID, elasticKeyID)
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

// GetElasticKeys retrieves ElasticKeys matching the provided query parameters.
func (s *BusinessLogicService) GetElasticKeys(ctx context.Context, elasticKeyQueryParams *cryptoutilOpenapiModel.ElasticKeysQueryParams) ([]cryptoutilKmsServer.ElasticKey, error) {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	ormElasticKeysQueryParams, err := s.oamOrmMapper.toOrmGetElasticKeysQueryParams(tenantID, elasticKeyQueryParams)
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

// GenerateMaterialKeyInElasticKey generates a new MaterialKey for an existing ElasticKey.
func (s *BusinessLogicService) GenerateMaterialKeyInElasticKey(ctx context.Context, elasticKeyID *googleUuid.UUID, _ *cryptoutilKmsServer.MaterialKeyGenerate) (*cryptoutilKmsServer.MaterialKey, error) {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	var ormElasticKey *cryptoutilOrmRepository.ElasticKey

	var ormMaterialKey *cryptoutilOrmRepository.MaterialKey

	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error

		ormElasticKey, err = sqlTransaction.GetElasticKey(tenantID, elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
		}

		if ormElasticKey.ElasticKeyStatus != cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingGenerate) && ormElasticKey.ElasticKeyStatus != cryptoutilKmsServer.Active {
			return fmt.Errorf("invalid ElasticKey Status: %w", err)
		}

		materialKeyID, _, _, clearMaterialKeyNonPublicJWKBytes, clearPublicJWKBytes, err := s.generateJWK(&ormElasticKey.ElasticKeyAlgorithm)
		if err != nil {
			return fmt.Errorf("failed to generate new MaterialKey for ElasticKey: %w", err)
		}

		materialKeyGenerateDate := time.Now().UTC()

		encryptedMaterialKeyPrivateOrPublicJWKBytes, err := s.barrierService.EncryptContentWithContext(ctx, clearMaterialKeyNonPublicJWKBytes)
		if err != nil {
			return fmt.Errorf("failed to encrypt new MaterialKey for ElasticKey: %w", err)
		}

		// Convert time.Time to Unix milliseconds for database storage
		generateDateMillis := materialKeyGenerateDate.UnixMilli()
		ormMaterialKey = &cryptoutilOrmRepository.MaterialKey{
			ElasticKeyID:                  *elasticKeyID,
			MaterialKeyID:                 *materialKeyID,
			MaterialKeyClearPublic:        clearPublicJWKBytes,
			MaterialKeyEncryptedNonPublic: encryptedMaterialKeyPrivateOrPublicJWKBytes,
			MaterialKeyGenerateDate:       &generateDateMillis,
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

// GetMaterialKeysForElasticKey retrieves MaterialKeys for a specific ElasticKey.
func (s *BusinessLogicService) GetMaterialKeysForElasticKey(ctx context.Context, elasticKeyID *googleUuid.UUID, elasticKeyMaterialKeysQueryParams *cryptoutilOpenapiModel.ElasticKeyMaterialKeysQueryParams) ([]cryptoutilKmsServer.MaterialKey, error) {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	ormElasticKeyMaterialKeysQueryParams, err := s.oamOrmMapper.toOrmGetMaterialKeysForElasticKeyQueryParams(elasticKeyMaterialKeysQueryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to map MaterialKeys for ElasticKey query parameters: %w", err)
	}

	var ormElasticKey *cryptoutilOrmRepository.ElasticKey

	var ormMaterialKeys []cryptoutilOrmRepository.MaterialKey

	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error

		ormElasticKey, err = sqlTransaction.GetElasticKey(tenantID, elasticKeyID)
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

// GetMaterialKeys retrieves MaterialKeys matching the provided query parameters.
func (s *BusinessLogicService) GetMaterialKeys(ctx context.Context, materialKeysQueryParams *cryptoutilOpenapiModel.MaterialKeysQueryParams) ([]cryptoutilKmsServer.MaterialKey, error) {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return nil, err
	}

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

		// Cache GetElasticKey calls to avoid redundant database queries for the same ElasticKeyID
		elasticKeyCache := make(map[googleUuid.UUID]*cryptoutilOrmRepository.ElasticKey)

		for _, ormMaterialKey := range ormMaterialKeys {
			elasticKeyID := ormMaterialKey.ElasticKeyID

			// Check cache first
			if cachedElasticKey, exists := elasticKeyCache[elasticKeyID]; exists {
				ormElasticKey = cachedElasticKey
			} else {
				// Cache miss - fetch from database
				ormElasticKey, err = sqlTransaction.GetElasticKey(tenantID, &elasticKeyID)
				if err != nil {
					return fmt.Errorf("failed to get ElasticKey by ElasticKeyID: %w", err)
				} else if ormElasticKey == nil {
					return fmt.Errorf("got nil ElasticKey by ElasticKeyID: %w", err)
				}
				// Cache the result
				elasticKeyCache[elasticKeyID] = ormElasticKey
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

// GetMaterialKeyByElasticKeyAndMaterialKeyID retrieves a MaterialKey by ElasticKey ID and MaterialKey ID.
func (s *BusinessLogicService) GetMaterialKeyByElasticKeyAndMaterialKeyID(ctx context.Context, elasticKeyID, materialKeyID *googleUuid.UUID) (*cryptoutilKmsServer.MaterialKey, error) {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	var ormMaterialKey *cryptoutilOrmRepository.MaterialKey

	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error

		_, err = sqlTransaction.GetElasticKey(tenantID, elasticKeyID)
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

// PostGenerateByElasticKeyID generates cryptographic key material using the active MaterialKey.
func (s *BusinessLogicService) PostGenerateByElasticKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID, generateParams *cryptoutilOpenapiModel.GenerateParams) ([]byte, []byte, []byte, error) {
	alg, err := cryptoutilSharedCryptoJose.ToGenerateAlgorithm((*string)(generateParams.Alg))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to map generate algorithm: %w", err)
	}

	_, _, _, clearNonPublicJWKBytes, clearPublicJWKBytes, err := cryptoutilSharedCryptoJose.GenerateJWKForAlg(alg)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate key from algorithm: %w", err)
	}

	encryptedNonPublicJWKBytes, err := s.PostEncryptByElasticKeyID(ctx, elasticKeyID, &cryptoutilOpenapiModel.EncryptParams{Context: generateParams.Context}, clearNonPublicJWKBytes)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to encrypt generated key: %w", err)
	}

	return encryptedNonPublicJWKBytes, clearNonPublicJWKBytes, clearPublicJWKBytes, nil
}

// PostEncryptByElasticKeyID encrypts a payload using the active MaterialKey in an ElasticKey.
func (s *BusinessLogicService) PostEncryptByElasticKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID, encryptParams *cryptoutilOpenapiModel.EncryptParams, clearPayloadBytes []byte) ([]byte, error) {
	elasticKey, _, decryptedMaterialKeyNonPublicJWEJWK, clearMaterialKeyPublicJWEJWK, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, elasticKeyID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt latest MaterialKey for ElasticKey: %w", err)
	}

	if elasticKey.ElasticKeyProvider != providerInternal {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}
	// Use encryptParams.Context for encryption
	var (
		jweMessageBytes []byte
		contextBytes    []byte
	)

	if encryptParams.Context != nil {
		contextBytes = []byte(*encryptParams.Context)
	}

	if clearMaterialKeyPublicJWEJWK != nil {
		_, jweMessageBytes, err = cryptoutilSharedCryptoJose.EncryptBytesWithContext([]joseJwk.Key{clearMaterialKeyPublicJWEJWK}, clearPayloadBytes, contextBytes) // asymmetric
	} else {
		_, jweMessageBytes, err = cryptoutilSharedCryptoJose.EncryptBytesWithContext([]joseJwk.Key{decryptedMaterialKeyNonPublicJWEJWK}, clearPayloadBytes, contextBytes) // symmetric
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encrypt bytes with latest MaterialKey for ElasticKeyID: %w", err)
	}

	return jweMessageBytes, nil
}

// PostDecryptByElasticKeyID decrypts a JWE message using the appropriate MaterialKey.
func (s *BusinessLogicService) PostDecryptByElasticKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID, jweMessageBytes []byte) ([]byte, error) {
	jweMessage, err := joseJwe.Parse(jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWE message bytes: %w", err)
	}

	materialKeyID, err := cryptoutilSharedCryptoJose.ExtractKidFromJWEMessage(jweMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWE message header kid: %w", err)
	}

	elasticKey, _, decryptedMaterialKeyNonPublicJWEJWK, _, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, elasticKeyID, materialKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt material key: %w", err)
	}

	if elasticKey.ElasticKeyProvider != providerInternal {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	} else if !cryptoutilSharedCryptoJose.IsJWE(&elasticKey.ElasticKeyAlgorithm) {
		return nil, fmt.Errorf("decrypt not supported by KeyMaterial with ElasticKeyAlgorithm %v", elasticKey.ElasticKeyAlgorithm)
	}

	decryptedJWEMessageBytes, err := cryptoutilSharedCryptoJose.DecryptBytes([]joseJwk.Key{decryptedMaterialKeyNonPublicJWEJWK}, jweMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt bytes with MaterialKey for ElasticKeyID : %w", err)
	}

	return decryptedJWEMessageBytes, nil
}

// PostSignByElasticKeyID signs a payload using the active MaterialKey in an ElasticKey.
func (s *BusinessLogicService) PostSignByElasticKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID, clearPayloadBytes []byte) ([]byte, error) {
	elasticKey, _, decryptedMaterialKeyNonPublicJWSJWK, _, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, elasticKeyID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt latest MaterialKey JWS JWK from ElasticKey for ElasticKeyID: %w", err)
	}

	if elasticKey.ElasticKeyProvider != providerInternal {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	}

	_, jwsMessageBytes, err := cryptoutilSharedCryptoJose.SignBytes([]joseJwk.Key{decryptedMaterialKeyNonPublicJWSJWK}, clearPayloadBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to sign bytes with latest MaterialKey for ElasticKeyID: %w", err)
	}

	return jwsMessageBytes, nil
}

// PostVerifyByElasticKeyID verifies a JWS signature using the appropriate MaterialKey.
func (s *BusinessLogicService) PostVerifyByElasticKeyID(ctx context.Context, elasticKeyID *googleUuid.UUID, jwsMessageBytes []byte) ([]byte, error) {
	jwsMessage, err := joseJws.Parse(jwsMessageBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWS message bytes: %w", err)
	}

	kidUUID, _, err := cryptoutilSharedCryptoJose.ExtractKidAlgFromJWSMessage(jwsMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWS message headers kid and alg: %w", err)
	}

	elasticKey, _, decryptedMaterialKeyNonPublicJWEJWK, clearMaterialKeyPublicJWEJWK, err := s.getAndDecryptMaterialKeyInElasticKey(ctx, elasticKeyID, kidUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get and decrypt material key: %w", err)
	}

	if elasticKey.ElasticKeyProvider != providerInternal {
		return nil, fmt.Errorf("provider not supported yet; use Internal for now")
	} else if !cryptoutilSharedCryptoJose.IsJWS(&elasticKey.ElasticKeyAlgorithm) {
		return nil, fmt.Errorf("verify not supported by KeyMaterial with ElasticKeyAlgorithm %v", elasticKey.ElasticKeyAlgorithm)
	}

	var verifiedJWSMessageBytes []byte
	if clearMaterialKeyPublicJWEJWK != nil {
		verifiedJWSMessageBytes, err = cryptoutilSharedCryptoJose.VerifyBytes([]joseJwk.Key{clearMaterialKeyPublicJWEJWK}, jwsMessageBytes) // asymmetric
	} else {
		verifiedJWSMessageBytes, err = cryptoutilSharedCryptoJose.VerifyBytes([]joseJwk.Key{decryptedMaterialKeyNonPublicJWEJWK}, jwsMessageBytes) // symmetric
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

	if cryptoutilSharedCryptoJose.IsJWE(elasticKeyAlgorithm) {
		enc, alg, err := cryptoutilSharedCryptoJose.ToJWEEncAndAlg(elasticKeyAlgorithm)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to map ElasticKeyAlgorithm: %w", err)
		}

		materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, err = s.jwkGenService.GenerateJWEJWK(enc, alg)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to generate MaterialKey JWE JWK: %w", err)
		}
	} else if cryptoutilSharedCryptoJose.IsJWS(elasticKeyAlgorithm) {
		alg, err := cryptoutilSharedCryptoJose.ToJWSAlg(elasticKeyAlgorithm)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to map JWS ElasticKey Algorithm: %w", err)
		}

		materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, err = s.jwkGenService.GenerateJWSJWK(*alg)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("failed to generate MaterialKey JWS JWK: %w", err)
		}
	} else {
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported ElasticKeyAlgorithm %v", elasticKeyAlgorithm)
	}

	return materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, nil
}

//nolint:unparam // Some callers ignore certain return values by design
func (s *BusinessLogicService) getAndDecryptMaterialKeyInElasticKey(ctx context.Context, elasticKeyID, materialKeyKidUUID *googleUuid.UUID) (*cryptoutilOrmRepository.ElasticKey, *cryptoutilOrmRepository.MaterialKey, joseJwk.Key, joseJwk.Key, error) {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var ormElasticKey *cryptoutilOrmRepository.ElasticKey

	var ormMaterialKey *cryptoutilOrmRepository.MaterialKey

	var materialKeyDecryptedNonPublicJWKBytes []byte

	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadOnly, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error

		ormElasticKey, err = sqlTransaction.GetElasticKey(tenantID, elasticKeyID)
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

		materialKeyDecryptedNonPublicJWKBytes, err = s.barrierService.DecryptContentWithContext(ctx, ormMaterialKey.MaterialKeyEncryptedNonPublic)
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
	if len(ormMaterialKey.MaterialKeyClearPublic) > 0 {
		clearMaterialKeyPublicJWK, err = joseJwk.ParseKey(ormMaterialKey.MaterialKeyClearPublic)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to parse MaterialKeyClearPublic: %w", err)
		}
	}

	return ormElasticKey, ormMaterialKey, decryptedMaterialKeyNonPublicJWK, clearMaterialKeyPublicJWK, nil
}

// UpdateElasticKey updates an existing ElasticKey.
func (s *BusinessLogicService) UpdateElasticKey(ctx context.Context, elasticKeyID *googleUuid.UUID, updateRequest *cryptoutilKmsServer.ElasticKeyUpdate) (*cryptoutilKmsServer.ElasticKey, error) {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	var ormElasticKey *cryptoutilOrmRepository.ElasticKey

	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error

		ormElasticKey, err = sqlTransaction.GetElasticKey(tenantID, elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey: %w", err)
		}

		ormElasticKey.ElasticKeyName = updateRequest.Name
		ormElasticKey.ElasticKeyDescription = *updateRequest.Description

		err = sqlTransaction.UpdateElasticKey(ormElasticKey)
		if err != nil {
			return fmt.Errorf("failed to update ElasticKey: %w", err)
		}

		ormElasticKey, err = sqlTransaction.GetElasticKey(tenantID, elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get updated ElasticKey: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update ElasticKey: %w", err)
	}

	return s.oamOrmMapper.toOamElasticKey(ormElasticKey), nil
}

// DeleteElasticKey deletes an ElasticKey and its associated MaterialKeys.
func (s *BusinessLogicService) DeleteElasticKey(ctx context.Context, elasticKeyID *googleUuid.UUID) error {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return err
	}

	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		ormElasticKey, err := sqlTransaction.GetElasticKey(tenantID, elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey: %w", err)
		}

		var deleteStatus cryptoutilKmsServer.ElasticKeyStatus

		switch ormElasticKey.ElasticKeyStatus {
		case cryptoutilKmsServer.Active:
			deleteStatus = cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingDeleteWasActive)
		case cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.Disabled):
			deleteStatus = cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingDeleteWasDisabled)
		case cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.ImportFailed):
			deleteStatus = cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingDeleteWasImportFailed)
		case cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingImport):
			deleteStatus = cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingDeleteWasPendingImport)
		case cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.GenerateFailed):
			deleteStatus = cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingDeleteWasGenerateFailed)
		default:
			return fmt.Errorf("cannot delete ElasticKey in status %s", ormElasticKey.ElasticKeyStatus)
		}

		err = sqlTransaction.UpdateElasticKeyStatus(ormElasticKey.ElasticKeyID, deleteStatus)
		if err != nil {
			return fmt.Errorf("failed to update ElasticKey status to %s: %w", deleteStatus, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to delete ElasticKey: %w", err)
	}

	return nil
}

// ImportMaterialKey imports a MaterialKey into an existing ElasticKey.
func (s *BusinessLogicService) ImportMaterialKey(ctx context.Context, elasticKeyID *googleUuid.UUID, importRequest *cryptoutilKmsServer.MaterialKeyImport) (*cryptoutilKmsServer.MaterialKey, error) {
	tenantID, err := getTenantID(ctx)
	if err != nil {
		return nil, err
	}

	var ormElasticKey *cryptoutilOrmRepository.ElasticKey

	var ormMaterialKey *cryptoutilOrmRepository.MaterialKey

	err = s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		var err error

		ormElasticKey, err = sqlTransaction.GetElasticKey(tenantID, elasticKeyID)
		if err != nil {
			return fmt.Errorf("failed to get ElasticKey: %w", err)
		}

		if !ormElasticKey.ElasticKeyImportAllowed {
			return fmt.Errorf("import not allowed for ElasticKey")
		}

		if ormElasticKey.ElasticKeyStatus != cryptoutilKmsServer.ElasticKeyStatus(cryptoutilOpenapiModel.PendingImport) && ormElasticKey.ElasticKeyStatus != cryptoutilKmsServer.Active {
			return fmt.Errorf("invalid ElasticKey status for import: %s", ormElasticKey.ElasticKeyStatus)
		}

		importedJWKBytes := []byte(importRequest.JWK)

		materialKeyID := googleUuid.New()

		materialKeyImportDate := time.Now().UTC()

		encryptedMaterialKeyBytes, err := s.barrierService.EncryptContentWithContext(ctx, importedJWKBytes)
		if err != nil {
			return fmt.Errorf("failed to encrypt imported MaterialKey: %w", err)
		}

		// Convert time.Time to Unix milliseconds for database storage
		importDateMillis := materialKeyImportDate.UnixMilli()
		ormMaterialKey = &cryptoutilOrmRepository.MaterialKey{
			ElasticKeyID:                  *elasticKeyID,
			MaterialKeyID:                 materialKeyID,
			MaterialKeyClearPublic:        nil,
			MaterialKeyEncryptedNonPublic: encryptedMaterialKeyBytes,
			MaterialKeyImportDate:         &importDateMillis,
		}

		err = sqlTransaction.AddElasticKeyMaterialKey(ormMaterialKey)
		if err != nil {
			return fmt.Errorf("failed to insert imported MaterialKey: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to import MaterialKey: %w", err)
	}

	oamMaterialKey, err := s.oamOrmMapper.toOamMaterialKey(ormMaterialKey)
	if err != nil {
		return nil, fmt.Errorf("failed to map imported MaterialKey: %w", err)
	}

	return oamMaterialKey, nil
}

// RevokeMaterialKey revokes a MaterialKey within an ElasticKey.
func (s *BusinessLogicService) RevokeMaterialKey(ctx context.Context, elasticKeyID, materialKeyID *googleUuid.UUID) error {
	err := s.ormRepository.WithTransaction(ctx, cryptoutilOrmRepository.ReadWrite, func(sqlTransaction *cryptoutilOrmRepository.OrmTransaction) error {
		ormMaterialKey, err := sqlTransaction.GetElasticKeyMaterialKeyVersion(elasticKeyID, materialKeyID)
		if err != nil {
			return fmt.Errorf("failed to get MaterialKey: %w", err)
		}

		if ormMaterialKey.MaterialKeyRevocationDate != nil {
			return fmt.Errorf("MaterialKey already revoked")
		}

		// Convert time.Time to Unix milliseconds for database storage
		revocationDate := time.Now().UTC()
		revocationDateMillis := revocationDate.UnixMilli()
		ormMaterialKey.MaterialKeyRevocationDate = &revocationDateMillis

		err = sqlTransaction.UpdateElasticKeyMaterialKeyRevoke(ormMaterialKey)
		if err != nil {
			return fmt.Errorf("failed to revoke MaterialKey: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to revoke MaterialKey: %w", err)
	}

	return nil
}

// DeleteMaterialKey deletes a MaterialKey within an ElasticKey.
func (s *BusinessLogicService) DeleteMaterialKey(ctx context.Context, elasticKeyID, materialKeyID *googleUuid.UUID) error {
	// TODO: Implement material key deletion. For now, return not implemented error.
	summary := "delete material key not implemented"

	return cryptoutilSharedApperr.NewHTTP500InternalServerError(&summary, nil)
}
