// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"bytes"
	"context"
	"crypto"
	"encoding/asn1"
	json "encoding/json"
	"encoding/pem"
	"io"
	http "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/sm-kms/server/pki/service/revocation"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	ocsp "golang.org/x/crypto/ocsp"
)

const (
	strictServerTestMissing       = "missing"
	strictServerTestBasePath      = "/pki"
	strictServerTestMalformedJSON = "{"
)

func TestStrictServer_AllOperations(t *testing.T) {
	t.Parallel()

	store := cryptoutilCAStorage.NewMemoryStore()
	server := NewStrictServer(&Handler{storage: store, profiles: map[string]*ProfileConfig{}})
	ctx := context.Background()

	requestID, err := googleUuid.NewV7()
	require.NoError(t, err)

	tests := []struct {
		name string
		call func() (any, error)
		want int
	}{
		{
			name: "ListCAs",
			call: func() (any, error) {
				return server.ListCAs(ctx, cryptoutilApiCaServer.ListCAsRequestObject{})
			},
			want: fiber.StatusInternalServerError,
		},
		{
			name: "GetCA",
			call: func() (any, error) {
				return server.GetCA(ctx, cryptoutilApiCaServer.GetCARequestObject{CAID: strictServerTestMissing})
			},
			want: fiber.StatusInternalServerError,
		},
		{
			name: "GetCRL",
			call: func() (any, error) {
				return server.GetCRL(ctx, cryptoutilApiCaServer.GetCRLRequestObject{CAID: strictServerTestMissing})
			},
			want: fiber.StatusServiceUnavailable,
		},
		{
			name: "ListCertificates",
			call: func() (any, error) {
				return server.ListCertificates(ctx, cryptoutilApiCaServer.ListCertificatesRequestObject{})
			},
			want: fiber.StatusOK,
		},
		{
			name: "GetCertificate",
			call: func() (any, error) {
				return server.GetCertificate(ctx, cryptoutilApiCaServer.GetCertificateRequestObject{SerialNumber: strictServerTestMissing})
			},
			want: fiber.StatusNotFound,
		},
		{
			name: "GetCertificateChain",
			call: func() (any, error) {
				return server.GetCertificateChain(ctx, cryptoutilApiCaServer.GetCertificateChainRequestObject{SerialNumber: strictServerTestMissing})
			},
			want: fiber.StatusNotFound,
		},
		{
			name: "RevokeCertificate",
			call: func() (any, error) {
				body := cryptoutilApiCaServer.RevocationRequest{Reason: cryptoutilApiCaServer.KeyCompromise}

				return server.RevokeCertificate(ctx, cryptoutilApiCaServer.RevokeCertificateRequestObject{SerialNumber: strictServerTestMissing, Body: &body})
			},
			want: fiber.StatusNotFound,
		},
		{
			name: "SubmitEnrollment",
			call: func() (any, error) {
				body := cryptoutilApiCaServer.EnrollmentRequest{}

				return server.SubmitEnrollment(ctx, cryptoutilApiCaServer.SubmitEnrollmentRequestObject{Body: &body})
			},
			want: fiber.StatusBadRequest,
		},
		{
			name: "GetEnrollmentStatus",
			call: func() (any, error) {
				return server.GetEnrollmentStatus(ctx, cryptoutilApiCaServer.GetEnrollmentStatusRequestObject{RequestID: requestID})
			},
			want: fiber.StatusNotFound,
		},
		{
			name: "ESTCACerts",
			call: func() (any, error) {
				return server.ESTCACerts(ctx, cryptoutilApiCaServer.ESTCACertsRequestObject{})
			},
			want: fiber.StatusInternalServerError,
		},
		{
			name: "ESTCSRAttrs",
			call: func() (any, error) {
				return server.ESTCSRAttrs(ctx, cryptoutilApiCaServer.ESTCSRAttrsRequestObject{})
			},
			want: fiber.StatusNoContent,
		},
		{
			name: "ESTServerKeyGen",
			call: func() (any, error) {
				return server.ESTServerKeyGen(ctx, cryptoutilApiCaServer.ESTServerKeyGenRequestObject{})
			},
			want: fiber.StatusServiceUnavailable,
		},
		{
			name: "ESTSimpleEnroll",
			call: func() (any, error) {
				return server.ESTSimpleEnroll(ctx, cryptoutilApiCaServer.ESTSimpleEnrollRequestObject{Body: strings.NewReader("")})
			},
			want: fiber.StatusBadRequest,
		},
		{
			name: "ESTSimpleReenroll",
			call: func() (any, error) {
				return server.ESTSimpleReenroll(ctx, cryptoutilApiCaServer.ESTSimpleReenrollRequestObject{Body: strings.NewReader("")})
			},
			want: fiber.StatusBadRequest,
		},
		{
			name: "HandleOCSP",
			call: func() (any, error) {
				return server.HandleOCSP(ctx, cryptoutilApiCaServer.HandleOCSPRequestObject{Body: strings.NewReader("")})
			},
			want: fiber.StatusServiceUnavailable,
		},
		{
			name: "ListProfiles",
			call: func() (any, error) {
				return server.ListProfiles(ctx, cryptoutilApiCaServer.ListProfilesRequestObject{})
			},
			want: fiber.StatusOK,
		},
		{
			name: "GetProfile",
			call: func() (any, error) {
				return server.GetProfile(ctx, cryptoutilApiCaServer.GetProfileRequestObject{ProfileID: strictServerTestMissing})
			},
			want: fiber.StatusNotFound,
		},
		{
			name: "TsaTimestamp",
			call: func() (any, error) {
				return server.TsaTimestamp(ctx, cryptoutilApiCaServer.TsaTimestampRequestObject{Body: strings.NewReader("")})
			},
			want: fiber.StatusServiceUnavailable,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			response, callErr := test.call()
			require.NoError(t, callErr)

			strict, ok := response.(strictResponse)
			require.True(t, ok, "strict handler response type = %T", response)
			require.Equal(t, test.want, strict.statusCode)
		})
	}
}

