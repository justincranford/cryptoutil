// Copyright (c) 2025 Justin Cranford
//
//

package businesslogic

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Create unique database name to avoid sharing between tests
	dbName := fmt.Sprintf("file:test_%s.db?mode=memory&cache=private", strings.ReplaceAll(t.Name(), "/", "_"))
	sqlDB, err := sql.Open("sqlite", dbName)
	require.NoError(t, err)

	// Enable WAL mode for better concurrency
	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	// Set busy timeout for concurrent writes
	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	// Pass to GORM with auto-transactions disabled
	dialector := sqlite.Dialector{Conn: sqlDB}
	db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	// Configure connection pool for GORM transactions
	sqlDB, err = db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(5) // Required for GORM transactions
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(0) // In-memory: never close

	// Auto-migrate session tables
	err = db.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSession{},
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK{},
	)
	require.NoError(t, err)

	// Verify tables were created
	var tableCount int

	err = db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('browser_sessions', 'service_sessions', 'browser_session_jwks', 'service_session_jwks')").Scan(&tableCount).Error
	require.NoError(t, err)
	require.Equal(t, 4, tableCount, "All 4 session tables should be created")

	return db
}

// setupSessionManager creates a SessionManager with test configuration.
func setupSessionManager(t *testing.T, browserAlg, serviceAlg cryptoutilMagic.SessionAlgorithmType) *SessionManager {
	t.Helper()

	db := setupTestDB(t)

	config := &cryptoutilConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    string(browserAlg),
		ServiceSessionAlgorithm:    string(serviceAlg),
		BrowserSessionExpiration:   24 * time.Hour,
		ServiceSessionExpiration:   7 * 24 * time.Hour,
		SessionIdleTimeout:         2 * time.Hour,
		SessionCleanupInterval:     time.Hour,
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

func TestSessionManager_NewSessionManager(t *testing.T) {
	db := setupTestDB(t)
	config := &cryptoutilConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm: string(cryptoutilMagic.SessionAlgorithmOPAQUE),
		ServiceSessionAlgorithm: string(cryptoutilMagic.SessionAlgorithmJWS),
	}

	// Use nil barrier service for tests (enables plain text JWK storage for testing)
	sm := NewSessionManager(db, nil, config)
	require.NotNil(t, sm)
	require.Equal(t, db, sm.db)
	require.Equal(t, config, sm.config)
}

func TestSessionManager_Initialize_OPAQUE(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)

	require.Equal(t, cryptoutilMagic.SessionAlgorithmOPAQUE, sm.browserAlgorithm)
	require.Equal(t, cryptoutilMagic.SessionAlgorithmOPAQUE, sm.serviceAlgorithm)
}

func TestSessionManager_IssueBrowserSession_OPAQUE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	token, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Verify token is valid UUIDv7
	_, parseErr := googleUuid.Parse(token)
	require.NoError(t, parseErr, "Token should be valid UUID")

	// Verify session stored in database
	var session cryptoutilAppsTemplateServiceServerRepository.BrowserSession

	findErr := sm.db.Where("user_id = ?", userID).First(&session).Error
	require.NoError(t, findErr)
	require.NotNil(t, session.TokenHash)
	require.NotEmpty(t, *session.TokenHash)
	require.Equal(t, tenantID, session.TenantID)
	require.Equal(t, realmID, session.RealmID)
	require.NotNil(t, session.UserID)
	require.Equal(t, userID, *session.UserID)
}

func TestSessionManager_ValidateBrowserSession_OPAQUE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	// Issue session
	token, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
	require.NoError(t, err)

	// Validate session
	session, validateErr := sm.ValidateBrowserSession(ctx, token)
	require.NoError(t, validateErr)
	require.NotNil(t, session)
	require.NotNil(t, session.UserID)
	require.Equal(t, userID, *session.UserID)
	require.Equal(t, tenantID, session.TenantID)
	require.Equal(t, realmID, session.RealmID)
}

func TestSessionManager_ValidateBrowserSession_OPAQUE_InvalidToken(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Validate with non-existent token
	invalidToken := googleUuid.Must(googleUuid.NewV7()).String()
	session, err := sm.ValidateBrowserSession(ctx, invalidToken)
	require.Error(t, err)
	require.Nil(t, session)
}

