// Copyright (c) 2025 Justin Cranford

package apperr_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cryptoutil/internal/common/apperr"
)

func TestNewHTTP400BadRequest(t *testing.T) {
	t.Parallel()

	summary := "Invalid request parameter"
	baseErr := errors.New("field 'name' is required")

	appErr := apperr.NewHTTP400BadRequest(&summary, baseErr)

	require.NotNil(t, appErr, "Error should not be nil")
	require.Equal(t, http.StatusBadRequest, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
	require.Equal(t, summary, appErr.Summary)
	require.Equal(t, baseErr, appErr.Err)
	require.NotEqual(t, "", appErr.ID.String(), "ID should be generated")
	require.WithinDuration(t, time.Now().UTC(), appErr.Timestamp, 1*time.Second, "Timestamp should be recent")
}

func TestNewHTTP401Unauthorized(t *testing.T) {
	t.Parallel()

	summary := "Authentication required"
	baseErr := errors.New("missing authorization header")

	appErr := apperr.NewHTTP401Unauthorized(&summary, baseErr)

	require.NotNil(t, appErr)
	require.Equal(t, http.StatusUnauthorized, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
	require.Equal(t, summary, appErr.Summary)
	require.Equal(t, baseErr, appErr.Err)
}

func TestNewHTTP403Forbidden(t *testing.T) {
	t.Parallel()

	summary := "Access denied"

	appErr := apperr.NewHTTP403Forbidden(&summary, nil) // No underlying error

	require.NotNil(t, appErr)
	require.Equal(t, http.StatusForbidden, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
	require.Equal(t, summary, appErr.Summary)
	require.Nil(t, appErr.Err, "Underlying error should be nil")
}

func TestNewHTTP404NotFound(t *testing.T) {
	t.Parallel()

	summary := "Resource not found"
	baseErr := errors.New("user with ID 123 does not exist")

	appErr := apperr.NewHTTP404NotFound(&summary, baseErr)

	require.NotNil(t, appErr)
	require.Equal(t, http.StatusNotFound, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
	require.Equal(t, summary, appErr.Summary)
	require.Equal(t, baseErr, appErr.Err)
}

func TestNewHTTP500InternalServerError(t *testing.T) {
	t.Parallel()

	summary := "Internal server error"
	baseErr := errors.New("database connection failed")

	appErr := apperr.NewHTTP500InternalServerError(&summary, baseErr)

	require.NotNil(t, appErr)
	require.Equal(t, http.StatusInternalServerError, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
	require.Equal(t, summary, appErr.Summary)
	require.Equal(t, baseErr, appErr.Err)
}

func TestError_ErrorMethod_WithUnderlyingError(t *testing.T) {
	t.Parallel()

	summary := "Bad request"
	baseErr := errors.New("validation failed")

	appErr := apperr.NewHTTP400BadRequest(&summary, baseErr)
	errorString := appErr.Error()

	// Verify error string format
	require.Contains(t, errorString, "CLIENT_ERROR_BAD_REQUEST", "Should contain proprietary code")
	require.Contains(t, errorString, summary, "Should contain summary")
	require.Contains(t, errorString, appErr.ID.String(), "Should contain correlation ID")
	require.Contains(t, errorString, baseErr.Error(), "Should contain underlying error message")
}

func TestError_ErrorMethod_WithoutUnderlyingError(t *testing.T) {
	t.Parallel()

	summary := "Not found"

	appErr := apperr.NewHTTP404NotFound(&summary, nil)
	errorString := appErr.Error()

	// Verify error string format without underlying error
	require.Contains(t, errorString, "CLIENT_ERROR_NOT_FOUND", "Should contain proprietary code")
	require.Contains(t, errorString, summary, "Should contain summary")
	require.Contains(t, errorString, appErr.ID.String(), "Should contain correlation ID")
	require.NotContains(t, errorString, ": ", "Should not have colon-space separator when no underlying error")
}

func TestNew_CustomError(t *testing.T) {
	t.Parallel()

	statusLineAndCode := &apperr.HTTP418StatusLineAndCodeTeapot
	summary := "I'm a teapot"
	baseErr := errors.New("coffee brewing not supported")

	appErr := apperr.New(statusLineAndCode, &summary, baseErr)

	require.NotNil(t, appErr)
	require.Equal(t, http.StatusTeapot, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
	require.Equal(t, summary, appErr.Summary)
	require.Equal(t, baseErr, appErr.Err)
	require.NotEqual(t, "", appErr.ID.String())
	require.WithinDuration(t, time.Now().UTC(), appErr.Timestamp, 1*time.Second)
}

func TestNewHTTPStatusLineAndCode(t *testing.T) {
	t.Parallel()

	statusCode := apperr.HTTPStatusCode(http.StatusOK)
	appCode := apperr.NewCode("CUSTOM_CODE")

	result := apperr.NewHTTPStatusLineAndCode(statusCode, &appCode)

	require.Equal(t, statusCode, result.StatusLine.StatusCode)
	require.Equal(t, appCode, result.Code)
}

func TestNewHTTPStatusLine(t *testing.T) {
	t.Parallel()

	statusCode := apperr.HTTPStatusCode(http.StatusCreated)
	reasonPhrase := apperr.HTTPReasonPhrase("Created")

	result := apperr.NewHTTPStatusLine(statusCode, reasonPhrase)

	require.Equal(t, statusCode, result.StatusCode)
	require.Equal(t, reasonPhrase, result.ReasonPhrase)
}

func TestNewCode(t *testing.T) {
	t.Parallel()

	message := "VALIDATION_ERROR"

	code := apperr.NewCode(message)

	require.Equal(t, apperr.ProprietaryAppCode(message), code)
}

func TestError_TimestampInUTC(t *testing.T) {
	t.Parallel()

	summary := "Test error"

	appErr := apperr.NewHTTP500InternalServerError(&summary, nil)

	// Verify timestamp is in UTC
	require.Equal(t, time.UTC, appErr.Timestamp.Location(), "Timestamp should be in UTC")
}

func TestError_ErrorMethod_Format(t *testing.T) {
	t.Parallel()

	summary := "Test error"
	baseErr := errors.New("underlying cause")

	appErr := apperr.NewHTTP400BadRequest(&summary, baseErr)
	errorString := appErr.Error()

	// Should contain timestamp in RFC3339Nano format
	require.True(t, strings.Contains(errorString, "T"), "Should contain ISO 8601 timestamp with T separator")
	require.True(t, strings.Contains(errorString, "Z"), "Should contain UTC timezone indicator Z")
}
