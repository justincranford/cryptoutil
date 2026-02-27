// Copyright (c) 2025 Justin Cranford

package notifications_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityNotifications "cryptoutil/internal/apps/identity/notifications"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Use unique in-memory database per test (no shared cache).
	dsn := cryptoutilSharedMagic.SQLiteMemoryPlaceholder
	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, dsn)
	require.NoError(t, err)

	ctx := context.Background()

	// Apply PRAGMA settings.
	_, err = sqlDB.ExecContext(ctx, "PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.ExecContext(ctx, "PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	// Pass existing connection to GORM.
	dialector := sqlite.Dialector{Conn: sqlDB}

	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Configure connection pool.
	sqlDB.SetMaxOpenConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	sqlDB.SetMaxIdleConns(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	sqlDB.SetConnMaxLifetime(0)
	sqlDB.SetConnMaxIdleTime(0)

	// Migrate schema.
	err = db.AutoMigrate(
		&cryptoutilIdentityDomain.Client{},
		&cryptoutilIdentityDomain.ClientSecretVersion{},
	)
	require.NoError(t, err)

	return db
}

// createTestClient creates a test client with a secret version.
func createTestClient(t *testing.T, db *gorm.DB, name string, expiresAt *time.Time) *cryptoutilIdentityDomain.Client {
	t.Helper()

	ctx := context.Background()

	client := &cryptoutilIdentityDomain.Client{
		ID:            googleUuid.New(),
		ClientID:      googleUuid.NewString(), // Unique OAuth 2.1 client identifier
		Name:          name,
		AllowedScopes: []string{cryptoutilSharedMagic.ScopeRead, cryptoutilSharedMagic.ScopeWrite},
	}

	err := db.WithContext(ctx).Create(client).Error
	require.NoError(t, err)

	// Create secret version.
	version := &cryptoutilIdentityDomain.ClientSecretVersion{
		ClientID:   client.ID,
		Version:    1,
		SecretHash: "fake-hash-for-testing",
		Status:     cryptoutilIdentityDomain.SecretStatusActive,
		ExpiresAt:  expiresAt,
	}

	err = db.WithContext(ctx).Create(version).Error
	require.NoError(t, err)

	return client
}

func TestCheckExpiringSecrets_NoExpiringSecrets(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	// Create client with secret expiring in 30 days (outside all thresholds).
	farFutureExpiration := time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * cryptoutilSharedMagic.HoursPerDay * time.Hour)
	createTestClient(t, db, "test-client", &farFutureExpiration)

	// Check for expiring secrets.
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{cryptoutilSharedMagic.GitRecentActivityDays, 3, 1},
		Channels:   []cryptoutilIdentityNotifications.NotificationChannel{cryptoutilIdentityNotifications.ChannelLog},
	}

	service := cryptoutilIdentityNotifications.NewNotificationService(db, config)
	notificationsSent, err := service.CheckExpiringSecrets(ctx)

	require.NoError(t, err)
	require.Equal(t, 0, notificationsSent, "Should send 0 notifications when no secrets expiring within thresholds")
}

func TestCheckExpiringSecrets_OneExpiringSecret_7DaysThreshold(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	// Create client with secret expiring in 7 days (within 7-day threshold).
	sevenDaysExpiration := time.Now().UTC().Add(cryptoutilSharedMagic.GitRecentActivityDays*cryptoutilSharedMagic.HoursPerDay*time.Hour + cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Minute)
	createTestClient(t, db, "test-client", &sevenDaysExpiration)

	// Check for expiring secrets.
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{cryptoutilSharedMagic.GitRecentActivityDays, 3, 1},
		Channels:   []cryptoutilIdentityNotifications.NotificationChannel{cryptoutilIdentityNotifications.ChannelLog},
	}

	service := cryptoutilIdentityNotifications.NewNotificationService(db, config)
	notificationsSent, err := service.CheckExpiringSecrets(ctx)

	require.NoError(t, err)
	require.Equal(t, 1, notificationsSent, "Should send 1 notification for secret expiring in 7 days")
}

