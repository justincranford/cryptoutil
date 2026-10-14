// Copyright (c) 2025-2026 Justin Cranford.
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package apis

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
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			valueErrorDescription:             valueClientIDIsRequired,
		})
	}

	if redirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			valueErrorDescription:             valueRedirectURIIsRequired,
		})
	}

	if responseType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			valueErrorDescription:             valueResponseTypeIsRequired,
		})
	}

	if responseType != cryptoutilSharedMagic.ResponseTypeCode {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorUnsupportedResponseType,
			valueErrorDescription:             valueOnlyCodeResponseTypeIsSupportedOAuth21NoImplicitFlow,
		})
	}

	// Validate PKCE parameters (required in OAuth 2.1).
	if codeChallenge == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			valueErrorDescription:             valueCodeChallengeIsRequiredOAuth21RequiresPkce,
		})
	}

	if codeChallengeMethod == "" {
		codeChallengeMethod = cryptoutilSharedMagic.PKCEMethodS256
	}

	if codeChallengeMethod != cryptoutilSharedMagic.PKCEMethodS256 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			valueErrorDescription:             valueOnlyS256CodeChallengeMethodIsSupported,
		})
	}

	// Validate client exists.
	ctx := c.Context()
	clientRepo := s.repoFactory.ClientRepository()

	client, err := clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrClientNotFound

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidClient,
			valueErrorDescription:             appErr.Message,
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
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			valueErrorDescription:             valueRedirectURIDoesNotMatchRegisteredUris,
		})
	}

	// Store authorization request with PKCE challenge in database.
	requestID, err := googleUuid.NewV7()
	if err != nil {
		slog.ErrorContext(ctx, valueFailedToGenerateRequestID, cryptoutilSharedMagic.StringError, err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			valueErrorDescription:             valueFailedToGenerateRequestID,
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
		slog.ErrorContext(ctx, valueFailedToStoreAuthorizationRequest, cryptoutilSharedMagic.StringError, err, "request_id", requestID)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			valueErrorDescription:             valueFailedToStoreAuthorizationRequest,
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
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			valueErrorDescription:             valueClientIDIsRequired,
		})
	}

	if redirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			valueErrorDescription:             valueRedirectURIIsRequired,
		})
	}

	if responseType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			valueErrorDescription:             valueResponseTypeIsRequired,
		})
	}

	if responseType != cryptoutilSharedMagic.ResponseTypeCode {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorUnsupportedResponseType,
			valueErrorDescription:             valueOnlyCodeResponseTypeIsSupportedOAuth21NoImplicitFlow,
		})
	}

	// Validate PKCE parameters.
	if codeChallenge == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			valueErrorDescription:             valueCodeChallengeIsRequiredOAuth21RequiresPkce,
		})
	}

	if codeChallengeMethod == "" {
		codeChallengeMethod = cryptoutilSharedMagic.PKCEMethodS256
	}

	if codeChallengeMethod != cryptoutilSharedMagic.PKCEMethodS256 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			valueErrorDescription:             valueOnlyS256CodeChallengeMethodIsSupported,
		})
	}

	// Validate client exists.
	ctx := c.Context()
	clientRepo := s.repoFactory.ClientRepository()

	client, err := clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		appErr := cryptoutilIdentityAppErr.ErrClientNotFound

		return c.Status(appErr.HTTPStatus).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidClient,
			valueErrorDescription:             appErr.Message,
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
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			valueErrorDescription:             valueRedirectURIDoesNotMatchRegisteredUris,
		})
	}

	// Store authorization request with PKCE challenge in database.
	requestID, err := googleUuid.NewV7()
	if err != nil {
		slog.ErrorContext(ctx, valueFailedToGenerateRequestID, cryptoutilSharedMagic.StringError, err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			valueErrorDescription:             valueFailedToGenerateRequestID,
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
		slog.ErrorContext(ctx, valueFailedToStoreAuthorizationRequest, cryptoutilSharedMagic.StringError, err, "request_id", requestID)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			valueErrorDescription:             valueFailedToStoreAuthorizationRequest,
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
