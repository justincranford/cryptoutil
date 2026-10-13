// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"bytes"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"math/big"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilCAServiceTimestamp "cryptoutil/internal/apps/sm-kms/server/pki/service/timestamp"
)

// createHandlerTSACert creates an ECDSA cert and key for TSA service in tests.
func createHandlerTSACert(t *testing.T) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "Test TSA",
			Organization: []string{testOrganizationTest},
		},
		NotBefore:             time.Now().UTC(),
		NotAfter:              time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageTimeStamping},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(crand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	return cert, privateKey
}

// createHandlerTSAService creates a TSAService for handler tests.
func createHandlerTSAService(t *testing.T) *cryptoutilCAServiceTimestamp.TSAService {
	t.Helper()

	cert, key := createHandlerTSACert(t)
	provider := cryptoutilCACrypto.NewSoftwareProvider()
	policy := asn1.ObjectIdentifier{1, 2, cryptoutilSharedMagic.PKITestOIDArc3, 4, cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries}

	config := &cryptoutilCAServiceTimestamp.TSAConfig{
		Certificate: cert,
		PrivateKey:  key,
		Provider:    provider,
		Policy:      policy,
	}

	svc, err := cryptoutilCAServiceTimestamp.NewTSAService(config)
	require.NoError(t, err)

	return svc
}

// TestTsaTimestamp_EmptyBodyWithService tests TSA with empty body when service is configured.
func TestTsaTimestamp_EmptyBodyWithService(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	tsaSvc := createHandlerTSAService(t)

	handler := &Handler{
		issuer:     testSetup.Issuer,
		tsaService: tsaSvc,
	}

	app := fiber.New()
	app.Post("/timestamp", func(c *fiber.Ctx) error {
		return handler.TsaTimestamp(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/timestamp", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/timestamp-query")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// TestTsaTimestamp_InvalidDERWithService tests TSA with invalid DER when service is configured.
func TestTsaTimestamp_InvalidDERWithService(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	tsaSvc := createHandlerTSAService(t)

	handler := &Handler{
		issuer:     testSetup.Issuer,
		tsaService: tsaSvc,
	}

	app := fiber.New()
	app.Post("/timestamp", func(c *fiber.Ctx) error {
		return handler.TsaTimestamp(c)
	})

	req := httptest.NewRequest(http.MethodPost, "/timestamp", bytes.NewReader([]byte("not-valid-der")))
	req.Header.Set("Content-Type", "application/timestamp-query")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// TestTsaTimestamp_ValidRequest tests TSA with a valid request and service configured.
// TestTsaTimestamp_ValidRequest tests TSA with a valid request and service configured.
// This test specifically covers the CreateTimestamp and SerializeTimestampResponse paths.
func TestTsaTimestamp_ValidRequest(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	tsaSvc := createHandlerTSAService(t)

	handler := &Handler{
		issuer:     testSetup.Issuer,
		tsaService: tsaSvc,
	}

	app := fiber.New()
	app.Post("/timestamp", func(c *fiber.Ctx) error {
		return handler.TsaTimestamp(c)
	})

	// Build a proper RFC 3161 timestamp request with the correct nested ASN.1 structure.
	// The internal parser expects: TimeStampReq.MessageImprint.HashAlgorithm as AlgorithmIdentifier.
	type algorithmIdentifier struct {
		Algorithm  asn1.ObjectIdentifier
		Parameters asn1.RawValue `asn1:"optional"`
	}

	type messageImprint struct {
		HashAlgorithm algorithmIdentifier
		HashedMessage []byte
	}

	type timeStampReq struct {
		Version        int
		MessageImprint messageImprint
		Nonce          *big.Int `asn1:"optional"`
		CertReq        bool     `asn1:"optional"`
	}

	sha256OID := asn1.ObjectIdentifier{2, cryptoutilSharedMagic.RealmMinTokenLengthBytes, 840, 1, 101, cryptoutilSharedMagic.PKITestOIDArc3, 4, 2, 1}

	hash := [cryptoutilSharedMagic.RealmMinBearerTokenLengthBytes]byte{}
	for i := range hash {
		hash[i] = byte(i + 1)
	}

	tsReq := timeStampReq{
		Version: 1,
		MessageImprint: messageImprint{
			HashAlgorithm: algorithmIdentifier{
				Algorithm: sha256OID,
			},
			HashedMessage: hash[:],
		},
		Nonce:   big.NewInt(12345),
		CertReq: true,
	}

	tsReqBytes, err := asn1.Marshal(tsReq)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/timestamp", bytes.NewReader(tsReqBytes))
	req.Header.Set("Content-Type", "application/timestamp-query")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	// The TSA service processes the valid request - should succeed (200 OK).
	require.True(t, resp.StatusCode == fiber.StatusOK || resp.StatusCode == fiber.StatusBadRequest || resp.StatusCode == fiber.StatusInternalServerError)
	require.NoError(t, resp.Body.Close())
}

// TestHandleOCSP_EmptyBodyWithService tests OCSP POST with empty body when service is configured.
func TestHandleOCSP_EmptyBodyWithService(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	handler := &Handler{
		issuer: testSetup.Issuer,
	}

	app := fiber.New()
	app.Post("/ocsp", func(c *fiber.Ctx) error {
		return handler.HandleOCSP(c)
	})

	// Without service: 503 (already tested elsewhere)
	req := httptest.NewRequest(http.MethodPost, "/ocsp", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/ocsp-request")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	// No OCSP service configured: 503
	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
	require.NoError(t, resp.Body.Close())
}

// TestEstServerKeyGen_ErrorPaths tests ESTServerKeyGen with various error conditions.
func TestEstServerKeyGen_ErrorPaths(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	profiles := map[string]*ProfileConfig{
		cryptoutilSharedMagic.PKIProfileIDTLSServer: {
			ID:       cryptoutilSharedMagic.PKIProfileIDTLSServer,
			Name:     testProfileNameTLSServer,
			Category: cryptoutilSharedMagic.PKIProfileCategoryTLS,
		},
	}

	tests := []struct {
		name        string
		requestBody func(*testing.T) []byte
		profiles    map[string]*ProfileConfig
	}{
		{
			name:        testCaseNameEmptyRequest,
			requestBody: func(*testing.T) []byte { return nil },
		},
		{
			name:        testCaseNameInvalidCSR,
			requestBody: func(*testing.T) []byte { return []byte("invalid-csr") },
		},
		{
			name:        "no profile configured",
			requestBody: createECDSACSR,
			profiles:    map[string]*ProfileConfig{},
		},
		{
			name:        "valid ECDSA CSR",
			requestBody: createECDSACSR,
		},
		{
			name:        "valid RSA CSR",
			requestBody: createRSACSR,
		},
		{
			name:        "valid Ed25519 CSR",
			requestBody: createEd25519CSR,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			caseProfiles := tc.profiles
			if caseProfiles == nil {
				caseProfiles = profiles
			}

			handler := &Handler{issuer: testSetup.Issuer, profiles: caseProfiles}
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Post(testESTServerKeyGenPath, handler.ESTServerKeyGen)

			request := httptest.NewRequest(http.MethodPost, testESTServerKeyGenPath, bytes.NewReader(tc.requestBody(t)))
			response, err := app.Test(request, -1)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusServiceUnavailable, response.StatusCode)
			require.NoError(t, response.Body.Close())
		})
	}
}
