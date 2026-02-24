// Copyright (c) 2025 Justin Cranford
//
//

package authz_test

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestHandleTokenAuthorizationCodeGrant_MissingParameters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create test dependencies.
	repoFactory, tokenSvc := setupAuthzTestDependencies(ctx, t)

	// Create authz service.
	cfg := &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://issuer.example.com",
		},
	}
	authzSvc := cryptoutilIdentityAuthz.NewService(cfg, repoFactory, tokenSvc)

	// Create fiber app.
	app := fiber.New()
	authzSvc.RegisterRoutes(app)

	tests := []struct {
		name           string
		formData       map[string]string
		wantStatusCode int
		wantError      string
	}{
		{
			name: "missing code",
			formData: map[string]string{
				"grant_type":    cryptoutilSharedMagic.GrantTypeAuthorizationCode,
				"redirect_uri":  "https://client.example.com/callback",
				"client_id":     "test-client",
				"code_verifier": "test-verifier",
			},
			wantStatusCode: 400,
			wantError:      "code is required",
		},
		{
			name: "missing redirect_uri",
			formData: map[string]string{
				"grant_type":    cryptoutilSharedMagic.GrantTypeAuthorizationCode,
				"code":          "test-code",
				"client_id":     "test-client",
				"code_verifier": "test-verifier",
			},
			wantStatusCode: 400,
			wantError:      "redirect_uri is required",
		},
		{
			name: "missing client_id",
			formData: map[string]string{
				"grant_type":    cryptoutilSharedMagic.GrantTypeAuthorizationCode,
				"code":          "test-code",
				"redirect_uri":  "https://client.example.com/callback",
				"code_verifier": "test-verifier",
			},
			wantStatusCode: 400,
			wantError:      "client_id is required",
		},
		{
			name: "missing code_verifier (PKCE required)",
			formData: map[string]string{
				"grant_type":   cryptoutilSharedMagic.GrantTypeAuthorizationCode,
				"code":         "test-code",
				"redirect_uri": "https://client.example.com/callback",
				"client_id":    "test-client",
			},
			wantStatusCode: 400,
			wantError:      "code_verifier is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Build form data.
			formData := url.Values{}
			for k, v := range tc.formData {
				formData.Set(k, v)
			}

			// Create request.
			req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			// Execute request.
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

			require.Equal(t, tc.wantStatusCode, resp.StatusCode)
		})
	}
}

func TestHandleTokenClientCredentialsGrant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create test dependencies.
	repoFactory, tokenSvc := setupAuthzTestDependencies(ctx, t)

	// Create test client.
	db := repoFactory.DB()
	testClient := &cryptoutilIdentityDomain.Client{
		ID:                      googleUuid.New(),
		ClientID:                "test-client-credentials",
		Name:                    "Test Client Credentials",
		ClientType:              cryptoutilIdentityDomain.ClientTypeConfidential,
		AllowedGrantTypes:       []string{cryptoutilSharedMagic.GrantTypeClientCredentials},
		AllowedScopes:           []string{"read", "write"},
		AccessTokenLifetime:     3600,
		RefreshTokenLifetime:    86400,
		TokenEndpointAuthMethod: cryptoutilIdentityDomain.ClientAuthMethodSecretBasic,
		CreatedAt:               time.Now().UTC(),
		UpdatedAt:               time.Now().UTC(),
	}
	err := db.Create(testClient).Error
	require.NoError(t, err)

	// Create authz service.
	cfg := &cryptoutilIdentityConfig.Config{
		Tokens: &cryptoutilIdentityConfig.TokenConfig{
			Issuer: "https://issuer.example.com",
		},
	}
	authzSvc := cryptoutilIdentityAuthz.NewService(cfg, repoFactory, tokenSvc)

	// Create fiber app.
	app := fiber.New()
	authzSvc.RegisterRoutes(app)

	// Build form data.
	formData := url.Values{}
	formData.Set("grant_type", cryptoutilSharedMagic.GrantTypeClientCredentials)
	formData.Set("client_id", "test-client-credentials")
	formData.Set("scope", "read write")

	// Create request.
	req := httptest.NewRequest("POST", "/oauth2/v1/token", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute request (will fail due to missing client authentication, but exercises handler code).
	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Test cleanup

	// Should fail with 401 (client authentication required).
	require.Equal(t, 401, resp.StatusCode)
}

// setupAuthzTestDependencies creates test repository and token service.
func setupAuthzTestDependencies(ctx context.Context, t *testing.T) (*cryptoutilIdentityRepository.RepositoryFactory, *cryptoutilIdentityIssuer.TokenService) {
	t.Helper()

	// Create test database.
	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}
	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)
	t.Cleanup(func() {
		sqlDB, _ := repoFactory.DB().DB() //nolint:errcheck // Test cleanup
		if sqlDB != nil {
			_ = sqlDB.Close() //nolint:errcheck // Test cleanup
		}
	})

	// Run migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.Client{},
		&cryptoutilIdentityDomain.User{},
		&cryptoutilIdentityDomain.AuthorizationRequest{},
	)
	require.NoError(t, err)

	// Create token service.
	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:               "https://issuer.example.com",
		AccessTokenFormat:    cryptoutilSharedMagic.TokenFormatJWS,
		AccessTokenLifetime:  cryptoutilSharedMagic.DefaultAccessTokenLifetime,
		RefreshTokenLifetime: cryptoutilSharedMagic.DefaultRefreshTokenLifetime,
		IDTokenLifetime:      cryptoutilSharedMagic.DefaultIDTokenLifetime,
		SigningAlgorithm:     "RS256",
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		&mockKeyGenerator{},
		nil,
	)
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		tokenConfig.Issuer,
		keyRotationMgr,
		tokenConfig.SigningAlgorithm,
		tokenConfig.AccessTokenLifetime,
		tokenConfig.IDTokenLifetime,
	)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	tokenSvc := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)

	return repoFactory, tokenSvc
}
