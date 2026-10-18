// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"bytes"
	"crypto"
	crand "crypto/rand"
	"crypto/x509"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

type pkiAuthorityTestProvider struct {
	cryptoutilCACrypto.Provider
	generateKeyPair func(cryptoutilCACrypto.KeySpec) (*cryptoutilCACrypto.KeyPair, error)
	signatureAlg    func(crypto.PublicKey) (x509.SignatureAlgorithm, error)
}

func (p pkiAuthorityTestProvider) GenerateKeyPair(spec cryptoutilCACrypto.KeySpec) (*cryptoutilCACrypto.KeyPair, error) {
	if p.generateKeyPair != nil {
		return p.generateKeyPair(spec)
	}

	return p.Provider.GenerateKeyPair(spec)
}

func (p pkiAuthorityTestProvider) GetSignatureAlgorithm(publicKey crypto.PublicKey) (x509.SignatureAlgorithm, error) {
	if p.signatureAlg != nil {
		return p.signatureAlg(publicKey)
	}

	return p.Provider.GetSignatureAlgorithm(publicKey)
}

type pkiAuthorityTestSigner struct {
	publicKey crypto.PublicKey
	err       error
}

func (s pkiAuthorityTestSigner) Public() crypto.PublicKey {
	return s.publicKey
}

func (s pkiAuthorityTestSigner) Sign(io.Reader, []byte, crypto.SignerOpts) ([]byte, error) {
	return nil, s.err
}

type pkiAuthorityFailingReader struct{}

func (pkiAuthorityFailingReader) Read([]byte) (int, error) {
	return 0, errors.New("injected authority random failure")
}

func TestCreateSelfSignedCAWithConfigFailures(t *testing.T) {
	t.Parallel()

	baseProvider := cryptoutilCACrypto.NewSoftwareProvider()
	keyPairFailure := errors.New("injected authority key generation failure")
	signatureFailure := errors.New("injected authority signature selection failure")
	signingFailure := errors.New("injected authority signing failure")
	tests := []struct {
		name         string
		provider     cryptoutilCACrypto.Provider
		randomReader io.Reader
		wantErr      string
	}{
		{name: pkiTestNilProvider, randomReader: crand.Reader, wantErr: cryptoutilSharedMagic.PKIErrorCryptoProviderRequired},
		{name: "nil random reader", provider: baseProvider, wantErr: pkiTestRandomReaderRequired},
		{
			name:         "serial random failure",
			provider:     baseProvider,
			randomReader: pkiAuthorityFailingReader{},
			wantErr:      pkiTestFailedGenerateSerialNumber,
		},
		{
			name:         pkiTestKeyGenerationFailure,
			provider:     pkiAuthorityTestProvider{Provider: baseProvider, generateKeyPair: func(cryptoutilCACrypto.KeySpec) (*cryptoutilCACrypto.KeyPair, error) { return nil, keyPairFailure }},
			randomReader: crand.Reader,
			wantErr:      pkiTestFailedGenerateCAKey,
		},
		{
			name:         "signature algorithm selection failure",
			provider:     pkiAuthorityTestProvider{Provider: baseProvider, signatureAlg: func(crypto.PublicKey) (x509.SignatureAlgorithm, error) { return 0, signatureFailure }},
			randomReader: crand.Reader,
			wantErr:      pkiTestFailedGetSignatureAlgorithm,
		},
		{
			name: pkiTestCertificateSigningFailure,
			provider: pkiAuthorityTestProvider{
				Provider: baseProvider,
				generateKeyPair: func(spec cryptoutilCACrypto.KeySpec) (*cryptoutilCACrypto.KeyPair, error) {
					pair, err := baseProvider.GenerateKeyPair(spec)
					if err != nil {
						return nil, err
					}

					pair.PrivateKey = pkiAuthorityTestSigner{publicKey: pair.PublicKey, err: signingFailure}

					return pair, nil
				},
			},
			randomReader: crand.Reader,
			wantErr:      pkiTestFailedCreateCACertificate,
		},
		{
			name: "generated key does not match configured curve",
			provider: pkiAuthorityTestProvider{Provider: baseProvider, generateKeyPair: func(cryptoutilCACrypto.KeySpec) (*cryptoutilCACrypto.KeyPair, error) {
				return baseProvider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: pkiAuthorityCurveP256})
			}},
			randomReader: crand.Reader,
			wantErr:      "generated CA failed lifecycle validation",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := DefaultPKIAuthorityConfig()
			_, _, err := createSelfSignedCAWithConfig(tc.provider, tc.randomReader, config)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestCreateSelfSignedCAWithConfigSerialFloor(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	zeroSerialThenRandom := io.MultiReader(bytes.NewReader(make([]byte, 64)), crand.Reader)
	certificate, _, err := createSelfSignedCAWithConfig(provider, zeroSerialThenRandom, DefaultPKIAuthorityConfig())
	require.NoError(t, err)
	require.Equal(t, "1", certificate.SerialNumber.String())
}
