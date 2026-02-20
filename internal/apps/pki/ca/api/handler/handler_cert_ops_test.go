// Copyright (c) 2025 Justin Cranford

package handler

import (
	"bytes"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilApiCaServer "cryptoutil/api/ca/server"
	cryptoutilCAStorage "cryptoutil/internal/apps/pki/ca/storage"
)
func TestEstEndpoints(t *testing.T) {
	t.Parallel()

	// Create a Fiber app for testing EST endpoints.
	app := fiber.New()

	mockStorage := cryptoutilCAStorage.NewMemoryStore()
	handler := &Handler{
		storage: mockStorage,
	}

	app.Get("/est/csrattrs", func(c *fiber.Ctx) error {
		return handler.EstCSRAttrs(c)
	})

	app.Post("/est/simpleenroll", func(c *fiber.Ctx) error {
		return handler.EstSimpleEnroll(c)
	})

	app.Post("/est/simplereenroll", func(c *fiber.Ctx) error {
		return handler.EstSimpleReenroll(c)
	})

	app.Post("/est/serverkeygen", func(c *fiber.Ctx) error {
		return handler.EstServerKeyGen(c)
	})

	t.Run("EST_CSRAttrs_returns_no_content", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/est/csrattrs", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNoContent, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("EST_SimpleEnroll_empty_body", func(t *testing.T) {
		t.Parallel()

		// With empty body, should return bad request (endpoint is implemented).
		req := httptest.NewRequest(http.MethodPost, "/est/simpleenroll", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("EST_SimpleReenroll_empty_body", func(t *testing.T) {
		t.Parallel()

		// With empty body, should return bad request (endpoint is implemented).
		req := httptest.NewRequest(http.MethodPost, "/est/simplereenroll", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("EST_ServerKeyGen_empty_body", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/est/serverkeygen", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestTsaTimestamp(t *testing.T) {
	t.Parallel()

	// Create a Fiber app for testing.
	app := fiber.New()

	mockStorage := cryptoutilCAStorage.NewMemoryStore()
	handler := &Handler{
		storage: mockStorage,
	}

	app.Post("/tsa/timestamp", func(c *fiber.Ctx) error {
		return handler.TsaTimestamp(c)
	})

	t.Run("TSA_no_service_configured", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/tsa/timestamp", bytes.NewReader([]byte("timestamp request")))
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestHandleOCSP(t *testing.T) {
	t.Parallel()

	// Create a Fiber app for testing.
	app := fiber.New()

	mockStorage := cryptoutilCAStorage.NewMemoryStore()
	handler := &Handler{
		storage: mockStorage,
	}

	app.Post("/ocsp", func(c *fiber.Ctx) error {
		return handler.HandleOCSP(c)
	})

	t.Run("OCSP_no_service_configured", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/ocsp", bytes.NewReader([]byte("ocsp request")))
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestGetCRL(t *testing.T) {
	t.Parallel()

	// Create a Fiber app for testing.
	app := fiber.New()

	mockStorage := cryptoutilCAStorage.NewMemoryStore()
	handler := &Handler{
		storage: mockStorage,
	}

	app.Get("/ca/:caId/crl", func(c *fiber.Ctx) error {
		return handler.GetCRL(c, c.Params("caId"), cryptoutilApiCaServer.GetCRLParams{})
	})

	t.Run("CRL_no_service_configured", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ca/test-ca/crl", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestListProfiles(t *testing.T) {
	t.Parallel()

	// Create a Fiber app for testing.
	app := fiber.New()

	profiles := map[string]*ProfileConfig{
		"tls-server": {
			ID:          "tls-server",
			Name:        "TLS Server",
			Description: "TLS Server Certificate Profile",
			Category:    "tls",
		},
		"code-signing": {
			ID:          "code-signing",
			Name:        "Code Signing",
			Description: "Code Signing Certificate Profile",
			Category:    "code_signing",
		},
	}

	handler := &Handler{
		profiles: profiles,
	}

	app.Get("/profiles", func(c *fiber.Ctx) error {
		return handler.ListProfiles(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/profiles", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.NoError(t, resp.Body.Close())

	require.Contains(t, string(body), "tls-server")
	require.Contains(t, string(body), "code-signing")
}

func TestGetProfile(t *testing.T) {
	t.Parallel()

	// Create a Fiber app for testing.
	app := fiber.New()

	profiles := map[string]*ProfileConfig{
		"tls-server": {
			ID:          "tls-server",
			Name:        "TLS Server",
			Description: "TLS Server Certificate Profile",
			Category:    "tls",
		},
	}

	handler := &Handler{
		profiles: profiles,
	}

	app.Get("/profiles/:profileId", func(c *fiber.Ctx) error {
		return handler.GetProfile(c, c.Params("profileId"))
	})

	t.Run("GetProfile_exists", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/profiles/tls-server", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.NoError(t, resp.Body.Close())

		require.Contains(t, string(body), "tls-server")
	})

	t.Run("GetProfile_not_found", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/profiles/unknown-profile", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestGetKeyInfoWithRSA4096(t *testing.T) {
	t.Parallel()

	// Test RSA 4096 key.
	key, err := rsa.GenerateKey(crand.Reader, 4096)
	require.NoError(t, err)

	cert := &x509.Certificate{PublicKey: &key.PublicKey}
	algo, size := getKeyInfo(cert)

	require.Equal(t, "RSA", algo)
	require.Equal(t, 4096, size)
}

func TestEnrollmentTracker(t *testing.T) {
	t.Parallel()

	t.Run("NewEnrollmentTracker", func(t *testing.T) {
		t.Parallel()

		tracker := newEnrollmentTracker(100)
		require.NotNil(t, tracker)
		require.NotNil(t, tracker.requests)
		require.Equal(t, 100, tracker.maxEntries)
	})

	t.Run("TrackAndGet", func(t *testing.T) {
		t.Parallel()

		tracker := newEnrollmentTracker(100)

		requestID := googleUuid.New()
		status := cryptoutilApiCaServer.EnrollmentStatusResponseStatusIssued
		serialNumber := "ABC123"

		tracker.track(requestID, status, serialNumber)

		entry, found := tracker.get(requestID)
		require.True(t, found)
		require.Equal(t, requestID, entry.RequestID)
		require.Equal(t, status, entry.Status)
		require.Equal(t, serialNumber, entry.SerialNumber)
	})

	t.Run("GetNotFound", func(t *testing.T) {
		t.Parallel()

		tracker := newEnrollmentTracker(100)

		requestID := googleUuid.New()
		entry, found := tracker.get(requestID)
		require.False(t, found)
		require.Nil(t, entry)
	})

	t.Run("MaxEntriesEnforced", func(t *testing.T) {
		t.Parallel()

		maxEntries := 3
		tracker := newEnrollmentTracker(maxEntries)

		// Add maxEntries items.
		ids := make([]googleUuid.UUID, maxEntries+1)
		for i := 0; i < maxEntries; i++ {
			ids[i] = googleUuid.New()
			tracker.track(ids[i], cryptoutilApiCaServer.EnrollmentStatusResponseStatusIssued, "serial")
			time.Sleep(time.Millisecond) // Ensure different timestamps.
		}

		// Add one more (should evict oldest).
		ids[maxEntries] = googleUuid.New()
		tracker.track(ids[maxEntries], cryptoutilApiCaServer.EnrollmentStatusResponseStatusIssued, "serial")

		// First entry should be evicted.
		_, found := tracker.get(ids[0])
		require.False(t, found, "oldest entry should be evicted")

		// Last entry should exist.
		_, found = tracker.get(ids[maxEntries])
		require.True(t, found, "newest entry should exist")
	})
}

func TestParseESTCSR(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	t.Run("ValidPEMCSR", func(t *testing.T) {
		t.Parallel()

		// Generate a test CSR.
		key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
		require.NoError(t, err)

		template := &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName: "test.example.com",
			},
		}
		csrDER, err := x509.CreateCertificateRequest(crand.Reader, template, key)
		require.NoError(t, err)

		csrPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE REQUEST",
			Bytes: csrDER,
		})

		csr, err := handler.parseESTCSR(csrPEM)
		require.NoError(t, err)
		require.NotNil(t, csr)
		require.Equal(t, "test.example.com", csr.Subject.CommonName)
	})

	t.Run("ValidDERCSR", func(t *testing.T) {
		t.Parallel()

		// Generate a test CSR.
		key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
		require.NoError(t, err)

		template := &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName: "test.example.com",
			},
		}
		csrDER, err := x509.CreateCertificateRequest(crand.Reader, template, key)
		require.NoError(t, err)

		csr, err := handler.parseESTCSR(csrDER)
		require.NoError(t, err)
		require.NotNil(t, csr)
	})

	t.Run("InvalidCSR", func(t *testing.T) {
		t.Parallel()

		csr, err := handler.parseESTCSR([]byte("invalid data"))
		require.Error(t, err)
		require.Nil(t, csr)
		require.Contains(t, err.Error(), "invalid format")
	})
}
