// Copyright (c) 2025 Justin Cranford
//
//

package rotation

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // Import modernc.org/sqlite driver.

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

func TestRevokeSecretVersion_NotFound(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	service := NewSecretRotationService(db)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7())

	// Attempt to revoke non-existent version.
	err := service.RevokeSecretVersion(
		ctx,
		clientID,
		999,
		"admin-user",
		"test",
	)
	require.Error(t, err, "Revoking non-existent version should fail")
	require.Contains(t, err.Error(), "not found")
}

// TestGetActiveSecretVersions tests retrieving all active secret versions.
func TestGetActiveSecretVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		setupVersions    int
		revokeVersions   []int
		expectedActive   int
		expectedVersions []int
	}{
		{
			name:             "no_secrets",
			setupVersions:    0,
			revokeVersions:   nil,
			expectedActive:   0,
			expectedVersions: []int{},
		},
		{
			name:             "single_active_secret",
			setupVersions:    1,
			revokeVersions:   nil,
			expectedActive:   1,
			expectedVersions: []int{1},
		},
		{
			name:             "multiple_active_secrets",
			setupVersions:    3,
			revokeVersions:   nil,
			expectedActive:   3,
			expectedVersions: []int{3, 2, 1}, // DESC order.
		},
		{
			name:             "mixed_active_and_revoked",
			setupVersions:    4,
			revokeVersions:   []int{2},
			expectedActive:   3,
			expectedVersions: []int{4, 3, 1}, // Excludes version 2.
		},
		{
			name:             "all_revoked",
			setupVersions:    2,
			revokeVersions:   []int{1, 2},
			expectedActive:   0,
			expectedVersions: []int{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			service := NewSecretRotationService(db)
			ctx := context.Background()

			clientID := googleUuid.Must(googleUuid.NewV7())
			gracePeriod := cryptoutilSharedMagic.HoursPerDay * time.Hour

			// Create secret versions.
			for i := 0; i < tc.setupVersions; i++ {
				_, err := service.RotateClientSecret(
					ctx,
					clientID,
					gracePeriod,
					"test-user",
					fmt.Sprintf("rotation-%d", i+1),
				)
				require.NoError(t, err)
			}

			// Revoke specified versions.
			for _, version := range tc.revokeVersions {
				err := service.RevokeSecretVersion(
					ctx,
					clientID,
					version,
					"admin-user",
					"test-revocation",
				)
				require.NoError(t, err)
			}

			// Get active secret versions.
			versions, err := service.GetActiveSecretVersions(ctx, clientID)
			require.NoError(t, err)

			require.Len(t, versions, tc.expectedActive, "Should return correct number of active versions")

			// Verify versions are in DESC order.
			for i, version := range versions {
				require.Equal(t, tc.expectedVersions[i], version.Version, "Version mismatch at index %d", i)
				require.Equal(t, cryptoutilIdentityDomain.SecretStatusActive, version.Status, "All returned versions should be active")
			}
		})
	}
}

