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

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// setupTestDB creates an in-memory SQLite database for testing.
// DEPRECATED: Use testDB from TestMain instead.
// Kept for backward compatibility with tests not yet refactored.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Create unique database name to avoid sharing between tests.
	dbName := fmt.Sprintf("file:test_%s.db?mode=memory&cache=private", strings.ReplaceAll(t.Name(), "/", "_"))
	sqlDB, err := sql.Open("sqlite", dbName)
	require.NoError(t, err)

	// Enable WAL mode for better concurrency.
	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	// Set busy timeout for concurrent writes.
	_, err = sqlDB.ExecContext(context.Background(), "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	// Pass to GORM with auto-transactions disabled.
	dialector := sqlite.Dialector{Conn: sqlDB}
	db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	// Configure connection pool for GORM transactions.
	sqlDB, err = db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(5) // Required for GORM transactions
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(0) // In-memory: never close

	// Auto-migrate session tables.
	err = db.AutoMigrate(
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSession{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSession{},
		&cryptoutilAppsTemplateServiceServerRepository.BrowserSessionJWK{},
		&cryptoutilAppsTemplateServiceServerRepository.ServiceSessionJWK{},
	)
	require.NoError(t, err)

	// Verify tables were created.
	var tableCount int

	err = db.Raw("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('browser_sessions', 'service_sessions', 'browser_session_jwks', 'service_session_jwks')").Scan(&tableCount).Error
	require.NoError(t, err)
	require.Equal(t, 4, tableCount, "All 4 session tables should be created")

	return db
}

// setupSessionManager creates a SessionManager with test configuration.
func setupSessionManager(t *testing.T, browserAlg, serviceAlg cryptoutilSharedMagic.SessionAlgorithmType) *SessionManager {
	t.Helper()

	db := setupTestDB(t)

	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
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
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm: string(cryptoutilSharedMagic.SessionAlgorithmOPAQUE),
		ServiceSessionAlgorithm: string(cryptoutilSharedMagic.SessionAlgorithmJWS),
	}

	// Use nil barrier service for tests (enables plain text JWK storage for testing)
	sm := NewSessionManager(db, nil, config)
	require.NotNil(t, sm)
	require.Equal(t, db, sm.db)
	require.Equal(t, config, sm.config)
}

func TestSessionManager_Initialize_OPAQUE(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)

	require.Equal(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, sm.browserAlgorithm)
	require.Equal(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, sm.serviceAlgorithm)
}

// TestSessionManager_Initialize_EmptyAlgorithm_UsesDefaults tests that empty algorithm strings
// fall back to default algorithms (OPAQUE for browser, JWS for service sessions).
func TestSessionManager_Initialize_EmptyAlgorithm_UsesDefaults(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	// Create config with empty algorithm strings - should use defaults.
	// Note: Must provide JWS/JWE algorithm settings since default service algorithm is JWS.
	config := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:    "", // Empty - should default to OPAQUE
		ServiceSessionAlgorithm:    "", // Empty - should default to JWS
		BrowserSessionExpiration:   24 * time.Hour,
		ServiceSessionExpiration:   7 * 24 * time.Hour,
		SessionIdleTimeout:         2 * time.Hour,
		SessionCleanupInterval:     time.Hour,
		BrowserSessionJWSAlgorithm: "RS256", // Needed if browser default changes to JWS
		ServiceSessionJWSAlgorithm: "RS256", // Required since default service algorithm is JWS
	}

	// Use nil barrier service for tests (enables plain text JWK storage for testing).
	sm := NewSessionManager(db, nil, config)

	err := sm.Initialize(context.Background())
	require.NoError(t, err)

	// Verify defaults were applied.
	require.Equal(t, cryptoutilSharedMagic.SessionAlgorithmType(cryptoutilSharedMagic.DefaultBrowserSessionAlgorithm), sm.browserAlgorithm,
		"Empty browser algorithm should default to %s", cryptoutilSharedMagic.DefaultBrowserSessionAlgorithm)
	require.Equal(t, cryptoutilSharedMagic.SessionAlgorithmType(cryptoutilSharedMagic.DefaultServiceSessionAlgorithm), sm.serviceAlgorithm,
		"Empty service algorithm should default to %s", cryptoutilSharedMagic.DefaultServiceSessionAlgorithm)
}

