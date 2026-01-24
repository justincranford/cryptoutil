// Copyright (c) 2025 Justin Cranford
//
//

package businesslogic

import (
	"context"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilOrmRepository "cryptoutil/internal/kms/server/repository/orm"
	cryptoutilBarrierService "cryptoutil/internal/shared/barrier"
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

	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, settings)
	testify.NoError(t, err)

	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
	testify.NoError(t, err)

	tests := []struct {
		name             string
		ctx              context.Context
		telemetryService *cryptoutilSharedTelemetry.TelemetryService
		jwkGenService    *cryptoutilSharedCryptoJose.JWKGenService
		ormRepository    *cryptoutilOrmRepository.OrmRepository
		barrierService   *cryptoutilBarrierService.BarrierService
		expectError      bool
		errorContains    string
	}{
		{
			"nil context",
			nil,
			telemetryService,
			jwkGenService,
			&cryptoutilOrmRepository.OrmRepository{},
			&cryptoutilBarrierService.BarrierService{},
			true,
			"ctx must be non-nil",
		},
		{
			"nil telemetry service",
			ctx,
			nil,
			jwkGenService,
			&cryptoutilOrmRepository.OrmRepository{},
			&cryptoutilBarrierService.BarrierService{},
			true,
			"telemetryService must be non-nil",
		},
		{
			"nil jwk gen service",
			ctx,
			telemetryService,
			nil,
			&cryptoutilOrmRepository.OrmRepository{},
			&cryptoutilBarrierService.BarrierService{},
			true,
			"jwkGenService must be non-nil",
		},
		{
			"nil orm repository",
			ctx,
			telemetryService,
			jwkGenService,
			nil,
			&cryptoutilBarrierService.BarrierService{},
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
			&cryptoutilBarrierService.BarrierService{},
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
	telemetryService, err := cryptoutilSharedTelemetry.NewTelemetryService(ctx, settings)
	testify.NoError(t, err)

	jwkGenService, err := cryptoutilSharedCryptoJose.NewJWKGenService(ctx, telemetryService, false)
	testify.NoError(t, err)

	service := &BusinessLogicService{
		telemetryService: telemetryService,
		jwkGenService:    jwkGenService,
		ormRepository:    &cryptoutilOrmRepository.OrmRepository{},
		oamOrmMapper:     NewOamOrmMapper(),
		barrierService:   &cryptoutilBarrierService.BarrierService{},
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

	// Comprehensive JWS algorithm tests - symmetric (HMAC) and asymmetric (RSA, EC, EdDSA).
	jwsSymmetricAlgorithms := []cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		cryptoutilOpenapiModel.HS256,
		cryptoutilOpenapiModel.HS384,
		cryptoutilOpenapiModel.HS512,
	}

	jwsAsymmetricRSAAlgorithms := []cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		cryptoutilOpenapiModel.RS256,
		cryptoutilOpenapiModel.RS384,
		cryptoutilOpenapiModel.RS512,
		cryptoutilOpenapiModel.PS256,
		cryptoutilOpenapiModel.PS384,
		cryptoutilOpenapiModel.PS512,
	}

	jwsAsymmetricECAlgorithms := []cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		cryptoutilOpenapiModel.ES256,
		cryptoutilOpenapiModel.ES384,
		cryptoutilOpenapiModel.ES512,
	}

	jwsAsymmetricEdDSAAlgorithms := []cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		cryptoutilOpenapiModel.EdDSA,
	}

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

	// Test JWS symmetric algorithms (HMAC) - no public key expected.
	for _, alg := range jwsSymmetricAlgorithms {
		t.Run("JWS_HMAC_"+string(alg), func(t *testing.T) {
			t.Parallel()

			materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, err := service.generateJWK(&alg)

			testify.NoError(t, err)
			testify.NotNil(t, materialKeyID)
			testify.NotNil(t, materialKeyNonPublicJWK)
			testify.NotEmpty(t, materialKeyNonPublicJWKBytes)
			// Symmetric HMAC algorithms should NOT have separate public key.
			testify.Nil(t, materialKeyPublicJWK)
			testify.Nil(t, materialKeyPublicJWKBytes)
		})
	}

	// Test JWS asymmetric RSA algorithms - public key expected.
	for _, alg := range jwsAsymmetricRSAAlgorithms {
		t.Run("JWS_RSA_"+string(alg), func(t *testing.T) {
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

	// Test JWS asymmetric EC algorithms - public key expected.
	for _, alg := range jwsAsymmetricECAlgorithms {
		t.Run("JWS_EC_"+string(alg), func(t *testing.T) {
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

	// Test JWS asymmetric EdDSA algorithms - public key expected.
	for _, alg := range jwsAsymmetricEdDSAAlgorithms {
		t.Run("JWS_EdDSA_"+string(alg), func(t *testing.T) {
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
