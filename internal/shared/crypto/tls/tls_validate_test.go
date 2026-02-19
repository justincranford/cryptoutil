// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"crypto/tls"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestValidateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		config      *tls.Config
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "TLS version too low",
			config: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
			expectError: true,
		},
		{
			name: "InsecureSkipVerify true",
			config: &tls.Config{
				MinVersion:         MinTLSVersion,
				InsecureSkipVerify: true,
			},
			expectError: true,
		},
		{
			name: "valid config",
			config: &tls.Config{
				MinVersion: MinTLSVersion,
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateConfig(tc.config)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStoreCertificatePEM(t *testing.T) {
	t.Parallel()

	// Create a CA chain and server certificate.
	chain, err := CreateCAChain(DefaultCAChainOptions("test.storage"))
	require.NoError(t, err)

	serverSubject, err := chain.CreateEndEntity(ServerEndEntityOptions(
		"server.test.local",
		[]string{"server.test.local"},
		[]net.IP{net.ParseIP("127.0.0.1")},
	))
	require.NoError(t, err)

	// Create temp directory.
	tempDir, err := os.MkdirTemp("", "tls-test-*")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	tests := []struct {
		name        string
		subject     any
		opts        *StorageOptions
		expectError bool
	}{
		{
			name:        "nil subject",
			subject:     nil,
			opts:        DefaultStorageOptions(tempDir),
			expectError: true,
		},
		{
			name:        "nil options",
			subject:     serverSubject,
			opts:        nil,
			expectError: true,
		},
		{
			name:    "empty directory",
			subject: serverSubject,
			opts: &StorageOptions{
				Format:    FormatPEM,
				Directory: "",
			},
			expectError: true,
		},
		{
			name:        "valid PEM storage",
			subject:     serverSubject,
			opts:        DefaultStorageOptions(filepath.Join(tempDir, "pem")),
			expectError: false,
		},
		{
			name:    "PKCS12 not implemented",
			subject: serverSubject,
			opts: &StorageOptions{
				Format:              FormatPKCS12,
				Directory:           filepath.Join(tempDir, "p12"),
				CertificateFilename: "cert.p12",
				IncludePrivateKey:   true,
				FileMode:            cryptoutilSharedMagic.FilePermOwnerReadWriteOnly,
				DirMode:             cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Note: Not parallel due to file system operations in same temp dir.
			var stored *StoredCertificate

			var err error

			if subject, ok := tc.subject.(*cryptoutilSharedCryptoCertificate.Subject); ok {
				stored, err = StoreCertificate(subject, tc.opts)
			} else {
				stored, err = StoreCertificate(nil, tc.opts)
			}

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, stored)
			} else {
				require.NoError(t, err)
				require.NotNil(t, stored)
				require.FileExists(t, stored.CertificatePath)

				if tc.opts.IncludePrivateKey {
					require.FileExists(t, stored.PrivateKeyPath)
				}
			}
		})
	}
}

func TestLoadCertificatePEM(t *testing.T) {
	t.Parallel()

	// Create a CA chain and server certificate.
	chain, err := CreateCAChain(DefaultCAChainOptions("test.load"))
	require.NoError(t, err)

	serverSubject, err := chain.CreateEndEntity(ServerEndEntityOptions(
		"server.test.local",
		[]string{"server.test.local"},
		[]net.IP{net.ParseIP("127.0.0.1")},
	))
	require.NoError(t, err)

	// Create temp directory and store certificate.
	tempDir, err := os.MkdirTemp("", "tls-load-test-*")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.RemoveAll(tempDir)
	})

	stored, err := StoreCertificate(serverSubject, DefaultStorageOptions(tempDir))
	require.NoError(t, err)

	tests := []struct {
		name        string
		certPath    string
		keyPath     string
		expectError bool
	}{
		{
			name:        "nonexistent certificate",
			certPath:    "/nonexistent/cert.pem",
			keyPath:     "",
			expectError: true,
		},
		{
			name:        "valid certificate without key",
			certPath:    stored.CertificatePath,
			keyPath:     "",
			expectError: false,
		},
		{
			name:        "valid certificate with key",
			certPath:    stored.CertificatePath,
			keyPath:     stored.PrivateKeyPath,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			subject, err := LoadCertificatePEM(tc.certPath, tc.keyPath)

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, subject)
			} else {
				require.NoError(t, err)
				require.NotNil(t, subject)
				require.NotEmpty(t, subject.KeyMaterial.CertificateChain)

				if tc.keyPath != "" {
					require.NotNil(t, subject.KeyMaterial.PrivateKey)
				}
			}
		})
	}
}

func TestRootCAsPool(t *testing.T) {
	t.Parallel()

	chain, err := CreateCAChain(DefaultCAChainOptions("test.pool"))
	require.NoError(t, err)

	pool := chain.RootCAsPool()
	require.NotNil(t, pool)
	// The pool should contain the root CA.
}

func TestIntermediateCAsPool(t *testing.T) {
	t.Parallel()

	// Test with chain length 3 (root + 2 intermediates).
	chain, err := CreateCAChain(&CAChainOptions{
		ChainLength:      3,
		CommonNamePrefix: "test.intermediate",
		Duration:         time.Hour,
	})
	require.NoError(t, err)

	pool := chain.IntermediateCAsPool()
	require.NotNil(t, pool)
	// The pool should contain intermediate CAs (excluding root).
}

func TestCreateCAChain_AllCurves(t *testing.T) {
	t.Parallel()

	curves := []struct {
		name  string
		curve ECCurve
	}{
		{"P256 default", CurveP256},
		{"P384", CurveP384},
		{"P521", CurveP521},
	}

	for _, tc := range curves {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			chain, err := CreateCAChain(&CAChainOptions{
				ChainLength:      1,
				CommonNamePrefix: "test.curve",
				Duration:         time.Hour,
				Curve:            tc.curve,
			})
			require.NoError(t, err)
			require.NotNil(t, chain)
			require.Len(t, chain.CAs, 1)

			// Verify chain created successfully with specified curve.
			require.NotNil(t, chain.IssuingCA)
			require.NotNil(t, chain.RootCA)
		})
	}
}
