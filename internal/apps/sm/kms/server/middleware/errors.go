// Copyright (c) 2025 Justin Cranford
//
//

package middleware

import (
	"fmt"

	fiber "github.com/gofiber/fiber/v2"
)

// ErrorFormat specifies the error response format.
type ErrorFormat string

const (
	// ErrorFormatOAuth2 uses standard OAuth2 error format.
	ErrorFormatOAuth2 ErrorFormat = "oauth2"

	// ErrorFormatProblem uses RFC 7807 Problem Details format.
	ErrorFormatProblem ErrorFormat = "problem"

	// ErrorFormatHybrid uses a combination of OAuth2 and Problem Details.
	ErrorFormatHybrid ErrorFormat = "hybrid"
)

// OAuth2Error represents a standard OAuth2 error response.
type OAuth2Error struct {
	// Error is the OAuth2 error code.
	Error string `json:"error"`

	// ErrorDescription provides additional details.
	ErrorDescription string `json:"error_description,omitempty"`

	// ErrorURI links to documentation about the error.
	ErrorURI string `json:"error_uri,omitempty"`
}

// ProblemDetails represents an RFC 7807 Problem Details error.
type ProblemDetails struct {
	// Type is a URI reference that identifies the problem type.
	Type string `json:"type,omitempty"`

	// Title is a short, human-readable summary.
	Title string `json:"title"`

	// Status is the HTTP status code.
	Status int `json:"status"`

	// Detail is a human-readable explanation specific to this occurrence.
	Detail string `json:"detail,omitempty"`

	// Instance is a URI reference that identifies the specific occurrence.
	Instance string `json:"instance,omitempty"`

	// Extensions can contain additional problem-specific properties.
	Extensions map[string]any `json:"-"`
}