func TestSessionManager_ValidateBrowserSession_OPAQUE_ExpiredSession(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	// Issue session
	token, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
	require.NoError(t, err)

	// Manually expire the session by updating database
	var session cryptoutilAppsTemplateServiceServerRepository.BrowserSession

	findErr := sm.db.Where("user_id = ?", userID).First(&session).Error
	require.NoError(t, findErr)

	pastTime := time.Now().Add(-1 * time.Hour)
	updateErr := sm.db.Model(&session).Update("expiration", pastTime).Error
	require.NoError(t, updateErr)

	// Validate should fail
	validatedSession, validateErr := sm.ValidateBrowserSession(ctx, token)
	require.Error(t, validateErr)
	require.Nil(t, validatedSession)
}

func TestSessionManager_IssueServiceSession_OPAQUE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	token, err := sm.IssueServiceSession(ctx, clientID, tenantID, realmID)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Verify token is valid UUIDv7
	_, parseErr := googleUuid.Parse(token)
	require.NoError(t, parseErr)

	// Verify session stored in database
	var session cryptoutilAppsTemplateServiceServerRepository.ServiceSession

	findErr := sm.db.Where("client_id = ?", clientID).First(&session).Error
	require.NoError(t, findErr)
	require.NotNil(t, session.TokenHash)
	require.NotEmpty(t, *session.TokenHash)
	require.Equal(t, tenantID, session.TenantID)
	require.Equal(t, realmID, session.RealmID)
	require.NotNil(t, session.ClientID)
	require.Equal(t, clientID, *session.ClientID)
}

func TestSessionManager_ValidateServiceSession_OPAQUE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())

	// Issue session
	token, err := sm.IssueServiceSession(ctx, clientID, tenantID, realmID)
	require.NoError(t, err)

	// Validate session
	session, validateErr := sm.ValidateServiceSession(ctx, token)
	require.NoError(t, validateErr)
	require.NotNil(t, session)
	require.NotNil(t, session.ClientID)
	require.Equal(t, clientID, *session.ClientID)
	require.Equal(t, tenantID, session.TenantID)
	require.Equal(t, realmID, session.RealmID)
}

func TestSessionManager_CleanupExpiredSessions_ExpiredByTime(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Create expired session
	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())
	token, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
	require.NoError(t, err)

	// Manually expire it
	var session cryptoutilAppsTemplateServiceServerRepository.BrowserSession

	findErr := sm.db.Where("user_id = ?", userID).First(&session).Error
	require.NoError(t, findErr)

	pastTime := time.Now().Add(-1 * time.Hour)
	updateErr := sm.db.Model(&session).Update("expiration", pastTime).Error
	require.NoError(t, updateErr)

	// Cleanup should remove it
	cleanupErr := sm.CleanupExpiredSessions(ctx)
	require.NoError(t, cleanupErr)

	// Verify session removed
	var count int64

	sm.db.Model(&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}).Where("user_id = ?", userID).Count(&count)
	require.Equal(t, int64(0), count)

	// Validate should fail
	_, validateErr := sm.ValidateBrowserSession(ctx, token)
	require.Error(t, validateErr)
}

func TestSessionManager_CleanupExpiredSessions_IdleTimeout(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Create session
	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID := googleUuid.Must(googleUuid.NewV7())
	_, err := sm.IssueBrowserSession(ctx, userID, tenantID, realmID)
	require.NoError(t, err)

	// Manually set last_activity to past idle threshold
	var session cryptoutilAppsTemplateServiceServerRepository.BrowserSession

	findErr := sm.db.Where("user_id = ?", userID).First(&session).Error
	require.NoError(t, findErr)

	pastActivity := time.Now().Add(-3 * time.Hour) // Beyond 2h idle timeout
	updateErr := sm.db.Model(&session).Update("last_activity", pastActivity).Error
	require.NoError(t, updateErr)

	// Cleanup should remove it
	cleanupErr := sm.CleanupExpiredSessions(ctx)
	require.NoError(t, cleanupErr)

	// Verify session removed
	var count int64

	sm.db.Model(&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}).Where("user_id = ?", userID).Count(&count)
	require.Equal(t, int64(0), count)
}

