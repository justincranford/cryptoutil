// Copyright (c) 2025 Justin Cranford
//

package service

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestAuditLogService_LogOperation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name         string
		operation    string
		success      bool
		errorMessage *string
	}{
		{
			name:         "log successful generate operation",
			operation:    cryptoutilAppsJoseJaDomain.OperationGenerate,
			success:      true,
			errorMessage: nil,
		},
		{
			name:         "log successful sign operation",
			operation:    cryptoutilAppsJoseJaDomain.OperationSign,
			success:      true,
			errorMessage: nil,
		},
		{
			name:         "log failed verify operation",
			operation:    cryptoutilAppsJoseJaDomain.OperationVerify,
			success:      false,
			errorMessage: stringPtr("verification failed"),
		},
		{
			name:         "log successful encrypt operation",
			operation:    cryptoutilAppsJoseJaDomain.OperationEncrypt,
			success:      true,
			errorMessage: nil,
		},
		{
			name:         "log failed decrypt operation",
			operation:    cryptoutilAppsJoseJaDomain.OperationDecrypt,
			success:      false,
			errorMessage: stringPtr("decryption failed"),
		},
		{
			name:         "log successful rotate operation",
			operation:    cryptoutilAppsJoseJaDomain.OperationRotate,
			success:      true,
			errorMessage: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewAuditLogService(testAuditLogRepo, testAuditConfigRepo, testElasticRepo)
			tenantID := googleUuid.New()

			// Set up 100% audit config for reliable testing.
			config := &cryptoutilAppsJoseJaDomain.AuditConfig{
				TenantID:     tenantID,
				Operation:    tt.operation,
				Enabled:      true,
				SamplingRate: 1.0,
			}
			err := svc.UpdateAuditConfig(ctx, tenantID, config)
			require.NoError(t, err)

			requestID := googleUuid.New().String()

			err = svc.LogOperation(ctx, tenantID, nil, tt.operation, requestID, tt.success, tt.errorMessage)
			require.NoError(t, err)

			// Verify log was created.
			logs, total, listErr := svc.ListAuditLogs(ctx, tenantID, 0, 10)
			require.NoError(t, listErr)
			require.Equal(t, int64(1), total)
			require.Len(t, logs, 1)
			require.Equal(t, tt.operation, logs[0].Operation)
			require.Equal(t, tt.success, logs[0].Success)
			require.Equal(t, requestID, logs[0].RequestID)
		})
	}
}

func TestAuditLogService_ListAuditLogs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name       string
		numLogs    int
		offset     int
		limit      int
		expectLogs int
	}{
		{
			name:       "list all logs",
			numLogs:    5,
			offset:     0,
			limit:      10,
			expectLogs: 5,
		},
		{
			name:       "list with pagination",
			numLogs:    10,
			offset:     0,
			limit:      5,
			expectLogs: 5,
		},
		{
			name:       "list with offset",
			numLogs:    10,
			offset:     5,
			limit:      10,
			expectLogs: 5,
		},
		{
			name:       "empty tenant",
			numLogs:    0,
			offset:     0,
			limit:      10,
			expectLogs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewAuditLogService(testAuditLogRepo, testAuditConfigRepo, testElasticRepo)
			tenantID := googleUuid.New()

			// Set up 100% audit config for reliable testing.
			if tt.numLogs > 0 {
				config := &cryptoutilAppsJoseJaDomain.AuditConfig{
					TenantID:     tenantID,
					Operation:    cryptoutilAppsJoseJaDomain.OperationGenerate,
					Enabled:      true,
					SamplingRate: 1.0,
				}
				err := svc.UpdateAuditConfig(ctx, tenantID, config)
				require.NoError(t, err)
			}

			// Create test logs.
			for i := 0; i < tt.numLogs; i++ {
				requestID := googleUuid.New().String()
				err := svc.LogOperation(ctx, tenantID, nil, cryptoutilAppsJoseJaDomain.OperationGenerate, requestID, true, nil)
				require.NoError(t, err)
			}

			// List logs.
			logs, total, err := svc.ListAuditLogs(ctx, tenantID, tt.offset, tt.limit)
			require.NoError(t, err)
			require.Equal(t, int64(tt.numLogs), total)
			require.Len(t, logs, tt.expectLogs)
		})
	}
}

