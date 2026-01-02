// Copyright (c) 2025 Justin Cranford
//

package realms_test

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver

	"cryptoutil/internal/learn/domain"
	"cryptoutil/internal/learn/repository"
	"cryptoutil/internal/learn/server"
	"cryptoutil/internal/learn/server/util"
	cryptoutilUnsealKeysService "cryptoutil/internal/shared/barrier/unsealkeysservice"
	cryptoutilConfig "cryptoutil/internal/shared/config"
	cryptoutilTLSGenerator "cryptoutil/internal/shared/config/tls_generator"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
	cryptoutilRandom "cryptoutil/internal/shared/util/random"
	cryptoutilBarrier "cryptoutil/internal/template/server/barrier"
)

// Test helpers duplicated from parent server package for subdirectory test isolation.

// initTestDB creates an in-memory SQLite database for testing.
func initTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Use unique in-memory database per test to avoid table conflicts.
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", googleUuid.New().String())
	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	require.NoError(t, err)

	// Migrate learn-im domain tables.
	err = db.AutoMigrate(&domain.User{}, &domain.Message{}, &domain.MessageRecipientJWK{})
	require.NoError(t, err)

	// Migrate barrier service tables.
	err = db.AutoMigrate(&cryptoutilBarrier.BarrierRootKey{}, &cryptoutilBarrier.BarrierIntermediateKey{}, &cryptoutilBarrier.BarrierContentKey{})
	require.NoError(t, err)

	return db
}

// createTestPublicServer creates a test PublicServer instance.
func createTestPublicServer(t *testing.T, db *gorm.DB) (*server.PublicServer, string) {
	t.Helper()

	ctx := context.Background()

	// Initialize dependencies.
	telemetrySettings := &cryptoutilConfig.ServerSettings{
		LogLevel:     "info",
		OTLPService:  "learn-im-realms-test",
		OTLPEnabled:  false,
		OTLPEndpoint: "grpc://" + cryptoutilSharedMagic.HostnameLocalhost + ":4317",
	}
	telemetryService, err := cryptoutilTelemetry.NewTelemetryService(ctx, telemetrySettings)
	require.NoError(t, err)

	jwkGenService, err := cryptoutilJose.NewJWKGenService(ctx, telemetryService, false)
	require.NoError(t, err)

	// Generate unseal JWK for testing.
	_, unsealJWK, _, _, _, err := jwkGenService.GenerateJWEJWK(&cryptoutilJose.EncA256GCM, &cryptoutilJose.AlgA256KW)
	require.NoError(t, err)

	unsealService, err := cryptoutilUnsealKeysService.NewUnsealKeysServiceSimple([]joseJwk.Key{unsealJWK})
	require.NoError(t, err)

	barrierRepo, err := cryptoutilBarrier.NewGormBarrierRepository(db)
	require.NoError(t, err)

	barrierService, err := cryptoutilBarrier.NewBarrierService(ctx, telemetryService, jwkGenService, barrierRepo, unsealService)
	require.NoError(t, err)

	tlsCfg, err := cryptoutilTLSGenerator.GenerateAutoTLSGeneratedSettings(
		[]string{cryptoutilSharedMagic.HostnameLocalhost},
		[]string{cryptoutilSharedMagic.IPv4Loopback},
		cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year,
	)
	require.NoError(t, err)

	// Initialize repositories.
	userRepo := repository.NewUserRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(db, barrierService)

	jwtSecret := "test-secret-key-minimum-32-chars-required-for-hs256"

	srv, err := server.NewPublicServer(ctx, 0, userRepo, messageRepo, messageRecipientJWKRepo, jwkGenService, jwtSecret, tlsCfg)
	require.NoError(t, err)

	go func() {
		_ = srv.Start(context.Background())
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	actualPort := srv.ActualPort()
	baseURL := fmt.Sprintf("https://%s:%d", cryptoutilSharedMagic.HostnameLocalhost, actualPort)

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
