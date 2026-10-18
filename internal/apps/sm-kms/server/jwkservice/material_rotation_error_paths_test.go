// Copyright (c) 2025-2026 Justin Cranford.
//

package jwkservice

import (
	"context"
	"testing"

	cryptoutilAppsSmKmsModel "cryptoutil/internal/apps/sm-kms/server/model"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMaterialRotationService_RotateFailurePaths(t *testing.T) {
	t.Parallel()

	tenantID := newTestUUID(t)
	elasticJWK, _, err := createTestElasticKey(t, tenantID, cryptoutilSharedMagic.JoseAlgES256, cryptoutilAppsSmKmsModel.KeyUseSig, 2)
	require.NoError(t, err)

	tests := []struct {
		name           string
		getError       bool
		wrongTenant    bool
		countError     bool
		count          int64
		maxMaterials   int
		algorithm      string
		rotateError    bool
		incrementError bool
		wantError      string
	}{
		{name: testErrorCaseElasticRepo, getError: true, wantError: testErrorFailedGetElastic},
		{name: testErrorCaseTenantMismatch, wrongTenant: true, wantError: testErrorElasticJWKNotFound},
		{name: "material count error", countError: true, wantError: "failed to count materials"},
		{name: testErrorCaseMaxMaterials, count: 2, maxMaterials: 2, wantError: testErrorCaseMaxMaterials},
		{name: testErrorCaseUnsupportedAlg, algorithm: testErrorUnsupported, maxMaterials: 2, wantError: "failed to create new material"},
		{name: "repository rotation error", maxMaterials: 2, rotateError: true, wantError: "failed to rotate material"},
		{name: "count update warning", maxMaterials: 2, incrementError: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testElastic := *elasticJWK
			if tc.wrongTenant {
				testElastic.TenantID = newTestUUID(t)
			}

			if tc.maxMaterials != 0 {
				testElastic.MaxMaterials = tc.maxMaterials
			}

			if tc.algorithm != "" {
				testElastic.Algorithm = tc.algorithm
			}

			elasticRepo := &elasticJWKRepositoryStub{
				getByIDFn: func(context.Context, googleUuid.UUID) (*cryptoutilAppsSmKmsModel.ElasticJWK, error) {
					if tc.getError {
						return nil, errJWKServiceInjected
					}

					return &testElastic, nil
				},
				incrementCountFn: func(context.Context, googleUuid.UUID) error {
					if tc.incrementError {
						return errJWKServiceInjected
					}

					return nil
				},
			}
			materialRepo := &materialJWKRepositoryStub{
				countMaterialsFn: func(context.Context, googleUuid.UUID) (int64, error) {
					if tc.countError {
						return 0, errJWKServiceInjected
					}

					return tc.count, nil
				},
				rotateMaterialFn: func(context.Context, googleUuid.UUID, *cryptoutilAppsSmKmsModel.MaterialJWK) error {
					if tc.rotateError {
						return errJWKServiceInjected
					}

					return nil
				},
			}

			rotated, err := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService).
				RotateMaterial(testContext, tenantID, testElastic.ID)
			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, rotated)
		})
	}
}
