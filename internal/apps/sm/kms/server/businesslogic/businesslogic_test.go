// Copyright (c) 2025 Justin Cranford
//
//

package businesslogic

import (
	"context"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilOrmRepository "cryptoutil/internal/apps/sm/kms/server/repository/orm"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	testify "github.com/stretchr/testify/require"
)

// Test constants for configuration.
const (
	testOTLPService  = "test-service"
	testOTLPEndpoint = "http://localhost:4318"
	testLogLevel     = "INFO"
)

func TestNewBusinessLogicService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	settings.OTLPService = testOTLPService
	settings.OTLPEndpoint = testOTLPEndpoint
	settings.LogLevel = testLogLevel

	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, settings.ToTelemetrySettings())
	testify.NoError(t, err)

	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
	testify.NoError(t, err)

	tests := []struct {
		name             string
		ctx              context.Context
		telemetryService *cryptoutilSharedTelemetry.TelemetryService
		jwkGenService    *cryptoutilSharedCryptoJose.JWKGenService
		ormRepository    *cryptoutilOrmRepository.OrmRepository
		barrierService   *cryptoutilAppsTemplateServiceServerBarrier.Service
		expectError      bool
		errorContains    string
	}{
		{
			"nil context",
			nil,
			telemetryService,
			jwkGenService,
			&cryptoutilOrmRepository.OrmRepository{},
			&cryptoutilAppsTemplateServiceServerBarrier.Service{},
			true,
			"ctx must be non-nil",
		},
		{
			"nil telemetry service",
			ctx,
			nil,
			jwkGenService,
			&cryptoutilOrmRepository.OrmRepository{},
			&cryptoutilAppsTemplateServiceServerBarrier.Service{},
			true,
			"telemetryService must be non-nil",
		},
		{
			"nil jwk gen service",
			ctx,
			telemetryService,
			nil,
			&cryptoutilOrmRepository.OrmRepository{},
			&cryptoutilAppsTemplateServiceServerBarrier.Service{},
			true,
			"jwkGenService must be non-nil",
		},
		{
			"nil orm repository",
			ctx,
			telemetryService,
			jwkGenService,
			nil,
			&cryptoutilAppsTemplateServiceServerBarrier.Service{},
			true,
			"ormRepository must be non-nil",
		},
		{
			"nil barrier service",
			ctx,
			telemetryService,
			jwkGenService,
			&cryptoutilOrmRepository.OrmRepository{},
			nil,
			true,
			"barrierService must be non-nil",
		},
		{
			"valid parameters",
			ctx,
			telemetryService,
			jwkGenService,
			&cryptoutilOrmRepository.OrmRepository{},
			&cryptoutilAppsTemplateServiceServerBarrier.Service{},
			false,
			"",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service, err := NewBusinessLogicService(tc.ctx, tc.telemetryService, tc.jwkGenService, tc.ormRepository, tc.barrierService)

			if tc.expectError {
				testify.Error(t, err)
				testify.Contains(t, err.Error(), tc.errorContains)
				testify.Nil(t, service)
			} else {
				testify.NoError(t, err)
				testify.NotNil(t, service)
				testify.NotNil(t, service.telemetryService)
				testify.NotNil(t, service.jwkGenService)
				testify.NotNil(t, service.ormRepository)
				testify.NotNil(t, service.oamOrmMapper)
				testify.NotNil(t, service.barrierService)
			}
		})
	}
}

func TestGenerateJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	settings := cryptoutilAppsTemplateServiceConfig.NewTestConfig(cryptoutilSharedMagic.IPv4Loopback, 0, true)
	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, settings.ToTelemetrySettings())
	testify.NoError(t, err)

	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
	testify.NoError(t, err)

	service := &BusinessLogicService{
		telemetryService: telemetryService,
		jwkGenService:    jwkGenService,
		ormRepository:    &cryptoutilOrmRepository.OrmRepository{},
		oamOrmMapper:     NewOamOrmMapper(),
		barrierService:   &cryptoutilAppsTemplateServiceServerBarrier.Service{},
	}

	// Comprehensive JWE algorithm tests - symmetric (dir) and asymmetric (RSA, ECDH).
	jweSymmetricAlgorithms := []cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		cryptoutilOpenapiModel.A128CBCHS256Dir,
		cryptoutilOpenapiModel.A128GCMDir,
		cryptoutilOpenapiModel.A192CBCHS384Dir,
		cryptoutilOpenapiModel.A192GCMDir,
		cryptoutilOpenapiModel.A256CBCHS512Dir,
		cryptoutilOpenapiModel.A256GCMDir,
	}

	jweAsymmetricRSAAlgorithms := []cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		cryptoutilOpenapiModel.A128CBCHS256RSAOAEP,
		cryptoutilOpenapiModel.A128CBCHS256RSAOAEP256,
		cryptoutilOpenapiModel.A256GCMRSAOAEP,
		cryptoutilOpenapiModel.A256GCMRSAOAEP256,
	}

	jweAsymmetricECDHAlgorithms := []cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		cryptoutilOpenapiModel.A128CBCHS256ECDHES,
		cryptoutilOpenapiModel.A128CBCHS256ECDHESA128KW,
		cryptoutilOpenapiModel.A256GCMECDHES,
		cryptoutilOpenapiModel.A256GCMECDHESA256KW,
	}

	// JWS NOT IN API SPEC: 	// Comprehensive JWS algorithm tests - symmetric (HMAC) and asymmetric (RSA, EC, EdDSA).
	// JWS NOT IN API SPEC: 	jwsSymmetricAlgorithms := []cryptoutilModel.ElasticKeyAlgorithm{
	// JWS NOT IN API SPEC: 		cryptoutilKmsServer.HS256,
	// JWS NOT IN API SPEC: 		cryptoutilKmsServer.HS384,
	// JWS NOT IN API SPEC: 		cryptoutilKmsServer.HS512,
	// JWS NOT IN API SPEC: 	}
	// JWS NOT IN API SPEC:
	// JWS NOT IN API SPEC: 	jwsAsymmetricRSAAlgorithms := []cryptoutilModel.ElasticKeyAlgorithm{
	// JWS NOT IN API SPEC: 		cryptoutilKmsServer.RS256,
	// JWS NOT IN API SPEC: 		cryptoutilKmsServer.RS384,
	// JWS NOT IN API SPEC: 		cryptoutilKmsServer.RS512,
	// JWS NOT IN API SPEC: 		cryptoutilKmsServer.PS256,
	// JWS NOT IN API SPEC: 		cryptoutilKmsServer.PS384,
	// JWS NOT IN API SPEC: 		cryptoutilKmsServer.PS512,
	// JWS NOT IN API SPEC: 	}
	// JWS NOT IN API SPEC:
	// JWS NOT IN API SPEC: 	jwsAsymmetricECAlgorithms := []cryptoutilModel.ElasticKeyAlgorithm{
	// JWS NOT IN API SPEC: 		cryptoutilKmsServer.ES256,
	// JWS NOT IN API SPEC: 		cryptoutilKmsServer.ES384,
	// JWS NOT IN API SPEC: 		cryptoutilKmsServer.ES512,
	// JWS NOT IN API SPEC: 	}
	// JWS NOT IN API SPEC:
	// JWS NOT IN API SPEC: 	jwsAsymmetricEdDSAAlgorithms := []cryptoutilModel.ElasticKeyAlgorithm{
	// JWS NOT IN API SPEC: 		cryptoutilKmsServer.EdDSA,
	// JWS NOT IN API SPEC: 	}
	// JWS NOT IN API SPEC:
	// Test JWE symmetric algorithms (dir) - no public key expected.
	for _, alg := range jweSymmetricAlgorithms {
		t.Run("JWE_Symmetric_"+string(alg), func(t *testing.T) {
			t.Parallel()

			materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, err := service.generateJWK(&alg)

			testify.NoError(t, err)
			testify.NotNil(t, materialKeyID)
			testify.NotNil(t, materialKeyNonPublicJWK)
			testify.NotEmpty(t, materialKeyNonPublicJWKBytes)
			// Symmetric algorithms should NOT have separate public key.
			testify.Nil(t, materialKeyPublicJWK)
			testify.Nil(t, materialKeyPublicJWKBytes)
		})
	}

	// Test JWE asymmetric RSA algorithms - public key expected.
	for _, alg := range jweAsymmetricRSAAlgorithms {
		t.Run("JWE_RSA_"+string(alg), func(t *testing.T) {
			t.Parallel()

			materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, err := service.generateJWK(&alg)

			testify.NoError(t, err)
			testify.NotNil(t, materialKeyID)
			testify.NotNil(t, materialKeyNonPublicJWK)
			testify.NotEmpty(t, materialKeyNonPublicJWKBytes)
			testify.NotNil(t, materialKeyPublicJWK)
			testify.NotEmpty(t, materialKeyPublicJWKBytes)
		})
	}

	// Test JWE asymmetric ECDH algorithms - public key expected.
	for _, alg := range jweAsymmetricECDHAlgorithms {
		t.Run("JWE_ECDH_"+string(alg), func(t *testing.T) {
			t.Parallel()

			materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, err := service.generateJWK(&alg)

			testify.NoError(t, err)
			testify.NotNil(t, materialKeyID)
			testify.NotNil(t, materialKeyNonPublicJWK)
			testify.NotEmpty(t, materialKeyNonPublicJWKBytes)
			testify.NotNil(t, materialKeyPublicJWK)
			testify.NotEmpty(t, materialKeyPublicJWKBytes)
		})
	}

	// JWS NOT IN API SPEC: 	// Test JWS symmetric algorithms (HMAC) - no public key expected.
	// JWS NOT IN API SPEC: 	for _, alg := range jwsSymmetricAlgorithms {
	// JWS NOT IN API SPEC: 		t.Run("JWS_HMAC_"+string(alg), func(t *testing.T) {
	// JWS NOT IN API SPEC: 			t.Parallel()
	// JWS NOT IN API SPEC:
	// JWS NOT IN API SPEC: 			materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, err := service.generateJWK(&alg)
	// JWS NOT IN API SPEC:
	// JWS NOT IN API SPEC: 			testify.NoError(t, err)
	// JWS NOT IN API SPEC: 			testify.NotNil(t, materialKeyID)
	// JWS NOT IN API SPEC: 			testify.NotNil(t, materialKeyNonPublicJWK)
	// JWS NOT IN API SPEC: 			testify.NotEmpty(t, materialKeyNonPublicJWKBytes)
	// JWS NOT IN API SPEC: 			// Symmetric HMAC algorithms should NOT have separate public key.
	// JWS NOT IN API SPEC: 			testify.Nil(t, materialKeyPublicJWK)
	// JWS NOT IN API SPEC: 			testify.Nil(t, materialKeyPublicJWKBytes)
	// JWS NOT IN API SPEC: 		})
	// JWS NOT IN API SPEC: 	}
	// JWS NOT IN API SPEC:
	// JWS NOT IN API SPEC: 	// Test JWS asymmetric RSA algorithms - public key expected.
	// JWS NOT IN API SPEC: 	for _, alg := range jwsAsymmetricRSAAlgorithms {
	// JWS NOT IN API SPEC: 		t.Run("JWS_RSA_"+string(alg), func(t *testing.T) {
	// JWS NOT IN API SPEC: 			t.Parallel()
	// JWS NOT IN API SPEC:
	// JWS NOT IN API SPEC: 			materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, err := service.generateJWK(&alg)
	// JWS NOT IN API SPEC:
	// JWS NOT IN API SPEC: 			testify.NoError(t, err)
	// JWS NOT IN API SPEC: 			testify.NotNil(t, materialKeyID)
	// JWS NOT IN API SPEC: 			testify.NotNil(t, materialKeyNonPublicJWK)
	// JWS NOT IN API SPEC: 			testify.NotEmpty(t, materialKeyNonPublicJWKBytes)
	// JWS NOT IN API SPEC: 			testify.NotNil(t, materialKeyPublicJWK)
	// JWS NOT IN API SPEC: 			testify.NotEmpty(t, materialKeyPublicJWKBytes)
	// JWS NOT IN API SPEC: 		})
	// JWS NOT IN API SPEC: 	}
	// JWS NOT IN API SPEC:
	// JWS NOT IN API SPEC: 	// Test JWS asymmetric EC algorithms - public key expected.
	// JWS NOT IN API SPEC: 	for _, alg := range jwsAsymmetricECAlgorithms {
	// JWS NOT IN API SPEC: 		t.Run("JWS_EC_"+string(alg), func(t *testing.T) {
	// JWS NOT IN API SPEC: 			t.Parallel()
	// JWS NOT IN API SPEC:
	// JWS NOT IN API SPEC: 			materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, err := service.generateJWK(&alg)
	// JWS NOT IN API SPEC:
	// JWS NOT IN API SPEC: 			testify.NoError(t, err)
	// JWS NOT IN API SPEC: 			testify.NotNil(t, materialKeyID)
	// JWS NOT IN API SPEC: 			testify.NotNil(t, materialKeyNonPublicJWK)
	// JWS NOT IN API SPEC: 			testify.NotEmpty(t, materialKeyNonPublicJWKBytes)
	// JWS NOT IN API SPEC: 			testify.NotNil(t, materialKeyPublicJWK)
	// JWS NOT IN API SPEC: 			testify.NotEmpty(t, materialKeyPublicJWKBytes)
	// JWS NOT IN API SPEC: 		})
	// JWS NOT IN API SPEC: 	}
	// JWS NOT IN API SPEC:
	// JWS NOT IN API SPEC: 	// Test JWS asymmetric EdDSA algorithms - public key expected.
	// JWS NOT IN API SPEC: 	for _, alg := range jwsAsymmetricEdDSAAlgorithms {
	// JWS NOT IN API SPEC: 		t.Run("JWS_EdDSA_"+string(alg), func(t *testing.T) {
	// JWS NOT IN API SPEC: 			t.Parallel()
	// JWS NOT IN API SPEC:
	// JWS NOT IN API SPEC: 			materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, err := service.generateJWK(&alg)
	// JWS NOT IN API SPEC:
	// JWS NOT IN API SPEC: 			testify.NoError(t, err)
	// JWS NOT IN API SPEC: 			testify.NotNil(t, materialKeyID)
	// JWS NOT IN API SPEC: 			testify.NotNil(t, materialKeyNonPublicJWK)
	// JWS NOT IN API SPEC: 			testify.NotEmpty(t, materialKeyNonPublicJWKBytes)
	// JWS NOT IN API SPEC: 			testify.NotNil(t, materialKeyPublicJWK)
	// JWS NOT IN API SPEC: 			testify.NotEmpty(t, materialKeyPublicJWKBytes)
	// JWS NOT IN API SPEC: 		})
	// JWS NOT IN API SPEC: 	}

	t.Run("unsupported algorithm", func(t *testing.T) {
		t.Parallel()

		unsupported := cryptoutilOpenapiModel.ElasticKeyAlgorithm("UNSUPPORTED")
		materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, err := service.generateJWK(&unsupported)

		testify.Error(t, err)
		testify.Contains(t, err.Error(), "unsupported ElasticKeyAlgorithm")
		testify.Nil(t, materialKeyID)
		testify.Nil(t, materialKeyNonPublicJWK)
		testify.Nil(t, materialKeyPublicJWK)
		testify.Nil(t, materialKeyNonPublicJWKBytes)
		testify.Nil(t, materialKeyPublicJWKBytes)
	})
}
