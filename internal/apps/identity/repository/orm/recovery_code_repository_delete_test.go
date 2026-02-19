// Copyright (c) 2025 Justin Cranford
//
//

package orm_test

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

func TestRecoveryCodeRepository_DeleteByUserID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		userID         googleUuid.UUID
		setupCodes     []*cryptoutilIdentityDomain.RecoveryCode
		expectedError  string
		validateResult func(t *testing.T, repo cryptoutilIdentityRepository.RecoveryCodeRepository, userID googleUuid.UUID)
	}{
		{
			name:   "DeleteMultipleCodes",
			userID: googleUuid.New(),
			setupCodes: []*cryptoutilIdentityDomain.RecoveryCode{
				{
					ID:        googleUuid.New(),
					UserID:    googleUuid.UUID{}, // Overwritten below.
					CodeHash:  "hash-delete-1",
					Used:      false,
					CreatedAt: time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(24 * time.Hour).UTC(),
				},
				{
					ID:        googleUuid.New(),
					UserID:    googleUuid.UUID{}, // Overwritten below.
					CodeHash:  "hash-delete-2",
					Used:      true,
					CreatedAt: time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(24 * time.Hour).UTC(),
				},
			},
			expectedError: "",
			validateResult: func(t *testing.T, repo cryptoutilIdentityRepository.RecoveryCodeRepository, userID googleUuid.UUID) {
				t.Helper()

				codes, err := repo.GetByUserID(context.Background(), userID)
				require.NoError(t, err)
				require.Empty(t, codes)
			},
		},
		{
			name:           "NoCodestoDelete",
			userID:         googleUuid.New(),
			setupCodes:     nil,
			expectedError:  "",
			validateResult: nil,
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

			err := repo.DeleteByUserID(context.Background(), tt.userID)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)

				if tt.validateResult != nil {
					tt.validateResult(t, repo, tt.userID)
				}
			}
		})
	}
}

func TestRecoveryCodeRepository_DeleteExpired(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		setupCodes           []*cryptoutilIdentityDomain.RecoveryCode
		expectedError        string
		expectedRowsAffected int64
		validateResult       func(t *testing.T, repo cryptoutilIdentityRepository.RecoveryCodeRepository)
	}{
		{
			name: "DeleteExpiredCodes",
			setupCodes: []*cryptoutilIdentityDomain.RecoveryCode{
				{
					ID:        googleUuid.New(),
					UserID:    googleUuid.New(),
					CodeHash:  "hash-expired-1",
					Used:      false,
					CreatedAt: time.Now().UTC().Add(-48 * time.Hour).UTC(),
					ExpiresAt: time.Now().UTC().Add(-24 * time.Hour).UTC(),
				},
				{
					ID:        googleUuid.New(),
					UserID:    googleUuid.New(),
					CodeHash:  "hash-valid-1",
					Used:      false,
					CreatedAt: time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(24 * time.Hour).UTC(),
				},
			},
			expectedError:        "",
			expectedRowsAffected: 1,
			validateResult: func(t *testing.T, _ cryptoutilIdentityRepository.RecoveryCodeRepository) {
				t.Helper()

				// Validate expired code deleted by attempting to retrieve it.
			},
		},
		{
			name: "NoExpiredCodes",
			setupCodes: []*cryptoutilIdentityDomain.RecoveryCode{
				{
					ID:        googleUuid.New(),
					UserID:    googleUuid.New(),
					CodeHash:  "hash-valid-2",
					Used:      false,
					CreatedAt: time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(48 * time.Hour).UTC(),
				},
			},
			expectedError:        "",
			expectedRowsAffected: 0,
			validateResult:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := setupRecoveryCodeTestDB(t)

			if tt.setupCodes != nil {
				for _, code := range tt.setupCodes {
					err := repo.Create(context.Background(), code)
					require.NoError(t, err)
				}
			}

			rowsAffected, err := repo.DeleteExpired(context.Background())

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedRowsAffected, rowsAffected)

				if tt.validateResult != nil {
					tt.validateResult(t, repo)
				}
			}
		})
	}
}

func TestRecoveryCodeRepository_CountUnused(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		userID        googleUuid.UUID
		setupCodes    []*cryptoutilIdentityDomain.RecoveryCode
		expectedCount int64
		expectedError string
	}{
		{
			name:   "CountUnusedCodes",
			userID: googleUuid.New(),
			setupCodes: []*cryptoutilIdentityDomain.RecoveryCode{
				{
					ID:        googleUuid.New(),
					UserID:    googleUuid.UUID{}, // Overwritten below.
					CodeHash:  "hash-unused-1",
					Used:      false,
					CreatedAt: time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(24 * time.Hour).UTC(),
				},
				{
					ID:        googleUuid.New(),
					UserID:    googleUuid.UUID{}, // Overwritten below.
					CodeHash:  "hash-unused-2",
					Used:      false,
					CreatedAt: time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(24 * time.Hour).UTC(),
				},
				{
					ID:        googleUuid.New(),
					UserID:    googleUuid.UUID{}, // Overwritten below.
					CodeHash:  "hash-used-1",
					Used:      true,
					CreatedAt: time.Now().UTC(),
					ExpiresAt: time.Now().UTC().Add(24 * time.Hour).UTC(),
				},
				{
					ID:        googleUuid.New(),
					UserID:    googleUuid.UUID{}, // Overwritten below.
					CodeHash:  "hash-expired-1",
					Used:      false,
					CreatedAt: time.Now().UTC().Add(-48 * time.Hour).UTC(),
					ExpiresAt: time.Now().UTC().Add(-24 * time.Hour).UTC(),
				},
			},
			expectedCount: 2,
			expectedError: "",
		},
		{
			name:          "NoUnusedCodes",
			userID:        googleUuid.New(),
			setupCodes:    nil,
			expectedCount: 0,
			expectedError: "",
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

			count, err := repo.CountUnused(context.Background(), tt.userID)

			if tt.expectedError != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expectedCount, count)
			}
		})
	}
}
