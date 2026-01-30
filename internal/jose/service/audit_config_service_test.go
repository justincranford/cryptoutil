// Copyright (c) 2025 Justin Cranford
//
//

package service

import (
	"context"
	"database/sql"
	"testing"

	cryptoutilJoseDomain "cryptoutil/internal/jose/domain"
	cryptoutilJoseRepository "cryptoutil/internal/jose/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver.
)

// setupAuditConfigTestDB creates an in-memory SQLite database for testing.
func setupAuditConfigTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()

	// Open SQLite database with unique name per test.
	sqlDB, err := sql.Open("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)

	// Configure SQLite for concurrent access.
	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.SQLiteMaxOpenConnections)
	sqlDB.SetConnMaxLifetime(0)

	// Wrap with GORM.
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Auto-migrate the AuditConfig table.
	err = db.AutoMigrate(&cryptoutilJoseDomain.AuditConfig{})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

// TestNewAuditConfigService tests the service constructor.
func TestNewAuditConfigService(t *testing.T) {
	t.Parallel()

	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	svc := NewAuditConfigService(repo)

	require.NotNil(t, svc)
	require.NotNil(t, svc.repo)
}

// TestAuditConfigService_GetConfig tests getting audit configuration.
func TestAuditConfigService_GetConfig(t *testing.T) {
	t.Parallel()

	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	svc := NewAuditConfigService(repo)
	ctx := context.Background()
	tenantID := googleUuid.New()

	tests := []struct {
		name        string
		operation   string
		setupConfig func()
		wantEnabled bool
		wantRate    float64
		wantErr     bool
	}{
		{
			name:        "returns default when no config exists",
			operation:   AuditOperationEncrypt,
			setupConfig: func() {},
			wantEnabled: DefaultAuditEnabled,
			wantRate:    DefaultAuditSamplingRate,
			wantErr:     false,
		},
		{
			name:      "returns existing config",
			operation: AuditOperationSign,
			setupConfig: func() {
				config := &cryptoutilJoseDomain.AuditConfig{
					TenantID:     tenantID,
					Operation:    AuditOperationSign,
					Enabled:      false,
					SamplingRate: 0.5,
				}
				err := repo.Upsert(ctx, config)
				require.NoError(t, err)
			},
			wantEnabled: false,
			wantRate:    0.5,
			wantErr:     false,
		},
		{
			name:        "invalid operation returns error",
			operation:   "invalid_operation",
			setupConfig: func() {},
			wantEnabled: false,
			wantRate:    0,
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupConfig()

			config, err := svc.GetConfig(ctx, tenantID, tc.operation)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				require.Equal(t, tc.wantEnabled, config.Enabled)
				require.Equal(t, tc.wantRate, config.SamplingRate)
			}
		})
	}
}

// TestAuditConfigService_SetConfig tests setting audit configuration.
func TestAuditConfigService_SetConfig(t *testing.T) {
	t.Parallel()

	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	svc := NewAuditConfigService(repo)
	ctx := context.Background()
	tenantID := googleUuid.New()

	tests := []struct {
		name         string
		operation    string
		enabled      bool
		samplingRate float64
		wantErr      bool
	}{
		{
			name:         "creates new config",
			operation:    AuditOperationEncrypt,
			enabled:      true,
			samplingRate: 0.05,
			wantErr:      false,
		},
		{
			name:         "updates existing config",
			operation:    AuditOperationEncrypt,
			enabled:      false,
			samplingRate: 0.1,
			wantErr:      false,
		},
		{
			name:         "invalid operation returns error",
			operation:    "bad_operation",
			enabled:      true,
			samplingRate: 0.01,
			wantErr:      true,
		},
		{
			name:         "negative sampling rate returns error",
			operation:    AuditOperationDecrypt,
			enabled:      true,
			samplingRate: -0.1,
			wantErr:      true,
		},
		{
			name:         "sampling rate > 1 returns error",
			operation:    AuditOperationDecrypt,
			enabled:      true,
			samplingRate: 1.5,
			wantErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.SetConfig(ctx, tenantID, tc.operation, tc.enabled, tc.samplingRate)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify the config was saved.
				config, err := svc.GetConfig(ctx, tenantID, tc.operation)
				require.NoError(t, err)
				require.Equal(t, tc.enabled, config.Enabled)
				require.Equal(t, tc.samplingRate, config.SamplingRate)
			}
		})
	}
}

