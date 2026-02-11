// Copyright (c) 2025 Justin Cranford

package handler

import (
	"context"
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

// TestGetEnrollmentStatus_WithCertificate tests enrollment status when certificate is issued and found.
func TestGetEnrollmentStatus_WithCertificate(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	storage := cryptoutilCAStorage.NewMemoryStore()

	// Create tracker and track an issued enrollment.
	tracker := newEnrollmentTracker(100)
	requestID := googleUuid.New()
	serialNumber := googleUuid.NewString()
	tracker.track(requestID, cryptoutilApiCaServer.EnrollmentStatusResponseStatusIssued, serialNumber)

	// Store the issued certificate in storage.
	cert := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		SerialNumber:   serialNumber,
		SubjectDN:      "CN=test.example.com",
		IssuerDN:       "CN=Test CA",
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * 24 * 365),
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      "tls-server",
		CertificatePEM: "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
	}
	err := storage.Store(context.Background(), cert)
	require.NoError(t, err)

	handler := &Handler{
		storage:           storage,
		enrollmentTracker: tracker,
	}

	app.Get("/enroll/:requestId", func(c *fiber.Ctx) error {
		idStr := c.Params("requestId")

		id, parseErr := googleUuid.Parse(idStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
		}

		return handler.GetEnrollmentStatus(c, id)
	})

	req := httptest.NewRequest(http.MethodGet, "/enroll/"+requestID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// TestGetEnrollmentStatus_IssuedNoCertificate tests enrollment status when issued but certificate not found in storage.
func TestGetEnrollmentStatus_IssuedNoCertificate(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	storage := cryptoutilCAStorage.NewMemoryStore()

	// Create tracker and track an issued enrollment with non-existent serial.
	tracker := newEnrollmentTracker(100)
	requestID := googleUuid.New()
	tracker.track(requestID, cryptoutilApiCaServer.EnrollmentStatusResponseStatusIssued, "NONEXISTENT")

	handler := &Handler{
		storage:           storage,
		enrollmentTracker: tracker,
	}

	app.Get("/enroll/:requestId", func(c *fiber.Ctx) error {
		idStr := c.Params("requestId")

		id, parseErr := googleUuid.Parse(idStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
		}

		return handler.GetEnrollmentStatus(c, id)
	})

	req := httptest.NewRequest(http.MethodGet, "/enroll/"+requestID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	// Should still return 200 OK, just without certificate details.
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// TestGetEnrollmentStatus_Pending tests enrollment status for pending request.
func TestGetEnrollmentStatus_Pending(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	storage := cryptoutilCAStorage.NewMemoryStore()

	// Create tracker and track a pending enrollment.
	tracker := newEnrollmentTracker(100)
	requestID := googleUuid.New()
	tracker.track(requestID, cryptoutilApiCaServer.EnrollmentStatusResponseStatusPending, "")

	handler := &Handler{
		storage:           storage,
		enrollmentTracker: tracker,
	}

	app.Get("/enroll/:requestId", func(c *fiber.Ctx) error {
		idStr := c.Params("requestId")

		id, parseErr := googleUuid.Parse(idStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
		}

		return handler.GetEnrollmentStatus(c, id)
	})

	req := httptest.NewRequest(http.MethodGet, "/enroll/"+requestID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// TestGetCertificateChain_Success tests successful certificate chain retrieval.
func TestGetCertificateChain_Success(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	storage := cryptoutilCAStorage.NewMemoryStore()

	// Store test certificate.
	serialNumber := googleUuid.NewString()
	cert := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		SerialNumber:   serialNumber,
		SubjectDN:      "CN=test.example.com,O=Test Org,C=US",
		IssuerDN:       "CN=Test CA,O=Test Org,C=US",
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * 24 * 365),
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      "tls-server",
		CertificatePEM: "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
	}
	err := storage.Store(context.Background(), cert)
	require.NoError(t, err)

	handler := &Handler{storage: storage}

	app.Get("/certificates/:serialNumber/chain", func(c *fiber.Ctx) error {
		return handler.GetCertificateChain(c, c.Params("serialNumber"))
	})

	req := httptest.NewRequest(http.MethodGet, "/certificates/"+serialNumber+"/chain", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// TestGetCertificateChain_NotFound tests certificate chain when certificate not found.
func TestGetCertificateChain_NotFound(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	storage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{storage: storage}

	app.Get("/certificates/:serialNumber/chain", func(c *fiber.Ctx) error {
		return handler.GetCertificateChain(c, c.Params("serialNumber"))
	})

	req := httptest.NewRequest(http.MethodGet, "/certificates/NONEXISTENT/chain", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}
