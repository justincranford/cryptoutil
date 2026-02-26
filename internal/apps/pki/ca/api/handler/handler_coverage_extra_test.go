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
	json "encoding/json"
	"encoding/pem"
	"math/big"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ocsp"

	cryptoutilApiCaServer "cryptoutil/api/ca/server"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/pki/ca/service/revocation"
	cryptoutilCAStorage "cryptoutil/internal/apps/pki/ca/storage"
)

// TestSubmitEnrollment_WithOverrides tests SubmitEnrollment with SubjectOverride,
// SANOverride, and ValidityDays to cover the buildIssueRequest conditional branches.
func TestSubmitEnrollment_WithOverrides(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
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

	app := fiber.New()
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

	csrPEM := string(pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCSR,
		Bytes: csrDER,
	}))

	// Build the request body with overrides.
	dnsNames := []string{"override.example.com"}
	org := []string{"Override Org"}
	validityDays := cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days

	reqBody := cryptoutilApiCaServer.EnrollmentRequest{
		CSR:     csrPEM,
		Profile: "tls-server",
		SubjectOverride: &cryptoutilApiCaServer.SubjectOverride{
			Organization: &org,
		},
		SANOverride: &cryptoutilApiCaServer.SANOverride{
			DNSNames: &dnsNames,
		},
		ValidityDays: &validityDays,
	}

	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/enrollments", bytes.NewReader(bodyBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusCreated, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// TestRevokeCertificate_EmptySerial tests RevokeCertificate with empty serial number.
func TestRevokeCertificate_EmptySerial(t *testing.T) {
	t.Parallel()

	mockStorage := cryptoutilCAStorage.NewMemoryStore()
	handler := &Handler{storage: mockStorage}

	app := fiber.New()
	// Use a fixed path, not a param-based route, to pass empty string.
	app.Post("/revoke-empty", func(c *fiber.Ctx) error {
		return handler.RevokeCertificate(c, "")
	})

	req := httptest.NewRequest(http.MethodPost, "/revoke-empty", bytes.NewBufferString(`{"reason":"key_compromise"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// TestHandleOCSP_ValidRequest tests HandleOCSP with a valid OCSP request
// to cover the RespondToRequest and Send paths.
func TestHandleOCSP_ValidRequest(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	signer, ok := testSetup.Issuer.GetCAConfig().PrivateKey.(crypto.Signer)
	require.True(t, ok)

	crlConfig := &cryptoutilCAServiceRevocation.CRLConfig{
		Issuer:     testSetup.Issuer.GetCAConfig().Certificate,
		PrivateKey: signer,
		Provider:   testSetup.Provider,
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	crlService, err := cryptoutilCAServiceRevocation.NewCRLService(crlConfig)
	require.NoError(t, err)

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

	// Create and store a test certificate for lookup.
	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	leafTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(12345),
		Subject: pkix.Name{
			CommonName: "ocsp-test.example.com",
		},
		NotBefore: time.Now().UTC().Add(-time.Hour),
		NotAfter:  time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay),
	}

	leafCertDER, err := x509.CreateCertificate(
		crand.Reader,
		leafTemplate,
		testSetup.Issuer.GetCAConfig().Certificate,
		&leafKey.PublicKey,
		signer,
	)
	require.NoError(t, err)

	leafCert, err := x509.ParseCertificate(leafCertDER)
	require.NoError(t, err)

	certPEM := string(pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: leafCertDER}))
	serialHex := leafCert.SerialNumber.Text(cryptoutilSharedMagic.RealmMinTokenLengthBytes)

	stored := &cryptoutilCAStorage.StoredCertificate{
		SerialNumber:   serialHex,
		SubjectDN:      leafCert.Subject.String(),
		IssuerDN:       leafCert.Issuer.String(),
		NotBefore:      leafCert.NotBefore,
		NotAfter:       leafCert.NotAfter,
		CertificatePEM: certPEM,
		Status:         cryptoutilCAStorage.StatusActive,
	}

	err = mockStorage.Store(context.Background(), stored)
	require.NoError(t, err)

	// Create a valid OCSP request.
	requestBytes, err := ocsp.CreateRequest(leafCert, testSetup.Issuer.GetCAConfig().Certificate, nil)
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/ocsp", func(c *fiber.Ctx) error {
		return handler.HandleOCSP(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/ocsp", bytes.NewReader(requestBytes))
	req.Header.Set("Content-Type", "application/ocsp-request")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, "application/ocsp-response", resp.Header.Get("Content-Type"))
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// TestHandleOCSP_ValidRequest_CertNotInStorage tests OCSP with cert not found in storage.
// This covers the lookupCertificateBySerial storage error/not-found path.
func TestHandleOCSP_ValidRequest_CertNotInStorage(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	signer, ok := testSetup.Issuer.GetCAConfig().PrivateKey.(crypto.Signer)
	require.True(t, ok)

	crlConfig := &cryptoutilCAServiceRevocation.CRLConfig{
		Issuer:     testSetup.Issuer.GetCAConfig().Certificate,
		PrivateKey: signer,
		Provider:   testSetup.Provider,
		Validity:   cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	crlService, err := cryptoutilCAServiceRevocation.NewCRLService(crlConfig)
	require.NoError(t, err)

	ocspConfig := &cryptoutilCAServiceRevocation.OCSPConfig{
		Issuer:       testSetup.Issuer.GetCAConfig().Certificate,
		Responder:    testSetup.Issuer.GetCAConfig().Certificate,
		ResponderKey: signer,
		Provider:     testSetup.Provider,
		Validity:     time.Hour,
	}

	ocspService, err := cryptoutilCAServiceRevocation.NewOCSPService(ocspConfig, crlService)
	require.NoError(t, err)

	// Empty storage — cert will NOT be found.
	handler := &Handler{
		storage:     mockStorage,
		issuer:      testSetup.Issuer,
		ocspService: ocspService,
	}

	// Create a leaf cert that is NOT stored.
	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	leafTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(99999),
		Subject: pkix.Name{
			CommonName: "not-stored.example.com",
		},
		NotBefore: time.Now().UTC().Add(-time.Hour),
		NotAfter:  time.Now().UTC().Add(time.Hour * cryptoutilSharedMagic.HoursPerDay),
	}

	leafCertDER, err := x509.CreateCertificate(
		crand.Reader,
		leafTemplate,
		testSetup.Issuer.GetCAConfig().Certificate,
		&leafKey.PublicKey,
		signer,
	)
	require.NoError(t, err)

	leafCert, err := x509.ParseCertificate(leafCertDER)
	require.NoError(t, err)

	// Create OCSP request for cert not in storage.
	requestBytes, err := ocsp.CreateRequest(leafCert, testSetup.Issuer.GetCAConfig().Certificate, nil)
	require.NoError(t, err)

	app := fiber.New()
	app.Post("/ocsp", func(c *fiber.Ctx) error {
		return handler.HandleOCSP(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/ocsp", bytes.NewReader(requestBytes))
	req.Header.Set("Content-Type", "application/ocsp-request")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	// OCSP response in any form (not-found etc.)
	require.Equal(t, "application/ocsp-response", resp.Header.Get("Content-Type"))
	require.NoError(t, resp.Body.Close())
}