func TestStrictServer_MountedFiberPaths(t *testing.T) {
	t.Parallel()

	store := cryptoutilCAStorage.NewMemoryStore()
	strictServer := NewStrictServer(&Handler{storage: store, profiles: map[string]*ProfileConfig{}})
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	cryptoutilApiCaServer.RegisterHandlersWithOptions(
		app,
		cryptoutilApiCaServer.NewStrictHandler(strictServer, nil),
		cryptoutilApiCaServer.FiberServerOptions{
			BaseURL:     strictServerTestBasePath,
			Middlewares: []cryptoutilApiCaServer.MiddlewareFunc{StrictRequestBodyErrorMiddleware},
		},
	)

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		contentType    string
		wantStatus     int
		wantType       string
		wantParseError bool
	}{
		{name: "profiles success", method: http.MethodGet, path: "/pki/profiles", wantStatus: fiber.StatusOK, wantType: testContentTypeJSON},
		{name: "CA error", method: http.MethodGet, path: "/pki/cas", wantStatus: fiber.StatusInternalServerError, wantType: testContentTypeJSON},
		{name: "certificate not found", method: http.MethodGet, path: "/pki/certificates/missing", wantStatus: fiber.StatusNotFound, wantType: testContentTypeJSON},
		{name: "malformed enrollment JSON", method: http.MethodPost, path: "/pki/enrollments", body: strictServerTestMalformedJSON, contentType: testContentTypeJSON, wantStatus: fiber.StatusBadRequest, wantType: testContentTypeJSON, wantParseError: true},
		{name: "malformed revocation JSON", method: http.MethodPost, path: "/pki/certificates/missing/revoke", body: strictServerTestMalformedJSON, contentType: testContentTypeJSON, wantStatus: fiber.StatusBadRequest, wantType: testContentTypeJSON, wantParseError: true},
		{name: "EST attributes no content", method: http.MethodGet, path: "/pki/est/csrattrs", wantStatus: fiber.StatusNoContent},
		{name: "OCSP unavailable", method: http.MethodPost, path: "/pki/ocsp", wantStatus: fiber.StatusServiceUnavailable, wantType: cryptoutilSharedMagic.PKIContentTypeOCSPResponse},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(test.method, test.path, strings.NewReader(test.body))
			if test.contentType != "" {
				request.Header.Set("Content-Type", test.contentType)
			}

			response, err := app.Test(request, -1)
			require.NoError(t, err)
			require.Equal(t, test.wantStatus, response.StatusCode)

			if test.wantType != "" {
				require.Equal(t, test.wantType, response.Header.Get("Content-Type"))
			}

			if test.wantParseError {
				var responseBody cryptoutilApiCaServer.ErrorResponse
				require.NoError(t, json.NewDecoder(response.Body).Decode(&responseBody))
				require.Equal(t, cryptoutilSharedMagic.ErrorInvalidRequest, responseBody.Error)
				require.NotNil(t, responseBody.Message)
				require.Equal(t, cryptoutilSharedMagic.PKIMessageFailedToParseRequestBody, *responseBody.Message)
			}

			_, err = io.Copy(io.Discard, response.Body)
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())
		})
	}
}

