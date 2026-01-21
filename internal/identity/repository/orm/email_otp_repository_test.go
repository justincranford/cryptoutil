// Copyright (c) 2025 Justin Cranford

package orm_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityORM "cryptoutil/internal/identity/repository/orm"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

// TestEmailOTPRepository_Create tests all Create paths.
func TestEmailOTPRepository_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupOTP  func() *cryptoutilIdentityDomain.EmailOTP
		expectErr bool
	}{
		{
			name: "SuccessfulCreate",
			setupOTP: func() *cryptoutilIdentityDomain.EmailOTP {
				return &cryptoutilIdentityDomain.EmailOTP{
					ID:        googleUuid.Must(googleUuid.NewV7()),
					UserID:    googleUuid.Must(googleUuid.NewV7()),
					CodeHash:  "hash123",
					ExpiresAt: time.Now().Add(5 * time.Minute),
					CreatedAt: time.Now(),
				}
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := setupTestDB(t)
			repo := cryptoutilIdentityORM.NewEmailOTPRepository(db)

			otp := tt.setupOTP()
			err := repo.Create(ctx, otp)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestEmailOTPRepository_GetByUserID tests all GetByUserID paths.
func TestEmailOTPRepository_GetByUserID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupOTPs  func() (googleUuid.UUID, []*cryptoutilIdentityDomain.EmailOTP)
		expectErr  error
		expectFunc func(*testing.T, *cryptoutilIdentityDomain.EmailOTP)
	}{
		{
			name: "Found_MostRecent",
			setupOTPs: func() (googleUuid.UUID, []*cryptoutilIdentityDomain.EmailOTP) {
				userID := googleUuid.Must(googleUuid.NewV7())
				older := &cryptoutilIdentityDomain.EmailOTP{
					ID:        googleUuid.Must(googleUuid.NewV7()),
					UserID:    userID,
					CodeHash:  "older",
					ExpiresAt: time.Now().Add(5 * time.Minute),
					CreatedAt: time.Now().Add(-10 * time.Minute),
				}
				newer := &cryptoutilIdentityDomain.EmailOTP{
					ID:        googleUuid.Must(googleUuid.NewV7()),
					UserID:    userID,
					CodeHash:  "newer",
					ExpiresAt: time.Now().Add(5 * time.Minute),
					CreatedAt: time.Now(),
				}

				return userID, []*cryptoutilIdentityDomain.EmailOTP{older, newer}
			},
			expectErr: nil,
			expectFunc: func(t *testing.T, otp *cryptoutilIdentityDomain.EmailOTP) {
				require.Equal(t, "newer", otp.CodeHash)
			},
		},
		{
			name: "NotFound",
			setupOTPs: func() (googleUuid.UUID, []*cryptoutilIdentityDomain.EmailOTP) {
				return googleUuid.Must(googleUuid.NewV7()), nil
			},
			expectErr: cryptoutilIdentityAppErr.ErrEmailOTPNotFound,
			expectFunc: func(t *testing.T, otp *cryptoutilIdentityDomain.EmailOTP) {
				require.Nil(t, otp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := setupTestDB(t)
			repo := cryptoutilIdentityORM.NewEmailOTPRepository(db)

			userID, otps := tt.setupOTPs()
			for _, otp := range otps {
				require.NoError(t, repo.Create(ctx, otp))
			}

			result, err := repo.GetByUserID(ctx, userID)

			if tt.expectErr != nil {
				require.ErrorIs(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
			}

			tt.expectFunc(t, result)
		})
	}
}

// TestEmailOTPRepository_GetByID tests all GetByID paths.
func TestEmailOTPRepository_GetByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupOTP  func() *cryptoutilIdentityDomain.EmailOTP
		queryID   func(*cryptoutilIdentityDomain.EmailOTP) googleUuid.UUID
		expectErr error
	}{
		{
			name: "Found",
			setupOTP: func() *cryptoutilIdentityDomain.EmailOTP {
				return &cryptoutilIdentityDomain.EmailOTP{
					ID:        googleUuid.Must(googleUuid.NewV7()),
					UserID:    googleUuid.Must(googleUuid.NewV7()),
					CodeHash:  "hash123",
					ExpiresAt: time.Now().Add(5 * time.Minute),
					CreatedAt: time.Now(),
				}
			},
			queryID: func(otp *cryptoutilIdentityDomain.EmailOTP) googleUuid.UUID {
				return otp.ID
			},
			expectErr: nil,
		},
		{
			name: "NotFound",
			setupOTP: func() *cryptoutilIdentityDomain.EmailOTP {
				return &cryptoutilIdentityDomain.EmailOTP{
					ID:        googleUuid.Must(googleUuid.NewV7()),
					UserID:    googleUuid.Must(googleUuid.NewV7()),
					CodeHash:  "hash123",
					ExpiresAt: time.Now().Add(5 * time.Minute),
					CreatedAt: time.Now(),
				}
			},
			queryID:   func(_ *cryptoutilIdentityDomain.EmailOTP) googleUuid.UUID { return googleUuid.Must(googleUuid.NewV7()) },
			expectErr: cryptoutilIdentityAppErr.ErrEmailOTPNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := setupTestDB(t)
			repo := cryptoutilIdentityORM.NewEmailOTPRepository(db)

			otp := tt.setupOTP()
			require.NoError(t, repo.Create(ctx, otp))

			result, err := repo.GetByID(ctx, tt.queryID(otp))

			if tt.expectErr != nil {
				require.ErrorIs(t, err, tt.expectErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, otp.ID, result.ID)
			}
		})
	}
}

