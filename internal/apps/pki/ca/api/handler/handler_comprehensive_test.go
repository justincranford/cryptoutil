// Copyright (c) 2025 Justin Cranford

package handler

import (
	"bytes"
	"context"
	"crypto"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	http "net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilApiCaServer "cryptoutil/api/ca/server"
	cryptoutilCABootstrap "cryptoutil/internal/apps/pki/ca/bootstrap"
	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
	cryptoutilCAIntermediate "cryptoutil/internal/apps/pki/ca/intermediate"
	cryptoutilCAProfileCertificate "cryptoutil/internal/apps/pki/ca/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/apps/pki/ca/profile/subject"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/pki/ca/service/issuer"
	cryptoutilCAServiceRevocation "cryptoutil/internal/apps/pki/ca/service/revocation"
	cryptoutilCAServiceTimestamp "cryptoutil/internal/apps/pki/ca/service/timestamp"
	cryptoutilCAStorage "cryptoutil/internal/apps/pki/ca/storage"
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
		ValidityDuration:  20 * 365 * 24 * time.Hour,
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
		ValidityDuration:  10 * 365 * 24 * time.Hour,
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
		{"email", "email", cryptoutilApiCaServer.Email},
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
			input:    []net.IP{net.ParseIP("::1")},
			expected: []string{"::1"},
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

func TestApplySubjectOverrides(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	tests := []struct {
		name           string
		initial        *cryptoutilCAProfileSubject.Request
		override       *cryptoutilApiCaServer.SubjectOverride
		expectedOrg    []string
		expectedOU     []string
		expectedCounty []string
		expectedState  []string
		expectedLocale []string
	}{
		{
			name: "override_organization",
			initial: &cryptoutilCAProfileSubject.Request{
				Organization: []string{"Original"},
			},
			override: &cryptoutilApiCaServer.SubjectOverride{
				Organization: &[]string{"Override Inc."},
			},
			expectedOrg: []string{"Override Inc."},
		},
		{
			name: "override_all_fields",
			initial: &cryptoutilCAProfileSubject.Request{
				Organization:       []string{"Original"},
				OrganizationalUnit: []string{"Original OU"},
				Country:            []string{"US"},
				State:              []string{"California"},
				Locality:           []string{"San Francisco"},
			},
			override: &cryptoutilApiCaServer.SubjectOverride{
				Organization:       &[]string{"New Org"},
				OrganizationalUnit: &[]string{"New OU"},
				Country:            &[]string{"CA"},
				State:              &[]string{"Ontario"},
				Locality:           &[]string{"Toronto"},
			},
			expectedOrg:    []string{"New Org"},
			expectedOU:     []string{"New OU"},
			expectedCounty: []string{"CA"},
			expectedState:  []string{"Ontario"},
			expectedLocale: []string{"Toronto"},
		},
		{
			name: "empty_override_no_change",
			initial: &cryptoutilCAProfileSubject.Request{
				Organization: []string{"Keep Me"},
			},
			override: &cryptoutilApiCaServer.SubjectOverride{
				Organization: &[]string{}, // Empty slice should not override.
			},
			expectedOrg: []string{"Keep Me"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			subjectReq := &cryptoutilCAProfileSubject.Request{
				Organization:       tc.initial.Organization,
				OrganizationalUnit: tc.initial.OrganizationalUnit,
				Country:            tc.initial.Country,
				State:              tc.initial.State,
				Locality:           tc.initial.Locality,
			}

			handler.applySubjectOverrides(subjectReq, tc.override)

			if tc.expectedOrg != nil {
				require.Equal(t, tc.expectedOrg, subjectReq.Organization)
			}

			if tc.expectedOU != nil {
				require.Equal(t, tc.expectedOU, subjectReq.OrganizationalUnit)
			}

			if tc.expectedCounty != nil {
				require.Equal(t, tc.expectedCounty, subjectReq.Country)
			}

			if tc.expectedState != nil {
				require.Equal(t, tc.expectedState, subjectReq.State)
			}

			if tc.expectedLocale != nil {
				require.Equal(t, tc.expectedLocale, subjectReq.Locality)
			}
		})
	}
}

func TestApplySANOverrides(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	tests := []struct {
		name           string
		initial        *cryptoutilCAProfileSubject.Request
		override       *cryptoutilApiCaServer.SANOverride
		expectedDNS    []string
		expectedIPs    []string
		expectedEmails []string
		expectedURIs   []string
	}{
		{
			name: "override_dns_names",
			initial: &cryptoutilCAProfileSubject.Request{
				DNSNames: []string{"original.example.com"},
			},
			override: &cryptoutilApiCaServer.SANOverride{
				DNSNames: &[]string{"override.example.com"},
			},
			expectedDNS: []string{"override.example.com"},
		},
		{
			name: "override_all_sans",
			initial: &cryptoutilCAProfileSubject.Request{
				DNSNames:       []string{"original.example.com"},
				IPAddresses:    []string{"192.168.1.1"},
				EmailAddresses: []string{"original@example.com"},
				URIs:           []string{"https://original.example.com"},
			},
			override: &cryptoutilApiCaServer.SANOverride{
				DNSNames:       &[]string{"new.example.com"},
				IPAddresses:    &[]string{"10.0.0.1"},
				EmailAddresses: &[]string{"new@example.com"},
				Uris:           &[]string{"https://new.example.com"},
			},
			expectedDNS:    []string{"new.example.com"},
			expectedIPs:    []string{"10.0.0.1"},
			expectedEmails: []string{"new@example.com"},
			expectedURIs:   []string{"https://new.example.com"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			subjectReq := &cryptoutilCAProfileSubject.Request{
				DNSNames:       tc.initial.DNSNames,
				IPAddresses:    tc.initial.IPAddresses,
				EmailAddresses: tc.initial.EmailAddresses,
				URIs:           tc.initial.URIs,
			}

			handler.applySANOverrides(subjectReq, tc.override)

			if tc.expectedDNS != nil {
				require.Equal(t, tc.expectedDNS, subjectReq.DNSNames)
			}

			if tc.expectedIPs != nil {
				require.Equal(t, tc.expectedIPs, subjectReq.IPAddresses)
			}

			if tc.expectedEmails != nil {
				require.Equal(t, tc.expectedEmails, subjectReq.EmailAddresses)
			}

			if tc.expectedURIs != nil {
				require.Equal(t, tc.expectedURIs, subjectReq.URIs)
			}
		})
	}
}

func TestParseCSR(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	// Generate a test key.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	// Create a valid CSR.
	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   "test.example.com",
			Organization: []string{"Test Org"},
		},
		DNSNames: []string{"test.example.com", "www.test.example.com"},
	}

	csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, privateKey)
	require.NoError(t, err)

	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDER,
	})

	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid_csr",
			input:   string(csrPEM),
			wantErr: false,
		},
		{
			name:        "invalid_pem",
			input:       "not a valid PEM",
			wantErr:     true,
			errContains: "failed to decode PEM block",
		},
		{
			name: "wrong_pem_type",
			input: string(pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: []byte("not a cert"),
			})),
			wantErr:     true,
			errContains: "expected CERTIFICATE REQUEST",
		},
		{
			name: "invalid_csr_content",
			input: string(pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE REQUEST",
				Bytes: []byte("invalid csr data"),
			})),
			wantErr:     true,
			errContains: "failed to parse CSR",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, err := handler.parseCSR(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

func TestBuildEnrollmentResponse(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	// Generate a test certificate.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "test.example.com",
			Organization: []string{"Test Org"},
			Country:      []string{"US"},
		},
		NotBefore:   time.Now().UTC(),
		NotAfter:    time.Now().UTC().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		DNSNames:    []string{"test.example.com"},
		IPAddresses: []net.IP{net.ParseIP("192.168.1.1")},
	}

	certDER, err := x509.CreateCertificate(crand.Reader, certTemplate, certTemplate, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	issued := &cryptoutilCAServiceIssuer.IssuedCertificate{
		Certificate:    cert,
		CertificatePEM: certPEM,
		ChainPEM:       certPEM,
		SerialNumber:   "01",
		Fingerprint:    "abc123",
	}

	result := handler.buildEnrollmentResponse(issued)
	require.NotNil(t, result)
	require.Equal(t, cryptoutilApiCaServer.Issued, result.Status)
	require.Equal(t, "01", result.Certificate.SerialNumber)
	require.NotNil(t, result.Certificate.FingerprintSha256)
	require.Equal(t, "abc123", *result.Certificate.FingerprintSha256)
}

