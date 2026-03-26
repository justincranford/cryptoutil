// Copyright (c) 2025 Justin Cranford
//
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package idp

import (
	"fmt"
	"net/url"

	fiber "github.com/gofiber/fiber/v2"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// handleLogout handles POST /logout - Terminate user session.
func (s *Service) handleLogout(c *fiber.Ctx) error {
	ctx := c.Context()

	// Extract session cookie.
	sessionID := c.Cookies(s.config.Sessions.CookieName)

	if sessionID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "No active session found",
		})
	}

	// Retrieve session from database.
	sessionRepo := s.repoFactory.SessionRepository()

	session, err := sessionRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Session not found",
		})
	}

	// Delete session from database.
	if err := sessionRepo.Delete(ctx, session.ID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorServerError,
			"error_description":               "Failed to delete session",
		})
	}

	// Clear session cookie.
	c.ClearCookie(s.config.Sessions.CookieName)

	// Return success response.
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// handleEndSession handles GET /endsession - OpenID Connect RP-Initiated Logout (RFC).
// See: https://openid.net/specs/openid-connect-rpinitiated-1_0.html
//
// Parameters:
// - id_token_hint: Previously issued ID token (recommended for identifying the user)
// - client_id: Client identifier (required if id_token_hint not provided)
// - post_logout_redirect_uri: URI to redirect after logout (optional, must be registered)
// - state: Opaque value for maintaining state between request and callback (optional).
func (s *Service) handleEndSession(c *fiber.Ctx) error {
	ctx := c.Context()

	// Extract parameters.
	idTokenHint := c.Query("id_token_hint")
	clientID := c.Query(cryptoutilSharedMagic.ClaimClientID)
	postLogoutRedirectURI := c.Query("post_logout_redirect_uri")
	state := c.Query(cryptoutilSharedMagic.ParamState)

	// Validate: either id_token_hint or client_id must be provided.
	if idTokenHint == "" && clientID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
			"error_description":               "Either id_token_hint or client_id is required",
		})
	}

	// Extract session cookie if present.
	sessionID := c.Cookies(s.config.Sessions.CookieName)

	// Clear session if one exists.
	if sessionID != "" {
		sessionRepo := s.repoFactory.SessionRepository()

		session, err := sessionRepo.GetBySessionID(ctx, sessionID)
		if err == nil {
			// Delete session from database.
			_ = sessionRepo.Delete(ctx, session.ID) //nolint:errcheck // Best effort cleanup
		}

		// Clear session cookie.
		c.ClearCookie(s.config.Sessions.CookieName)
	}

	// Validate post_logout_redirect_uri if provided.
	if postLogoutRedirectURI != "" {
		// Validate URI is well-formed.
		parsedURI, parseErr := url.Parse(postLogoutRedirectURI)
		if parseErr != nil || parsedURI.Scheme == "" || parsedURI.Host == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
				"error_description":               "Invalid post_logout_redirect_uri",
			})
		}

		// If client_id provided, validate redirect URI is registered for client.
		if clientID != "" {
			clientRepo := s.repoFactory.ClientRepository()

			client, err := clientRepo.GetByClientID(ctx, clientID)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
					"error_description":               "Client not found",
				})
			}

			// Check if post_logout_redirect_uri is in allowed list.
			uriAllowed := false

			for _, allowedURI := range client.PostLogoutRedirectURIs {
				if allowedURI == postLogoutRedirectURI {
					uriAllowed = true

					break
				}
			}

			if !uriAllowed {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					cryptoutilSharedMagic.StringError: cryptoutilSharedMagic.ErrorInvalidRequest,
					"error_description":               "post_logout_redirect_uri not registered for client",
				})
			}
		}

		// Build redirect URL with optional state.
		redirectURL := postLogoutRedirectURI
		if state != "" {
			if parsedURI, err := url.Parse(redirectURL); err == nil {
				query := parsedURI.Query()
				query.Set(cryptoutilSharedMagic.ParamState, state)
				parsedURI.RawQuery = query.Encode()
				redirectURL = parsedURI.String()
			}
		}

		return c.Redirect(redirectURL, fiber.StatusFound)
	}

	// No redirect URI - return success page.
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)

	return c.Status(fiber.StatusOK).SendString(fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Logged Out</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
               display: flex; justify-content: center; align-items: center;
               min-height: 100vh; margin: 0; background: #f5f5f5; }
        .container { text-align: center; padding: 40px; background: white;
                     border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #333; margin-bottom: 16px; }
        p { color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Logged Out</h1>
        <p>You have been successfully logged out.</p>
        <p>%s</p>
    </div>
</body>
</html>`, getLogoutMessage(clientID)))
}

// getLogoutMessage returns a contextual message for the logout page.
func getLogoutMessage(clientID string) string {
	if clientID != "" {
		return "You may close this window or return to the application."
	}

	return "You may close this window."
}
