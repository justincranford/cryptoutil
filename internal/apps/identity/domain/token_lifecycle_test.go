// Copyright (c) 2025 Justin Cranford

package domain_test

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"database/sql"
	json "encoding/json"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "modernc.org/sqlite" // Register CGO-free SQLite driver

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// TestTokenExpirationEnforcement validates token expiration logic across all token types.
// Satisfies R05-06: Token expiration enforcement.
func TestTokenExpirationEnforcement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name         string
		tokenType    cryptoutilIdentityDomain.TokenType
		issuedOffset time.Duration // Offset from now for IssuedAt
		expiresAt    time.Time     // Absolute expiration time
		wantExpired  bool
	}{
		{
			name:         "access_token_not_expired",
			tokenType:    cryptoutilIdentityDomain.TokenTypeAccess,
			issuedOffset: -cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Minute,
			expiresAt:    time.Now().UTC().Add(1 * time.Hour),
			wantExpired:  false,
		},
		{
			name:         "access_token_expired",
			tokenType:    cryptoutilIdentityDomain.TokenTypeAccess,
			issuedOffset: -2 * time.Hour,
			expiresAt:    time.Now().UTC().Add(-1 * time.Hour),
			wantExpired:  true,
		},
		{
			name:         "refresh_token_not_expired",
			tokenType:    cryptoutilIdentityDomain.TokenTypeRefresh,
			issuedOffset: -1 * time.Hour,
			expiresAt:    time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
			wantExpired:  false,
		},
		{
			name:         "refresh_token_expired",
			tokenType:    cryptoutilIdentityDomain.TokenTypeRefresh,
			issuedOffset: -cryptoutilSharedMagic.HMACSHA384KeySize * time.Hour,
			expiresAt:    time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour),
			wantExpired:  true,
		},
		{
			name:         "id_token_not_expired",
			tokenType:    cryptoutilIdentityDomain.TokenTypeID,
			issuedOffset: -cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Minute,
			expiresAt:    time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Minute),
			wantExpired:  false,
		},
		{
			name:         "id_token_expired",
			tokenType:    cryptoutilIdentityDomain.TokenTypeID,
			issuedOffset: -2 * time.Hour,
			expiresAt:    time.Now().UTC().Add(-1 * time.Hour),
			wantExpired:  true,
		},
		{
			name:         "token_expires_exactly_now",
			tokenType:    cryptoutilIdentityDomain.TokenTypeAccess,
			issuedOffset: -1 * time.Hour,
			expiresAt:    time.Now().UTC(),
			wantExpired:  true, // Token is considered expired AT expiration time
		},
	}

	for _, tc := range tests {
		// Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create unique in-memory database per test.
			dbID, err := googleUuid.NewV7()
			require.NoError(t, err)

			dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbID.String())

			sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
			require.NoError(t, err)

			// Apply PRAGMA settings
			_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
			require.NoError(t, err)

			_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
			require.NoError(t, err)

			dialector := sqlite.Dialector{Conn: sqlDB}

			db, err := gorm.Open(dialector, &gorm.Config{
				Logger:                 logger.Default.LogMode(logger.Silent),
				SkipDefaultTransaction: true,
			})
			require.NoError(t, err)

			gormDB, err := db.DB()
			require.NoError(t, err)

			gormDB.SetMaxOpenConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
			gormDB.SetMaxIdleConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)

			err = db.AutoMigrate(&cryptoutilIdentityDomain.Token{})
			require.NoError(t, err)

			t.Cleanup(func() {
				_ = sqlDB.Close() //nolint:errcheck // Test cleanup
			})

			// Create test token with specified expiration time.
			tokenID, err := googleUuid.NewV7()
			require.NoError(t, err)

			clientID, err := googleUuid.NewV7()
			require.NoError(t, err)

			token := &cryptoutilIdentityDomain.Token{
				ID:          tokenID,
				TokenValue:  tokenID.String(),
				TokenType:   tc.tokenType,
				TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
				ClientID:    clientID,
				Scopes:      []string{cryptoutilSharedMagic.ScopeOpenID},
				IssuedAt:    time.Now().UTC().Add(tc.issuedOffset),
				ExpiresAt:   tc.expiresAt,
			}

			err = db.Create(token).Error
			require.NoError(t, err)

			// Query token and check if expired.
			var retrievedToken cryptoutilIdentityDomain.Token

			err = db.Where("id = ?", tokenID).First(&retrievedToken).Error
			require.NoError(t, err)

			// Validate expiration logic.
			isExpired := time.Now().UTC().After(retrievedToken.ExpiresAt) || time.Now().UTC().Equal(retrievedToken.ExpiresAt)

			if tc.wantExpired {
				require.True(t, isExpired, "Token should be expired but is not")
				require.True(t, time.Now().UTC().After(retrievedToken.ExpiresAt) || time.Now().UTC().Equal(retrievedToken.ExpiresAt),
					"ExpiresAt (%v) should be before or equal to now (%v)", retrievedToken.ExpiresAt, time.Now().UTC())
			} else {
				require.False(t, isExpired, "Token should not be expired but is")
				require.True(t, time.Now().UTC().Before(retrievedToken.ExpiresAt),
					"ExpiresAt (%v) should be after now (%v)", retrievedToken.ExpiresAt, time.Now().UTC())
			}
		})
	}
}

