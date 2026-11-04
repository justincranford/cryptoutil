// Copyright (c) 2025-2026 Justin Cranford.
//
//

package orm

import (
	"context"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	cryptoutilTestHelpDB "cryptoutil/internal/apps-framework/service/test_help_db"
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

	db := cryptoutilTestHelpDB.NewInMemorySQLiteDB(t)
	db.Logger = logger.Default.LogMode(logger.Silent)

	// Get underlying sql.DB for connection pool configuration.
	gormDB, err := db.DB()
	require.NoError(t, err)

	// Configure connection pool for GORM transaction pattern.
	gormDB.SetMaxOpenConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries) // Allows transaction + operations concurrently.
	gormDB.SetMaxIdleConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
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
		Email:             testTestExampleCom,
		EmailVerified:     true,
		PasswordHash:      "dummy-hash",
		Enabled:           true,
	}

	require.NoError(t, db.WithContext(ctx).Create(user).Error)

	return user
}
