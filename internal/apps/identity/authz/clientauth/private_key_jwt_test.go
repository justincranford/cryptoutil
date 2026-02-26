// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	rsa "crypto/rsa"
	"database/sql"
	json "encoding/json"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	joseJwa "github.com/lestrrat-go/jwx/v3/jwa"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJwt "github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityRepository "cryptoutil/internal/apps/identity/repository"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
)

// setupPrivateKeyTestDB creates an in-memory SQLite database for testing (using modernc.org/sqlite).
func setupPrivateKeyTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Create unique in-memory database per test.
	dbID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbID.String())

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	// Configure SQLite for concurrent writes (WAL mode, busy timeout).
	if _, err := sqlDB.ExecContext(context.Background(), "PRAGMA journal_mode=WAL;"); err != nil {
		require.NoError(t, fmt.Errorf("failed to enable WAL mode: %w", err))
	}

	if _, err := sqlDB.ExecContext(context.Background(), "PRAGMA busy_timeout = 30000;"); err != nil {
		require.NoError(t, fmt.Errorf("failed to set busy timeout: %w", err))
	}

	// Open GORM with modernc SQLite dialector.
	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto-migrate JTI replay cache table.
	err = db.AutoMigrate(&cryptoutilIdentityDomain.JTIReplayCache{})
	require.NoError(t, err)

	return db
}

func TestPrivateKeyJWTValidator_JTIReplayProtection(t *testing.T) {
	t.Parallel()

	db := setupPrivateKeyTestDB(t)
	jtiRepo := cryptoutilIdentityRepository.NewJTIReplayCacheRepository(db)

	ctx := context.Background()
	validator := NewPrivateKeyJWTValidator(testTokenEndpointURL, jtiRepo)

	// Generate RSA key pair.
	keyID := googleUuid.NewString()
	rsaKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	rsaPrivateKey, ok := rsaKeyPair.Private.(*rsa.PrivateKey)
	require.True(t, ok)

	privateJWK, err := joseJwk.Import(rsaPrivateKey)
	require.NoError(t, err)
	require.NoError(t, privateJWK.Set(joseJwk.KeyIDKey, keyID))
	require.NoError(t, privateJWK.Set(joseJwk.AlgorithmKey, joseJwa.RS256()))

	publicJWK, err := joseJwk.PublicKeyOf(privateJWK)
	require.NoError(t, err)

	publicKeySet := joseJwk.NewSet()
	require.NoError(t, publicKeySet.AddKey(publicJWK))

	publicKeySetJSON, err := json.Marshal(publicKeySet)
	require.NoError(t, err)

	clientID := googleUuid.New()
	client := &cryptoutilIdentityDomain.Client{
		ID:       clientID,
		ClientID: clientID.String(),
		JWKs:     string(publicKeySetJSON),
	}

	// Create two tokens with the same JTI.
	jti := googleUuid.NewString()
	now := time.Now().UTC()

	createToken := func() string {
		token := joseJwt.New()
		require.NoError(t, token.Set(joseJwt.IssuerKey, client.ClientID))
		require.NoError(t, token.Set(joseJwt.SubjectKey, client.ClientID))
		require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
		require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute)))
		require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
		require.NoError(t, token.Set(joseJwt.JwtIDKey, jti))

		signedToken, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.RS256(), privateJWK))
		require.NoError(t, err)

		return string(signedToken)
	}

	// First token should succeed.
	firstToken := createToken()
	_, err = validator.ValidateJWT(ctx, firstToken, client)
	require.NoError(t, err, "First token with unique JTI should succeed")

	// Second token with same JTI should fail (replay detected).
	secondToken := createToken()
	_, err = validator.ValidateJWT(ctx, secondToken, client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JTI replay detected", "Second token with same JTI should fail")
}

