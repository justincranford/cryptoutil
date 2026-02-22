// Copyright (c) 2025 Justin Cranford

package businesslogic

import (
	"context"
	"fmt"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerDomain "cryptoutil/internal/apps/template/service/server/domain"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const testUnsupportedAlgorithm = "UNSUPPORTED_ALG"

// =============================================================================
// Mock repo implementations for error injection
// =============================================================================

// errorUserRepo wraps a real UserRepository and overrides Create to return an error.
type errorUserRepo struct {
	cryptoutilAppsTemplateServiceServerRepository.UserRepository
	createErr error
}

func (r *errorUserRepo) Create(_ context.Context, _ *cryptoutilAppsTemplateServiceServerRepository.User) error {
	if r.createErr != nil {
		return r.createErr
	}

	return nil
}

// errorJoinRequestRepo wraps a real TenantJoinRequestRepository with injectable errors.
type errorJoinRequestRepo struct {
	cryptoutilAppsTemplateServiceServerRepository.TenantJoinRequestRepository
	createErr error
	updateErr error
}

func (r *errorJoinRequestRepo) Create(_ context.Context, _ *cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest) error {
	if r.createErr != nil {
		return r.createErr
	}

	return nil
}

func (r *errorJoinRequestRepo) Update(_ context.Context, _ *cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest) error {
	if r.updateErr != nil {
		return r.updateErr
	}

	return nil
}

// =============================================================================
// TenantRegistrationService mock-repo error coverage tests
// =============================================================================

// TestRegisterUserWithTenant_UserCreate_Error verifies that a user creation
// failure is propagated (covers line 81 in tenant_registration_service.go).
func TestRegisterUserWithTenant_UserCreate_Error(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testDB)
	realUserRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testDB)
	joinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testDB)

	mockUserRepo := &errorUserRepo{
		UserRepository: realUserRepo,
		createErr:      fmt.Errorf("simulated user create error"),
	}

	service := NewTenantRegistrationService(testDB, tenantRepo, mockUserRepo, joinRequestRepo)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7())
	username := fmt.Sprintf("erruser_%s", userID.String()[:8])
	email := fmt.Sprintf("err_%s@example.com", userID.String()[:8])

	_, err := service.RegisterUserWithTenant(ctx, userID, username, email, testPasswordHash, "Error Test Tenant", true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create user")
}

// TestRegisterClientWithTenant_JoinRequest_Create_Error verifies that a join
// request creation failure is propagated (covers line 118 in tenant_registration_service.go).
func TestRegisterClientWithTenant_JoinRequest_Create_Error(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testDB)

	mockJoinRequestRepo := &errorJoinRequestRepo{
		TenantJoinRequestRepository: cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testDB),
		createErr:                   fmt.Errorf("simulated join request create error"),
	}

	service := NewTenantRegistrationService(testDB, tenantRepo, userRepo, mockJoinRequestRepo)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7())
	err := service.RegisterClientWithTenant(ctx, clientID, testTenantID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create client join request")
}

// TestAuthorizeJoinRequest_Update_Error verifies that a join request update
// failure is propagated (covers line 157 in tenant_registration_service.go).
func TestAuthorizeJoinRequest_Update_Error(t *testing.T) {
	t.Parallel()

	tenantRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantRepository(testDB)
	userRepo := cryptoutilAppsTemplateServiceServerRepository.NewUserRepository(testDB)

	realJoinRequestRepo := cryptoutilAppsTemplateServiceServerRepository.NewTenantJoinRequestRepository(testDB)
	mockJoinRequestRepo := &errorJoinRequestRepo{
		TenantJoinRequestRepository: realJoinRequestRepo,
		updateErr:                   fmt.Errorf("simulated update error"),
	}

	service := NewTenantRegistrationService(testDB, tenantRepo, userRepo, mockJoinRequestRepo)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7())
	tenantID := testTenantID
	requestID := googleUuid.Must(googleUuid.NewV7())

	joinRequest := &cryptoutilAppsTemplateServiceServerDomain.TenantJoinRequest{
		ID:          requestID,
		ClientID:    &clientID,
		TenantID:    tenantID,
		Status:      cryptoutilAppsTemplateServiceServerDomain.JoinRequestStatusPending,
		RequestedAt: time.Now().UTC(),
	}

	err := realJoinRequestRepo.Create(ctx, joinRequest)
	require.NoError(t, err)

	adminUserID := googleUuid.Must(googleUuid.NewV7())
	err = service.AuthorizeJoinRequest(ctx, requestID, adminUserID, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to update join request")
}

