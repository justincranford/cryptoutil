// Copyright (c) 2025 Justin Cranford
//
//

package businesslogic

import (
	"context"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"
	cryptoutilBarrierService "cryptoutil/internal/server/barrier"
	cryptoutilOrmRepository "cryptoutil/internal/server/repository/orm"

	testify "github.com/stretchr/testify/require"
)

func TestNewBusinessLogicService(t *testing.T) {
	ctx := context.Background()

	settings := &cryptoutilConfig.Settings{
		OTLPService:  "test-service",
		OTLPEndpoint: "http://localhost:4318",
		LogLevel:     "INFO",
	}
	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
	testify.NoError(t, err)

	jwkGenService, err := cryptoutilJose.NewJWKGenService(ctx, telemetryService, false)
	testify.NoError(t, err)

	tests := []struct {
		name             string
		ctx              context.Context
		telemetryService *cryptoutilTelemetry.TelemetryService
		jwkGenService    *cryptoutilJose.JWKGenService
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
	ctx := context.Background()

	settings := &cryptoutilConfig.Settings{
		OTLPService:  "test-service",
		OTLPEndpoint: "http://localhost:4318",
		LogLevel:     "INFO",
	}
	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, settings)
	testify.NoError(t, err)

	jwkGenService, err := cryptoutilJose.NewJWKGenService(ctx, telemetryService, false)
	testify.NoError(t, err)

	service := &BusinessLogicService{
		telemetryService: telemetryService,
		jwkGenService:    jwkGenService,
		ormRepository:    &cryptoutilOrmRepository.OrmRepository{},
		oamOrmMapper:     NewOamOrmMapper(),
		barrierService:   &cryptoutilBarrierService.BarrierService{},
	}

	jweAlgorithms := []cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		cryptoutilOpenapiModel.A128CBCHS256Dir,
		cryptoutilOpenapiModel.A128GCMDir,
		cryptoutilOpenapiModel.A128CBCHS256RSAOAEP,
		cryptoutilOpenapiModel.A128CBCHS256ECDHES,
	}

	jwsAlgorithms := []cryptoutilOpenapiModel.ElasticKeyAlgorithm{
		cryptoutilOpenapiModel.HS256,
		cryptoutilOpenapiModel.RS256,
		cryptoutilOpenapiModel.ES256,
		cryptoutilOpenapiModel.EdDSA,
	}

	for _, alg := range jweAlgorithms {
		t.Run("JWE_"+string(alg), func(t *testing.T) {
			materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, err := service.generateJWK(&alg)

			testify.NoError(t, err)
			testify.NotNil(t, materialKeyID)
			testify.NotNil(t, materialKeyNonPublicJWK)
			testify.NotEmpty(t, materialKeyNonPublicJWKBytes)

			// Asymmetric algorithms should have public key
			if alg == cryptoutilOpenapiModel.A128CBCHS256RSAOAEP || alg == cryptoutilOpenapiModel.A128CBCHS256ECDHES {
				testify.NotNil(t, materialKeyPublicJWK)
				testify.NotEmpty(t, materialKeyPublicJWKBytes)
			}
		})
	}

	for _, alg := range jwsAlgorithms {
		t.Run("JWS_"+string(alg), func(t *testing.T) {
			materialKeyID, materialKeyNonPublicJWK, materialKeyPublicJWK, materialKeyNonPublicJWKBytes, materialKeyPublicJWKBytes, err := service.generateJWK(&alg)

			testify.NoError(t, err)
			testify.NotNil(t, materialKeyID)
			testify.NotNil(t, materialKeyNonPublicJWK)
			testify.NotEmpty(t, materialKeyNonPublicJWKBytes)

			// Asymmetric algorithms should have public key
			if alg == cryptoutilOpenapiModel.RS256 || alg == cryptoutilOpenapiModel.ES256 || alg == cryptoutilOpenapiModel.EdDSA {
				testify.NotNil(t, materialKeyPublicJWK)
				testify.NotEmpty(t, materialKeyPublicJWKBytes)
			}
		})
	}

	t.Run("unsupported algorithm", func(t *testing.T) {
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
