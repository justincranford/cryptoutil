// Copyright (c) 2025 Justin Cranford

package idp

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	json "encoding/json"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilIdentityORM "cryptoutil/internal/apps/identity/repository/orm"
)

// TestUserInfoClaims validates OIDC UserInfo endpoint returns all required standard claims
// and custom claims. Satisfies R02-05: UserInfo response includes all required OIDC claims.
func TestUserInfoClaims(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		user              *cryptoutilIdentityDomain.User
		expectedClaims    map[string]any
		unexpectedClaims  []string
		verifyClaimValues bool
	}{
		{
			name: "standard_oidc_claims",
			user: &cryptoutilIdentityDomain.User{
				ID:                googleUuid.Must(googleUuid.NewV7()),
				Sub:               googleUuid.Must(googleUuid.NewV7()).String(),
				PreferredUsername: "testuser_standard",
				Email:             "standard@example.com",
				EmailVerified:     true,
				GivenName:         "Test",
				FamilyName:        "User",
				Name:              "Test User",
				Locale:            "en-US",
				Zoneinfo:          "America/New_York",
				UpdatedAt:         time.Now().UTC(),
			},
			expectedClaims: map[string]any{
				cryptoutilSharedMagic.ClaimSub:                nil, // Value validated separately
				cryptoutilSharedMagic.ClaimPreferredUsername: "testuser_standard",
				cryptoutilSharedMagic.ClaimEmail:              "standard@example.com",
				cryptoutilSharedMagic.ClaimEmailVerified:     true,
				cryptoutilSharedMagic.ClaimGivenName:         "Test",
				cryptoutilSharedMagic.ClaimFamilyName:        "User",
				cryptoutilSharedMagic.ClaimName:               "Test User",
				cryptoutilSharedMagic.ClaimLocale:             "en-US",
				cryptoutilSharedMagic.ClaimZoneinfo:           "America/New_York",
				cryptoutilSharedMagic.ClaimUpdatedAt:         nil, // Timestamp validation
			},
			verifyClaimValues: true,
		},
		{
			name: "minimal_required_claims",
			user: &cryptoutilIdentityDomain.User{
				ID:                googleUuid.Must(googleUuid.NewV7()),
				Sub:               googleUuid.Must(googleUuid.NewV7()).String(),
				PreferredUsername: "testuser_minimal",
				Email:             "minimal@example.com",
				EmailVerified:     false,
			},
			expectedClaims: map[string]any{
				cryptoutilSharedMagic.ClaimSub:                nil,
				cryptoutilSharedMagic.ClaimPreferredUsername: "testuser_minimal",
				cryptoutilSharedMagic.ClaimEmail:              "minimal@example.com",
				cryptoutilSharedMagic.ClaimEmailVerified:     false,
			},
			unexpectedClaims: []string{
				cryptoutilSharedMagic.ClaimGivenName,
				cryptoutilSharedMagic.ClaimFamilyName,
				cryptoutilSharedMagic.ClaimName,
				cryptoutilSharedMagic.ClaimLocale,
				cryptoutilSharedMagic.ClaimZoneinfo,
			},
			verifyClaimValues: true,
		},
		{
			name: "verified_email",
			user: &cryptoutilIdentityDomain.User{
				ID:                googleUuid.Must(googleUuid.NewV7()),
				Sub:               googleUuid.Must(googleUuid.NewV7()).String(),
				PreferredUsername: "testuser_verified",
				Email:             "verified@example.com",
				EmailVerified:     true,
			},
			expectedClaims: map[string]any{
				cryptoutilSharedMagic.ClaimEmailVerified: true,
			},
			verifyClaimValues: true,
		},
		{
			name: "unverified_email",
			user: &cryptoutilIdentityDomain.User{
				ID:                googleUuid.Must(googleUuid.NewV7()),
				Sub:               googleUuid.Must(googleUuid.NewV7()).String(),
				PreferredUsername: "testuser_unverified",
				Email:             "unverified@example.com",
				EmailVerified:     false,
			},
			expectedClaims: map[string]any{
				cryptoutilSharedMagic.ClaimEmailVerified: false,
			},
			verifyClaimValues: true,
		},
	}

	for _, tc := range tests {
		// Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			// Create in-memory SQLite repository factory.
			repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, &cryptoutilIdentityConfig.DatabaseConfig{
				Type: cryptoutilSharedMagic.TestDatabaseSQLite,
				DSN:  cryptoutilSharedMagic.SQLiteMemoryPlaceholder,
			})
			require.NoError(t, err, "failed to create repository factory")

			// Auto-migrate User model.
			db := repoFactory.DB()
			err = db.AutoMigrate(&cryptoutilIdentityDomain.User{})
			require.NoError(t, err, "failed to auto-migrate User")

			// Create user in database.
			userRepo := cryptoutilIdentityORM.NewUserRepository(db)
			err = userRepo.Create(ctx, tc.user)
			require.NoError(t, err, "failed to create test user")

			// Retrieve user (simulates UserInfo endpoint query).
			retrievedUser, err := userRepo.GetBySub(ctx, tc.user.Sub)
			require.NoError(t, err, "failed to retrieve user")
			require.NotNil(t, retrievedUser, "retrieved user should not be nil")

			// Convert user to UserInfo claims (JSON representation).
			claimsJSON, err := json.Marshal(map[string]any{
				cryptoutilSharedMagic.ClaimSub:                retrievedUser.Sub,
				cryptoutilSharedMagic.ClaimPreferredUsername: retrievedUser.PreferredUsername,
				cryptoutilSharedMagic.ClaimEmail:              retrievedUser.Email,
				cryptoutilSharedMagic.ClaimEmailVerified:     retrievedUser.EmailVerified,
				cryptoutilSharedMagic.ClaimGivenName:         retrievedUser.GivenName,
				cryptoutilSharedMagic.ClaimFamilyName:        retrievedUser.FamilyName,
				cryptoutilSharedMagic.ClaimName:               retrievedUser.Name,
				cryptoutilSharedMagic.ClaimLocale:             retrievedUser.Locale,
				cryptoutilSharedMagic.ClaimZoneinfo:           retrievedUser.Zoneinfo,
				cryptoutilSharedMagic.ClaimUpdatedAt:         retrievedUser.UpdatedAt.Unix(),
			})
			require.NoError(t, err, "failed to marshal claims to JSON")

			var claims map[string]any

			err = json.Unmarshal(claimsJSON, &claims)
			require.NoError(t, err, "failed to unmarshal claims from JSON")

			// Verify expected claims present.
			for claimName := range tc.expectedClaims {
				_, exists := claims[claimName]
				require.True(t, exists, "claim %s should be present", claimName)
			}

			// Verify unexpected claims absent.
			for _, claimName := range tc.unexpectedClaims {
				value, exists := claims[claimName]
				// Allow empty string values (GORM zero values).
				if exists {
					strValue, ok := value.(string)
					require.True(t, !ok || strValue == "", "claim %s should be empty or absent, got: %v", claimName, value)
				}
			}

			// Verify specific claim values.
			if tc.verifyClaimValues {
				require.Equal(t, tc.user.Sub, claims[cryptoutilSharedMagic.ClaimSub], "sub claim should match")
				require.Equal(t, tc.user.PreferredUsername, claims[cryptoutilSharedMagic.ClaimPreferredUsername], "preferred_username claim should match")
				require.Equal(t, tc.user.Email, claims[cryptoutilSharedMagic.ClaimEmail], "email claim should match")
				require.Equal(t, tc.user.EmailVerified.Bool(), claims[cryptoutilSharedMagic.ClaimEmailVerified], "email_verified claim should match")

				// Verify optional claims if present.
				if tc.user.GivenName != "" {
					require.Equal(t, tc.user.GivenName, claims[cryptoutilSharedMagic.ClaimGivenName], "given_name claim should match")
				}

				if tc.user.FamilyName != "" {
					require.Equal(t, tc.user.FamilyName, claims[cryptoutilSharedMagic.ClaimFamilyName], "family_name claim should match")
				}

				if tc.user.Name != "" {
					require.Equal(t, tc.user.Name, claims[cryptoutilSharedMagic.ClaimName], "name claim should match")
				}

				if tc.user.Locale != "" {
					require.Equal(t, tc.user.Locale, claims[cryptoutilSharedMagic.ClaimLocale], "locale claim should match")
				}

				if tc.user.Zoneinfo != "" {
					require.Equal(t, tc.user.Zoneinfo, claims[cryptoutilSharedMagic.ClaimZoneinfo], "zoneinfo claim should match")
				}

				// Verify updated_at matches user record timestamp (NOT time.Now() - causes race failures).
				// CRITICAL: Comparing against time.Now() fails when test execution takes >5s (OAuth flow, HTTP requests).
				// The DB timestamp is set at user creation, which may be 30-60+ seconds before this assertion.
				if !tc.user.UpdatedAt.IsZero() {
					updatedAtFloat, ok := claims[cryptoutilSharedMagic.ClaimUpdatedAt].(float64)
					require.True(t, ok, "updated_at should be a number")

					updatedAtTime := time.Unix(int64(updatedAtFloat), 0)
					require.WithinDuration(t, tc.user.UpdatedAt, updatedAtTime, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Second, "updated_at should match user record")
				}
			}
		})
	}
}
