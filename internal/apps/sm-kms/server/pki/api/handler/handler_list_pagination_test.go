// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
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

// TestListCertificates_Pagination tests pagination logic with page > 1.
func TestListCertificates_Pagination(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	// Add 5 test certificates to storage.
	for i := 1; i <= cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries; i++ {
		cert := &cryptoutilCAStorage.StoredCertificate{
			ID:             googleUuid.New(),
			SerialNumber:   googleUuid.NewString(),
			SubjectDN:      testSubjectDNExampleOrg,
			IssuerDN:       testIssuerDNTestCAOrg,
			NotBefore:      time.Now().UTC().Add(-time.Hour),
			NotAfter:       time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
			Status:         cryptoutilCAStorage.StatusActive,
			ProfileID:      cryptoutilSharedMagic.PKIProfileIDTLSServer,
			CertificatePEM: testCertificatePEMMalformed,
		}
		err := mockStorage.Store(context.Background(), cert)
		require.NoError(t, err)
	}

	handler := &Handler{storage: mockStorage}

	app.Get("/certificates", func(c *fiber.Ctx) error {
		params := cryptoutilApiCaServer.ListCertificatesParams{}

		// Parse page parameter.
		if pageStr := c.Query(cryptoutilSharedMagic.DisplayPage); pageStr != "" {
			page := 2
			params.Page = &page
		}

		// Parse page_size parameter.
		if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
			pageSize := 2
			params.PageSize = &pageSize
		}

		return handler.ListCertificates(c, params)
	})

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "page_2_with_page_size_2",
			queryParams:    "?page=2&page_size=2",
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "page_2_default_page_size",
			queryParams:    "?page=2",
			expectedStatus: fiber.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/certificates"+tc.queryParams, nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			err = resp.Body.Close()
			require.NoError(t, err)
		})
	}
}

// TestListCertificates_Filtering tests filtering by profile and status.
func TestListCertificates_Filtering(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	// Add test certificates with different profiles and statuses.
	cert1 := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		SerialNumber:   googleUuid.NewString(),
		SubjectDN:      testSubjectDNTest1Org,
		IssuerDN:       testIssuerDNTestCAOrg,
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      cryptoutilSharedMagic.PKIProfileIDTLSServer,
		CertificatePEM: testCertificatePEMMalformed,
	}
	err := mockStorage.Store(context.Background(), cert1)
	require.NoError(t, err)

	cert2 := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		SerialNumber:   googleUuid.NewString(),
		SubjectDN:      testSubjectDNTest2Org,
		IssuerDN:       testIssuerDNTestCAOrg,
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
		Status:         cryptoutilCAStorage.StatusRevoked,
		ProfileID:      cryptoutilSharedMagic.PKIProfileIDTLSClient,
		CertificatePEM: testCertificatePEMMalformed,
	}
	err = mockStorage.Store(context.Background(), cert2)
	require.NoError(t, err)

	handler := &Handler{storage: mockStorage}

	app.Get("/certificates", func(c *fiber.Ctx) error {
		params := cryptoutilApiCaServer.ListCertificatesParams{}

		// Parse profile parameter.
		if profileStr := c.Query(cryptoutilSharedMagic.ClaimProfile); profileStr != "" {
			params.Profile = &profileStr
		}

		// Parse status parameter.
		if statusStr := c.Query(cryptoutilSharedMagic.StringStatus); statusStr != "" {
			status := cryptoutilApiCaServer.CertificateStatus(statusStr)
			params.Status = &status
		}

		return handler.ListCertificates(c, params)
	})

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "filter_by_profile_tls_server",
			queryParams:    "?profile=tls-server",
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "filter_by_status_active",
			queryParams:    "?status=active",
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "filter_by_status_revoked",
			queryParams:    "?status=revoked",
			expectedStatus: fiber.StatusOK,
		},
		{
			name:           "filter_by_profile_and_status",
			queryParams:    "?profile=tls-client&status=revoked",
			expectedStatus: fiber.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, "/certificates"+tc.queryParams, nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			err = resp.Body.Close()
			require.NoError(t, err)
		})
	}
}
