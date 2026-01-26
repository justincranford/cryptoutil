// Copyright (c) 2025 Justin Cranford
//

package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver.

	cryptoutilAppsJoseJaDomain "cryptoutil/internal/apps/jose/ja/domain"
	cryptoutilAppsJoseJaRepository "cryptoutil/internal/apps/jose/ja/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"
)

// createClosedServiceDependencies creates a new in-memory SQLite database,
// applies migrations, creates all repositories and services, then closes
// the database connection to force database errors in service operations.
func createClosedServiceDependencies() (*gorm.DB, cryptoutilAppsJoseJaRepository.ElasticJWKRepository, cryptoutilAppsJoseJaRepository.MaterialJWKRepository, cryptoutilAppsJoseJaRepository.AuditLogRepository, cryptoutilAppsJoseJaRepository.AuditConfigRepository, error) {
	ctx := context.Background()
	dbID, _ := cryptoutilSharedUtilRandom.GenerateUUIDv7()
	dsn := "file:" + dbID.String() + "?mode=memory&cache=shared"

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to open SQLite: %w", err)
	}

	// Configure SQLite.
	_, _ = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	_, _ = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")

	// Create GORM DB.
	gormDB, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err != nil {
		_ = sqlDB.Close()

		return nil, nil, nil, nil, nil, fmt.Errorf("failed to create GORM DB: %w", err)
	}

	// Apply migrations.
	if err := cryptoutilAppsJoseJaRepository.ApplyJoseJAMigrations(sqlDB, cryptoutilAppsJoseJaRepository.DatabaseTypeSQLite); err != nil {
		_ = sqlDB.Close()

		return nil, nil, nil, nil, nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Create repositories BEFORE closing.
	elasticRepo := cryptoutilAppsJoseJaRepository.NewElasticJWKRepository(gormDB)
	materialRepo := cryptoutilAppsJoseJaRepository.NewMaterialJWKRepository(gormDB)
	auditLogRepo := cryptoutilAppsJoseJaRepository.NewAuditLogRepository(gormDB)
	auditConfigRepo := cryptoutilAppsJoseJaRepository.NewAuditConfigRepository(gormDB)

	// Close the underlying connection to force errors.
	if err := sqlDB.Close(); err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("failed to close database: %w", err)
	}

	return gormDB, elasticRepo, materialRepo, auditLogRepo, auditConfigRepo, nil
}

// ====================
// ElasticJWK Service Database Error Tests
// ====================

func TestElasticJWKService_CreateDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	// Valid parameters, but database is closed.
	_, _, err = svc.CreateElasticJWK(ctx, googleUuid.New(), "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to create elastic JWK"))
}

func TestElasticJWKService_GetDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err = svc.GetElasticJWK(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get elastic JWK"))
}

func TestElasticJWKService_ListDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, _, err = svc.ListElasticJWKs(ctx, googleUuid.New(), 0, 10)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to list elastic JWKs"))
}

func TestElasticJWKService_DeleteDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewElasticJWKService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	err = svc.DeleteElasticJWK(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on any database operation.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

// ====================
// AuditLog Service Database Error Tests
// ====================

func TestAuditLogService_LogOperationDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, _, auditLogRepo, auditConfigRepo, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	elasticJWKID := googleUuid.New()
	err = svc.LogOperation(ctx, googleUuid.New(), &elasticJWKID, "test-operation", "test-request-id", true, nil)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to create audit log"))
}

func TestAuditLogService_ListAuditLogsDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, _, auditLogRepo, auditConfigRepo, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, _, err = svc.ListAuditLogs(ctx, googleUuid.New(), 0, 10)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to list audit logs"))
}

func TestAuditLogService_ListAuditLogsByElasticJWKDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, _, auditLogRepo, auditConfigRepo, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, _, err = svc.ListAuditLogsByElasticJWK(ctx, googleUuid.New(), googleUuid.New(), 0, 10)
	require.Error(t, err)
	// Could fail on get elastic JWK or list logs.
	require.True(t,
		strings.Contains(err.Error(), "failed to get elastic JWK") ||
			strings.Contains(err.Error(), "failed to list audit logs"),
		"Expected get or list error, got: %v", err)
}

