// Copyright (c) 2025 Justin Cranford

package apperr_test

import (
	"errors"
	http "net/http"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
)

const testSummaryText = "test summary"

func TestAllHTTP4xxErrorConstructors(t *testing.T) {
	t.Parallel()

	summary := testSummaryText
	baseErr := errors.New("test error")

	tests := []struct {
		name           string
		constructor    func(summary *string, err error) *cryptoutilSharedApperr.Error
		wantStatusCode int
	}{
		{"http402-payment-required", cryptoutilSharedApperr.NewHTTP402PaymentRequired, http.StatusPaymentRequired},
		{"http405-method-not-allowed", cryptoutilSharedApperr.NewHTTP405MethodNotAllowed, http.StatusMethodNotAllowed},
		{"http406-not-acceptable", cryptoutilSharedApperr.NewHTTP406NotAcceptable, http.StatusNotAcceptable},
		{"http407-proxy-auth-required", cryptoutilSharedApperr.NewHTTP407ProxyAuthRequired, http.StatusProxyAuthRequired},
		{"http408-request-timeout", cryptoutilSharedApperr.NewHTTP408RequestTimeout, http.StatusRequestTimeout},
		{"http409-conflict", cryptoutilSharedApperr.NewHTTP409Conflict, http.StatusConflict},
		{"http410-gone", cryptoutilSharedApperr.NewHTTP410Gone, http.StatusGone},
		{"http411-length-required", cryptoutilSharedApperr.NewHTTP411LengthRequired, http.StatusLengthRequired},
		{"http412-precondition-failed", cryptoutilSharedApperr.NewHTTP412PreconditionFailed, http.StatusPreconditionFailed},
		{"http413-payload-too-large", cryptoutilSharedApperr.NewHTTP413PayloadTooLarge, http.StatusRequestEntityTooLarge},
		{"http414-uri-too-long", cryptoutilSharedApperr.NewHTTP414URITooLong, http.StatusRequestURITooLong},
		{"http415-unsupported-media-type", cryptoutilSharedApperr.NewHTTP415UnsupportedMediaType, http.StatusUnsupportedMediaType},
		{"http416-range-not-satisfiable", cryptoutilSharedApperr.NewHTTP416RangeNotSatisfiable, http.StatusRequestedRangeNotSatisfiable},
		{"http417-expectation-failed", cryptoutilSharedApperr.NewHTTP417ExpectationFailed, http.StatusExpectationFailed},
		{"http418-teapot", cryptoutilSharedApperr.NewHTTP418Teapot, http.StatusTeapot},
		{"http421-misdirected-request", cryptoutilSharedApperr.NewHTTP421MisdirectedRequest, http.StatusMisdirectedRequest},
		{"http422-unprocessable-entity", cryptoutilSharedApperr.NewHTTP422UnprocessableEntity, http.StatusUnprocessableEntity},
		{"http423-locked", cryptoutilSharedApperr.NewHTTP423Locked, http.StatusLocked},
		{"http424-failed-dependency", cryptoutilSharedApperr.NewHTTP424FailedDependency, http.StatusFailedDependency},
		{"http425-too-early", cryptoutilSharedApperr.NewHTTP425TooEarly, http.StatusTooEarly},
		{"http426-upgrade-required", cryptoutilSharedApperr.NewHTTP426UpgradeRequired, http.StatusUpgradeRequired},
		{"http428-precondition-required", cryptoutilSharedApperr.NewHTTP428PreconditionRequired, http.StatusPreconditionRequired},
		{"http429-too-many-requests", cryptoutilSharedApperr.NewHTTP429TooManyRequests, http.StatusTooManyRequests},
		{"http431-request-header-fields-too-large", cryptoutilSharedApperr.NewHTTP431RequestHeaderFieldsTooLarge, http.StatusRequestHeaderFieldsTooLarge},
		{"http451-unavailable-for-legal-reasons", cryptoutilSharedApperr.NewHTTP451UnavailableForLegalReasons, http.StatusUnavailableForLegalReasons},
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
		constructor    func(summary *string, err error) *cryptoutilSharedApperr.Error
		wantStatusCode int
	}{
		{"http501-not-implemented", cryptoutilSharedApperr.NewHTTP501StatusLineAndCodeNotImplemented, http.StatusNotImplemented},
		{"http502-bad-gateway", cryptoutilSharedApperr.NewHTTP502StatusLineAndCodeBadGateway, http.StatusBadGateway},
		{"http503-service-unavailable", cryptoutilSharedApperr.NewHTTP503StatusLineAndCodeServiceUnavailable, http.StatusServiceUnavailable},
		{"http504-gateway-timeout", cryptoutilSharedApperr.NewHTTP504StatusLineAndCodeGatewayTimeout, http.StatusGatewayTimeout},
		{"http505-http-version-not-supported", cryptoutilSharedApperr.NewHTTP505StatusLineAndCodeHTTPVersionNotSupported, http.StatusHTTPVersionNotSupported},
		{"http506-variant-also-negotiates", cryptoutilSharedApperr.NewHTTP506StatusLineAndCodeVariantAlsoNegotiates, http.StatusVariantAlsoNegotiates},
		{"http507-insufficient-storage", cryptoutilSharedApperr.NewHTTP507StatusLineAndCodeInsufficientStorage, http.StatusInsufficientStorage},
		{"http508-loop-detected", cryptoutilSharedApperr.NewHTTP508StatusLineAndCodeLoopDetected, http.StatusLoopDetected},
		{"http510-not-extended", cryptoutilSharedApperr.NewHTTP510StatusLineAndCodeNotExtended, http.StatusNotExtended},
		{"http511-network-authentication-required", cryptoutilSharedApperr.NewHTTP511StatusLineAndCodeNetworkAuthenticationRequired, http.StatusNetworkAuthenticationRequired},
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
	appErr := cryptoutilSharedApperr.NewHTTP400BadRequest(&summary, baseErr)
	require.NotNil(t, appErr)
	require.Equal(t, http.StatusBadRequest, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
	require.Equal(t, "", appErr.Summary)
	require.Equal(t, baseErr, appErr.Err)
}

func TestHTTPErrorConstructorsWithNilError(t *testing.T) {
	t.Parallel()

	summary := testSummaryText

	// Test with nil error - should handle gracefully.
	appErr := cryptoutilSharedApperr.NewHTTP500InternalServerError(&summary, nil)
	require.NotNil(t, appErr)
	require.Equal(t, http.StatusInternalServerError, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
	require.Equal(t, summary, appErr.Summary)
	require.Nil(t, appErr.Err)
}

func TestHTTPErrorConstructorsWithEmptySummaryAndNilError(t *testing.T) {
	t.Parallel()

	summary := ""

	// Test with empty summary and nil error - should handle gracefully.
	appErr := cryptoutilSharedApperr.NewHTTP404NotFound(&summary, nil)
	require.NotNil(t, appErr)
	require.Equal(t, http.StatusNotFound, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
	require.Equal(t, "", appErr.Summary)
	require.Nil(t, appErr.Err)
}
