// Copyright (c) 2025 Justin Cranford
//
//

package orm_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityORM "cryptoutil/internal/apps/identity/repository/orm"
)

// setupDeviceAuthTestDB creates an in-memory SQLite database for device authorization tests.
func setupDeviceAuthTestDB(t *testing.T) *cryptoutilIdentityORM.DeviceAuthorizationRepository {
	t.Helper()

	sqlDB, err := sql.Open("sqlite", testDSNInMemory)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	dialector := sqlite.Dialector{Conn: sqlDB}

	db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	err = db.AutoMigrate(&cryptoutilIdentityDomain.DeviceAuthorization{})
	require.NoError(t, err)

	underlyingDB, err := db.DB()
	require.NoError(t, err)

	underlyingDB.SetMaxOpenConns(5)
	underlyingDB.SetMaxIdleConns(5)
	underlyingDB.SetConnMaxLifetime(0)

	return cryptoutilIdentityORM.NewDeviceAuthorizationRepository(db)
}

func TestDeviceAuthorizationRepository_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		auth          *cryptoutilIdentityDomain.DeviceAuthorization
		expectedError string
	}{
		{
			name: "SuccessfulCreate",
			auth: &cryptoutilIdentityDomain.DeviceAuthorization{
				ID:         googleUuid.New(),
				ClientID:   "test-client-123",
				DeviceCode: "device-" + googleUuid.NewString(),
				UserCode:   "ABCD-1234",
				Scope:      "openid profile",
				Status:     cryptoutilIdentityDomain.DeviceAuthStatusPending,
				CreatedAt:  time.Now().UTC(),
				ExpiresAt:  time.Now().UTC().Add(10 * time.Minute).UTC(),
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := setupDeviceAuthTestDB(t)
			err := repo.Create(context.Background(), tt.auth)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDeviceAuthorizationRepository_GetByDeviceCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		deviceCode     string
		setupAuth      *cryptoutilIdentityDomain.DeviceAuthorization
		expectedError  error
		validateResult func(t *testing.T, result *cryptoutilIdentityDomain.DeviceAuthorization)
	}{
		{
			name:       "Found",
			deviceCode: "device-test-code-123",
			setupAuth: &cryptoutilIdentityDomain.DeviceAuthorization{
				ID:         googleUuid.New(),
				ClientID:   "test-client",
				DeviceCode: "device-test-code-123",
				UserCode:   "WXYZ-9876",
				Status:     cryptoutilIdentityDomain.DeviceAuthStatusPending,
				CreatedAt:  time.Now().UTC(),
				ExpiresAt:  time.Now().UTC().Add(5 * time.Minute).UTC(),
			},
			expectedError: nil,
			validateResult: func(t *testing.T, result *cryptoutilIdentityDomain.DeviceAuthorization) {
				t.Helper()
				require.NotNil(t, result)
				require.Equal(t, "device-test-code-123", result.DeviceCode)
				require.Equal(t, "test-client", result.ClientID)
			},
		},
		{
			name:           "NotFound",
			deviceCode:     "nonexistent-device-code",
			setupAuth:      nil,
			expectedError:  cryptoutilIdentityAppErr.ErrDeviceAuthorizationNotFound,
			validateResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := setupDeviceAuthTestDB(t)

			if tt.setupAuth != nil {
				err := repo.Create(context.Background(), tt.setupAuth)
				require.NoError(t, err)
			}

			result, err := repo.GetByDeviceCode(context.Background(), tt.deviceCode)

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)

				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestDeviceAuthorizationRepository_GetByUserCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		userCode       string
		setupAuth      *cryptoutilIdentityDomain.DeviceAuthorization
		expectedError  error
		validateResult func(t *testing.T, result *cryptoutilIdentityDomain.DeviceAuthorization)
	}{
		{
			name:     "Found",
			userCode: "ABCD-5678",
			setupAuth: &cryptoutilIdentityDomain.DeviceAuthorization{
				ID:         googleUuid.New(),
				ClientID:   "test-client-user-code",
				DeviceCode: "device-" + googleUuid.NewString(),
				UserCode:   "ABCD-5678",
				Status:     cryptoutilIdentityDomain.DeviceAuthStatusPending,
				CreatedAt:  time.Now().UTC(),
				ExpiresAt:  time.Now().UTC().Add(15 * time.Minute).UTC(),
			},
			expectedError: nil,
			validateResult: func(t *testing.T, result *cryptoutilIdentityDomain.DeviceAuthorization) {
				t.Helper()
				require.NotNil(t, result)
				require.Equal(t, "ABCD-5678", result.UserCode)
				require.Equal(t, "test-client-user-code", result.ClientID)
			},
		},
		{
			name:           "NotFound",
			userCode:       "INVALID-CODE",
			setupAuth:      nil,
			expectedError:  cryptoutilIdentityAppErr.ErrDeviceAuthorizationNotFound,
			validateResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := setupDeviceAuthTestDB(t)

			if tt.setupAuth != nil {
				err := repo.Create(context.Background(), tt.setupAuth)
				require.NoError(t, err)
			}

			result, err := repo.GetByUserCode(context.Background(), tt.userCode)

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)

				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestDeviceAuthorizationRepository_GetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		id             googleUuid.UUID
		setupAuth      *cryptoutilIdentityDomain.DeviceAuthorization
		expectedError  error
		validateResult func(t *testing.T, result *cryptoutilIdentityDomain.DeviceAuthorization)
	}{
		{
			name: "Found",
			id:   googleUuid.New(),
			setupAuth: &cryptoutilIdentityDomain.DeviceAuthorization{
				ID:         googleUuid.UUID{}, // Overwritten below.
				ClientID:   "test-client-id",
				DeviceCode: "device-" + googleUuid.NewString(),
				UserCode:   "QRST-4321",
				Status:     cryptoutilIdentityDomain.DeviceAuthStatusAuthorized,
				CreatedAt:  time.Now().UTC(),
				ExpiresAt:  time.Now().UTC().Add(20 * time.Minute).UTC(),
			},
			expectedError: nil,
			validateResult: func(t *testing.T, result *cryptoutilIdentityDomain.DeviceAuthorization) {
				t.Helper()
				require.NotNil(t, result)
				require.Equal(t, cryptoutilIdentityDomain.DeviceAuthStatusAuthorized, result.Status)
			},
		},
		{
			name:           "NotFound",
			id:             googleUuid.New(),
			setupAuth:      nil,
			expectedError:  cryptoutilIdentityAppErr.ErrDeviceAuthorizationNotFound,
			validateResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := setupDeviceAuthTestDB(t)

			if tt.setupAuth != nil {
				tt.setupAuth.ID = tt.id

				err := repo.Create(context.Background(), tt.setupAuth)
				require.NoError(t, err)
			}

			result, err := repo.GetByID(context.Background(), tt.id)

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)

				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestDeviceAuthorizationRepository_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupAuth      *cryptoutilIdentityDomain.DeviceAuthorization
		updateAuth     *cryptoutilIdentityDomain.DeviceAuthorization
		expectedError  error
		validateResult func(t *testing.T, repo *cryptoutilIdentityORM.DeviceAuthorizationRepository, id googleUuid.UUID)
	}{
		{
			name: "MarkAsAuthorized",
			setupAuth: &cryptoutilIdentityDomain.DeviceAuthorization{
				ID:         googleUuid.New(),
				ClientID:   "update-test-client",
				DeviceCode: "device-" + googleUuid.NewString(),
				UserCode:   "UPDATE-1234",
				Status:     cryptoutilIdentityDomain.DeviceAuthStatusPending,
				CreatedAt:  time.Now().UTC(),
				ExpiresAt:  time.Now().UTC().Add(10 * time.Minute).UTC(),
			},
			updateAuth: &cryptoutilIdentityDomain.DeviceAuthorization{
				Status: cryptoutilIdentityDomain.DeviceAuthStatusAuthorized,
				UserID: cryptoutilIdentityDomain.NullableUUID{
					UUID:  googleUuid.New(),
					Valid: true,
				},
			},
			expectedError: nil,
			validateResult: func(t *testing.T, repo *cryptoutilIdentityORM.DeviceAuthorizationRepository, id googleUuid.UUID) {
				t.Helper()

				updated, err := repo.GetByID(context.Background(), id)
				require.NoError(t, err)
				require.Equal(t, cryptoutilIdentityDomain.DeviceAuthStatusAuthorized, updated.Status)
				require.True(t, updated.UserID.Valid)
			},
		},
		{
			name:      "UpdateNonExistent",
			setupAuth: nil,
			updateAuth: &cryptoutilIdentityDomain.DeviceAuthorization{
				ID:     googleUuid.New(),
				Status: cryptoutilIdentityDomain.DeviceAuthStatusDenied,
			},
			expectedError:  cryptoutilIdentityAppErr.ErrDeviceAuthorizationNotFound,
			validateResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := setupDeviceAuthTestDB(t)

			var id googleUuid.UUID

			if tt.setupAuth != nil {
				err := repo.Create(context.Background(), tt.setupAuth)
				require.NoError(t, err)

				id = tt.setupAuth.ID
				tt.updateAuth.ID = id
			}

			err := repo.Update(context.Background(), tt.updateAuth)

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)

				if tt.validateResult != nil {
					tt.validateResult(t, repo, id)
				}
			}
		})
	}
}

