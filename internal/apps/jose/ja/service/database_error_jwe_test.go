// Copyright (c) 2025 Justin Cranford
//

package service

import (
	"context"
	"strings"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	_ "modernc.org/sqlite" // CGO-free SQLite driver.
)

func TestJWEService_EncryptDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWEService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.Encrypt(ctx, googleUuid.New(), googleUuid.New(), []byte("test plaintext"))
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWEService_DecryptDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWEService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.Decrypt(ctx, googleUuid.New(), googleUuid.New(), "eyJhbGciOiJSU0EtT0FFUC0yNTYiLCJlbmMiOiJBMjU2R0NNIn0.test.test.test.test")
	require.Error(t, err)
	// Could fail on parse, get elastic JWK, or decrypt.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "parse"),
		"Expected database, not-found, or parse error, got: %v", err)
}

func TestJWEService_EncryptWithKIDDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWEService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.EncryptWithKID(ctx, googleUuid.New(), googleUuid.New(), "test-kid", []byte("test plaintext"))
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

// ====================
// JWS Service Database Error Tests
// ====================

func TestJWSService_SignDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWSService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.Sign(ctx, googleUuid.New(), googleUuid.New(), []byte("test payload"))
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWSService_VerifyDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWSService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.Verify(ctx, googleUuid.New(), googleUuid.New(), "eyJhbGciOiJSUzI1NiJ9.dGVzdA.test")
	require.Error(t, err)
	// Could fail on parse, get elastic JWK, or verify.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "parse"),
		"Expected database, not-found, or parse error, got: %v", err)
}

func TestJWSService_SignWithKIDDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWSService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.SignWithKID(ctx, googleUuid.New(), googleUuid.New(), "test-kid", []byte("test payload"))
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

// ====================
// JWT Service Database Error Tests
// ====================

func TestJWTService_CreateJWTDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWTService(elasticRepo, materialRepo, testBarrierService)

	claims := &JWTClaims{Issuer: "test-issuer"}
	_, err = svc.CreateJWT(ctx, googleUuid.New(), googleUuid.New(), claims)
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWTService_ValidateJWTDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWTService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.ValidateJWT(ctx, googleUuid.New(), googleUuid.New(), "eyJhbGciOiJSUzI1NiJ9.eyJpc3MiOiJ0ZXN0In0.test")
	require.Error(t, err)
	// Could fail on parse, get elastic JWK, or validate.
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found") ||
			strings.Contains(err.Error(), "parse"),
		"Expected database, not-found, or parse error, got: %v", err)
}

func TestJWTService_CreateEncryptedJWTDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWTService(elasticRepo, materialRepo, testBarrierService)

	claims := &JWTClaims{Issuer: "test-issuer"}
	_, err = svc.CreateEncryptedJWT(ctx, googleUuid.New(), googleUuid.New(), googleUuid.New(), claims)
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

// ====================
// JWKS Service Database Error Tests
// ====================

func TestJWKSService_GetJWKSDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWKSService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.GetJWKS(ctx, googleUuid.New())
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWKSService_GetJWKSForElasticKeyDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWKSService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.GetJWKSForElasticKey(ctx, googleUuid.New(), googleUuid.New())
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

func TestJWKSService_GetPublicJWKDatabaseError(t *testing.T) {
	t.Parallel()

	_, elasticRepo, materialRepo, _, _, err := createClosedServiceDependencies()
	require.NoError(t, err)

	ctx := context.Background()
	svc := NewJWKSService(elasticRepo, materialRepo, testBarrierService)

	_, err = svc.GetPublicJWK(ctx, googleUuid.New(), "test-kid")
	require.Error(t, err)
	require.True(t,
		strings.Contains(err.Error(), "failed to") ||
			strings.Contains(err.Error(), "not found"),
		"Expected database or not-found error, got: %v", err)
}

// ============================================================================
// Crypto Error Path Tests
// ============================================================================
// These tests exercise error paths in crypto operations by corrupting data
// in the database. This is the only way to test these paths since the service
// reads data from the repository before performing crypto operations.

// TestJWSService_Sign_CorruptedBase64 tests that Sign returns error when
// material's PrivateJWKJWE contains invalid base64.