func TestAuditLogService_ListAuditLogsByOperationDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, _, auditLogRepo, auditConfigRepo, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, _, err = svc.ListAuditLogsByOperation(ctx, googleUuid.New(), "test-operation", 0, 10)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to list audit logs"))
}

func TestAuditLogService_GetAuditConfigDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, _, auditLogRepo, auditConfigRepo, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, err = svc.GetAuditConfig(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to get audit config"))
}

func TestAuditLogService_UpdateAuditConfigDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, _, auditLogRepo, auditConfigRepo, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	tenantID := googleUuid.New()
	config := &cryptoutilAppsJoseJaDomain.AuditConfig{
		TenantID:     tenantID,
		Operation:    "test-operation",
		Enabled:      true,
		SamplingRate: 1.0,
	}

	err = svc.UpdateAuditConfig(ctx, tenantID, config)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to update audit config"))
}

func TestAuditLogService_CleanupOldLogsDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, _, auditLogRepo, auditConfigRepo, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewAuditLogService(auditLogRepo, auditConfigRepo, elasticRepo)

	_, err = svc.CleanupOldLogs(ctx, googleUuid.New(), 30)
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to delete old audit logs"))
}

// ====================
// MaterialRotation Service Database Error Tests
// ====================

func TestMaterialRotationService_ListMaterialsDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err = svc.ListMaterials(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on get elastic JWK or list materials.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_GetActiveMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err = svc.GetActiveMaterial(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on get elastic JWK or get active material.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_GetMaterialByKIDDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err = svc.GetMaterialByKID(ctx, googleUuid.New(), googleUuid.New(), "test-kid")
	require.Error(t, err)
	// Could fail on get elastic JWK or get material by KID.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_RotateMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	_, err = svc.RotateMaterial(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on get elastic JWK or later operations.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestMaterialRotationService_RetireMaterialDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewMaterialRotationService(elasticRepo, materialRepo, testJWKGenService, testBarrierService)

	err = svc.RetireMaterial(ctx, googleUuid.New(), googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	// Could fail on get elastic JWK, get material by KID, or update.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

// ====================
// JWE Service Database Error Tests
// ====================

func TestJWEService_EncryptDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWEService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.Encrypt(ctx, googleUuid.New(), googleUuid.New(), []byte("test plaintext"))
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWEService_DecryptDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWEService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.Decrypt(ctx, googleUuid.New(), googleUuid.New(), "eyJhbGciOiJSU0EtT0FFUC0yNTYiLCJlbmMiOiJBMjU2R0NNIn0.test.test.test.test")
	require.Error(t, err)
	// Could fail on parse, get elastic JWK, or decrypt.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "parse"),
		"Expected database, not-found, or parse error, got: %v", err)
}

func TestJWEService_EncryptWithKIDDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWEService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.EncryptWithKID(ctx, googleUuid.New(), googleUuid.New(), "test-kid", []byte("test plaintext"))
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

// ====================
// JWS Service Database Error Tests
// ====================

func TestJWSService_SignDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWSService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.Sign(ctx, googleUuid.New(), googleUuid.New(), []byte("test payload"))
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWSService_VerifyDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWSService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.Verify(ctx, googleUuid.New(), googleUuid.New(), "eyJhbGciOiJSUzI1NiJ9.dGVzdA.test")
	require.Error(t, err)
	// Could fail on parse, get elastic JWK, or verify.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "parse"),
		"Expected database, not-found, or parse error, got: %v", err)
}

func TestJWSService_SignWithKIDDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWSService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.SignWithKID(ctx, googleUuid.New(), googleUuid.New(), "test-kid", []byte("test payload"))
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

// ====================
// JWT Service Database Error Tests
// ====================