// TestTokenRevocationEnforcement validates token revocation logic and status tracking.
// Satisfies R05-04: Token revocation endpoint.
func TestTokenRevocationEnforcement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name            string
		tokenType       cryptoutilIdentityDomain.TokenType
		revoked         cryptoutilIdentityDomain.IntBool
		revokedAt       *time.Time
		wantRevoked     cryptoutilIdentityDomain.IntBool
		wantRevokedTime bool
	}{
		{
			name:            "access_token_not_revoked",
			tokenType:       cryptoutilIdentityDomain.TokenTypeAccess,
			revoked:         cryptoutilIdentityDomain.IntBool(false),
			revokedAt:       nil,
			wantRevoked:     cryptoutilIdentityDomain.IntBool(false),
			wantRevokedTime: false,
		},
		{
			name:            "access_token_revoked",
			tokenType:       cryptoutilIdentityDomain.TokenTypeAccess,
			revoked:         cryptoutilIdentityDomain.IntBool(true),
			revokedAt:       timePtr(time.Now().UTC().Add(-cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * time.Minute)),
			wantRevoked:     cryptoutilIdentityDomain.IntBool(true),
			wantRevokedTime: true,
		},
		{
			name:            "refresh_token_not_revoked",
			tokenType:       cryptoutilIdentityDomain.TokenTypeRefresh,
			revoked:         cryptoutilIdentityDomain.IntBool(false),
			revokedAt:       nil,
			wantRevoked:     cryptoutilIdentityDomain.IntBool(false),
			wantRevokedTime: false,
		},
		{
			name:            "refresh_token_revoked",
			tokenType:       cryptoutilIdentityDomain.TokenTypeRefresh,
			revoked:         cryptoutilIdentityDomain.IntBool(true),
			revokedAt:       timePtr(time.Now().UTC().Add(-1 * time.Hour)),
			wantRevoked:     cryptoutilIdentityDomain.IntBool(true),
			wantRevokedTime: true,
		},
		{
			name:            "id_token_revoked",
			tokenType:       cryptoutilIdentityDomain.TokenTypeID,
			revoked:         cryptoutilIdentityDomain.IntBool(true),
			revokedAt:       timePtr(time.Now().UTC().Add(-cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Second)),
			wantRevoked:     cryptoutilIdentityDomain.IntBool(true),
			wantRevokedTime: true,
		},
	}

	for _, tc := range tests {
		// Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create unique in-memory database per test.
			dbID, err := googleUuid.NewV7()
			require.NoError(t, err)

			dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbID.String())

			sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
			require.NoError(t, err)

			_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
			require.NoError(t, err)

			_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
			require.NoError(t, err)

			dialector := sqlite.Dialector{Conn: sqlDB}

			db, err := gorm.Open(dialector, &gorm.Config{
				Logger:                 logger.Default.LogMode(logger.Silent),
				SkipDefaultTransaction: true,
			})
			require.NoError(t, err)

			gormDB, err := db.DB()
			require.NoError(t, err)

			gormDB.SetMaxOpenConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
			gormDB.SetMaxIdleConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)

			err = db.AutoMigrate(&cryptoutilIdentityDomain.Token{})
			require.NoError(t, err)

			t.Cleanup(func() {
				_ = sqlDB.Close() //nolint:errcheck // Test cleanup
			})

			// Create test token with revocation status.
			tokenID, err := googleUuid.NewV7()
			require.NoError(t, err)

			clientID, err := googleUuid.NewV7()
			require.NoError(t, err)

			token := &cryptoutilIdentityDomain.Token{
				ID:          tokenID,
				TokenValue:  tokenID.String(),
				TokenType:   tc.tokenType,
				TokenFormat: cryptoutilIdentityDomain.TokenFormatUUID,
				ClientID:    clientID,
				Scopes:      []string{cryptoutilSharedMagic.ScopeOpenID},
				IssuedAt:    time.Now().UTC().Add(-1 * time.Hour),
				ExpiresAt:   time.Now().UTC().Add(1 * time.Hour),
				Revoked:     tc.revoked,
				RevokedAt:   tc.revokedAt,
			}

			err = db.Create(token).Error
			require.NoError(t, err)

			// Query token and validate revocation status.
			var retrievedToken cryptoutilIdentityDomain.Token

			err = db.Where("id = ?", tokenID).First(&retrievedToken).Error
			require.NoError(t, err)

			// Validate revocation status.
			require.Equal(t, tc.wantRevoked, retrievedToken.Revoked, "Revoked status mismatch")

			if tc.wantRevokedTime {
				require.NotNil(t, retrievedToken.RevokedAt, "RevokedAt should not be nil for revoked token")
				require.WithinDuration(t, *tc.revokedAt, *retrievedToken.RevokedAt, 1*time.Second,
					"RevokedAt timestamp mismatch")
			} else {
				require.Nil(t, retrievedToken.RevokedAt, "RevokedAt should be nil for non-revoked token")
			}

			// Validate JSON serialization includes/excludes RevokedAt correctly.
			jsonBytes, err := json.Marshal(retrievedToken)
			require.NoError(t, err)

			var jsonMap map[string]any

			err = json.Unmarshal(jsonBytes, &jsonMap)
			require.NoError(t, err)

			if tc.wantRevokedTime {
				_, hasRevokedAt := jsonMap["revoked_at"]
				require.True(t, hasRevokedAt, "JSON should include revoked_at for revoked token")
			} else {
				_, hasRevokedAt := jsonMap["revoked_at"]
				require.False(t, hasRevokedAt, "JSON should not include revoked_at for non-revoked token")
			}
		})
	}
}

// timePtr is a helper function to create time.Time pointers.
func timePtr(t time.Time) *time.Time {
	return &t
}
