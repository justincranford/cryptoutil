package authz

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
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

	// TODO: Store authorization request with PKCE challenge.
	// TODO: Redirect to login/consent flow.
	// TODO: Generate authorization code after user consent.

	// Store authorization request with PKCE challenge.
	requestID, err := googleUuid.NewV7()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate request ID", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to generate request ID",
		})
	}

	authRequest := &AuthorizationRequest{
		RequestID:           requestID,
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		ResponseType:        responseType,
		Scope:               scope,
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		CreatedAt:           time.Now(),
		ExpiresAt:           time.Now().Add(cryptoutilIdentityMagic.DefaultCodeLifetime),
		ConsentGranted:      false,
	}

	if err := s.authReqStore.Store(ctx, authRequest); err != nil {
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

	// Placeholder response - redirect to consent screen.
	// TODO: In future tasks, integrate with IdP for login/consent flow.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":    "Authorization request accepted - user authentication and consent required",
		"request_id": requestID.String(),
		"client_id":  clientID,
		"scope":      scope,
		"state":      state,
	})
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

	// TODO: Store authorization request with PKCE challenge.
	// TODO: Generate authorization code.

	// Store authorization request with PKCE challenge.
	requestID, err := googleUuid.NewV7()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate request ID", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to generate request ID",
		})
	}

	authRequest := &AuthorizationRequest{
		RequestID:           requestID,
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		ResponseType:        responseType,
		Scope:               scope,
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		CreatedAt:           time.Now(),
		ExpiresAt:           time.Now().Add(cryptoutilIdentityMagic.DefaultCodeLifetime),
		ConsentGranted:      false,
	}

	if err := s.authReqStore.Store(ctx, authRequest); err != nil {
		slog.ErrorContext(ctx, "Failed to store authorization request", "error", err, "request_id", requestID)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to store authorization request",
		})
	}

	// Generate authorization code (simulating consent being granted immediately for now).
	// TODO: In future tasks, integrate with IdP for login/consent flow before generating code.
	code, err := GenerateAuthorizationCode()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate authorization code", "error", err, "request_id", requestID)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to generate authorization code",
		})
	}

	// Update authorization request with code and consent.
	authRequest.Code = code
	authRequest.ConsentGranted = true

	if err := s.authReqStore.Update(ctx, authRequest); err != nil {
		slog.ErrorContext(ctx, "Failed to update authorization request", "error", err, "request_id", requestID)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Failed to update authorization request",
		})
	}

	slog.InfoContext(ctx, "Authorization code generated",
		"request_id", requestID,
		"client_id", clientID,
		"scope", scope,
	)

	// Build redirect URI with authorization code and state.
	redirectURL := redirectURI + "?code=" + code
	if state != "" {
		redirectURL += "&state=" + state
	}

	// Return 302 redirect to client's redirect_uri with authorization code.
	if err := c.Redirect(redirectURL, fiber.StatusFound); err != nil {
		return fmt.Errorf("failed to redirect to callback: %w", err)
	}

	return nil
}
