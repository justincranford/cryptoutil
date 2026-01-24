// Copyright (c) 2025 Justin Cranford
//
//

package service

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"
	cryptoutilJoseRepository "cryptoutil/internal/jose/repository"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "modernc.org/sqlite" // Pure Go SQLite driver.
)

// setupAuditLogTestDB creates an in-memory SQLite database for testing audit log service.
func setupAuditLogTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file::memory:?cache=shared"
	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err, "failed to open SQLite connection")

	// Configure for testing.
	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA journal_mode=WAL;")
	require.NoError(t, err, "failed to set WAL mode")
	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err, "failed to set busy timeout")

	dialector := sqlite.Dialector{Conn: sqlDB}
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err, "failed to initialize GORM")

	underlyingDB, err := db.DB()
	require.NoError(t, err, "failed to get underlying DB")

	underlyingDB.SetMaxOpenConns(5)
	underlyingDB.SetMaxIdleConns(5)
	underlyingDB.SetConnMaxLifetime(0)

	// Create tenant_realms table (required for foreign key).
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tenant_realms (
			tenant_id TEXT NOT NULL,
			realm_id TEXT NOT NULL,
			name TEXT NOT NULL,
			PRIMARY KEY (tenant_id, realm_id)
		)
	`).Error
	require.NoError(t, err, "failed to create tenant_realms table")

	// Auto-migrate audit tables.
	err = db.AutoMigrate(&cryptoutilJoseDomain.AuditConfig{}, &cryptoutilJoseDomain.AuditLogEntry{})
	require.NoError(t, err, "failed to auto-migrate audit tables")

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

// createTestTenantRealm creates a test tenant realm in the database.
func createTestTenantRealm(t *testing.T, db *gorm.DB, tenantID, realmID googleUuid.UUID) {
	t.Helper()

	err := db.Exec(
		"INSERT INTO tenant_realms (tenant_id, realm_id, name) VALUES (?, ?, ?)",
		tenantID.String(), realmID.String(), "test-realm",
	).Error
	require.NoError(t, err, "failed to create test tenant realm")
}

func TestNewAuditLogService(t *testing.T) {
	db := setupAuditLogTestDB(t)
	configRepo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	configService := NewAuditConfigService(configRepo)
	logRepo := cryptoutilJoseRepository.NewAuditLogGormRepository(db)

	service := NewAuditLogService(configService, logRepo)
	require.NotNil(t, service)
}

func TestAuditLogService_Log_WhenEnabled(t *testing.T) {
	db := setupAuditLogTestDB(t)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	createTestTenantRealm(t, db, tenantID, realmID)

	configRepo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	configService := NewAuditConfigService(configRepo)
	logRepo := cryptoutilJoseRepository.NewAuditLogGormRepository(db)
	service := NewAuditLogService(configService, logRepo)

	// Initialize defaults (enabled=true, samplingRate=100% for deterministic test).
	err := configService.InitializeDefaults(ctx, tenantID)
	require.NoError(t, err)

	// Set sampling rate to 100% for deterministic test.
	err = configService.SetConfig(ctx, tenantID, AuditOperationSign, true, 1.0)
	require.NoError(t, err)

	// Log an operation.
	logged, err := service.Log(ctx, AuditLogParams{
		TenantID:     tenantID,
		RealmID:      realmID,
		Operation:    AuditOperationSign,
		ResourceType: "elastic_jwk",
		ResourceID:   "test-kid",
		Success:      true,
		Metadata:     map[string]any{"test": "value"},
	})
	require.NoError(t, err)
	require.True(t, logged, "audit log should be created when enabled at 100% sampling")

	// Verify the log entry was created.
	entries, err := logRepo.ListByTenantRealm(ctx, tenantID, realmID, 0, 10)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, AuditOperationSign, entries[0].Operation)
	require.Equal(t, "elastic_jwk", entries[0].ResourceType)
	require.Equal(t, "test-kid", entries[0].ResourceID)
	require.True(t, entries[0].Success)
}

func TestAuditLogService_Log_WhenDisabled(t *testing.T) {
	db := setupAuditLogTestDB(t)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	createTestTenantRealm(t, db, tenantID, realmID)

	configRepo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	configService := NewAuditConfigService(configRepo)
	logRepo := cryptoutilJoseRepository.NewAuditLogGormRepository(db)
	service := NewAuditLogService(configService, logRepo)

	// Disable audit logging for this operation.
	err := configService.SetConfig(ctx, tenantID, AuditOperationSign, false, 0.0)
	require.NoError(t, err)

	// Attempt to log an operation.
	logged, err := service.Log(ctx, AuditLogParams{
		TenantID:     tenantID,
		RealmID:      realmID,
		Operation:    AuditOperationSign,
		ResourceType: "elastic_jwk",
		ResourceID:   "test-kid",
		Success:      true,
	})
	require.NoError(t, err)
	require.False(t, logged, "audit log should NOT be created when disabled")

	// Verify no log entry was created.
	entries, err := logRepo.ListByTenantRealm(ctx, tenantID, realmID, 0, 10)
	require.NoError(t, err)
	require.Len(t, entries, 0)
}

func TestAuditLogService_Log_WithSampling(t *testing.T) {
	db := setupAuditLogTestDB(t)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()

	createTestTenantRealm(t, db, tenantID, realmID)

	configRepo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	configService := NewAuditConfigService(configRepo)
	logRepo := cryptoutilJoseRepository.NewAuditLogGormRepository(db)
	service := NewAuditLogService(configService, logRepo)

	// Set sampling rate to 50%.
	err := configService.SetConfig(ctx, tenantID, AuditOperationEncrypt, true, 0.5)
	require.NoError(t, err)

	// Log 100 operations and count how many were actually logged.
	loggedCount := 0

	for i := 0; i < 100; i++ {
		logged, err := service.Log(ctx, AuditLogParams{
			TenantID:     tenantID,
			RealmID:      realmID,
			Operation:    AuditOperationEncrypt,
			ResourceType: "elastic_jwk",
			ResourceID:   "test-kid",
			Success:      true,
		})
		require.NoError(t, err)

		if logged {
			loggedCount++
		}
	}

	// With 50% sampling, we expect roughly 50 entries, but allow variance.
	// Statistical test: 50% sampling on 100 operations should give 30-70 entries (99% CI).
	require.Greater(t, loggedCount, 20, "at 50% sampling, should log at least some entries")
	require.Less(t, loggedCount, 80, "at 50% sampling, should not log too many entries")
}

func TestAuditLogService_LogSuccess(t *testing.T) {
	db := setupAuditLogTestDB(t)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	createTestTenantRealm(t, db, tenantID, realmID)

	configRepo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	configService := NewAuditConfigService(configRepo)
	logRepo := cryptoutilJoseRepository.NewAuditLogGormRepository(db)
	service := NewAuditLogService(configService, logRepo)

	// Enable with 100% sampling.
	err := configService.SetConfig(ctx, tenantID, AuditOperationSign, true, 1.0)
	require.NoError(t, err)

	// Log a successful operation.
	logged, err := service.LogSuccess(ctx, tenantID, realmID, AuditOperationSign, "elastic_jwk", "test-kid", map[string]any{
		"alg": "RS256",
	})
	require.NoError(t, err)
	require.True(t, logged)

	// Verify the entry.
	entries, err := logRepo.ListByTenantRealm(ctx, tenantID, realmID, 0, 10)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.True(t, entries[0].Success)
	require.Nil(t, entries[0].ErrorMessage)
}

func TestAuditLogService_LogFailure(t *testing.T) {
	db := setupAuditLogTestDB(t)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	createTestTenantRealm(t, db, tenantID, realmID)

	configRepo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	configService := NewAuditConfigService(configRepo)
	logRepo := cryptoutilJoseRepository.NewAuditLogGormRepository(db)
	service := NewAuditLogService(configService, logRepo)

	// Enable with 100% sampling.
	err := configService.SetConfig(ctx, tenantID, AuditOperationDecrypt, true, 1.0)
	require.NoError(t, err)

	// Log a failed operation.
	testErr := fmt.Errorf("elastic JWK not found")
	logged, err := service.LogFailure(ctx, tenantID, realmID, AuditOperationDecrypt, "elastic_jwk", "test-kid", testErr, nil)
	require.NoError(t, err)
	require.True(t, logged)

	// Verify the entry.
	entries, err := logRepo.ListByTenantRealm(ctx, tenantID, realmID, 0, 10)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.False(t, entries[0].Success)
	require.NotNil(t, entries[0].ErrorMessage)
	require.Contains(t, *entries[0].ErrorMessage, "not found")
}

func TestAuditLogService_WithUserAndSessionFromContext(t *testing.T) {
	db := setupAuditLogTestDB(t)
	ctx := context.Background()

	tenantID := googleUuid.New()
	realmID := googleUuid.New()
	userID := googleUuid.New()
	sessionID := googleUuid.New()

	createTestTenantRealm(t, db, tenantID, realmID)

	configRepo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	configService := NewAuditConfigService(configRepo)
	logRepo := cryptoutilJoseRepository.NewAuditLogGormRepository(db)
	service := NewAuditLogService(configService, logRepo)

	// Enable with 100% sampling.
	err := configService.SetConfig(ctx, tenantID, AuditOperationVerify, true, 1.0)
	require.NoError(t, err)

	// Add user and session to context.
	ctx = ContextWithUser(ctx, userID)
	ctx = ContextWithSession(ctx, sessionID)

	// Log using LogSuccess which extracts user/session from context.
	logged, err := service.LogSuccess(ctx, tenantID, realmID, AuditOperationVerify, "elastic_jwk", "test-kid", nil)
	require.NoError(t, err)
	require.True(t, logged)

	// Verify the entry has user_id and session_id.
	entries, err := logRepo.ListByTenantRealm(ctx, tenantID, realmID, 0, 10)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.NotNil(t, entries[0].UserID)
	require.Equal(t, userID, *entries[0].UserID)
	require.NotNil(t, entries[0].SessionID)
	require.Equal(t, sessionID, *entries[0].SessionID)
}

func TestContextWithUser(t *testing.T) {
	userID := googleUuid.New()
	ctx := context.Background()

	// Before adding user.
	require.Nil(t, UserFromContext(ctx))

	// After adding user.
	ctx = ContextWithUser(ctx, userID)
	result := UserFromContext(ctx)
	require.NotNil(t, result)
	require.Equal(t, userID, *result)
}

func TestContextWithSession(t *testing.T) {
	sessionID := googleUuid.New()
	ctx := context.Background()

	// Before adding session.
	require.Nil(t, SessionFromContext(ctx))

	// After adding session.
	ctx = ContextWithSession(ctx, sessionID)
	result := SessionFromContext(ctx)
	require.NotNil(t, result)
	require.Equal(t, sessionID, *result)
}
