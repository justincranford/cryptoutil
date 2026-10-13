// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"bytes"
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/sm-kms/server/pki/service/revocation"
	cryptoutilCAServiceTimestamp "cryptoutil/internal/apps/sm-kms/server/pki/service/timestamp"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
)

func TestEstSimpleEnrollWithRealIssuer(t *testing.T) {
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

	app.Post("/est/simpleenroll", func(c *fiber.Ctx) error {
		return handler.ESTSimpleEnroll(c)
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

	t.Run("SuccessfulEnrollmentWithDER", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/est/simpleenroll", bytes.NewReader(csrDER))
		req.Header.Set("Content-Type", "application/pkcs10")
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)
		require.Contains(t, string(body), "-----BEGIN CERTIFICATE-----")

		require.NoError(t, resp.Body.Close())
	})

	t.Run("SuccessfulEnrollmentWithBase64", func(t *testing.T) {
		t.Parallel()

		csrBase64 := base64.StdEncoding.EncodeToString(csrDER)
		req := httptest.NewRequest(http.MethodPost, "/est/simpleenroll", bytes.NewBufferString(csrBase64))
		req.Header.Set("Content-Type", "application/pkcs10")
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)
		require.Contains(t, string(body), "-----BEGIN CERTIFICATE-----")

		require.NoError(t, resp.Body.Close())
	})

	t.Run("SuccessfulEnrollmentWithPEM", func(t *testing.T) {
		t.Parallel()

		csrPEM := pem.EncodeToMemory(&pem.Block{
			Type:  cryptoutilSharedMagic.StringPEMTypeCSR,
			Bytes: csrDER,
		})
		req := httptest.NewRequest(http.MethodPost, "/est/simpleenroll", bytes.NewReader(csrPEM))
		req.Header.Set("Content-Type", "application/pkcs10")
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)
		require.Contains(t, string(body), "-----BEGIN CERTIFICATE-----")

		require.NoError(t, resp.Body.Close())
	})

	t.Run("InvalidCSR", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/est/simpleenroll", bytes.NewBufferString("invalid-csr-data"))
		req.Header.Set("Content-Type", "application/pkcs10")
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestEstSimpleEnrollNoProfile(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	// No profiles configured.
	handler := &Handler{
		storage:           mockStorage,
		issuer:            testSetup.Issuer,
		profiles:          map[string]*ProfileConfig{},
		enrollmentTracker: newEnrollmentTracker(cryptoutilSharedMagic.JoseJAMaxMaterials),
	}

	app.Post("/est/simpleenroll", func(c *fiber.Ctx) error {
		return handler.ESTSimpleEnroll(c)
	})

	// Generate a test CSR.
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: testDomainNameExample,
		},
	}
	csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, key)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/est/simpleenroll", bytes.NewReader(csrDER))
	req.Header.Set("Content-Type", "application/pkcs10")
	resp, testErr := app.Test(req, -1)
	require.NoError(t, testErr)
	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestTsaTimestampWithService(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	// Type assert private key to crypto.Signer.
	signer, ok := testSetup.Issuer.GetCAConfig().PrivateKey.(crypto.Signer)
	require.True(t, ok, "private key must implement crypto.Signer")

	// Create TSA service.
	tsaConfig := &cryptoutilCAServiceTimestamp.TSAConfig{
		Certificate:        testSetup.Issuer.GetCAConfig().Certificate,
		PrivateKey:         signer,
		Provider:           testSetup.Provider,
		Policy:             []int{1, cryptoutilSharedMagic.PKITestOIDArc3, cryptoutilSharedMagic.DefaultEmailOTPLength, 1, 4, 1, 99999, 1},
		AcceptedAlgorithms: []cryptoutilCAServiceTimestamp.HashAlgorithm{cryptoutilCAServiceTimestamp.HashAlgorithmSHA256},
	}
	tsaService, err := cryptoutilCAServiceTimestamp.NewTSAService(tsaConfig)
	require.NoError(t, err)

	handler := &Handler{
		storage:    cryptoutilCAStorage.NewMemoryStore(),
		issuer:     testSetup.Issuer,
		tsaService: tsaService,
	}

	app := fiber.New()
	app.Post("/tsa/timestamp", func(c *fiber.Ctx) error {
		return handler.TsaTimestamp(c)
	})

	t.Run("EmptyRequest", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/tsa/timestamp", bytes.NewReader([]byte{}))
		req.Header.Set("Content-Type", "application/timestamp-query")
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("InvalidDERRequest", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/tsa/timestamp", bytes.NewReader([]byte("not-valid-der")))
		req.Header.Set("Content-Type", "application/timestamp-query")
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

// TestGetCRLWithService tests CRL generation with a real CRL service.
func TestGetCRLWithService(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	// Type assert private key to crypto.Signer.
	signer, ok := testSetup.Issuer.GetCAConfig().PrivateKey.(crypto.Signer)
	require.True(t, ok, "private key must implement crypto.Signer")

	// Create CRL service.
	crlConfig := &cryptoutilCAServiceRevocation.CRLConfig{
		Issuer:           testSetup.Issuer.GetCAConfig().Certificate,
		PrivateKey:       signer,
		Provider:         testSetup.Provider,
		Validity:         cryptoutilSharedMagic.HoursPerDay * time.Hour,
		NextUpdateBuffer: time.Hour,
	}
	crlService, err := cryptoutilCAServiceRevocation.NewCRLService(crlConfig)
	require.NoError(t, err)

	handler := &Handler{
		storage:    cryptoutilCAStorage.NewMemoryStore(),
		issuer:     testSetup.Issuer,
		crlService: crlService,
	}

	app := fiber.New()
	app.Get("/ca/:caId/crl", func(c *fiber.Ctx) error {
		caID := c.Params("caId")
		formatParam := c.Query("format", "der")
		format := cryptoutilApiCaServer.GetCRLParamsFormat(formatParam)

		return handler.GetCRL(c, caID, cryptoutilApiCaServer.GetCRLParams{Format: &format})
	})

	t.Run("CANotFound", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ca/wrong-ca/crl", nil)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		closeErr := resp.Body.Close()
		require.NoError(t, closeErr)
	})

	t.Run("DERFormat", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ca/test-ca/crl?format=der", nil)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		require.Equal(t, "application/pkix-crl", resp.Header.Get("Content-Type"))
		require.Contains(t, resp.Header.Get("Content-Disposition"), ".crl")

		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)
		require.NotEmpty(t, body)

		closeErr := resp.Body.Close()
		require.NoError(t, closeErr)
	})

	t.Run("PEMFormat", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ca/test-ca/crl?format=pem", nil)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		require.Equal(t, "application/x-pem-file", resp.Header.Get("Content-Type"))
		require.Contains(t, resp.Header.Get("Content-Disposition"), ".crl.pem")

		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)
		require.Contains(t, string(body), "-----BEGIN X509 CRL-----")

		closeErr := resp.Body.Close()
		require.NoError(t, closeErr)
	})
}