func TestAuditLogService_ListAuditLogsByElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("list logs for elastic JWK", func(t *testing.T) {
		t.Parallel()

		elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
		auditSvc := NewAuditLogService(testAuditLogRepo, testAuditConfigRepo, testElasticRepo)
		tenantID := googleUuid.New()

		// Set up 100% audit config for reliable testing.
		config := &cryptoutilAppsJoseJaDomain.AuditConfig{
			TenantID:     tenantID,
			Operation:    cryptoutilAppsJoseJaDomain.OperationSign,
			Enabled:      true,
			SamplingRate: 1.0,
		}
		err := auditSvc.UpdateAuditConfig(ctx, tenantID, config)
		require.NoError(t, err)

		// Create elastic JWK.
		elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
		require.NoError(t, err)

		// Log operations for this JWK.
		for i := 0; i < 3; i++ {
			requestID := googleUuid.New().String()
			err := auditSvc.LogOperation(ctx, tenantID, &elasticJWK.ID, cryptoutilAppsJoseJaDomain.OperationSign, requestID, true, nil)
			require.NoError(t, err)
		}

		// List logs by elastic JWK.
		logs, total, err := auditSvc.ListAuditLogsByElasticJWK(ctx, tenantID, elasticJWK.ID, 0, 10)
		require.NoError(t, err)
		require.Equal(t, int64(3), total)
		require.Len(t, logs, 3)

		for _, log := range logs {
			require.NotNil(t, log.ElasticJWKID)
			require.Equal(t, elasticJWK.ID, *log.ElasticJWKID)
		}
	})

	t.Run("wrong tenant returns error", func(t *testing.T) {
		t.Parallel()

		elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
		auditSvc := NewAuditLogService(testAuditLogRepo, testAuditConfigRepo, testElasticRepo)
		tenantID := googleUuid.New()

		// Create elastic JWK.
		elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
		require.NoError(t, err)

		// Try to list with wrong tenant.
		wrongTenantID := googleUuid.New()

		_, _, err = auditSvc.ListAuditLogsByElasticJWK(ctx, wrongTenantID, elasticJWK.ID, 0, 10)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
	})
}

func TestAuditLogService_ListAuditLogsByOperation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("list logs by operation", func(t *testing.T) {
		t.Parallel()

		svc := NewAuditLogService(testAuditLogRepo, testAuditConfigRepo, testElasticRepo)
		tenantID := googleUuid.New()

		// Set up 100% audit config for all operations.
		for _, op := range []string{cryptoutilAppsJoseJaDomain.OperationGenerate, cryptoutilAppsJoseJaDomain.OperationSign, cryptoutilAppsJoseJaDomain.OperationVerify} {
			config := &cryptoutilAppsJoseJaDomain.AuditConfig{
				TenantID:     tenantID,
				Operation:    op,
				Enabled:      true,
				SamplingRate: 1.0,
			}
			err := svc.UpdateAuditConfig(ctx, tenantID, config)
			require.NoError(t, err)
		}

		// Create various operations.
		operations := []string{
			cryptoutilAppsJoseJaDomain.OperationGenerate,
			cryptoutilAppsJoseJaDomain.OperationSign,
			cryptoutilAppsJoseJaDomain.OperationSign,
			cryptoutilAppsJoseJaDomain.OperationVerify,
			cryptoutilAppsJoseJaDomain.OperationSign,
		}

		for _, op := range operations {
			requestID := googleUuid.New().String()
			err := svc.LogOperation(ctx, tenantID, nil, op, requestID, true, nil)
			require.NoError(t, err)
		}

		// List by sign operation.
		logs, total, err := svc.ListAuditLogsByOperation(ctx, tenantID, cryptoutilAppsJoseJaDomain.OperationSign, 0, 10)
		require.NoError(t, err)
		require.Equal(t, int64(3), total)
		require.Len(t, logs, 3)

		for _, log := range logs {
			require.Equal(t, cryptoutilAppsJoseJaDomain.OperationSign, log.Operation)
		}
	})
}

func TestAuditLogService_GetAuditConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("get default config for new tenant", func(t *testing.T) {
		t.Parallel()

		svc := NewAuditLogService(testAuditLogRepo, testAuditConfigRepo, testElasticRepo)
		tenantID := googleUuid.New()

		config, err := svc.GetAuditConfig(ctx, tenantID)
		require.NoError(t, err)
		require.NotNil(t, config)
		require.Equal(t, tenantID, config.TenantID)
		require.True(t, config.Enabled)
		require.Equal(t, float64(1.0), config.SamplingRate)
	})

	t.Run("get existing config", func(t *testing.T) {
		t.Parallel()

		svc := NewAuditLogService(testAuditLogRepo, testAuditConfigRepo, testElasticRepo)
		tenantID := googleUuid.New()

		// Create custom config.
		customConfig := &cryptoutilAppsJoseJaDomain.AuditConfig{
			TenantID:     tenantID,
			Operation:    cryptoutilAppsJoseJaDomain.OperationSign,
			Enabled:      false,
			SamplingRate: 0.5,
		}
		err := svc.UpdateAuditConfig(ctx, tenantID, customConfig)
		require.NoError(t, err)

		// Get config.
		config, getErr := svc.GetAuditConfig(ctx, tenantID)
		require.NoError(t, getErr)
		require.NotNil(t, config)
		require.Equal(t, tenantID, config.TenantID)
	})
}

