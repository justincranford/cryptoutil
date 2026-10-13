// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"crypto"
	crand "crypto/rand"
	"crypto/x509"
	"errors"
	"io"
	"testing"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

type pkiComponentTestProvider struct {
	cryptoutilCACrypto.Provider
	generateKeyPairFn       func(cryptoutilCACrypto.KeySpec) (*cryptoutilCACrypto.KeyPair, error)
	getSignatureAlgorithmFn func(crypto.PublicKey) (x509.SignatureAlgorithm, error)
}

func (p pkiComponentTestProvider) GenerateKeyPair(spec cryptoutilCACrypto.KeySpec) (*cryptoutilCACrypto.KeyPair, error) {
	if p.generateKeyPairFn != nil {
		return p.generateKeyPairFn(spec)
	}

	return p.Provider.GenerateKeyPair(spec)
}

func (p pkiComponentTestProvider) GetSignatureAlgorithm(publicKey crypto.PublicKey) (x509.SignatureAlgorithm, error) {
	if p.getSignatureAlgorithmFn != nil {
		return p.getSignatureAlgorithmFn(publicKey)
	}

	return p.Provider.GetSignatureAlgorithm(publicKey)
}

type pkiComponentErrorSigner struct {
	publicKey crypto.PublicKey
}

func (s pkiComponentErrorSigner) Public() crypto.PublicKey {
	return s.publicKey
}

func (pkiComponentErrorSigner) Sign(io.Reader, []byte, crypto.SignerOpts) ([]byte, error) {
	return nil, errors.New("injected certificate signing failure")
}

type pkiComponentErrorReader struct{}

func (pkiComponentErrorReader) Read([]byte) (int, error) {
	return 0, io.EOF
}

type pkiComponentZeroPrefixReader struct {
	zeroBytesRemaining int
}

func (r *pkiComponentZeroPrefixReader) Read(buffer []byte) (int, error) {
	if r.zeroBytesRemaining == 0 {
		return crand.Reader.Read(buffer)
	}

	count := len(buffer)
	if count > r.zeroBytesRemaining {
		count = r.zeroBytesRemaining
	}

	clear(buffer[:count])

	r.zeroBytesRemaining -= count
	if count == len(buffer) {
		return count, nil
	}

	read, err := crand.Reader.Read(buffer[count:])

	return count + read, err
}

func TestCreateSelfSignedCA_Results(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		provider         func() cryptoutilCACrypto.Provider
		randomReader     func() io.Reader
		wantErr          string
		wantSerialNumber string
	}{
		{
			name: "success",
			provider: func() cryptoutilCACrypto.Provider {
				return cryptoutilCACrypto.NewSoftwareProvider()
			},
			randomReader: func() io.Reader { return crand.Reader },
		},
		{
			name: "nil randomness source",
			provider: func() cryptoutilCACrypto.Provider {
				return cryptoutilCACrypto.NewSoftwareProvider()
			},
			randomReader: func() io.Reader { return nil },
			wantErr:      "random reader is required",
		},
		{
			name: "zero serial number is normalized",
			provider: func() cryptoutilCACrypto.Provider {
				return cryptoutilCACrypto.NewSoftwareProvider()
			},
			randomReader: func() io.Reader {
				return &pkiComponentZeroPrefixReader{zeroBytesRemaining: cryptoutilSharedMagic.SerialNumberLength}
			},
			wantSerialNumber: "1",
		},
		{
			name: "serial number randomness failure",
			provider: func() cryptoutilCACrypto.Provider {
				return cryptoutilCACrypto.NewSoftwareProvider()
			},
			randomReader: func() io.Reader { return pkiComponentErrorReader{} },
			wantErr:      "failed to generate serial number",
		},
		{
			name: "key generation failure",
			provider: func() cryptoutilCACrypto.Provider {
				return pkiComponentTestProvider{
					Provider: cryptoutilCACrypto.NewSoftwareProvider(),
					generateKeyPairFn: func(_ cryptoutilCACrypto.KeySpec) (*cryptoutilCACrypto.KeyPair, error) {
						return nil, errors.New("injected key generation failure")
					},
				}
			},
			randomReader: func() io.Reader { return crand.Reader },
			wantErr:      "failed to generate CA key",
		},
		{
			name: "generated private key is not a signer",
			provider: func() cryptoutilCACrypto.Provider {
				return pkiComponentTestProvider{
					Provider: cryptoutilCACrypto.NewSoftwareProvider(),
					generateKeyPairFn: func(_ cryptoutilCACrypto.KeySpec) (*cryptoutilCACrypto.KeyPair, error) {
						return &cryptoutilCACrypto.KeyPair{PrivateKey: struct{}{}}, nil
					},
				}
			},
			randomReader: func() io.Reader { return crand.Reader },
			wantErr:      "private key is not a signer",
		},
		{
			name: "signature algorithm lookup failure",
			provider: func() cryptoutilCACrypto.Provider {
				return pkiComponentTestProvider{
					Provider: cryptoutilCACrypto.NewSoftwareProvider(),
					getSignatureAlgorithmFn: func(_ crypto.PublicKey) (x509.SignatureAlgorithm, error) {
						return x509.UnknownSignatureAlgorithm, errors.New("injected signature lookup failure")
					},
				}
			},
			randomReader: func() io.Reader { return crand.Reader },
			wantErr:      "failed to get signature algorithm",
		},
		{
			name: "certificate signing failure",
			provider: func() cryptoutilCACrypto.Provider {
				baseProvider := cryptoutilCACrypto.NewSoftwareProvider()

				return pkiComponentTestProvider{
					Provider: baseProvider,
					generateKeyPairFn: func(spec cryptoutilCACrypto.KeySpec) (*cryptoutilCACrypto.KeyPair, error) {
						keyPair, err := baseProvider.GenerateKeyPair(spec)
						if err != nil {
							return nil, err
						}

						keyPair.PrivateKey = pkiComponentErrorSigner{publicKey: keyPair.PublicKey}

						return keyPair, nil
					},
				}
			},
			randomReader: func() io.Reader { return crand.Reader },
			wantErr:      "failed to create CA certificate",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			certificate, privateKey, err := createSelfSignedCA(tc.provider(), tc.randomReader())
			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
				require.Nil(t, certificate)
				require.Nil(t, privateKey)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, certificate)
			require.True(t, certificate.IsCA)
			require.Greater(t, certificate.SerialNumber.Sign(), 0)
			require.LessOrEqual(t, certificate.SerialNumber.BitLen(), cryptoutilSharedMagic.PKICASerialNumberBits)

			if tc.wantSerialNumber != "" {
				require.Equal(t, tc.wantSerialNumber, certificate.SerialNumber.String())
			}

			_, ok := privateKey.(crypto.Signer)
			require.True(t, ok)
		})
	}
}
