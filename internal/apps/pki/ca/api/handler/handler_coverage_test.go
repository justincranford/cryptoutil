// Copyright (c) 2025 Justin Cranford

package handler

import (
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"github.com/stretchr/testify/require"
)

// TestGenerateKeyPairFromCSR_AllAlgorithms tests key pair generation for all supported algorithms.
func TestGenerateKeyPairFromCSR_AllAlgorithms(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	t.Run("RSA_2048", func(t *testing.T) {
		t.Parallel()

		// Create RSA CSR template.
		rsaKey, err := rsa.GenerateKey(crand.Reader, 2048)
		require.NoError(t, err)

		csrTemplate := &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName: "test-rsa.example.com",
			},
			PublicKeyAlgorithm: x509.RSA,
		}

		csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, rsaKey)
		require.NoError(t, err)

		csr, parseErr := x509.ParseCertificateRequest(csrDER)
		require.NoError(t, parseErr)

		// Generate key pair from CSR.
		privateKey, publicKey, genErr := handler.generateKeyPairFromCSR(csr)
		require.NoError(t, genErr)
		require.NotNil(t, privateKey)
		require.NotNil(t, publicKey)

		// Verify key types.
		_, isRSAPrivate := privateKey.(*rsa.PrivateKey)
		require.True(t, isRSAPrivate, "private key should be *rsa.PrivateKey")

		_, isRSAPublic := publicKey.(*rsa.PublicKey)
		require.True(t, isRSAPublic, "public key should be *rsa.PublicKey")
	})

	t.Run("ECDSA_P256", func(t *testing.T) {
		t.Parallel()

		// Create ECDSA CSR template.
		ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
		require.NoError(t, err)

		csrTemplate := &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName: "test-ecdsa.example.com",
			},
			PublicKeyAlgorithm: x509.ECDSA,
		}

		csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, ecdsaKey)
		require.NoError(t, err)

		csr, parseErr := x509.ParseCertificateRequest(csrDER)
		require.NoError(t, parseErr)

		// Generate key pair from CSR.
		privateKey, publicKey, genErr := handler.generateKeyPairFromCSR(csr)
		require.NoError(t, genErr)
		require.NotNil(t, privateKey)
		require.NotNil(t, publicKey)

		// Verify key types.
		_, isECDSAPrivate := privateKey.(*ecdsa.PrivateKey)
		require.True(t, isECDSAPrivate, "private key should be *ecdsa.PrivateKey")

		_, isECDSAPublic := publicKey.(*ecdsa.PublicKey)
		require.True(t, isECDSAPublic, "public key should be *ecdsa.PublicKey")
	})

	t.Run("Ed25519", func(t *testing.T) {
		t.Parallel()

		// Create Ed25519 CSR template.
		ed25519Public, ed25519Private, err := ed25519.GenerateKey(crand.Reader)
		require.NoError(t, err)

		csrTemplate := &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName: "test-ed25519.example.com",
			},
			PublicKeyAlgorithm: x509.Ed25519,
		}

		csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, ed25519Private)
		require.NoError(t, err)

		csr, parseErr := x509.ParseCertificateRequest(csrDER)
		require.NoError(t, parseErr)

		// Generate key pair from CSR.
		privateKey, publicKey, genErr := handler.generateKeyPairFromCSR(csr)
		require.NoError(t, genErr)
		require.NotNil(t, privateKey)
		require.NotNil(t, publicKey)

		// Verify key types.
		_, isEd25519Private := privateKey.(ed25519.PrivateKey)
		require.True(t, isEd25519Private, "private key should be ed25519.PrivateKey")

		_, isEd25519Public := publicKey.(ed25519.PublicKey)
		require.True(t, isEd25519Public, "public key should be ed25519.PublicKey")

		// Verify keys are different from original (new keys generated).
		require.NotEqual(t, ed25519Private, privateKey)
		require.NotEqual(t, ed25519Public, publicKey)
	})

	t.Run("UnsupportedAlgorithm_DSA", func(t *testing.T) {
		t.Parallel()

		// Create CSR with unsupported algorithm (DSA).
		csr := &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName: "test-unsupported.example.com",
			},
			PublicKeyAlgorithm: x509.DSA, // Unsupported.
		}

		// Attempt to generate key pair.
		privateKey, publicKey, genErr := handler.generateKeyPairFromCSR(csr)
		require.Error(t, genErr)
		require.Nil(t, privateKey)
		require.Nil(t, publicKey)
		require.Contains(t, genErr.Error(), "unsupported public key algorithm")
	})
}