// TestEmailOTPRepository_Update tests Update paths.
func TestEmailOTPRepository_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupOTP  func() *cryptoutilIdentityDomain.EmailOTP
		modifyOTP func(*cryptoutilIdentityDomain.EmailOTP)
	}{
		{
			name: "UpdateHashSuccess",
			setupOTP: func() *cryptoutilIdentityDomain.EmailOTP {
				return &cryptoutilIdentityDomain.EmailOTP{
					ID:        googleUuid.Must(googleUuid.NewV7()),
					UserID:    googleUuid.Must(googleUuid.NewV7()),
					CodeHash:  "oldhash",
					ExpiresAt: time.Now().Add(5 * time.Minute),
					CreatedAt: time.Now(),
				}
			},
			modifyOTP: func(otp *cryptoutilIdentityDomain.EmailOTP) {
				otp.CodeHash = "newhash"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := setupTestDB(t)
			repo := cryptoutilIdentityORM.NewEmailOTPRepository(db)

			otp := tt.setupOTP()
			require.NoError(t, repo.Create(ctx, otp))

			tt.modifyOTP(otp)
			require.NoError(t, repo.Update(ctx, otp))

			retrieved, err := repo.GetByID(ctx, otp.ID)
			require.NoError(t, err)
			require.Equal(t, otp.CodeHash, retrieved.CodeHash)
		})
	}
}

// TestEmailOTPRepository_DeleteByUserID tests DeleteByUserID paths.
func TestEmailOTPRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupOTPs func() (googleUuid.UUID, []*cryptoutilIdentityDomain.EmailOTP)
		expectErr bool
	}{
		{
			name: "DeleteMultiple",
			setupOTPs: func() (googleUuid.UUID, []*cryptoutilIdentityDomain.EmailOTP) {
				userID := googleUuid.Must(googleUuid.NewV7())
				otp1 := &cryptoutilIdentityDomain.EmailOTP{
					ID:        googleUuid.Must(googleUuid.NewV7()),
					UserID:    userID,
					CodeHash:  "hash1",
					ExpiresAt: time.Now().Add(5 * time.Minute),
					CreatedAt: time.Now(),
				}
				otp2 := &cryptoutilIdentityDomain.EmailOTP{
					ID:        googleUuid.Must(googleUuid.NewV7()),
					UserID:    userID,
					CodeHash:  "hash2",
					ExpiresAt: time.Now().Add(5 * time.Minute),
					CreatedAt: time.Now(),
				}

				return userID, []*cryptoutilIdentityDomain.EmailOTP{otp1, otp2}
			},
			expectErr: false,
		},
		{
			name: "NoOTPsToDelete",
			setupOTPs: func() (googleUuid.UUID, []*cryptoutilIdentityDomain.EmailOTP) {
				return googleUuid.Must(googleUuid.NewV7()), nil
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := setupTestDB(t)
			repo := cryptoutilIdentityORM.NewEmailOTPRepository(db)

			userID, otps := tt.setupOTPs()
			for _, otp := range otps {
				require.NoError(t, repo.Create(ctx, otp))
			}

			err := repo.DeleteByUserID(ctx, userID)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				_, err := repo.GetByUserID(ctx, userID)
				require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrEmailOTPNotFound)
			}
		})
	}
}

