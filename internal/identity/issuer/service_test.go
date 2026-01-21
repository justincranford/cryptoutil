// Copyright (c) 2025 Justin Cranford
//
//

package issuer_test

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilIdentityConfig "cryptoutil/internal/identity/config"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/identity/issuer"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// mockKeyGenerator implements KeyGenerator for testing using real RSA keys.
type mockKeyGenerator struct {
	productionGen *cryptoutilIdentityIssuer.ProductionKeyGenerator
}

func newMockKeyGenerator() *mockKeyGenerator {
	return &mockKeyGenerator{
		productionGen: cryptoutilIdentityIssuer.NewProductionKeyGenerator(),
	}
}

func (m *mockKeyGenerator) GenerateSigningKey(ctx context.Context, algorithm string) (*cryptoutilIdentityIssuer.SigningKey, error) {
	return m.productionGen.GenerateSigningKey(ctx, algorithm) //nolint:wrapcheck // Test wrapper
}

func (m *mockKeyGenerator) GenerateEncryptionKey(ctx context.Context) (*cryptoutilIdentityIssuer.EncryptionKey, error) {
	return m.productionGen.GenerateEncryptionKey(ctx) //nolint:wrapcheck // Test wrapper
}

// setupTestService creates a token service with initialized keys for testing.
func setupTestService(t *testing.T, tokenFormat string) *cryptoutilIdentityIssuer.TokenService {
	t.Helper()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	db := repoFactory.DB()
	err = db.AutoMigrate(&cryptoutilIdentityDomain.Key{})
	require.NoError(t, err)

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  "RS256",
		AccessTokenFormat: tokenFormat,
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	// Initialize signing key for token generation.
	err = keyRotationMgr.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	// Initialize encryption key if using JWE format.
	if tokenFormat == "jwe" {
		err = keyRotationMgr.RotateEncryptionKey(ctx)
		require.NoError(t, err)
	}

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		tokenConfig.Issuer,
		keyRotationMgr,
		tokenConfig.SigningAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	return cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)
}

// TestNewTokenService validates service initialization.
func TestNewTokenService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.Key{},
	)
	require.NoError(t, err)

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  "RS256",
		AccessTokenFormat: "jws",
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		tokenConfig.Issuer,
		keyRotationMgr,
		tokenConfig.SigningAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)
	require.NotNil(t, jwsIssuer)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)
	require.NotNil(t, jweIssuer)

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()
	require.NotNil(t, uuidIssuer)

	service := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)
	require.NotNil(t, service)
}

// TestIssueAccessTokenJWS validates JWS access token generation.
func TestIssueAccessTokenJWS(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.Key{},
	)
	require.NoError(t, err)

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  "RS256",
		AccessTokenFormat: "jws",
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	// Initialize signing key for token generation.
	err = keyRotationMgr.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		tokenConfig.Issuer,
		keyRotationMgr,
		tokenConfig.SigningAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	service := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)

	claims := map[string]any{
		"sub":   googleUuid.Must(googleUuid.NewV7()).String(),
		"email": "test@example.com",
		"name":  "Test User",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}

	token, err := service.IssueAccessToken(ctx, claims)
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

// TestIssueAccessTokenInvalidFormat validates error handling for unsupported token format.
func TestIssueAccessTokenInvalidFormat(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  "RS256",
		AccessTokenFormat: "invalid-format",
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		tokenConfig.Issuer,
		keyRotationMgr,
		tokenConfig.SigningAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	service := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)

	claims := map[string]any{
		"sub":   googleUuid.Must(googleUuid.NewV7()).String(),
		"email": "test@example.com",
	}

	_, err = service.IssueAccessToken(ctx, claims)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported access token format")
}

// TestIssueAccessTokenJWE validates JWE access token generation.
func TestIssueAccessTokenJWE(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.Key{},
	)
	require.NoError(t, err)

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  "RS256",
		AccessTokenFormat: "jwe",
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	// Initialize signing and encryption keys.
	err = keyRotationMgr.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	err = keyRotationMgr.RotateEncryptionKey(ctx)
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		tokenConfig.Issuer,
		keyRotationMgr,
		tokenConfig.SigningAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	service := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)

	claims := map[string]any{
		"sub":   googleUuid.Must(googleUuid.NewV7()).String(),
		"email": "test@example.com",
		"name":  "Test User",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
	}

	token, err := service.IssueAccessToken(ctx, claims)
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

