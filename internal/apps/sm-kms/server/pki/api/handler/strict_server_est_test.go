// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"bytes"
	"context"
	"crypto"
	"encoding/pem"
	"errors"
	"io"
	http "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/sm-kms/server/pki/service/issuer"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/sm-kms/server/pki/service/revocation"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	ocsp "golang.org/x/crypto/ocsp"
)

func TestStrictServer_ESTSimpleEnrollPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       func(*testing.T) io.Reader
		profile    bool
		issuer     bool
		storageErr bool
		wantStatus int
		wantError  string
	}{
		{name: strictHandlerTestMissingBody, body: func(*testing.T) io.Reader { return nil }, profile: true, issuer: true, wantStatus: http.StatusBadRequest, wantError: strictHandlerTestBadRequest},
		{name: "reader failure", body: func(*testing.T) io.Reader { return strictFailingReader{} }, profile: true, issuer: true, wantStatus: http.StatusBadRequest, wantError: strictHandlerTestInvalidCSR},
		{name: strictHandlerTestEmptyBody, body: func(*testing.T) io.Reader { return strings.NewReader("") }, profile: true, issuer: true, wantStatus: http.StatusBadRequest, wantError: strictHandlerTestBadRequest},
		{name: testCaseNameInvalidCSR, body: func(*testing.T) io.Reader { return strings.NewReader("invalid CSR") }, profile: true, issuer: true, wantStatus: http.StatusBadRequest, wantError: strictHandlerTestInvalidCSR},
		{name: "invalid signature", body: strictInvalidCSRSignatureReader, profile: true, issuer: true, wantStatus: http.StatusBadRequest, wantError: "invalid_signature"},
		{name: "no profile", body: strictValidCSRReader, issuer: true, wantStatus: http.StatusServiceUnavailable, wantError: "no_profile"},
		{name: "no issuer", body: strictValidCSRReader, profile: true, wantStatus: http.StatusServiceUnavailable, wantError: strictHandlerTestServiceUnavailable},
		{name: strictHandlerTestStorageError, body: strictValidCSRReader, profile: true, issuer: true, storageErr: true, wantStatus: http.StatusInternalServerError, wantError: "storage_error"},
		{name: "successful enrollment", body: strictValidCSRReader, profile: true, issuer: true, wantStatus: http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				issuerSetup *testIssuerSetup
				storage     cryptoutilCAStorage.Store = cryptoutilCAStorage.NewMemoryStore()
			)

			if tc.issuer || tc.storageErr {
				issuerSetup = createTestIssuer(t)
			}

			if tc.storageErr {
				storage = &errorStore{
					inner:    storage,
					storeErr: errors.New("injected EST storage failure"),
				}
			}

			handler := &Handler{storage: storage}
			if tc.profile {
				handler.profiles = map[string]*ProfileConfig{
					cryptoutilSharedMagic.PKIProfileIDTLSServer: {
						ID: cryptoutilSharedMagic.PKIProfileIDTLSServer,
					},
				}
			}

			if tc.issuer {
				handler.issuer = issuerSetup.Issuer
			}

			response := callStrictESTEnroll(t, NewStrictServer(handler), tc.body(t), false, tc.wantStatus)

			if tc.wantError != "" {
				require.Equal(t, tc.wantError, strictResponseBody[cryptoutilApiCaServer.ErrorResponse](t, response).Error)

				return
			}

			require.Equal(t, cryptoutilSharedMagic.PKIContentTypePKCS7CertsOnly, response.contentType)
			require.Equal(t, cryptoutilSharedMagic.PKIContentTransferEncodingBase64, response.headers[0][1])
			require.NotEmpty(t, response.body)
		})
	}
}

func TestStrictServer_ESTSimpleReenrollSuccess(t *testing.T) {
	t.Parallel()

	_, handler, strict := newStrictServerFixture(t)
	response := callStrictESTEnroll(t, strict, strings.NewReader(createValidCSRPEM(t)), true, http.StatusOK)
	require.Equal(t, cryptoutilSharedMagic.PKIContentTypePKCS7CertsOnly, response.contentType)
	require.NotEmpty(t, response.body)

	certificates, total, err := handler.storage.List(context.Background(), &cryptoutilCAStorage.ListFilter{Limit: 10})
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, certificates, 1)
}