func TestStrictServer_ESTCACertsBinaryResponse(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	strictServer := NewStrictServer(&Handler{issuer: testSetup.Issuer})
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	cryptoutilApiCaServer.RegisterHandlersWithOptions(
		app,
		cryptoutilApiCaServer.NewStrictHandler(strictServer, nil),
		cryptoutilApiCaServer.FiberServerOptions{BaseURL: strictServerTestBasePath},
	)

	request := httptest.NewRequest(http.MethodGet, "/pki/est/cacerts", nil)
	response, err := app.Test(request, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, response.StatusCode)
	require.Equal(t, cryptoutilSharedMagic.PKIContentTypePKCS7CertsOnly, response.Header.Get("Content-Type"))
	require.Equal(t, cryptoutilSharedMagic.PKIContentTransferEncodingBase64, response.Header.Get(cryptoutilSharedMagic.PKIHeaderContentTransferEncoding))

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())

	certificateBlock, _ := pem.Decode(body)
	require.NotNil(t, certificateBlock)
	require.Equal(t, testSetup.Issuer.GetCAConfig().Certificate.Raw, certificateBlock.Bytes)
}

func TestStrictServer_OCSPBinaryResponse(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	caConfig := testSetup.Issuer.GetCAConfig()
	signer, ok := caConfig.PrivateKey.(crypto.Signer)
	require.True(t, ok)

	crlService, err := cryptoutilCAServiceRevocation.NewCRLService(&cryptoutilCAServiceRevocation.CRLConfig{
		Issuer:     caConfig.Certificate,
		PrivateKey: signer,
		Provider:   testSetup.Provider,
		Validity:   time.Hour,
	})
	require.NoError(t, err)
	ocspService, err := cryptoutilCAServiceRevocation.NewOCSPService(&cryptoutilCAServiceRevocation.OCSPConfig{
		Issuer:       caConfig.Certificate,
		Responder:    caConfig.Certificate,
		ResponderKey: signer,
		Provider:     testSetup.Provider,
		Validity:     time.Hour,
	}, crlService)
	require.NoError(t, err)

	strictServer := NewStrictServer(&Handler{
		issuer:      testSetup.Issuer,
		ocspService: ocspService,
		storage:     cryptoutilCAStorage.NewMemoryStore(),
	})
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	cryptoutilApiCaServer.RegisterHandlersWithOptions(
		app,
		cryptoutilApiCaServer.NewStrictHandler(strictServer, nil),
		cryptoutilApiCaServer.FiberServerOptions{BaseURL: strictServerTestBasePath},
	)

	requestBytes, err := ocsp.CreateRequest(caConfig.Certificate, caConfig.Certificate, nil)
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "/pki/ocsp", bytes.NewReader(requestBytes))
	request.Header.Set("Content-Type", "application/ocsp-request")

	response, err := app.Test(request, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, response.StatusCode)
	require.Equal(t, cryptoutilSharedMagic.PKIContentTypeOCSPResponse, response.Header.Get("Content-Type"))

	responseBytes, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())

	parsedResponse, err := ocsp.ParseResponse(responseBytes, testSetup.RootIssuer.GetCAConfig().Certificate)
	require.NoError(t, err)
	require.Equal(t, ocsp.Unknown, parsedResponse.Status)
}

func TestStrictServer_TSABinaryResponse(t *testing.T) {
	t.Parallel()

	handler := &Handler{}
	handler.SetTSAService(createHandlerTSAService(t))
	strictServer := NewStrictServer(handler)
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	cryptoutilApiCaServer.RegisterHandlersWithOptions(
		app,
		cryptoutilApiCaServer.NewStrictHandler(strictServer, nil),
		cryptoutilApiCaServer.FiberServerOptions{BaseURL: strictServerTestBasePath},
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/pki/tsa/timestamp",
		bytes.NewReader(createValidTimestampRequest(t)),
	)
	request.Header.Set("Content-Type", "application/timestamp-query")

	response, err := app.Test(request, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, response.StatusCode)
	require.Equal(t, cryptoutilSharedMagic.PKIContentTypeTimestampReply, response.Header.Get("Content-Type"))

	responseBytes, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.NotEmpty(t, responseBytes)

	var responseASN1 asn1.RawValue

	remaining, err := asn1.Unmarshal(responseBytes, &responseASN1)
	require.NoError(t, err)
	require.Empty(t, remaining)
	require.Equal(t, asn1.TagSequence, responseASN1.Tag)
}
