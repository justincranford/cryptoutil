// Copyright (c) 2025 Justin Cranford

package handler

import (
	"bytes"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"encoding/pem"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilCAStorage "cryptoutil/internal/apps/pki/ca/storage"
)

func TestListCAsNilIssuer(t *testing.T) {
	t.Parallel()

	// Tests the error path when issuer is nil.
	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  nil,
	}

	app.Get("/ca", func(c *fiber.Ctx) error {
		return handler.ListCAs(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/ca", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	// When issuer is nil, GetCAConfig returns nil, causing an internal server error.
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestGetCANilIssuer(t *testing.T) {
	t.Parallel()

	// Tests the error path when issuer is nil.
	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  nil,
	}

	app.Get("/ca/:caId", func(c *fiber.Ctx) error {
		return handler.GetCA(c, c.Params("caId"))
	})

	req := httptest.NewRequest(http.MethodGet, "/ca/test-ca", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	// When issuer is nil, GetCAConfig returns nil, causing an internal server error.
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestEstCACertsNilIssuer(t *testing.T) {
	t.Parallel()

	// Tests the error path when issuer is nil.
	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  nil,
	}

	app.Get("/est/cacerts", func(c *fiber.Ctx) error {
		return handler.EstCACerts(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/est/cacerts", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	// When issuer is nil, GetCAConfig returns nil, causing an internal server error.
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// Tests with real issuer follow.

func TestListCAsWithRealIssuer(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  testSetup.Issuer,
	}

	app.Get("/ca", func(c *fiber.Ctx) error {
		return handler.ListCAs(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/ca", nil)
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

	app.Get("/ca/:caId", func(c *fiber.Ctx) error {
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

	app.Get("/est/cacerts", func(c *fiber.Ctx) error {
		return handler.EstCACerts(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/est/cacerts", nil)
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
		"tls-server": {
			ID:          "tls-server",
			Name:        "TLS Server",
			Description: "TLS Server Certificate Profile",
			Category:    "tls",
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
			CommonName:   "test.example.com",
			Organization: []string{"Test Org"},
		},
		DNSNames: []string{"test.example.com"},
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
		req.Header.Set("Content-Type", "application/json")
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("MissingCSR", func(t *testing.T) {
		t.Parallel()

		reqBody := `{"profile": "tls-server"}`
		req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("MissingProfile", func(t *testing.T) {
		t.Parallel()

		reqBody := `{"csr": "test-csr"}`
		req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
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
		req.Header.Set("Content-Type", "application/json")
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("InvalidCSR", func(t *testing.T) {
		t.Parallel()

		reqBody := `{"csr": "invalid-csr-data", "profile": "tls-server"}`
		req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}
