// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsCipherImServer "cryptoutil/internal/apps/cipher/im/server"
)

// TestPublicServer_PublicBaseURL tests the PublicBaseURL accessor method.
// This method is a delegation to the underlying PublicServerBase.
func TestPublicServer_PublicBaseURL(t *testing.T) {
	t.Parallel()

	// The test server from TestMain has a PublicServer embedded in the CipherIMServer.
	// We cannot access it directly, but we can test it via the CipherIMServer wrapper.
	publicURL := testCipherIMServer.PublicBaseURL()
	require.NotEmpty(t, publicURL, "Public base URL should not be empty")
	require.Contains(t, publicURL, "https://", "Public base URL should use HTTPS")
	require.Contains(t, publicURL, "127.0.0.1:", "Public base URL should bind to 127.0.0.1")
}

// TestNewPublicServer_NilBase tests that NewPublicServer rejects nil base parameter.
func TestNewPublicServer_NilBase(t *testing.T) {
	t.Parallel()

	_, err := cryptoutilAppsCipherImServer.NewPublicServer(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "public server base cannot be nil")
}

// TestNewPublicServer_NilSessionManager tests that NewPublicServer rejects nil session manager.
func TestNewPublicServer_NilSessionManager(t *testing.T) {
	t.Parallel()

	// Get working base from the test server.
	base := testCipherIMServer.PublicServerBase()

	_, err := cryptoutilAppsCipherImServer.NewPublicServer(base, nil, nil, nil, nil, nil, nil, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "session manager service cannot be nil")
}

// TestNewPublicServer_NilRealmService tests that NewPublicServer rejects nil realm service.
func TestNewPublicServer_NilRealmService(t *testing.T) {
	t.Parallel()

	base := testCipherIMServer.PublicServerBase()
	sessionMgr := testCipherIMServer.SessionManager()

	_, err := cryptoutilAppsCipherImServer.NewPublicServer(base, sessionMgr, nil, nil, nil, nil, nil, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "realm service cannot be nil")
}

// TestNewPublicServer_NilRegistrationService tests that NewPublicServer rejects nil registration service.
func TestNewPublicServer_NilRegistrationService(t *testing.T) {
	t.Parallel()

	base := testCipherIMServer.PublicServerBase()
	sessionMgr := testCipherIMServer.SessionManager()
	realmSvc := testCipherIMServer.RealmService()

	_, err := cryptoutilAppsCipherImServer.NewPublicServer(base, sessionMgr, realmSvc, nil, nil, nil, nil, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "registration service cannot be nil")
}

// TestNewPublicServer_NilUserRepo tests that NewPublicServer rejects nil user repository.
func TestNewPublicServer_NilUserRepo(t *testing.T) {
	t.Parallel()

	base := testCipherIMServer.PublicServerBase()
	sessionMgr := testCipherIMServer.SessionManager()
	realmSvc := testCipherIMServer.RealmService()
	regSvc := testCipherIMServer.RegistrationService()

	_, err := cryptoutilAppsCipherImServer.NewPublicServer(base, sessionMgr, realmSvc, regSvc, nil, nil, nil, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "user repository cannot be nil")
}

// TestNewPublicServer_NilMessageRepo tests that NewPublicServer rejects nil message repository.
func TestNewPublicServer_NilMessageRepo(t *testing.T) {
	t.Parallel()

	base := testCipherIMServer.PublicServerBase()
	sessionMgr := testCipherIMServer.SessionManager()
	realmSvc := testCipherIMServer.RealmService()
	regSvc := testCipherIMServer.RegistrationService()
	userRepo := testCipherIMServer.UserRepo()

	_, err := cryptoutilAppsCipherImServer.NewPublicServer(base, sessionMgr, realmSvc, regSvc, userRepo, nil, nil, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "message repository cannot be nil")
}

// TestNewPublicServer_NilMessageRecipientJWKRepo tests that NewPublicServer rejects nil JWK repository.
func TestNewPublicServer_NilMessageRecipientJWKRepo(t *testing.T) {
	t.Parallel()

	base := testCipherIMServer.PublicServerBase()
	sessionMgr := testCipherIMServer.SessionManager()
	realmSvc := testCipherIMServer.RealmService()
	regSvc := testCipherIMServer.RegistrationService()
	userRepo := testCipherIMServer.UserRepo()
	msgRepo := testCipherIMServer.MessageRepo()

	_, err := cryptoutilAppsCipherImServer.NewPublicServer(base, sessionMgr, realmSvc, regSvc, userRepo, msgRepo, nil, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "message recipient JWK repository cannot be nil")
}

// TestNewPublicServer_NilJWKGenService tests that NewPublicServer rejects nil JWK generation service.
func TestNewPublicServer_NilJWKGenService(t *testing.T) {
	t.Parallel()

	base := testCipherIMServer.PublicServerBase()
	sessionMgr := testCipherIMServer.SessionManager()
	realmSvc := testCipherIMServer.RealmService()
	regSvc := testCipherIMServer.RegistrationService()
	userRepo := testCipherIMServer.UserRepo()
	msgRepo := testCipherIMServer.MessageRepo()
	jwkRepo := testCipherIMServer.MessageRecipientJWKRepo()

	_, err := cryptoutilAppsCipherImServer.NewPublicServer(base, sessionMgr, realmSvc, regSvc, userRepo, msgRepo, jwkRepo, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWK generation service cannot be nil")
}

// TestNewPublicServer_NilBarrierService tests that NewPublicServer rejects nil barrier service.
func TestNewPublicServer_NilBarrierService(t *testing.T) {
	t.Parallel()

	base := testCipherIMServer.PublicServerBase()
	sessionMgr := testCipherIMServer.SessionManager()
	realmSvc := testCipherIMServer.RealmService()
	regSvc := testCipherIMServer.RegistrationService()
	userRepo := testCipherIMServer.UserRepo()
	msgRepo := testCipherIMServer.MessageRepo()
	jwkRepo := testCipherIMServer.MessageRecipientJWKRepo()
	jwkGenSvc := testCipherIMServer.JWKGen()

	_, err := cryptoutilAppsCipherImServer.NewPublicServer(base, sessionMgr, realmSvc, regSvc, userRepo, msgRepo, jwkRepo, jwkGenSvc, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "barrier service cannot be nil")
}
