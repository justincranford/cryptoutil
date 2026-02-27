// Copyright (c) 2025 Justin Cranford

package handler

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"

	cryptoutilApiCaServer "cryptoutil/api/ca/server"
	cryptoutilCAProfileCertificate "cryptoutil/internal/apps/pki/ca/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/apps/pki/ca/profile/subject"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/pki/ca/service/issuer"
	cryptoutilCAStorage "cryptoutil/internal/apps/pki/ca/storage"
)

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
		Type:  cryptoutilSharedMagic.StringPEMTypeCSR,
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
				Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
				Bytes: []byte("not a cert"),
			})),
			wantErr:     true,
			errContains: "expected CERTIFICATE REQUEST",
		},
		{
			name: "invalid_csr_content",
			input: string(pem.EncodeToMemory(&pem.Block{
				Type:  cryptoutilSharedMagic.StringPEMTypeCSR,
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
		NotAfter:    time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		DNSNames:    []string{"test.example.com"},
		IPAddresses: []net.IP{net.ParseIP("192.168.1.1")},
	}

	certDER, err := x509.CreateCertificate(crand.Reader, certTemplate, certTemplate, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
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
	require.Equal(t, cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year, *result.MaxValidityDays)
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
	resp, err := app.Test(req, -1)
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
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.NoError(t, resp.Body.Close())

	require.Contains(t, string(body), "test_error")
	require.Contains(t, string(body), "This is a test error message")
}