// TestGenerateRandomSecret tests the internal secret generation function.
func TestGenerateRandomSecret(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		length     int
		wantErr    bool
		minEncoded int // Minimum base64-encoded length.
	}{
		{
			name:       "standard_length_32",
			length:     cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes,
			wantErr:    false,
			minEncoded: 40, // base64(32 bytes) = ~43 chars.
		},
		{
			name:       "standard_length_64",
			length:     cryptoutilSharedMagic.MinSerialNumberBits,
			wantErr:    false,
			minEncoded: cryptoutilSharedMagic.LineWidth, // base64(64 bytes) = ~86 chars.
		},
		{
			name:       "minimum_length_16",
			length:     cryptoutilSharedMagic.RealmMinTokenLengthBytes,
			wantErr:    false,
			minEncoded: cryptoutilSharedMagic.MaxErrorDisplay, // base64(16 bytes) = ~22 chars.
		},
		{
			name:       "zero_length",
			length:     0,
			wantErr:    false,
			minEncoded: 0, // base64(0 bytes) = empty string (expected behavior).
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			secret, err := generateRandomSecret(tc.length)

			if tc.wantErr {
				require.Error(t, err)
				require.Empty(t, secret)
			} else {
				require.NoError(t, err)

				// Special case: zero-length generates empty string.
				if tc.length == 0 {
					require.Empty(t, secret, "Zero-length secret should be empty string")
				} else {
					require.NotEmpty(t, secret, "Should generate non-empty secret")
					require.GreaterOrEqual(t, len(secret), tc.minEncoded, "Base64-encoded secret should meet minimum length")

					// Verify secret is URL-safe base64.
					require.NotContains(t, secret, "+", "Should use URL-safe encoding")
					require.NotContains(t, secret, "/", "Should use URL-safe encoding")
				}
			}
		})
	}

	// Test uniqueness - generate multiple secrets and verify they're different.
	t.Run("uniqueness", func(t *testing.T) {
		t.Parallel()

		const (
			numSecrets   = 10
			secretLength = 32
		)

		secrets := make(map[string]bool)

		for i := 0; i < numSecrets; i++ {
			secret, err := generateRandomSecret(secretLength)
			require.NoError(t, err)

			// Check for duplicates.
			require.False(t, secrets[secret], "Generated secret should be unique (duplicate found)")

			secrets[secret] = true
		}

		require.Len(t, secrets, numSecrets, "Should generate exactly %d unique secrets", numSecrets)
	})
}

func TestSecretRotationService_GetActiveSecretVersions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := setupTestDB(t)
	service := NewSecretRotationService(db)

	// Create test client.
	clientID, err := googleUuid.NewV7()
	require.NoError(t, err)

	// Create multiple secret versions with different statuses.
	activeVersion1 := &cryptoutilIdentityDomain.ClientSecretVersion{
		ID:        googleUuid.New(),
		ClientID:  clientID,
		Version:   1,
		Status:    cryptoutilIdentityDomain.SecretStatusActive,
		CreatedAt: time.Now().UTC().Add(-cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour),
	}

	activeVersion2 := &cryptoutilIdentityDomain.ClientSecretVersion{
		ID:        googleUuid.New(),
		ClientID:  clientID,
		Version:   2,
		Status:    cryptoutilIdentityDomain.SecretStatusActive,
		CreatedAt: time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour),
	}

	revokedAt := time.Now().UTC().Add(-cryptoutilSharedMagic.DefaultEmailOTPLength * time.Hour)

	revokedVersion := &cryptoutilIdentityDomain.ClientSecretVersion{
		ID:        googleUuid.New(),
		ClientID:  clientID,
		Version:   3,
		Status:    cryptoutilIdentityDomain.SecretStatusRevoked,
		CreatedAt: time.Now().UTC().Add(-cryptoutilSharedMagic.HashPrefixLength * time.Hour),
		RevokedAt: &revokedAt,
	}

	// Insert versions.
	require.NoError(t, db.WithContext(ctx).Create(activeVersion1).Error)
	require.NoError(t, db.WithContext(ctx).Create(activeVersion2).Error)
	require.NoError(t, db.WithContext(ctx).Create(revokedVersion).Error)

	// Get active versions - should only return active ones, ordered by version DESC.
	activeVersions, err := service.GetActiveSecretVersions(ctx, clientID)
	require.NoError(t, err)
	require.Len(t, activeVersions, 2, "Should return only active versions")

	// Verify order (DESC by version).
	require.Equal(t, 2, activeVersions[0].Version, "First version should be 2")
	require.Equal(t, 1, activeVersions[1].Version, "Second version should be 1")
	require.Equal(t, cryptoutilIdentityDomain.SecretStatusActive, activeVersions[0].Status)
	require.Equal(t, cryptoutilIdentityDomain.SecretStatusActive, activeVersions[1].Status)
}
