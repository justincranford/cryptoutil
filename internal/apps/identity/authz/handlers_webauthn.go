// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck
package authz

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"net/url"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityMfa "cryptoutil/internal/apps/identity/mfa"
)

// WebAuthn Request/Response types.

type BeginWebAuthnRegistrationRequest struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type BeginWebAuthnRegistrationResponse struct {
	Options   any `json:"options"`
	SessionID string      `json:"session_id"`
}

type FinishWebAuthnRegistrationRequest struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	SessionID   string `json:"session_id"`
	Response    string `json:"response"`
}

type FinishWebAuthnRegistrationResponse struct {
	Success      bool   `json:"success"`
	CredentialID string `json:"credential_id"`
}

type BeginWebAuthnAuthenticationRequest struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

type BeginWebAuthnAuthenticationResponse struct {
	Options   any `json:"options"`
	SessionID string      `json:"session_id"`
}

type FinishWebAuthnAuthenticationRequest struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	SessionID string `json:"session_id"`
	Response  string `json:"response"`
}

type FinishWebAuthnAuthenticationResponse struct {
	Success      bool   `json:"success"`
	CredentialID string `json:"credential_id"`
	LastUsedAt   string `json:"last_used_at,omitempty"`
}

type WebAuthnCredentialResponse struct {
	ID              string `json:"id"`
	CredentialID    string `json:"credential_id"`
	DisplayName     string `json:"display_name"`
	CreatedAt       string `json:"created_at"`
	LastUsedAt      string `json:"last_used_at,omitempty"`
	SignCount       uint32 `json:"sign_count"`
	AttestationType string `json:"attestation_type,omitempty"`
	Transports      string `json:"transports,omitempty"`
	AAGUID          string `json:"aaguid,omitempty"`
}

type ListWebAuthnCredentialsResponse struct {
	Credentials []WebAuthnCredentialResponse `json:"credentials"`
}

type DeleteWebAuthnCredentialRequest struct {
	UserID       string `json:"user_id"`
	CredentialID string `json:"credential_id"`
}

// getWebAuthnService creates a WebAuthn service instance.
func (s *Service) getWebAuthnService() (*cryptoutilIdentityMfa.WebAuthnService, error) {
	db := s.repoFactory.DB()

	// Derive RP config from issuer URL.
	issuerURL, err := url.Parse(s.config.Tokens.Issuer)
	if err != nil {
		return nil, err
	}

	config := cryptoutilIdentityMfa.WebAuthnConfig{
		RPDisplayName: "Cryptoutil Identity",
		RPID:          issuerURL.Hostname(),
		RPOrigins:     []string{s.config.Tokens.Issuer},
	}

	return cryptoutilIdentityMfa.NewWebAuthnService(db, config)
}

