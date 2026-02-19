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
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm/kms/server/repository/orm"
	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"

	googleUuid "github.com/google/uuid"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"
)

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
