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
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// handleAuthorizeGET handles GET /authorize - OAuth 2.1 authorization endpoint.
func (s *Service) handleAuthorizeGET(c *fiber.Ctx) error {
	// Extract query parameters.
	clientID := c.Query(cryptoutilSharedMagic.ParamClientID)
	redirectURI := c.Query(cryptoutilSharedMagic.ParamRedirectURI)
	responseType := c.Query(cryptoutilSharedMagic.ParamResponseType)
	scope := c.Query(cryptoutilSharedMagic.ParamScope)
	state := c.Query(cryptoutilSharedMagic.ParamState)
	codeChallenge := c.Query(cryptoutilSharedMagic.ParamCodeChallenge)
	codeChallengeMethod := c.Query(cryptoutilSharedMagic.ParamCodeChallengeMethod)

	// Validate required parameters.
	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "client_id is required",
		})
	}

	if redirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "redirect_uri is required",
		})
	}

	if responseType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "response_type is required",
		})
	}

	if responseType != cryptoutilSharedMagic.ResponseTypeCode {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorUnsupportedResponseType,
			"error_description": "Only 'code' response_type is supported (OAuth 2.1 - no implicit flow)",
		})
	}

	// Validate PKCE parameters (required in OAuth 2.1).
	if codeChallenge == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "code_challenge is required (OAuth 2.1 requires PKCE)",
		})
	}

	if codeChallengeMethod == "" {
		codeChallengeMethod = cryptoutilSharedMagic.PKCEMethodS256
	}

	if codeChallengeMethod != cryptoutilSharedMagic.PKCEMethodS256 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
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
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidClient,
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
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "redirect_uri does not match registered URIs",
		})
	}

	// Store authorization request with PKCE challenge in database.
	requestID, err := googleUuid.NewV7()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate request ID", cryptoutilSharedMagic.StringError, err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorServerError,
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
		ExpiresAt:           time.Now().UTC().Add(cryptoutilSharedMagic.DefaultCodeLifetime),
		ConsentGranted:      false,
	}

	authzReqRepo := s.repoFactory.AuthorizationRequestRepository()
	if err := authzReqRepo.Create(ctx, authRequest); err != nil {
		slog.ErrorContext(ctx, "Failed to store authorization request", cryptoutilSharedMagic.StringError, err, "request_id", requestID)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorServerError,
			"error_description": "Failed to store authorization request",
		})
	}

	slog.InfoContext(ctx, "Authorization request created",
		"request_id", requestID,
		cryptoutilSharedMagic.ClaimClientID, clientID,
		cryptoutilSharedMagic.ClaimScope, scope,
	)

	// Redirect to IdP login page with request_id parameter.
	loginURL := fmt.Sprintf("/oidc/v1/login?request_id=%s", requestID.String())

	return c.Redirect(loginURL, fiber.StatusFound)
}

// handleAuthorizePOST handles POST /authorize - OAuth 2.1 authorization endpoint (form submission).
func (s *Service) handleAuthorizePOST(c *fiber.Ctx) error {
	// Extract form parameters.
	clientID := c.FormValue(cryptoutilSharedMagic.ParamClientID)
	redirectURI := c.FormValue(cryptoutilSharedMagic.ParamRedirectURI)
	responseType := c.FormValue(cryptoutilSharedMagic.ParamResponseType)
	scope := c.FormValue(cryptoutilSharedMagic.ParamScope)
	state := c.FormValue(cryptoutilSharedMagic.ParamState)
	codeChallenge := c.FormValue(cryptoutilSharedMagic.ParamCodeChallenge)
	codeChallengeMethod := c.FormValue(cryptoutilSharedMagic.ParamCodeChallengeMethod)

	// Validate required parameters (same as GET).
	if clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "client_id is required",
		})
	}

	if redirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "redirect_uri is required",
		})
	}

	if responseType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "response_type is required",
		})
	}

	if responseType != cryptoutilSharedMagic.ResponseTypeCode {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorUnsupportedResponseType,
			"error_description": "Only 'code' response_type is supported (OAuth 2.1 - no implicit flow)",
		})
	}

	// Validate PKCE parameters.
	if codeChallenge == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "code_challenge is required (OAuth 2.1 requires PKCE)",
		})
	}

	if codeChallengeMethod == "" {
		codeChallengeMethod = cryptoutilSharedMagic.PKCEMethodS256
	}

	if codeChallengeMethod != cryptoutilSharedMagic.PKCEMethodS256 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
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
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidClient,
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
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description": "redirect_uri does not match registered URIs",
		})
	}

	// Store authorization request with PKCE challenge in database.
	requestID, err := googleUuid.NewV7()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate request ID", cryptoutilSharedMagic.StringError, err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorServerError,
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
		ExpiresAt:           time.Now().UTC().Add(cryptoutilSharedMagic.DefaultCodeLifetime),
		ConsentGranted:      false,
	}

	authzReqRepo := s.repoFactory.AuthorizationRequestRepository()
	if err := authzReqRepo.Create(ctx, authRequest); err != nil {
		slog.ErrorContext(ctx, "Failed to store authorization request", cryptoutilSharedMagic.StringError, err, "request_id", requestID)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError:             cryptoutilSharedMagic.ErrorServerError,
			"error_description": "Failed to store authorization request",
		})
	}

	slog.InfoContext(ctx, "Authorization request created",
		"request_id", requestID,
		cryptoutilSharedMagic.ClaimClientID, clientID,
		cryptoutilSharedMagic.ClaimScope, scope,
	)

	// Redirect to IdP login page with request_id parameter.
	loginURL := fmt.Sprintf("/oidc/v1/login?request_id=%s", requestID.String())

	return c.Redirect(loginURL, fiber.StatusFound)
}
