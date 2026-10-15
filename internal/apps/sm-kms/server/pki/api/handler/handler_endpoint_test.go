// Copyright (c) 2025-2026 Justin Cranford.
package handler

import (
	"context"
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	http "net/http"
	"net/http/httptest"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAProfileCertificate "cryptoutil/internal/apps/sm-kms/server/pki/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/apps/sm-kms/server/pki/profile/subject"
	cryptoutilCAServiceIssuer "cryptoutil/internal/apps/sm-kms/server/pki/service/issuer"
	cryptoutilCAStorage "cryptoutil/internal/apps/sm-kms/server/pki/storage"
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
				Organization: []string{testOrganizationOriginal},
			},
			override: &cryptoutilApiCaServer.SubjectOverride{
				Organization: &[]string{testOrganizationOverride},
			},
			expectedOrg: []string{testOrganizationOverride},
		},
		{
			name: "override_all_fields",
			initial: &cryptoutilCAProfileSubject.Request{
				Organization:       []string{testOrganizationOriginal},
				OrganizationalUnit: []string{"Original OU"},
				Country:            []string{testCountryUS},
				State:              []string{testStateCalifornia},
				Locality:           []string{testLocalitySanFrancisco},
			},
			override: &cryptoutilApiCaServer.SubjectOverride{
				Organization:       &[]string{testOrganizationNew},
				OrganizationalUnit: &[]string{testOrganizationalUnitNew},
				Country:            &[]string{testCountryCA},
				State:              &[]string{testStateOntario},
				Locality:           &[]string{testLocalityToronto},
			},
			expectedOrg:    []string{testOrganizationNew},
			expectedOU:     []string{testOrganizationalUnitNew},
			expectedCounty: []string{testCountryCA},
			expectedState:  []string{testStateOntario},
			expectedLocale: []string{testLocalityToronto},
		},
		{
			name: "empty_override_no_change",
			initial: &cryptoutilCAProfileSubject.Request{
				Organization: []string{testOrganizationKeep},
			},
			override: &cryptoutilApiCaServer.SubjectOverride{
				Organization: &[]string{}, // Empty slice should not override.
			},
			expectedOrg: []string{testOrganizationKeep},
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
				DNSNames: []string{testDomainNameOriginal},
			},
			override: &cryptoutilApiCaServer.SANOverride{
				DNSNames: &[]string{testDomainNameOverride},
			},
			expectedDNS: []string{testDomainNameOverride},
		},
		{
			name: "override_all_sans",
			initial: &cryptoutilCAProfileSubject.Request{
				DNSNames:       []string{testDomainNameOriginal},
				IPAddresses:    []string{testIPAddress192},
				EmailAddresses: []string{"original@example.com"},
				URIs:           []string{"https://original.example.com"},
			},
			override: &cryptoutilApiCaServer.SANOverride{
				DNSNames:       &[]string{testDomainNameNew},
				IPAddresses:    &[]string{testIPAddress10},
				EmailAddresses: &[]string{testEmailAddressNew},
				Uris:           &[]string{testURLNew},
			},
			expectedDNS:    []string{testDomainNameNew},
			expectedIPs:    []string{testIPAddress10},
			expectedEmails: []string{testEmailAddressNew},
			expectedURIs:   []string{testURLNew},
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
			CommonName:   testDomainNameExample,
			Organization: []string{testOrganizationTest},
		},
		DNSNames: []string{testDomainNameExample, testDomainNameWwwExample},
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
			CommonName:   testDomainNameExample,
			Organization: []string{testOrganizationTest},
			Country:      []string{testCountryUS},
		},
		NotBefore:   time.Now().UTC(),
		NotAfter:    time.Now().UTC().Add(cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		DNSNames:    []string{testDomainNameExample},
		IPAddresses: []net.IP{net.ParseIP(testIPAddress192)},
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

	requestID := googleUuid.Must(googleUuid.NewV7())
	result := handler.buildEnrollmentResponse(issued, requestID)
	require.NotNil(t, result)
	require.Equal(t, cryptoutilApiCaServer.Issued, result.Status)
	require.Equal(t, requestID, result.RequestID)
	require.Equal(t, "01", result.Certificate.SerialNumber)
	require.NotNil(t, result.Certificate.FingerprintSha256)
	require.Equal(t, "abc123", *result.Certificate.FingerprintSha256)
}

func TestCertSubjectToAPI(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	cert := &x509.Certificate{
		Subject: pkix.Name{
			CommonName:         testDomainNameExample,
			Organization:       []string{testOrganizationTest},
			OrganizationalUnit: []string{"Test OU"},
			Country:            []string{testCountryUS},
			Province:           []string{testStateCalifornia},
			Locality:           []string{testLocalitySanFrancisco},
		},
		DNSNames:       []string{testDomainNameExample, testDomainNameWwwExample},
		IPAddresses:    []net.IP{net.ParseIP(testIPAddress192)},
		EmailAddresses: []string{"admin@example.com"},
	}

	result := handler.certSubjectToAPI(cert)
	require.NotNil(t, result.CommonName)
	require.Equal(t, testDomainNameExample, *result.CommonName)
	require.NotNil(t, result.Organization)
	require.Equal(t, []string{testOrganizationTest}, *result.Organization)
	require.NotNil(t, result.DNSNames)
	require.Equal(t, []string{testDomainNameExample, testDomainNameWwwExample}, *result.DNSNames)
}

func TestBuildProfileResponse(t *testing.T) {
	t.Parallel()

	handler := &Handler{}

	profile := &ProfileConfig{
		ID:          cryptoutilSharedMagic.PKIProfileIDTLSServer,
		Name:        "TLS Server Certificate",
		Description: "A profile for TLS server certificates",
		Category:    cryptoutilSharedMagic.PKIProfileCategoryTLS,
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
	require.Equal(t, cryptoutilSharedMagic.PKIProfileIDTLSServer, result.ID)
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

func TestLookupCertificateBySerial(t *testing.T) {
	t.Parallel()

	handler := &Handler{storage: cryptoutilCAStorage.NewMemoryStore()}

	tests := []struct {
		name   string
		serial *big.Int
	}{
		{name: "nil_serial", serial: nil},
		{name: testErrorCodeNotFound, serial: big.NewInt(12345)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := handler.lookupCertificateBySerial(context.Background(), tc.serial)
			require.Nil(t, result)
		})
	}
}

func TestHandlerSetServices(t *testing.T) {
	t.Parallel()

	mockStorage := cryptoutilCAStorage.NewMemoryStore()
	handler := &Handler{
		storage: mockStorage,
	}

	handler.SetTSAService(nil)
	handler.SetOCSPService(nil)
	handler.SetCRLService(nil)

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
