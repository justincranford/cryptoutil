// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"strings"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

// TestValidateFQDN_ExactMaxLength kills the boundary mutant at chain.go:39
// that changes `len(name) > FQDNMaxLength` to `>= FQDNMaxLength`.
func TestValidateFQDN_ExactMaxLength(t *testing.T) {
	t.Parallel()

	// Build a valid FQDN with exactly FQDNMaxLength (253) characters.
	// Using short labels to stay within label limits (63 chars each).
	// Pattern: "aaa.aaa.aaa..." to reach exactly 253 characters.
	labelLen := 62
	label := strings.Repeat("a", labelLen)
	// Each label+dot = 63 chars. 253/63 = 4 segments, with 1 char remaining.
	// 4 segments * (62 chars + 1 dot) = 252 chars + 1 final char = 253.
	name := label + "." + label + "." + label + "." + label + ".a"
	require.Equal(t, cryptoutilSharedMagic.FQDNMaxLength, len(name), "test name must be exactly FQDN max length")

	err := ValidateFQDN(name)
	require.NoError(t, err, "FQDN with exactly max length should be valid")
}

// TestValidateFQDN_ExactLabelMaxLength kills the boundary mutant at chain.go:50
// that changes `len(label) > FQDNLabelMaxLength` to `>= FQDNLabelMaxLength`.
func TestValidateFQDN_ExactLabelMaxLength(t *testing.T) {
	t.Parallel()

	// Build a valid FQDN with a label of exactly FQDNLabelMaxLength (63) characters.
	label := strings.Repeat("a", cryptoutilSharedMagic.FQDNLabelMaxLength)
	name := label + ".com"
	require.Equal(t, cryptoutilSharedMagic.FQDNLabelMaxLength, len(label), "label must be exactly label max length")

	err := ValidateFQDN(name)
	require.NoError(t, err, "label with exactly max length should be valid")
}

// TestCreateCAChain_ZeroDuration kills the boundary mutant at chain.go:223
// that changes `opts.Duration <= 0` to `opts.Duration < 0`.
func TestCreateCAChain_ZeroDuration(t *testing.T) {
	t.Parallel()

	chain, err := CreateCAChain(&CAChainOptions{
		ChainLength:      1,
		CommonNamePrefix: "test.zero.duration",
		Duration:         0, // Zero duration must be rejected.
	})
	require.Error(t, err)
	require.Nil(t, chain)
	require.Contains(t, err.Error(), "duration must be positive")
}

// TestCreateEndEntity_ZeroDuration kills the boundary mutant at chain.go:295
// that changes `duration <= 0` to `duration < 0`.
func TestCreateEndEntity_ZeroDuration(t *testing.T) {
	t.Parallel()

	chain, err := CreateCAChain(DefaultCAChainOptions("test.ee.zerodur"))
	require.NoError(t, err)

	subject, err := chain.CreateEndEntity(&EndEntityOptions{
		SubjectName: "test.ee.zerodur",
		Duration:    0, // Zero duration should get default, not pass through as zero.
	})
	require.NoError(t, err)
	require.NotNil(t, subject)

	// Verify cert has meaningful validity (DefaultEndEntityDuration, not zero).
	cert := subject.KeyMaterial.CertificateChain[0]
	validity := cert.NotAfter.Sub(cert.NotBefore)
	require.Greater(t, validity, time.Hour, "cert with Duration=0 should get default duration, not zero")
}

// TestCreateEndEntity_CustomDuration kills the negation mutant at chain.go:295
// that changes `duration <= 0` to `duration > 0`.
func TestCreateEndEntity_CustomDuration(t *testing.T) {
	t.Parallel()

	chain, err := CreateCAChain(DefaultCAChainOptions("test.ee.custom"))
	require.NoError(t, err)

	customDuration := 2 * time.Hour
	subject, err := chain.CreateEndEntity(&EndEntityOptions{
		SubjectName: "test.ee.custom",
		Duration:    customDuration,
	})
	require.NoError(t, err)
	require.NotNil(t, subject)

	// Verify cert validity is close to the custom duration (not the default 365 days).
	cert := subject.KeyMaterial.CertificateChain[0]
	validity := cert.NotAfter.Sub(cert.NotBefore)
	// Allow some tolerance for not-before randomization.
	require.Less(t, validity, 24*time.Hour, "custom 2h duration cert should not have default 365-day validity")
}

