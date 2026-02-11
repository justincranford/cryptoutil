// Copyright (c) 2025 Justin Cranford
//
//

// Package idp provides the Identity Provider (IdP) implementation.
package idp

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	http "net/http"
	"net/url"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityIssuer "cryptoutil/internal/apps/identity/issuer"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

// BackChannelLogoutService handles back-channel logout token generation and delivery.
type BackChannelLogoutService struct {
	tokenService *cryptoutilIdentityIssuer.TokenService
	httpClient   *http.Client
	logger       *slog.Logger
	issuer       string
}

// NewBackChannelLogoutService creates a new back-channel logout service.
func NewBackChannelLogoutService(tokenService *cryptoutilIdentityIssuer.TokenService, issuer string, logger *slog.Logger) *BackChannelLogoutService {
	return &BackChannelLogoutService{
		tokenService: tokenService,
		issuer:       issuer,
		httpClient: &http.Client{
			Timeout: cryptoutilIdentityMagic.BackChannelLogoutTimeout,
		},
		logger: logger,
	}
}

// SendBackChannelLogout sends back-channel logout tokens to all registered clients.
// Returns the number of successful notifications and any errors encountered.
func (s *BackChannelLogoutService) SendBackChannelLogout(ctx context.Context, session *cryptoutilIdentityDomain.Session, clients []*cryptoutilIdentityDomain.Client) (int, []error) {
	var (
		successCount int
		errors       []error
	)

	for _, client := range clients {
		if client.BackChannelLogoutURI == "" {
			continue
		}

		// Generate logout token for this client.
		logoutToken, err := s.generateLogoutToken(ctx, session, client)
		if err != nil {
			s.logger.Error("Failed to generate logout token",
				"client_id", client.ClientID,
				"error", err,
			)

			errors = append(errors, fmt.Errorf("failed to generate logout token for %s: %w", client.ClientID, err))

			continue
		}

		// Send logout token to client's back-channel logout URI.
		if err := s.deliverLogoutToken(ctx, client.BackChannelLogoutURI, logoutToken); err != nil {
			s.logger.Error("Failed to deliver logout token",
				"client_id", client.ClientID,
				"uri", client.BackChannelLogoutURI,
				"error", err,
			)

			errors = append(errors, fmt.Errorf("failed to deliver logout token to %s: %w", client.ClientID, err))

			continue
		}

		s.logger.Info("Back-channel logout successful",
			"client_id", client.ClientID,
			"uri", client.BackChannelLogoutURI,
		)

		successCount++
	}

	return successCount, errors
}

// generateLogoutToken creates a JWT logout token per OpenID Connect Back-Channel Logout 1.0.
func (s *BackChannelLogoutService) generateLogoutToken(ctx context.Context, session *cryptoutilIdentityDomain.Session, client *cryptoutilIdentityDomain.Client) (string, error) {
	now := time.Now().UTC()

	// Build logout token claims as a map for the token service.
	claims := map[string]any{
		cryptoutilIdentityMagic.ClaimIss: s.issuer,
		cryptoutilIdentityMagic.ClaimAud: client.ClientID,
		cryptoutilIdentityMagic.ClaimIat: now.Unix(),
		"jti":                            googleUuid.Must(googleUuid.NewV7()).String(),
		"events": map[string]any{
			"http://schemas.openid.net/event/backchannel-logout": map[string]any{},
		},
	}

	// Add subject if available from session.
	if session.UserID != googleUuid.Nil {
		claims[cryptoutilIdentityMagic.ClaimSub] = session.UserID.String()
	}

	// Add session ID if client requires it.
	if client.BackChannelLogoutSessionRequired != nil && *client.BackChannelLogoutSessionRequired {
		claims["sid"] = session.SessionID
	}

	// Issue as ID token (uses same signing mechanism).
	token, err := s.tokenService.IssueIDToken(ctx, claims)
	if err != nil {
		return "", fmt.Errorf("failed to issue logout token: %w", err)
	}

	return token, nil
}

// deliverLogoutToken sends the logout token to the client's back-channel logout URI.
func (s *BackChannelLogoutService) deliverLogoutToken(ctx context.Context, logoutURI string, token string) error {
	// Prepare form data.
	formData := url.Values{}
	formData.Set("logout_token", token)

	// Create POST request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, logoutURI, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute request.
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }() //nolint:errcheck // Best effort cleanup

	// Check response status.
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("back-channel logout failed with status %d", resp.StatusCode)
	}

	return nil
}

// GenerateFrontChannelLogoutIframes generates HTML iframes for front-channel logout.
func GenerateFrontChannelLogoutIframes(clients []*cryptoutilIdentityDomain.Client, sessionID string) string {
	var iframes string

	for _, client := range clients {
		if client.FrontChannelLogoutURI == "" {
			continue
		}

		// Build logout URI with optional session ID.
		logoutURI := client.FrontChannelLogoutURI
		if client.FrontChannelLogoutSessionRequired != nil && *client.FrontChannelLogoutSessionRequired && sessionID != "" {
			parsedURI, err := url.Parse(logoutURI)
			if err == nil {
				query := parsedURI.Query()
				query.Set("sid", sessionID)
				parsedURI.RawQuery = query.Encode()
				logoutURI = parsedURI.String()
			}
		}

		// Generate hidden iframe.
		iframes += fmt.Sprintf(`<iframe src="%s" style="display:none;" width="0" height="0"></iframe>`, logoutURI)
	}

	return iframes
}