func TestJWTService_CreateJWTDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWTService(elasticRepo, materialRepo, testBarrierService)

	claims := &JWTClaims{Issuer: "test-issuer"}
	_, err = svc.CreateJWT(ctx, googleUuid.New(), googleUuid.New(), claims)
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWTService_ValidateJWTDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWTService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.ValidateJWT(ctx, googleUuid.New(), googleUuid.New(), "eyJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJ0ZXN0In0.test")
	require.Error(t, err)
	// Could fail on parse, get elastic JWK, or validate.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "parse"),
		"Expected database, not-found, or parse error, got: %v", err)
}

func TestJWTService_CreateEncryptedJWTDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWTService(elasticRepo, materialRepo, testBarrierService)

	claims := &JWTClaims{Issuer: "test-issuer"}
	_, err = svc.CreateEncryptedJWT(ctx, googleUuid.New(), googleUuid.New(), googleUuid.New(), claims)
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

// ====================
// JWKS Service Database Error Tests
// ====================

func TestJWKSService_GetJWKSDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWKSService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.GetJWKS(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWKSService_GetJWKSForElasticKeyDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWKSService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.GetJWKSForElasticKey(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWKSService_GetPublicJWKDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWKSService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.GetPublicJWK(ctx, googleUuid.New(), "test-kid")
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

// ============================================================================
// Crypto Error Path Tests
// ============================================================================
// These tests exercise error paths in crypto operations by corrupting data
// in the database. This is the only way to test these paths since the service
// reads data from the repository before performing crypto operations.

// TestJWSService_Sign_CorruptedBase64 tests that Sign returns error when
// material's PrivateJWKJWE contains invalid base64.
func TestJWSService_Sign_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PrivateJWKJWE with invalid base64 directly in DB.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to sign - should fail on base64 decode.
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode private JWK JWE")
}

// TestJWSService_Sign_CorruptedJWE tests that Sign returns error when
// material's PrivateJWKJWE contains valid base64 but invalid JWE content.
func TestJWSService_Sign_CorruptedJWE(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PrivateJWKJWE with valid base64 but invalid JWE.
	// Use base64 encoding of "not a valid JWE" string.
	invalidJWE := "bm90IGEgdmFsaWQgSldF" // base64("not a valid JWE")
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", invalidJWE).Error
	require.NoError(t, err)

	// Try to sign - should fail on barrier decrypt.
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decrypt private JWK")
}

// TestJWEService_Encrypt_CorruptedBase64 tests that Encrypt returns error when
// material's PublicJWKJWE contains invalid base64.
func TestJWEService_Encrypt_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	// Use RSA/2048 key type which maps to RSA-OAEP-256 algorithm for encryption.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PublicJWKJWE with invalid base64.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("public_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to encrypt - should fail on base64 decode.
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

// TestJWEService_Decrypt_CorruptedBase64 tests that Decrypt returns error when
// material's PrivateJWKJWE contains invalid base64.
// NOTE: Decrypt loops through all materials and catches errors silently,
// so when decoding fails it returns "no matching key found" not the decode error.
func TestJWEService_Decrypt_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	// Use RSA/2048 key type which maps to RSA-OAEP-256 algorithm for encryption.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// First encrypt something valid.
	jweCompact, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.NoError(t, err)
	require.NotEmpty(t, jweCompact)

	// Now corrupt the material's PrivateJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to decrypt - Decrypt catches decode errors and tries next material.
	// Since we only have one material and it fails, returns "no matching key found".
	_, err = jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, jweCompact)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

// TestJWSService_Verify_CorruptedBase64 tests that Verify returns error when
// material's PublicJWKJWE contains invalid base64.
// NOTE: Verify loops through all materials and catches errors silently,
// so when decoding fails it returns "no matching key found" not the decode error.
func TestJWSService_Verify_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// First sign something valid.
	jwsCompact, err := jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test payload"))
	require.NoError(t, err)
	require.NotEmpty(t, jwsCompact)

	// Now corrupt the material's PublicJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("public_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to verify - Verify catches decode errors and tries next material.
	// Since we only have one material and it fails, returns "no matching key found".
	_, err = jwsSvc.Verify(ctx, tenantID, elasticJWK.ID, jwsCompact)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no matching key found")
}