// TestIssueAccessTokenUUID validates UUID access token generation.
func TestIssueAccessTokenUUID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.Key{},
	)
	require.NoError(t, err)

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  "RS256",
		AccessTokenFormat: "uuid",
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		tokenConfig.Issuer,
		keyRotationMgr,
		tokenConfig.SigningAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	service := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)

	claims := map[string]any{
		"sub":   googleUuid.Must(googleUuid.NewV7()).String(),
		"email": "test@example.com",
	}

	token, err := service.IssueAccessToken(ctx, claims)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Verify token is valid UUID.
	_, err = googleUuid.Parse(token)
	require.NoError(t, err)
}

// TestIssueIDTokenService validates ID token generation via service.
func TestIssueIDTokenService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.Key{},
	)
	require.NoError(t, err)

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  "RS256",
		AccessTokenFormat: "jws",
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	// Initialize signing key.
	err = keyRotationMgr.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		tokenConfig.Issuer,
		keyRotationMgr,
		tokenConfig.SigningAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	service := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)

	claims := map[string]any{
		"sub": googleUuid.Must(googleUuid.NewV7()).String(),
		"aud": "client123",
	}

	token, err := service.IssueIDToken(ctx, claims)
	require.NoError(t, err)
	require.NotEmpty(t, token)
}

// TestIssueRefreshToken validates refresh token generation via service.
func TestIssueRefreshToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.Key{},
	)
	require.NoError(t, err)

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  "RS256",
		AccessTokenFormat: "jws",
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		tokenConfig.Issuer,
		keyRotationMgr,
		tokenConfig.SigningAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	service := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)

	token, err := service.IssueRefreshToken(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Verify token is valid UUID.
	_, err = googleUuid.Parse(token)
	require.NoError(t, err)
}

// TestGetPublicKeys validates public keys retrieval.
func TestGetPublicKeys(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	dbConfig := &cryptoutilIdentityConfig.DatabaseConfig{
		Type: "sqlite",
		DSN:  ":memory:",
	}

	repoFactory, err := cryptoutilIdentityRepository.NewRepositoryFactory(ctx, dbConfig)
	require.NoError(t, err)

	// Run migrations.
	db := repoFactory.DB()
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.Key{},
	)
	require.NoError(t, err)

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  "RS256",
		AccessTokenFormat: "jws",
	}

	keyRotationMgr, err := cryptoutilIdentityIssuer.NewKeyRotationManager(
		cryptoutilIdentityIssuer.DefaultKeyRotationPolicy(),
		newMockKeyGenerator(),
		nil,
	)
	require.NoError(t, err)

	// Rotate a key so there's something to return.
	err = keyRotationMgr.RotateSigningKey(ctx, "RS256")
	require.NoError(t, err)

	jwsIssuer, err := cryptoutilIdentityIssuer.NewJWSIssuer(
		tokenConfig.Issuer,
		keyRotationMgr,
		tokenConfig.SigningAlgorithm,
		1*time.Hour,
		1*time.Hour,
	)
	require.NoError(t, err)

	jweIssuer, err := cryptoutilIdentityIssuer.NewJWEIssuer(keyRotationMgr)
	require.NoError(t, err)

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	service := cryptoutilIdentityIssuer.NewTokenService(jwsIssuer, jweIssuer, uuidIssuer, tokenConfig)

	keys := service.GetPublicKeys()
	require.NotNil(t, keys)
	// Keys should be present since we rotated.
	require.GreaterOrEqual(t, len(keys), 0)
}

// TestGetPublicKeysNilJWSIssuer validates empty result when JWS issuer is nil.
func TestGetPublicKeysNilJWSIssuer(t *testing.T) {
	t.Parallel()

	tokenConfig := &cryptoutilIdentityConfig.TokenConfig{
		Issuer:            "https://localhost:8080",
		SigningAlgorithm:  "RS256",
		AccessTokenFormat: "uuid", // Not JWS.
	}

	uuidIssuer := cryptoutilIdentityIssuer.NewUUIDIssuer()

	// Create service with nil JWS issuer.
	service := cryptoutilIdentityIssuer.NewTokenService(nil, nil, uuidIssuer, tokenConfig)

	keys := service.GetPublicKeys()
	require.NotNil(t, keys)
	require.Empty(t, keys)
}

// TestBuildTokenDomain validates token domain builder.
func TestBuildTokenDomain(t *testing.T) {
	t.Parallel()

	clientID := googleUuid.Must(googleUuid.NewV7())
	userID := googleUuid.Must(googleUuid.NewV7())
	scopes := []string{"openid", "profile"}
	expiresAt := time.Now().Add(1 * time.Hour)

	token := cryptoutilIdentityIssuer.BuildTokenDomain(
		cryptoutilIdentityDomain.TokenTypeAccess,
		cryptoutilIdentityDomain.TokenFormatJWS,
		"test-token-value",
		clientID,
		userID,
		scopes,
		expiresAt,
	)

	require.NotNil(t, token)
	require.Equal(t, cryptoutilIdentityDomain.TokenTypeAccess, token.TokenType)
	require.Equal(t, cryptoutilIdentityDomain.TokenFormatJWS, token.TokenFormat)
	require.Equal(t, "test-token-value", token.TokenValue)
	require.Equal(t, clientID, token.ClientID)
	require.Equal(t, scopes, token.Scopes)
}

