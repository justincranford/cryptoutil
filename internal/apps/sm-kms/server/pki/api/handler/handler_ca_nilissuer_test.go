// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"bytes"
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
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

func TestBuildESTIssueRequest(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	// Generate a test CSR.
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   testDomainNameExample,
			Organization: []string{testOrganizationTest},
		},
		DNSNames: []string{testDomainNameExample, testDomainNameWwwExample},
	}
	csrDER, err := x509.CreateCertificateRequest(crand.Reader, template, key)
	require.NoError(t, err)

	csr, err := x509.ParseCertificateRequest(csrDER)
	require.NoError(t, err)

	profile := &ProfileConfig{
		ID:   testProfileIDTest,
		Name: testProfileNameTest,
	}

	issueReq := handler.buildESTIssueRequest(csr, profile)
	require.NotNil(t, issueReq)
	require.NotNil(t, issueReq.SubjectRequest)
	require.Equal(t, testDomainNameExample, issueReq.SubjectRequest.CommonName)
	require.Equal(t, []string{testOrganizationTest}, issueReq.SubjectRequest.Organization)
	require.Equal(t, []string{testDomainNameExample, testDomainNameWwwExample}, issueReq.SubjectRequest.DNSNames)
}

func TestListCertificates(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	// Add test certificates to storage.
	cert1 := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		RequestID:      googleUuid.Must(googleUuid.NewV7()),
		SerialNumber:   testSerialABC123,
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
		RequestID:      googleUuid.Must(googleUuid.NewV7()),
		SerialNumber:   "DEF456",
		SubjectDN:      testSubjectDNTest2Org,
		IssuerDN:       testIssuerDNTestCAOrg,
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
		Status:         cryptoutilCAStorage.StatusRevoked,
		ProfileID:      cryptoutilSharedMagic.PKIProfileIDTLSServer,
		CertificatePEM: testCertificatePEMMalformed,
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
		RequestID:      googleUuid.Must(googleUuid.NewV7()),
		SerialNumber:   testCertificateSerialNumber,
		SubjectDN:      testSubjectDNExampleOrg,
		IssuerDN:       testIssuerDNTestCAOrg,
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      cryptoutilSharedMagic.PKIProfileIDTLSServer,
		CertificatePEM: testCertificatePEMInvalid,
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
		RequestID:      googleUuid.Must(googleUuid.NewV7()),
		SerialNumber:   "CHAIN123",
		SubjectDN:      testSubjectDNExample,
		IssuerDN:       testIssuerDNTestCA,
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      cryptoutilSharedMagic.PKIProfileIDTLSServer,
		CertificatePEM: testCertificatePEMInvalid,
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

	testIssuer := createTestIssuer(t).Issuer
	caConfig := testIssuer.GetCAConfig()
	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	testCert := &cryptoutilCAStorage.StoredCertificate{
		ID:                googleUuid.Must(googleUuid.NewV7()),
		RequestID:         googleUuid.Must(googleUuid.NewV7()),
		SerialNumber:      "REVOKE123",
		SubjectDN:         testSubjectDNExample,
		IssuerDN:          caConfig.Certificate.Subject.String(),
		IssuerAuthorityID: caConfig.AuthorityID,
		AuthorityKeyID:    fmt.Sprintf("%x", caConfig.Certificate.SubjectKeyId),
		NotBefore:         time.Now().UTC().Add(-time.Hour),
		NotAfter:          time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year),
		Status:            cryptoutilCAStorage.StatusActive,
		ProfileID:         cryptoutilSharedMagic.PKIProfileIDTLSServer,
		CertificatePEM:    testCertificatePEMInvalid,
	}
	err := mockStorage.Store(context.Background(), testCert)
	require.NoError(t, err)

	handler := &Handler{issuer: testIssuer, storage: mockStorage}

	app.Post("/certificates/:serialNumber/revoke", func(c *fiber.Ctx) error {
		return handler.RevokeCertificate(c, c.Params("serialNumber"))
	})

	revokeBody := testRequestBodyKeyCompromise

	t.Run("RevokeSuccess", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/certificates/REVOKE123/revoke", bytes.NewBufferString(revokeBody))
		req.Header.Set("Content-Type", testContentTypeJSON)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/certificates/NONEXISTENT/revoke", bytes.NewBufferString(revokeBody))
		req.Header.Set("Content-Type", testContentTypeJSON)
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
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{cryptoutilSharedMagic.StringError: testErrorMessageInvalidRequestID})
		}

		return handler.GetEnrollmentStatus(c, id)
	})

	t.Run("TrackerOnly", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/enroll/"+requestID.String(), nil)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

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