func TestCertSubjectToAPI(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	cert := &x509.Certificate{
		Subject: pkix.Name{
			CommonName:         "test.example.com",
			Organization:       []string{"Test Org"},
			OrganizationalUnit: []string{"Test OU"},
			Country:            []string{"US"},
			Province:           []string{"California"},
			Locality:           []string{"San Francisco"},
		},
		DNSNames:       []string{"test.example.com", "www.test.example.com"},
		IPAddresses:    []net.IP{net.ParseIP("192.168.1.1")},
		EmailAddresses: []string{"admin@example.com"},
	}

	result := handler.certSubjectToAPI(cert)
	require.NotNil(t, result.CommonName)
	require.Equal(t, "test.example.com", *result.CommonName)
	require.NotNil(t, result.Organization)
	require.Equal(t, []string{"Test Org"}, *result.Organization)
	require.NotNil(t, result.DNSNames)
	require.Equal(t, []string{"test.example.com", "www.test.example.com"}, *result.DNSNames)
}

func TestBuildProfileResponse(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	profile := &ProfileConfig{
		ID:          "tls-server",
		Name:        "TLS Server Certificate",
		Description: "A profile for TLS server certificates",
		Category:    "tls",
		CertificateProfile: &cryptoutilCAProfileCertificate.Profile{
			KeyUsage: cryptoutilCAProfileCertificate.KeyUsageConfig{
				DigitalSignature: true,
				KeyEncipherment:  true,
			},
			ExtendedKeyUsage: cryptoutilCAProfileCertificate.ExtKeyUsageConfig{
				ServerAuth: true,
			},
			Validity: cryptoutilCAProfileCertificate.ValidityConfig{
				MaxDuration: "8760h",
			},
		},
		SubjectProfile: &cryptoutilCAProfileSubject.Profile{
			Constraints: cryptoutilCAProfileSubject.Constraints{
				RequireCommonName: true,
			},
		},
	}

	result := handler.buildProfileResponse(profile)
	require.NotNil(t, result)
	require.Equal(t, "tls-server", result.ID)
	require.Equal(t, "TLS Server Certificate", result.Name)
	require.NotNil(t, result.Description)
	require.Equal(t, "A profile for TLS server certificates", *result.Description)
	require.NotNil(t, result.MaxValidityDays)
	require.Equal(t, 365, *result.MaxValidityDays)
}

func TestOcspErrorResponse(t *testing.T) {
	t.Parallel()

	// Create a Fiber app for testing.
	app := fiber.New()

	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
	}

	app.Get("/ocsp-error", func(c *fiber.Ctx) error {
		return handler.ocspErrorResponse(c, fiber.StatusBadRequest)
	})

	req := httptest.NewRequest(http.MethodGet, "/ocsp-error", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	require.Equal(t, "application/ocsp-response", resp.Header.Get("Content-Type"))

	require.NoError(t, resp.Body.Close())
}

func TestLookupCertificateBySerialNilSerial(t *testing.T) {
	t.Parallel()

	mockStorage := cryptoutilCAStorage.NewMemoryStore()
	handler := &Handler{
		storage: mockStorage,
	}

	// Test with nil serial number.
	result := handler.lookupCertificateBySerial(context.Background(), nil)
	require.Nil(t, result)
}

func TestLookupCertificateBySerialNotFound(t *testing.T) {
	t.Parallel()

	mockStorage := cryptoutilCAStorage.NewMemoryStore()
	handler := &Handler{
		storage: mockStorage,
	}

	// Test with serial number that doesn't exist.
	serial := big.NewInt(12345)
	result := handler.lookupCertificateBySerial(context.Background(), serial)
	require.Nil(t, result)
}

