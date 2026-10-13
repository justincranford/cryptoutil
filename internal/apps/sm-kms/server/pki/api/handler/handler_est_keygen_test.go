// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"bytes"
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

const testESTServerKeyGenPath = "/est/serverkeygen"

// createECDSACSR creates a valid ECDSA CSR in DER format.
func createECDSACSR(t *testing.T) []byte {
	t.Helper()

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

	return csrDER
}

// createRSACSR creates a valid RSA CSR in DER format.
func createRSACSR(t *testing.T) []byte {
	t.Helper()

	key, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
	require.NoError(t, err)

	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: testDomainNameExample,
		},
	}

	csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, key)
	require.NoError(t, err)

	return csrDER
}

// createEd25519CSR creates a valid Ed25519 CSR in DER format.
func createEd25519CSR(t *testing.T) []byte {
	t.Helper()

	_, key, err := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, err)

	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   testDomainNameExample,
			Organization: []string{testOrganizationTest},
		},
	}

	csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, key)
	require.NoError(t, err)

	return csrDER
}

func TestGenerateKeyPairFromCSR_UnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	handler := &Handler{}
	csr := &x509.CertificateRequest{PublicKeyAlgorithm: x509.UnknownPublicKeyAlgorithm}

	_, _, err := handler.generateKeyPairFromCSR(csr)
	require.ErrorContains(t, err, "unsupported public key algorithm")
}

func TestEncodePrivateKeyPEM_UnsupportedKeyType(t *testing.T) {
	t.Parallel()

	handler := &Handler{}
	_, err := handler.encodePrivateKeyPEM(struct{}{})
	require.ErrorContains(t, err, "unsupported private key type")
}

func TestEstServerKeyGen_EncryptionUnavailable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		requestBody func(*testing.T) []byte
	}{
		{
			name:        testCaseNameEmptyRequest,
			requestBody: func(*testing.T) []byte { return nil },
		},
		{
			name:        testCaseNameInvalidCSR,
			requestBody: func(*testing.T) []byte { return []byte("not-a-csr") },
		},
		{
			name:        "DER CSR",
			requestBody: createECDSACSR,
		},
		{
			name: "base64 CSR",
			requestBody: func(t *testing.T) []byte {
				return []byte(base64.StdEncoding.EncodeToString(createECDSACSR(t)))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler := &Handler{}
			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Post(testESTServerKeyGenPath, handler.ESTServerKeyGen)

			request := httptest.NewRequest(http.MethodPost, testESTServerKeyGenPath, bytes.NewReader(tc.requestBody(t)))
			response, err := app.Test(request, -1)
			require.NoError(t, err)
			require.Equal(t, fiber.StatusServiceUnavailable, response.StatusCode)

			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			require.NoError(t, response.Body.Close())
			require.Contains(t, string(body), "secure server-side key generation is unavailable")
			require.NotContains(t, string(body), cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey)
		})
	}
}
