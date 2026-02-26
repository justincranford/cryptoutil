// Copyright (c) 2025 Justin Cranford

package handler

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"bytes"
	"context"
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/pki/ca/service/issuer"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/pki/ca/service/revocation"
	cryptoutilCAStorage "cryptoutil/internal/apps/pki/ca/storage"
)
func TestHandleOCSPWithService(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

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

	// Use the issuing CA cert as OCSP responder (self-signed responder).
	ocspConfig := &cryptoutilCAServiceRevocation.OCSPConfig{
		Issuer:       testSetup.Issuer.GetCAConfig().Certificate,
		Responder:    testSetup.Issuer.GetCAConfig().Certificate,
		ResponderKey: signer,
		Provider:     testSetup.Provider,
		Validity:     time.Hour,
	}
	ocspService, err := cryptoutilCAServiceRevocation.NewOCSPService(ocspConfig, crlService)
	require.NoError(t, err)

	handler := &Handler{
		storage:     mockStorage,
		issuer:      testSetup.Issuer,
		ocspService: ocspService,
	}

	app := fiber.New()
	app.Post("/ocsp", func(c *fiber.Ctx) error {
		return handler.HandleOCSP(c)
	})

	t.Run("EmptyRequest", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/ocsp", bytes.NewReader([]byte{}))
		req.Header.Set("Content-Type", "application/ocsp-request")
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		require.Equal(t, "application/ocsp-response", resp.Header.Get("Content-Type"))

		require.NoError(t, resp.Body.Close())
	})

	t.Run("InvalidOCSPRequest", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/ocsp", bytes.NewReader([]byte("not-valid-ocsp")))
		req.Header.Set("Content-Type", "application/ocsp-request")
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		// Invalid request returns error status with OCSP content type.
		require.Equal(t, "application/ocsp-response", resp.Header.Get("Content-Type"))

		require.NoError(t, resp.Body.Close())
	})
}

// TestNewHandlerValidation tests NewHandler validation paths.
func TestNewHandlerValidation(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	tests := []struct {
		name      string
		issuer    *cryptoutilCAServiceIssuer.Issuer
		storage   cryptoutilCAStorage.Store
		profiles  map[string]*ProfileConfig
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "NilStorage",
			issuer:    testSetup.Issuer,
			storage:   nil,
			profiles:  map[string]*ProfileConfig{},
			wantErr:   true,
			errSubstr: "storage",
		},
		{
			name:      "NilIssuer",
			issuer:    nil,
			storage:   mockStorage,
			profiles:  map[string]*ProfileConfig{},
			wantErr:   true,
			errSubstr: "issuer",
		},
		{
			name:     "ValidWithNoProfiles",
			issuer:   testSetup.Issuer,
			storage:  mockStorage,
			profiles: map[string]*ProfileConfig{},
			wantErr:  false,
		},
		{
			name:    "ValidWithProfiles",
			issuer:  testSetup.Issuer,
			storage: mockStorage,
			profiles: map[string]*ProfileConfig{
				"test-profile": {
					ID:          "test-profile",
					Name:        "Test Profile",
					Description: "A test profile",
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler, err := NewHandler(tc.issuer, tc.storage, tc.profiles)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errSubstr)
				require.Nil(t, handler)
			} else {
				require.NoError(t, err)
				require.NotNil(t, handler)
			}
		})
	}
}

// TestLookupCertificateBySerialWithCert tests serial lookup with stored certificate.
func TestLookupCertificateBySerialWithCert(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  testSetup.Issuer,
	}

	// Generate a test certificate.
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	serialNumber := big.NewInt(12345)

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "test.example.com",
		},
		NotBefore: time.Now().UTC(),
		NotAfter:  time.Now().UTC().Add(cryptoutilSharedMagic.HoursPerDay * time.Hour),
	}

	// Self-sign for testing.
	certDER, err := x509.CreateCertificate(crand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
		Bytes: certDER,
	})

	// Store the certificate.
	storedCert := &cryptoutilCAStorage.StoredCertificate{
		SerialNumber:   serialNumber.Text(cryptoutilSharedMagic.RealmMinTokenLengthBytes),
		CertificatePEM: string(certPEM),
	}
	ctx := context.Background()
	err = mockStorage.Store(ctx, storedCert)
	require.NoError(t, err)

	// Look up by serial.
	foundCert := handler.lookupCertificateBySerial(ctx, serialNumber)
	require.NotNil(t, foundCert)
	require.Equal(t, serialNumber.Int64(), foundCert.SerialNumber.Int64())
}

// TestLookupCertificateBySerialInvalidPEM tests serial lookup with invalid PEM.
func TestLookupCertificateBySerialInvalidPEM(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  testSetup.Issuer,
	}

	serialNumber := big.NewInt(99999)

	// Store certificate with invalid PEM.
	storedCert := &cryptoutilCAStorage.StoredCertificate{
		SerialNumber:   serialNumber.Text(cryptoutilSharedMagic.RealmMinTokenLengthBytes),
		CertificatePEM: "not-valid-pem-data",
	}
	ctx := context.Background()
	err := mockStorage.Store(ctx, storedCert)
	require.NoError(t, err)

	// Should return nil for invalid PEM.
	foundCert := handler.lookupCertificateBySerial(ctx, serialNumber)
	require.Nil(t, foundCert)
}

// TestEstCSRAttrsHandler tests the EST CSR attributes endpoint.
func TestEstCSRAttrsHandler(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	handler := &Handler{
		storage: cryptoutilCAStorage.NewMemoryStore(),
		issuer:  testSetup.Issuer,
		profiles: map[string]*ProfileConfig{
			"server": {
				ID:          "server",
				Name:        "Server Profile",
				Description: "Server certificate profile",
			},
		},
	}

	app := fiber.New()
	app.Get("/est/csrattrs", func(c *fiber.Ctx) error {
		return handler.EstCSRAttrs(c)
	})

	t.Run("WithProfile", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/est/csrattrs?profile=server", nil)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		// CSR attrs returns 200 or 204 based on implementation.
		require.True(t, resp.StatusCode == fiber.StatusOK || resp.StatusCode == fiber.StatusNoContent)

		err := resp.Body.Close()
		require.NoError(t, err)
	})

	t.Run("WithoutProfile", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/est/csrattrs", nil)
		resp, testErr := app.Test(req, -1)
		require.NoError(t, testErr)
		// CSR attrs returns 200 or 204 based on implementation.
		require.True(t, resp.StatusCode == fiber.StatusOK || resp.StatusCode == fiber.StatusNoContent)

		err := resp.Body.Close()
		require.NoError(t, err)
	})
}