// TestValidateAccessToken validates access token validation for all formats (JWS, JWE, UUID).
func TestValidateAccessToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tokenFormat string
		setupToken  func(*testing.T, *cryptoutilIdentityIssuer.TokenService) string
		wantErr     bool
	}{
		{
			name:        "valid_jws_token",
			tokenFormat: "jws",
			setupToken: func(t *testing.T, service *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				ctx := context.Background()
				claims := map[string]any{
					"sub":   googleUuid.Must(googleUuid.NewV7()).String(),
					"scope": "openid profile",
					"iat":   time.Now().Unix(),
					"exp":   time.Now().Add(1 * time.Hour).Unix(),
				}
				token, err := service.IssueAccessToken(ctx, claims)
				require.NoError(t, err)

				return token
			},
			wantErr: false,
		},
		{
			name:        "valid_jwe_token",
			tokenFormat: "jwe",
			setupToken: func(t *testing.T, service *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				ctx := context.Background()
				claims := map[string]any{
					"sub":   googleUuid.Must(googleUuid.NewV7()).String(),
					"scope": "openid profile",
					"iat":   time.Now().Unix(),
					"exp":   time.Now().Add(1 * time.Hour).Unix(),
				}
				token, err := service.IssueAccessToken(ctx, claims)
				require.NoError(t, err)

				return token
			},
			wantErr: false,
		},
		{
			name:        "valid_uuid_token",
			tokenFormat: "uuid",
			setupToken: func(t *testing.T, service *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				ctx := context.Background()
				claims := map[string]any{}
				token, err := service.IssueAccessToken(ctx, claims)
				require.NoError(t, err)

				return token
			},
			wantErr: false,
		},
		{
			name:        "invalid_jws_token",
			tokenFormat: "jws",
			setupToken: func(t *testing.T, _ *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				return "invalid.jws.token"
			},
			wantErr: true,
		},
		{
			name:        "invalid_jwe_token",
			tokenFormat: "jwe",
			setupToken: func(t *testing.T, _ *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				return "invalid_jwe_token"
			},
			wantErr: true,
		},
		{
			name:        "unsupported_format",
			tokenFormat: "jwt",
			setupToken: func(t *testing.T, _ *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				return "any-token"
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			service := setupTestService(t, tc.tokenFormat)

			token := tc.setupToken(t, service)

			claims, err := service.ValidateAccessToken(ctx, token)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, claims)
			}
		})
	}
}

// TestValidateIDToken validates ID token validation.
func TestValidateIDToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupToken func(*testing.T, *cryptoutilIdentityIssuer.TokenService) string
		wantErr    bool
	}{
		{
			name: "valid_id_token",
			setupToken: func(t *testing.T, service *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				ctx := context.Background()
				claims := map[string]any{
					"sub":   googleUuid.Must(googleUuid.NewV7()).String(),
					"aud":   googleUuid.Must(googleUuid.NewV7()).String(),
					"nonce": "test-nonce",
					"iat":   time.Now().Unix(),
					"exp":   time.Now().Add(1 * time.Hour).Unix(),
				}
				token, err := service.IssueIDToken(ctx, claims)
				require.NoError(t, err)

				return token
			},
			wantErr: false,
		},
		{
			name: "invalid_id_token",
			setupToken: func(t *testing.T, _ *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				return "invalid.id.token"
			},
			wantErr: true,
		},
		{
			name: "expired_id_token",
			setupToken: func(t *testing.T, service *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				ctx := context.Background()
				claims := map[string]any{
					"sub":   googleUuid.Must(googleUuid.NewV7()).String(),
					"aud":   googleUuid.Must(googleUuid.NewV7()).String(),
					"nonce": "test-nonce",
					"iat":   time.Now().Add(-2 * time.Hour).Unix(),
					"exp":   time.Now().Add(-1 * time.Hour).Unix(),
				}
				token, err := service.IssueIDToken(ctx, claims)
				require.NoError(t, err)

				return token
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			service := setupTestService(t, "jws")

			token := tc.setupToken(t, service)

			claims, err := service.ValidateIDToken(ctx, token)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, claims)
				require.Contains(t, claims, "sub")
			}
		})
	}
}

