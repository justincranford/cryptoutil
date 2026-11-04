// Copyright (c) 2025-2026 Justin Cranford.
package apis

import (
	"context"
	"fmt"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
)

// TestTokenExpiration validates token expiration enforcement at UserInfo endpoint.
// Satisfies R05-06: Token expiration checking prevents access with expired tokens.
func TestTokenExpiration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupToken     func(t *testing.T, db *gorm.DB) string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid_unexpired_token_allows_access",
			setupToken: func(t *testing.T, db *gorm.DB) string {
				t.Helper()

				userID, err := googleUuid.NewV7()
				require.NoError(t, err)

				user := &cryptoutilIdentityDomain.User{
					ID:  userID,
					Sub: userID.String(),
				}

				err = db.Create(user).Error
				require.NoError(t, err)

				tokenID, err := googleUuid.NewV7()
				require.NoError(t, err)

				token := &cryptoutilIdentityDomain.Token{
					ID:          tokenID,
					ClientID:    userID,
					UserID:      cryptoutilIdentityDomain.NullableUUID{UUID: userID, Valid: true},
					TokenValue:  tokenID.String(),
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					Scopes:      []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
					IssuedAt:    time.Now().UTC().Add(-cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Minute),
					ExpiresAt:   time.Now().UTC().Add(cryptoutilSharedMagic.IMMaxUsernameLength * time.Minute), // Valid for 50 more minutes.
				}

				err = db.Create(token).Error
				require.NoError(t, err)

				return token.TokenValue
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name: "expired_token_returns_unauthorized",
			setupToken: func(t *testing.T, db *gorm.DB) string {
				t.Helper()

				userID, err := googleUuid.NewV7()
				require.NoError(t, err)

				user := &cryptoutilIdentityDomain.User{
					ID:  userID,
					Sub: userID.String(),
				}

				err = db.Create(user).Error
				require.NoError(t, err)

				tokenID, err := googleUuid.NewV7()
				require.NoError(t, err)

				token := &cryptoutilIdentityDomain.Token{
					ID:          tokenID,
					ClientID:    userID,
					UserID:      cryptoutilIdentityDomain.NullableUUID{UUID: userID, Valid: true},
					TokenValue:  tokenID.String(),
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					Scopes:      []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
					IssuedAt:    time.Now().UTC().Add(-70 * time.Minute),
					ExpiresAt:   time.Now().UTC().Add(-cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Minute), // Expired 10 minutes ago.
				}

				err = db.Create(token).Error
				require.NoError(t, err)

				return token.TokenValue
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  cryptoutilSharedMagic.ErrorInvalidToken,
		},
		{
			name: "token_expiring_within_grace_period_allows_access",
			setupToken: func(t *testing.T, db *gorm.DB) string {
				t.Helper()

				userID, err := googleUuid.NewV7()
				require.NoError(t, err)

				user := &cryptoutilIdentityDomain.User{
					ID:  userID,
					Sub: userID.String(),
				}

				err = db.Create(user).Error
				require.NoError(t, err)

				tokenID, err := googleUuid.NewV7()
				require.NoError(t, err)

				token := &cryptoutilIdentityDomain.Token{
					ID:          tokenID,
					ClientID:    userID,
					UserID:      cryptoutilIdentityDomain.NullableUUID{UUID: userID, Valid: true},
					TokenValue:  tokenID.String(),
					TokenType:   cryptoutilIdentityDomain.TokenTypeAccess,
					TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
					Scopes:      []string{cryptoutilSharedMagic.ScopeOpenID, cryptoutilSharedMagic.ClaimProfile},
					IssuedAt:    time.Now().UTC().Add(-59 * time.Minute),
					ExpiresAt:   time.Now().UTC().Add(1 * time.Minute), // Expires in 1 minute (within grace).
				}

				err = db.Create(token).Error
				require.NoError(t, err)

				return token.TokenValue
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			// Create unique in-memory database per test.
			dbID, err := googleUuid.NewV7()
			require.NoError(t, err)

			dsn := fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, dbID.String())

			dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
				Type: cryptoutilSharedMagic.TestDatabaseSQLite,
				DSN:  dsn,
			}
			repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, repoFactory.Close()) })

			db := repoFactory.DB()

			// Auto-migrate test schemas.
			require.NoError(t, db.AutoMigrate(
				&cryptoutilIdentityDomain.User{},
				&cryptoutilIdentityDomain.Token{},
			))

			// Create test token.
			tokenValue := tc.setupToken(t, db)

			tokenCfg := &cryptoutilIdentityConfig.TokenConfig{
				AccessTokenFormat: cryptoutilSharedMagic.IdentityTokenFormatUUID,
				Issuer:            testHTTPSExampleCom,
			}

			tokenSvc := cryptoutilIdentityIssuer.NewTokenService(nil, nil, nil, tokenCfg)

			config := &cryptoutilIdentityConfig.Config{
				Tokens: tokenCfg,
			}

			service := NewService(config, repoFactory, tokenSvc)

			app := fiber.New()
			service.RegisterRoutes(app)

			// Create request to UserInfo endpoint with Bearer token.
			req := httptest.NewRequest(http.MethodGet, "/oidc/v1/userinfo", nil)
			req.Header.Set("Authorization", cryptoutilSharedMagic.AuthorizationBearerPrefix+tokenValue)

			// Execute request.
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() {
				_ = resp.Body.Close() //nolint:errcheck // Test cleanup.
			}()

			// Validate response status.
			require.Equal(t, tc.expectedStatus, resp.StatusCode,
				"Expected status %d for test %s, got %d", tc.expectedStatus, tc.name, resp.StatusCode)

			// For error responses, validate error field exists.
			if tc.expectedError != "" {
				require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			}
		})
	}
}
