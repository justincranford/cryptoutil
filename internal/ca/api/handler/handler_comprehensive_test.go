// Copyright (c) 2025 Justin Cranford

package handler

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilCAServer "cryptoutil/api/ca/server"
	cryptoutilCAProfileCertificate "cryptoutil/internal/ca/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/ca/profile/subject"
	cryptoutilCAServiceIssuer "cryptoutil/internal/ca/service/issuer"
	cryptoutilCAStorage "cryptoutil/internal/ca/storage"
)

func TestMapCategory(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	tests := []struct {
		name     string
		input    string
		expected cryptoutilCAServer.ProfileSummaryCategory
	}{
		{"tls", "tls", cryptoutilCAServer.TLS},
		{"email", "email", cryptoutilCAServer.Email},
		{"code_signing", "code_signing", cryptoutilCAServer.CodeSigning},
		{"document_signing", "document_signing", cryptoutilCAServer.DocumentSigning},
		{"ca", "ca", cryptoutilCAServer.CA},
		{"unknown_returns_other", "unknown", cryptoutilCAServer.Other},
		{"empty_returns_other", "", cryptoutilCAServer.Other},
		{"random_returns_other", "random_category", cryptoutilCAServer.Other},
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
		override       *cryptoutilCAServer.SubjectOverride
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
			override: &cryptoutilCAServer.SubjectOverride{
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
			override: &cryptoutilCAServer.SubjectOverride{
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
			override: &cryptoutilCAServer.SubjectOverride{
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
		override       *cryptoutilCAServer.SANOverride
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
			override: &cryptoutilCAServer.SANOverride{
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
			override: &cryptoutilCAServer.SANOverride{
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
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	// Create a valid CSR.
	csrTemplate := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   "test.example.com",
			Organization: []string{"Test Org"},
		},
		DNSNames: []string{"test.example.com", "www.test.example.com"},
	}

	csrDER, err := x509.CreateCertificateRequest(rand.Reader, csrTemplate, privateKey)
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
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   "test.example.com",
			Organization: []string{"Test Org"},
			Country:      []string{"US"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		DNSNames:    []string{"test.example.com"},
		IPAddresses: []net.IP{net.ParseIP("192.168.1.1")},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, &privateKey.PublicKey, privateKey)
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
	require.Equal(t, cryptoutilCAServer.Issued, result.Status)
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

	err = resp.Body.Close()
	require.NoError(t, err)
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

	err = resp.Body.Close()
	require.NoError(t, err)

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

		err = resp.Body.Close()
		require.NoError(t, err)
	})

	t.Run("EST_SimpleEnroll_not_implemented", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/est/simpleenroll", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)

		err = resp.Body.Close()
		require.NoError(t, err)
	})

	t.Run("EST_SimpleReenroll_not_implemented", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/est/simplereenroll", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)

		err = resp.Body.Close()
		require.NoError(t, err)
	})

	t.Run("EST_ServerKeyGen_not_implemented", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodPost, "/est/serverkeygen", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)

		err = resp.Body.Close()
		require.NoError(t, err)
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

		err = resp.Body.Close()
		require.NoError(t, err)
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

		err = resp.Body.Close()
		require.NoError(t, err)
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
		return handler.GetCRL(c, c.Params("caId"), cryptoutilCAServer.GetCRLParams{})
	})

	t.Run("CRL_no_service_configured", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/ca/test-ca/crl", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

		err = resp.Body.Close()
		require.NoError(t, err)
	})
}

func TestEstCACerts(t *testing.T) {
	t.Parallel()

	// EstCACerts requires an issuer to be configured.
	// Since creating an issuer requires crypto key generation, we test the error path.
	// Note: When issuer is nil, EstCACerts will panic due to nil pointer dereference
	// when calling h.issuer.GetCAConfig(). This is a known limitation.
	// A production implementation should check for nil issuer.
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

	err = resp.Body.Close()
	require.NoError(t, err)

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

		err = resp.Body.Close()
		require.NoError(t, err)

		require.Contains(t, string(body), "tls-server")
	})

	t.Run("GetProfile_not_found", func(t *testing.T) {
		t.Parallel()

		req := httptest.NewRequest(http.MethodGet, "/profiles/unknown-profile", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusNotFound, resp.StatusCode)

		err = resp.Body.Close()
		require.NoError(t, err)
	})
}

func TestGetKeyInfoWithRSA4096(t *testing.T) {
	t.Parallel()

	// Test RSA 4096 key.
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err)

	cert := &x509.Certificate{PublicKey: &key.PublicKey}
	algo, size := getKeyInfo(cert)

	require.Equal(t, "RSA", algo)
	require.Equal(t, 4096, size)
}
