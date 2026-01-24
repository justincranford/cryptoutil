// Copyright (c) 2025 Justin Cranford

package handler

import (
	"bytes"
	"context"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilApiCaServer "cryptoutil/api/ca/server"
	cryptoutilCAStorage "cryptoutil/internal/ca/storage"
)

// TestRevokeCertificate_ErrorPaths tests various error conditions in RevokeCertificate.
func TestRevokeCertificate_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupStorage   func() cryptoutilCAStorage.Store
		serialNumber   string
		requestBody    string
		expectedStatus int
	}{
		{
			name: "invalid_json_body",
			setupStorage: func() cryptoutilCAStorage.Store {
				return cryptoutilCAStorage.NewMemoryStore()
			},
			serialNumber:   "SERIAL123",
			requestBody:    `{invalid json}`,
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name: "certificate_not_found",
			setupStorage: func() cryptoutilCAStorage.Store {
				return cryptoutilCAStorage.NewMemoryStore()
			},
			serialNumber:   "NONEXISTENT",
			requestBody:    `{"reason": "key_compromise"}`,
			expectedStatus: fiber.StatusNotFound,
		},
		{
			name: "already_revoked",
			setupStorage: func() cryptoutilCAStorage.Store {
				storage := cryptoutilCAStorage.NewMemoryStore()
				cert := &cryptoutilCAStorage.StoredCertificate{
					ID:             googleUuid.New(),
					SerialNumber:   "ALREADY_REVOKED",
					SubjectDN:      "CN=test.example.com",
					IssuerDN:       "CN=Test CA",
					NotBefore:      time.Now().Add(-time.Hour),
					NotAfter:       time.Now().Add(time.Hour * 24 * 365),
					Status:         cryptoutilCAStorage.StatusRevoked,
					ProfileID:      "tls-server",
					CertificatePEM: "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
				}
				_ = storage.Store(context.Background(), cert)

				return storage
			},
			serialNumber:   "ALREADY_REVOKED",
			requestBody:    `{"reason": "key_compromise"}`,
			expectedStatus: fiber.StatusConflict,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			storage := tc.setupStorage()
			handler := &Handler{storage: storage}

			app.Post("/certificates/:serialNumber/revoke", func(c *fiber.Ctx) error {
				serial := c.Params("serialNumber")

				return handler.RevokeCertificate(c, serial)
			})

			req := httptest.NewRequest(http.MethodPost, "/certificates/"+tc.serialNumber+"/revoke", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			err = resp.Body.Close()
			require.NoError(t, err)
		})
	}
}

// TestRevokeCertificate_AllReasons tests all revocation reasons are mapped correctly.
func TestRevokeCertificate_AllReasons(t *testing.T) {
	t.Parallel()

	reasons := []struct {
		name   string
		reason cryptoutilApiCaServer.RevocationReason
	}{
		{"key_compromise", cryptoutilApiCaServer.KeyCompromise},
		{"ca_compromise", cryptoutilApiCaServer.CACompromise},
		{"affiliation_changed", cryptoutilApiCaServer.AffiliationChanged},
		{"superseded", cryptoutilApiCaServer.Superseded},
		{"cessation_of_operation", cryptoutilApiCaServer.CessationOfOperation},
		{"certificate_hold", cryptoutilApiCaServer.CertificateHold},
		{"remove_from_crl", cryptoutilApiCaServer.RemoveFromCRL},
		{"privilege_withdrawn", cryptoutilApiCaServer.PrivilegeWithdrawn},
		{"aa_compromise", cryptoutilApiCaServer.AaCompromise},
	}

	for _, tc := range reasons {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			storage := cryptoutilCAStorage.NewMemoryStore()

			// Create unique certificate for each test case.
			cert := &cryptoutilCAStorage.StoredCertificate{
				ID:             googleUuid.New(),
				SerialNumber:   googleUuid.NewString(),
				SubjectDN:      "CN=test.example.com",
				IssuerDN:       "CN=Test CA",
				NotBefore:      time.Now().Add(-time.Hour),
				NotAfter:       time.Now().Add(time.Hour * 24 * 365),
				Status:         cryptoutilCAStorage.StatusActive,
				ProfileID:      "tls-server",
				CertificatePEM: "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
			}
			err := storage.Store(context.Background(), cert)
			require.NoError(t, err)

			handler := &Handler{storage: storage}

			app.Post("/certificates/:serialNumber/revoke", func(c *fiber.Ctx) error {
				return handler.RevokeCertificate(c, c.Params("serialNumber"))
			})

			requestBody := `{"reason": "` + tc.name + `"}`
			req := httptest.NewRequest(http.MethodPost, "/certificates/"+cert.SerialNumber+"/revoke", bytes.NewBufferString(requestBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusOK, resp.StatusCode)

			err = resp.Body.Close()
			require.NoError(t, err)
		})
	}
}
