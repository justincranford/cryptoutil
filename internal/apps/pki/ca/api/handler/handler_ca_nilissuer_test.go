// Copyright (c) 2025 Justin Cranford

package handler

import (
	"bytes"
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
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

func TestBuildESTIssueRequest(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	// Generate a test CSR.
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   "test.example.com",
			Organization: []string{"Test Org"},
		},
		DNSNames: []string{"test.example.com", "www.test.example.com"},
	}
	csrDER, err := x509.CreateCertificateRequest(crand.Reader, template, key)
	require.NoError(t, err)

	csr, err := x509.ParseCertificateRequest(csrDER)
	require.NoError(t, err)

	profile := &ProfileConfig{
		ID:   "test-profile",
		Name: "Test Profile",
	}

	issueReq := handler.buildESTIssueRequest(csr, profile)
	require.NotNil(t, issueReq)
	require.NotNil(t, issueReq.SubjectRequest)
	require.Equal(t, "test.example.com", issueReq.SubjectRequest.CommonName)
	require.Equal(t, []string{"Test Org"}, issueReq.SubjectRequest.Organization)
	require.Equal(t, []string{"test.example.com", "www.test.example.com"}, issueReq.SubjectRequest.DNSNames)
}

func TestListCertificates(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	// Add test certificates to storage.
	cert1 := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		SerialNumber:   "ABC123",
		SubjectDN:      "CN=test1.example.com,O=Test Org",
		IssuerDN:       "CN=Test CA,O=Test Org",
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      "tls-server",
		CertificatePEM: "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----",
	}
	err := mockStorage.Store(context.Background(), cert1)
	require.NoError(t, err)

	cert2 := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		SerialNumber:   "DEF456",
		SubjectDN:      "CN=test2.example.com,O=Test Org",
		IssuerDN:       "CN=Test CA,O=Test Org",
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
		Status:         cryptoutilCAStorage.StatusRevoked,
		ProfileID:      "tls-server",
		CertificatePEM: "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----",
	}
	err = mockStorage.Store(context.Background(), cert2)
	require.NoError(t, err)

	handler := &Handler{storage: mockStorage}

	app.Get("/certificates", func(c *fiber.Ctx) error {
		params := cryptoutilApiCaServer.ListCertificatesParams{}

		return handler.ListCertificates(c, params)
	})

	t.Run("ListAll", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/certificates", nil)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestGetCertificate(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	testCert := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		SerialNumber:   "SERIAL123",
		SubjectDN:      "CN=test.example.com,O=Test Org",
		IssuerDN:       "CN=Test CA,O=Test Org",
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      "tls-server",
		CertificatePEM: "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
	}
	err := mockStorage.Store(context.Background(), testCert)
	require.NoError(t, err)

	handler := &Handler{storage: mockStorage}

	app.Get("/certificates/:serialNumber", func(c *fiber.Ctx) error {
		return handler.GetCertificate(c, c.Params("serialNumber"))
	})

	t.Run("Found", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/certificates/SERIAL123", nil)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/certificates/NONEXISTENT", nil)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("EmptySerial", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/certificates/", nil)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)

		require.NoError(t, resp.Body.Close())
	})
}

func TestGetCertificateChain(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	testCert := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		SerialNumber:   "CHAIN123",
		SubjectDN:      "CN=test.example.com",
		IssuerDN:       "CN=Test CA",
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      "tls-server",
		CertificatePEM: "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
	}
	err := mockStorage.Store(context.Background(), testCert)
	require.NoError(t, err)

	handler := &Handler{storage: mockStorage}

	app.Get("/certificates/:serialNumber/chain", func(c *fiber.Ctx) error {
		return handler.GetCertificateChain(c, c.Params("serialNumber"))
	})

	t.Run("Found", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/certificates/CHAIN123/chain", nil)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/certificates/NONEXISTENT/chain", nil)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestRevokeCertificate(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	testCert := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		SerialNumber:   "REVOKE123",
		SubjectDN:      "CN=test.example.com",
		IssuerDN:       "CN=Test CA",
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      "tls-server",
		CertificatePEM: "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
	}
	err := mockStorage.Store(context.Background(), testCert)
	require.NoError(t, err)

	handler := &Handler{storage: mockStorage}

	app.Post("/certificates/:serialNumber/revoke", func(c *fiber.Ctx) error {
		return handler.RevokeCertificate(c, c.Params("serialNumber"))
	})

	revokeBody := `{"reason": "key_compromise"}`

	t.Run("RevokeSuccess", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/certificates/REVOKE123/revoke", bytes.NewBufferString(revokeBody))
		req.Header.Set("Content-Type", "application/json")
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/certificates/NONEXISTENT/revoke", bytes.NewBufferString(revokeBody))
		req.Header.Set("Content-Type", "application/json")
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestGetEnrollmentStatusHandler(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	tracker := newEnrollmentTracker(cryptoutilSharedMagic.JoseJAMaxMaterials)
	requestID := googleUuid.New()
	tracker.track(requestID, cryptoutilApiCaServer.EnrollmentStatusResponseStatusIssued, "ISSUED123")

	handler := &Handler{
		storage:           mockStorage,
		enrollmentTracker: tracker,
	}

	app.Get("/enroll/:requestId", func(c *fiber.Ctx) error {
		idStr := c.Params("requestId")

		id, parseErr := googleUuid.Parse(idStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: "invalid request ID"})
		}

		return handler.GetEnrollmentStatus(c, id)
	})

	t.Run("Found", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/enroll/"+requestID.String(), nil)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		unknownID := googleUuid.New()
		req := httptest.NewRequest(http.MethodGet, "/enroll/"+unknownID.String(), nil)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		err := resp.Body.Close()
		require.NoError(t, err)
	})
}