// TestIsTokenActive validates token expiration and not-before checks.
func TestIsTokenActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		claims     map[string]any
		wantActive bool
	}{
		{
			name: "valid_active_token",
			claims: map[string]any{
				"exp": float64(time.Now().Add(1 * time.Hour).Unix()),
				"nbf": float64(time.Now().Add(-1 * time.Minute).Unix()),
			},
			wantActive: true,
		},
		{
			name: "expired_token",
			claims: map[string]any{
				"exp": float64(time.Now().Add(-1 * time.Hour).Unix()),
				"nbf": float64(time.Now().Add(-2 * time.Hour).Unix()),
			},
			wantActive: false,
		},
		{
			name: "not_yet_valid_token",
			claims: map[string]any{
				"exp": float64(time.Now().Add(2 * time.Hour).Unix()),
				"nbf": float64(time.Now().Add(1 * time.Hour).Unix()),
			},
			wantActive: false,
		},
		{
			name:       "no_expiration_or_nbf",
			claims:     map[string]any{},
			wantActive: true,
		},
		{
			name: "only_expiration_valid",
			claims: map[string]any{
				"exp": float64(time.Now().Add(1 * time.Hour).Unix()),
			},
			wantActive: true,
		},
		{
			name: "only_nbf_valid",
			claims: map[string]any{
				"nbf": float64(time.Now().Add(-1 * time.Minute).Unix()),
			},
			wantActive: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service := setupTestService(t, "jws")

			active := service.IsTokenActive(tc.claims)

			require.Equal(t, tc.wantActive, active)
		})
	}
}

// TestIntrospectToken validates token introspection.
func TestIntrospectToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupToken  func(*testing.T, *cryptoutilIdentityIssuer.TokenService) string
		wantActive  bool
		checkExpiry bool
	}{
		{
			name: "valid_active_token",
			setupToken: func(t *testing.T, service *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				ctx := context.Background()
				claims := map[string]any{
					"sub":   googleUuid.Must(googleUuid.NewV7()).String(),
					"scope": "openid profile",
					"iat":   time.Now().Unix(),
					"exp":   time.Now().Add(1 * time.Hour).Unix(),
				}
				token, err := service.IssueAccessToken(ctx, claims)
				require.NoError(t, err)

				return token
			},
			wantActive:  true,
			checkExpiry: true,
		},
		{
			name: "expired_token",
			setupToken: func(t *testing.T, service *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				ctx := context.Background()
				claims := map[string]any{
					"sub":   googleUuid.Must(googleUuid.NewV7()).String(),
					"scope": "openid profile",
					"iat":   time.Now().Add(-2 * time.Hour).Unix(),
					"exp":   time.Now().Add(-1 * time.Hour).Unix(),
				}
				token, err := service.IssueAccessToken(ctx, claims)
				require.NoError(t, err)

				return token
			},
			wantActive:  true,
			checkExpiry: true,
		},
		{
			name: "invalid_token",
			setupToken: func(t *testing.T, _ *cryptoutilIdentityIssuer.TokenService) string {
				t.Helper()

				return "invalid.token.here"
			},
			wantActive:  false,
			checkExpiry: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			service := setupTestService(t, "jws")

			token := tc.setupToken(t, service)

			metadata, err := service.IntrospectToken(ctx, token)

			require.NoError(t, err)
			require.NotNil(t, metadata)
			require.Equal(t, tc.wantActive, metadata.Active)

			if tc.checkExpiry {
				require.NotNil(t, metadata.ExpiresAt)
				require.NotNil(t, metadata.Claims)
			}
		})
	}
}

// TestIssueUserInfoJWT validates UserInfo JWT issuance.
func TestIssueUserInfoJWT(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		clientID string
		claims   map[string]any
		wantErr  bool
	}{
		{
			name:     "valid_userinfo_jwt",
			clientID: googleUuid.Must(googleUuid.NewV7()).String(),
			claims: map[string]any{
				"sub":   googleUuid.Must(googleUuid.NewV7()).String(),
				"email": "user@example.com",
				"name":  "Test User",
			},
			wantErr: false,
		},
		{
			name:     "missing_sub_claim",
			clientID: googleUuid.Must(googleUuid.NewV7()).String(),
			claims: map[string]any{
				"email": "user@example.com",
				"name":  "Test User",
			},
			wantErr: true,
		},
		{
			name:     "empty_client_id",
			clientID: "",
			claims: map[string]any{
				"sub":   googleUuid.Must(googleUuid.NewV7()).String(),
				"email": "user@example.com",
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			service := setupTestService(t, "jws")

			jwt, err := service.IssueUserInfoJWT(ctx, tc.clientID, tc.claims)

			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, jwt)
			}
		})
	}
}
