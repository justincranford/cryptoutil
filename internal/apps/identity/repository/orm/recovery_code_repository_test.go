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
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilIdentityORM "cryptoutil/internal/apps/identity/repository/orm"
)

// setupRecoveryCodeTestDB creates an in-memory SQLite database for recovery code tests.
func setupRecoveryCodeTestDB(t *testing.T) cryptoutilIdentityRepository.RecoveryCodeRepository {
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

	err = db.AutoMigrate(&cryptoutilIdentityDomain.RecoveryCode{})
	require.NoError(t, err)

	underlyingDB, err := db.DB()
	require.NoError(t, err)

	underlyingDB.SetMaxOpenConns(5)
	underlyingDB.SetMaxIdleConns(5)
	underlyingDB.SetConnMaxLifetime(0)

	return cryptoutilIdentityORM.NewRecoveryCodeRepository(db)
}

func TestRecoveryCodeRepository_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		code          *cryptoutilIdentityDomain.RecoveryCode
		expectedError string
	}{
		{
			name: "SuccessfulCreate",
			code: &cryptoutilIdentityDomain.RecoveryCode{
				ID:        googleUuid.New(),
				UserID:    googleUuid.New(),
				CodeHash:  "hash-" + googleUuid.NewString(),
				Used:      false,
				CreatedAt: time.Now().UTC(),
				ExpiresAt: time.Now().UTC().Add(24 * time.Hour).UTC(),
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := setupRecoveryCodeTestDB(t)
			err := repo.Create(context.Background(), tt.code)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRecoveryCodeRepository_CreateBatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		codes         []*cryptoutilIdentityDomain.RecoveryCode
		expectedError string
	}{
		{
			name: "SuccessfulBatchCreate",
			codes: []*cryptoutilIdentityDomain.RecoveryCode{
				{
					ID:        googleUuid.New(),
					UserID:    googleUuid.New(),
					CodeHash:  "hash-1",
					Used:      false,
					CreatedAt: time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(24 * time.Hour).UTC(),
				},
				{
					ID:        googleUuid.New(),
					UserID:    googleUuid.New(),
					CodeHash:  "hash-2",
					Used:      false,
					CreatedAt: time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(24 * time.Hour).UTC(),
				},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := setupRecoveryCodeTestDB(t)
			err := repo.CreateBatch(context.Background(), tt.codes)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRecoveryCodeRepository_GetByUserID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		userID         googleUuid.UUID
		setupCodes     []*cryptoutilIdentityDomain.RecoveryCode
		expectedError  string
		validateResult func(t *testing.T, result []*cryptoutilIdentityDomain.RecoveryCode)
	}{
		{
			name:   "Found_MultipleCodes",
			userID: googleUuid.New(),
			setupCodes: []*cryptoutilIdentityDomain.RecoveryCode{
				{
					ID:        googleUuid.New(),
					UserID:    googleUuid.UUID{}, // Overwritten below.
					CodeHash:  "hash-code-1",
					Used:      false,
					CreatedAt: time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(24 * time.Hour).UTC(),
				},
				{
					ID:        googleUuid.New(),
					UserID:    googleUuid.UUID{}, // Overwritten below.
					CodeHash:  "hash-code-2",
					Used:      true,
					CreatedAt: time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(24 * time.Hour).UTC(),
				},
			},
			expectedError: "",
			validateResult: func(t *testing.T, result []*cryptoutilIdentityDomain.RecoveryCode) {
				t.Helper()
				require.Len(t, result, 2)
				require.Equal(t, "hash-code-1", result[0].CodeHash)
				require.Equal(t, "hash-code-2", result[1].CodeHash)
			},
		},
		{
			name:          "NotFound_NoCodes",
			userID:        googleUuid.New(),
			setupCodes:    nil,
			expectedError: "",
			validateResult: func(t *testing.T, result []*cryptoutilIdentityDomain.RecoveryCode) {
				t.Helper()
				require.Empty(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := setupRecoveryCodeTestDB(t)

			if tt.setupCodes != nil {
				for _, code := range tt.setupCodes {
					code.UserID = tt.userID

					err := repo.Create(context.Background(), code)
					require.NoError(t, err)
				}
			}

			result, err := repo.GetByUserID(context.Background(), tt.userID)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)

				if tt.validateResult != nil {
					tt.validateResult(t, result)
				}
			}
		})
	}
}

func TestRecoveryCodeRepository_GetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		id             googleUuid.UUID
		setupCode      *cryptoutilIdentityDomain.RecoveryCode
		expectedError  error
		validateResult func(t *testing.T, result *cryptoutilIdentityDomain.RecoveryCode)
	}{
		{
			name: "Found",
			id:   googleUuid.New(),
			setupCode: &cryptoutilIdentityDomain.RecoveryCode{
				ID:        googleUuid.UUID{}, // Overwritten below.
				UserID:    googleUuid.New(),
				CodeHash:  "hash-found-test",
				Used:      false,
				CreatedAt: time.Now().UTC(),
				ExpiresAt: time.Now().UTC().Add(24 * time.Hour).UTC(),
			},
			expectedError: nil,
			validateResult: func(t *testing.T, result *cryptoutilIdentityDomain.RecoveryCode) {
				t.Helper()
				require.NotNil(t, result)
				require.Equal(t, "hash-found-test", result.CodeHash)
			},
		},
		{
			name:           "NotFound",
			id:             googleUuid.New(),
			setupCode:      nil,
			expectedError:  cryptoutilIdentityAppErr.ErrRecoveryCodeNotFound,
			validateResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := setupRecoveryCodeTestDB(t)

			if tt.setupCode != nil {
				tt.setupCode.ID = tt.id

				err := repo.Create(context.Background(), tt.setupCode)
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

func TestRecoveryCodeRepository_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupCode      *cryptoutilIdentityDomain.RecoveryCode
		updateCode     *cryptoutilIdentityDomain.RecoveryCode
		expectedError  string
		validateResult func(t *testing.T, repo cryptoutilIdentityRepository.RecoveryCodeRepository, id googleUuid.UUID)
	}{
		{
			name: "MarkAsUsed",
			setupCode: &cryptoutilIdentityDomain.RecoveryCode{
				ID:        googleUuid.New(),
				UserID:    googleUuid.New(),
				CodeHash:  "hash-to-mark-used",
				Used:      false,
				CreatedAt: time.Now().UTC(),
				ExpiresAt: time.Now().UTC().Add(24 * time.Hour).UTC(),
			},
			updateCode:    nil, // Populated below.
			expectedError: "",
			validateResult: func(t *testing.T, repo cryptoutilIdentityRepository.RecoveryCodeRepository, id googleUuid.UUID) {
				t.Helper()

				updated, err := repo.GetByID(context.Background(), id)
				require.NoError(t, err)
				require.True(t, updated.Used)
				require.NotNil(t, updated.UsedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := setupRecoveryCodeTestDB(t)

			if tt.setupCode != nil {
				err := repo.Create(context.Background(), tt.setupCode)
				require.NoError(t, err)

				tt.setupCode.MarkAsUsed()

				tt.updateCode = tt.setupCode
			}

			err := repo.Update(context.Background(), tt.updateCode)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)

				if tt.validateResult != nil {
					tt.validateResult(t, repo, tt.setupCode.ID)
				}
			}
		})
	}
}
