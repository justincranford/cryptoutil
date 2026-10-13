// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"bytes"
	"context"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
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
			name: testCaseNameInvalidJSONBody,
			setupStorage: func() cryptoutilCAStorage.Store {
				return cryptoutilCAStorage.NewMemoryStore()
			},
			serialNumber:   testCertificateSerialNumber,
			requestBody:    `{invalid json}`,
			expectedStatus: fiber.StatusBadRequest,
		},
		{
			name: "certificate_not_found",
			setupStorage: func() cryptoutilCAStorage.Store {
				return cryptoutilCAStorage.NewMemoryStore()
			},
			serialNumber:   "NONEXISTENT",
			requestBody:    testRequestBodyKeyCompromise,
			expectedStatus: fiber.StatusNotFound,
		},
		{
			name: "already_revoked",
			setupStorage: func() cryptoutilCAStorage.Store {
				storage := cryptoutilCAStorage.NewMemoryStore()
				cert := &cryptoutilCAStorage.StoredCertificate{
					ID:             googleUuid.New(),
					SerialNumber:   testSerialAlreadyRevoked,
					SubjectDN:      testSubjectDNExample,
					IssuerDN:       testIssuerDNTestCA,
					NotBefore:      time.Now().UTC().Add(-time.Hour),
					NotAfter:       time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
					Status:         cryptoutilCAStorage.StatusRevoked,
					ProfileID:      cryptoutilSharedMagic.PKIProfileIDTLSServer,
					CertificatePEM: testCertificatePEMInvalid,
				}
				_ = storage.Store(context.Background(), cert)

				return storage
			},
			serialNumber:   testSerialAlreadyRevoked,
			requestBody:    testRequestBodyKeyCompromise,
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

			req := httptest.NewRequest(http.MethodPost, "/certificates/"+tc.serialNumber+cryptoutilSharedMagic.PathRevoke, bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", testContentTypeJSON)

			resp, err := app.Test(req, -1)
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
		{testRevocationCaseKeyCompromise, cryptoutilApiCaServer.KeyCompromise},
		{testRevocationCaseCACompromise, cryptoutilApiCaServer.CACompromise},
		{testRevocationCaseAffiliationChanged, cryptoutilApiCaServer.AffiliationChanged},
		{cryptoutilSharedMagic.RFC5280RevocationReasonNameSuperseded, cryptoutilApiCaServer.Superseded},
		{testRevocationCaseCessationOfOperation, cryptoutilApiCaServer.CessationOfOperation},
		{testRevocationCaseCertificateHold, cryptoutilApiCaServer.CertificateHold},
		{testRevocationCaseRemoveFromCRL, cryptoutilApiCaServer.RemoveFromCRL},
		{testRevocationCasePrivilegeWithdrawn, cryptoutilApiCaServer.PrivilegeWithdrawn},
		{testRevocationCaseAACompromise, cryptoutilApiCaServer.AaCompromise},
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
				SubjectDN:      testSubjectDNExample,
				IssuerDN:       testIssuerDNTestCA,
				NotBefore:      time.Now().UTC().Add(-time.Hour),
				NotAfter:       time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
				Status:         cryptoutilCAStorage.StatusActive,
				ProfileID:      cryptoutilSharedMagic.PKIProfileIDTLSServer,
				CertificatePEM: testCertificatePEMInvalid,
			}
			err := storage.Store(context.Background(), cert)
			require.NoError(t, err)

			handler := &Handler{storage: storage}

			app.Post("/certificates/:serialNumber/revoke", func(c *fiber.Ctx) error {
				return handler.RevokeCertificate(c, c.Params("serialNumber"))
			})

			requestBody := `{"reason": "` + tc.name + `"}`
			req := httptest.NewRequest(http.MethodPost, "/certificates/"+cert.SerialNumber+cryptoutilSharedMagic.PathRevoke, bytes.NewBufferString(requestBody))
			req.Header.Set("Content-Type", testContentTypeJSON)

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusOK, resp.StatusCode)

			err = resp.Body.Close()
			require.NoError(t, err)
		})
	}
}
