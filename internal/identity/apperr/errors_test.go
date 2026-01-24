// Copyright (c) 2025 Justin Cranford
//
//

package apperr

import (
	"errors"
	http "net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIdentityError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *IdentityError
		expected string
	}{
		{
			name: "Error without internal error",
			err: &IdentityError{
				Code:       "test_error",
				Message:    "Test error message",
				HTTPStatus: http.StatusBadRequest,
			},
			expected: "test_error: Test error message",
		},
		{
			name: "Error with internal error",
			err: &IdentityError{
				Code:       "test_error",
				Message:    "Test error message",
				HTTPStatus: http.StatusBadRequest,
				Internal:   errors.New("internal issue"),
			},
			expected: "test_error: Test error message (internal: internal issue)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.err.Error()
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestIdentityError_Unwrap(t *testing.T) {
	t.Parallel()

	internalErr := errors.New("internal error")
	err := &IdentityError{
		Code:       "test",
		Message:    "Test",
		HTTPStatus: http.StatusInternalServerError,
		Internal:   internalErr,
	}

	unwrapped := err.Unwrap()
	require.Equal(t, internalErr, unwrapped)
}

func TestIdentityError_Is(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *IdentityError
		target   error
		expected bool
	}{
		{
			name:     "Same error code",
			err:      ErrUserNotFound,
			target:   ErrUserNotFound,
			expected: true,
		},
		{
			name:     "Different error code",
			err:      ErrUserNotFound,
			target:   ErrUserDisabled,
			expected: false,
		},
		{
			name:     "Non-identity error",
			err:      ErrUserNotFound,
			target:   errors.New("standard error"),
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.err.Is(tc.target)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestNewIdentityError(t *testing.T) {
	t.Parallel()

	internalErr := errors.New("database connection lost")
	err := NewIdentityError("db_error", "Database error occurred", http.StatusInternalServerError, internalErr)

	require.Equal(t, "db_error", err.Code)
	require.Equal(t, "Database error occurred", err.Message)
	require.Equal(t, http.StatusInternalServerError, err.HTTPStatus)
	require.Equal(t, internalErr, err.Internal)
}

func TestWrapError(t *testing.T) {
	t.Parallel()

	internalErr := errors.New("connection timeout")
	wrapped := WrapError(ErrDatabaseConnection, internalErr)

	require.Equal(t, ErrDatabaseConnection.Code, wrapped.Code)
	require.Equal(t, ErrDatabaseConnection.Message, wrapped.Message)
	require.Equal(t, ErrDatabaseConnection.HTTPStatus, wrapped.HTTPStatus)
	require.Equal(t, internalErr, wrapped.Internal)
}

func TestPredefinedErrors_UserErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        *IdentityError
		code       string
		httpStatus int
	}{
		{"UserNotFound", ErrUserNotFound, "user_not_found", http.StatusNotFound},
		{"UserAlreadyExists", ErrUserAlreadyExists, "user_already_exists", http.StatusConflict},
		{"UserDisabled", ErrUserDisabled, "user_disabled", http.StatusForbidden},
		{"UserLocked", ErrUserLocked, "user_locked", http.StatusForbidden},
		{"InvalidCredentials", ErrInvalidCredentials, "invalid_credentials", http.StatusUnauthorized},
		{"PasswordHashFailed", ErrPasswordHashFailed, "password_hash_failed", http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.code, tc.err.Code)
			require.Equal(t, tc.httpStatus, tc.err.HTTPStatus)
			require.Nil(t, tc.err.Internal)
		})
	}
}

func TestPredefinedErrors_ClientErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        *IdentityError
		code       string
		httpStatus int
	}{
		{"ClientNotFound", ErrClientNotFound, "client_not_found", http.StatusNotFound},
		{"ClientAlreadyExists", ErrClientAlreadyExists, "client_already_exists", http.StatusConflict},
		{"ClientDisabled", ErrClientDisabled, "client_disabled", http.StatusForbidden},
		{"InvalidClientAuth", ErrInvalidClientAuth, "invalid_client", http.StatusUnauthorized},
		{"InvalidClientSecret", ErrInvalidClientSecret, "invalid_client_secret", http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.code, tc.err.Code)
			require.Equal(t, tc.httpStatus, tc.err.HTTPStatus)
			require.Nil(t, tc.err.Internal)
		})
	}
}

func TestPredefinedErrors_TokenErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        *IdentityError
		code       string
		httpStatus int
	}{
		{"TokenNotFound", ErrTokenNotFound, "token_not_found", http.StatusNotFound},
		{"TokenExpired", ErrTokenExpired, "token_expired", http.StatusUnauthorized},
		{"TokenRevoked", ErrTokenRevoked, "token_revoked", http.StatusUnauthorized},
		{"InvalidToken", ErrInvalidToken, "invalid_token", http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.code, tc.err.Code)
			require.Equal(t, tc.httpStatus, tc.err.HTTPStatus)
			require.Nil(t, tc.err.Internal)
		})
	}
}

func TestPredefinedErrors_SessionErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        *IdentityError
		code       string
		httpStatus int
	}{
		{"SessionNotFound", ErrSessionNotFound, "session_not_found", http.StatusNotFound},
		{"SessionExpired", ErrSessionExpired, "session_expired", http.StatusUnauthorized},
		{"SessionTerminated", ErrSessionTerminated, "session_terminated", http.StatusUnauthorized},
		{"InvalidSession", ErrInvalidSession, "invalid_session", http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.code, tc.err.Code)
			require.Equal(t, tc.httpStatus, tc.err.HTTPStatus)
			require.Nil(t, tc.err.Internal)
		})
	}
}

func TestPredefinedErrors_OAuthErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        *IdentityError
		code       string
		httpStatus int
	}{
		{"InvalidRequest", ErrInvalidRequest, "invalid_request", http.StatusBadRequest},
		{"InvalidGrant", ErrInvalidGrant, "invalid_grant", http.StatusBadRequest},
		{"UnauthorizedClient", ErrUnauthorizedClient, "unauthorized_client", http.StatusUnauthorized},
		{"AccessDenied", ErrAccessDenied, "access_denied", http.StatusForbidden},
		{"UnsupportedGrantType", ErrUnsupportedGrantType, "unsupported_grant_type", http.StatusBadRequest},
		{"InvalidScope", ErrInvalidScope, "invalid_scope", http.StatusBadRequest},
		{"ServerError", ErrServerError, "server_error", http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.code, tc.err.Code)
			require.Equal(t, tc.httpStatus, tc.err.HTTPStatus)
			require.Nil(t, tc.err.Internal)
		})
	}
}
