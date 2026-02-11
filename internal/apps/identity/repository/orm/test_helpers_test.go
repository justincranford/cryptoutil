// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite" // Import CGO-free SQLite driver

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
)

// boolPtr converts bool to *bool for struct literals requiring pointer fields.
func boolPtr(b bool) *bool {
	return &b
}

// testDB wraps a GORM database connection for testing.
type testDB struct {
	db *gorm.DB
}

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *testDB {
	t.Helper()

	// Create unique in-memory database per test using UUIDv7.
	dbID, err := googleUuid.NewV7()
	require.NoError(t, err)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbID.String())

	// Open database connection using modernc.org/sqlite (CGO-free).
	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	// Apply SQLite PRAGMA settings for WAL mode and busy timeout.
	if _, err := sqlDB.ExecContext(context.Background(), "PRAGMA journal_mode=WAL;"); err != nil {
		require.FailNowf(t, "failed to enable WAL mode", "%v", err)
	}

	if _, err := sqlDB.ExecContext(context.Background(), "PRAGMA busy_timeout = 30000;"); err != nil {
		require.FailNowf(t, "failed to set busy timeout", "%v", err)
	}

	// Create GORM database with explicit connection.
	dialector := sqlite.Dialector{Conn: sqlDB}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true, // Disable automatic transactions.
	})
	require.NoError(t, err)

	// Get underlying sql.DB for connection pool configuration.
	gormDB, err := db.DB()
	require.NoError(t, err)

	// Configure connection pool for GORM transaction pattern.
	gormDB.SetMaxOpenConns(5) // Allows transaction + operations concurrently.
	gormDB.SetMaxIdleConns(5)
	gormDB.SetConnMaxLifetime(0) // In-memory DB: never close connections.
	gormDB.SetConnMaxIdleTime(0)

	// Auto-migrate test schemas.
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.User{},
		&cryptoutilIdentityDomain.Client{},
		&cryptoutilIdentityDomain.Token{},
		&cryptoutilIdentityDomain.Key{},
		&WebAuthnCredential{},
		&cryptoutilIdentityDomain.AuthFlow{},
		&cryptoutilIdentityDomain.AuthProfile{},
		&cryptoutilIdentityDomain.AuthorizationRequest{},
		&cryptoutilIdentityDomain.Session{},
		&cryptoutilIdentityDomain.ClientProfile{},
		&cryptoutilIdentityDomain.ConsentDecision{},
		&cryptoutilIdentityDomain.MFAFactor{},
		&cryptoutilIdentityDomain.ClientSecretVersion{},
		&cryptoutilIdentityDomain.ClientSecretHistory{},
		&cryptoutilIdentityDomain.KeyRotationEvent{},
	)
	require.NoError(t, err)

	// Cleanup function to close database.
	t.Cleanup(func() {
		_ = sqlDB.Close() //nolint:errcheck // Test cleanup - error not critical for test teardown
	})

	return &testDB{db: db}
}

// seedTestUser creates a test user in the database.
func seedTestUser(ctx context.Context, t *testing.T, db *gorm.DB, userID string) *cryptoutilIdentityDomain.User {
	t.Helper()

	uid, err := googleUuid.Parse(userID)
	require.NoError(t, err)

	user := &cryptoutilIdentityDomain.User{
		ID:                uid,
		Sub:               userID,
		PreferredUsername: "testuser",
		Email:             "test@example.com",
		EmailVerified:     true,
		PasswordHash:      "dummy-hash",
		Enabled:           true,
	}

	require.NoError(t, db.WithContext(ctx).Create(user).Error)

	return user
}
