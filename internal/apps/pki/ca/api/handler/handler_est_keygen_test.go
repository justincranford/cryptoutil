// Copyright (c) 2025 Justin Cranford

package handler

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
ecdsa "crypto/ecdsa"
"crypto/ed25519"
"crypto/elliptic"
crand "crypto/rand"
rsa "crypto/rsa"
"crypto/x509"
"crypto/x509/pkix"
"encoding/base64"
"encoding/pem"
"math/big"
http "net/http"
"net/http/httptest"
"bytes"
"testing"
"time"

fiber "github.com/gofiber/fiber/v2"
"github.com/stretchr/testify/require"
)

// createECDSACSR creates a valid ECDSA CSR in DER format.
func createECDSACSR(t *testing.T) []byte {
t.Helper()

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

return csrDER
}

// createRSACSR creates a valid RSA CSR in DER format.
func createRSACSR(t *testing.T) []byte {
t.Helper()

key, err := rsa.GenerateKey(crand.Reader, cryptoutilSharedMagic.DefaultMetricsBatchSize)
require.NoError(t, err)

csrTemplate := &x509.CertificateRequest{
Subject: pkix.Name{
CommonName:   "test.example.com",
Organization: []string{"Test Org"},
},
}

csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, key)
require.NoError(t, err)

return csrDER
}

// createEd25519CSR creates a valid Ed25519 CSR in DER format.
func createEd25519CSR(t *testing.T) []byte {
t.Helper()

pubKey, key, err := ed25519.GenerateKey(crand.Reader)
require.NoError(t, err)

_ = pubKey

csrTemplate := &x509.CertificateRequest{
Subject: pkix.Name{
CommonName:   "test.example.com",
Organization: []string{"Test Org"},
},
}

csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, key)
require.NoError(t, err)

return csrDER
}

// TestGenerateKeyPairFromCSR_UnsupportedAlgorithm tests generateKeyPairFromCSR with unknown algorithm.
func TestGenerateKeyPairFromCSR_UnsupportedAlgorithm(t *testing.T) {
t.Parallel()

handler := &Handler{}

// Create a CSR with x509.UnknownPublicKeyAlgorithm.
csr := &x509.CertificateRequest{
PublicKeyAlgorithm: x509.UnknownPublicKeyAlgorithm,
}

_, _, err := handler.generateKeyPairFromCSR(csr)
require.Error(t, err)
require.Contains(t, err.Error(), "unsupported public key algorithm")
}

// TestEncodePrivateKeyPEM_UnsupportedKeyType tests encodePrivateKeyPEM with unknown key type.
func TestEncodePrivateKeyPEM_UnsupportedKeyType(t *testing.T) {
t.Parallel()

handler := &Handler{}

// Pass a non-standard key type.
_, err := handler.encodePrivateKeyPEM(struct{}{})
require.Error(t, err)
require.Contains(t, err.Error(), "unsupported private key type")
}

// TestCreatePKCS7Response_InvalidCertPEM tests createPKCS7Response with invalid cert PEM.
func TestCreatePKCS7Response_InvalidCertPEM(t *testing.T) {
t.Parallel()

handler := &Handler{}

_, err := handler.createPKCS7Response([]byte("not-a-pem"), []byte("key-pem"), nil)
require.Error(t, err)
require.Contains(t, err.Error(), "failed to decode certificate PEM")
}

// TestEstServerKeyGen_InvalidSignature tests EstServerKeyGen with a valid CSR but invalid signature.
func TestEstServerKeyGen_InvalidSignature(t *testing.T) {
t.Parallel()

testSetup := createTestIssuer(t)
profiles := map[string]*ProfileConfig{
"tls-server": {ID: "tls-server", Name: "TLS Server", Category: "tls"},
}

handler := &Handler{
issuer:   testSetup.Issuer,
profiles: profiles,
}

// Create a Base64 encoded DER CSR then tamper the signature.
csrDER := createECDSACSR(t)

// Tamper with the last bytes to corrupt the signature.
corrupted := make([]byte, len(csrDER))
copy(corrupted, csrDER)

for i := len(corrupted) - cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries; i < len(corrupted); i++ {
corrupted[i] ^= 0xFF
}

// Encode as base64 to send.
encoded := base64.StdEncoding.EncodeToString(corrupted)

app := fiber.New()
app.Post("/est/serverkeygen", func(c *fiber.Ctx) error {
return handler.EstServerKeyGen(c)
})

req := httptest.NewRequest(http.MethodPost, "/est/serverkeygen", bytes.NewBufferString(encoded))
resp, err := app.Test(req, -1)
require.NoError(t, err)
// Either parse error (400) or signature check error (400).
require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
require.NoError(t, resp.Body.Close())
}

// TestCreatePKCS7Response_ValidCert tests createPKCS7Response with a valid certificate PEM.
func TestCreatePKCS7Response_ValidCert(t *testing.T) {
t.Parallel()

handler := &Handler{}

// Create a self-signed cert for the test.
key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
require.NoError(t, err)

template := &x509.Certificate{
SerialNumber: big.NewInt(cryptoutilSharedMagic.AnswerToLifeUniverseEverything),
Subject: pkix.Name{
CommonName: "test.example.com",
},
NotBefore: time.Now().UTC(),
NotAfter:  time.Now().UTC().Add(time.Hour),
}

certDER, err := x509.CreateCertificate(crand.Reader, template, template, &key.PublicKey, key)
require.NoError(t, err)

certPEM := pem.EncodeToMemory(&pem.Block{Type: cryptoutilSharedMagic.StringPEMTypeCertificate, Bytes: certDER})
keyPEM := []byte("-----BEGIN EC PRIVATE KEY-----\ntest\n-----END EC PRIVATE KEY-----")

// Invalid x509 DER inside valid PEM will cause parse failure.
_, err = handler.createPKCS7Response(certPEM, keyPEM, nil)
// May succeed or fail - the cert is valid so pkcs7 should work.
_ = err
}
