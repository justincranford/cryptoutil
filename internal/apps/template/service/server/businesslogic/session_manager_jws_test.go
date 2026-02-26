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

func TestSessionManager_JWS_Issue_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		isBrowser bool
		jwsAlg    string
	}{
		{"Browser_RS256", true, cryptoutilSharedMagic.SessionJWSAlgorithmRS256},
		{"Browser_ES256", true, cryptoutilSharedMagic.SessionJWSAlgorithmES256},
		{"Browser_EdDSA", true, cryptoutilSharedMagic.SessionJWSAlgorithmEdDSA},
		{"Service_RS256", false, cryptoutilSharedMagic.SessionJWSAlgorithmRS256},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				BrowserSessionExpiration: cryptoutilSharedMagic.HoursPerDay * time.Hour,
				ServiceSessionExpiration: cryptoutilSharedMagic.GitRecentActivityDays * cryptoutilSharedMagic.HoursPerDay * time.Hour,
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
			require.NoError(t, err)

			ctx := context.Background()
			userID := googleUuid.Must(googleUuid.NewV7()).String()
			tenantID := googleUuid.Must(googleUuid.NewV7())
			realmID := googleUuid.Must(googleUuid.NewV7())

			// Issue session
			var token string
			if tt.isBrowser {
				token, err = sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
			} else {
				token, err = sm.IssueServiceSession(ctx, userID, tenantID, realmID)
			}

			require.NoError(t, err)
			require.NotEmpty(t, token)

			// Validate session
			var session any
			if tt.isBrowser {
				session, err = sm.ValidateBrowserSession(ctx, token)
			} else {
				session, err = sm.ValidateServiceSession(ctx, token)
			}

			require.NoError(t, err)
			require.NotNil(t, session)
		})
	}
}

// TestSessionManager_JWE_Issue_Validate tests JWE session issue and validation lifecycle.
func TestSessionManager_JWE_Issue_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		isBrowser bool
		jweAlg    string
	}{
		{"Browser_DirA256GCM", true, cryptoutilSharedMagic.SessionJWEAlgorithmDirA256GCM},
		{"Browser_A256GCMKW", true, cryptoutilSharedMagic.SessionJWEAlgorithmA256GCMKWA256GCM},
		{"Service_DirA256GCM", false, cryptoutilSharedMagic.SessionJWEAlgorithmDirA256GCM},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db := setupTestDB(t)
			config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
				BrowserSessionExpiration: cryptoutilSharedMagic.HoursPerDay * time.Hour,
				ServiceSessionExpiration: cryptoutilSharedMagic.GitRecentActivityDays * cryptoutilSharedMagic.HoursPerDay * time.Hour,
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
			require.NoError(t, err)

			ctx := context.Background()
			userID := googleUuid.Must(googleUuid.NewV7()).String()
			tenantID := googleUuid.Must(googleUuid.NewV7())
			realmID := googleUuid.Must(googleUuid.NewV7())

			// Issue session
			var token string
			if tt.isBrowser {
				token, err = sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
			} else {
				token, err = sm.IssueServiceSession(ctx, userID, tenantID, realmID)
			}

			require.NoError(t, err)
			require.NotEmpty(t, token)

			// Validate session
			var session any
			if tt.isBrowser {
				session, err = sm.ValidateBrowserSession(ctx, token)
			} else {
				session, err = sm.ValidateServiceSession(ctx, token)
			}

			require.NoError(t, err)
			require.NotNil(t, session)
		})
	}
}

// TestSessionManager_JWS_Validation_InvalidToken tests JWS validation error paths.
func TestSessionManager_JWS_Validation_InvalidToken(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWS),
		ServiceSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
		ServiceSessionExpiration:   cryptoutilSharedMagic.GitRecentActivityDays * cryptoutilSharedMagic.HoursPerDay * time.Hour,
		SessionIdleTimeout:         2 * time.Hour,
		SessionCleanupInterval:     time.Hour,
		BrowserSessionJWSAlgorithm: cryptoutilSharedMagic.SessionJWSAlgorithmRS256,
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	ctx := context.Background()

	// Test with completely invalid token
	_, err = sm.ValidateBrowserSession(ctx, "invalid-token")
	require.Error(t, err)

	// Test with malformed JWT (wrong structure)
	_, err = sm.ValidateBrowserSession(ctx, "eyJhbGciOiJSUzI1NiJ9.invalid.signature")
	require.Error(t, err)

	// Test with empty token
	_, err = sm.ValidateBrowserSession(ctx, "")
	require.Error(t, err)
}

// TestSessionManager_JWE_Validation_InvalidToken tests JWE validation error paths.
func TestSessionManager_JWE_Validation_InvalidToken(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWE),
		ServiceSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
		ServiceSessionExpiration:   cryptoutilSharedMagic.GitRecentActivityDays * cryptoutilSharedMagic.HoursPerDay * time.Hour,
		SessionIdleTimeout:         2 * time.Hour,
		SessionCleanupInterval:     time.Hour,
		BrowserSessionJWEAlgorithm: cryptoutilSharedMagic.SessionJWEAlgorithmDirA256GCM,
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	ctx := context.Background()

	// Test with completely invalid token
	_, err = sm.ValidateBrowserSession(ctx, "invalid-token")
	require.Error(t, err)

	// Test with empty token
	_, err = sm.ValidateBrowserSession(ctx, "")
	require.Error(t, err)
}