func TestSessionManager_MultipleSessionsPerUser(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	tenantID := googleUuid.Must(googleUuid.NewV7())
	realmID1 := googleUuid.Must(googleUuid.NewV7())
	realmID2 := googleUuid.Must(googleUuid.NewV7())

	// Issue multiple sessions for same user
	token1, err1 := sm.IssueBrowserSession(ctx, userID, tenantID, realmID1)
	require.NoError(t, err1)

	token2, err2 := sm.IssueBrowserSession(ctx, userID, tenantID, realmID2)
	require.NoError(t, err2)

	require.NotEqual(t, token1, token2)

	// Both sessions should validate
	session1, validate1 := sm.ValidateBrowserSession(ctx, token1)
	require.NoError(t, validate1)
	require.NotNil(t, session1)

	session2, validate2 := sm.ValidateBrowserSession(ctx, token2)
	require.NoError(t, validate2)
	require.NotNil(t, session2)

	// Verify both in database
	var count int64

	sm.db.Model(&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{}).Where("user_id = ?", userID).Count(&count)
	require.Equal(t, int64(2), count)
}

// TestSessionManager_GenerateSessionJWK tests JWK generation.
func TestSessionManager_GenerateSessionJWK(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)

	// Test browser JWK generation for JWS
	browserJWK, err := sm.generateSessionJWK(true, cryptoutilMagic.SessionAlgorithmJWS)
	require.NoError(t, err)
	require.NotNil(t, browserJWK)

	// Test service JWK generation for JWE
	serviceJWK, err := sm.generateSessionJWK(false, cryptoutilMagic.SessionAlgorithmJWE)
	require.NoError(t, err)
	require.NotNil(t, serviceJWK)

	// Test error case: unsupported algorithm (cast to SessionAlgorithmType)
	_, err = sm.generateSessionJWK(true, cryptoutilMagic.SessionAlgorithmType("invalid-algorithm"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported session algorithm")
}

// TestSessionManager_GenerateJWSKey tests JWS key generation.
func TestSessionManager_GenerateJWSKey(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)

	privateKey, err := sm.generateJWSKey(cryptoutilMagic.SessionJWSAlgorithmRS256)
	require.NoError(t, err)
	require.NotNil(t, privateKey)

	// Test unsupported algorithm
	_, err = sm.generateJWSKey("invalid-algorithm")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported JWS algorithm")
}

// TestSessionManager_GenerateJWSKey_AllAlgorithms tests JWS key generation for all supported algorithms.
func TestSessionManager_GenerateJWSKey_AllAlgorithms(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)

	tests := []struct {
		name      string
		algorithm string
	}{
		{"RS256", cryptoutilMagic.SessionJWSAlgorithmRS256},
		{"RS384", cryptoutilMagic.SessionJWSAlgorithmRS384},
		{"RS512", cryptoutilMagic.SessionJWSAlgorithmRS512},
		{"ES256", cryptoutilMagic.SessionJWSAlgorithmES256},
		{"ES384", cryptoutilMagic.SessionJWSAlgorithmES384},
		{"ES512", cryptoutilMagic.SessionJWSAlgorithmES512},
		{"EdDSA", cryptoutilMagic.SessionJWSAlgorithmEdDSA},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privateKey, err := sm.generateJWSKey(tt.algorithm)
			require.NoError(t, err, "generateJWSKey should succeed for %s", tt.algorithm)
			require.NotNil(t, privateKey, "privateKey should not be nil for %s", tt.algorithm)
		})
	}
}

// TestSessionManager_GenerateJWEKey_AllAlgorithms tests JWE key generation for all supported algorithms.
func TestSessionManager_GenerateJWEKey_AllAlgorithms(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)

	tests := []struct {
		name      string
		algorithm string
	}{
		{"DirA256GCM", cryptoutilMagic.SessionJWEAlgorithmDirA256GCM},
		{"A256GCMKWA256GCM", cryptoutilMagic.SessionJWEAlgorithmA256GCMKWA256GCM},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privateKey, err := sm.generateJWEKey(tt.algorithm)
			require.NoError(t, err, "generateJWEKey should succeed for %s", tt.algorithm)
			require.NotNil(t, privateKey, "privateKey should not be nil for %s", tt.algorithm)
		})
	}
}

// TestSessionManager_GenerateJWEKey tests JWE key generation.
func TestSessionManager_GenerateJWEKey(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)

	privateKey, err := sm.generateJWEKey(cryptoutilMagic.SessionJWEAlgorithmDirA256GCM)
	require.NoError(t, err)
	require.NotNil(t, privateKey)

	// Test unsupported algorithm
	_, err = sm.generateJWEKey("invalid-algorithm")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported JWE algorithm")
}

// TestSessionManager_StartCleanupTask tests the cleanup task startup.
func TestSessionManager_StartCleanupTask(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)

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
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)
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