// TestRootCAsPool_ContainsRootCert kills negation mutants at chain.go:321
// that change `c.RootCA != nil` to `== nil` or negate cert chain length checks.
func TestRootCAsPool_ContainsRootCert(t *testing.T) {
	t.Parallel()

	chain, err := CreateCAChain(&CAChainOptions{
		ChainLength:      3,
		CommonNamePrefix: "test.rootpool",
		Duration:         time.Hour,
	})
	require.NoError(t, err)

	pool := chain.RootCAsPool()
	require.NotNil(t, pool)

	// Verify pool contains exactly the root CA cert (not empty, not all CAs).
	expectedPool := x509.NewCertPool()
	expectedPool.AddCert(chain.RootCA.KeyMaterial.CertificateChain[0])
	require.True(t, pool.Equal(expectedPool), "RootCAsPool should contain exactly the root CA cert")
}

// TestIntermediateCAsPool_ExcludesRoot kills boundary/negation mutants at chain.go:332-333
// that would include root in pool or make pool empty.
func TestIntermediateCAsPool_ExcludesRoot(t *testing.T) {
	t.Parallel()

	chain, err := CreateCAChain(&CAChainOptions{
		ChainLength:      3,
		CommonNamePrefix: "test.intpool",
		Duration:         time.Hour,
	})
	require.NoError(t, err)

	pool := chain.IntermediateCAsPool()
	require.NotNil(t, pool)

	// Build expected pool with only intermediate CAs (first N-1 CAs, excluding root which is last).
	expectedPool := x509.NewCertPool()
	for i := 0; i < len(chain.CAs)-1; i++ {
		expectedPool.AddCert(chain.CAs[i].KeyMaterial.CertificateChain[0])
	}

	require.True(t, pool.Equal(expectedPool), "IntermediateCAsPool should contain only intermediate CAs, not root")

	// Verify root is NOT in the intermediate pool.
	rootOnlyPool := x509.NewCertPool()
	rootOnlyPool.AddCert(chain.RootCA.KeyMaterial.CertificateChain[0])
	require.False(t, pool.Equal(rootOnlyPool), "IntermediateCAsPool should not equal root-only pool")
}

// TestNewServerConfig_ClientAuthFallback kills negation mutants at config.go:84-85
// that negate ClientAuth comparisons for VerifyClientCertIfGiven and RequireAnyClientCert.
func TestNewServerConfig_ClientAuthFallback(t *testing.T) {
	t.Parallel()

	subject := testSubjectHelper(t)

	tests := []struct {
		name       string
		clientAuth tls.ClientAuthType
		expectCAs  bool // Whether ClientCAs should be populated from rootCAsPool.
	}{
		{
			name:       "VerifyClientCertIfGiven with nil ClientCAs uses rootCAsPool",
			clientAuth: tls.VerifyClientCertIfGiven,
			expectCAs:  true,
		},
		{
			name:       "RequireAnyClientCert with nil ClientCAs uses rootCAsPool",
			clientAuth: tls.RequireAnyClientCert,
			expectCAs:  true,
		},
		{
			name:       "NoClientCert with nil ClientCAs stays nil",
			clientAuth: tls.NoClientCert,
			expectCAs:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config, err := NewServerConfig(&ServerConfigOptions{
				Subject:    subject,
				ClientAuth: tc.clientAuth,
				ClientCAs:  nil, // Force fallback logic.
			})
			require.NoError(t, err)
			require.NotNil(t, config)

			if tc.expectCAs {
				require.NotNil(t, config.TLSConfig.ClientCAs, "ClientCAs should be populated via fallback for %s", tc.name)
			} else {
				require.Nil(t, config.TLSConfig.ClientCAs, "ClientCAs should remain nil for %s", tc.name)
			}
		})
	}
}
