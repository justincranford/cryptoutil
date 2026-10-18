// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"errors"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateSelfSignedTLSConfigWithGenerators_Errors(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected generation failure")

	tests := []struct {
		name                   string
		generatePrivateKeyFn   func() (*ecdsa.PrivateKey, error)
		generateSerialNumberFn func() (*big.Int, error)
		createCertificateFn    func(*x509.Certificate, *ecdsa.PrivateKey) ([]byte, error)
		wantError              string
	}{
		{
			name: "key generation",
			generatePrivateKeyFn: func() (*ecdsa.PrivateKey, error) {
				return nil, injectedErr
			},
			wantError: "failed to generate private key",
		},
		{
			name: "serial number generation",
			generateSerialNumberFn: func() (*big.Int, error) {
				return nil, injectedErr
			},
			wantError: "failed to generate serial number",
		},
		{
			name: "certificate creation",
			createCertificateFn: func(*x509.Certificate, *ecdsa.PrivateKey) ([]byte, error) {
				return nil, injectedErr
			},
			wantError: "failed to create certificate",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			generatePrivateKeyFn := func() (*ecdsa.PrivateKey, error) {
				return ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
			}
			generateSerialNumberFn := func() (*big.Int, error) {
				return big.NewInt(1), nil
			}
			createCertificateFn := func(*x509.Certificate, *ecdsa.PrivateKey) ([]byte, error) {
				return nil, nil
			}

			if tc.generatePrivateKeyFn != nil {
				generatePrivateKeyFn = tc.generatePrivateKeyFn
			}

			if tc.generateSerialNumberFn != nil {
				generateSerialNumberFn = tc.generateSerialNumberFn
			}

			if tc.createCertificateFn != nil {
				createCertificateFn = tc.createCertificateFn
			}

			tlsConfig, err := generateSelfSignedTLSConfigWithGenerators(
				generatePrivateKeyFn,
				generateSerialNumberFn,
				createCertificateFn,
			)

			require.ErrorContains(t, err, tc.wantError)
			require.ErrorIs(t, err, injectedErr)
			require.Nil(t, tlsConfig)
		})
	}
}