func TestCheckExpiringSecrets_OneExpiringSecret_3DaysThreshold(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	// Create client with secret expiring in 3 days (within 3-day threshold).
	threeDaysExpiration := time.Now().UTC().Add(3*cryptoutilSharedMagic.HoursPerDay*time.Hour + cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Minute)
	createTestClient(t, db, "test-client", &threeDaysExpiration)

	// Check for expiring secrets.
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{cryptoutilSharedMagic.GitRecentActivityDays, 3, 1},
		Channels:   []cryptoutilIdentityNotifications.NotificationChannel{cryptoutilIdentityNotifications.ChannelLog},
	}

	service := cryptoutilIdentityNotifications.NewNotificationService(db, config)
	notificationsSent, err := service.CheckExpiringSecrets(ctx)

	require.NoError(t, err)
	require.Equal(t, 1, notificationsSent, "Should send 1 notification for secret expiring in 3 days")
}

func TestCheckExpiringSecrets_OneExpiringSecret_1DayThreshold(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	// Create client with secret expiring in 1 day (within 1-day threshold).
	oneDayExpiration := time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay*time.Hour + cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Minute)
	createTestClient(t, db, "test-client", &oneDayExpiration)

	// Check for expiring secrets.
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{cryptoutilSharedMagic.GitRecentActivityDays, 3, 1},
		Channels:   []cryptoutilIdentityNotifications.NotificationChannel{cryptoutilIdentityNotifications.ChannelLog},
	}

	service := cryptoutilIdentityNotifications.NewNotificationService(db, config)
	notificationsSent, err := service.CheckExpiringSecrets(ctx)

	require.NoError(t, err)
	require.Equal(t, 1, notificationsSent, "Should send 1 notification for secret expiring in 1 day")
}

func TestCheckExpiringSecrets_MultipleClients(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	// Create 3 clients with different expiration times.
	sevenDaysExpiration := time.Now().UTC().Add(cryptoutilSharedMagic.GitRecentActivityDays*cryptoutilSharedMagic.HoursPerDay*time.Hour + cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Minute)
	threeDaysExpiration := time.Now().UTC().Add(3*cryptoutilSharedMagic.HoursPerDay*time.Hour + cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Minute)
	oneDayExpiration := time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay*time.Hour + cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Minute)

	createTestClient(t, db, "client-1", &sevenDaysExpiration)
	createTestClient(t, db, "client-2", &threeDaysExpiration)
	createTestClient(t, db, "client-3", &oneDayExpiration)

	// Check for expiring secrets.
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{cryptoutilSharedMagic.GitRecentActivityDays, 3, 1},
		Channels:   []cryptoutilIdentityNotifications.NotificationChannel{cryptoutilIdentityNotifications.ChannelLog},
	}

	service := cryptoutilIdentityNotifications.NewNotificationService(db, config)
	notificationsSent, err := service.CheckExpiringSecrets(ctx)

	require.NoError(t, err)
	require.Equal(t, 3, notificationsSent, "Should send 3 notifications (1 per client)")
}

func TestCheckExpiringSecrets_MultipleChannels(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	// Create client with secret expiring in 7 days.
	sevenDaysExpiration := time.Now().UTC().Add(cryptoutilSharedMagic.GitRecentActivityDays*cryptoutilSharedMagic.HoursPerDay*time.Hour + cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Minute)
	createTestClient(t, db, "test-client", &sevenDaysExpiration)

	// Check with multiple channels (log + webhook).
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{cryptoutilSharedMagic.GitRecentActivityDays, 3, 1},
		Channels: []cryptoutilIdentityNotifications.NotificationChannel{
			cryptoutilIdentityNotifications.ChannelLog,
			cryptoutilIdentityNotifications.ChannelWebhook,
		},
		WebhookURL: "https://example.com/webhook",
	}

	service := cryptoutilIdentityNotifications.NewNotificationService(db, config)
	notificationsSent, err := service.CheckExpiringSecrets(ctx)

	require.NoError(t, err)
	require.Equal(t, 2, notificationsSent, "Should send 2 notifications (log + webhook)")
}