func TestSessionManager_IssueBrowserSession_OPAQUE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
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
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
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
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Validate with non-existent token
	invalidToken := googleUuid.Must(googleUuid.NewV7()).String()
	session, err := sm.ValidateBrowserSession(ctx, invalidToken)
	require.Error(t, err)
	require.Nil(t, session)
}

func TestSessionManager_ValidateBrowserSession_OPAQUE_ExpiredSession(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
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

	pastTime := time.Now().UTC().Add(-1 * time.Hour)
	updateErr := sm.db.Model(&session).Update("expiration", pastTime).Error
	require.NoError(t, updateErr)

	// Validate should fail
	validatedSession, validateErr := sm.ValidateBrowserSession(ctx, token)
	require.Error(t, validateErr)
	require.Nil(t, validatedSession)
}

func TestSessionManager_IssueServiceSession_OPAQUE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
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
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
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
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
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

	pastTime := time.Now().UTC().Add(-1 * time.Hour)
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
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
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

	pastActivity := time.Now().UTC().Add(-3 * time.Hour) // Beyond 2h idle timeout
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
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)
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
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)

	// Test browser JWK generation for JWS
	browserJWK, err := sm.generateSessionJWK(true, cryptoutilSharedMagic.SessionAlgorithmJWS)
	require.NoError(t, err)
	require.NotNil(t, browserJWK)

	// Test service JWK generation for JWE
	serviceJWK, err := sm.generateSessionJWK(false, cryptoutilSharedMagic.SessionAlgorithmJWE)
	require.NoError(t, err)
	require.NotNil(t, serviceJWK)

	// Test error case: unsupported algorithm (cast to SessionAlgorithmType)
	_, err = sm.generateSessionJWK(true, cryptoutilSharedMagic.SessionAlgorithmType("invalid-algorithm"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "unsupported session algorithm")
}

// TestSessionManager_GenerateJWSKey tests JWS key generation.
func TestSessionManager_GenerateJWSKey(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)

	privateKey, err := sm.generateJWSKey(cryptoutilSharedMagic.SessionJWSAlgorithmRS256)
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

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)

	tests := []struct {
		name      string
		algorithm string
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
			privateKey, err := sm.generateJWSKey(tt.algorithm)
			require.NoError(t, err, "generateJWSKey should succeed for %s", tt.algorithm)
			require.NotNil(t, privateKey, "privateKey should not be nil for %s", tt.algorithm)
		})
	}
}

// TestSessionManager_GenerateJWEKey_AllAlgorithms tests JWE key generation for all supported algorithms.
func TestSessionManager_GenerateJWEKey_AllAlgorithms(t *testing.T) {
	t.Parallel()

	sm := setupSessionManager(t, cryptoutilSharedMagic.SessionAlgorithmOPAQUE, cryptoutilSharedMagic.SessionAlgorithmOPAQUE)

	tests := []struct {
		name      string
		algorithm string
	}{
		{"DirA256GCM", cryptoutilSharedMagic.SessionJWEAlgorithmDirA256GCM},
		{"A256GCMKWA256GCM", cryptoutilSharedMagic.SessionJWEAlgorithmA256GCMKWA256GCM},
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
		BrowserSessionExpiration:   24 * time.Hour,
		ServiceSessionExpiration:   7 * 24 * time.Hour,
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
		BrowserSessionExpiration:   24 * time.Hour,
		ServiceSessionExpiration:   7 * 24 * time.Hour,
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
		BrowserSessionExpiration:   24 * time.Hour,
		ServiceSessionExpiration:   7 * 24 * time.Hour,
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
		BrowserSessionExpiration:   24 * time.Hour,
		ServiceSessionExpiration:   7 * 24 * time.Hour,
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
		BrowserSessionExpiration: 24 * time.Hour,
		ServiceSessionExpiration: 7 * 24 * time.Hour,
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
		BrowserSessionExpiration: 24 * time.Hour,
		ServiceSessionExpiration: 7 * 24 * time.Hour,
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
