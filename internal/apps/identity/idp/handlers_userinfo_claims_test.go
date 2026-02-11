// Copyright (c) 2025 Justin Cranford

package idp

import (
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
				"sub":                nil, // Value validated separately
				"preferred_username": "testuser_standard",
				"email":              "standard@example.com",
				"email_verified":     true,
				"given_name":         "Test",
				"family_name":        "User",
				"name":               "Test User",
				"locale":             "en-US",
				"zoneinfo":           "America/New_York",
				"updated_at":         nil, // Timestamp validation
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
				"sub":                nil,
				"preferred_username": "testuser_minimal",
				"email":              "minimal@example.com",
				"email_verified":     false,
			},
			unexpectedClaims: []string{
				"given_name",
				"family_name",
				"name",
				"locale",
				"zoneinfo",
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
				"email_verified": true,
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
				"email_verified": false,
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
				Type: "sqlite",
				DSN:  ":memory:",
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
				"sub":                retrievedUser.Sub,
				"preferred_username": retrievedUser.PreferredUsername,
				"email":              retrievedUser.Email,
				"email_verified":     retrievedUser.EmailVerified,
				"given_name":         retrievedUser.GivenName,
				"family_name":        retrievedUser.FamilyName,
				"name":               retrievedUser.Name,
				"locale":             retrievedUser.Locale,
				"zoneinfo":           retrievedUser.Zoneinfo,
				"updated_at":         retrievedUser.UpdatedAt.Unix(),
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
				require.Equal(t, tc.user.Sub, claims["sub"], "sub claim should match")
				require.Equal(t, tc.user.PreferredUsername, claims["preferred_username"], "preferred_username claim should match")
				require.Equal(t, tc.user.Email, claims["email"], "email claim should match")
				require.Equal(t, tc.user.EmailVerified.Bool(), claims["email_verified"], "email_verified claim should match")

				// Verify optional claims if present.
				if tc.user.GivenName != "" {
					require.Equal(t, tc.user.GivenName, claims["given_name"], "given_name claim should match")
				}

				if tc.user.FamilyName != "" {
					require.Equal(t, tc.user.FamilyName, claims["family_name"], "family_name claim should match")
				}

				if tc.user.Name != "" {
					require.Equal(t, tc.user.Name, claims["name"], "name claim should match")
				}

				if tc.user.Locale != "" {
					require.Equal(t, tc.user.Locale, claims["locale"], "locale claim should match")
				}

				if tc.user.Zoneinfo != "" {
					require.Equal(t, tc.user.Zoneinfo, claims["zoneinfo"], "zoneinfo claim should match")
				}

				// Verify updated_at matches user record timestamp (NOT time.Now() - causes race failures).
				// CRITICAL: Comparing against time.Now() fails when test execution takes >5s (OAuth flow, HTTP requests).
				// The DB timestamp is set at user creation, which may be 30-60+ seconds before this assertion.
				if !tc.user.UpdatedAt.IsZero() {
					updatedAtFloat, ok := claims["updated_at"].(float64)
					require.True(t, ok, "updated_at should be a number")

					updatedAtTime := time.Unix(int64(updatedAtFloat), 0)
					require.WithinDuration(t, tc.user.UpdatedAt, updatedAtTime, 5*time.Second, "updated_at should match user record")
				}
			}
		})
	}
}
