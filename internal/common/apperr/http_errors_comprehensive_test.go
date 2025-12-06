// Copyright (c) 2025 Justin Cranford

package apperr_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppErr "cryptoutil/internal/common/apperr"
)

func TestAllHTTP4xxErrorConstructors(t *testing.T) {
	t.Parallel()

	summary := "test summary"
	baseErr := errors.New("test error")

	tests := []struct {
		name           string
		constructor    func(summary *string, err error) *cryptoutilAppErr.Error
		wantStatusCode int
	}{
		{"http402-payment-required", cryptoutilAppErr.NewHTTP402PaymentRequired, http.StatusPaymentRequired},
		{"http405-method-not-allowed", cryptoutilAppErr.NewHTTP405MethodNotAllowed, http.StatusMethodNotAllowed},
		{"http406-not-acceptable", cryptoutilAppErr.NewHTTP406NotAcceptable, http.StatusNotAcceptable},
		{"http407-proxy-auth-required", cryptoutilAppErr.NewHTTP407ProxyAuthRequired, http.StatusProxyAuthRequired},
		{"http408-request-timeout", cryptoutilAppErr.NewHTTP408RequestTimeout, http.StatusRequestTimeout},
		{"http409-conflict", cryptoutilAppErr.NewHTTP409Conflict, http.StatusConflict},
		{"http410-gone", cryptoutilAppErr.NewHTTP410Gone, http.StatusGone},
		{"http411-length-required", cryptoutilAppErr.NewHTTP411LengthRequired, http.StatusLengthRequired},
		{"http412-precondition-failed", cryptoutilAppErr.NewHTTP412PreconditionFailed, http.StatusPreconditionFailed},
		{"http413-payload-too-large", cryptoutilAppErr.NewHTTP413PayloadTooLarge, http.StatusRequestEntityTooLarge},
		{"http414-uri-too-long", cryptoutilAppErr.NewHTTP414URITooLong, http.StatusRequestURITooLong},
		{"http415-unsupported-media-type", cryptoutilAppErr.NewHTTP415UnsupportedMediaType, http.StatusUnsupportedMediaType},
		{"http416-range-not-satisfiable", cryptoutilAppErr.NewHTTP416RangeNotSatisfiable, http.StatusRequestedRangeNotSatisfiable},
		{"http417-expectation-failed", cryptoutilAppErr.NewHTTP417ExpectationFailed, http.StatusExpectationFailed},
		{"http418-teapot", cryptoutilAppErr.NewHTTP418Teapot, http.StatusTeapot},
		{"http421-misdirected-request", cryptoutilAppErr.NewHTTP421MisdirectedRequest, http.StatusMisdirectedRequest},
		{"http422-unprocessable-entity", cryptoutilAppErr.NewHTTP422UnprocessableEntity, http.StatusUnprocessableEntity},
		{"http423-locked", cryptoutilAppErr.NewHTTP423Locked, http.StatusLocked},
		{"http424-failed-dependency", cryptoutilAppErr.NewHTTP424FailedDependency, http.StatusFailedDependency},
		{"http425-too-early", cryptoutilAppErr.NewHTTP425TooEarly, http.StatusTooEarly},
		{"http426-upgrade-required", cryptoutilAppErr.NewHTTP426UpgradeRequired, http.StatusUpgradeRequired},
		{"http428-precondition-required", cryptoutilAppErr.NewHTTP428PreconditionRequired, http.StatusPreconditionRequired},
		{"http429-too-many-requests", cryptoutilAppErr.NewHTTP429TooManyRequests, http.StatusTooManyRequests},
		{"http431-request-header-fields-too-large", cryptoutilAppErr.NewHTTP431RequestHeaderFieldsTooLarge, http.StatusRequestHeaderFieldsTooLarge},
		{"http451-unavailable-for-legal-reasons", cryptoutilAppErr.NewHTTP451UnavailableForLegalReasons, http.StatusUnavailableForLegalReasons},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			appErr := tc.constructor(&summary, baseErr)
			require.NotNil(t, appErr)
			require.Equal(t, tc.wantStatusCode, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
			require.Equal(t, summary, appErr.Summary)
			require.Equal(t, baseErr, appErr.Err)
		})
	}
}