// TestAuditConfigService_GetAllConfigs tests getting all configurations.
func TestAuditConfigService_GetAllConfigs(t *testing.T) {
	t.Parallel()

	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	svc := NewAuditConfigService(repo)
	ctx := context.Background()
	tenantID := googleUuid.New()

	// Set a few configs, leave others as defaults.
	err := svc.SetConfig(ctx, tenantID, AuditOperationEncrypt, false, 0.5)
	require.NoError(t, err)

	err = svc.SetConfig(ctx, tenantID, AuditOperationSign, true, 0.25)
	require.NoError(t, err)

	// Get all configs.
	configs, err := svc.GetAllConfigs(ctx, tenantID)
	require.NoError(t, err)
	require.Len(t, configs, len(AllAuditOperations))

	// Check specific configs.
	configMap := make(map[string]cryptoutilJoseDomain.AuditConfig)

	for _, c := range configs {
		configMap[c.Operation] = c
	}

	// Verify custom configs.
	encryptConfig := configMap[AuditOperationEncrypt]
	require.False(t, encryptConfig.Enabled)
	require.Equal(t, 0.5, encryptConfig.SamplingRate)

	signConfig := configMap[AuditOperationSign]
	require.True(t, signConfig.Enabled)
	require.Equal(t, 0.25, signConfig.SamplingRate)

	// Verify default configs for operations not explicitly set.
	decryptConfig := configMap[AuditOperationDecrypt]
	require.True(t, decryptConfig.Enabled) // Default enabled.
	require.Equal(t, DefaultAuditSamplingRate, decryptConfig.SamplingRate)
}

// TestAuditConfigService_InitializeDefaults tests initializing default configs.
func TestAuditConfigService_InitializeDefaults(t *testing.T) {
	t.Parallel()

	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	svc := NewAuditConfigService(repo)
	ctx := context.Background()
	tenantID := googleUuid.New()

	// Initialize defaults.
	err := svc.InitializeDefaults(ctx, tenantID)
	require.NoError(t, err)

	// Verify all operations have configs.
	for _, op := range AllAuditOperations {
		config, err := svc.GetConfig(ctx, tenantID, op)
		require.NoError(t, err)
		require.NotNil(t, config)
		require.Equal(t, tenantID, config.TenantID)
		require.Equal(t, op, config.Operation)
		require.True(t, config.Enabled)
		require.Equal(t, DefaultAuditSamplingRate, config.SamplingRate)
	}
}

// TestAuditConfigService_IsEnabled tests checking if audit is enabled.
func TestAuditConfigService_IsEnabled(t *testing.T) {
	t.Parallel()

	db := setupAuditConfigTestDB(t)
	repo := cryptoutilJoseRepository.NewAuditConfigGormRepository(db)
	svc := NewAuditConfigService(repo)
	ctx := context.Background()
	tenantID := googleUuid.New()

	tests := []struct {
		name        string
		operation   string
		setupConfig func()
		wantEnabled bool
		wantRate    float64
		wantErr     bool
	}{
		{
			name:        "returns defaults when no config exists",
			operation:   AuditOperationVerify,
			setupConfig: func() {},
			wantEnabled: DefaultAuditEnabled,
			wantRate:    DefaultAuditSamplingRate,
			wantErr:     false,
		},
		{
			name:      "returns custom config",
			operation: AuditOperationRotate,
			setupConfig: func() {
				err := svc.SetConfig(ctx, tenantID, AuditOperationRotate, false, 0.75)
				require.NoError(t, err)
			},
			wantEnabled: false,
			wantRate:    0.75,
			wantErr:     false,
		},
		{
			name:        "invalid operation returns error",
			operation:   "invalid",
			setupConfig: func() {},
			wantEnabled: false,
			wantRate:    0,
			wantErr:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupConfig()

			enabled, rate, err := svc.IsEnabled(ctx, tenantID, tc.operation)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantEnabled, enabled)
				require.Equal(t, tc.wantRate, rate)
			}
		})
	}
}

// TestIsValidOperation tests the operation validation helper.
func TestIsValidOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		operation string
		want      bool
	}{
		{AuditOperationEncrypt, true},
		{AuditOperationDecrypt, true},
		{AuditOperationSign, true},
		{AuditOperationVerify, true},
		{AuditOperationKeyGen, true},
		{AuditOperationRotate, true},
		{AuditOperationGetJWKS, true},
		{AuditOperationGetKey, true},
		{AuditOperationListKeys, true},
		{"invalid", false},
		{"", false},
		{"ENCRYPT", false}, // Case-sensitive.
	}

	for _, tc := range tests {
		t.Run(tc.operation, func(t *testing.T) {
			t.Parallel()

			result := isValidOperation(tc.operation)
			require.Equal(t, tc.want, result)
		})
	}
}

// TestIsNotFoundAuditConfigError tests the isNotFoundAuditConfigError helper function.
func TestIsNotFoundAuditConfigError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		want   bool
	}{
		{
			name: "nil error returns false",
			err:  nil,
			want: false,
		},
		{
			name: "error containing 'not found' returns true",
			err:  &testAuditError{msg: "audit config not found"},
			want: true,
		},
		{
			name: "unrelated error returns false",
			err:  &testAuditError{msg: "database connection failed"},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isNotFoundAuditConfigError(tc.err)
			require.Equal(t, tc.want, result)
		})
	}
}

// testAuditError is a simple error type for testing.
type testAuditError struct {
	msg string
}

func (e *testAuditError) Error() string {
	return e.msg
}