func TestStrictServer_ESTCACertsConfiguration(t *testing.T) {
	t.Parallel()

	setup := createTestIssuer(t)
	strict := NewStrictServer(&Handler{issuer: setup.Issuer})
	response := assertStrictCall(t, http.StatusOK, func() (any, error) {
		return strict.ESTCACerts(strictTestContext(), cryptoutilApiCaServer.ESTCACertsRequestObject{})
	})
	require.Equal(t, cryptoutilSharedMagic.PKIContentTypePKCS7CertsOnly, response.contentType)
	require.Equal(t, cryptoutilSharedMagic.PKIContentTransferEncodingBase64, response.headers[0][1])

	unconfigured := NewStrictServer(&Handler{issuer: &cryptoutilCAServiceIssuer.Issuer{}})
	errorResponse := assertStrictCall(t, http.StatusServiceUnavailable, func() (any, error) {
		return unconfigured.ESTCACerts(strictTestContext(), cryptoutilApiCaServer.ESTCACertsRequestObject{})
	})
	require.Equal(t, strictHandlerTestServiceUnavailable, strictResponseBody[cryptoutilApiCaServer.ErrorResponse](t, errorResponse).Error)
}

func TestStrictServer_HandleOCSPRequestErrors(t *testing.T) {
	t.Parallel()

	setup := createTestIssuer(t)
	ocspService := newStrictOCSPService(t, setup)

	tests := []struct {
		name       string
		body       io.Reader
		issuer     bool
		wantStatus int
	}{
		{name: "missing body", body: nil, issuer: true, wantStatus: http.StatusBadRequest},
		{name: strictHandlerTestReaderError, body: strictFailingReader{}, issuer: true, wantStatus: http.StatusBadRequest},
		{name: strictHandlerTestEmptyBody, body: strings.NewReader(""), issuer: true, wantStatus: http.StatusBadRequest},
		{name: "invalid OCSP request", body: strings.NewReader("not an OCSP request"), issuer: true, wantStatus: http.StatusInternalServerError},
		{name: strictHandlerTestIssuerAbsent, body: strictValidOCSPRequest(t, setup), wantStatus: http.StatusServiceUnavailable},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{
				storage:     cryptoutilCAStorage.NewMemoryStore(),
				issuer:      setup.Issuer,
				ocspService: ocspService,
			}
			if !tc.issuer {
				handler.issuer = nil
			}

			response := assertStrictCall(t, tc.wantStatus, func() (any, error) {
				return NewStrictServer(handler).HandleOCSP(strictTestContext(), cryptoutilApiCaServer.HandleOCSPRequestObject{Body: tc.body})
			})
			require.Equal(t, cryptoutilSharedMagic.PKIContentTypeOCSPResponse, response.contentType)
		})
	}
}

func TestStrictServer_TsaTimestampRequestErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       io.Reader
		wantStatus int
		wantError  string
	}{
		{name: strictHandlerTestMissingBody, body: nil, wantStatus: http.StatusBadRequest, wantError: strictHandlerTestBadRequest},
		{name: strictHandlerTestReaderError, body: strictFailingReader{}, wantStatus: http.StatusBadRequest, wantError: cryptoutilSharedMagic.ErrorInvalidRequest},
		{name: strictHandlerTestEmptyBody, body: strings.NewReader(""), wantStatus: http.StatusBadRequest, wantError: strictHandlerTestBadRequest},
		{name: "invalid request", body: strings.NewReader("invalid timestamp request"), wantStatus: http.StatusBadRequest, wantError: cryptoutilSharedMagic.ErrorInvalidRequest},
		{name: "valid request", body: bytes.NewReader(createValidTimestampRequest(t)), wantStatus: http.StatusOK},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{tsaService: createHandlerTSAService(t)}
			response := assertStrictCall(t, tc.wantStatus, func() (any, error) {
				return NewStrictServer(handler).TsaTimestamp(strictTestContext(), cryptoutilApiCaServer.TsaTimestampRequestObject{Body: tc.body})
			})

			if tc.wantStatus == http.StatusOK {
				require.Equal(t, cryptoutilSharedMagic.PKIContentTypeTimestampReply, response.contentType)
				require.NotEmpty(t, response.body)
			} else {
				require.Equal(t, tc.wantError, strictResponseBody[cryptoutilApiCaServer.ErrorResponse](t, response).Error)
			}
		})
	}
}