// TestJWKSService_GetPublicJWK_CorruptedBase64 tests that GetPublicJWK returns
// error when material's PublicJWKJWE contains invalid base64.
func TestJWKSService_GetPublicJWK_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PublicJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("public_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to get public JWK - should fail on base64 decode.
	// GetPublicJWK signature: (ctx, tenantID, kid).
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	_, err = jwksSvc.GetPublicJWK(ctx, tenantID, material.MaterialKID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

// TestJWKSService_GetJWKS_CorruptedBase64 tests that GetJWKS skips materials
// with corrupted PublicJWKJWE (graceful degradation).
// NOTE: GetJWKS uses continue pattern - it skips corrupted materials rather than failing.
func TestJWKSService_GetJWKS_CorruptedBase64(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PublicJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("public_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// GetJWKS skips corrupted materials and returns empty JWKS (graceful degradation).
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	jwks, err := jwksSvc.GetJWKS(ctx, tenantID)
	require.NoError(t, err)
	// The corrupted material is skipped, resulting in empty keys.
	require.Empty(t, jwks.Keys)
}

// ====================
// JWT Service Corrupted Key Tests
// ====================

// TestJWTService_ValidateJWT_CorruptedPublicKeyDB tests that ValidateJWT returns error
// when material's PublicJWKJWE is corrupted in the database.
func TestJWTService_ValidateJWT_CorruptedPublicKeyDB(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Create a valid JWT first.
	claims := &JWTClaims{
		Subject:  "test-subject",
		Issuer:   "test-issuer",
		Audience: []string{"test-audience"},
	}
	token, err := jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Corrupt the material's PublicJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("public_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to validate - should fail on base64 decode.
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, elasticJWK.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode public JWK JWE")
}

// TestJWTService_CreateJWT_CorruptedPrivateKeyDB tests that CreateJWT returns error
// when material's PrivateJWKJWE is corrupted in the database.
func TestJWTService_CreateJWT_CorruptedPrivateKeyDB(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create elastic JWK with material using real services.
	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PrivateJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", "not-valid-base64!!!").Error
	require.NoError(t, err)

	// Try to create JWT - should fail on base64 decode.
	claims := &JWTClaims{
		Subject:  "test-subject",
		Issuer:   "test-issuer",
		Audience: []string{"test-audience"},
	}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, elasticJWK.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode private JWK JWE")
}

// ====================
// CreateElasticJWK Algorithm Validation Tests
// ====================

// TestElasticJWKService_CreateElasticJWK_UnsupportedAlgorithm tests that CreateElasticJWK
// returns error for unsupported algorithm.
func TestElasticJWKService_CreateElasticJWK_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Try to create with an unsupported algorithm.
	_, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "INVALID-ALGORITHM", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid algorithm")
}

// TestElasticJWKService_CreateElasticJWK_EmptyAlgorithm tests that CreateElasticJWK
// returns error for empty algorithm.
func TestElasticJWKService_CreateElasticJWK_EmptyAlgorithm(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Try to create with empty algorithm.
	_, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.Error(t, err)
}

// ====================
// Delete Cascading Error Tests
// ====================

// TestElasticJWKService_DeleteElasticJWK_WithMultipleMaterials tests that DeleteElasticJWK
// correctly deletes all materials before deleting the elastic JWK.
func TestElasticJWKService_DeleteElasticJWK_WithMultipleMaterials(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK with initial material.
	elasticJWK, material1, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material1)

	// Rotate to create second material.
	material2, err := rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.NotNil(t, material2)

	// Rotate to create third material.
	material3, err := rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.NotNil(t, material3)

	// Verify we have 3 materials.
	materials, err := rotationSvc.ListMaterials(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Len(t, materials, 3)

	// Delete elastic JWK should cascade delete all materials.
	err = elasticSvc.DeleteElasticJWK(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)

	// Verify elastic JWK is deleted.
	_, err = elasticSvc.GetElasticJWK(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not found")
}

// ====================
// JWE/JWS Decrypt/Verify with Barrier Errors
// ====================

// TestJWEService_Decrypt_CorruptedJWEInDB tests that Decrypt returns error
// when the decrypted key is malformed.
func TestJWEService_Decrypt_CorruptedJWEInDB(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Create valid JWE first.
	jweCompact, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test data"))
	require.NoError(t, err)
	require.NotEmpty(t, jweCompact)

	// Corrupt the material's PrivateJWKJWE (base64 decode will fail).
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", "invalid-base64!!!").Error
	require.NoError(t, err)

	// Try to decrypt - should fail due to corrupted stored key.
	_, err = jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, jweCompact)
	require.Error(t, err)
	// Should get "no matching key found" since decode fails and it continues to next material.
	require.Contains(t, err.Error(), "no matching key found")
}

// TestJWSService_Sign_CorruptedPrivateKeyInDB tests that Sign returns error
// when the stored private key is corrupted.
func TestJWSService_Sign_CorruptedPrivateKeyInDB(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Corrupt the material's PrivateJWKJWE.
	err = testDB.Model(&cryptoutilAppsJoseJaDomain.MaterialJWK{}).
		Where("id = ?", material.ID).
		Update("private_jwk_jwe", "invalid-base64!!!").Error
	require.NoError(t, err)

	// Try to sign - should fail due to corrupted stored key.
	_, err = jwsSvc.Sign(ctx, tenantID, elasticJWK.ID, []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to decode private JWK JWE")
}

// ====================
// Material Rotation with Corrupted Data Tests
// ====================

// TestMaterialRotationService_RotateMaterial_CreatesMaterialSuccessfully tests that
// RotateMaterial creates a new material and marks it as active.
func TestMaterialRotationService_RotateMaterial_CreatesMaterialSuccessfully(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK with initial material.
	elasticJWK, initialMaterial, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, initialMaterial)
	require.True(t, initialMaterial.Active)

	// Rotate creates a new active material.
	newMaterial, err := rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.NotNil(t, newMaterial)
	require.True(t, newMaterial.Active)
	require.NotEqual(t, initialMaterial.ID, newMaterial.ID)
}

// ====================
// Symmetric Key Tests (oct type)
// ====================

// TestElasticJWKService_CreateElasticJWK_SymmetricKey tests that CreateElasticJWK
// correctly handles symmetric (oct) keys where there's no separate public key.
func TestElasticJWKService_CreateElasticJWK_SymmetricKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create symmetric key.
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeOct256, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)
	require.NotNil(t, elasticJWK)
	require.NotNil(t, material)

	// Verify the key was created.
	retrieved, err := elasticSvc.GetElasticJWK(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)
	require.Equal(t, elasticJWK.ID, retrieved.ID)
}

