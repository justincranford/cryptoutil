// Copyright (c) 2025 Justin Cranford

package apperr_test

import (
	"errors"
	http "net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

func TestHTTPErrorConstructors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		constructor    func(summary *string, err error) *cryptoutilSharedApperr.Error
		wantStatusCode int
		summary        string
		baseErr        error
		wantCode       string
	}{
		{
			name:           "http400_bad_request",
			constructor:    cryptoutilSharedApperr.NewHTTP400BadRequest,
			wantStatusCode: http.StatusBadRequest,
			summary:        "Invalid request parameter",
			baseErr:        errors.New("field 'name' is required"),
			wantCode:       "CLIENT_ERROR_BAD_REQUEST",
		},
		{
			name:           "http401_unauthorized",
			constructor:    cryptoutilSharedApperr.NewHTTP401Unauthorized,
			wantStatusCode: http.StatusUnauthorized,
			summary:        "Authentication required",
			baseErr:        errors.New("missing authorization header"),
			wantCode:       "CLIENT_ERROR_UNAUTHORIZED",
		},
		{
			name:           "http403_forbidden_no_base_error",
			constructor:    cryptoutilSharedApperr.NewHTTP403Forbidden,
			wantStatusCode: http.StatusForbidden,
			summary:        "Access denied",
			baseErr:        nil,
			wantCode:       "CLIENT_ERROR_FORBIDDEN",
		},
		{
			name:           "http404_not_found",
			constructor:    cryptoutilSharedApperr.NewHTTP404NotFound,
			wantStatusCode: http.StatusNotFound,
			summary:        "Resource not found",
			baseErr:        errors.New("user with ID 123 does not exist"),
			wantCode:       "CLIENT_ERROR_NOT_FOUND",
		},
		{
			name:           "http500_internal_server_error",
			constructor:    cryptoutilSharedApperr.NewHTTP500InternalServerError,
			wantStatusCode: http.StatusInternalServerError,
			summary:        "Internal server error",
			baseErr:        errors.New("database connection failed"),
			wantCode:       "SERVER_ERROR_INTERNAL_SERVER_ERROR",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			appErr := tc.constructor(&tc.summary, tc.baseErr)

			require.NotNil(t, appErr, "Error should not be nil")
			require.Equal(t, tc.wantStatusCode, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
			require.Equal(t, tc.summary, appErr.Summary)
			require.Equal(t, tc.baseErr, appErr.Err)
			require.NotEqual(t, "", appErr.ID.String(), "ID should be generated")
			require.WithinDuration(t, time.Now().UTC(), appErr.Timestamp, 1*time.Second, "Timestamp should be recent")
			require.Equal(t, time.UTC, appErr.Timestamp.Location(), "Timestamp should be in UTC")

			errorString := appErr.Error()
			require.Contains(t, errorString, tc.wantCode, "Should contain proprietary code")
			require.Contains(t, errorString, tc.summary, "Should contain summary")
			require.Contains(t, errorString, appErr.ID.String(), "Should contain correlation ID")

			if tc.baseErr != nil {
				require.Contains(t, errorString, tc.baseErr.Error(), "Should contain underlying error message")
			} else {
				require.NotContains(t, errorString, ": ", "Should not have colon-space separator when no underlying error")
			}
		})
	}
}

func TestError_CustomError(t *testing.T) {
	t.Parallel()

	statusLineAndCode := &cryptoutilSharedApperr.HTTP418StatusLineAndCodeTeapot
	summary := "I'm a teapot"
	baseErr := errors.New("coffee brewing not supported")

	appErr := cryptoutilSharedApperr.New(statusLineAndCode, &summary, baseErr)

	require.NotNil(t, appErr)
	require.Equal(t, http.StatusTeapot, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
	require.Equal(t, summary, appErr.Summary)
	require.Equal(t, baseErr, appErr.Err)
	require.NotEqual(t, "", appErr.ID.String())
	require.WithinDuration(t, time.Now().UTC(), appErr.Timestamp, 1*time.Second)
}

func TestNewHTTPStatusLineAndCode(t *testing.T) {
	t.Parallel()

	statusCode := cryptoutilSharedApperr.HTTPStatusCode(http.StatusOK)
	appCode := cryptoutilSharedApperr.NewCode("CUSTOM_CODE")

	result := cryptoutilSharedApperr.NewHTTPStatusLineAndCode(statusCode, &appCode)

	require.Equal(t, statusCode, result.StatusLine.StatusCode)
	require.Equal(t, appCode, result.Code)
}

func TestNewHTTPStatusLine(t *testing.T) {
	t.Parallel()

	statusCode := cryptoutilSharedApperr.HTTPStatusCode(http.StatusCreated)
	reasonPhrase := cryptoutilSharedApperr.HTTPReasonPhrase("Created")

	result := cryptoutilSharedApperr.NewHTTPStatusLine(statusCode, reasonPhrase)

	require.Equal(t, statusCode, result.StatusCode)
	require.Equal(t, reasonPhrase, result.ReasonPhrase)
}

func TestNewCode(t *testing.T) {
	t.Parallel()

	message := "VALIDATION_ERROR"

	code := cryptoutilSharedApperr.NewCode(message)

	require.Equal(t, cryptoutilSharedApperr.ProprietaryAppCode(message), code)
}

func TestError_ErrorMethod_Format(t *testing.T) {
	t.Parallel()

	summary := "Test error"
	baseErr := errors.New("underlying cause")

	appErr := cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, baseErr)
	errorString := appErr.Error()

	// Should contain timestamp in RFC3339Nano format
	require.True(t, strings.Contains(errorString, "T"), "Should contain ISO 8601 timestamp with T separator")
	require.True(t, strings.Contains(errorString, "Z"), "Should contain UTC timezone indicator Z")
}
