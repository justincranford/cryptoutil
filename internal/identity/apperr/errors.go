// Copyright (c) 2025 Justin Cranford
//
//

// Package apperr provides application-level error definitions for the identity service.
package apperr

import (
	"fmt"
	"net/http"
)

// IdentityError represents an identity module-specific error.
type IdentityError struct {
	Code       string // Error code.
	Message    string // Error message.
	HTTPStatus int    // HTTP status code.
	Internal   error  // Internal error (if any).
}

// Error implements the error interface.
func (e *IdentityError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("%s: %s (internal: %v)", e.Code, e.Message, e.Internal)
	}

	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the internal error for error chain unwrapping.
func (e *IdentityError) Unwrap() error {
	return e.Internal
}

// Is enables error comparison for error wrapping.
func (e *IdentityError) Is(target error) bool {
	if t, ok := target.(*IdentityError); ok {
		return e.Code == t.Code
	}

	return false
}

// NewIdentityError creates a new identity error.
func NewIdentityError(code, message string, httpStatus int, internal error) *IdentityError {
	return &IdentityError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Internal:   internal,
	}
}

// Common identity errors.
var (
	// User errors.
	ErrUserNotFound       = NewIdentityError("user_not_found", "User not found", http.StatusNotFound, nil)
	ErrUserAlreadyExists  = NewIdentityError("user_already_exists", "User already exists", http.StatusConflict, nil)
	ErrUserDisabled       = NewIdentityError("user_disabled", "User account is disabled", http.StatusForbidden, nil)
	ErrUserLocked         = NewIdentityError("user_locked", "User account is locked", http.StatusForbidden, nil)
	ErrInvalidCredentials = NewIdentityError("invalid_credentials", "Invalid username or password", http.StatusUnauthorized, nil)
	ErrPasswordHashFailed = NewIdentityError("password_hash_failed", "Failed to hash password", http.StatusInternalServerError, nil)

	// Client errors.
	ErrClientNotFound        = NewIdentityError("client_not_found", "Client not found", http.StatusNotFound, nil)
	ErrClientAlreadyExists   = NewIdentityError("client_already_exists", "Client already exists", http.StatusConflict, nil)
	ErrClientDisabled        = NewIdentityError("client_disabled", "Client is disabled", http.StatusForbidden, nil)
	ErrInvalidClientAuth     = NewIdentityError("invalid_client", "Invalid client authentication", http.StatusUnauthorized, nil)
	ErrInvalidClientSecret   = NewIdentityError("invalid_client_secret", "Invalid client secret", http.StatusUnauthorized, nil)
	ErrClientProfileNotFound = NewIdentityError("client_profile_not_found", "Client profile not found", http.StatusNotFound, nil)
	ErrAuthFlowNotFound      = NewIdentityError("auth_flow_not_found", "Authorization flow not found", http.StatusNotFound, nil)
	ErrAuthProfileNotFound   = NewIdentityError("auth_profile_not_found", "Authentication profile not found", http.StatusNotFound, nil)
	ErrMFAFactorNotFound     = NewIdentityError("mfa_factor_not_found", "MFA factor not found", http.StatusNotFound, nil)

	// Token errors.
	ErrTokenNotFound         = NewIdentityError("token_not_found", "Token not found", http.StatusNotFound, nil)
	ErrTokenExpired          = NewIdentityError("token_expired", "Token has expired", http.StatusUnauthorized, nil)
	ErrTokenRevoked          = NewIdentityError("token_revoked", "Token has been revoked", http.StatusUnauthorized, nil)
	ErrInvalidToken          = NewIdentityError("invalid_token", "Invalid token", http.StatusUnauthorized, nil)
	ErrTokenIssuanceFailed   = NewIdentityError("token_issuance_failed", "Failed to issue token", http.StatusInternalServerError, nil)
	ErrTokenValidationFailed = NewIdentityError("token_validation_failed", "Failed to validate token", http.StatusInternalServerError, nil)
	ErrTokenRevocationFailed = NewIdentityError("token_revocation_failed", "Failed to revoke token", http.StatusInternalServerError, nil)

	// Key management errors.
	ErrKeyNotFound         = NewIdentityError("key_not_found", "Cryptographic key not found", http.StatusNotFound, nil)
	ErrKeyRotationFailed   = NewIdentityError("key_rotation_failed", "Key rotation failed", http.StatusInternalServerError, nil)
	ErrKeyGenerationFailed = NewIdentityError("key_generation_failed", "Key generation failed", http.StatusInternalServerError, nil)

	// Credential errors.
	ErrCredentialNotFound   = NewIdentityError("credential_not_found", "Credential not found", http.StatusNotFound, nil)
	ErrAuthenticationFailed = NewIdentityError("authentication_failed", "Authentication failed", http.StatusUnauthorized, nil)

	// Session errors.
	ErrSessionNotFound   = NewIdentityError("session_not_found", "Session not found", http.StatusNotFound, nil)
	ErrSessionExpired    = NewIdentityError("session_expired", "Session has expired", http.StatusUnauthorized, nil)
	ErrSessionTerminated = NewIdentityError("session_terminated", "Session has been terminated", http.StatusUnauthorized, nil)
	ErrInvalidSession    = NewIdentityError("invalid_session", "Invalid session", http.StatusUnauthorized, nil)

	// OAuth errors.
	ErrInvalidRequest       = NewIdentityError("invalid_request", "Invalid OAuth request", http.StatusBadRequest, nil)
	ErrInvalidGrant         = NewIdentityError("invalid_grant", "Invalid authorization grant", http.StatusBadRequest, nil)
	ErrUnauthorizedClient   = NewIdentityError("unauthorized_client", "Client is not authorized", http.StatusUnauthorized, nil)
	ErrAccessDenied         = NewIdentityError("access_denied", "Access denied", http.StatusForbidden, nil)
	ErrUnsupportedGrantType = NewIdentityError("unsupported_grant_type", "Unsupported grant type", http.StatusBadRequest, nil)
	ErrInvalidScope         = NewIdentityError("invalid_scope", "Invalid scope", http.StatusBadRequest, nil)
	ErrServerError          = NewIdentityError("server_error", "Internal server error", http.StatusInternalServerError, nil)

	// Authorization request errors.
	ErrAuthorizationRequestNotFound       = NewIdentityError("authorization_request_not_found", "Authorization request not found", http.StatusNotFound, nil)
	ErrConsentNotFound                    = NewIdentityError("consent_not_found", "Consent decision not found", http.StatusNotFound, nil)
	ErrDeviceAuthorizationNotFound        = NewIdentityError("device_authorization_not_found", "Device authorization not found", http.StatusNotFound, nil)
	ErrPushedAuthorizationRequestNotFound = NewIdentityError("pushed_authorization_request_not_found", "Pushed authorization request not found", http.StatusNotFound, nil)
	ErrRecoveryCodeNotFound               = NewIdentityError("recovery_code_not_found", "Recovery code not found", http.StatusNotFound, nil)
	ErrEmailOTPNotFound                   = NewIdentityError("email_otp_not_found", "Email OTP not found", http.StatusNotFound, nil)

	// MFA errors.
	ErrInvalidOTP        = NewIdentityError("invalid_otp", "Invalid OTP code", http.StatusUnauthorized, nil)
	ErrExpiredOTP        = NewIdentityError("expired_otp", "OTP has expired", http.StatusUnauthorized, nil)
	ErrOTPAlreadyUsed    = NewIdentityError("otp_already_used", "OTP has already been used", http.StatusUnauthorized, nil)
	ErrRateLimitExceeded = NewIdentityError("rate_limit_exceeded", "Rate limit exceeded", http.StatusTooManyRequests, nil)

	// PKCE errors.
	ErrPKCERequired         = NewIdentityError("pkce_required", "PKCE is required for this flow", http.StatusBadRequest, nil)
	ErrInvalidCodeChallenge = NewIdentityError("invalid_code_challenge", "Invalid PKCE code challenge", http.StatusBadRequest, nil)
	ErrInvalidCodeVerifier  = NewIdentityError("invalid_code_verifier", "Invalid PKCE code verifier", http.StatusBadRequest, nil)

	// OIDC errors.
	ErrInvalidIDToken     = NewIdentityError("invalid_id_token", "Invalid ID token", http.StatusUnauthorized, nil)
	ErrInvalidNonce       = NewIdentityError("invalid_nonce", "Invalid nonce", http.StatusBadRequest, nil)
	ErrInvalidRedirectURI = NewIdentityError("invalid_redirect_uri", "Invalid redirect URI", http.StatusBadRequest, nil)

	// Configuration errors.
	ErrInvalidConfiguration = NewIdentityError("invalid_configuration", "Invalid configuration", http.StatusInternalServerError, nil)
	ErrMissingConfiguration = NewIdentityError("missing_configuration", "Missing required configuration", http.StatusInternalServerError, nil)

	// Database errors.
	ErrDatabaseConnection  = NewIdentityError("database_connection", "Database connection failed", http.StatusInternalServerError, nil)
	ErrDatabaseQuery       = NewIdentityError("database_query", "Database query failed", http.StatusInternalServerError, nil)
	ErrDatabaseTransaction = NewIdentityError("database_transaction", "Database transaction failed", http.StatusInternalServerError, nil)
)

// WrapError wraps an internal error with an identity error.
func WrapError(identityErr *IdentityError, internal error) *IdentityError {
	return &IdentityError{
		Code:       identityErr.Code,
		Message:    identityErr.Message,
		HTTPStatus: identityErr.HTTPStatus,
		Internal:   internal,
	}
}