// TestJWEService_EncryptDecrypt_SymmetricKey tests JWE operations with symmetric keys.
func TestJWEService_EncryptDecrypt_SymmetricKey(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create symmetric key for direct encryption.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeOct256, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)

	plaintext := []byte("secret message for symmetric key")
	jweCompact, err := jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, plaintext)
	require.NoError(t, err)
	require.NotEmpty(t, jweCompact)

	decrypted, err := jweSvc.Decrypt(ctx, tenantID, elasticJWK.ID, jweCompact)
	require.NoError(t, err)
	require.Equal(t, plaintext, decrypted)
}

// ====================
// Additional Error Path Tests
// ====================

// TestJWEService_Encrypt_WrongKeyUse tests that Encrypt fails when key is not for encryption.
func TestJWEService_Encrypt_WrongKeyUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create a signing key (not encryption).
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to encrypt with signing key - should fail.
	_, err = jweSvc.Encrypt(ctx, tenantID, elasticJWK.ID, []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured for encryption")
}

// Removed TestJWEService_Decrypt_WrongKeyUse - Decrypt doesn't validate key use, just fails with "no matching key"

// TestJWEService_EncryptWithKID_WrongKeyUse tests that EncryptWithKID fails when key is not for encryption.
func TestJWEService_EncryptWithKID_WrongKeyUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create a signing key (not encryption).
	elasticJWK, material, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to encrypt with signing key - should fail.
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, material.MaterialKID, []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured for encryption")
}

