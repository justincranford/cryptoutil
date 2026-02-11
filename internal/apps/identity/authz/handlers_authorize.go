// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package authz

import (
	"fmt"
	"log/slog"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

// handleAuthorizeGET handles GET /authorize - OAuth 2.1 authorization endpoint.
func (s *Service) handleAuthorizeGET(c *fiber.Ctx) error {
	// Extract query parameters.
	clientID := c.Query(cryptoutilIdentityMagic.ParamClientID)
	redirectURI := c.Query(cryptoutilIdentityMagic.ParamRedirectURI)
	responseType := c.Query(cryptoutilIdentityMagic.ParamResponseType)
	scope := c.Query(cryptoutilIdentityMagic.ParamScope)
	state := c.Query(cryptoutilIdentityMagic.ParamState)
	codeChallenge := c.Query(cryptoutilIdentityMagic.ParamCodeChallenge)
	codeChallengeMethod := c.Query(cryptoutilIdentityMagic.ParamCodeChallengeMethod)

	// Validate required parameters.
	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "client_id is required",
		})
	}

	if redirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "redirect_uri is required",
		})
	}

	if responseType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "response_type is required",
		})
	}

	if responseType != cryptoutilIdentityMagic.ResponseTypeCode {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorUnsupportedResponseType,
			"error_description": "Only 'code' response_type is supported (OAuth 2.1 - no implicit flow)",
		})
	}

	// Validate PKCE parameters (required in OAuth 2.1).
	if codeChallenge == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "code_challenge is required (OAuth 2.1 requires PKCE)",
		})
	}

	if codeChallengeMethod == "" {
		codeChallengeMethod = cryptoutilIdentityMagic.PKCEMethodS256
	}

	if codeChallengeMethod != cryptoutilIdentityMagic.PKCEMethodS256 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Only 'S256' code_challenge_method is supported",
		})
	}

	// Validate client exists.
	ctx := c.Context()
	clientRepo := s.repoFactory.ClientRepository()

	client, err := clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrClientNotFound

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidClient,
			"error_description": appErr.Message,
		})
	}

	// Validate redirect URI matches registered URIs.
	validRedirectURI := false

	for _, uri := range client.RedirectURIs {
		if uri == redirectURI {
			validRedirectURI = true

			break
		}
	}

	if !validRedirectURI {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "redirect_uri does not match registered URIs",
		})
	}

	// Store authorization request with PKCE challenge in database.
	requestID, err := googleUuid.NewV7()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate request ID", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to generate request ID",
		})
	}

	authRequest := &cryptoutilIdentityDomain.AuthorizationRequest{
		ID:                  requestID,
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		ResponseType:        responseType,
		Scope:               scope,
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Code:                googleUuid.Must(googleUuid.NewV7()).String(), // Pre-generate authorization code (GET handler)
		CreatedAt:           time.Now().UTC(),
		ExpiresAt:           time.Now().UTC().Add(cryptoutilIdentityMagic.DefaultCodeLifetime),
		ConsentGranted:      false,
	}

	authzReqRepo := s.repoFactory.AuthorizationRequestRepository()
	if err := authzReqRepo.Create(ctx, authRequest); err != nil {
		slog.ErrorContext(ctx, "Failed to store authorization request", "error", err, "request_id", requestID)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to store authorization request",
		})
	}

	slog.InfoContext(ctx, "Authorization request created",
		"request_id", requestID,
		"client_id", clientID,
		"scope", scope,
	)

	// Redirect to IdP login page with request_id parameter.
	loginURL := fmt.Sprintf("/oidc/v1/login?request_id=%s", requestID.String())

	return c.Redirect(loginURL, fiber.StatusFound)
}

// handleAuthorizePOST handles POST /authorize - OAuth 2.1 authorization endpoint (form submission).
func (s *Service) handleAuthorizePOST(c *fiber.Ctx) error {
	// Extract form parameters.
	clientID := c.FormValue(cryptoutilIdentityMagic.ParamClientID)
	redirectURI := c.FormValue(cryptoutilIdentityMagic.ParamRedirectURI)
	responseType := c.FormValue(cryptoutilIdentityMagic.ParamResponseType)
	scope := c.FormValue(cryptoutilIdentityMagic.ParamScope)
	state := c.FormValue(cryptoutilIdentityMagic.ParamState)
	codeChallenge := c.FormValue(cryptoutilIdentityMagic.ParamCodeChallenge)
	codeChallengeMethod := c.FormValue(cryptoutilIdentityMagic.ParamCodeChallengeMethod)

	// Validate required parameters (same as GET).
	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "client_id is required",
		})
	}

	if redirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "redirect_uri is required",
		})
	}

	if responseType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "response_type is required",
		})
	}

	if responseType != cryptoutilIdentityMagic.ResponseTypeCode {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorUnsupportedResponseType,
			"error_description": "Only 'code' response_type is supported (OAuth 2.1 - no implicit flow)",
		})
	}

	// Validate PKCE parameters.
	if codeChallenge == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "code_challenge is required (OAuth 2.1 requires PKCE)",
		})
	}

	if codeChallengeMethod == "" {
		codeChallengeMethod = cryptoutilIdentityMagic.PKCEMethodS256
	}

	if codeChallengeMethod != cryptoutilIdentityMagic.PKCEMethodS256 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Only 'S256' code_challenge_method is supported",
		})
	}

	// Validate client exists.
	ctx := c.Context()
	clientRepo := s.repoFactory.ClientRepository()

	client, err := clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrClientNotFound

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidClient,
			"error_description": appErr.Message,
		})
	}

	// Validate redirect URI.
	validRedirectURI := false

	for _, uri := range client.RedirectURIs {
		if uri == redirectURI {
			validRedirectURI = true

			break
		}
	}

	if !validRedirectURI {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "redirect_uri does not match registered URIs",
		})
	}

	// Store authorization request with PKCE challenge in database.
	requestID, err := googleUuid.NewV7()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate request ID", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to generate request ID",
		})
	}

	authRequest := &cryptoutilIdentityDomain.AuthorizationRequest{
		ID:                  requestID,
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		ResponseType:        responseType,
		Scope:               scope,
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Code:                googleUuid.Must(googleUuid.NewV7()).String(), // Pre-generate authorization code (POST handler)
		CreatedAt:           time.Now().UTC(),
		ExpiresAt:           time.Now().UTC().Add(cryptoutilIdentityMagic.DefaultCodeLifetime),
		ConsentGranted:      false,
	}

	authzReqRepo := s.repoFactory.AuthorizationRequestRepository()
	if err := authzReqRepo.Create(ctx, authRequest); err != nil {
		slog.ErrorContext(ctx, "Failed to store authorization request", "error", err, "request_id", requestID)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to store authorization request",
		})
	}

	slog.InfoContext(ctx, "Authorization request created",
		"request_id", requestID,
		"client_id", clientID,
		"scope", scope,
	)

	// Redirect to IdP login page with request_id parameter.
	loginURL := fmt.Sprintf("/oidc/v1/login?request_id=%s", requestID.String())

	return c.Redirect(loginURL, fiber.StatusFound)
}
