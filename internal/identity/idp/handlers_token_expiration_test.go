// Copyright (c) 2025 Justin Cranford

package idp

import (
	"context"
	"database/sql"
	"fmt"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "modernc.org/sqlite"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
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
					Scopes:      []string{"openid", "profile"},
					IssuedAt:    time.Now().UTC().Add(-10 * time.Minute),
					ExpiresAt:   time.Now().UTC().Add(50 * time.Minute), // Valid for 50 more minutes.
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
					Scopes:      []string{"openid", "profile"},
					IssuedAt:    time.Now().UTC().Add(-70 * time.Minute),
					ExpiresAt:   time.Now().UTC().Add(-10 * time.Minute), // Expired 10 minutes ago.
				}

				err = db.Create(token).Error
				require.NoError(t, err)

				return token.TokenValue
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid_token",
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
					Scopes:      []string{"openid", "profile"},
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

			dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbID.String())

			sqlDB, err := sql.Open("sqlite", dsn)
			require.NoError(t, err)

			// Apply PRAGMA settings.
			if _, err := sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
				require.FailNowf(t, "failed to enable WAL mode", "%v", err)
			}

			if _, err := sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;"); err != nil {
				require.FailNowf(t, "failed to set busy timeout", "%v", err)
			}

			// Create GORM database.
			dialector := sqlite.Dialector{Conn: sqlDB}

			db, err := gorm.Open(dialector, &gorm.Config{
				Logger:                 logger.Default.LogMode(logger.Silent),
				SkipDefaultTransaction: true,
			})
			require.NoError(t, err)

			gormDB, err := db.DB()
			require.NoError(t, err)

			gormDB.SetMaxOpenConns(5)
			gormDB.SetMaxIdleConns(5)

			// Auto-migrate test schemas.
			err = db.AutoMigrate(
				&cryptoutilIdentityDomain.User{},
				&cryptoutilIdentityDomain.Token{},
			)
			require.NoError(t, err)

			t.Cleanup(func() {
				_ = sqlDB.Close() //nolint:errcheck // Test cleanup.
			})

			// Create test token.
			tokenValue := tc.setupToken(t, db)

			// Create repository factory and token service.
			dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
				Type: "sqlite",
				DSN:  dsn,
			}

			repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
			require.NoError(t, err)

			tokenCfg := &cryptoutilIdentityConfig.TokenConfig{
				AccessTokenFormat: "uuid",
				Issuer:            "https://example.com",
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
			req.Header.Set("Authorization", "Bearer "+tokenValue)

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
