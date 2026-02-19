// Copyright (c) 2025 Justin Cranford
//
//

package businesslogic

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestSessionManager_GenerateJWEKey(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)

	privateKey, err := sm.generateJWEKey(cryptoutilSharedMagic.SessionJWEAlgorithmDirA256GCM)
	require.NoError(t, err)
	require.NotNil(t, privateKey)

	// Test unsupported algorithm
	_, err = sm.generateJWEKey("invalid-algorithm")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported JWE algorithm")
}

// TestSessionManager_Initialize_JWS_AllAlgorithms tests initialization with all JWS algorithm variants.
// This exercises initializeSessionJWK for JWS algorithms via Initialize().
func TestSessionManager_Initialize_JWS_AllAlgorithms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		jwsAlg    string
		isBrowser bool
	}{
		{"Browser_RS256", cryptoutilSharedMagic.SessionJWSAlgorithmRS256, true},
		{"Browser_RS384", cryptoutilSharedMagic.SessionJWSAlgorithmRS384, true},
		{"Browser_RS512", cryptoutilSharedMagic.SessionJWSAlgorithmRS512, true},
		{"Browser_ES256", cryptoutilSharedMagic.SessionJWSAlgorithmES256, true},
		{"Browser_ES384", cryptoutilSharedMagic.SessionJWSAlgorithmES384, true},
		{"Browser_ES512", cryptoutilSharedMagic.SessionJWSAlgorithmES512, true},
		{"Browser_EdDSA", cryptoutilSharedMagic.SessionJWSAlgorithmEdDSA, true},
		// HMAC algorithms (HS256/HS384/HS512) are supported in initializeSessionJWK
		// but constants are not exported; test with raw strings
		{"Browser_HS256", "HS256", true},
		{"Browser_HS384", "HS384", true},
		{"Browser_HS512", "HS512", true},
		{"Service_RS256", cryptoutilSharedMagic.SessionJWSAlgorithmRS256, false},
		{"Service_ES256", cryptoutilSharedMagic.SessionJWSAlgorithmES256, false},
		{"Service_HS256", "HS256", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				BrowserSessionExpiration: 24 * time.Hour,
				ServiceSessionExpiration: 7 * 24 * time.Hour,
				SessionIdleTimeout:       2 * time.Hour,
				SessionCleanupInterval:   time.Hour,
			}

			if tt.isBrowser {
				config.BrowserSessionAlgorithm = string(cryptoutilSharedMagic.SessionAlgorithmJWS)
				config.ServiceSessionAlgorithm = string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
				config.BrowserSessionJWSAlgorithm = tt.jwsAlg
			} else {
				config.BrowserSessionAlgorithm = string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
				config.ServiceSessionAlgorithm = string(cryptoutilSharedMagic.SessionAlgorithmJWS)
				config.ServiceSessionJWSAlgorithm = tt.jwsAlg
			}

			sm := NewSessionManager(db, nil, config)
			err := sm.Initialize(context.Background())
			require.NoError(t, err, "Initialize should succeed for %s", tt.name)

			// Verify JWK was created
			if tt.isBrowser {
				require.NotNil(t, sm.browserJWKID, "browserJWKID should be set for %s", tt.name)
			} else {
				require.NotNil(t, sm.serviceJWKID, "serviceJWKID should be set for %s", tt.name)
			}
		})
	}
}

// TestSessionManager_Initialize_JWE_AllAlgorithms tests initialization with all JWE algorithm variants.
// This exercises initializeSessionJWK for JWE algorithms via Initialize().
func TestSessionManager_Initialize_JWE_AllAlgorithms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		jweAlg    string
		isBrowser bool
	}{
		{"Browser_DirA256GCM", cryptoutilSharedMagic.SessionJWEAlgorithmDirA256GCM, true},
		{"Browser_A256GCMKWA256GCM", cryptoutilSharedMagic.SessionJWEAlgorithmA256GCMKWA256GCM, true},
		{"Service_DirA256GCM", cryptoutilSharedMagic.SessionJWEAlgorithmDirA256GCM, false},
		{"Service_A256GCMKWA256GCM", cryptoutilSharedMagic.SessionJWEAlgorithmA256GCMKWA256GCM, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				BrowserSessionExpiration: 24 * time.Hour,
				ServiceSessionExpiration: 7 * 24 * time.Hour,
				SessionIdleTimeout:       2 * time.Hour,
				SessionCleanupInterval:   time.Hour,
			}

			if tt.isBrowser {
				config.BrowserSessionAlgorithm = string(cryptoutilSharedMagic.SessionAlgorithmJWE)
				config.ServiceSessionAlgorithm = string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
				config.BrowserSessionJWEAlgorithm = tt.jweAlg
			} else {
				config.BrowserSessionAlgorithm = string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
				config.ServiceSessionAlgorithm = string(cryptoutilSharedMagic.SessionAlgorithmJWE)
				config.ServiceSessionJWEAlgorithm = tt.jweAlg
			}

			sm := NewSessionManager(db, nil, config)
			err := sm.Initialize(context.Background())
			require.NoError(t, err, "Initialize should succeed for %s", tt.name)

			// Verify JWK was created
			if tt.isBrowser {
				require.NotNil(t, sm.browserJWKID, "browserJWKID should be set for %s", tt.name)
			} else {
				require.NotNil(t, sm.serviceJWKID, "serviceJWKID should be set for %s", tt.name)
			}
		})
	}
}