func TestHandlerSetServices(t *testing.T) {
	t.Parallel()

	mockStorage := cryptoutilCAStorage.NewMemoryStore()
	handler := &Handler{
		storage: mockStorage,
	}

	// Test SetTSAService.
	handler.SetTSAService(nil)

	// Test SetOCSPService.
	handler.SetOCSPService(nil)

	// Test SetCRLService.
	handler.SetCRLService(nil)

	// Verify services can be set (no panic).
	require.NotNil(t, handler)
}

func TestErrorResponseHandler(t *testing.T) {
	t.Parallel()

	// Create a Fiber app for testing.
	app := fiber.New()

	mockStorage := cryptoutilCAStorage.NewMemoryStore()
	handler := &Handler{
		storage: mockStorage,
	}

	app.Get("/error-test", func(c *fiber.Ctx) error {
		return handler.errorResponse(c, fiber.StatusBadRequest, "test_error", "This is a test error message")
	})

	req := httptest.NewRequest(http.MethodGet, "/error-test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.NoError(t, resp.Body.Close())

	require.Contains(t, string(body), "test_error")
	require.Contains(t, string(body), "This is a test error message")
}

func TestEstEndpoints(t *testing.T) {
	t.Parallel()

	// Create a Fiber app for testing EST endpoints.
	app := fiber.New()

	mockStorage := cryptoutilCAStorage.NewMemoryStore()
	handler := &Handler{
		storage: mockStorage,
	}

	app.Get("/est/csrattrs", func(c *fiber.Ctx) error {
		return handler.EstCSRAttrs(c)
	})

	app.Post("/est/simpleenroll", func(c *fiber.Ctx) error {
		return handler.EstSimpleEnroll(c)
	})

	app.Post("/est/simplereenroll", func(c *fiber.Ctx) error {
		return handler.EstSimpleReenroll(c)
	})

	app.Post("/est/serverkeygen", func(c *fiber.Ctx) error {
		return handler.EstServerKeyGen(c)
	})

	t.Run("EST_CSRAttrs_returns_no_content", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/est/csrattrs", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNoContent, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("EST_SimpleEnroll_empty_body", func(t *testing.T) {
		t.Parallel()

		// With empty body, should return bad request (endpoint is implemented).
		req := httptest.NewRequest(http.MethodPost, "/est/simpleenroll", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("EST_SimpleReenroll_empty_body", func(t *testing.T) {
		t.Parallel()

		// With empty body, should return bad request (endpoint is implemented).
		req := httptest.NewRequest(http.MethodPost, "/est/simplereenroll", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("EST_ServerKeyGen_empty_body", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/est/serverkeygen", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestTsaTimestamp(t *testing.T) {
	t.Parallel()

	// Create a Fiber app for testing.
	app := fiber.New()

	mockStorage := cryptoutilCAStorage.NewMemoryStore()
	handler := &Handler{
		storage: mockStorage,
	}

	app.Post("/tsa/timestamp", func(c *fiber.Ctx) error {
		return handler.TsaTimestamp(c)
	})

	t.Run("TSA_no_service_configured", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/tsa/timestamp", bytes.NewReader([]byte("timestamp request")))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestHandleOCSP(t *testing.T) {
	t.Parallel()

	// Create a Fiber app for testing.
	app := fiber.New()

	mockStorage := cryptoutilCAStorage.NewMemoryStore()
	handler := &Handler{
		storage: mockStorage,
	}

	app.Post("/ocsp", func(c *fiber.Ctx) error {
		return handler.HandleOCSP(c)
	})

	t.Run("OCSP_no_service_configured", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/ocsp", bytes.NewReader([]byte("ocsp request")))
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestGetCRL(t *testing.T) {
	t.Parallel()

	// Create a Fiber app for testing.
	app := fiber.New()

	mockStorage := cryptoutilCAStorage.NewMemoryStore()
	handler := &Handler{
		storage: mockStorage,
	}

	app.Get("/ca/:caId/crl", func(c *fiber.Ctx) error {
		return handler.GetCRL(c, c.Params("caId"), cryptoutilApiCaServer.GetCRLParams{})
	})

	t.Run("CRL_no_service_configured", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ca/test-ca/crl", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestListProfiles(t *testing.T) {
	t.Parallel()

	// Create a Fiber app for testing.
	app := fiber.New()

	profiles := map[string]*ProfileConfig{
		"tls-server": {
			ID:          "tls-server",
			Name:        "TLS Server",
			Description: "TLS Server Certificate Profile",
			Category:    "tls",
		},
		"code-signing": {
			ID:          "code-signing",
			Name:        "Code Signing",
			Description: "Code Signing Certificate Profile",
			Category:    "code_signing",
		},
	}

	handler := &Handler{
		profiles: profiles,
	}

	app.Get("/profiles", func(c *fiber.Ctx) error {
		return handler.ListProfiles(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/profiles", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.NoError(t, resp.Body.Close())

	require.Contains(t, string(body), "tls-server")
	require.Contains(t, string(body), "code-signing")
}

func TestGetProfile(t *testing.T) {
	t.Parallel()

	// Create a Fiber app for testing.
	app := fiber.New()

	profiles := map[string]*ProfileConfig{
		"tls-server": {
			ID:          "tls-server",
			Name:        "TLS Server",
			Description: "TLS Server Certificate Profile",
			Category:    "tls",
		},
	}

	handler := &Handler{
		profiles: profiles,
	}

	app.Get("/profiles/:profileId", func(c *fiber.Ctx) error {
		return handler.GetProfile(c, c.Params("profileId"))
	})

	t.Run("GetProfile_exists", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/profiles/tls-server", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.NoError(t, resp.Body.Close())

		require.Contains(t, string(body), "tls-server")
	})

	t.Run("GetProfile_not_found", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/profiles/unknown-profile", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestGetKeyInfoWithRSA4096(t *testing.T) {
	t.Parallel()

	// Test RSA 4096 key.
	key, err := rsa.GenerateKey(crand.Reader, 4096)
	require.NoError(t, err)

	cert := &x509.Certificate{PublicKey: &key.PublicKey}
	algo, size := getKeyInfo(cert)

	require.Equal(t, "RSA", algo)
	require.Equal(t, 4096, size)
}

func TestEnrollmentTracker(t *testing.T) {
	t.Parallel()

	t.Run("NewEnrollmentTracker", func(t *testing.T) {
		t.Parallel()

		tracker := newEnrollmentTracker(100)
		require.NotNil(t, tracker)
		require.NotNil(t, tracker.requests)
		require.Equal(t, 100, tracker.maxEntries)
	})

	t.Run("TrackAndGet", func(t *testing.T) {
		t.Parallel()

		tracker := newEnrollmentTracker(100)

		requestID := googleUuid.New()
		status := cryptoutilApiCaServer.EnrollmentStatusResponseStatusIssued
		serialNumber := "ABC123"

		tracker.track(requestID, status, serialNumber)

		entry, found := tracker.get(requestID)
		require.True(t, found)
		require.Equal(t, requestID, entry.RequestID)
		require.Equal(t, status, entry.Status)
		require.Equal(t, serialNumber, entry.SerialNumber)
	})

	t.Run("GetNotFound", func(t *testing.T) {
		t.Parallel()

		tracker := newEnrollmentTracker(100)

		requestID := googleUuid.New()
		entry, found := tracker.get(requestID)
		require.False(t, found)
		require.Nil(t, entry)
	})

	t.Run("MaxEntriesEnforced", func(t *testing.T) {
		t.Parallel()

		maxEntries := 3
		tracker := newEnrollmentTracker(maxEntries)

		// Add maxEntries items.
		ids := make([]googleUuid.UUID, maxEntries+1)
		for i := 0; i < maxEntries; i++ {
			ids[i] = googleUuid.New()
			tracker.track(ids[i], cryptoutilApiCaServer.EnrollmentStatusResponseStatusIssued, "serial")
			time.Sleep(time.Millisecond) // Ensure different timestamps.
		}

		// Add one more (should evict oldest).
		ids[maxEntries] = googleUuid.New()
		tracker.track(ids[maxEntries], cryptoutilApiCaServer.EnrollmentStatusResponseStatusIssued, "serial")

		// First entry should be evicted.
		_, found := tracker.get(ids[0])
		require.False(t, found, "oldest entry should be evicted")

		// Last entry should exist.
		_, found = tracker.get(ids[maxEntries])
		require.True(t, found, "newest entry should exist")
	})
}

func TestParseESTCSR(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	t.Run("ValidPEMCSR", func(t *testing.T) {
		t.Parallel()

		// Generate a test CSR.
		key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
		require.NoError(t, err)

		template := &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName: "test.example.com",
			},
		}
		csrDER, err := x509.CreateCertificateRequest(crand.Reader, template, key)
		require.NoError(t, err)

		csrPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE REQUEST",
			Bytes: csrDER,
		})

		csr, err := handler.parseESTCSR(csrPEM)
		require.NoError(t, err)
		require.NotNil(t, csr)
		require.Equal(t, "test.example.com", csr.Subject.CommonName)
	})

	t.Run("ValidDERCSR", func(t *testing.T) {
		t.Parallel()

		// Generate a test CSR.
		key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
		require.NoError(t, err)

		template := &x509.CertificateRequest{
			Subject: pkix.Name{
				CommonName: "test.example.com",
			},
		}
		csrDER, err := x509.CreateCertificateRequest(crand.Reader, template, key)
		require.NoError(t, err)

		csr, err := handler.parseESTCSR(csrDER)
		require.NoError(t, err)
		require.NotNil(t, csr)
	})

	t.Run("InvalidCSR", func(t *testing.T) {
		t.Parallel()

		csr, err := handler.parseESTCSR([]byte("invalid data"))
		require.Error(t, err)
		require.Nil(t, csr)
		require.Contains(t, err.Error(), "invalid format")
	})
}

func TestBuildESTIssueRequest(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	// Generate a test CSR.
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   "test.example.com",
			Organization: []string{"Test Org"},
		},
		DNSNames: []string{"test.example.com", "www.test.example.com"},
	}
	csrDER, err := x509.CreateCertificateRequest(crand.Reader, template, key)
	require.NoError(t, err)

	csr, err := x509.ParseCertificateRequest(csrDER)
	require.NoError(t, err)

	profile := &ProfileConfig{
		ID:   "test-profile",
		Name: "Test Profile",
	}

	issueReq := handler.buildESTIssueRequest(csr, profile)
	require.NotNil(t, issueReq)
	require.NotNil(t, issueReq.SubjectRequest)
	require.Equal(t, "test.example.com", issueReq.SubjectRequest.CommonName)
	require.Equal(t, []string{"Test Org"}, issueReq.SubjectRequest.Organization)
	require.Equal(t, []string{"test.example.com", "www.test.example.com"}, issueReq.SubjectRequest.DNSNames)
}

func TestListCertificates(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	// Add test certificates to storage.
	cert1 := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		SerialNumber:   "ABC123",
		SubjectDN:      "CN=test1.example.com,O=Test Org",
		IssuerDN:       "CN=Test CA,O=Test Org",
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * 24 * 365),
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      "tls-server",
		CertificatePEM: "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----",
	}
	err := mockStorage.Store(context.Background(), cert1)
	require.NoError(t, err)

	cert2 := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		SerialNumber:   "DEF456",
		SubjectDN:      "CN=test2.example.com,O=Test Org",
		IssuerDN:       "CN=Test CA,O=Test Org",
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * 24 * 365),
		Status:         cryptoutilCAStorage.StatusRevoked,
		ProfileID:      "tls-server",
		CertificatePEM: "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----",
	}
	err = mockStorage.Store(context.Background(), cert2)
	require.NoError(t, err)

	handler := &Handler{storage: mockStorage}

	app.Get("/certificates", func(c *fiber.Ctx) error {
		params := cryptoutilApiCaServer.ListCertificatesParams{}

		return handler.ListCertificates(c, params)
	})

	t.Run("ListAll", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/certificates", nil)
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestGetCertificate(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	testCert := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		SerialNumber:   "SERIAL123",
		SubjectDN:      "CN=test.example.com,O=Test Org",
		IssuerDN:       "CN=Test CA,O=Test Org",
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * 24 * 365),
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      "tls-server",
		CertificatePEM: "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
	}
	err := mockStorage.Store(context.Background(), testCert)
	require.NoError(t, err)

	handler := &Handler{storage: mockStorage}

	app.Get("/certificates/:serialNumber", func(c *fiber.Ctx) error {
		return handler.GetCertificate(c, c.Params("serialNumber"))
	})

	t.Run("Found", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/certificates/SERIAL123", nil)
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/certificates/NONEXISTENT", nil)
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("EmptySerial", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/certificates/", nil)
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)

		require.NoError(t, resp.Body.Close())
	})
}

func TestGetCertificateChain(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	testCert := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		SerialNumber:   "CHAIN123",
		SubjectDN:      "CN=test.example.com",
		IssuerDN:       "CN=Test CA",
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * 24 * 365),
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      "tls-server",
		CertificatePEM: "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
	}
	err := mockStorage.Store(context.Background(), testCert)
	require.NoError(t, err)

	handler := &Handler{storage: mockStorage}

	app.Get("/certificates/:serialNumber/chain", func(c *fiber.Ctx) error {
		return handler.GetCertificateChain(c, c.Params("serialNumber"))
	})

	t.Run("Found", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/certificates/CHAIN123/chain", nil)
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/certificates/NONEXISTENT/chain", nil)
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestRevokeCertificate(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	testCert := &cryptoutilCAStorage.StoredCertificate{
		ID:             googleUuid.New(),
		SerialNumber:   "REVOKE123",
		SubjectDN:      "CN=test.example.com",
		IssuerDN:       "CN=Test CA",
		NotBefore:      time.Now().UTC().Add(-time.Hour),
		NotAfter:       time.Now().UTC().Add(time.Hour * 24 * 365),
		Status:         cryptoutilCAStorage.StatusActive,
		ProfileID:      "tls-server",
		CertificatePEM: "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----",
	}
	err := mockStorage.Store(context.Background(), testCert)
	require.NoError(t, err)

	handler := &Handler{storage: mockStorage}

	app.Post("/certificates/:serialNumber/revoke", func(c *fiber.Ctx) error {
		return handler.RevokeCertificate(c, c.Params("serialNumber"))
	})

	revokeBody := `{"reason": "key_compromise"}`

	t.Run("RevokeSuccess", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/certificates/REVOKE123/revoke", bytes.NewBufferString(revokeBody))
		req.Header.Set("Content-Type", "application/json")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/certificates/NONEXISTENT/revoke", bytes.NewBufferString(revokeBody))
		req.Header.Set("Content-Type", "application/json")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestGetEnrollmentStatusHandler(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	tracker := newEnrollmentTracker(100)
	requestID := googleUuid.New()
	tracker.track(requestID, cryptoutilApiCaServer.EnrollmentStatusResponseStatusIssued, "ISSUED123")

	handler := &Handler{
		storage:           mockStorage,
		enrollmentTracker: tracker,
	}

	app.Get("/enroll/:requestId", func(c *fiber.Ctx) error {
		idStr := c.Params("requestId")

		id, parseErr := googleUuid.Parse(idStr)
		if parseErr != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request ID"})
		}

		return handler.GetEnrollmentStatus(c, id)
	})

	t.Run("Found", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/enroll/"+requestID.String(), nil)
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		unknownID := googleUuid.New()
		req := httptest.NewRequest(http.MethodGet, "/enroll/"+unknownID.String(), nil)
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		err := resp.Body.Close()
		require.NoError(t, err)
	})
}

func TestListCAsNilIssuer(t *testing.T) {
	t.Parallel()

	// Tests the error path when issuer is nil.
	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  nil,
	}

	app.Get("/ca", func(c *fiber.Ctx) error {
		return handler.ListCAs(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/ca", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	// When issuer is nil, GetCAConfig returns nil, causing an internal server error.
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestGetCANilIssuer(t *testing.T) {
	t.Parallel()

	// Tests the error path when issuer is nil.
	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  nil,
	}

	app.Get("/ca/:caId", func(c *fiber.Ctx) error {
		return handler.GetCA(c, c.Params("caId"))
	})

	req := httptest.NewRequest(http.MethodGet, "/ca/test-ca", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	// When issuer is nil, GetCAConfig returns nil, causing an internal server error.
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestEstCACertsNilIssuer(t *testing.T) {
	t.Parallel()

	// Tests the error path when issuer is nil.
	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  nil,
	}

	app.Get("/est/cacerts", func(c *fiber.Ctx) error {
		return handler.EstCACerts(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/est/cacerts", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	// When issuer is nil, GetCAConfig returns nil, causing an internal server error.
	require.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// Tests with real issuer follow.

func TestListCAsWithRealIssuer(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  testSetup.Issuer,
	}

	app.Get("/ca", func(c *fiber.Ctx) error {
		return handler.ListCAs(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/ca", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "test-ca")

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestGetCAWithRealIssuer(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  testSetup.Issuer,
	}

	app.Get("/ca/:caId", func(c *fiber.Ctx) error {
		return handler.GetCA(c, c.Params("caId"))
	})

	t.Run("Found", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ca/test-ca", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), "test-ca")

		require.NoError(t, resp.Body.Close())
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ca/unknown-ca", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestEstCACertsWithRealIssuer(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  testSetup.Issuer,
	}

	app.Get("/est/cacerts", func(c *fiber.Ctx) error {
		return handler.EstCACerts(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/est/cacerts", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "-----BEGIN CERTIFICATE-----")

	err = resp.Body.Close()
	require.NoError(t, err)
}

func TestSubmitEnrollmentWithRealIssuer(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	profiles := map[string]*ProfileConfig{
		"tls-server": {
			ID:          "tls-server",
			Name:        "TLS Server",
			Description: "TLS Server Certificate Profile",
			Category:    "tls",
		},
	}

	handler := &Handler{
		storage:           mockStorage,
		issuer:            testSetup.Issuer,
		profiles:          profiles,
		enrollmentTracker: newEnrollmentTracker(100),
	}

	app.Post("/enrollments", func(c *fiber.Ctx) error {
		return handler.SubmitEnrollment(c)
	})

	// Generate a test CSR.
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   "test.example.com",
			Organization: []string{"Test Org"},
		},
		DNSNames: []string{"test.example.com"},
	}
	csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, key)
	require.NoError(t, err)

	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrDER,
	})

	t.Run("SuccessfulEnrollment", func(t *testing.T) {
		t.Parallel()

		reqBody := `{"csr": "` + string(csrPEM) + `", "profile": "tls-server"}`
		// Escape newlines for JSON.
		reqBodyEscaped := bytes.ReplaceAll([]byte(reqBody), []byte("\n"), []byte("\\n"))

		req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewReader(reqBodyEscaped))
		req.Header.Set("Content-Type", "application/json")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusCreated, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("MissingCSR", func(t *testing.T) {
		t.Parallel()

		reqBody := `{"profile": "tls-server"}`
		req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("MissingProfile", func(t *testing.T) {
		t.Parallel()

		reqBody := `{"csr": "test-csr"}`
		req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("UnknownProfile", func(t *testing.T) {
		t.Parallel()

		reqBody := `{"csr": "` + string(csrPEM) + `", "profile": "unknown-profile"}`
		reqBody = string(bytes.ReplaceAll([]byte(reqBody), []byte("\n"), []byte("\\n")))

		req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("InvalidCSR", func(t *testing.T) {
		t.Parallel()

		reqBody := `{"csr": "invalid-csr-data", "profile": "tls-server"}`
		req := httptest.NewRequest(http.MethodPost, "/enrollments", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestEstSimpleEnrollWithRealIssuer(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	profiles := map[string]*ProfileConfig{
		"tls-server": {
			ID:          "tls-server",
			Name:        "TLS Server",
			Description: "TLS Server Certificate Profile",
			Category:    "tls",
		},
	}

	handler := &Handler{
		storage:           mockStorage,
		issuer:            testSetup.Issuer,
		profiles:          profiles,
		enrollmentTracker: newEnrollmentTracker(100),
	}

	app.Post("/est/simpleenroll", func(c *fiber.Ctx) error {
		return handler.EstSimpleEnroll(c)
	})

	// Generate a test CSR.
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   "test.example.com",
			Organization: []string{"Test Org"},
		},
		DNSNames: []string{"test.example.com"},
	}
	csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, key)
	require.NoError(t, err)

	t.Run("SuccessfulEnrollmentWithDER", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/est/simpleenroll", bytes.NewReader(csrDER))
		req.Header.Set("Content-Type", "application/pkcs10")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)
		require.Contains(t, string(body), "-----BEGIN CERTIFICATE-----")

		require.NoError(t, resp.Body.Close())
	})

	t.Run("SuccessfulEnrollmentWithBase64", func(t *testing.T) {
		t.Parallel()

		csrBase64 := base64.StdEncoding.EncodeToString(csrDER)
		req := httptest.NewRequest(http.MethodPost, "/est/simpleenroll", bytes.NewBufferString(csrBase64))
		req.Header.Set("Content-Type", "application/pkcs10")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)
		require.Contains(t, string(body), "-----BEGIN CERTIFICATE-----")

		require.NoError(t, resp.Body.Close())
	})

	t.Run("SuccessfulEnrollmentWithPEM", func(t *testing.T) {
		t.Parallel()

		csrPEM := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE REQUEST",
			Bytes: csrDER,
		})
		req := httptest.NewRequest(http.MethodPost, "/est/simpleenroll", bytes.NewReader(csrPEM))
		req.Header.Set("Content-Type", "application/pkcs10")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)
		require.Contains(t, string(body), "-----BEGIN CERTIFICATE-----")

		require.NoError(t, resp.Body.Close())
	})

	t.Run("InvalidCSR", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/est/simpleenroll", bytes.NewBufferString("invalid-csr-data"))
		req.Header.Set("Content-Type", "application/pkcs10")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

func TestEstSimpleEnrollNoProfile(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	// No profiles configured.
	handler := &Handler{
		storage:           mockStorage,
		issuer:            testSetup.Issuer,
		profiles:          map[string]*ProfileConfig{},
		enrollmentTracker: newEnrollmentTracker(100),
	}

	app.Post("/est/simpleenroll", func(c *fiber.Ctx) error {
		return handler.EstSimpleEnroll(c)
	})

	// Generate a test CSR.
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: "test.example.com",
		},
	}
	csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, key)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/est/simpleenroll", bytes.NewReader(csrDER))
	req.Header.Set("Content-Type", "application/pkcs10")
	resp, testErr := app.Test(req)
	require.NoError(t, testErr)
	require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// TestEstServerKeyGenWithRealIssuer tests the EST serverkeygen endpoint with a real issuer.
func TestEstServerKeyGenWithRealIssuer(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	app := fiber.New()
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	profiles := map[string]*ProfileConfig{
		"tls-server": {
			ID:          "tls-server",
			Name:        "TLS Server",
			Description: "TLS Server Certificate Profile",
			Category:    "tls",
		},
	}

	handler := &Handler{
		storage:           mockStorage,
		issuer:            testSetup.Issuer,
		profiles:          profiles,
		enrollmentTracker: newEnrollmentTracker(100),
	}

	app.Post("/est/serverkeygen", func(c *fiber.Ctx) error {
		return handler.EstServerKeyGen(c)
	})

	// Generate a test CSR template (server will replace the key).
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   "serverkeygen.example.com",
			Organization: []string{"Test Org"},
		},
		DNSNames: []string{"serverkeygen.example.com"},
	}
	csrDER, err := x509.CreateCertificateRequest(crand.Reader, csrTemplate, key)
	require.NoError(t, err)

	t.Run("SuccessfulServerKeyGenWithDER", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/est/serverkeygen", bytes.NewReader(csrDER))
		req.Header.Set("Content-Type", "application/pkcs10")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)
		require.NotEmpty(t, body)

		// Response should be PKCS#7 format containing certificate and key.
		require.Equal(t, "application/pkcs7-mime; smime-type=server-generated-key", resp.Header.Get("Content-Type"))

		require.NoError(t, resp.Body.Close())
	})

	t.Run("SuccessfulServerKeyGenWithBase64", func(t *testing.T) {
		t.Parallel()

		csrBase64 := base64.StdEncoding.EncodeToString(csrDER)
		req := httptest.NewRequest(http.MethodPost, "/est/serverkeygen", bytes.NewBufferString(csrBase64))
		req.Header.Set("Content-Type", "application/pkcs10")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)

		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)
		require.NotEmpty(t, body)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("InvalidCSRFormat", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/est/serverkeygen", bytes.NewBufferString("invalid-csr-data"))
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

// TestTsaTimestampWithService tests the TSA endpoint with a real TSA service.
func TestTsaTimestampWithService(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	// Type assert private key to crypto.Signer.
	signer, ok := testSetup.Issuer.GetCAConfig().PrivateKey.(crypto.Signer)
	require.True(t, ok, "private key must implement crypto.Signer")

	// Create TSA service.
	tsaConfig := &cryptoutilCAServiceTimestamp.TSAConfig{
		Certificate:        testSetup.Issuer.GetCAConfig().Certificate,
		PrivateKey:         signer,
		Provider:           testSetup.Provider,
		Policy:             []int{1, 3, 6, 1, 4, 1, 99999, 1},
		AcceptedAlgorithms: []cryptoutilCAServiceTimestamp.HashAlgorithm{cryptoutilCAServiceTimestamp.HashAlgorithmSHA256},
	}
	tsaService, err := cryptoutilCAServiceTimestamp.NewTSAService(tsaConfig)
	require.NoError(t, err)

	handler := &Handler{
		storage:    cryptoutilCAStorage.NewMemoryStore(),
		issuer:     testSetup.Issuer,
		tsaService: tsaService,
	}

	app := fiber.New()
	app.Post("/tsa/timestamp", func(c *fiber.Ctx) error {
		return handler.TsaTimestamp(c)
	})

	t.Run("EmptyRequest", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/tsa/timestamp", bytes.NewReader([]byte{}))
		req.Header.Set("Content-Type", "application/timestamp-query")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})

	t.Run("InvalidDERRequest", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/tsa/timestamp", bytes.NewReader([]byte("not-valid-der")))
		req.Header.Set("Content-Type", "application/timestamp-query")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		require.NoError(t, resp.Body.Close())
	})
}

// TestGetCRLWithService tests CRL generation with a real CRL service.
func TestGetCRLWithService(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	// Type assert private key to crypto.Signer.
	signer, ok := testSetup.Issuer.GetCAConfig().PrivateKey.(crypto.Signer)
	require.True(t, ok, "private key must implement crypto.Signer")

	// Create CRL service.
	crlConfig := &cryptoutilCAServiceRevocation.CRLConfig{
		Issuer:           testSetup.Issuer.GetCAConfig().Certificate,
		PrivateKey:       signer,
		Provider:         testSetup.Provider,
		Validity:         24 * time.Hour,
		NextUpdateBuffer: time.Hour,
	}
	crlService, err := cryptoutilCAServiceRevocation.NewCRLService(crlConfig)
	require.NoError(t, err)

	handler := &Handler{
		storage:    cryptoutilCAStorage.NewMemoryStore(),
		issuer:     testSetup.Issuer,
		crlService: crlService,
	}

	app := fiber.New()
	app.Get("/ca/:caId/crl", func(c *fiber.Ctx) error {
		caID := c.Params("caId")
		formatParam := c.Query("format", "der")
		format := cryptoutilApiCaServer.GetCRLParamsFormat(formatParam)

		return handler.GetCRL(c, caID, cryptoutilApiCaServer.GetCRLParams{Format: &format})
	})

	t.Run("CANotFound", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ca/wrong-ca/crl", nil)
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		closeErr := resp.Body.Close()
		require.NoError(t, closeErr)
	})

	t.Run("DERFormat", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ca/test-ca/crl?format=der", nil)
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		require.Equal(t, "application/pkix-crl", resp.Header.Get("Content-Type"))
		require.Contains(t, resp.Header.Get("Content-Disposition"), ".crl")

		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)
		require.NotEmpty(t, body)

		closeErr := resp.Body.Close()
		require.NoError(t, closeErr)
	})

	t.Run("PEMFormat", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ca/test-ca/crl?format=pem", nil)
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusOK, resp.StatusCode)
		require.Equal(t, "application/x-pem-file", resp.Header.Get("Content-Type"))
		require.Contains(t, resp.Header.Get("Content-Disposition"), ".crl.pem")

		body, readErr := io.ReadAll(resp.Body)
		require.NoError(t, readErr)
		require.Contains(t, string(body), "-----BEGIN X509 CRL-----")

		closeErr := resp.Body.Close()
		require.NoError(t, closeErr)
	})
}

// TestHandleOCSPWithService tests OCSP handling with a real OCSP service.
func TestHandleOCSPWithService(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	// Type assert private key to crypto.Signer.
	signer, ok := testSetup.Issuer.GetCAConfig().PrivateKey.(crypto.Signer)
	require.True(t, ok, "private key must implement crypto.Signer")

	// Create CRL service.
	crlConfig := &cryptoutilCAServiceRevocation.CRLConfig{
		Issuer:           testSetup.Issuer.GetCAConfig().Certificate,
		PrivateKey:       signer,
		Provider:         testSetup.Provider,
		Validity:         24 * time.Hour,
		NextUpdateBuffer: time.Hour,
	}
	crlService, err := cryptoutilCAServiceRevocation.NewCRLService(crlConfig)
	require.NoError(t, err)

	// Use the issuing CA cert as OCSP responder (self-signed responder).
	ocspConfig := &cryptoutilCAServiceRevocation.OCSPConfig{
		Issuer:       testSetup.Issuer.GetCAConfig().Certificate,
		Responder:    testSetup.Issuer.GetCAConfig().Certificate,
		ResponderKey: signer,
		Provider:     testSetup.Provider,
		Validity:     time.Hour,
	}
	ocspService, err := cryptoutilCAServiceRevocation.NewOCSPService(ocspConfig, crlService)
	require.NoError(t, err)

	handler := &Handler{
		storage:     mockStorage,
		issuer:      testSetup.Issuer,
		ocspService: ocspService,
	}

	app := fiber.New()
	app.Post("/ocsp", func(c *fiber.Ctx) error {
		return handler.HandleOCSP(c)
	})

	t.Run("EmptyRequest", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/ocsp", bytes.NewReader([]byte{}))
		req.Header.Set("Content-Type", "application/ocsp-request")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		require.Equal(t, "application/ocsp-response", resp.Header.Get("Content-Type"))

		require.NoError(t, resp.Body.Close())
	})

	t.Run("InvalidOCSPRequest", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/ocsp", bytes.NewReader([]byte("not-valid-ocsp")))
		req.Header.Set("Content-Type", "application/ocsp-request")
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		// Invalid request returns error status with OCSP content type.
		require.Equal(t, "application/ocsp-response", resp.Header.Get("Content-Type"))

		require.NoError(t, resp.Body.Close())
	})
}

// TestNewHandlerValidation tests NewHandler validation paths.
func TestNewHandlerValidation(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	tests := []struct {
		name      string
		issuer    *cryptoutilCAServiceIssuer.Issuer
		storage   cryptoutilCAStorage.Store
		profiles  map[string]*ProfileConfig
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "NilStorage",
			issuer:    testSetup.Issuer,
			storage:   nil,
			profiles:  map[string]*ProfileConfig{},
			wantErr:   true,
			errSubstr: "storage",
		},
		{
			name:      "NilIssuer",
			issuer:    nil,
			storage:   mockStorage,
			profiles:  map[string]*ProfileConfig{},
			wantErr:   true,
			errSubstr: "issuer",
		},
		{
			name:     "ValidWithNoProfiles",
			issuer:   testSetup.Issuer,
			storage:  mockStorage,
			profiles: map[string]*ProfileConfig{},
			wantErr:  false,
		},
		{
			name:    "ValidWithProfiles",
			issuer:  testSetup.Issuer,
			storage: mockStorage,
			profiles: map[string]*ProfileConfig{
				"test-profile": {
					ID:          "test-profile",
					Name:        "Test Profile",
					Description: "A test profile",
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			handler, err := NewHandler(tc.issuer, tc.storage, tc.profiles)
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errSubstr)
				require.Nil(t, handler)
			} else {
				require.NoError(t, err)
				require.NotNil(t, handler)
			}
		})
	}
}

