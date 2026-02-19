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
	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestSessionManager_StartCleanupTask_CleansExpiredSessions(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:  string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		ServiceSessionAlgorithm:  string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration: 24 * time.Hour,
		ServiceSessionExpiration: 7 * 24 * time.Hour,
		SessionIdleTimeout:       2 * time.Hour,
		SessionCleanupInterval:   100 * time.Millisecond, // Very short for testing
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	// Issue an expired session
	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	token, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Expire the session immediately
	expiredTime := time.Now().UTC().Add(-1 * time.Hour)
	updateErr := sm.db.Model(&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}).
		Where("user_id = ?", userID).
		Update("expiration", expiredTime).
		Error
	require.NoError(t, updateErr)

	// Start cleanup task in background
	go sm.StartCleanupTask(ctx)

	// Wait for at least 2 cleanup cycles
	time.Sleep(300 * time.Millisecond)

	// Cancel context to stop the task
	cancel()

	// Give task time to stop
	time.Sleep(50 * time.Millisecond)

	// Session should have been cleaned up
	var count int64
	sm.db.Model(&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}).
		Where("user_id = ?", userID).
		Count(&count)
	require.Equal(t, int64(0), count, "Expired session should be cleaned up by background task")
}

// TestSessionManager_ServiceSession_JWS_FullCycle tests service session JWS full lifecycle.
func TestSessionManager_ServiceSession_JWS_FullCycle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		jwsAlg string
	}{
		{"RS256", cryptoutilSharedMagic.SessionJWSAlgorithmRS256},
		{"RS384", cryptoutilSharedMagic.SessionJWSAlgorithmRS384},
		{"RS512", cryptoutilSharedMagic.SessionJWSAlgorithmRS512},
		{"ES256", cryptoutilSharedMagic.SessionJWSAlgorithmES256},
		{"ES384", cryptoutilSharedMagic.SessionJWSAlgorithmES384},
		{"ES512", cryptoutilSharedMagic.SessionJWSAlgorithmES512},
		{"EdDSA", cryptoutilSharedMagic.SessionJWSAlgorithmEdDSA},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				BrowserSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
				ServiceSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWS),
				BrowserSessionExpiration:   24 * time.Hour,
				ServiceSessionExpiration:   7 * 24 * time.Hour,
				SessionIdleTimeout:         2 * time.Hour,
				SessionCleanupInterval:     time.Hour,
				ServiceSessionJWSAlgorithm: tt.jwsAlg,
			}

			sm := NewSessionManager(db, nil, config)
			err := sm.Initialize(context.Background())
			require.NoError(t, err)

			ctx := context.Background()
			clientID := googleUuid.Must(googleUuid.NewV7()).String()
			tenantID := googleUuid.Must(googleUuid.NewV7())
			realmID := googleUuid.Must(googleUuid.NewV7())

			// Issue
			token, err := sm.IssueServiceSession(ctx, clientID, tenantID, realmID)
			require.NoError(t, err)
			require.NotEmpty(t, token)

			// Validate
			session, err := sm.ValidateServiceSession(ctx, token)
			require.NoError(t, err)
			require.NotNil(t, session)

			// Validate again (should update last_activity)
			session2, err := sm.ValidateServiceSession(ctx, token)
			require.NoError(t, err)
			require.NotNil(t, session2)
		})
	}
}

// TestSessionManager_ServiceSession_JWE_FullCycle tests service session JWE full lifecycle.
func TestSessionManager_ServiceSession_JWE_FullCycle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		jweAlg string
	}{
		{"DirA256GCM", cryptoutilSharedMagic.SessionJWEAlgorithmDirA256GCM},
		{"A256GCMKW", cryptoutilSharedMagic.SessionJWEAlgorithmA256GCMKWA256GCM},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				BrowserSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
				ServiceSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWE),
				BrowserSessionExpiration:   24 * time.Hour,
				ServiceSessionExpiration:   7 * 24 * time.Hour,
				SessionIdleTimeout:         2 * time.Hour,
				SessionCleanupInterval:     time.Hour,
				ServiceSessionJWEAlgorithm: tt.jweAlg,
			}

			sm := NewSessionManager(db, nil, config)
			err := sm.Initialize(context.Background())
			require.NoError(t, err)

			ctx := context.Background()
			clientID := googleUuid.Must(googleUuid.NewV7()).String()
			tenantID := googleUuid.Must(googleUuid.NewV7())
			realmID := googleUuid.Must(googleUuid.NewV7())

			// Issue
			token, err := sm.IssueServiceSession(ctx, clientID, tenantID, realmID)
			require.NoError(t, err)
			require.NotEmpty(t, token)

			// Validate
			session, err := sm.ValidateServiceSession(ctx, token)
			require.NoError(t, err)
			require.NotNil(t, session)
		})
	}
}

// TestSessionManager_CleanupServiceSessions tests cleanup of expired service sessions.
func TestSessionManager_CleanupServiceSessions(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:  string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		ServiceSessionAlgorithm:  string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration: 24 * time.Hour,
		ServiceSessionExpiration: 7 * 24 * time.Hour,
		SessionIdleTimeout:       2 * time.Hour,
		SessionCleanupInterval:   time.Hour,
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	ctx := context.Background()

	// Issue service session
	clientID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	token, err := sm.IssueServiceSession(ctx, clientID, tenantID, realmID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Manually expire the session
	expiredTime := time.Now().UTC().Add(-1 * time.Hour)
	updateErr := sm.db.Model(&cryptoutilAppsTemplateServiceServerRepository.ServiceSession{}).
		Where("client_id = ?", clientID).
		Update("expiration", expiredTime).
		Error
	require.NoError(t, updateErr)

	// Run cleanup
	cleanupErr := sm.CleanupExpiredSessions(ctx)
	require.NoError(t, cleanupErr)

	// Session should be cleaned up
	var count int64
	sm.db.Model(&cryptoutilAppsTemplateServiceServerRepository.ServiceSession{}).
		Where("client_id = ?", clientID).
		Count(&count)
	require.Equal(t, int64(0), count, "Expired service session should be cleaned up")
}
