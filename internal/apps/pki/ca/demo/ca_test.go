// Copyright (c) 2025 Justin Cranford
//
//

package demo

import (
	"testing"

	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"

	"github.com/stretchr/testify/require"
)

func TestGetDemoCA(t *testing.T) {
	t.Parallel()

	ca, err := GetDemoCA()
	require.NoError(t, err)
	require.NotNil(t, ca)
	require.NotNil(t, ca.Chain)
	require.NotNil(t, ca.Chain.RootCA)
	require.NotNil(t, ca.Chain.IssuingCA)
	require.Len(t, ca.Chain.CAs, cryptoutilSharedCryptoTls.DefaultCAChainLength)

	// Second call should return same instance.
	ca2, err := GetDemoCA()
	require.NoError(t, err)
	require.Same(t, ca, ca2)
}

func TestCreateDemoCA(t *testing.T) {
	t.Parallel()

	ca, err := CreateDemoCA()
	require.NoError(t, err)
	require.NotNil(t, ca)
	require.NotNil(t, ca.Chain)

	// Should create different instance each time.
	ca2, err := CreateDemoCA()
	require.NoError(t, err)
	require.NotSame(t, ca, ca2)
}

func TestCreateDemoCAWithOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		opts        *cryptoutilSharedCryptoTls.CAChainOptions
		expectError bool
	}{
		{
			name:        "nil options uses default",
			opts:        nil,
			expectError: false,
		},
		{
			name: "custom valid options",
			opts: &cryptoutilSharedCryptoTls.CAChainOptions{
				ChainLength:      2,
				CommonNamePrefix: "custom.demo.local",
				CNStyle:          cryptoutilSharedCryptoTls.CNStyleFQDN,
				Duration:         cryptoutilSharedCryptoTls.DefaultCADuration,
				Curve:            cryptoutilSharedCryptoTls.DefaultECCurve,
			},
			expectError: false,
		},
		{
			name: "descriptive style",
			opts: &cryptoutilSharedCryptoTls.CAChainOptions{
				ChainLength:      1,
				CommonNamePrefix: "Custom Demo CA",
				CNStyle:          cryptoutilSharedCryptoTls.CNStyleDescriptive,
				Duration:         cryptoutilSharedCryptoTls.DefaultCADuration,
				Curve:            cryptoutilSharedCryptoTls.DefaultECCurve,
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ca, err := CreateDemoCAWithOptions(tc.opts)

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, ca)
			} else {
				require.NoError(t, err)
				require.NotNil(t, ca)
				require.NotNil(t, ca.Chain)
			}
		})
	}
}

func TestDemoCACreateServerCertificate(t *testing.T) {
	t.Parallel()

	ca, err := CreateDemoCA()
	require.NoError(t, err)

	tests := []struct {
		name        string
		serverName  string
		expectError bool
	}{
		{
			name:        "empty server name",
			serverName:  "",
			expectError: true,
		},
		{
			name:        "valid server name",
			serverName:  "kms.cryptoutil.demo.local",
			expectError: false,
		},
		{
			name:        "another valid server",
			serverName:  "identity.cryptoutil.demo.local",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			subject, err := ca.CreateServerCertificate(tc.serverName)

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, subject)
			} else {
				require.NoError(t, err)
				require.NotNil(t, subject)
				require.Equal(t, tc.serverName, subject.SubjectName)
			}
		})
	}
}

func TestDemoCACreateClientCertificate(t *testing.T) {
	t.Parallel()

	ca, err := CreateDemoCA()
	require.NoError(t, err)

	tests := []struct {
		name        string
		clientName  string
		expectError bool
	}{
		{
			name:        "empty client name",
			clientName:  "",
			expectError: true,
		},
		{
			name:        "valid client name",
			clientName:  "client.cryptoutil.demo.local",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			subject, err := ca.CreateClientCertificate(tc.clientName)

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, subject)
			} else {
				require.NoError(t, err)
				require.NotNil(t, subject)
				require.Equal(t, tc.clientName, subject.SubjectName)
			}
		})
	}
}

func TestDemoCAPoolMethods(t *testing.T) {
	t.Parallel()

	ca, err := CreateDemoCA()
	require.NoError(t, err)

	t.Run("RootCAsPool", func(t *testing.T) {
		t.Parallel()

		pool := ca.RootCAsPool()
		require.NotNil(t, pool)
	})

	t.Run("IntermediateCAsPool", func(t *testing.T) {
		t.Parallel()

		pool := ca.IntermediateCAsPool()
		require.NotNil(t, pool)
	})
}

func TestGetDemoCAMultipleCalls(t *testing.T) {
	t.Parallel()

	ca1, err1 := GetDemoCA()
	require.NoError(t, err1)
	require.NotNil(t, ca1)

	ca2, err2 := GetDemoCA()
	require.NoError(t, err2)
	require.NotNil(t, ca2)
	require.Same(t, ca1, ca2)
	require.Same(t, ca1.Chain, ca2.Chain)
}

func TestCreateDemoCAChainValidation(t *testing.T) {
	t.Parallel()

	ca, err := CreateDemoCA()
	require.NoError(t, err)
	require.NotNil(t, ca)
	require.NotNil(t, ca.Chain)
	require.NotNil(t, ca.Chain.RootCA)
	require.NotNil(t, ca.Chain.IssuingCA)
	require.Len(t, ca.Chain.CAs, cryptoutilSharedCryptoTls.DefaultCAChainLength)
	require.NotEmpty(t, ca.Chain.RootCA.KeyMaterial.CertificateChain)
	require.NotEmpty(t, ca.Chain.IssuingCA.KeyMaterial.CertificateChain)
}

func TestCreateDemoCAWithOptionsDefaultsWhenNil(t *testing.T) {
	t.Parallel()

	ca, err := CreateDemoCAWithOptions(nil)
	require.NoError(t, err)
	require.NotNil(t, ca)
	require.NotNil(t, ca.Chain)
	require.Len(t, ca.Chain.CAs, cryptoutilSharedCryptoTls.DefaultCAChainLength)
}

func TestCreateServerCertificateFullPath(t *testing.T) {
	t.Parallel()

	ca, err := CreateDemoCA()
	require.NoError(t, err)

	subject, err := ca.CreateServerCertificate("testserver.local")
	require.NoError(t, err)
	require.NotNil(t, subject)
	require.NotNil(t, subject.KeyMaterial)
	require.NotNil(t, subject.KeyMaterial.CertificateChain)
	require.NotEmpty(t, subject.KeyMaterial.CertificateChain)
	require.NotNil(t, subject.KeyMaterial.PrivateKey)
	require.Equal(t, "testserver.local", subject.SubjectName)
}

func TestCreateClientCertificateFullPath(t *testing.T) {
	t.Parallel()

	ca, err := CreateDemoCA()
	require.NoError(t, err)

	subject, err := ca.CreateClientCertificate("testclient.local")
	require.NoError(t, err)
	require.NotNil(t, subject)
	require.NotNil(t, subject.KeyMaterial)
	require.NotNil(t, subject.KeyMaterial.CertificateChain)
	require.NotEmpty(t, subject.KeyMaterial.CertificateChain)
	require.NotNil(t, subject.KeyMaterial.PrivateKey)
	require.Equal(t, "testclient.local", subject.SubjectName)
}