// TestHandleOCSPWithService tests OCSP handling with a real OCSP service.

func TestESTSimpleEnroll_MissingIssuer(t *testing.T) {
	t.Parallel()

	profiles := map[string]*ProfileConfig{
		cryptoutilSharedMagic.PKIProfileIDTLSServer: {
			ID:       cryptoutilSharedMagic.PKIProfileIDTLSServer,
			Name:     testProfileNameTLSServer,
			Category: cryptoutilSharedMagic.PKIProfileCategoryTLS,
		},
	}

	tests := []struct {
		name          string
		requestBody   func(*testing.T) []byte
		wantStatus    int
		wantErrorText string
	}{
		{
			name:          testCaseNameEmptyRequest,
			requestBody:   func(*testing.T) []byte { return nil },
			wantStatus:    fiber.StatusBadRequest,
			wantErrorText: "empty CSR",
		},
		{
			name:          "valid CSR",
			requestBody:   createECDSACSR,
			wantStatus:    fiber.StatusServiceUnavailable,
			wantErrorText: testCAIssuerNotConfiguredMessage,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{profiles: profiles}
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Post("/est/simpleenroll", handler.ESTSimpleEnroll)

			request := httptest.NewRequest(http.MethodPost, "/est/simpleenroll", bytes.NewReader(tc.requestBody(t)))
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
