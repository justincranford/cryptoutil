// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"crypto/tls"
	"net"
	"testing"
	"time"


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
