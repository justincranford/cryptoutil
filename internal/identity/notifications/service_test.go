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

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityNotifications "cryptoutil/internal/identity/notifications"
)

// setupTestDB creates an in-memory SQLite database for testing.
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Use unique in-memory database per test (no shared cache).
	dsn := ":memory:"
	sqlDB, err := sql.Open("sqlite", dsn)
	require.NoError(t, err)

	// Apply PRAGMA settings.
	_, err = sqlDB.Exec("PRAGMA journal_mode=WAL;")
	require.NoError(t, err)

	_, err = sqlDB.Exec("PRAGMA busy_timeout = 30000;")
	require.NoError(t, err)

	// Pass existing connection to GORM.
	dialector := sqlite.Dialector{Conn: sqlDB}

	db, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	// Configure connection pool.
	sqlDB.SetMaxOpenConns(5)
	sqlDB.SetMaxIdleConns(5)
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
		AllowedScopes: []string{"read", "write"},
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
	farFutureExpiration := time.Now().Add(30 * 24 * time.Hour)
	createTestClient(t, db, "test-client", &farFutureExpiration)

	// Check for expiring secrets.
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{7, 3, 1},
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
	sevenDaysExpiration := time.Now().Add(7*24*time.Hour + 30*time.Minute)
	createTestClient(t, db, "test-client", &sevenDaysExpiration)

	// Check for expiring secrets.
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{7, 3, 1},
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
	threeDaysExpiration := time.Now().Add(3*24*time.Hour + 30*time.Minute)
	createTestClient(t, db, "test-client", &threeDaysExpiration)

	// Check for expiring secrets.
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{7, 3, 1},
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
	oneDayExpiration := time.Now().Add(24*time.Hour + 30*time.Minute)
	createTestClient(t, db, "test-client", &oneDayExpiration)

	// Check for expiring secrets.
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{7, 3, 1},
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
	sevenDaysExpiration := time.Now().Add(7*24*time.Hour + 30*time.Minute)
	threeDaysExpiration := time.Now().Add(3*24*time.Hour + 30*time.Minute)
	oneDayExpiration := time.Now().Add(24*time.Hour + 30*time.Minute)

	createTestClient(t, db, "client-1", &sevenDaysExpiration)
	createTestClient(t, db, "client-2", &threeDaysExpiration)
	createTestClient(t, db, "client-3", &oneDayExpiration)

	// Check for expiring secrets.
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{7, 3, 1},
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
	sevenDaysExpiration := time.Now().Add(7*24*time.Hour + 30*time.Minute)
	createTestClient(t, db, "test-client", &sevenDaysExpiration)

	// Check with multiple channels (log + webhook).
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{7, 3, 1},
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
	sevenDaysExpiration := time.Now().Add(7*24*time.Hour + 30*time.Minute)
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
		Thresholds: []int{7, 3, 1},
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
	pastExpiration := time.Now().Add(-24 * time.Hour)
	createTestClient(t, db, "test-client", &pastExpiration)

	// Check for expiring secrets.
	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{7, 3, 1},
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
		AllowedScopes: []string{"read"},
	}

	err := db.WithContext(context.Background()).Create(client).Error
	require.NoError(t, err)

	// Create revoked secret version.
	sevenDaysExpiration := time.Now().Add(7*24*time.Hour + 30*time.Minute)
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
		Thresholds: []int{7, 3, 1},
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
	require.Equal(t, []int{7, 3, 1}, config.Thresholds, "Default thresholds should be 7, 3, 1 days")
	require.Equal(t, []cryptoutilIdentityNotifications.NotificationChannel{cryptoutilIdentityNotifications.ChannelLog}, config.Channels, "Default channel should be log")
	require.Empty(t, config.WebhookURL, "Default webhook URL should be empty")
	require.Empty(t, config.EmailRecipients, "Default email recipients should be empty")
}

func TestNewNotificationService_WithWebhook(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)

	config := &cryptoutilIdentityNotifications.NotificationConfig{
		Thresholds: []int{7},
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
		Thresholds:      []int{7},
		Channels:        []cryptoutilIdentityNotifications.NotificationChannel{cryptoutilIdentityNotifications.ChannelEmail},
		EmailRecipients: []string{"admin@example.com"},
	}

	service := cryptoutilIdentityNotifications.NewNotificationService(db, config)
	require.NotNil(t, service)
}