// TestSessionManager_JWS_ExpiredSession tests JWS session expiration handling.
func TestSessionManager_JWS_ExpiredSession(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWS),
		ServiceSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
		ServiceSessionExpiration:   cryptoutilSharedMagic.GitRecentActivityDays * cryptoutilSharedMagic.HoursPerDay * time.Hour,
		SessionIdleTimeout:         2 * time.Hour,
		SessionCleanupInterval:     time.Hour,
		BrowserSessionJWSAlgorithm: cryptoutilSharedMagic.SessionJWSAlgorithmRS256,
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	ctx := context.Background()
	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	// Issue session
	token, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
	require.NoError(t, err)

	// Manually expire the session in database
	var session cryptoutilAppsTemplateServiceServerRepository.BrowserSession

	findErr := sm.db.Where("user_id = ?", userID).First(&session).Error
	require.NoError(t, findErr)

	pastTime := time.Now().UTC().Add(-1 * time.Hour)
	updateErr := sm.db.Model(&session).Update("expiration", pastTime).Error
	require.NoError(t, updateErr)

	// Validation should fail because database session is expired
	// (Even though JWT exp claim might still be valid, DB lookup filters by expiration > now)
	_, validateErr := sm.ValidateBrowserSession(ctx, token)
	require.Error(t, validateErr)
}

// TestSessionManager_JWE_ExpiredSession tests JWE session revocation handling (delete session from DB).
func TestSessionManager_JWE_ExpiredSession(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmJWE),
		ServiceSessionAlgorithm:    string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
		ServiceSessionExpiration:   cryptoutilSharedMagic.GitRecentActivityDays * cryptoutilSharedMagic.HoursPerDay * time.Hour,
		SessionIdleTimeout:         2 * time.Hour,
		SessionCleanupInterval:     time.Hour,
		BrowserSessionJWEAlgorithm: cryptoutilSharedMagic.SessionJWEAlgorithmDirA256GCM,
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	ctx := context.Background()
	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	// Issue session
	token, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
	require.NoError(t, err)

	// Validate works initially
	session, validateErr := sm.ValidateBrowserSession(ctx, token)
	require.NoError(t, validateErr)
	require.NotNil(t, session)

	// Revoke session by deleting from database (JWE validates jti lookup)
	deleteErr := sm.db.Where("user_id = ?", userID).Delete(&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}).Error
	require.NoError(t, deleteErr)

	// Validation should fail because session is revoked (not found in DB)
	_, validateErr2 := sm.ValidateBrowserSession(ctx, token)
	require.Error(t, validateErr2)
}

// TestSessionManager_CleanupExpiredSessions tests cleanup of expired and idle sessions.
func TestSessionManager_CleanupExpiredSessions(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:  string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		ServiceSessionAlgorithm:  string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration: cryptoutilSharedMagic.HoursPerDay * time.Hour,
		ServiceSessionExpiration: cryptoutilSharedMagic.GitRecentActivityDays * cryptoutilSharedMagic.HoursPerDay * time.Hour,
		SessionIdleTimeout:       2 * time.Hour,
		SessionCleanupInterval:   time.Hour,
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	ctx := context.Background()

	// Issue sessions
	userID1 := googleUuid.Must(googleUuid.NewV7()).String()
	userID2 := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	token1, err := sm.IssueBrowserSession(ctx, userID1, tenantID, realmID)
	require.NoError(t, err)
	require.NotEmpty(t, token1)

	token2, err := sm.IssueBrowserSession(ctx, userID2, tenantID, realmID)
	require.NoError(t, err)
	require.NotEmpty(t, token2)

	// Manually expire session 1
	expiredTime := time.Now().UTC().Add(-1 * time.Hour)
	updateErr := sm.db.Model(&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}).
		Where("user_id = ?", userID1).
		Update("expiration", expiredTime).
		Error
	require.NoError(t, updateErr)

	// Run cleanup
	cleanupErr := sm.CleanupExpiredSessions(ctx)
	require.NoError(t, cleanupErr)

	// Session 1 should be cleaned up
	var count1 int64
	sm.db.Model(&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}).
		Where("user_id = ?", userID1).
		Count(&count1)
	require.Equal(t, int64(0), count1, "Expired session should be cleaned up")

	// Session 2 should still exist
	var count2 int64
	sm.db.Model(&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}).
		Where("user_id = ?", userID2).
		Count(&count2)
	require.Equal(t, int64(1), count2, "Non-expired session should remain")
}

// TestSessionManager_CleanupIdleSessions tests cleanup of idle sessions.
func TestSessionManager_CleanupIdleSessions(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:  string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		ServiceSessionAlgorithm:  string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		BrowserSessionExpiration: cryptoutilSharedMagic.HoursPerDay * time.Hour,
		ServiceSessionExpiration: cryptoutilSharedMagic.GitRecentActivityDays * cryptoutilSharedMagic.HoursPerDay * time.Hour,
		SessionIdleTimeout:       2 * time.Hour,
		SessionCleanupInterval:   time.Hour,
	}

	sm := NewSessionManager(db, nil, config)
	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	ctx := context.Background()

	// Issue session
	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	token, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Set last_activity to be beyond idle timeout
	idleTime := time.Now().UTC().Add(-3 * time.Hour) // 3 hours ago, exceeds 2 hour timeout
	updateErr := sm.db.Model(&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}).
		Where("user_id = ?", userID).
		Update("last_activity", idleTime).
		Error
	require.NoError(t, updateErr)

	// Run cleanup
	cleanupErr := sm.CleanupExpiredSessions(ctx)
	require.NoError(t, cleanupErr)

	// Session should be cleaned up due to idle timeout
	var count int64
	sm.db.Model(&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}).
		Where("user_id = ?", userID).
		Count(&count)
	require.Equal(t, int64(0), count, "Idle session should be cleaned up")
}

// TestSessionManager_StartCleanupTask_CleansExpiredSessions tests cleanup task removes expired sessions.
