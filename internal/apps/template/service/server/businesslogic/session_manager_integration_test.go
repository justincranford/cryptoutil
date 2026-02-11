// Copyright (c) 2025 Justin Cranford
//
//

package businesslogic

import (
	"context"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestSessionManager_Integration_CrossAlgorithm tests interactions between different session algorithms.
func TestSessionManager_Integration_CrossAlgorithm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		browserAlgorithm   cryptoutilSharedMagic.SessionAlgorithmType
		serviceAlgorithm   cryptoutilSharedMagic.SessionAlgorithmType
		expectedBrowserJWK bool
		expectedServiceJWK bool
	}{
		{
			name:               "OPAQUE_OPAQUE_NoJWKs",
			browserAlgorithm:   cryptoutilSharedMagic.SessionAlgorithmOPAQUE,
			serviceAlgorithm:   cryptoutilSharedMagic.SessionAlgorithmOPAQUE,
			expectedBrowserJWK: false,
			expectedServiceJWK: false,
		},
		{
			name:               "JWS_JWE_BothJWKs",
			browserAlgorithm:   cryptoutilSharedMagic.SessionAlgorithmJWS,
			serviceAlgorithm:   cryptoutilSharedMagic.SessionAlgorithmJWE,
			expectedBrowserJWK: true,
			expectedServiceJWK: true,
		},
		{
			name:               "OPAQUE_JWS_ServiceJWKOnly",
			browserAlgorithm:   cryptoutilSharedMagic.SessionAlgorithmOPAQUE,
			serviceAlgorithm:   cryptoutilSharedMagic.SessionAlgorithmJWS,
			expectedBrowserJWK: false,
			expectedServiceJWK: true,
		},
		{
			name:               "JWE_OPAQUE_BrowserJWKOnly",
			browserAlgorithm:   cryptoutilSharedMagic.SessionAlgorithmJWE,
			serviceAlgorithm:   cryptoutilSharedMagic.SessionAlgorithmOPAQUE,
			expectedBrowserJWK: true,
			expectedServiceJWK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := setupSessionManager(t, tt.browserAlgorithm, tt.serviceAlgorithm)
			ctx := context.Background()

			// Issue browser session
			userID := googleUuid.Must(googleUuid.NewV7()).String()
			tenantID := googleUuid.Must(googleUuid.NewV7())
			realmID := googleUuid.Must(googleUuid.NewV7())
			browserToken, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
			require.NoError(t, err)
			require.NotEmpty(t, browserToken)

			// Issue service session
			clientID := googleUuid.Must(googleUuid.NewV7()).String()
			serviceToken, err := sm.IssueServiceSession(ctx, clientID, tenantID, realmID)
			require.NoError(t, err)
			require.NotEmpty(t, serviceToken)

			// Validate both sessions
			browserSession, err := sm.ValidateBrowserSession(ctx, browserToken)
			require.NoError(t, err)
			require.NotNil(t, browserSession)

			serviceSession, err := sm.ValidateServiceSession(ctx, serviceToken)
			require.NoError(t, err)
			require.NotNil(t, serviceSession)

			// Check JWK initialization state
			if tt.expectedBrowserJWK {
				require.NotNil(t, sm.browserJWKID, "Expected browser JWK to be initialized")
			} else {
				require.Nil(t, sm.browserJWKID, "Expected browser JWK to remain nil")
			}

			if tt.expectedServiceJWK {
				require.NotNil(t, sm.serviceJWKID, "Expected service JWK to be initialized")
			} else {
				require.Nil(t, sm.serviceJWKID, "Expected service JWK to remain nil")
			}
		})
	}
}

// TestSessionManager_Integration_SessionLifecycle tests complete session lifecycle scenarios.
func TestSessionManager_Integration_SessionLifecycle(t *testing.T) {
	t.Parallel()

	algorithms := []struct {
		name             string
		browserAlgorithm cryptoutilSharedMagic.SessionAlgorithmType
		serviceAlgorithm cryptoutilSharedMagic.SessionAlgorithmType
	}{
		{"OPAQUE", cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE},
		{"JWS", cryptoutilSharedMagic.SessionAlgorithmJWS, cryptoutilSharedMagic.SessionAlgorithmJWS},
		{"JWE", cryptoutilSharedMagic.SessionAlgorithmJWE, cryptoutilSharedMagic.SessionAlgorithmJWE},
		{"Mixed", cryptoutilSharedMagic.SessionAlgorithmJWS, cryptoutilSharedMagic.SessionAlgorithmJWE},
	}

	for _, alg := range algorithms {
		t.Run(alg.name, func(t *testing.T) {
			sm := setupSessionManager(t, alg.browserAlgorithm, alg.serviceAlgorithm)
			ctx := context.Background()

			// 1. Issue multiple sessions for same user
			userID := googleUuid.Must(googleUuid.NewV7()).String()
			tenantID := googleUuid.Must(googleUuid.NewV7())
			realmID := googleUuid.Must(googleUuid.NewV7())

			var tokens []string

			for i := 0; i < 3; i++ {
				token, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
				require.NoError(t, err)

				tokens = append(tokens, token)
			}

			// 2. Validate all sessions are active
			for i, token := range tokens {
				session, err := sm.ValidateBrowserSession(ctx, token)
				require.NoError(t, err, "Session %d should be valid", i)
				require.NotNil(t, session)
			}

			// 3. Verify all sessions are still valid (no revocation test since method doesn't exist yet)
			for i, token := range tokens {
				_, err := sm.ValidateBrowserSession(ctx, token)
				require.NoError(t, err, "Session %d should still be valid", i)
			}

			// 4. Test cleanup of expired sessions
			// Create short-lived session for expiration test
			shortSM := setupShortSessionManager(t, alg.browserAlgorithm, alg.serviceAlgorithm)
			shortToken, err := shortSM.IssueBrowserSession(ctx, userID, tenantID, realmID)
			require.NoError(t, err)

			// Wait for expiration
			time.Sleep(100 * time.Millisecond)

			// Verify session is expired
			var validationErr error

			_, validationErr = shortSM.ValidateBrowserSession(ctx, shortToken)
			require.Error(t, validationErr, "Expired session should be invalid")

			// Run cleanup
			cleanupErr := shortSM.CleanupExpiredSessions(ctx)
			require.NoError(t, cleanupErr)
		})
	}
}

