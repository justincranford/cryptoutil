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

func TestValidateFQDN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fqdn        string
		expectError bool
	}{
		{
			name:        "empty string",
			fqdn:        "",
			expectError: true,
		},
		{
			name:        "valid simple",
			fqdn:        "example.com",
			expectError: false,
		},
		{
			name:        "valid with subdomain",
			fqdn:        "kms.cryptoutil.demo.local",
			expectError: false,
		},
		{
			name:        "valid single label",
			fqdn:        "localhost",
			expectError: false,
		},
		{
			name:        "invalid starts with hyphen",
			fqdn:        "-invalid.com",
			expectError: true,
		},
		{
			name:        "invalid ends with hyphen",
			fqdn:        "invalid-.com",
			expectError: true,
		},
		{
			name:        "invalid has underscore",
			fqdn:        "invalid_name.com",
			expectError: true,
		},
		{
			name:        "valid with hyphen",
			fqdn:        "my-service.example.com",
			expectError: false,
		},
		{
			name:        "valid alphanumeric",
			fqdn:        "service123.example456.com",
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateFQDN(tc.fqdn)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCreateCAChain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		opts        *CAChainOptions
		expectError bool
	}{
		{
			name:        "nil options",
			opts:        nil,
			expectError: true,
		},
		{
			name: "zero chain length",
			opts: &CAChainOptions{
				ChainLength:      0,
				CommonNamePrefix: "test.chain",
				Duration:         time.Hour,
			},
			expectError: true,
		},
		{
			name: "empty common name prefix",
			opts: &CAChainOptions{
				ChainLength:      1,
				CommonNamePrefix: "",
				Duration:         time.Hour,
			},
			expectError: true,
		},
		{
			name: "negative duration",
			opts: &CAChainOptions{
				ChainLength:      1,
				CommonNamePrefix: "test.chain",
				Duration:         -time.Hour,
			},
			expectError: true,
		},
		{
			name:        "valid single CA FQDN style",
			opts:        DefaultCAChainOptions("test.single"),
			expectError: false,
		},
		{
			name: "valid chain length 3 FQDN style",
			opts: &CAChainOptions{
				ChainLength:      3,
				CommonNamePrefix: "test.chain3",
				CNStyle:          CNStyleFQDN,
				Duration:         time.Hour,
			},
			expectError: false,
		},
		{
			name: "valid chain length 3 descriptive style",
			opts: &CAChainOptions{
				ChainLength:      3,
				CommonNamePrefix: "Test CA Chain",
				CNStyle:          CNStyleDescriptive,
				Duration:         time.Hour,
			},
			expectError: false,
		},
		{
			name: "invalid FQDN prefix with FQDN style",
			opts: &CAChainOptions{
				ChainLength:      1,
				CommonNamePrefix: "invalid_prefix",
				CNStyle:          CNStyleFQDN,
				Duration:         time.Hour,
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			chain, err := CreateCAChain(tc.opts)

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, chain)
			} else {
				require.NoError(t, err)
				require.NotNil(t, chain)
				require.NotNil(t, chain.IssuingCA)
				require.NotNil(t, chain.RootCA)
				require.Len(t, chain.CAs, tc.opts.ChainLength)
			}
		})
	}
}

func TestCreateEndEntity(t *testing.T) {
	t.Parallel()

	// Create a CA chain first.
	chain, err := CreateCAChain(DefaultCAChainOptions("test.ee"))
	require.NoError(t, err)

	tests := []struct {
		name        string
		opts        *EndEntityOptions
		expectError bool
	}{
		{
			name:        "nil options",
			opts:        nil,
			expectError: true,
		},
		{
			name: "empty subject name",
			opts: &EndEntityOptions{
				SubjectName: "",
			},
			expectError: true,
		},
		{
			name:        "valid server certificate",
			opts:        ServerEndEntityOptions("server.test.local", []string{"server.test.local", "localhost"}, []net.IP{net.ParseIP("127.0.0.1")}),
			expectError: false,
		},
		{
			name:        "valid client certificate",
			opts:        ClientEndEntityOptions("client.test.local"),
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			subject, err := chain.CreateEndEntity(tc.opts)

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, subject)
			} else {
				require.NoError(t, err)
				require.NotNil(t, subject)
				require.Equal(t, tc.opts.SubjectName, subject.SubjectName)
				require.NotNil(t, subject.KeyMaterial.CertificateChain)
				require.NotNil(t, subject.KeyMaterial.PrivateKey)
			}
		})
	}
}

func TestNewServerConfig(t *testing.T) {
	t.Parallel()

	// Create a CA chain and server certificate.
	chain, err := CreateCAChain(DefaultCAChainOptions("test.server"))
	require.NoError(t, err)

	serverSubject, err := chain.CreateEndEntity(ServerEndEntityOptions(
		"server.test.local",
		[]string{"server.test.local"},
		[]net.IP{net.ParseIP("127.0.0.1")},
	))
	require.NoError(t, err)

	tests := []struct {
		name        string
		opts        *ServerConfigOptions
		expectError bool
	}{
		{
			name:        "nil options",
			opts:        nil,
			expectError: true,
		},
		{
			name: "nil subject",
			opts: &ServerConfigOptions{
				Subject: nil,
			},
			expectError: true,
		},
		{
			name: "valid server config no client auth",
			opts: &ServerConfigOptions{
				Subject:    serverSubject,
				ClientAuth: tls.NoClientCert,
			},
			expectError: false,
		},
		{
			name: "valid server config with mTLS",
			opts: &ServerConfigOptions{
				Subject:    serverSubject,
				ClientAuth: tls.RequireAndVerifyClientCert,
				ClientCAs:  chain.RootCAsPool(),
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, err := NewServerConfig(tc.opts)

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				require.NotNil(t, config.TLSConfig)
				require.Equal(t, uint16(MinTLSVersion), config.TLSConfig.MinVersion)
			}
		})
	}
}

func TestNewClientConfig(t *testing.T) {
	t.Parallel()

	// Create a CA chain and client certificate.
	chain, err := CreateCAChain(DefaultCAChainOptions("test.client"))
	require.NoError(t, err)

	clientSubject, err := chain.CreateEndEntity(ClientEndEntityOptions("client.test.local"))
	require.NoError(t, err)

	tests := []struct {
		name        string
		opts        *ClientConfigOptions
		expectError bool
	}{
		{
			name:        "nil options",
			opts:        nil,
			expectError: true,
		},
		{
			name: "valid client config no mTLS",
			opts: &ClientConfigOptions{
				RootCAs:    chain.RootCAsPool(),
				ServerName: "server.test.local",
			},
			expectError: false,
		},
		{
			name: "valid client config with mTLS",
			opts: &ClientConfigOptions{
				ClientSubject: clientSubject,
				RootCAs:       chain.RootCAsPool(),
				ServerName:    "server.test.local",
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, err := NewClientConfig(tc.opts)

			if tc.expectError {
				require.Error(t, err)
				require.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				require.NotNil(t, config.TLSConfig)
				require.Equal(t, uint16(MinTLSVersion), config.TLSConfig.MinVersion)
			}
		})
	}
}

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