func TestCheckExpiringSecrets_DefaultConfig(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	// Create client with secret expiring in 7 days.
	sevenDaysExpiration := time.Now().UTC().Add(cryptoutilSharedMagic.GitRecentActivityDays*cryptoutilSharedMagic.HoursPerDay*time.Hour + cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Minute)
	createTestClient(t, db, "test-client", &sevenDaysExpiration)

	// Check with nil config (uses defaults).
	service := cryptoutilIdentityNotifications.NewNotificationService(db, nil)
	notificationsSent, err := service.CheckExpiringSecrets(ctx)

	require.NoError(t, err)
	require.Equal(t, 1, notificationsSent, "Should send notification using default config")
}

func TestCheckExpiringSecrets_NoExpiration(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	// Create client with secret that never expires (expires_at = nil).
	createTestClient(t, db, "test-client", nil)

	// Check for expiring secrets.
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{cryptoutilSharedMagic.GitRecentActivityDays, 3, 1},
		Channels:   []cryptoutilIdentityNotifications.NotificationChannel{cryptoutilIdentityNotifications.ChannelLog},
	}

	service := cryptoutilIdentityNotifications.NewNotificationService(db, config)
	notificationsSent, err := service.CheckExpiringSecrets(ctx)

	require.NoError(t, err)
	require.Equal(t, 0, notificationsSent, "Should send 0 notifications for secrets with no expiration")
}

func TestCheckExpiringSecrets_AlreadyExpiredSecrets(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	// Create client with already-expired secret.
	pastExpiration := time.Now().UTC().Add(-cryptoutilSharedMagic.HoursPerDay * time.Hour)
	createTestClient(t, db, "test-client", &pastExpiration)

	// Check for expiring secrets.
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{cryptoutilSharedMagic.GitRecentActivityDays, 3, 1},
		Channels:   []cryptoutilIdentityNotifications.NotificationChannel{cryptoutilIdentityNotifications.ChannelLog},
	}

	service := cryptoutilIdentityNotifications.NewNotificationService(db, config)
	notificationsSent, err := service.CheckExpiringSecrets(ctx)

	require.NoError(t, err)
	require.Equal(t, 0, notificationsSent, "Should send 0 notifications for already-expired secrets")
}

func TestCheckExpiringSecrets_RevokedSecrets(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	ctx := context.Background()

	client := &cryptoutilIdentityDomain.Client{
		ID:            googleUuid.New(),
		Name:          "test-client",
		AllowedScopes: []string{cryptoutilSharedMagic.ScopeRead},
	}

	err := db.WithContext(context.Background()).Create(client).Error
	require.NoError(t, err)

	// Create revoked secret version.
	sevenDaysExpiration := time.Now().UTC().Add(cryptoutilSharedMagic.GitRecentActivityDays*cryptoutilSharedMagic.HoursPerDay*time.Hour + cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Minute)
	version := &cryptoutilIdentityDomain.ClientSecretVersion{
		ClientID:   client.ID,
		Version:    1,
		SecretHash: "fake-hash",
		Status:     cryptoutilIdentityDomain.SecretStatusRevoked,
		ExpiresAt:  &sevenDaysExpiration,
	}

	err = db.WithContext(context.Background()).Create(version).Error
	require.NoError(t, err)

	// Check for expiring secrets.
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{cryptoutilSharedMagic.GitRecentActivityDays, 3, 1},
		Channels:   []cryptoutilIdentityNotifications.NotificationChannel{cryptoutilIdentityNotifications.ChannelLog},
	}

	service := cryptoutilIdentityNotifications.NewNotificationService(db, config)
	notificationsSent, err := service.CheckExpiringSecrets(ctx)

	require.NoError(t, err)
	require.Equal(t, 0, notificationsSent, "Should send 0 notifications for revoked secrets")
}