func TestDeviceAuthorizationRepository_DeleteExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupAuths     []*cryptoutilIdentityDomain.DeviceAuthorization
		expectedError  error
		validateResult func(t *testing.T, repo *cryptoutilIdentityORM.DeviceAuthorizationRepository, ids []googleUuid.UUID)
	}{
		{
			name: "DeleteExpiredDeviceAuths",
			setupAuths: []*cryptoutilIdentityDomain.DeviceAuthorization{
				{
					ID:         googleUuid.New(),
					ClientID:   "expired-1",
					DeviceCode: "device-expired-1",
					UserCode:   "EXP1-0000",
					Status:     cryptoutilIdentityDomain.DeviceAuthStatusPending,
					CreatedAt:  time.Now().UTC().Add(-20 * time.Minute).UTC(),
					ExpiresAt:  time.Now().UTC().Add(-10 * time.Minute).UTC(),
				},
				{
					ID:         googleUuid.New(),
					ClientID:   "valid-1",
					DeviceCode: "device-valid-1",
					UserCode:   "VAL1-9999",
					Status:     cryptoutilIdentityDomain.DeviceAuthStatusPending,
					CreatedAt:  time.Now().UTC(),
					ExpiresAt:  time.Now().UTC().Add(10 * time.Minute).UTC(),
				},
			},
			expectedError: nil,
			validateResult: func(t *testing.T, repo *cryptoutilIdentityORM.DeviceAuthorizationRepository, _ []googleUuid.UUID) {
				t.Helper()

				expired, err := repo.GetByDeviceCode(context.Background(), "device-expired-1")
				require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrDeviceAuthorizationNotFound)
				require.Nil(t, expired)

				valid, err := repo.GetByDeviceCode(context.Background(), "device-valid-1")
				require.NoError(t, err)
				require.Equal(t, "valid-1", valid.ClientID)
			},
		},
		{
			name: "NoExpiredDeviceAuths",
			setupAuths: []*cryptoutilIdentityDomain.DeviceAuthorization{
				{
					ID:         googleUuid.New(),
					ClientID:   "valid-2",
					DeviceCode: "device-valid-2",
					UserCode:   "VAL2-5555",
					Status:     cryptoutilIdentityDomain.DeviceAuthStatusPending,
					CreatedAt:  time.Now().UTC(),
					ExpiresAt:  time.Now().UTC().Add(15 * time.Minute).UTC(),
				},
			},
			expectedError: nil,
			validateResult: func(t *testing.T, repo *cryptoutilIdentityORM.DeviceAuthorizationRepository, _ []googleUuid.UUID) {
				t.Helper()

				valid, err := repo.GetByDeviceCode(context.Background(), "device-valid-2")
				require.NoError(t, err)
				require.Equal(t, "valid-2", valid.ClientID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := setupDeviceAuthTestDB(t)
			ids := make([]googleUuid.UUID, len(tt.setupAuths))

			for i, auth := range tt.setupAuths {
				err := repo.Create(context.Background(), auth)
				require.NoError(t, err)

				ids[i] = auth.ID
			}

			err := repo.DeleteExpired(context.Background())

			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)

				if tt.validateResult != nil {
					tt.validateResult(t, repo, ids)
				}
			}
		})
	}
}