// TestEmailOTPRepository_DeleteExpired tests DeleteExpired paths.
func TestEmailOTPRepository_DeleteExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupOTPs      func() []*cryptoutilIdentityDomain.EmailOTP
		expectedCount  int64
		remainingCount int
	}{
		{
			name: "DeleteExpiredOTPs",
			setupOTPs: func() []*cryptoutilIdentityDomain.EmailOTP {
				expired1 := &cryptoutilIdentityDomain.EmailOTP{
					ID:        googleUuid.Must(googleUuid.NewV7()),
					UserID:    googleUuid.Must(googleUuid.NewV7()),
					CodeHash:  "expired1",
					ExpiresAt: time.Now().Add(-10 * time.Minute),
					CreatedAt: time.Now().Add(-20 * time.Minute),
				}
				expired2 := &cryptoutilIdentityDomain.EmailOTP{
					ID:        googleUuid.Must(googleUuid.NewV7()),
					UserID:    googleUuid.Must(googleUuid.NewV7()),
					CodeHash:  "expired2",
					ExpiresAt: time.Now().Add(-5 * time.Minute),
					CreatedAt: time.Now().Add(-15 * time.Minute),
				}
				valid := &cryptoutilIdentityDomain.EmailOTP{
					ID:        googleUuid.Must(googleUuid.NewV7()),
					UserID:    googleUuid.Must(googleUuid.NewV7()),
					CodeHash:  "valid",
					ExpiresAt: time.Now().Add(5 * time.Minute),
					CreatedAt: time.Now(),
				}

				return []*cryptoutilIdentityDomain.EmailOTP{expired1, expired2, valid}
			},
			expectedCount:  2,
			remainingCount: 1,
		},
		{
			name: "NoExpiredOTPs",
			setupOTPs: func() []*cryptoutilIdentityDomain.EmailOTP {
				valid := &cryptoutilIdentityDomain.EmailOTP{
					ID:        googleUuid.Must(googleUuid.NewV7()),
					UserID:    googleUuid.Must(googleUuid.NewV7()),
					CodeHash:  "valid",
					ExpiresAt: time.Now().Add(5 * time.Minute),
					CreatedAt: time.Now(),
				}

				return []*cryptoutilIdentityDomain.EmailOTP{valid}
			},
			expectedCount:  0,
			remainingCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			db := setupTestDB(t)
			repo := cryptoutilIdentityORM.NewEmailOTPRepository(db)

			otps := tt.setupOTPs()
			for _, otp := range otps {
				require.NoError(t, repo.Create(ctx, otp))
			}

			count, err := repo.DeleteExpired(ctx)
			require.NoError(t, err)
			require.Equal(t, tt.expectedCount, count)

			// Verify remaining count.
			var remaining []cryptoutilIdentityDomain.EmailOTP

			require.NoError(t, db.Find(&remaining).Error)
			require.Len(t, remaining, tt.remainingCount)
		})
	}
}

// setupTestDB creates an in-memory SQLite database for testing (CGO-free).
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()

	// Use testDSNInMemory for isolated per-test database (no shared cache for parallel test safety).
	dsn := testDSNInMemory
	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(0)

	dialector := sqlite.Dialector{Conn: sqlDB}
	db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	err = db.AutoMigrate(&cryptoutilIdentityDomain.EmailOTP{})
	require.NoError(t, err)

	return db
}
