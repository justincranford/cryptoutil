// Copyright (c) 2025 Justin Cranford

package notifications_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityNotifications "cryptoutil/internal/apps/identity/notifications"
)

func setupCoverageTestDB(t *testing.T) (*gorm.DB, *sql.DB) {
	t.Helper()

	dsn := cryptoutilSharedMagic.SQLiteMemoryPlaceholder
	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	ctx := context.Background()

	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)
	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	dialector := sqlite.Dialector{Conn: sqlDB}

	db, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(0)
	sqlDB.SetConnMaxIdleTime(0)

	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.Client{},
		&cryptoutilIdentityDomain.ClientSecretVersion{},
	)
	require.NoError(t, err)

	return db, sqlDB
}

func createCoverageTestClient(t *testing.T, db *gorm.DB, expiresAt *time.Time) *cryptoutilIdentityDomain.Client {
	t.Helper()

	ctx := context.Background()

	client := &cryptoutilIdentityDomain.Client{
		ID:            googleUuid.New(),
		ClientID:      googleUuid.NewString(),
		Name:          "coverage-client-" + googleUuid.NewString(),
		AllowedScopes: []string{"read"},
	}

	err := db.WithContext(ctx).Create(client).Error
	require.NoError(t, err)

	version := &cryptoutilIdentityDomain.ClientSecretVersion{
		ClientID:   client.ID,
		Version:    1,
		SecretHash: "fake-hash",
		Status:     cryptoutilIdentityDomain.SecretStatusActive,
		ExpiresAt:  expiresAt,
	}

	err = db.WithContext(ctx).Create(version).Error
	require.NoError(t, err)

	return client
}

// TestCheckExpiringSecrets_NotifierError covers the "failed to send notification" path.
// EmailNotifier always returns an error, so using ChannelEmail with an expiring secret triggers it.
func TestCheckExpiringSecrets_NotifierError(t *testing.T) {
	t.Parallel()

	db, _ := setupCoverageTestDB(t)

	// Secret expiring in 7 days + 30min hits the 7-day threshold window.
	expiresAt := time.Now().UTC().Add(7*24*time.Hour + 30*time.Minute)
	createCoverageTestClient(t, db, &expiresAt)

	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds:      []int{7},
		Channels:        []cryptoutilIdentityNotifications.NotificationChannel{cryptoutilIdentityNotifications.ChannelEmail},
		EmailRecipients: []string{"admin@example.com"},
	}

	service := cryptoutilIdentityNotifications.NewNotificationService(db, config)
	_, err := service.CheckExpiringSecrets(context.Background())
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to send")
}

// TestCheckExpiringSecrets_DBQueryError covers the "failed to query expiring secrets" path.
func TestCheckExpiringSecrets_DBQueryError(t *testing.T) {
	t.Parallel()

	db, sqlDB := setupCoverageTestDB(t)

	// Close the underlying SQL connection to force DB errors.
	require.NoError(t, sqlDB.Close())

	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{7},
		Channels:   []cryptoutilIdentityNotifications.NotificationChannel{cryptoutilIdentityNotifications.ChannelLog},
	}

	service := cryptoutilIdentityNotifications.NewNotificationService(db, config)
	_, err := service.CheckExpiringSecrets(context.Background())
	require.Error(t, err)
	require.ErrorContains(t, err, "failed to query expiring secrets")
}