func TestAuditLogService_UpdateAuditConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("update audit config", func(t *testing.T) {
		t.Parallel()

		svc := NewAuditLogService(testAuditLogRepo, testAuditConfigRepo, testElasticRepo)
		tenantID := googleUuid.New()

		// Create config.
		config := &cryptoutilAppsJoseJaDomain.AuditConfig{
			TenantID:     tenantID,
			Operation:    cryptoutilAppsJoseJaDomain.OperationEncrypt,
			Enabled:      true,
			SamplingRate: 0.75,
		}

		err := svc.UpdateAuditConfig(ctx, tenantID, config)
		require.NoError(t, err)
	})

	t.Run("update sets tenant ID", func(t *testing.T) {
		t.Parallel()

		svc := NewAuditLogService(testAuditLogRepo, testAuditConfigRepo, testElasticRepo)
		tenantID := googleUuid.New()

		// Create config without tenant ID.
		config := &cryptoutilAppsJoseJaDomain.AuditConfig{
			Operation:    cryptoutilAppsJoseJaDomain.OperationDecrypt,
			Enabled:      false,
			SamplingRate: 0.25,
		}

		err := svc.UpdateAuditConfig(ctx, tenantID, config)
		require.NoError(t, err)

		// Verify tenant ID was set.
		require.Equal(t, tenantID, config.TenantID)
	})
}

func TestAuditLogService_CleanupOldLogs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("cleanup with no old logs", func(t *testing.T) {
		t.Parallel()

		svc := NewAuditLogService(testAuditLogRepo, testAuditConfigRepo, testElasticRepo)
		tenantID := googleUuid.New()

		// Set up 100% audit config for reliable testing.
		config := &cryptoutilAppsJoseJaDomain.AuditConfig{
			TenantID:     tenantID,
			Operation:    cryptoutilAppsJoseJaDomain.OperationGenerate,
			Enabled:      true,
			SamplingRate: 1.0,
		}
		err := svc.UpdateAuditConfig(ctx, tenantID, config)
		require.NoError(t, err)

		// Create some recent logs.
		for i := 0; i < 3; i++ {
			requestID := googleUuid.New().String()
			err := svc.LogOperation(ctx, tenantID, nil, cryptoutilAppsJoseJaDomain.OperationGenerate, requestID, true, nil)
			require.NoError(t, err)
		}

		// Cleanup logs older than 30 days.
		count, cleanupErr := svc.CleanupOldLogs(ctx, tenantID, 30)
		require.NoError(t, cleanupErr)
		require.Equal(t, int64(0), count) // No old logs to delete.

		// Verify logs still exist.
		logs, total, listErr := svc.ListAuditLogs(ctx, tenantID, 0, 10)
		require.NoError(t, listErr)
		require.Equal(t, int64(3), total)
		require.Len(t, logs, 3)
	})
}

func TestAuditLogService_LogOperation_AuditDisabled(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("audit disabled for operation", func(t *testing.T) {
		t.Parallel()

		svc := NewAuditLogService(testAuditLogRepo, testAuditConfigRepo, testElasticRepo)
		tenantID := googleUuid.New()

		// Disable audit for the operation.
		config := &cryptoutilAppsJoseJaDomain.AuditConfig{
			TenantID:     tenantID,
			Operation:    cryptoutilAppsJoseJaDomain.OperationSign,
			Enabled:      false,
			SamplingRate: 0.0,
		}
		err := svc.UpdateAuditConfig(ctx, tenantID, config)
		require.NoError(t, err)

		// Log operation - should return nil without creating log.
		requestID := googleUuid.New().String()
		err = svc.LogOperation(ctx, tenantID, nil, cryptoutilAppsJoseJaDomain.OperationSign, requestID, true, nil)
		require.NoError(t, err)

		// Verify no log was created.
		logs, total, listErr := svc.ListAuditLogs(ctx, tenantID, 0, 10)
		require.NoError(t, listErr)
		require.Equal(t, int64(0), total)
		require.Empty(t, logs)
	})
}

func TestAuditLogService_ListAuditLogsByElasticJWK_WrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("wrong tenant should fail", func(t *testing.T) {
		t.Parallel()

		auditSvc := NewAuditLogService(testAuditLogRepo, testAuditConfigRepo, testElasticRepo)
		elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
		tenantID := googleUuid.New()

		// Create an elastic JWK.
		elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseAlgRS256, cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
		require.NoError(t, err)

		// Try to list logs with wrong tenant - should fail.
		wrongTenantID := googleUuid.New()
		_, _, err = auditSvc.ListAuditLogsByElasticJWK(ctx, wrongTenantID, elasticJWK.ID, 0, 10)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
	})
}

func TestAuditLogService_ListAuditLogsByElasticJWK_NonExistent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("non-existent elastic JWK should fail", func(t *testing.T) {
		t.Parallel()

		auditSvc := NewAuditLogService(testAuditLogRepo, testAuditConfigRepo, testElasticRepo)
		tenantID := googleUuid.New()

		// Try to list logs for non-existent elastic JWK - should fail.
		_, _, err := auditSvc.ListAuditLogsByElasticJWK(ctx, tenantID, googleUuid.New(), 0, 10)
		require.Error(t, err)
	})
}

// stringPtr returns a pointer to the given string.
func stringPtr(s string) *string {
	return &s
}
