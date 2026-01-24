// Copyright (c) 2025 Justin Cranford

package notifications

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// NotificationChannel defines the delivery method for notifications.
type NotificationChannel string

// Notification channel constants for delivery methods.
const (
	// ChannelWebhook delivers notifications via HTTP webhook.
	ChannelWebhook NotificationChannel = "webhook"
	// ChannelEmail delivers notifications via email.
	ChannelEmail NotificationChannel = "email"
	// ChannelLog delivers notifications to application logs.
	ChannelLog NotificationChannel = "log"
)

// NotificationConfig configures the notification service.
type NotificationConfig struct {
	// Thresholds for sending notifications (days before expiration).
	Thresholds []int

	// Channels to use for notification delivery.
	Channels []NotificationChannel

	// WebhookURL for webhook notifications (optional).
	WebhookURL string

	// EmailRecipients for email notifications (optional).
	EmailRecipients []string
}

// DefaultNotificationConfig returns the default notification configuration.
func DefaultNotificationConfig() *NotificationConfig {
	return &NotificationConfig{
		Thresholds:      []int{7, 3, 1}, // 7 days, 3 days, 1 day before expiration
		Channels:        []NotificationChannel{ChannelLog},
		WebhookURL:      "",
		EmailRecipients: []string{},
	}
}

// Notifier defines the interface for notification delivery.
type Notifier interface {
	// Send sends a notification about an expiring secret.
	Send(ctx context.Context, notification *ExpirationNotification) error
}

// ExpirationNotification contains details about an expiring secret.
type ExpirationNotification struct {
	ClientID      googleUuid.UUID
	ClientName    string
	Version       int
	ExpiresAt     time.Time
	DaysRemaining int
	Channel       NotificationChannel
}

// NotificationService manages pre-expiration notifications.
type NotificationService struct {
	db        *gorm.DB
	config    *NotificationConfig
	notifiers map[NotificationChannel]Notifier
}

// NewNotificationService creates a new notification service.
func NewNotificationService(db *gorm.DB, config *NotificationConfig) *NotificationService {
	if config == nil {
		config = DefaultNotificationConfig()
	}

	service := &NotificationService{
		db:        db,
		config:    config,
		notifiers: make(map[NotificationChannel]Notifier),
	}

	// Register default notifiers.
	service.notifiers[ChannelLog] = NewLogNotifier()

	if config.WebhookURL != "" {
		service.notifiers[ChannelWebhook] = NewWebhookNotifier(config.WebhookURL)
	}

	if len(config.EmailRecipients) > 0 {
		service.notifiers[ChannelEmail] = NewEmailNotifier(config.EmailRecipients)
	}

	return service
}

// CheckExpiringSecrets checks for secrets approaching expiration and sends notifications.
func (s *NotificationService) CheckExpiringSecrets(ctx context.Context) (int, error) {
	now := time.Now()
	notificationsSent := 0

	// Check each threshold (7, 3, 1 days).
	for _, threshold := range s.config.Thresholds {
		// Calculate time window for this threshold (e.g., 7 days ± 1 hour).
		expirationStart := now.Add(time.Duration(threshold) * 24 * time.Hour)
		expirationEnd := expirationStart.Add(cryptoutilSharedMagic.SecretRotationCheckInterval)

		// Find active secrets expiring within this threshold window.
		var secrets []cryptoutilIdentityDomain.ClientSecretVersion

		err := s.db.WithContext(ctx).
			Where("status = ? AND expires_at IS NOT NULL AND expires_at >= ? AND expires_at < ?",
				cryptoutilIdentityDomain.SecretStatusActive, expirationStart, expirationEnd).
			Order("expires_at ASC").
			Find(&secrets).Error
		if err != nil {
			return notificationsSent, fmt.Errorf("failed to query expiring secrets: %w", err)
		}

		// Send notifications for each secret.
		for _, secret := range secrets {
			// Get client details.
			var client cryptoutilIdentityDomain.Client

			err := s.db.WithContext(ctx).
				Where("id = ?", secret.ClientID).
				First(&client).Error
			if err != nil {
				return notificationsSent, fmt.Errorf("failed to query client %s: %w", secret.ClientID, err)
			}

			// Create notification.
			notification := &ExpirationNotification{
				ClientID:      client.ID,
				ClientName:    client.Name,
				Version:       secret.Version,
				ExpiresAt:     *secret.ExpiresAt,
				DaysRemaining: threshold,
			}

			// Send via all configured channels.
			for _, channel := range s.config.Channels {
				notification.Channel = channel

				if notifier, exists := s.notifiers[channel]; exists {
					if err := notifier.Send(ctx, notification); err != nil {
						return notificationsSent, fmt.Errorf("failed to send %s notification: %w", channel, err)
					}

					notificationsSent++
				}
			}
		}
	}

	return notificationsSent, nil
}