// TestLookupCertificateBySerialWithCert tests serial lookup with stored certificate.
func TestLookupCertificateBySerialWithCert(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  testSetup.Issuer,
	}

	// Generate a test certificate.
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	serialNumber := big.NewInt(12345)

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "test.example.com",
		},
		NotBefore: time.Now().UTC(),
		NotAfter:  time.Now().UTC().Add(24 * time.Hour),
	}

	// Self-sign for testing.
	certDER, err := x509.CreateCertificate(crand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	// Store the certificate.
	storedCert := &cryptoutilCAStorage.StoredCertificate{
		SerialNumber:   serialNumber.Text(16),
		CertificatePEM: string(certPEM),
	}
	ctx := context.Background()
	err = mockStorage.Store(ctx, storedCert)
	require.NoError(t, err)

	// Look up by serial.
	foundCert := handler.lookupCertificateBySerial(ctx, serialNumber)
	require.NotNil(t, foundCert)
	require.Equal(t, serialNumber.Int64(), foundCert.SerialNumber.Int64())
}

// TestLookupCertificateBySerialInvalidPEM tests serial lookup with invalid PEM.
func TestLookupCertificateBySerialInvalidPEM(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)
	mockStorage := cryptoutilCAStorage.NewMemoryStore()

	handler := &Handler{
		storage: mockStorage,
		issuer:  testSetup.Issuer,
	}

	serialNumber := big.NewInt(99999)

	// Store certificate with invalid PEM.
	storedCert := &cryptoutilCAStorage.StoredCertificate{
		SerialNumber:   serialNumber.Text(16),
		CertificatePEM: "not-valid-pem-data",
	}
	ctx := context.Background()
	err := mockStorage.Store(ctx, storedCert)
	require.NoError(t, err)

	// Should return nil for invalid PEM.
	foundCert := handler.lookupCertificateBySerial(ctx, serialNumber)
	require.Nil(t, foundCert)
}