// TestSessionManager_Initialize_UnsupportedAlgorithm tests error handling for unsupported algorithms.
func TestSessionManager_Initialize_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		browserAlg    cryptoutilSharedMagic.SessionAlgorithmType
		serviceAlg    cryptoutilSharedMagic.SessionAlgorithmType
		browserJWSAlg string
		browserJWEAlg string
		serviceJWSAlg string
		serviceJWEAlg string
		expectErrPart string
	}{
		{
			name:          "UnsupportedBrowserJWSAlg",
			browserAlg:    cryptoutilSharedMagic.SessionAlgorithmJWS,
			serviceAlg:    cryptoutilSharedMagic.SessionAlgorithmOPAQUE,
			browserJWSAlg: "INVALID_ALG",
			expectErrPart: "unsupported JWS algorithm",
		},
		{
			name:          "UnsupportedBrowserJWEAlg",
			browserAlg:    cryptoutilSharedMagic.SessionAlgorithmJWE,
			serviceAlg:    cryptoutilSharedMagic.SessionAlgorithmOPAQUE,
			browserJWEAlg: "INVALID_ALG",
			expectErrPart: "unsupported JWE algorithm",
		},
		{
			name:          "UnsupportedServiceJWSAlg",
			browserAlg:    cryptoutilSharedMagic.SessionAlgorithmOPAQUE,
			serviceAlg:    cryptoutilSharedMagic.SessionAlgorithmJWS,
			serviceJWSAlg: "INVALID_ALG",
			expectErrPart: "unsupported JWS algorithm",
		},
		{
			name:          "UnsupportedServiceJWEAlg",
			browserAlg:    cryptoutilSharedMagic.SessionAlgorithmOPAQUE,
			serviceAlg:    cryptoutilSharedMagic.SessionAlgorithmJWE,
			serviceJWEAlg: "INVALID_ALG",
			expectErrPart: "unsupported JWE algorithm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				BrowserSessionAlgorithm:    string(tt.browserAlg),
				ServiceSessionAlgorithm:    string(tt.serviceAlg),
				BrowserSessionExpiration:   24 * time.Hour,
				ServiceSessionExpiration:   7 * 24 * time.Hour,
				SessionIdleTimeout:         2 * time.Hour,
				SessionCleanupInterval:     time.Hour,
				BrowserSessionJWSAlgorithm: tt.browserJWSAlg,
				BrowserSessionJWEAlgorithm: tt.browserJWEAlg,
				ServiceSessionJWSAlgorithm: tt.serviceJWSAlg,
				ServiceSessionJWEAlgorithm: tt.serviceJWEAlg,
			}

			sm := NewSessionManager(db, nil, config)
			err := sm.Initialize(context.Background())
			require.Error(t, err, "Initialize should fail for %s", tt.name)
			require.Contains(t, err.Error(), tt.expectErrPart, "Error should contain '%s' for %s", tt.expectErrPart, tt.name)
		})
	}
}

// TestSessionManager_Initialize_ExistingJWK tests reusing existing JWK from database.
func TestSessionManager_Initialize_ExistingJWK(t *testing.T) {
	t.Parallel()

	// Create database and first session manager
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

	// First initialization creates JWK
	sm1 := NewSessionManager(db, nil, config)
	err1 := sm1.Initialize(context.Background())
	require.NoError(t, err1)
	require.NotNil(t, sm1.browserJWKID)
	firstJWKID := *sm1.browserJWKID

	// Second initialization should reuse existing JWK
	sm2 := NewSessionManager(db, nil, config)
	err2 := sm2.Initialize(context.Background())
	require.NoError(t, err2)
	require.NotNil(t, sm2.browserJWKID)
	secondJWKID := *sm2.browserJWKID

	// Both should use the same JWK
	require.Equal(t, firstJWKID, secondJWKID, "Second initialization should reuse existing JWK")
}

// TestSessionManager_StartCleanupTask tests the cleanup task startup.
func TestSessionManager_StartCleanupTask(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)

	// Create a context that we can cancel to stop the cleanup task
	ctx, cancel := context.WithCancel(context.Background())

	// Start cleanup task in a goroutine
	done := make(chan bool, 1)

	go func() {
		sm.StartCleanupTask(ctx)

		done <- true
	}()

	// Let it run for a brief moment
	time.Sleep(10 * time.Millisecond)

	// Cancel context to stop the cleanup task
	cancel()

	// Wait for cleanup task to finish
	select {
	case <-done:
		// Task completed successfully
	case <-time.After(1 * time.Second):
		t.Fatal("Cleanup task did not stop within timeout")
	}
}

// TestSessionManager_ErrorCases tests various error scenarios for better coverage.
func TestSessionManager_ErrorCases(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Test validation with empty token (will fail in hash function)
	_, err := sm.ValidateBrowserSession(ctx, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to hash session token")

	_, err = sm.ValidateServiceSession(ctx, "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to hash session token")

	// Test validation with invalid token format
	_, err = sm.ValidateBrowserSession(ctx, "invalid-token")
	require.Error(t, err) // Should fail validation (either hash failure or record not found)

	_, err = sm.ValidateServiceSession(ctx, "invalid-token")
	require.Error(t, err) // Should fail validation

	// Test with malformed UUID-like token
	_, err = sm.ValidateBrowserSession(ctx, "not-a-valid-uuid-format-that-is-long-enough")
	require.Error(t, err) // Should fail validation

	_, err = sm.ValidateServiceSession(ctx, "not-a-valid-uuid-format-that-is-long-enough")
	require.Error(t, err) // Should fail validation
}

// TestSessionManager_JWS_Issue_Validate tests JWS session issue and validation lifecycle.