// =============================================================================
// Unsupported algorithm default-branch coverage tests
// =============================================================================

func setupJWSSessionManager(t *testing.T) *SessionManager {
	t.Helper()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWS),
		ServiceSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWS),
		BrowserSessionExpiration:   24 * time.Hour,
		ServiceSessionExpiration:   7 * 24 * time.Hour,
		SessionIdleTimeout:         2 * time.Hour,
		SessionCleanupInterval:     time.Hour,
		BrowserSessionJWSAlgorithm: cryptoutilSharedMagic.SessionJWSAlgorithmRS256,
		ServiceSessionJWSAlgorithm: cryptoutilSharedMagic.SessionJWSAlgorithmRS256,
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	return sm
}

// TestIssueBrowserSession_UnsupportedAlgorithm covers the default branch in
// IssueBrowserSession (line 165 in session_manager_session.go).
func TestIssueBrowserSession_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	sm.browserAlgorithm = testUnsupportedAlgorithm

	ctx := context.Background()
	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	_, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported browser session algorithm")
}

// TestValidateBrowserSession_UnsupportedAlgorithm covers the default branch in
// ValidateBrowserSession (line 193 in session_manager_session.go).
func TestValidateBrowserSession_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	sm.browserAlgorithm = testUnsupportedAlgorithm

	ctx := context.Background()

	_, err := sm.ValidateBrowserSession(ctx, "some-token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported browser session algorithm")
}

// TestIssueServiceSession_UnsupportedAlgorithm covers the default branch in
// IssueServiceSession (line 220 in session_manager_session.go).
func TestIssueServiceSession_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	sm.serviceAlgorithm = testUnsupportedAlgorithm

	ctx := context.Background()
	clientID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	_, err := sm.IssueServiceSession(ctx, clientID, tenantID, realmID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported service session algorithm")
}

// TestValidateServiceSession_UnsupportedAlgorithm covers the default branch in
// ValidateServiceSession (line 240 in session_manager_session.go).
func TestValidateServiceSession_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	sm := setupJWSSessionManager(t)
	sm.serviceAlgorithm = testUnsupportedAlgorithm

	ctx := context.Background()

	_, err := sm.ValidateServiceSession(ctx, "some-token")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported service session algorithm")
}

// TestInitialize_UnsupportedJWSSubAlgorithm covers the default branch in
// generateJWSKey (line 103 in session_manager_session.go).
func TestInitialize_UnsupportedJWSSubAlgorithm(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWS),
		ServiceSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration:   24 * time.Hour,
		ServiceSessionExpiration:   7 * 24 * time.Hour,
		SessionIdleTimeout:         2 * time.Hour,
		SessionCleanupInterval:     time.Hour,
		BrowserSessionJWSAlgorithm: "UNSUPPORTED_JWS_ALG",
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported JWS algorithm")
}

// TestInitialize_UnsupportedJWESubAlgorithm covers the default branch in
// generateJWEKey (line 120 in session_manager_session.go).
func TestInitialize_UnsupportedJWESubAlgorithm(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWE),
		ServiceSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration:   24 * time.Hour,
		ServiceSessionExpiration:   7 * 24 * time.Hour,
		SessionIdleTimeout:         2 * time.Hour,
		SessionCleanupInterval:     time.Hour,
		BrowserSessionJWEAlgorithm: "UNSUPPORTED_JWE_ALG",
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported JWE algorithm")
}