// loadWebAuthnUser loads a user and their WebAuthn credentials.
func (s *Service) loadWebAuthnUser(ctx context.Context, userID googleUuid.UUID, username, displayName string, webauthnSvc *cryptoutilIdentityMfa.WebAuthnService) (*cryptoutilIdentityMfa.WebAuthnUser, error) {
	// Load existing credentials for the user.
	creds, err := webauthnSvc.GetCredentials(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &cryptoutilIdentityMfa.WebAuthnUser{
		ID:          userID,
		Name:        username,
		DisplayName: displayName,
		Credentials: creds,
	}, nil
}

// BeginWebAuthnRegistration starts WebAuthn credential registration.
func (s *Service) BeginWebAuthnRegistration(c *fiber.Ctx) error {
	var req BeginWebAuthnRegistrationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid request body"})
	}

	// Validate required fields.
	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "user_id is required"})
	}

	if req.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "username is required"})
	}

	if req.DisplayName == "" {
		req.DisplayName = req.Username
	}

	// Parse user ID.
	userID, err := googleUuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid user_id format"})
	}

	// Get WebAuthn service.
	webauthnSvc, err := s.getWebAuthnService()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "WebAauthn service unavailable"})
	}

	// Load WebAuthn user.
	ctx := c.Context()

	user, err := s.loadWebAuthnUser(ctx, userID, req.Username, req.DisplayName, webauthnSvc)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Failed to load user"})
	}

	// Begin registration.
	options, sessionID, err := webauthnSvc.BeginRegistration(ctx, user)
	if err != nil {
		// Check for specific errors.
		switch {
		case errors.Is(err, cryptoutilIdentityAppErr.ErrWebAuthnSessionAlreadyExists):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Reauthentication session already exists"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Failed to begin registration"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(BeginWebAuthnRegistrationResponse{
		Options:   options,
		SessionID: sessionID.String(),
	})
}

// FinishWebAuthnRegistration completes WebAuthn credential registration.
func (s *Service) FinishWebAuthnRegistration(c *fiber.Ctx) error {
	var req FinishWebAuthnRegistrationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid request body"})
	}

	// Validate required fields.
	if req.UserID == "" || req.Username == "" || req.SessionID == "" || req.Response == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Missing required fields"})
	}

	if req.DisplayName == "" {
		req.DisplayName = req.Username
	}

	// Parse IDs.
	userID, err := googleUuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid user_id format"})
	}

	sessionID, err := googleUuid.Parse(req.SessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid session_id format"})
	}

	// Parse the credential creation response.
	responseBytes, err := base64.StdEncoding.DecodeString(req.Response)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid response encoding"})
	}

	parsedResponse, err := protocol.ParseCredentialCreationResponseBody(bytes.NewReader(responseBytes))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid credential response"})
	}

	// Get WebAuthn service.
	webauthnSvc, err := s.getWebAuthnService()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "WebAauthn service unavailable"})
	}

	// Load WebAuthn user.
	ctx := c.Context()

	user, err := s.loadWebAuthnUser(ctx, userID, req.Username, req.DisplayName, webauthnSvc)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Failed to load user"})
	}

	// Finish registration.
	credential, err := webauthnSvc.FinishRegistration(ctx, user, sessionID, parsedResponse, req.DisplayName)
	if err != nil {
		switch {
		case errors.Is(err, cryptoutilIdentityAppErr.ErrWebAuthnSessionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Session not found or expired"})
		case errors.Is(err, cryptoutilIdentityAppErr.ErrWebAuthnSessionExpired):
			return c.Status(fiber.StatusGone).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Session has expired"})
		case errors.Is(err, cryptoutilIdentityAppErr.ErrWebAuthnCredentialAlreadyExists):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Credential already exists"})
		case errors.Is(err, cryptoutilIdentityAppErr.ErrWebAuthnVerificationFailed):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Credential verification failed"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Failed to complete registration"})
		}
	}

	return c.Status(fiber.StatusCreated).JSON(FinishWebAuthnRegistrationResponse{
		Success:      true,
		CredentialID: base64.StdEncoding.EncodeToString(credential.CredentialID),
	})
}

