// Copyright (c) 2025-2026 Justin Cranford.

package repository

import (
	"context"
	"errors"
	"strings"
	"testing"

	cryptoutilAppsSmKmsModel "cryptoutil/internal/apps/sm-kms/server/model"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const testFailedToListAuditLogEntriesError = "failed to list audit log entries"

func TestAuditConfigRepository_ShouldAuditSamplingPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := newRepositoryTestDB(t)
	repo := NewAuditConfigRepository(db)
	tenantID := googleUuid.Must(googleUuid.NewV7())

	_, err := repo.ShouldAudit(ctx, tenantID, "missing")
	require.NoError(t, err)

	config := &cryptoutilAppsSmKmsModel.AuditConfig{
		TenantID: tenantID, Operation: "invalid-sampling-rate", Enabled: true, SamplingRate: 2,
	}
	require.NoError(t, repo.Upsert(ctx, config))

	_, err = repo.ShouldAudit(ctx, tenantID, config.Operation)
	require.ErrorContains(t, err, "outside [0,1]")
}

func TestRepositories_ListFetchErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := newRepositoryTestDB(t)
	require.NoError(t, db.Callback().Query().After("gorm:query").Register("test:fail_list_fetch", func(tx *gorm.DB) {
		if strings.HasPrefix(strings.TrimSpace(tx.Statement.SQL.String()), "SELECT *") {
			_ = tx.AddError(errors.New("injected list fetch failure"))
		}
	}))

	tenantID := googleUuid.Must(googleUuid.NewV7())
	elasticJWK := newRepositoryTestElasticJWK(tenantID)
	require.NoError(t, db.Create(elasticJWK).Error)

	tests := []struct {
		name    string
		call    func() error
		wantErr string
	}{
		{
			name: "audit log by tenant",
			call: func() error {
				_, _, err := NewAuditLogRepository(db).List(ctx, tenantID, 0, 10)

				return err
			},
			wantErr: testFailedToListAuditLogEntriesError,
		},
		{
			name: "audit log by elastic JWK",
			call: func() error {
				_, _, err := NewAuditLogRepository(db).ListByElasticJWK(ctx, elasticJWK.ID, 0, 10)

				return err
			},
			wantErr: testFailedToListAuditLogEntriesError,
		},
		{
			name: "audit log by operation",
			call: func() error {
				_, _, err := NewAuditLogRepository(db).ListByOperation(ctx, tenantID, cryptoutilAppsSmKmsModel.OperationSign, 0, 10)

				return err
			},
			wantErr: testFailedToListAuditLogEntriesError,
		},
		{
			name: "elastic JWK",
			call: func() error {
				_, _, err := NewElasticJWKRepository(db).List(ctx, tenantID, 0, 10)

				return err
			},
			wantErr: "failed to list elastic JWKs",
		},
		{
			name: "material JWK",
			call: func() error {
				_, _, err := NewMaterialJWKRepository(db).ListByElasticJWK(ctx, elasticJWK.ID, 0, 10)

				return err
			},
			wantErr: "failed to list material JWKs",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.ErrorContains(t, tc.call(), tc.wantErr)
		})
	}
}

func TestMaterialJWKRepository_RotateMaterialCreateErrorRollsBack(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := newRepositoryTestDB(t)
	tenantID := googleUuid.Must(googleUuid.NewV7())
	elasticJWK := newRepositoryTestElasticJWK(tenantID)
	require.NoError(t, db.Create(elasticJWK).Error)

	active := newRepositoryTestMaterial(elasticJWK.ID, true)
	require.NoError(t, db.Create(active).Error)

	err := NewMaterialJWKRepository(db).RotateMaterial(ctx, elasticJWK.ID, active)
	require.ErrorContains(t, err, "failed to rotate material JWK")

	got, getErr := NewMaterialJWKRepository(db).GetActiveMaterial(ctx, elasticJWK.ID)
	require.NoError(t, getErr)
	require.Equal(t, active.ID, got.ID)
	require.Nil(t, got.RetiredAt)
}