// =============================================================================
// JWS claims validation error coverage tests
// =============================================================================

// signCustomJWSClaims issues a JWS token using the SessionManager's active
// browser JWK with a custom claims payload.
func signCustomJWSClaims(t *testing.T, sm *SessionManager, claimsJSON []byte) string {
	t.Helper()
	require.NotNil(t, sm.browserJWKID)

	ctx := context.Background()

	var browserJWK cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK

	err := sm.db.WithContext(ctx).Where("id = ?", *sm.browserJWKID).First(&browserJWK).Error
	require.NoError(t, err)

	jwkBytes := []byte(browserJWK.EncryptedJWK)
	privateJWK, err := joseJwk.ParseKey(jwkBytes)
	require.NoError(t, err)

	_, signedBytes, err := cryptoutilSharedCryptoJose.SignBytes([]joseJwk.Key{privateJWK}, claimsJSON)
	require.NoError(t, err)

	return string(signedBytes)
}

// TestValidateBrowserSession_JWS_NoExpClaim covers the missing exp claim branch
// (line 248 in session_manager_jws.go).
func TestValidateBrowserSession_JWS_NoExpClaim(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWS),
		ServiceSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration:   24 * time.Hour,
		ServiceSessionExpiration:   7 * 24 * time.Hour,
		SessionIdleTimeout:         2 * time.Hour,
		SessionCleanupInterval:     time.Hour,
		BrowserSessionJWSAlgorithm: cryptoutilSharedMagic.SessionJWSAlgorithmRS256,
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	claimsJSON := []byte(`{"jti":"valid-id","sub":"user123"}`)
	token := signCustomJWSClaims(t, sm, claimsJSON)

	ctx := context.Background()
	_, err = sm.ValidateBrowserSession(ctx, token)
	require.Error(t, err)
}

// TestValidateBrowserSession_JWS_NoJTIClaim covers the missing jti claim branch
// (line 263 in session_manager_jws.go).
func TestValidateBrowserSession_JWS_NoJTIClaim(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWS),
		ServiceSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration:   24 * time.Hour,
		ServiceSessionExpiration:   7 * 24 * time.Hour,
		SessionIdleTimeout:         2 * time.Hour,
		SessionCleanupInterval:     time.Hour,
		BrowserSessionJWSAlgorithm: cryptoutilSharedMagic.SessionJWSAlgorithmRS256,
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	futureExp := time.Now().Add(24 * time.Hour).Unix()
	claimsJSON := []byte(fmt.Sprintf(`{"exp":%d,"sub":"user123"}`, futureExp))
	token := signCustomJWSClaims(t, sm, claimsJSON)

	ctx := context.Background()
	_, err = sm.ValidateBrowserSession(ctx, token)
	require.Error(t, err)
}

// TestValidateBrowserSession_JWS_InvalidJTI covers the invalid jti UUID branch
// (line 270 in session_manager_jws.go).
func TestValidateBrowserSession_JWS_InvalidJTI(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWS),
		ServiceSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration:   24 * time.Hour,
		ServiceSessionExpiration:   7 * 24 * time.Hour,
		SessionIdleTimeout:         2 * time.Hour,
		SessionCleanupInterval:     time.Hour,
		BrowserSessionJWSAlgorithm: cryptoutilSharedMagic.SessionJWSAlgorithmRS256,
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	futureExp := time.Now().Add(24 * time.Hour).Unix()
	claimsJSON := []byte(fmt.Sprintf(`{"exp":%d,"jti":"not-a-valid-uuid"}`, futureExp))
	token := signCustomJWSClaims(t, sm, claimsJSON)

	ctx := context.Background()
	_, err = sm.ValidateBrowserSession(ctx, token)
	require.Error(t, err)
}