// TestEstCSRAttrsHandler tests the EST CSR attributes endpoint.
func TestEstCSRAttrsHandler(t *testing.T) {
	t.Parallel()

	testSetup := createTestIssuer(t)

	handler := &Handler{
		storage: cryptoutilCAStorage.NewMemoryStore(),
		issuer:  testSetup.Issuer,
		profiles: map[string]*ProfileConfig{
			"server": {
				ID:          "server",
				Name:        "Server Profile",
				Description: "Server certificate profile",
			},
		},
	}

	app := fiber.New()
	app.Get("/est/csrattrs", func(c *fiber.Ctx) error {
		return handler.EstCSRAttrs(c)
	})

	t.Run("WithProfile", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/est/csrattrs?profile=server", nil)
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		// CSR attrs returns 200 or 204 based on implementation.
		require.True(t, resp.StatusCode == fiber.StatusOK || resp.StatusCode == fiber.StatusNoContent)

		err := resp.Body.Close()
		require.NoError(t, err)
	})

	t.Run("WithoutProfile", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/est/csrattrs", nil)
		resp, testErr := app.Test(req)
		require.NoError(t, testErr)
		// CSR attrs returns 200 or 204 based on implementation.
		require.True(t, resp.StatusCode == fiber.StatusOK || resp.StatusCode == fiber.StatusNoContent)

		err := resp.Body.Close()
		require.NoError(t, err)
	})
}
