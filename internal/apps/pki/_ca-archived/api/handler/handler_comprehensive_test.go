// Copyright (c) 2025 Justin Cranford

package handler

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilApiCaServer "cryptoutil/api/ca/server"
	cryptoutilCABootstrap "cryptoutil/internal/apps/pki/ca/bootstrap"
	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
	cryptoutilCAIntermediate "cryptoutil/internal/apps/pki/ca/intermediate"
	cryptoutilCAProfileCertificate "cryptoutil/internal/apps/pki/ca/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/apps/pki/ca/profile/subject"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/pki/ca/service/issuer"
)

// testIssuerSetup contains a test issuer and related configuration.
type testIssuerSetup struct {
	Issuer   *cryptoutilCAServiceIssuer.Issuer
	Provider cryptoutilCACrypto.Provider
}

// createTestIssuer creates a real issuer for testing.
func createTestIssuer(t *testing.T) *testIssuerSetup {
	t.Helper()

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	// Create root CA.
	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootConfig := &cryptoutilCABootstrap.RootCAConfig{
		Name: "Test Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  cryptoutilSharedMagic.MaxErrorDisplay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour,
		PathLenConstraint: 2,
	}

	rootCA, _, err := bootstrapper.Bootstrap(rootConfig)
	require.NoError(t, err)

	// Create intermediate/issuing CA.
	provisioner := cryptoutilCAIntermediate.NewProvisioner(provider)
	intermediateConfig := &cryptoutilCAIntermediate.IntermediateCAConfig{
		Name: "Test Issuing CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  cryptoutilSharedMagic.JoseJADefaultMaxMaterials * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}

	issuingCA, _, err := provisioner.Provision(intermediateConfig)
	require.NoError(t, err)

	// Create issuer.
	caConfig := &cryptoutilCAServiceIssuer.IssuingCAConfig{
		Name:        "test-ca",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuer, err := cryptoutilCAServiceIssuer.NewIssuer(provider, caConfig)
	require.NoError(t, err)

	return &testIssuerSetup{
		Issuer:   issuer,
		Provider: provider,
	}
}

func TestMapCategory(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	tests := []struct {
		name     string
		input    string
		expected cryptoutilApiCaServer.ProfileSummaryCategory
	}{
		{"tls", "tls", cryptoutilApiCaServer.TLS},
		{cryptoutilSharedMagic.ClaimEmail, cryptoutilSharedMagic.ClaimEmail, cryptoutilApiCaServer.Email},
		{"code_signing", "code_signing", cryptoutilApiCaServer.CodeSigning},
		{"document_signing", "document_signing", cryptoutilApiCaServer.DocumentSigning},
		{"ca", "ca", cryptoutilApiCaServer.CA},
		{"unknown_returns_other", "unknown", cryptoutilApiCaServer.Other},
		{"empty_returns_other", "", cryptoutilApiCaServer.Other},
		{"random_returns_other", "random_category", cryptoutilApiCaServer.Other},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := handler.mapCategory(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestMapKeyUsage(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	tests := []struct {
		name     string
		profile  *cryptoutilCAProfileCertificate.Profile
		expected []string
	}{
		{
			name:     "nil_profile_returns_nil",
			profile:  nil,
			expected: nil,
		},
		{
			name: "digital_signature_only",
			profile: &cryptoutilCAProfileCertificate.Profile{
				KeyUsage: cryptoutilCAProfileCertificate.KeyUsageConfig{
					DigitalSignature: true,
				},
			},
			expected: []string{"digitalSignature"},
		},
		{
			name: "all_key_usages",
			profile: &cryptoutilCAProfileCertificate.Profile{
				KeyUsage: cryptoutilCAProfileCertificate.KeyUsageConfig{
					DigitalSignature:  true,
					ContentCommitment: true,
					KeyEncipherment:   true,
					DataEncipherment:  true,
					KeyAgreement:      true,
					CertSign:          true,
					CRLSign:           true,
				},
			},
			expected: []string{
				"digitalSignature",
				"contentCommitment",
				"keyEncipherment",
				"dataEncipherment",
				"keyAgreement",
				"keyCertSign",
				"cRLSign",
			},
		},
		{
			name: "partial_key_usages",
			profile: &cryptoutilCAProfileCertificate.Profile{
				KeyUsage: cryptoutilCAProfileCertificate.KeyUsageConfig{
					DigitalSignature: true,
					KeyEncipherment:  true,
				},
			},
			expected: []string{"digitalSignature", "keyEncipherment"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := handler.mapKeyUsage(tc.profile)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestMapExtKeyUsage(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	tests := []struct {
		name     string
		profile  *cryptoutilCAProfileCertificate.Profile
		expected []string
	}{
		{
			name:     "nil_profile_returns_nil",
			profile:  nil,
			expected: nil,
		},
		{
			name: "server_auth_only",
			profile: &cryptoutilCAProfileCertificate.Profile{
				ExtendedKeyUsage: cryptoutilCAProfileCertificate.ExtKeyUsageConfig{
					ServerAuth: true,
				},
			},
			expected: []string{"serverAuth"},
		},
		{
			name: "all_ext_key_usages",
			profile: &cryptoutilCAProfileCertificate.Profile{
				ExtendedKeyUsage: cryptoutilCAProfileCertificate.ExtKeyUsageConfig{
					ServerAuth:      true,
					ClientAuth:      true,
					CodeSigning:     true,
					EmailProtection: true,
					TimeStamping:    true,
					OCSPSigning:     true,
				},
			},
			expected: []string{
				"serverAuth",
				"clientAuth",
				"codeSigning",
				"emailProtection",
				"timeStamping",
				"ocspSigning",
			},
		},
		{
			name: "tls_usages",
			profile: &cryptoutilCAProfileCertificate.Profile{
				ExtendedKeyUsage: cryptoutilCAProfileCertificate.ExtKeyUsageConfig{
					ServerAuth: true,
					ClientAuth: true,
				},
			},
			expected: []string{"serverAuth", "clientAuth"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := handler.mapExtKeyUsage(tc.profile)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestMapSubjectRequirements(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	tests := []struct {
		name      string
		profile   *cryptoutilCAProfileSubject.Profile
		expectNil bool
	}{
		{
			name:      "nil_profile_returns_nil",
			profile:   nil,
			expectNil: true,
		},
		{
			name: "basic_requirements",
			profile: &cryptoutilCAProfileSubject.Profile{
				Constraints: cryptoutilCAProfileSubject.Constraints{
					RequireCommonName:   true,
					RequireOrganization: true,
					RequireCountry:      false,
					ValidCountries:      []string{"US", "CA"},
				},
			},
			expectNil: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := handler.mapSubjectRequirements(tc.profile)
			if tc.expectNil {
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)
			}
		})
	}
}

func TestMapSANRequirements(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	tests := []struct {
		name      string
		profile   *cryptoutilCAProfileSubject.Profile
		expectNil bool
	}{
		{
			name:      "nil_profile_returns_nil",
			profile:   nil,
			expectNil: true,
		},
		{
			name: "basic_san_requirements",
			profile: &cryptoutilCAProfileSubject.Profile{
				SubjectAltNames: cryptoutilCAProfileSubject.SANConfig{
					DNSNames: cryptoutilCAProfileSubject.SANPatterns{
						Allowed:  true,
						Required: true,
					},
					IPAddresses: cryptoutilCAProfileSubject.SANPatterns{
						Allowed: true,
					},
					EmailAddresses: cryptoutilCAProfileSubject.SANPatterns{
						Allowed: true,
					},
					URIs: cryptoutilCAProfileSubject.SANPatterns{
						Allowed: true,
					},
				},
			},
			expectNil: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := handler.mapSANRequirements(tc.profile)
			if tc.expectNil {
				require.Nil(t, result)
			} else {
				require.NotNil(t, result)
			}
		})
	}
}

func TestIpsToStrings(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	tests := []struct {
		name     string
		input    []net.IP
		expected []string
	}{
		{
			name:     "empty_slice",
			input:    []net.IP{},
			expected: []string{},
		},
		{
			name:     "single_ipv4",
			input:    []net.IP{net.ParseIP("192.168.1.1")},
			expected: []string{"192.168.1.1"},
		},
		{
			name:     "multiple_ips",
			input:    []net.IP{net.ParseIP("192.168.1.1"), net.ParseIP("10.0.0.1")},
			expected: []string{"192.168.1.1", "10.0.0.1"},
		},
		{
			name:     "ipv6_address",
			input:    []net.IP{net.ParseIP(cryptoutilSharedMagic.IPv6Loopback)},
			expected: []string{cryptoutilSharedMagic.IPv6Loopback},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := handler.ipsToStrings(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestUrisToStrings(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	tests := []struct {
		name     string
		input    []*url.URL
		expected []string
	}{
		{
			name:     "empty_slice",
			input:    []*url.URL{},
			expected: []string{},
		},
		{
			name: "single_uri",
			input: func() []*url.URL {
				u, _ := url.Parse("https://example.com")

				return []*url.URL{u}
			}(),
			expected: []string{"https://example.com"},
		},
		{
			name: "multiple_uris",
			input: func() []*url.URL {
				u1, _ := url.Parse("https://example.com")
				u2, _ := url.Parse("spiffe://trust-domain/workload")

				return []*url.URL{u1, u2}
			}(),
			expected: []string{"https://example.com", "spiffe://trust-domain/workload"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := handler.urisToStrings(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestPtrSlice(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"empty_slice", []string{}, []string{}},
		{"single_element", []string{"a"}, []string{"a"}},
		{"multiple_elements", []string{"a", "b", "c"}, []string{"a", "b", "c"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := handler.ptrSlice(tc.input)
			require.NotNil(t, result)
			require.Equal(t, tc.expected, *result)
		})
	}
}