// TestEncodePrivateKeyPEM_AllKeyTypes tests PEM encoding for all supported private key types.
func TestEncodePrivateKeyPEM_AllKeyTypes(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	t.Run("RSA_PrivateKey", func(t *testing.T) {
		t.Parallel()

		// Generate RSA private key.
		rsaKey, err := rsa.GenerateKey(crand.Reader, 2048)
		require.NoError(t, err)

		// Encode to PEM.
		pemBytes, encErr := handler.encodePrivateKeyPEM(rsaKey)
		require.NoError(t, encErr)
		require.NotEmpty(t, pemBytes)

		// Verify PEM format (PKCS#1 for RSA).
		require.Contains(t, string(pemBytes), "-----BEGIN "+cryptoutilSharedMagic.StringPEMTypeRSAPrivateKey+"-----") // pragma: allowlist secret
		require.Contains(t, string(pemBytes), "-----END "+cryptoutilSharedMagic.StringPEMTypeRSAPrivateKey+"-----")   // pragma: allowlist secret
	})

	t.Run("ECDSA_PrivateKey", func(t *testing.T) {
		t.Parallel()

		// Generate ECDSA private key.
		ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
		require.NoError(t, err)

		// Encode to PEM.
		pemBytes, encErr := handler.encodePrivateKeyPEM(ecdsaKey)
		require.NoError(t, encErr)
		require.NotEmpty(t, pemBytes)

		// Verify PEM format (EC format for ECDSA).
		require.Contains(t, string(pemBytes), "-----BEGIN "+cryptoutilSharedMagic.StringPEMTypeECPrivateKey+"-----") // pragma: allowlist secret
		require.Contains(t, string(pemBytes), "-----END "+cryptoutilSharedMagic.StringPEMTypeECPrivateKey+"-----")   // pragma: allowlist secret
	})

	t.Run("Ed25519_PrivateKey", func(t *testing.T) {
		t.Parallel()

		// Generate Ed25519 private key.
		_, ed25519Private, err := ed25519.GenerateKey(crand.Reader)
		require.NoError(t, err)

		// Encode to PEM.
		pemBytes, encErr := handler.encodePrivateKeyPEM(ed25519Private)
		require.NoError(t, encErr)
		require.NotEmpty(t, pemBytes)

		// Verify PEM format.
		require.Contains(t, string(pemBytes), "-----BEGIN "+cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey+"-----") // pragma: allowlist secret
		require.Contains(t, string(pemBytes), "-----END "+cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey+"-----")   // pragma: allowlist secret
	})

	t.Run("UnsupportedKeyType_String", func(t *testing.T) {
		t.Parallel()

		// Attempt to encode unsupported key type.
		pemBytes, encErr := handler.encodePrivateKeyPEM("not-a-key")
		require.Error(t, encErr)
		require.Nil(t, pemBytes)
		require.Contains(t, encErr.Error(), "unsupported private key type")
	})
}

// TestCreateCSRWithKey_AllKeyTypes tests CSR creation with generated keys.
func TestCreateCSRWithKey_AllKeyTypes(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	t.Run("RSA_Key", func(t *testing.T) {
		t.Parallel()

		// Generate RSA key.
		rsaKey, err := rsa.GenerateKey(crand.Reader, 2048)
		require.NoError(t, err)

		// Create CSR template.
		csrTemplate := &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName:   "rsa-keygen.example.com",
				Organization: []string{"Test Org"},
			},
			DNSNames: []string{"rsa-keygen.example.com"},
		}

		// Create CSR with generated key.
		newCSR, csrErr := handler.createCSRWithKey(csrTemplate, rsaKey)
		require.NoError(t, csrErr)
		require.NotNil(t, newCSR)

		// Verify CSR attributes.
		require.Equal(t, "rsa-keygen.example.com", newCSR.Subject.CommonName)
		require.Equal(t, []string{"rsa-keygen.example.com"}, newCSR.DNSNames)
		require.Equal(t, x509.RSA, newCSR.PublicKeyAlgorithm)
	})

	t.Run("ECDSA_Key", func(t *testing.T) {
		t.Parallel()

		// Generate ECDSA key.
		ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
		require.NoError(t, err)

		// Create CSR template.
		csrTemplate := &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName:   "ecdsa-keygen.example.com",
				Organization: []string{"Test Org"},
			},
			DNSNames: []string{"ecdsa-keygen.example.com"},
		}

		// Create CSR with generated key.
		newCSR, csrErr := handler.createCSRWithKey(csrTemplate, ecdsaKey)
		require.NoError(t, csrErr)
		require.NotNil(t, newCSR)

		// Verify CSR attributes.
		require.Equal(t, "ecdsa-keygen.example.com", newCSR.Subject.CommonName)
		require.Equal(t, []string{"ecdsa-keygen.example.com"}, newCSR.DNSNames)
		require.Equal(t, x509.ECDSA, newCSR.PublicKeyAlgorithm)
	})

	t.Run("Ed25519_Key", func(t *testing.T) {
		t.Parallel()

		// Generate Ed25519 key.
		_, ed25519Private, err := ed25519.GenerateKey(crand.Reader)
		require.NoError(t, err)

		// Create CSR template.
		csrTemplate := &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName:   "ed25519-keygen.example.com",
				Organization: []string{"Test Org"},
			},
			DNSNames: []string{"ed25519-keygen.example.com"},
		}

		// Create CSR with generated key.
		newCSR, csrErr := handler.createCSRWithKey(csrTemplate, ed25519Private)
		require.NoError(t, csrErr)
		require.NotNil(t, newCSR)

		// Verify CSR attributes.
		require.Equal(t, "ed25519-keygen.example.com", newCSR.Subject.CommonName)
		require.Equal(t, []string{"ed25519-keygen.example.com"}, newCSR.DNSNames)
		require.Equal(t, x509.Ed25519, newCSR.PublicKeyAlgorithm)
	})

	t.Run("InvalidKey_NilPrivateKey", func(t *testing.T) {
		t.Parallel()

		csrTemplate := &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName: "test.example.com",
			},
		}

		// Attempt to create CSR with nil private key.
		newCSR, csrErr := handler.createCSRWithKey(csrTemplate, nil)
		require.Error(t, csrErr)
		require.Nil(t, newCSR)
	})
}
