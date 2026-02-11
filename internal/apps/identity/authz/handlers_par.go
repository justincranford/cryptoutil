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

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/apps/identity/magic"
)

// PARResponse represents the response from POST /par (RFC 9126 Section 2.1).
type PARResponse struct {
	RequestURI string `json:"request_uri"`
	ExpiresIn  int    `json:"expires_in"`
}

// handlePAR handles POST /oauth2/v1/par - Pushed Authorization Request endpoint (RFC 9126).
//
// PAR allows OAuth clients to push authorization request parameters directly to the
// authorization server before redirecting the user agent. This provides:
// - Request integrity: Parameters cannot be tampered with in transit
// - Confidentiality: Sensitive parameters (code_challenge) not exposed in URLs
// - Phishing resistance: Prevents interception of authorization parameters
//
// Request parameters (all from OAuth 2.1 /authorize):
// - client_id (required for public clients)
// - response_type (required) - "code" for authorization code flow
// - redirect_uri (required)
// - scope (optional)
// - state (optional but recommended)
// - code_challenge (required per PKCE)
// - code_challenge_method (required) - "S256"
// - nonce (optional for OIDC)
//
// Response (201 Created):
// - request_uri: urn:ietf:params:oauth:request_uri:<opaque-identifier>
// - expires_in: Lifetime in seconds (default: 90)
//
// Error Response (400 Bad Request):
// - error: OAuth 2.1 error code
// - error_description: Human-readable error message.
func (s *Service) handlePAR(c *fiber.Ctx) error {
	ctx := c.Context()

	// Extract authorization request parameters from form body.
	clientIDStr := c.FormValue(cryptoutilIdentityMagic.ParamClientID)
	responseType := c.FormValue(cryptoutilIdentityMagic.ParamResponseType)
	redirectURI := c.FormValue(cryptoutilIdentityMagic.ParamRedirectURI)
	scope := c.FormValue(cryptoutilIdentityMagic.ParamScope)
	state := c.FormValue(cryptoutilIdentityMagic.ParamState)
	codeChallenge := c.FormValue(cryptoutilIdentityMagic.ParamCodeChallenge)
	codeChallengeMethod := c.FormValue(cryptoutilIdentityMagic.ParamCodeChallengeMethod)
	nonce := c.FormValue("nonce")

	// Validate required parameters.
	if clientIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Missing required parameter: client_id",
		})
	}

	if responseType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Missing required parameter: response_type",
		})
	}

	if redirectURI == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Missing required parameter: redirect_uri",
		})
	}

	if codeChallenge == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Missing required parameter: code_challenge",
		})
	}

	if codeChallengeMethod == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "Missing required parameter: code_challenge_method",
		})
	}

	// Validate response_type (only "code" supported).
	if responseType != cryptoutilIdentityMagic.ResponseTypeCode {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorUnsupportedResponseType,
			"error_description": fmt.Sprintf("Only response_type=%s is supported", cryptoutilIdentityMagic.ResponseTypeCode),
		})
	}

	// Validate code_challenge_method (only S256 supported per PKCE).
	if codeChallengeMethod != cryptoutilIdentityMagic.PKCEMethodS256 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": fmt.Sprintf("Only code_challenge_method=%s is supported", cryptoutilIdentityMagic.PKCEMethodS256),
		})
	}

	// Validate client exists and redirect_uri is registered.
	clientRepo := s.repoFactory.ClientRepository()

	client, err := clientRepo.GetByClientID(ctx, clientIDStr)
	if err != nil {
		slog.ErrorContext(ctx, "Client not found for PAR request", "client_id", clientIDStr, "error", err)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidClient,
			"error_description": "Invalid client_id",
		})
	}

	// Validate redirect_uri against client configuration.
	validRedirectURI := false

	for _, registeredURI := range client.RedirectURIs {
		if registeredURI == redirectURI {
			validRedirectURI = true

			break
		}
	}

	if !validRedirectURI {
		slog.ErrorContext(ctx, "Invalid redirect_uri for client", "client_id", clientIDStr, "redirect_uri", redirectURI)

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorInvalidRequest,
			"error_description": "redirect_uri not registered for this client",
		})
	}

	// Generate cryptographically random request_uri.
	requestURI, err := GenerateRequestURI()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate request_uri", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Internal server error",
		})
	}

	// Create and store PushedAuthorizationRequest.
	now := time.Now().UTC()
	expiresAt := now.Add(cryptoutilIdentityMagic.DefaultPARLifetime)

	par := &cryptoutilIdentityDomain.PushedAuthorizationRequest{
		ID:                  googleUuid.Must(googleUuid.NewV7()),
		RequestURI:          requestURI,
		ClientID:            client.ID,
		ResponseType:        responseType,
		RedirectURI:         redirectURI,
		Scope:               scope,
		State:               state,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Nonce:               nonce,
		Used:                false,
		ExpiresAt:           expiresAt,
		CreatedAt:           now,
	}

	parRepo := s.repoFactory.PushedAuthorizationRequestRepository()
	if err := parRepo.Create(ctx, par); err != nil {
		slog.ErrorContext(ctx, "Failed to create pushed authorization request", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":             cryptoutilIdentityMagic.ErrorServerError,
			"error_description": "Internal server error",
		})
	}

	// Return response (201 Created).
	response := PARResponse{
		RequestURI: requestURI,
		ExpiresIn:  int(cryptoutilIdentityMagic.DefaultPARLifetime.Seconds()),
	}

	slog.InfoContext(ctx, "Pushed authorization request created",
		"request_uri", requestURI,
		"client_id", clientIDStr,
		"expires_at", expiresAt)

	return c.Status(fiber.StatusCreated).JSON(response)
}
