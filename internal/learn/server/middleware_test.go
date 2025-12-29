// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"cryptoutil/internal/learn/server"
)

// TestJWTMiddleware_InvalidTokens tests various invalid JWT scenarios.
func TestJWTMiddleware_InvalidTokens(t *testing.T) {
	const jwtSecret = "learn-im-dev-secret-change-in-production"

	tests := []struct {
		name         string
		setupToken   func(t *testing.T) string
		expectedCode int
	}{
		{
			name: "invalid signing method (none)",
			setupToken: func(t *testing.T) string {
				userID := googleUuid.New()
				claims := &server.Claims{
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
				claims := &server.Claims{
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
				claims := &server.Claims{
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