func TestDefaultNotificationConfig(t *testing.T) {
	t.Parallel()

	config := cryptoutilIdentityNotifications.DefaultNotificationConfig()

	require.NotNil(t, config)
	require.Equal(t, []int{cryptoutilSharedMagic.GitRecentActivityDays, 3, 1}, config.Thresholds, "Default thresholds should be 7, 3, 1 days")
	require.Equal(t, []cryptoutilIdentityNotifications.NotificationChannel{cryptoutilIdentityNotifications.ChannelLog}, config.Channels, "Default channel should be log")
	require.Empty(t, config.WebhookURL, "Default webhook URL should be empty")
	require.Empty(t, config.EmailRecipients, "Default email recipients should be empty")
}

func TestNewNotificationService_WithWebhook(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{cryptoutilSharedMagic.GitRecentActivityDays},
		Channels:   []cryptoutilIdentityNotifications.NotificationChannel{cryptoutilIdentityNotifications.ChannelWebhook},
		WebhookURL: "https://example.com/webhook",
	}

	service := cryptoutilIdentityNotifications.NewNotificationService(db, config)
	require.NotNil(t, service)
}

func TestNewNotificationService_WithEmail(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds:      []int{cryptoutilSharedMagic.GitRecentActivityDays},
		Channels:        []cryptoutilIdentityNotifications.NotificationChannel{cryptoutilIdentityNotifications.ChannelEmail},
		EmailRecipients: []string{"admin@example.com"},
	}

	service := cryptoutilIdentityNotifications.NewNotificationService(db, config)
	require.NotNil(t, service)
}

// TestEmailNotifier_Send tests EmailNotifier stub implementation.
func TestEmailNotifier_Send(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		recipients   []string
		notification *cryptoutilIdentityNotifications.ExpirationNotification
		wantErr      bool
		errContains  string
	}{
		{
			name:       "single_recipient",
			recipients: []string{"admin@example.com"},
			notification: &cryptoutilIdentityNotifications.ExpirationNotification{
				ClientID:      googleUuid.New(),
				ClientName:    "Test Client",
				DaysRemaining: cryptoutilSharedMagic.GitRecentActivityDays,
				ExpiresAt:     time.Now().UTC().Add(cryptoutilSharedMagic.GitRecentActivityDays * cryptoutilSharedMagic.HoursPerDay * time.Hour),
			},
			wantErr:     true,
			errContains: "email notifications not yet implemented",
		},
		{
			name:       "multiple_recipients",
			recipients: []string{"admin@example.com", "security@example.com"},
			notification: &cryptoutilIdentityNotifications.ExpirationNotification{
				ClientID:      googleUuid.New(),
				ClientName:    "Critical Service",
				DaysRemaining: 1,
				ExpiresAt:     time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
			},
			wantErr:     true,
			errContains: "email notifications not yet implemented",
		},
		{
			name:       "no_recipients",
			recipients: []string{},
			notification: &cryptoutilIdentityNotifications.ExpirationNotification{
				ClientID:      googleUuid.New(),
				ClientName:    "Orphan Client",
				DaysRemaining: 3,
				ExpiresAt:     time.Now().UTC().Add(3 * cryptoutilSharedMagic.HoursPerDay * time.Hour),
			},
			wantErr:     true,
			errContains: "email notifications not yet implemented",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			notifier := cryptoutilIdentityNotifications.NewEmailNotifier(tc.recipients)
			require.NotNil(t, notifier)

			ctx := context.Background()

			err := notifier.Send(ctx, tc.notification)

			if tc.wantErr {
				require.Error(t, err)

				if tc.errContains != "" {
					require.ErrorContains(t, err, tc.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestNewEmailNotifier tests EmailNotifier constructor.
func TestNewEmailNotifier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		recipients []string
	}{
		{
			name:       "single_recipient",
			recipients: []string{"admin@example.com"},
		},
		{
			name:       "multiple_recipients",
			recipients: []string{"admin@example.com", "security@example.com", "ops@example.com"},
		},
		{
			name:       "no_recipients",
			recipients: []string{},
		},
		{
			name:       "nil_recipients",
			recipients: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			notifier := cryptoutilIdentityNotifications.NewEmailNotifier(tc.recipients)
			require.NotNil(t, notifier, "NewEmailNotifier should never return nil")
		})
	}
}