// TestSessionManager_Integration_MultiAlgorithmWorkflow tests realistic multi-algorithm workflows.
func TestSessionManager_Integration_MultiAlgorithmWorkflow(t *testing.T) {
	t.Parallel()
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmJWS, cryptoutilSharedMagic.SessionAlgorithmJWE)
	ctx := context.Background()

	// Test realistic workflow: user logs in (browser session) and API client connects (service session)
	userID := googleUuid.Must(googleUuid.NewV7()).String()
	clientID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	browserRealmID := googleUuid.Must(googleUuid.NewV7())
	serviceRealmID := googleUuid.Must(googleUuid.NewV7())

	// Step 1: User logs in via browser
	browserToken, err := sm.IssueBrowserSession(ctx, userID, tenantID, browserRealmID)
	require.NoError(t, err)
	require.NotEmpty(t, browserToken)

	// Step 2: Validate browser session works
	browserSession, err := sm.ValidateBrowserSession(ctx, browserToken)
	require.NoError(t, err)
	require.NotNil(t, browserSession)
	require.Equal(t, userID, *browserSession.UserID)
	require.Equal(t, tenantID, browserSession.TenantID)
	require.Equal(t, browserRealmID, browserSession.RealmID)

	// Step 3: API client authenticates
	serviceToken, err := sm.IssueServiceSession(ctx, clientID, tenantID, serviceRealmID)
	require.NoError(t, err)
	require.NotEmpty(t, serviceToken)

	// Step 4: Validate service session works
	serviceSession, err := sm.ValidateServiceSession(ctx, serviceToken)
	require.NoError(t, err)
	require.NotNil(t, serviceSession)
	require.Equal(t, clientID, *serviceSession.ClientID)
	require.Equal(t, tenantID, serviceSession.TenantID)
	require.Equal(t, serviceRealmID, serviceSession.RealmID)

	// Step 5: Issue multiple sessions for load testing scenario
	var (
		browserTokens []string
		serviceTokens []string
	)

	testRealmID := googleUuid.Must(googleUuid.NewV7())

	for i := 0; i < 10; i++ {
		userID := googleUuid.Must(googleUuid.NewV7()).String()
		clientID := googleUuid.Must(googleUuid.NewV7()).String()

		browserToken, err := sm.IssueBrowserSession(ctx, userID, tenantID, testRealmID)
		require.NoError(t, err)

		browserTokens = append(browserTokens, browserToken)

		serviceToken, err := sm.IssueServiceSession(ctx, clientID, tenantID, testRealmID)
		require.NoError(t, err)

		serviceTokens = append(serviceTokens, serviceToken)
	}

	// Step 6: Validate all sessions are unique and valid
	tokenSet := make(map[string]bool)
	for _, token := range browserTokens {
		require.False(t, tokenSet[token], "Browser tokens should be unique")
		tokenSet[token] = true

		_, err := sm.ValidateBrowserSession(ctx, token)
		require.NoError(t, err, "All browser sessions should be valid")
	}

	for _, token := range serviceTokens {
		require.False(t, tokenSet[token], "Service tokens should be unique")
		tokenSet[token] = true

		_, err := sm.ValidateServiceSession(ctx, token)
		require.NoError(t, err, "All service sessions should be valid")
	}

	// Step 7: Verify JWK infrastructure is working
	require.NotNil(t, sm.browserJWKID, "Browser JWK should be initialized for JWS")
	require.NotNil(t, sm.serviceJWKID, "Service JWK should be initialized for JWE")

	// Step 8: Test error scenarios
	_, err = sm.ValidateBrowserSession(ctx, "invalid-token")
	require.Error(t, err, "Invalid browser token should fail validation")

	_, err = sm.ValidateServiceSession(ctx, "invalid-token")
	require.Error(t, err, "Invalid service token should fail validation")
}

// setupShortSessionManager creates a SessionManager with very short expiration times for testing.
func setupShortSessionManager(t *testing.T, browserAlg, serviceAlg cryptoutilSharedMagic.SessionAlgorithmType) *SessionManager {
	t.Helper()

	db := setupTestDB(t)

	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    string(browserAlg),
		ServiceSessionAlgorithm:    string(serviceAlg),
		BrowserSessionExpiration:   50 * time.Millisecond, // Very short
		ServiceSessionExpiration:   50 * time.Millisecond, // Very short
		SessionIdleTimeout:         25 * time.Millisecond,
		SessionCleanupInterval:     10 * time.Millisecond,
		BrowserSessionJWSAlgorithm: "RS256",
		BrowserSessionJWEAlgorithm: "dir+A256GCM",
		ServiceSessionJWSAlgorithm: "RS256",
		ServiceSessionJWEAlgorithm: "dir+A256GCM",
	}

	// Use nil barrier service for tests (enables plain text JWK storage for testing)
	sm := NewSessionManager(db, nil, config)

	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	return sm
}