// TestJWEService_EncryptWithKID_MaterialNotFound tests that EncryptWithKID fails when material not found.
func TestJWEService_EncryptWithKID_MaterialNotFound(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Try to encrypt with non-existent material KID.
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK.ID, "nonexistent-kid", []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get material")
}

// TestJWEService_EncryptWithKID_MaterialWrongElasticJWK tests EncryptWithKID fails when material
// belongs to different elastic JWK.
func TestJWEService_EncryptWithKID_MaterialWrongElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jweSvc := NewJWEService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create two encryption keys.
	elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)

	elasticJWK2, material2, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Try to encrypt using elasticJWK1 but with material2's KID.
	_, err = jweSvc.EncryptWithKID(ctx, tenantID, elasticJWK1.ID, material2.MaterialKID, []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "material key does not belong to elastic JWK")

	_ = elasticJWK2 // Use variable to avoid unused warning.
}

// TestMaterialRotationService_RotateMaterial_MaxMaterialsReached tests that RotateMaterial fails
// when max materials is reached.
func TestMaterialRotationService_RotateMaterial_MaxMaterialsReached(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create elastic JWK with max 2 materials.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 2)
	require.NoError(t, err)

	// Rotate once - should succeed (now have 2 materials).
	_, err = rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.NoError(t, err)

	// Rotate again - should fail (max reached).
	_, err = rotationSvc.RotateMaterial(ctx, tenantID, elasticJWK.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "max materials reached")
}

// TestMaterialRotationService_RetireMaterial_MaterialWrongElasticJWK tests that RetireMaterial fails
// when material doesn't belong to the specified elastic JWK.
func TestMaterialRotationService_RetireMaterial_MaterialWrongElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	rotationSvc := NewMaterialRotationService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	tenantID := googleUuid.New()

	// Create two elastic JWKs.
	elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	elasticJWK2, material2, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to retire material2 via elasticJWK1 - should fail.
	err = rotationSvc.RetireMaterial(ctx, tenantID, elasticJWK1.ID, material2.ID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "material not found for this elastic JWK")

	_ = elasticJWK2 // Use variable to avoid unused warning.
}

// TestJWTService_CreateEncryptedJWT_EncryptionKeyWrongTenant tests that CreateEncryptedJWT fails
// when encryption key has wrong tenant.
func TestJWTService_CreateEncryptedJWT_EncryptionKeyWrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	otherTenantID := googleUuid.New()

	// Create signing key for tenant1.
	signingKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Create encryption key for tenant2.
	encryptionKey, _, err := elasticSvc.CreateElasticJWK(ctx, otherTenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Try to create encrypted JWT using tenant1's signing key but tenant2's encryption key.
	claims := &JWTClaims{
		Subject:   "test-user",
		ExpiresAt: timePtr(time.Now().UTC().Add(time.Hour)),
	}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, signingKey.ID, encryptionKey.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "encryption key not found")
}

// TestJWTService_CreateEncryptedJWT_EncryptionKeyWrongUse tests that CreateEncryptedJWT fails
// when encryption key is not configured for encryption.
func TestJWTService_CreateEncryptedJWT_EncryptionKeyWrongUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	signingKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Create another signing key (not encryption).
	anotherSigningKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to create encrypted JWT using signing key for encryption.
	claims := &JWTClaims{
		Subject:   "test-user",
		ExpiresAt: timePtr(time.Now().UTC().Add(time.Hour)),
	}
	_, err = jwtSvc.CreateEncryptedJWT(ctx, tenantID, signingKey.ID, anotherSigningKey.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured for encryption")
}

// TestJWTService_ValidateJWT_NoHeaders tests that ValidateJWT fails when JWT has no headers.
func TestJWTService_ValidateJWT_NoHeaders(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	signingKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Create a malformed token (not a valid JWT structure).
	invalidToken := "invalid.token.structure"

	// Try to validate - should fail parsing.
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, signingKey.ID, invalidToken)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse JWT")
}