// BeginWebAuthnAuthentication starts WebAuthn authentication.
func (s *Service) BeginWebAuthnAuthentication(c *fiber.Ctx) error {
	var req BeginWebAuthnAuthenticationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid request body"})
	}

	// Validate required fields.
	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "user_id is required"})
	}

	if req.Username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "username is required"})
	}

	// Parse user ID.
	userID, err := googleUuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid user_id format"})
	}

	// Get WebAuthn service.
	webauthnSvc, err := s.getWebAuthnService()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "WebAuthn service unavailable"})
	}

	// Load WebAuthn user with credentials.
	ctx := c.Context()

	user, err := s.loadWebAuthnUser(ctx, userID, req.Username, "", webauthnSvc)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Failed to load user"})
	}

	// Check if user has any credentials.
	if len(user.Credentials) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "No credentials registered for this user"})
	}

	// Begin authentication.
	options, sessionID, err := webauthnSvc.BeginAuthentication(ctx, user)
	if err != nil {
		switch {
		case errors.Is(err, cryptoutilIdentityAppErr.ErrWebAuthnSessionAlreadyExists):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Authentication session already exists"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Failed to begin authentication"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(BeginWebAuthnAuthenticationResponse{
		Options:   options,
		SessionID: sessionID.String(),
	})
}

// FinishWebAuthnAuthentication completes WebAuthn authentication.
func (s *Service) FinishWebAuthnAuthentication(c *fiber.Ctx) error {
	var req FinishWebAuthnAuthenticationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid request body"})
	}

	// Validate required fields.
	if req.UserID == "" || req.Username == "" || req.SessionID == "" || req.Response == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Missing required fields"})
	}

	// Parse IDs.
	userID, err := googleUuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid user_id format"})
	}

	sessionID, err := googleUuid.Parse(req.SessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid session_id format"})
	}

	// Parse the credential assertion response.
	responseBytes, err := base64.StdEncoding.DecodeString(req.Response)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid response encoding"})
	}

	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(bytes.NewReader(responseBytes))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid credential response"})
	}

	// Get WebAuthn service.
	webauthnSvc, err := s.getWebAuthnService()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "WebAuthn service unavailable"})
	}

	// Load WebAuthn user with credentials.
	ctx := c.Context()

	user, err := s.loadWebAuthnUser(ctx, userID, req.Username, "", webauthnSvc)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Failed to load user"})
	}

	// Finish authentication.
	credential, err := webauthnSvc.FinishAuthentication(ctx, user, sessionID, parsedResponse)
	if err != nil {
		switch {
		case errors.Is(err, cryptoutilIdentityAppErr.ErrWebAuthnSessionNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Session not found or expired"})
		case errors.Is(err, cryptoutilIdentityAppErr.ErrWebAuthnSessionExpired):
			return c.Status(fiber.StatusGone).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Session has expired"})
		case errors.Is(err, cryptoutilIdentityAppErr.ErrWebAuthnCredentialNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Credential not found"})
		case errors.Is(err, cryptoutilIdentityAppErr.ErrWebAuthnVerificationFailed):
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Authentication verification failed"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Failed to complete authentication"})
		}
	}

	lastUsed := ""
	if credential.LastUsedAt != nil {
		lastUsed = credential.LastUsedAt.Format(time.RFC3339)
	}

	return c.Status(fiber.StatusOK).JSON(FinishWebAuthnAuthenticationResponse{
		Success:      true,
		CredentialID: base64.StdEncoding.EncodeToString(credential.CredentialID),
		LastUsedAt:   lastUsed,
	})
}

// ListWebAuthnCredentials lists all WebAuthn credentials for a user.
func (s *Service) ListWebAuthnCredentials(c *fiber.Ctx) error {
	userIDString := c.Params("user_id")
	if userIDString == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "user_id is required"})
	}

	userID, err := googleUuid.Parse(userIDString)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid user_id format"})
	}

	// Get WebAuthn service.
	webauthnSvc, err := s.getWebAuthnService()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "WebAuthn service unavailable"})
	}

	// Get credentials for the user.
	ctx := c.Context()

	credentials, err := webauthnSvc.GetCredentials(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Failed to get credentials"})
	}

	// Convert to response format.
	responseCreds := make([]WebAuthnCredentialResponse, 0, len(credentials))
	for _, cred := range credentials {
		lastUsed := ""
		if cred.LastUsedAt != nil {
			lastUsed = cred.LastUsedAt.Format(time.RFC3339)
		}

		responseCreds = append(responseCreds, WebAuthnCredentialResponse{
			ID:          cred.ID.String(),
			DisplayName: cred.DisplayName,
			CreatedAt:   cred.CreatedAt.Format(time.RFC3339),
			LastUsedAt:  lastUsed,
		})
	}

	return c.Status(fiber.StatusOK).JSON(ListWebAuthnCredentialsResponse{
		Credentials: responseCreds,
	})
}

// DeleteWebAuthnCredential deletes a WebAuthn credential.
func (s *Service) DeleteWebAuthnCredential(c *fiber.Ctx) error {
	var req DeleteWebAuthnCredentialRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid request body"})
	}

	// Validate required fields.
	if req.UserID == "" || req.CredentialID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Missing required fields"})
	}

	// Parse IDs.
	userID, err := googleUuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid user_id format"})
	}

	credentialID, err := googleUuid.Parse(req.CredentialID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Invalid credential_id format"})
	}

	// Get WebAuthn service.
	webauthnSvc, err := s.getWebAuthnService()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "WebAuthn service unavailable"})
	}

	// Delete the credential.
	ctx := c.Context()

	err = webauthnSvc.DeleteCredential(ctx, userID, credentialID)
	if err != nil {
		switch {
		case errors.Is(err, cryptoutilIdentityAppErr.ErrWebAuthnCredentialNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Credential not found"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "Failed to delete credential"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"success": true})
}
