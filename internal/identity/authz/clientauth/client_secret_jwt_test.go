// Copyright (c) 2025 Justin Cranford
//
//

package clientauth

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	joseJwt "github.com/lestrrat-go/jwx/v3/jwt"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Create unique in-memory database per test.
	dbID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbID.String())

	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	if _, err := sqlDB.ExecContext(context.Background(), "PRAGMA journal_mode=WAL;"); err != nil {
		require.FailNowf(t, "failed to enable WAL mode", "%v", err)
	}

	if _, err := sqlDB.ExecContext(context.Background(), "PRAGMA busy_timeout = 30000;"); err != nil {
		require.FailNowf(t, "failed to set busy timeout", "%v", err)
	}

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
	gormDB.SetConnMaxLifetime(0)
	gormDB.SetConnMaxIdleTime(0)

	// Auto-migrate JTI replay cache table.
	err = db.AutoMigrate(&cryptoutilIdentityDomain.JTIReplayCache{})
	require.NoError(t, err)

	return db
}

// TestClientSecretJWTValidator_JTIReplayProtection tests JTI replay attack prevention.
func TestClientSecretJWTValidator_JTIReplayProtection(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	jtiRepo := cryptoutilIdentityRepository.NewJTIReplayCacheRepository(db)
	validator := NewClientSecretJWTValidator("https://auth.example.com/token", jtiRepo)

	clientID := googleUuid.New()
	client := &cryptoutilIdentityDomain.Client{
		ID:           clientID,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}

	// Create first JWT with jti claim
	now := time.Now().UTC()
	firstToken := joseJwt.New()
	require.NoError(t, firstToken.Set(joseJwt.IssuerKey, client.ClientID))
	require.NoError(t, firstToken.Set(joseJwt.SubjectKey, client.ClientID))
	require.NoError(t, firstToken.Set(joseJwt.AudienceKey, []string{"https://auth.example.com/token"}))
	require.NoError(t, firstToken.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, firstToken.Set(joseJwt.ExpirationKey, now.Add(5*time.Minute)))
	require.NoError(t, firstToken.Set(joseJwt.JwtIDKey, "unique-jti-1"))

	// First validation should succeed (JTI stored in cache)
	err := validator.validateClaims(context.Background(), firstToken, client)
	require.NoError(t, err)

	// Second validation with same JTI should fail (replay detected)
	secondToken := joseJwt.New()
	require.NoError(t, secondToken.Set(joseJwt.IssuerKey, client.ClientID))
	require.NoError(t, secondToken.Set(joseJwt.SubjectKey, client.ClientID))
	require.NoError(t, secondToken.Set(joseJwt.AudienceKey, []string{"https://auth.example.com/token"}))
	require.NoError(t, secondToken.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, secondToken.Set(joseJwt.ExpirationKey, now.Add(5*time.Minute)))
	require.NoError(t, secondToken.Set(joseJwt.JwtIDKey, "unique-jti-1")) // Same JTI!

	err = validator.validateClaims(context.Background(), secondToken, client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JTI replay detected")
}

// TestClientSecretJWTValidator_AssertionLifetimeValidation tests assertion lifetime validation.
func TestClientSecretJWTValidator_AssertionLifetimeValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		lifetime  time.Duration
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "valid lifetime (5 minutes)",
			lifetime: 5 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "valid lifetime (maximum 10 minutes)",
			lifetime: cryptoutilIdentityMagic.JWTAssertionMaxLifetime,
			wantErr:  false,
		},
		{
			name:      "invalid lifetime (15 minutes)",
			lifetime:  15 * time.Minute,
			wantErr:   true,
			errSubstr: "assertion lifetime",
		},
		{
			name:      "invalid lifetime (1 hour)",
			lifetime:  time.Hour,
			wantErr:   true,
			errSubstr: "exceeds maximum",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create validator without JTI repo for isolated assertion lifetime testing
			validator := NewClientSecretJWTValidator("https://auth.example.com/token", nil)

			clientID := googleUuid.New()
			client := &cryptoutilIdentityDomain.Client{
				ID:           clientID,
				ClientID:     "test-client",
				ClientSecret: "test-secret",
			}

			now := time.Now().UTC()
			token := joseJwt.New()
			require.NoError(t, token.Set(joseJwt.IssuerKey, client.ClientID))
			require.NoError(t, token.Set(joseJwt.SubjectKey, client.ClientID))
			require.NoError(t, token.Set(joseJwt.AudienceKey, []string{"https://auth.example.com/token"}))
			require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
			require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(tc.lifetime)))
			require.NoError(t, token.Set(joseJwt.JwtIDKey, googleUuid.NewString()))

			err := validator.validateClaims(context.Background(), token, client)

			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errSubstr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestClientSecretJWTValidator_MissingJTI tests validation fails when jti claim is missing.
func TestClientSecretJWTValidator_MissingJTI(t *testing.T) {
	t.Parallel()

	validator := NewClientSecretJWTValidator("https://auth.example.com/token", nil)

	clientID := googleUuid.New()
	client := &cryptoutilIdentityDomain.Client{
		ID:           clientID,
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	}

	now := time.Now().UTC()
	token := joseJwt.New()
	require.NoError(t, token.Set(joseJwt.IssuerKey, client.ClientID))
	require.NoError(t, token.Set(joseJwt.SubjectKey, client.ClientID))
	require.NoError(t, token.Set(joseJwt.AudienceKey, []string{"https://auth.example.com/token"}))
	require.NoError(t, token.Set(joseJwt.IssuedAtKey, now))
	require.NoError(t, token.Set(joseJwt.ExpirationKey, now.Add(5*time.Minute)))
	// No jti claim set!

	err := validator.validateClaims(context.Background(), token, client)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing jti")
}