// TestJWTService_ValidateJWT_MaterialKIDNotBelongToElasticJWK tests that ValidateJWT fails
// when the material KID doesn't belong to the expected elastic JWK.
func TestJWTService_ValidateJWT_MaterialKIDNotBelongToElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create two signing keys.
	signingKey1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	signingKey2, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Create JWT with signing key 1.
	claims := &JWTClaims{
		Subject:   "test-user",
		ExpiresAt: timePtr(time.Now().UTC().Add(time.Hour)),
	}
	token, err := jwtSvc.CreateJWT(ctx, tenantID, signingKey1.ID, claims)
	require.NoError(t, err)

	// Try to validate with signing key 2 - should fail because KID doesn't belong.
	_, err = jwtSvc.ValidateJWT(ctx, tenantID, signingKey2.ID, token)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not belong to this elastic JWK")
}

// TestJWKSService_GetJWKS_WrongTenant tests that GetJWKS returns empty for wrong tenant.
func TestJWKSService_GetJWKS_EmptyForWrongTenant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()
	otherTenantID := googleUuid.New()

	// Create signing key for one tenant.
	_, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Get JWKS with different tenant - should return empty (not an error).
	jwks, err := jwksSvc.GetJWKS(ctx, otherTenantID)
	require.NoError(t, err)
	require.Empty(t, jwks.Keys)
}

// TestJWKSService_GetPublicJWK_WrongTenant tests that GetPublicJWK returns not found for wrong tenant.
func TestJWKSService_GetPublicJWK_WrongKID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	jwksSvc := NewJWKSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Try to get public JWK with non-existent KID.
	_, err := jwksSvc.GetPublicJWK(ctx, tenantID, "nonexistent-kid")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get material")
}

// TestJWSService_SignWithKID_WrongKID tests that SignWithKID fails when KID not found.
func TestJWSService_SignWithKID_WrongKID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create signing key.
	elasticJWK, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to sign with non-existent KID.
	_, err = jwsSvc.SignWithKID(ctx, tenantID, elasticJWK.ID, "nonexistent-kid", []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get material")
}

// TestJWSService_SignWithKID_MaterialWrongElasticJWK tests that SignWithKID fails
// when material belongs to different elastic JWK.
func TestJWSService_SignWithKID_MaterialWrongElasticJWK(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwsSvc := NewJWSService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create two signing keys.
	elasticJWK1, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	elasticJWK2, material2, err := elasticSvc.CreateElasticJWK(ctx, tenantID, "RS256", cryptoutilAppsJoseJaDomain.KeyUseSig, 10)
	require.NoError(t, err)

	// Try to sign using elasticJWK1 but with material2's KID.
	_, err = jwsSvc.SignWithKID(ctx, tenantID, elasticJWK1.ID, material2.MaterialKID, []byte("test data"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "material key does not belong to elastic JWK")

	_ = elasticJWK2 // Use variable to avoid unused warning.
}

// TestJWTService_CreateJWT_WrongKeyUse tests that CreateJWT fails when key is for encryption.
func TestJWTService_CreateJWT_WrongKeyUse(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	elasticSvc := NewElasticJWKService(testElasticRepo, testMaterialRepo, testJWKGenService, testBarrierService)
	jwtSvc := NewJWTService(testElasticRepo, testMaterialRepo, testBarrierService)
	tenantID := googleUuid.New()

	// Create encryption key (not signing).
	encryptionKey, _, err := elasticSvc.CreateElasticJWK(ctx, tenantID, cryptoutilSharedMagic.JoseKeyTypeRSA2048, cryptoutilAppsJoseJaDomain.KeyUseEnc, 10)
	require.NoError(t, err)

	// Try to create JWT with encryption key - should fail.
	claims := &JWTClaims{
		Subject:   "test-user",
		ExpiresAt: timePtr(time.Now().UTC().Add(time.Hour)),
	}
	_, err = jwtSvc.CreateJWT(ctx, tenantID, encryptionKey.ID, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not configured for signing")
}

// Helper function for time pointer.
func timePtr(t time.Time) *time.Time {
	return &t
}