func TestPrivateKeyJWTValidator_AssertionLifetimeValidation(t *testing.T) {
	t.Parallel()

	db := setupPrivateKeyTestDB(t)
	jtiRepo := cryptoutilIdentityRepository.NewJTIReplayCacheRepository(db)

	ctx := context.Background()
	validator := NewPrivateKeyJWTValidator(testTokenEndpointURL, jtiRepo)

	// Generate RSA key pair.
	keyID := googleUuid.NewString()
	rsaKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	rsaPrivateKey, ok := rsaKeyPair.Private.(*rsa.PrivateKey)
	require.True(t, ok)

	privateJWK, err := joseJwk.Import(rsaPrivateKey)
	require.NoError(t, err)
	require.NoError(t, privateJWK.Set(joseJwk.KeyIDKey, keyID))
	require.NoError(t, privateJWK.Set(joseJwk.AlgorithmKey, joseJwa.RS256()))

	publicJWK, err := joseJwk.PublicKeyOf(privateJWK)
	require.NoError(t, err)

	publicKeySet := joseJwk.NewSet()
	require.NoError(t, publicKeySet.AddKey(publicJWK))

	publicKeySetJSON, err := json.Marshal(publicKeySet)
	require.NoError(t, err)

	clientID := googleUuid.New()
	client := &cryptoutilIdentityDomain.Client{
		ID:       clientID,
		ClientID: clientID.String(),
		JWKs:     string(publicKeySetJSON),
	}

	tests := []struct {
		name      string
		lifetime  time.Duration
		wantError bool
	}{
		{"valid lifetime (cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries minutes)", 5 * time.Minute, false},
		{"valid lifetime (maximum cryptoutilSharedMagic.JoseJADefaultMaxMaterials minutes)", 10 * time.Minute, false},
		{"invalid lifetime (15 minutes)", 15 * time.Minute, true},
		{"invalid lifetime (1 hour)", time.Hour, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			now := time.Now().UTC()
			token := joseJwt.New()
			require.NoError(t, token.Set(joseJwt.IssuerKey, client.ClientID))
			require.NoError(t, token.Set(joseJwt.SubjectKey, client.ClientID))
			require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
			require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(tc.lifetime)))
			require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
			require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

			signedToken, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.RS256(), privateJWK))
			require.NoError(t, err)

			_, err = validator.ValidateJWT(ctx, string(signedToken), client)

			if tc.wantError {
				require.Error(t, err)
				require.Contains(t, err.Error(), "exceeds maximum")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPrivateKeyJWTValidator_MissingJTI(t *testing.T) {
	t.Parallel()

	db := setupPrivateKeyTestDB(t)
	jtiRepo := cryptoutilIdentityRepository.NewJTIReplayCacheRepository(db)

	ctx := context.Background()
	validator := NewPrivateKeyJWTValidator(testTokenEndpointURL, jtiRepo)

	// Generate RSA key pair.
	keyID := googleUuid.NewString()
	rsaKeyPair, err := cryptoutilSharedCryptoKeygen.GenerateRSAKeyPair(cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	rsaPrivateKey, ok := rsaKeyPair.Private.(*rsa.PrivateKey)
	require.True(t, ok)

	privateJWK, err := joseJwk.Import(rsaPrivateKey)
	require.NoError(t, err)
	require.NoError(t, privateJWK.Set(joseJwk.KeyIDKey, keyID))
	require.NoError(t, privateJWK.Set(joseJwk.AlgorithmKey, joseJwa.RS256()))

	publicJWK, err := joseJwk.PublicKeyOf(privateJWK)
	require.NoError(t, err)

	publicKeySet := joseJwk.NewSet()
	require.NoError(t, publicKeySet.AddKey(publicJWK))

	publicKeySetJSON, err := json.Marshal(publicKeySet)
	require.NoError(t, err)

	clientID := googleUuid.New()
	client := &cryptoutilIdentityDomain.Client{
		ID:       clientID,
		ClientID: clientID.String(),
		JWKs:     string(publicKeySetJSON),
	}

	now := time.Now().UTC()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, client.ClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, client.ClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{testTokenEndpointURL}))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries*time.Minute)))
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	// Deliberately omit JTI claim.

	signedToken, err := joseJwt.Sign(token, joseJwt.WithKey(joseJwa.RS256(), privateJWK))
	require.NoError(t, err)

	_, err = validator.ValidateJWT(ctx, string(signedToken), client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing jti", "Token without JTI should fail validation")
}
