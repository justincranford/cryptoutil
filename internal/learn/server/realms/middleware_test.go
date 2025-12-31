// Copyright (c) 2025 Justin Cranford
//

package realms_test

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"cryptoutil/internal/learn/domain"
	"cryptoutil/internal/learn/server"
	"cryptoutil/internal/learn/server/config"
	"cryptoutil/internal/learn/server/util"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilRandom "cryptoutil/internal/shared/util/random"
)

// Test helpers duplicated from parent server package for subdirectory test isolation.

// initTestDB creates an in-memory SQLite database for testing.
func initTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&domain.User{}, &domain.Message{}, &domain.MessagesRecipientJWK{})
	require.NoError(t, err)

	return db
}

// createTestPublicServer creates a test PublicServer instance.
func createTestPublicServer(t *testing.T, db *gorm.DB) (*server.PublicServer, string) {
	t.Helper()

	cfg := config.DefaultAppConfig()
	cfg.BindPublicPort = 0  // Dynamic port
	cfg.BindPrivatePort = 0 // Dynamic port
	cfg.OTLPService = "learn-im-test"
	cfg.LogLevel = "info"
	cfg.OTLPEndpoint = "grpc://" + cryptoutilSharedMagic.HostnameLocalhost + ":4317"
	cfg.OTLPEnabled = false

	srv, err := server.NewPublicServer(context.Background(), cfg, db)
	require.NoError(t, err)

	go func() {
		_ = srv.Start(context.Background())
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	baseURL := "https://" + cryptoutilSharedMagic.HostnameLocalhost + ":" + srv.GetPublicPort()

	return srv, baseURL
}

// createHTTPClient creates an HTTP client that trusts self-signed certificates.
func createHTTPClient(t *testing.T) *http.Client {
	t.Helper()

	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec // Test environment only.
			},
		},
		Timeout: cryptoutilSharedMagic.LearnDefaultTimeout,
	}
}

// TestJWTMiddleware_InvalidTokens tests various invalid JWT scenarios.
func TestJWTMiddleware_InvalidTokens(t *testing.T) {
	jwtSecret, err := cryptoutilRandom.GeneratePasswordSimple()
	require.NoError(t, err)

	tests := []struct {
		name         string
		setupToken   func(t *testing.T) string
		expectedCode int
	}{
		{
			name: "invalid signing method (none)",
			setupToken: func(t *testing.T) string {
				userID := googleUuid.New()
				claims := &util.Claims{
					UserID:   userID.String(),
					Username: "testuser",
				}
				token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
				tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
				require.NoError(t, err)

				return tokenString
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "malformed user ID in token",
			setupToken: func(t *testing.T) string {
				claims := &util.Claims{
					UserID:   "not-a-uuid",
					Username: "testuser",
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, err := token.SignedString([]byte(jwtSecret))
				require.NoError(t, err)

				return tokenString
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "expired token",
			setupToken: func(t *testing.T) string {
				userID := googleUuid.New()
				expirationTime := time.Now().Add(-1 * time.Hour)
				claims := &util.Claims{
					UserID:   userID.String(),
					Username: "testuser",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(expirationTime),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
						Issuer:    cryptoutilSharedMagic.LearnJWTIssuer,
					},
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, err := token.SignedString([]byte(jwtSecret))
				require.NoError(t, err)

				return tokenString
			},
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := initTestDB(t)
			_, baseURL := createTestPublicServer(t, db)
			client := createHTTPClient(t)

			tokenString := tt.setupToken(t)

			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/service/api/v1/messages/rx", http.NoBody)
			require.NoError(t, err)
			req.Header.Set("Authorization", "Bearer "+tokenString)

			resp, err := client.Do(req)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tt.expectedCode, resp.StatusCode)
		})
	}
}
