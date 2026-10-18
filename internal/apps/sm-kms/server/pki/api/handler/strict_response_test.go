// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"bytes"
	"errors"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestStrictResponse_WriteBodyVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		response   strictResponse
		wantStatus int
		wantType   string
		wantHeader string
		wantBody   string
	}{
		{
			name:       "empty status",
			response:   strictStatus(http.StatusNoContent, ""),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "binary body and header",
			response:   strictBinary(http.StatusOK, cryptoutilSharedMagic.PKIContentTypeCRLDER, []byte("binary"), [2]string{"X-Strict-Test", strictHandlerTestPresent}),
			wantStatus: http.StatusOK,
			wantType:   cryptoutilSharedMagic.PKIContentTypeCRLDER,
			wantHeader: "present",
			wantBody:   "binary",
		},
		{
			name:       "reader body",
			response:   strictResponse{statusCode: http.StatusOK, body: bytes.NewBufferString("stream")},
			wantStatus: http.StatusOK,
			wantType:   "text/plain; charset=utf-8",
			wantBody:   "stream",
		},
		{
			name:       "json body",
			response:   strictJSON(http.StatusCreated, map[string]string{strictHandlerTestResultKey: "created"}),
			wantStatus: http.StatusCreated,
			wantType:   "application/json",
			wantBody:   `{"result":"created"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Get("/", tc.response.write)

			response, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil), -1)
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, response.Body.Close())
			})

			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, response.StatusCode)
			require.Equal(t, tc.wantType, response.Header.Get("Content-Type"))
			require.Equal(t, tc.wantHeader, response.Header.Get("X-Strict-Test"))
			require.Equal(t, tc.wantBody, string(body))
		})
	}
}

func TestStrictResponse_WriteErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   fiber.Config
		response strictResponse
		wantErr  string
	}{
		{
			name:     strictHandlerTestReaderError,
			config:   fiber.Config{DisableStartupMessage: true},
			response: strictResponse{statusCode: http.StatusOK, body: strictFailingReader{}},
			wantErr:  "failed to write binary response",
		},
		{
			name: "json encoding error",
			config: fiber.Config{
				DisableStartupMessage: true,
				JSONEncoder: func(any) ([]byte, error) {
					return nil, errors.New("injected JSON encoding error")
				},
			},
			response: strictJSON(http.StatusOK, map[string]string{strictHandlerTestResultKey: "failed"}),
			wantErr:  "failed to send JSON response",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var handlerError error

			config := tc.config
			config.ErrorHandler = func(ctx *fiber.Ctx, err error) error {
				handlerError = err

				return ctx.SendStatus(http.StatusInternalServerError)
			}

			app := fiber.New(config)
			app.Get("/", tc.response.write)

			response, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil), -1)
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, response.Body.Close())
			})

			require.Equal(t, http.StatusInternalServerError, response.StatusCode)
			require.ErrorContains(t, handlerError, tc.wantErr)
		})
	}
}

func TestStrictResponse_Visitors(t *testing.T) {
	t.Parallel()

	response := strictJSON(http.StatusOK, map[string]string{strictHandlerTestResultKey: "visited"})
	visitors := []struct {
		name    string
		visitFn func(*fiber.Ctx) error
	}{
		{name: strictHandlerTestListCAs, visitFn: response.VisitListCAsResponse},
		{name: strictHandlerTestGetCA, visitFn: response.VisitGetCAResponse},
		{name: "get CRL", visitFn: response.VisitGetCRLResponse},
		{name: "list certificates", visitFn: response.VisitListCertificatesResponse},
		{name: "get certificate", visitFn: response.VisitGetCertificateResponse},
		{name: "get certificate chain", visitFn: response.VisitGetCertificateChainResponse},
		{name: "revoke certificate", visitFn: response.VisitRevokeCertificateResponse},
		{name: strictHandlerTestSubmitEnrollment, visitFn: response.VisitSubmitEnrollmentResponse},
		{name: strictHandlerTestGetEnrollmentStatus, visitFn: response.VisitGetEnrollmentStatusResponse},
		{name: "EST CA certs", visitFn: response.VisitESTCACertsResponse},
		{name: "EST CSR attrs", visitFn: response.VisitESTCSRAttrsResponse},
		{name: "EST server keygen", visitFn: response.VisitESTServerKeyGenResponse},
		{name: "EST simple enroll", visitFn: response.VisitESTSimpleEnrollResponse},
		{name: "EST simple reenroll", visitFn: response.VisitESTSimpleReenrollResponse},
		{name: "OCSP", visitFn: response.VisitHandleOCSPResponse},
		{name: strictHandlerTestListProfiles, visitFn: response.VisitListProfilesResponse},
		{name: strictHandlerTestGetProfile, visitFn: response.VisitGetProfileResponse},
		{name: "TSA timestamp", visitFn: response.VisitTsaTimestampResponse},
	}

	for _, tc := range visitors {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Get("/", tc.visitFn)

			result, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil), -1)
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, result.Body.Close())
			})

			require.Equal(t, http.StatusOK, result.StatusCode)
		})
	}
}

func TestStrictResponse_ResultAndHandledResponse(t *testing.T) {
	t.Parallel()

	response := strictJSON(http.StatusOK, cryptoutilApiCaServer.ErrorResponse{})
	typed, err := strictResponseResult[cryptoutilApiCaServer.ListCAsResponseObject](response)
	require.NoError(t, err)
	require.Equal(t, response, typed)

	_, err = strictResponseResult[string](response)
	require.ErrorContains(t, err, "does not implement generated response type")

	handled, err := strictHandledResponse(response)
	require.NoError(t, err)
	require.Equal(t, response, handled)
}

type strictFailingReader struct{}

func (strictFailingReader) Read([]byte) (int, error) {
	return 0, errors.New("injected reader error")
}
