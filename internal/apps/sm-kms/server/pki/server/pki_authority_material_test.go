// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"crypto"
	"crypto/ed25519"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"io"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
)

type pkiAuthorityInvalidPublicSigner struct{}

func (pkiAuthorityInvalidPublicSigner) Public() crypto.PublicKey {
	return struct{}{}
}

func (pkiAuthorityInvalidPublicSigner) Sign(io.Reader, []byte, crypto.SignerOpts) ([]byte, error) {
	return nil, nil
}

func TestValidateAuthorityMaterial(t *testing.T) {
	t.Parallel()

	config := DefaultPKIAuthorityConfig()
	certificate, privateKey, err := createSelfSignedCAWithConfig(cryptoutilCACrypto.NewSoftwareProvider(), crand.Reader, config)
	require.NoError(t, err)

	tests := []struct {
		name      string
		nilCert   bool
		configure func(*x509.Certificate, *PKIAuthorityConfig) crypto.PrivateKey
		wantErr   string
	}{
		{name: "nil certificate", nilCert: true, configure: func(*x509.Certificate, *PKIAuthorityConfig) crypto.PrivateKey { return privateKey }, wantErr: "certificate is required"},
		{
			name: "invalid configuration",
			configure: func(_ *x509.Certificate, config *PKIAuthorityConfig) crypto.PrivateKey {
				config.ID = ""

				return privateKey
			},
			wantErr: pkiTestInvalidAuthorityConfiguration,
		},
		{
			name: pkiTestCertificateIsNotCA,
			configure: func(cert *x509.Certificate, _ *PKIAuthorityConfig) crypto.PrivateKey {
				cert.IsCA = false

				return privateKey
			},
			wantErr: pkiTestMustBeCAWithValidBasicConstraints,
		},
		{
			name: "basic constraints are invalid",
			configure: func(cert *x509.Certificate, _ *PKIAuthorityConfig) crypto.PrivateKey {
				cert.BasicConstraintsValid = false

				return privateKey
			},
			wantErr: "must be a CA with valid basic constraints",
		},
		{
			name: "certificate key usage is insufficient",
			configure: func(cert *x509.Certificate, _ *PKIAuthorityConfig) crypto.PrivateKey {
				cert.KeyUsage = x509.KeyUsageDigitalSignature

				return privateKey
			},
			wantErr: "permit certificate and CRL signing",
		},
		{
			name: "subject key identifier is missing",
			configure: func(cert *x509.Certificate, _ *PKIAuthorityConfig) crypto.PrivateKey {
				cert.SubjectKeyId = nil

				return privateKey
			},
			wantErr: "subject key identifier is required",
		},
		{
			name: "self-signature is invalid",
			configure: func(cert *x509.Certificate, _ *PKIAuthorityConfig) crypto.PrivateKey {
				cert.Signature = []byte{0}

				return privateKey
			},
			wantErr: "not self-signed",
		},
		{
			name: pkiTestCertificateNotYetValid,
			configure: func(cert *x509.Certificate, _ *PKIAuthorityConfig) crypto.PrivateKey {
				cert.NotBefore = time.Now().UTC().Add(time.Hour)

				return privateKey
			},
			wantErr: pkiTestNotCurrentlyValid,
		},
		{
			name: "certificate is expired",
			configure: func(cert *x509.Certificate, _ *PKIAuthorityConfig) crypto.PrivateKey {
				cert.NotAfter = time.Now().UTC().Add(-time.Hour)

				return privateKey
			},
			wantErr: pkiTestNotCurrentlyValid,
		},
		{
			name: "configured subject differs",
			configure: func(_ *x509.Certificate, config *PKIAuthorityConfig) crypto.PrivateKey {
				config.Subject.CommonName = "different authority"

				return privateKey
			},
			wantErr: "subject does not match configured identity",
		},
		{
			name: "validity differs",
			configure: func(_ *x509.Certificate, config *PKIAuthorityConfig) crypto.PrivateKey {
				config.Validity = time.Hour

				return privateKey
			},
			wantErr: "validity does not match configured lifetime",
		},
		{
			name: "path length differs",
			configure: func(_ *x509.Certificate, config *PKIAuthorityConfig) crypto.PrivateKey {
				config.MaxPathLength++

				return privateKey
			},
			wantErr: "path length does not match configured identity",
		},
		{
			name: "public key algorithm differs",
			configure: func(_ *x509.Certificate, config *PKIAuthorityConfig) crypto.PrivateKey {
				config.KeySpec.ECDSACurve = pkiAuthorityCurveP256

				return privateKey
			},
			wantErr: pkiTestDoesNotMatchECDSACurve,
		},
		{
			name: pkiTestPrivateKeyIsNotSigner,
			configure: func(*x509.Certificate, *PKIAuthorityConfig) crypto.PrivateKey {
				return struct{}{}
			},
			wantErr: pkiTestPrivateKeyIsNotSigner,
		},
		{
			name: pkiTestCertificateAndPrivateKeyMismatch,
			configure: func(*x509.Certificate, *PKIAuthorityConfig) crypto.PrivateKey {
				_, differentKey, createErr := createSelfSignedCAWithConfig(cryptoutilCACrypto.NewSoftwareProvider(), crand.Reader, DefaultPKIAuthorityConfig())
				require.NoError(t, createErr)

				return differentKey
			},
			wantErr: pkiTestCertificateAndPrivateKeyMismatch,
		},
		{
			name: "private key public component cannot be encoded",
			configure: func(*x509.Certificate, *PKIAuthorityConfig) crypto.PrivateKey {
				return pkiAuthorityInvalidPublicSigner{}
			},
			wantErr: "encode authority private key public component",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			certCopy := *certificate
			configCopy := config
			testKey := tc.configure(&certCopy, &configCopy)

			testCert := &certCopy
			if tc.nilCert {
				testCert = nil
			}

			require.ErrorContains(t, validateAuthorityMaterial(testCert, testKey, configCopy), tc.wantErr)
		})
	}

	t.Run("valid material", func(t *testing.T) {
		t.Parallel()
		require.NoError(t, validateAuthorityMaterial(certificate, privateKey, config))
	})
}

func TestValidateAuthorityPublicKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		public  crypto.PublicKey
		keySpec cryptoutilCACrypto.KeySpec
		wantErr string
	}{
		{
			name:    "ECDSA public key has wrong type",
			public:  ed25519.PublicKey(make([]byte, ed25519.PublicKeySize)),
			keySpec: cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: pkiAuthorityCurveP384},
			wantErr: pkiTestDoesNotMatchECDSACurve,
		},
		{
			name:    "RSA public key has wrong type",
			public:  ed25519.PublicKey(make([]byte, ed25519.PublicKeySize)),
			keySpec: cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeRSA, RSABits: cryptoutilCACrypto.MinRSAKeyBits},
			wantErr: pkiTestDoesNotMatchRSAKeySize,
		},
		{
			name:    "RSA public key has wrong size",
			public:  &rsa.PublicKey{N: big.NewInt(pkiTestSmallRSAKeyComponent), E: pkiTestSmallRSAKeyComponent},
			keySpec: cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeRSA, RSABits: cryptoutilCACrypto.MinRSAKeyBits},
			wantErr: pkiTestDoesNotMatchRSAKeySize,
		},
		{
			name:    "EdDSA public key has wrong type",
			public:  &rsa.PublicKey{N: big.NewInt(pkiTestSmallRSAKeyComponent), E: pkiTestSmallRSAKeyComponent},
			keySpec: cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeEdDSA},
			wantErr: "does not match configured EdDSA curve",
		},
		{
			name:    "unsupported public key type",
			public:  &rsa.PublicKey{N: big.NewInt(pkiTestSmallRSAKeyComponent), E: pkiTestSmallRSAKeyComponent},
			keySpec: cryptoutilCACrypto.KeySpec{Type: pkiTestUnsupported},
			wantErr: "unsupported configured authority public key type",
		},
		{
			name:    "valid EdDSA public key",
			public:  ed25519.PublicKey(make([]byte, ed25519.PublicKeySize)),
			keySpec: cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeEdDSA},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateAuthorityPublicKey(tc.public, tc.keySpec)
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
		})
	}
}
