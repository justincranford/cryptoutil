// Copyright (c) 2025 Justin Cranford
//
//

package businesslogic

import (
	"context"
	"database/sql"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite" // CGO-free SQLite driver

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilRepository "cryptoutil/internal/apps/template/service/server/repository"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Open SQLite with modernc driver (CGO-free)
	dsn := "file::memory:?cache=shared"
	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	// Enable WAL mode for better concurrency
	_, err = sqlDB.Exec("PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	// Set busy timeout for concurrent writes
	_, err = sqlDB.Exec("PRAGMA busy_timeout = 30000;")
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
		&cryptoutilRepository.BrowserSession{},
		&cryptoutilRepository.ServiceSession{},
	)
	require.NoError(t, err)

	return db
}

// setupSessionManager creates a SessionManager with test configuration.
func setupSessionManager(t *testing.T, browserAlg, serviceAlg cryptoutilMagic.SessionAlgorithmType) *SessionManager {
	t.Helper()

	db := setupTestDB(t)

	config := &cryptoutilConfig.ServiceTemplateServerSettings{
		BrowserSessionAlgorithm:  string(browserAlg),
		ServiceSessionAlgorithm:  string(serviceAlg),
		BrowserSessionExpiration: 24 * time.Hour,
		ServiceSessionExpiration: 7 * 24 * time.Hour,
		SessionIdleTimeout:       2 * time.Hour,
		SessionCleanupInterval:   time.Hour,
	}

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
	realm := "test-realm"

	token, err := sm.IssueBrowserSession(ctx, userID, realm)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Verify token is valid UUIDv7
	_, parseErr := googleUuid.Parse(token)
	require.NoError(t, parseErr, "Token should be valid UUID")

	// Verify session stored in database
	var session cryptoutilRepository.BrowserSession
	findErr := sm.db.Where("user_id = ?", userID).First(&session).Error
	require.NoError(t, findErr)
	require.NotNil(t, session.TokenHash)
	require.NotEmpty(t, *session.TokenHash)
	require.NotNil(t, session.Realm)
	require.Equal(t, realm, *session.Realm)
	require.NotNil(t, session.UserID)
	require.Equal(t, userID, *session.UserID)
}

func TestSessionManager_ValidateBrowserSession_OPAQUE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()
	realm := "test-realm"

	// Issue session
	token, err := sm.IssueBrowserSession(ctx, userID, realm)
	require.NoError(t, err)

	// Validate session
	session, validateErr := sm.ValidateBrowserSession(ctx, token)
	require.NoError(t, validateErr)
	require.NotNil(t, session)
	require.NotNil(t, session.UserID)
	require.Equal(t, userID, *session.UserID)
	require.NotNil(t, session.Realm)
	require.Equal(t, realm, *session.Realm)
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
	realm := "test-realm"

	// Issue session
	token, err := sm.IssueBrowserSession(ctx, userID, realm)
	require.NoError(t, err)

	// Manually expire the session by updating database
	var session cryptoutilRepository.BrowserSession
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
	realm := "service-realm"

	token, err := sm.IssueServiceSession(ctx, clientID, realm)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// Verify token is valid UUIDv7
	_, parseErr := googleUuid.Parse(token)
	require.NoError(t, parseErr)

	// Verify session stored in database
	var session cryptoutilRepository.ServiceSession
	findErr := sm.db.Where("client_id = ?", clientID).First(&session).Error
	require.NoError(t, findErr)
	require.NotNil(t, session.TokenHash)
	require.NotEmpty(t, *session.TokenHash)
	require.NotNil(t, session.Realm)
	require.Equal(t, realm, *session.Realm)
	require.NotNil(t, session.ClientID)
	require.Equal(t, clientID, *session.ClientID)
}

func TestSessionManager_ValidateServiceSession_OPAQUE_Success(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	clientID := googleUuid.Must(googleUuid.NewV7()).String()
	realm := "service-realm"

	// Issue session
	token, err := sm.IssueServiceSession(ctx, clientID, realm)
	require.NoError(t, err)

	// Validate session
	session, validateErr := sm.ValidateServiceSession(ctx, token)
	require.NoError(t, validateErr)
	require.NotNil(t, session)
	require.NotNil(t, session.ClientID)
	require.Equal(t, clientID, *session.ClientID)
	require.NotNil(t, session.Realm)
	require.Equal(t, realm, *session.Realm)
}

func TestSessionManager_CleanupExpiredSessions_ExpiredByTime(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	// Create expired session
	userID := googleUuid.Must(googleUuid.NewV7()).String()
	token, err := sm.IssueBrowserSession(ctx, userID, "test-realm")
	require.NoError(t, err)

	// Manually expire it
	var session cryptoutilRepository.BrowserSession
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
	sm.db.Model(&cryptoutilRepository.BrowserSession{}).Where("user_id = ?", userID).Count(&count)
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
	_, err := sm.IssueBrowserSession(ctx, userID, "test-realm")
	require.NoError(t, err)

	// Manually set last_activity to past idle threshold
	var session cryptoutilRepository.BrowserSession
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
	sm.db.Model(&cryptoutilRepository.BrowserSession{}).Where("user_id = ?", userID).Count(&count)
	require.Equal(t, int64(0), count)
}

func TestSessionManager_MultipleSessionsPerUser(t *testing.T) {
	sm := setupSessionManager(t, cryptoutilMagic.SessionAlgorithmOPAQUE, cryptoutilMagic.SessionAlgorithmOPAQUE)
	ctx := context.Background()

	userID := googleUuid.Must(googleUuid.NewV7()).String()

	// Issue multiple sessions for same user
	token1, err1 := sm.IssueBrowserSession(ctx, userID, "realm1")
	require.NoError(t, err1)

	token2, err2 := sm.IssueBrowserSession(ctx, userID, "realm2")
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
	sm.db.Model(&cryptoutilRepository.BrowserSession{}).Where("user_id = ?", userID).Count(&count)
	require.Equal(t, int64(2), count)
}