// HybridError combines OAuth2 and Problem Details formats.
type HybridError struct {
	// OAuth2 fields.
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`

	// Problem Details fields.
	Type     string `json:"type,omitempty"`
	Title    string `json:"title,omitempty"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`

	// Extensions for additional context.
	Extensions map[string]any `json:"extensions,omitempty"`
}

// AuthErrorCode represents standard OAuth2 authentication error codes.
type AuthErrorCode string

// OAuth2 error codes.
const (
	AuthErrorInvalidRequest       AuthErrorCode = "invalid_request"
	AuthErrorUnauthorizedClient   AuthErrorCode = "unauthorized_client"
	AuthErrorAccessDenied         AuthErrorCode = "access_denied"
	AuthErrorInvalidToken         AuthErrorCode = "invalid_token"
	AuthErrorInsufficientScope    AuthErrorCode = "insufficient_scope"
	AuthErrorServerError          AuthErrorCode = "server_error"
	AuthErrorTemporarilyUnavail   AuthErrorCode = "temporarily_unavailable"
	AuthErrorInvalidGrant         AuthErrorCode = "invalid_grant"
	AuthErrorUnsupportedGrantType AuthErrorCode = "unsupported_grant_type"
)

// AuthErrorResponder creates authentication error responses.
type AuthErrorResponder struct {
	format      ErrorFormat
	detailLevel string
	baseTypeURI string
}

// NewAuthErrorResponder creates a new error responder.
func NewAuthErrorResponder(format ErrorFormat, detailLevel string) *AuthErrorResponder {
	return &AuthErrorResponder{
		format:      format,
		detailLevel: detailLevel,
		baseTypeURI: "https://cryptoutil.dev/errors/auth",
	}
}

// SendUnauthorized sends a 401 Unauthorized response.
func (r *AuthErrorResponder) SendUnauthorized(c *fiber.Ctx, code AuthErrorCode, description string) error {
	return r.sendError(c, fiber.StatusUnauthorized, code, description, nil)
}

// SendForbidden sends a 403 Forbidden response.
func (r *AuthErrorResponder) SendForbidden(c *fiber.Ctx, code AuthErrorCode, description string) error {
	return r.sendError(c, fiber.StatusForbidden, code, description, nil)
}

// SendBadRequest sends a 400 Bad Request response.
func (r *AuthErrorResponder) SendBadRequest(c *fiber.Ctx, code AuthErrorCode, description string) error {
	return r.sendError(c, fiber.StatusBadRequest, code, description, nil)
}

// SendServerError sends a 500 Internal Server Error response.
func (r *AuthErrorResponder) SendServerError(c *fiber.Ctx, description string) error {
	return r.sendError(c, fiber.StatusInternalServerError, AuthErrorServerError, description, nil)
}

// SendErrorWithExtensions sends an error with additional extensions.
func (r *AuthErrorResponder) SendErrorWithExtensions(c *fiber.Ctx, status int, code AuthErrorCode, description string, extensions map[string]any) error {
	return r.sendError(c, status, code, description, extensions)
}

// sendError sends an error response in the configured format.
func (r *AuthErrorResponder) sendError(c *fiber.Ctx, status int, code AuthErrorCode, description string, extensions map[string]any) error {
	// Adjust description based on detail level.
	adjustedDescription := r.adjustDescription(description)

	switch r.format {
	case ErrorFormatOAuth2:
		return r.sendOAuth2Error(c, status, code, adjustedDescription)
	case ErrorFormatProblem:
		return r.sendProblemDetails(c, status, code, adjustedDescription, extensions)
	case ErrorFormatHybrid:
		return r.sendHybridError(c, status, code, adjustedDescription, extensions)
	default:
		return r.sendOAuth2Error(c, status, code, adjustedDescription)
	}
}

// adjustDescription adjusts error description based on detail level.
func (r *AuthErrorResponder) adjustDescription(description string) string {
	switch r.detailLevel {
	case "minimal":
		return "" // No description in minimal mode.
	case "standard":
		return description
	case "verbose", "debug":
		return description
	default:
		return ""
	}
}

// sendOAuth2Error sends a standard OAuth2 error response.
func (r *AuthErrorResponder) sendOAuth2Error(c *fiber.Ctx, status int, code AuthErrorCode, description string) error {
	response := OAuth2Error{
		Error:            string(code),
		ErrorDescription: description,
	}

	// Set WWW-Authenticate header for 401.
	if status == fiber.StatusUnauthorized {
		c.Set("WWW-Authenticate", fmt.Sprintf(`Bearer error="%s"`, code))
	}

	c.Set("Content-Type", "application/json")

	if err := c.Status(status).JSON(response); err != nil {
		return fmt.Errorf("failed to send OAuth2 error response: %w", err)
	}

	return nil
}

// sendProblemDetails sends an RFC 7807 Problem Details response.
func (r *AuthErrorResponder) sendProblemDetails(c *fiber.Ctx, status int, code AuthErrorCode, description string, extensions map[string]any) error {
	response := ProblemDetails{
		Type:       fmt.Sprintf("%s/%s", r.baseTypeURI, code),
		Title:      codeToTitle(code),
		Status:     status,
		Detail:     description,
		Instance:   c.Path(),
		Extensions: extensions,
	}

	c.Status(status)

	// Set Content-Type AFTER status to ensure it's not overwritten.
	c.Set("Content-Type", "application/problem+json")

	if err := c.JSON(response); err != nil {
		return fmt.Errorf("failed to send Problem Details response: %w", err)
	}

	// Fiber's JSON() overwrites Content-Type, so set it again after.
	c.Set("Content-Type", "application/problem+json")

	return nil
}

// sendHybridError sends a hybrid OAuth2 + Problem Details response.
func (r *AuthErrorResponder) sendHybridError(c *fiber.Ctx, status int, code AuthErrorCode, description string, extensions map[string]any) error {
	response := HybridError{
		// OAuth2 fields.
		Error:            string(code),
		ErrorDescription: description,
		// Problem Details fields.
		Type:       fmt.Sprintf("%s/%s", r.baseTypeURI, code),
		Title:      codeToTitle(code),
		Status:     status,
		Detail:     description,
		Instance:   c.Path(),
		Extensions: extensions,
	}

	// Set WWW-Authenticate header for 401.
	if status == fiber.StatusUnauthorized {
		c.Set("WWW-Authenticate", fmt.Sprintf(`Bearer error="%s"`, code))
	}

	c.Set("Content-Type", "application/json")

	if err := c.Status(status).JSON(response); err != nil {
		return fmt.Errorf("failed to send hybrid error response: %w", err)
	}

	return nil
}

// codeToTitle converts an error code to a human-readable title.
func codeToTitle(code AuthErrorCode) string {
	switch code {
	case AuthErrorInvalidRequest:
		return "Invalid Request"
	case AuthErrorUnauthorizedClient:
		return "Unauthorized Client"
	case AuthErrorAccessDenied:
		return "Access Denied"
	case AuthErrorInvalidToken:
		return "Invalid Token"
	case AuthErrorInsufficientScope:
		return "Insufficient Scope"
	case AuthErrorServerError:
		return "Server Error"
	case AuthErrorTemporarilyUnavail:
		return "Temporarily Unavailable"
	case AuthErrorInvalidGrant:
		return "Invalid Grant"
	case AuthErrorUnsupportedGrantType:
		return "Unsupported Grant Type"
	default:
		return "Authentication Error"
	}
}

// MakeProblemDetails creates a ProblemDetails struct.
func MakeProblemDetails(typeURI string, title string, status int, detail string, instance string) ProblemDetails {
	return ProblemDetails{
		Type:     typeURI,
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	}
}

// WithExtension adds an extension to ProblemDetails.
func (p ProblemDetails) WithExtension(key string, value any) ProblemDetails {
	if p.Extensions == nil {
		p.Extensions = make(map[string]any)
	}

	p.Extensions[key] = value

	return p
}
