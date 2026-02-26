// Copyright (c) 2025 Justin Cranford

// Package notifications provides notification services for the identity platform.
package notifications

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"fmt"
	"log/slog"
)

// LogNotifier sends notifications to the application log.
type LogNotifier struct{}

// NewLogNotifier creates a new log notifier.
func NewLogNotifier() *LogNotifier {
	return &LogNotifier{}
}

// Send logs the expiration notification.
func (n *LogNotifier) Send(ctx context.Context, notification *ExpirationNotification) error {
	slog.WarnContext(ctx, "Client secret expiring soon",
		cryptoutilSharedMagic.ClaimClientID, notification.ClientID,
		"client_name", notification.ClientName,
		"version", notification.Version,
		"expires_at", notification.ExpiresAt,
		"days_remaining", notification.DaysRemaining,
	)

	return nil
}

// WebhookNotifier sends notifications via HTTP webhook.
type WebhookNotifier struct {
	webhookURL string
}

// NewWebhookNotifier creates a new webhook notifier.
func NewWebhookNotifier(webhookURL string) *WebhookNotifier {
	return &WebhookNotifier{
		webhookURL: webhookURL,
	}
}

// Send sends the notification to the webhook URL.
func (n *WebhookNotifier) Send(ctx context.Context, notification *ExpirationNotification) error {
	// TODO: Implement HTTP POST to webhook URL with notification payload.
	// For now, log that webhook would be called.
	slog.InfoContext(ctx, "Webhook notification (not implemented)",
		"webhook_url", n.webhookURL,
		cryptoutilSharedMagic.ClaimClientID, notification.ClientID,
		"client_name", notification.ClientName,
		"days_remaining", notification.DaysRemaining,
	)

	return nil
}

// EmailNotifier sends notifications via email.
type EmailNotifier struct {
	recipients []string
}

// NewEmailNotifier creates a new email notifier.
func NewEmailNotifier(recipients []string) *EmailNotifier {
	return &EmailNotifier{
		recipients: recipients,
	}
}

// Send sends the notification via email.
func (n *EmailNotifier) Send(ctx context.Context, notification *ExpirationNotification) error {
	// TODO: Implement SMTP email sending with notification details.
	// For now, log that email would be sent.
	slog.InfoContext(ctx, "Email notification (not implemented)",
		"recipients", n.recipients,
		cryptoutilSharedMagic.ClaimClientID, notification.ClientID,
		"client_name", notification.ClientName,
		"days_remaining", notification.DaysRemaining,
	)

	return fmt.Errorf("email notifications not yet implemented")
}
