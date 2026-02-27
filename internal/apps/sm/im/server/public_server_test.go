// Copyright (c) 2025 Justin Cranford
//

package server_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsSmImServer "cryptoutil/internal/apps/sm/im/server"
)

// TestPublicServer_PublicBaseURL tests the PublicBaseURL accessor method.
// This method is a delegation to the underlying PublicServerBase.
func TestPublicServer_PublicBaseURL(t *testing.T) {
	t.Parallel()

	// The test server from TestMain has a PublicServer embedded in the SmIMServer.
	// We cannot access it directly, but we can test it via the SmIMServer wrapper.
	publicURL := testSmIMServer.PublicBaseURL()
	require.NotEmpty(t, publicURL, "Public base URL should not be empty")
	require.Contains(t, publicURL, "https://", "Public base URL should use HTTPS")
	require.Contains(t, publicURL, "127.0.0.1:", "Public base URL should bind to 127.0.0.1")
}

// TestNewPublicServer_NilBase tests that NewPublicServer rejects nil base parameter.
func TestNewPublicServer_NilBase(t *testing.T) {
	t.Parallel()

	_, err := cryptoutilAppsSmImServer.NewPublicServer(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "public server base cannot be nil")
}

// TestNewPublicServer_NilSessionManager tests that NewPublicServer rejects nil session manager.
func TestNewPublicServer_NilSessionManager(t *testing.T) {
	t.Parallel()

	// Get working base from the test server.
	base := testSmIMServer.PublicServerBase()

	_, err := cryptoutilAppsSmImServer.NewPublicServer(base, nil, nil, nil, nil, nil, nil, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "session manager service cannot be nil")
}

// TestNewPublicServer_NilRealmService tests that NewPublicServer rejects nil realm service.
func TestNewPublicServer_NilRealmService(t *testing.T) {
	t.Parallel()

	base := testSmIMServer.PublicServerBase()
	sessionMgr := testSmIMServer.SessionManager()

	_, err := cryptoutilAppsSmImServer.NewPublicServer(base, sessionMgr, nil, nil, nil, nil, nil, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "realm service cannot be nil")
}

// TestNewPublicServer_NilRegistrationService tests that NewPublicServer rejects nil registration service.
func TestNewPublicServer_NilRegistrationService(t *testing.T) {
	t.Parallel()

	base := testSmIMServer.PublicServerBase()
	sessionMgr := testSmIMServer.SessionManager()
	realmSvc := testSmIMServer.RealmService()

	_, err := cryptoutilAppsSmImServer.NewPublicServer(base, sessionMgr, realmSvc, nil, nil, nil, nil, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "registration service cannot be nil")
}

// TestNewPublicServer_NilUserRepo tests that NewPublicServer rejects nil user repository.
func TestNewPublicServer_NilUserRepo(t *testing.T) {
	t.Parallel()

	base := testSmIMServer.PublicServerBase()
	sessionMgr := testSmIMServer.SessionManager()
	realmSvc := testSmIMServer.RealmService()
	regSvc := testSmIMServer.RegistrationService()

	_, err := cryptoutilAppsSmImServer.NewPublicServer(base, sessionMgr, realmSvc, regSvc, nil, nil, nil, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "user repository cannot be nil")
}

// TestNewPublicServer_NilMessageRepo tests that NewPublicServer rejects nil message repository.
func TestNewPublicServer_NilMessageRepo(t *testing.T) {
	t.Parallel()

	base := testSmIMServer.PublicServerBase()
	sessionMgr := testSmIMServer.SessionManager()
	realmSvc := testSmIMServer.RealmService()
	regSvc := testSmIMServer.RegistrationService()
	userRepo := testSmIMServer.UserRepo()

	_, err := cryptoutilAppsSmImServer.NewPublicServer(base, sessionMgr, realmSvc, regSvc, userRepo, nil, nil, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "message repository cannot be nil")
}

// TestNewPublicServer_NilMessageRecipientJWKRepo tests that NewPublicServer rejects nil JWK repository.
func TestNewPublicServer_NilMessageRecipientJWKRepo(t *testing.T) {
	t.Parallel()

	base := testSmIMServer.PublicServerBase()
	sessionMgr := testSmIMServer.SessionManager()
	realmSvc := testSmIMServer.RealmService()
	regSvc := testSmIMServer.RegistrationService()
	userRepo := testSmIMServer.UserRepo()
	msgRepo := testSmIMServer.MessageRepo()

	_, err := cryptoutilAppsSmImServer.NewPublicServer(base, sessionMgr, realmSvc, regSvc, userRepo, msgRepo, nil, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "message recipient JWK repository cannot be nil")
}

// TestNewPublicServer_NilJWKGenService tests that NewPublicServer rejects nil JWK generation service.
func TestNewPublicServer_NilJWKGenService(t *testing.T) {
	t.Parallel()

	base := testSmIMServer.PublicServerBase()
	sessionMgr := testSmIMServer.SessionManager()
	realmSvc := testSmIMServer.RealmService()
	regSvc := testSmIMServer.RegistrationService()
	userRepo := testSmIMServer.UserRepo()
	msgRepo := testSmIMServer.MessageRepo()
	jwkRepo := testSmIMServer.MessageRecipientJWKRepo()

	_, err := cryptoutilAppsSmImServer.NewPublicServer(base, sessionMgr, realmSvc, regSvc, userRepo, msgRepo, jwkRepo, nil, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWK generation service cannot be nil")
}

// TestNewPublicServer_NilBarrierService tests that NewPublicServer rejects nil barrier service.
func TestNewPublicServer_NilBarrierService(t *testing.T) {
	t.Parallel()

	base := testSmIMServer.PublicServerBase()
	sessionMgr := testSmIMServer.SessionManager()
	realmSvc := testSmIMServer.RealmService()
	regSvc := testSmIMServer.RegistrationService()
	userRepo := testSmIMServer.UserRepo()
	msgRepo := testSmIMServer.MessageRepo()
	jwkRepo := testSmIMServer.MessageRecipientJWKRepo()
	jwkGenSvc := testSmIMServer.JWKGen()

	_, err := cryptoutilAppsSmImServer.NewPublicServer(base, sessionMgr, realmSvc, regSvc, userRepo, msgRepo, jwkRepo, jwkGenSvc, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "barrier service cannot be nil")
}

// TestNewPublicServer_ValidCreationAndPublicBaseURL tests successful PublicServer
// creation and that PublicBaseURL delegates to the base.
func TestNewPublicServer_ValidCreationAndPublicBaseURL(t *testing.T) {
	t.Parallel()

	base := testSmIMServer.PublicServerBase()
	sessionMgr := testSmIMServer.SessionManager()
	realmSvc := testSmIMServer.RealmService()
	regSvc := testSmIMServer.RegistrationService()
	userRepo := testSmIMServer.UserRepo()
	msgRepo := testSmIMServer.MessageRepo()
	jwkRepo := testSmIMServer.MessageRecipientJWKRepo()
	jwkGenSvc := testSmIMServer.JWKGen()
	barrierSvc := testSmIMServer.BarrierService()

	ps, err := cryptoutilAppsSmImServer.NewPublicServer(base, sessionMgr, realmSvc, regSvc, userRepo, msgRepo, jwkRepo, jwkGenSvc, barrierSvc)
	require.NoError(t, err)
	require.NotNil(t, ps)

	// Call PublicBaseURL() to cover the delegation in public_server.go:184.
	url := ps.PublicBaseURL()
	require.NotEmpty(t, url)
	require.Contains(t, url, "https://")
}
