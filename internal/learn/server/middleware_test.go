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

	"cryptoutil/internal/learn/magic"
	"cryptoutil/internal/learn/server"
)

// TestJWTMiddleware_InvalidSigningMethod tests JWT with invalid signing method.
func TestJWTMiddleware_InvalidSigningMethod(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	userID := googleUuid.New()
	claims := &server.Claims{
		UserID:   userID.String(),
		Username: "testuser",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/service/api/v1/messages/rx", http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestJWTMiddleware_InvalidUserIDInToken tests JWT with malformed user ID.
func TestJWTMiddleware_InvalidUserIDInToken(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	claims := &server.Claims{
		UserID:   "not-a-uuid",
		Username: "testuser",
	}

	const jwtSecret = "learn-im-dev-secret-change-in-production"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/service/api/v1/messages/rx", http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestJWTMiddleware_ExpiredToken tests JWT with expired token.
func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	t.Parallel()

	db := initTestDB(t)
	_, baseURL := createTestPublicServer(t, db)
	client := createHTTPClient(t)

	userID := googleUuid.New()
	expirationTime := time.Now().Add(-1 * time.Hour)

	claims := &server.Claims{
		UserID:   userID.String(),
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    magic.JWTIssuer,
		},
	}

	const jwtSecret = "learn-im-dev-secret-change-in-production"

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, baseURL+"/service/api/v1/messages/rx", http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
