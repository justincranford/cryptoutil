// Copyright (c) 2025-2026 Justin Cranford.
package bootstrap

import (
	"crypto"
	"crypto/x509"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilCACrypto "cryptoutil/internal/apps/sm-kms/server/pki/crypto"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestBootstrapper_BootstrapSigningAndPersistenceErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		configure func(*testing.T, *bootstrapFaultProvider, *RootCAConfig)
		wantErr   string
	}{
		{
			name: "signature algorithm selection failure",
			configure: func(_ *testing.T, provider *bootstrapFaultProvider, _ *RootCAConfig) {
				provider.signatureErr = errors.New("injected signature algorithm failure")
			},
			wantErr: "failed to select signature algorithm",
		},
		{
			name: "certificate signing failure",
			configure: func(_ *testing.T, provider *bootstrapFaultProvider, _ *RootCAConfig) {
				provider.wrapPrivateKey = true
				provider.signErr = errors.New("injected certificate signing failure")
			},
			wantErr: "failed to create certificate",
		},
		{
			name: "unencodable signer persistence failure",
			configure: func(t *testing.T, provider *bootstrapFaultProvider, config *RootCAConfig) {
				provider.wrapPrivateKey = true
				config.OutputDir = t.TempDir()
			},
			wantErr: "failed to marshal private key",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			provider := &bootstrapFaultProvider{Provider: cryptoutilCACrypto.NewSoftwareProvider()}

			config := &RootCAConfig{
				Name: "Bootstrap Test Root",
				KeySpec: cryptoutilCACrypto.KeySpec{
					Type:       cryptoutilCACrypto.KeyTypeECDSA,
					ECDSACurve: cryptoutilSharedMagic.P256,
				},
				ValidityDuration:  time.Hour * 24 * 365,
				PathLenConstraint: 1,
			}
			if tc.configure != nil {
				tc.configure(t, provider, config)
			}

			rootCA, audit, err := NewBootstrapper(provider).Bootstrap(config)
			require.ErrorContains(t, err, tc.wantErr)
			require.Nil(t, rootCA)
			require.Nil(t, audit)
		})
	}
}

type bootstrapFaultProvider struct {
	cryptoutilCACrypto.Provider
	signatureErr   error
	wrapPrivateKey bool
	signErr        error
}

func (p *bootstrapFaultProvider) GenerateKeyPair(spec cryptoutilCACrypto.KeySpec) (*cryptoutilCACrypto.KeyPair, error) {
	keyPair, err := p.Provider.GenerateKeyPair(spec)
	if err != nil {
		return nil, err
	}

	if p.wrapPrivateKey {
		signer, ok := keyPair.PrivateKey.(crypto.Signer)
		if !ok {
			return nil, errors.New("generated private key is not a signer")
		}

		keyPair.PrivateKey = bootstrapTestSigner{signer: signer, signErr: p.signErr}
	}

	return keyPair, nil
}

func (p *bootstrapFaultProvider) GetSignatureAlgorithm(publicKey crypto.PublicKey) (x509.SignatureAlgorithm, error) {
	if p.signatureErr != nil {
		return x509.UnknownSignatureAlgorithm, p.signatureErr
	}

	return p.Provider.GetSignatureAlgorithm(publicKey)
}

type bootstrapTestSigner struct {
	signer  crypto.Signer
	signErr error
}

func (s bootstrapTestSigner) Public() crypto.PublicKey {
	return s.signer.Public()
}

func (s bootstrapTestSigner) Sign(random io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	if s.signErr != nil {
		return nil, s.signErr
	}

	return s.signer.Sign(random, digest, opts)
}
