// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"bytes"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	http "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/sm-kms/server/pki/service/issuer"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
)

func TestIssuerConfigurationErrors(t *testing.T) {
	t.Parallel()

	var unconfiguredIssuer cryptoutilCAServiceIssuer.Issuer

	tests := []struct {
		name          string
		method        string
		path          string
		issuer        *cryptoutilCAServiceIssuer.Issuer
		invoke        func(*Handler, *fiber.Ctx) error
		wantStatus    int
		wantErrorText string
	}{
		{
			name:          "list CAs without issuer",
			method:        http.MethodGet,
			path:          testPathCACollection,
			invoke:        func(handler *Handler, c *fiber.Ctx) error { return handler.ListCAs(c) },
			wantStatus:    fiber.StatusInternalServerError,
			wantErrorText: testCAIssuerNotConfiguredMessage,
		},
		{
			name:          "get CA without issuer",
			method:        http.MethodGet,
			path:          testPathCAByID,
			invoke:        func(handler *Handler, c *fiber.Ctx) error { return handler.GetCA(c, c.Params("caId")) },
			wantStatus:    fiber.StatusInternalServerError,
			wantErrorText: testCAIssuerNotConfiguredMessage,
		},
		{
			name:          "get EST CA certs without issuer",
			method:        http.MethodGet,
			path:          testPathESTCACerts,
			invoke:        func(handler *Handler, c *fiber.Ctx) error { return handler.ESTCACerts(c) },
			wantStatus:    fiber.StatusInternalServerError,
			wantErrorText: testCAIssuerNotConfiguredMessage,
		},
		{
			name:          "list CAs without CA configuration",
			method:        http.MethodGet,
			path:          testPathCACollection,
			issuer:        &unconfiguredIssuer,
			invoke:        func(handler *Handler, c *fiber.Ctx) error { return handler.ListCAs(c) },
			wantStatus:    fiber.StatusInternalServerError,
			wantErrorText: testCAConfigurationUnavailableMessage,
		},
		{
			name:          "get CA without CA configuration",
			method:        http.MethodGet,
			path:          testPathCAByID,
			issuer:        &unconfiguredIssuer,
			invoke:        func(handler *Handler, c *fiber.Ctx) error { return handler.GetCA(c, c.Params("caId")) },
			wantStatus:    fiber.StatusInternalServerError,
			wantErrorText: testCAConfigurationUnavailableMessage,
		},
		{
			name:          "get EST CA certs without CA configuration",
			method:        http.MethodGet,
			path:          testPathESTCACerts,
			issuer:        &unconfiguredIssuer,
			invoke:        func(handler *Handler, c *fiber.Ctx) error { return handler.ESTCACerts(c) },
			wantStatus:    fiber.StatusServiceUnavailable,
			wantErrorText: "CA not configured",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{issuer: tc.issuer}
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Add(tc.method, tc.path, func(c *fiber.Ctx) error {
				return tc.invoke(handler, c)
			})

			request := httptest.NewRequest(tc.method, strings.ReplaceAll(tc.path, ":caId", "test-ca"), nil)
			response, err := app.Test(request, -1)
			require.NoError(t, err)
			require.Equal(t, tc.wantStatus, response.StatusCode)

			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())
			require.Contains(t, string(body), tc.wantErrorText)
		})
	}
}

func TestListCAsWithRealIssuer(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  testSetup.Issuer,
	}

	app.Get(testPathCACollection, func(c *fiber.Ctx) error {
		return handler.ListCAs(c)
	})

	req := httptest.NewRequest(http.MethodGet, testPathCACollection, nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "test-ca")

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestGetCAWithRealIssuer(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  testSetup.Issuer,
	}

	app.Get(testPathCAByID, func(c *fiber.Ctx) error {
		return handler.GetCA(c, c.Params("caId"))
	})

	t.Run("Found", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ca/test-ca", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "test-ca")

		require.NoError(t, resp.Body.Close())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ca/unknown-ca", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestEstCACertsWithRealIssuer(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  testSetup.Issuer,
	}

	app.Get(testPathESTCACerts, func(c *fiber.Ctx) error {
		return handler.ESTCACerts(c)
	})

	req := httptest.NewRequest(http.MethodGet, testPathESTCACerts, nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "-----BEGIN CERTIFICATE-----")

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestSubmitEnrollmentWithRealIssuer(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	profiles := map[string]*ProfileConfig{
		cryptoutilSharedMagic.PKIProfileIDTLSServer: {
			ID:          cryptoutilSharedMagic.PKIProfileIDTLSServer,
			Name:        testProfileNameTLSServer,
			Description: testProfileDescriptionTLSServer,
			Category:    cryptoutilSharedMagic.PKIProfileCategoryTLS,
		},
	}

	handler := &Handler{
		storage:           mockStorage,
		issuer:            testSetup.Issuer,
		profiles:          profiles,
		enrollmentTracker: newEnrollmentTracker(cryptoutilSharedMagic.JoseJAMaxMaterials),
	}

	app.Post("/enrollments", func(c *fiber.Ctx) error {
		return handler.SubmitEnrollment(c)
	})

	// Generate a test CSR.
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   testDomainNameExample,
			Organization: []string{testOrganizationTest},
		},
		DNSNames: []string{testDomainNameExample},
	}
	csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, key)
	require.NoError(t, err)

	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCSR,
		Bytes: csrDER,
	})

	t.Run("SuccessfulEnrollment", func(t *testing.T) {
		t.Parallel()

		reqBody := `{"csr": "` + string(csrPEM) + `", "profile": "tls-server"}`
		// Escape newlines for JSON.
		reqBodyEscaped := bytes.ReplaceAll([]byte(reqBody), []byte("\n"), []byte("\\n"))

		req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewReader(reqBodyEscaped))
		req.Header.Set("Content-Type", testContentTypeJSON)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("MissingCSR", func(t *testing.T) {
		t.Parallel()

		reqBody := `{"profile": "tls-server"}`
		req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", testContentTypeJSON)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("MissingProfile", func(t *testing.T) {
		t.Parallel()

		reqBody := `{"csr": "test-csr"}`
		req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", testContentTypeJSON)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("UnknownProfile", func(t *testing.T) {
		t.Parallel()

		reqBody := `{"csr": "` + string(csrPEM) + `", "profile": "unknown-profile"}`
		reqBody = string(bytes.ReplaceAll([]byte(reqBody), []byte("\n"), []byte("\\n")))

		req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", testContentTypeJSON)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("InvalidCSR", func(t *testing.T) {
		t.Parallel()

		reqBody := `{"csr": "invalid-csr-data", "profile": "tls-server"}`
		req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", testContentTypeJSON)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestRootIssuerCAResponses(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	tests := []struct {
		name   string
		invoke func(*Handler, *fiber.Ctx) error
	}{
		{
			name:   "list root CA",
			invoke: func(handler *Handler, c *fiber.Ctx) error { return handler.ListCAs(c) },
		},
		{
			name: "get root CA",
			invoke: func(handler *Handler, c *fiber.Ctx) error {
				return handler.GetCA(c, handler.issuer.GetCAConfig().Name)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{issuer: testSetup.RootIssuer}
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Get(testPathCACollection, func(c *fiber.Ctx) error {
				return tc.invoke(handler, c)
			})

			request := httptest.NewRequest(http.MethodGet, testPathCACollection, nil)
			response, err := app.Test(request, -1)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusOK, response.StatusCode)

			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())
			require.Contains(t, string(body), "root")
		})
	}
}