func TestStrictRequestBodyErrorMiddleware_Branches(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		route      string
		path       string
		statusCode int
		wantStatus int
		wantParse  bool
	}{
		{name: "enrollment parser error", method: http.MethodPost, route: strictHandlerTestEnrollmentPath, path: strictHandlerTestEnrollmentPath, statusCode: http.StatusBadRequest, wantStatus: http.StatusBadRequest, wantParse: true},
		{name: "revocation parser error", method: http.MethodPost, route: "/pki/certificates/:serialNumber/revoke", path: "/pki/certificates/test/revoke", statusCode: http.StatusBadRequest, wantStatus: http.StatusBadRequest, wantParse: true},
		{name: "wrong status", method: http.MethodPost, route: strictHandlerTestEnrollmentPath, path: strictHandlerTestEnrollmentPath, statusCode: http.StatusConflict, wantStatus: http.StatusConflict},
		{name: "wrong method", method: http.MethodGet, route: strictHandlerTestEnrollmentPath, path: strictHandlerTestEnrollmentPath, statusCode: http.StatusBadRequest, wantStatus: http.StatusBadRequest},
		{name: "wrong path", method: http.MethodPost, route: strictHandlerTestOtherEnrollmentPath, path: strictHandlerTestOtherEnrollmentPath, statusCode: http.StatusBadRequest, wantStatus: http.StatusBadRequest},
		{name: "wrong suffix", method: http.MethodPost, route: strictHandlerTestOtherPKIPath, path: strictHandlerTestOtherPKIPath, statusCode: http.StatusBadRequest, wantStatus: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{
				DisableStartupMessage: true,
				ErrorHandler: func(ctx *fiber.Ctx, _ error) error {
					return ctx.SendStatus(tc.statusCode)
				},
			})
			app.Use(StrictRequestBodyErrorMiddleware)
			app.Add(tc.method, tc.route, func(*fiber.Ctx) error {
				return fiber.NewError(tc.statusCode, "injected route failure")
			})

			response, err := app.Test(httptest.NewRequest(tc.method, tc.path, nil), -1)
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, response.Body.Close())
			})

			require.Equal(t, tc.wantStatus, response.StatusCode)

			if tc.wantParse {
				body, err := io.ReadAll(response.Body)
				require.NoError(t, err)
				require.Contains(t, string(body), cryptoutilSharedMagic.PKIMessageFailedToParseRequestBody)
			}
		})
	}
}

func callStrictESTEnroll(t *testing.T, strict *StrictServer, body io.Reader, reenroll bool, statusCode int) strictResponse {
	t.Helper()

	return assertStrictCall(t, statusCode, func() (any, error) {
		if reenroll {
			return strict.ESTSimpleReenroll(strictTestContext(), cryptoutilApiCaServer.ESTSimpleReenrollRequestObject{Body: body})
		}

		return strict.ESTSimpleEnroll(strictTestContext(), cryptoutilApiCaServer.ESTSimpleEnrollRequestObject{Body: body})
	})
}

func strictValidCSRReader(t *testing.T) io.Reader {
	t.Helper()

	return strings.NewReader(createValidCSRPEM(t))
}

func strictInvalidCSRSignatureReader(t *testing.T) io.Reader {
	t.Helper()

	block, _ := pem.Decode([]byte(createValidCSRPEM(t)))
	require.NotNil(t, block)
	block.Bytes[len(block.Bytes)-1] ^= 1

	return bytes.NewReader(pem.EncodeToMemory(block))
}

func strictValidOCSPRequest(t *testing.T, setup *testIssuerSetup) io.Reader {
	t.Helper()

	request, err := ocsp.CreateRequest(setup.Issuer.GetCAConfig().Certificate, setup.Issuer.GetCAConfig().Certificate, nil)
	require.NoError(t, err)

	return bytes.NewReader(request)
}

func newStrictOCSPService(t *testing.T, setup *testIssuerSetup) *cryptoutilCAServiceRevocation.OCSPService {
	t.Helper()

	caConfig := setup.Issuer.GetCAConfig()
	signer, ok := caConfig.PrivateKey.(crypto.Signer)
	require.True(t, ok)
	crlService := newStrictCRLService(t, setup)
	service, err := cryptoutilCAServiceRevocation.NewOCSPService(&cryptoutilCAServiceRevocation.OCSPConfig{
		Issuer:       caConfig.Certificate,
		Responder:    caConfig.Certificate,
		ResponderKey: signer,
		Provider:     setup.Provider,
		Validity:     time.Hour,
	}, crlService)
	require.NoError(t, err)

	return service
}