func TestAllHTTP5xxErrorConstructors(t *testing.T) {
	t.Parallel()

	summary := "server error summary"
	baseErr := errors.New("server error")

	tests := []struct {
		name           string
		constructor    func(summary *string, err error) *cryptoutilAppErr.Error
		wantStatusCode int
	}{
		{"http501-not-implemented", cryptoutilAppErr.NewHTTP501StatusLineAndCodeNotImplemented, http.StatusNotImplemented},
		{"http502-bad-gateway", cryptoutilAppErr.NewHTTP502StatusLineAndCodeBadGateway, http.StatusBadGateway},
		{"http503-service-unavailable", cryptoutilAppErr.NewHTTP503StatusLineAndCodeServiceUnavailable, http.StatusServiceUnavailable},
		{"http504-gateway-timeout", cryptoutilAppErr.NewHTTP504StatusLineAndCodeGatewayTimeout, http.StatusGatewayTimeout},
		{"http505-http-version-not-supported", cryptoutilAppErr.NewHTTP505StatusLineAndCodeHTTPVersionNotSupported, http.StatusHTTPVersionNotSupported},
		{"http506-variant-also-negotiates", cryptoutilAppErr.NewHTTP506StatusLineAndCodeVariantAlsoNegotiates, http.StatusVariantAlsoNegotiates},
		{"http507-insufficient-storage", cryptoutilAppErr.NewHTTP507StatusLineAndCodeInsufficientStorage, http.StatusInsufficientStorage},
		{"http508-loop-detected", cryptoutilAppErr.NewHTTP508StatusLineAndCodeLoopDetected, http.StatusLoopDetected},
		{"http510-not-extended", cryptoutilAppErr.NewHTTP510StatusLineAndCodeNotExtended, http.StatusNotExtended},
		{"http511-network-authentication-required", cryptoutilAppErr.NewHTTP511StatusLineAndCodeNetworkAuthenticationRequired, http.StatusNetworkAuthenticationRequired},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			appErr := tc.constructor(&summary, baseErr)
			require.NotNil(t, appErr)
			require.Equal(t, tc.wantStatusCode, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
			require.Equal(t, summary, appErr.Summary)
			require.Equal(t, baseErr, appErr.Err)
		})
	}
}

func TestHTTPErrorConstructorsWithNilSummary(t *testing.T) {
	t.Parallel()

	summary := ""
	baseErr := errors.New("test error")

	// Test with empty summary - should handle gracefully.
	appErr := cryptoutilAppErr.NewHTTP400BadRequest(&summary, baseErr)
	require.NotNil(t, appErr)
	require.Equal(t, http.StatusBadRequest, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
	require.Equal(t, "", appErr.Summary)
	require.Equal(t, baseErr, appErr.Err)
}

func TestHTTPErrorConstructorsWithNilError(t *testing.T) {
	t.Parallel()

	summary := "test summary"

	// Test with nil error - should handle gracefully.
	appErr := cryptoutilAppErr.NewHTTP500InternalServerError(&summary, nil)
	require.NotNil(t, appErr)
	require.Equal(t, http.StatusInternalServerError, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
	require.Equal(t, summary, appErr.Summary)
	require.Nil(t, appErr.Err)
}

func TestHTTPErrorConstructorsWithEmptySummaryAndNilError(t *testing.T) {
	t.Parallel()

	summary := ""

	// Test with empty summary and nil error - should handle gracefully.
	appErr := cryptoutilAppErr.NewHTTP404NotFound(&summary, nil)
	require.NotNil(t, appErr)
	require.Equal(t, http.StatusNotFound, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
	require.Equal(t, "", appErr.Summary)
	require.Nil(t, appErr.Err)
}
