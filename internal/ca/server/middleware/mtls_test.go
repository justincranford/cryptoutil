// Copyright (c) 2025 Justin Cranford

package middleware

import (
	ecdsa "crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net/http/httptest"
	"testing"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// generateTestCertificate creates a test certificate with specified attributes.
func generateTestCertificate(t *testing.T, cn string, ous []string, dnsNames []string, emails []string, extKeyUsages []x509.ExtKeyUsage) (*x509.Certificate, *ecdsa.PrivateKey) {
	t.Helper()

	priv, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	serialNumber, err := crand.Int(crand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:         cn,
			OrganizationalUnit: ous,
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           extKeyUsages,
		DNSNames:              dnsNames,
		EmailAddresses:        emails,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(crand.Reader, template, template, &priv.PublicKey, priv)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	return cert, priv
}

func TestMTLSMiddleware_NoCertificate_NoTLSState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		requireClientCert  bool
		expectedStatusCode int
	}{
		{
			name:               "required_no_tls_returns_401",
			requireClientCert:  true,
			expectedStatusCode: fiber.StatusUnauthorized,
		},
		{
			name:               "optional_no_tls_returns_200",
			requireClientCert:  false,
			expectedStatusCode: fiber.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			config := &MTLSConfig{
				RequireClientCert: tc.requireClientCert,
			}
			app.Use(RequireClientCertWithConfig(config))
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			require.Equal(t, tc.expectedStatusCode, resp.StatusCode)
		})
	}
}

func TestMTLSMiddleware_ValidateCertificate(t *testing.T) {
	t.Parallel()

	cert, _ := generateTestCertificate(t, "test-client", []string{"Engineering"}, []string{"test.example.com"}, []string{"test@example.com"}, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})

	tests := []struct {
		name        string
		config      *MTLSConfig
		expectError bool
	}{
		{
			name:        "valid_default_config",
			config:      DefaultMTLSConfig(),
			expectError: false,
		},
		{
			name: "valid_allowed_cn",
			config: &MTLSConfig{
				RequireClientCert: true,
				AllowedCNs:        []string{"test-client"},
			},
			expectError: false,
		},
		{
			name: "valid_allowed_ou",
			config: &MTLSConfig{
				RequireClientCert: true,
				AllowedOUs:        []string{"Engineering"},
			},
			expectError: false,
		},
		{
			name: "valid_allowed_dns_san",
			config: &MTLSConfig{
				RequireClientCert: true,
				AllowedDNSSANs:    []string{"test.example.com"},
			},
			expectError: false,
		},
		{
			name: "valid_allowed_email_san",
			config: &MTLSConfig{
				RequireClientCert: true,
				AllowedEmailSANs:  []string{"test@example.com"},
			},
			expectError: false,
		},
		{
			name: "invalid_cn_not_in_list",
			config: &MTLSConfig{
				RequireClientCert: true,
				AllowedCNs:        []string{"other-client"},
			},
			expectError: true,
		},
		{
			name: "invalid_ou_not_in_list",
			config: &MTLSConfig{
				RequireClientCert: true,
				AllowedOUs:        []string{"Sales"},
			},
			expectError: true,
		},
		{
			name: "invalid_dns_san_not_in_list",
			config: &MTLSConfig{
				RequireClientCert: true,
				AllowedDNSSANs:    []string{"other.example.com"},
			},
			expectError: true,
		},
		{
			name: "invalid_email_san_not_in_list",
			config: &MTLSConfig{
				RequireClientCert: true,
				AllowedEmailSANs:  []string{"other@example.com"},
			},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mw := NewMTLSMiddleware(tc.config)
			err := mw.validateCertificate(cert)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMTLSMiddleware_ExtKeyUsageValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		certExtKeyUsages   []x509.ExtKeyUsage
		configExtKeyUsages []x509.ExtKeyUsage
		expectError        bool
	}{
		{
			name:               "has_required_client_auth",
			certExtKeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			configExtKeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			expectError:        false,
		},
		{
			name:               "missing_required_client_auth",
			certExtKeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			configExtKeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			expectError:        true,
		},
		{
			name:               "has_multiple_required",
			certExtKeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageEmailProtection},
			configExtKeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageEmailProtection},
			expectError:        false,
		},
		{
			name:               "no_validation_required",
			certExtKeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			configExtKeyUsages: nil,
			expectError:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cert, _ := generateTestCertificate(t, "test-client", nil, nil, nil, tc.certExtKeyUsages)

			config := &MTLSConfig{
				RequireClientCert:   true,
				ValidateExtKeyUsage: tc.configExtKeyUsages,
			}
			mw := NewMTLSMiddleware(config)
			err := mw.validateCertificate(cert)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetClientCertInfo_NoInfo(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Get("/test", func(c *fiber.Ctx) error {
		info := GetClientCertInfo(c)
		require.Nil(t, info)

		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestDefaultMTLSConfig(t *testing.T) {
	t.Parallel()

	config := DefaultMTLSConfig()
	require.True(t, config.RequireClientCert)
	require.Nil(t, config.AllowedCNs)
	require.Nil(t, config.AllowedOUs)
	require.Nil(t, config.AllowedDNSSANs)
	require.Nil(t, config.AllowedEmailSANs)
	require.Nil(t, config.TrustedCAs)
	require.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, config.ValidateExtKeyUsage)
	require.False(t, config.AllowExpired)
}

func TestNewMTLSMiddleware_NilConfig(t *testing.T) {
	t.Parallel()

	mw := NewMTLSMiddleware(nil)
	require.NotNil(t, mw)
	require.NotNil(t, mw.config)
	require.True(t, mw.config.RequireClientCert)
}

func TestCaseInsensitiveMatching(t *testing.T) {
	t.Parallel()

	cert, _ := generateTestCertificate(t, "Test-Client", []string{"Engineering"}, []string{"TEST.example.com"}, []string{"TEST@Example.Com"}, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})

	tests := []struct {
		name        string
		config      *MTLSConfig
		expectError bool
	}{
		{
			name: "cn_case_insensitive",
			config: &MTLSConfig{
				RequireClientCert: true,
				AllowedCNs:        []string{"test-client"},
			},
			expectError: false,
		},
		{
			name: "ou_case_insensitive",
			config: &MTLSConfig{
				RequireClientCert: true,
				AllowedOUs:        []string{"engineering"},
			},
			expectError: false,
		},
		{
			name: "dns_case_insensitive",
			config: &MTLSConfig{
				RequireClientCert: true,
				AllowedDNSSANs:    []string{"test.EXAMPLE.com"},
			},
			expectError: false,
		},
		{
			name: "email_case_insensitive",
			config: &MTLSConfig{
				RequireClientCert: true,
				AllowedEmailSANs:  []string{"test@example.com"},
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mw := NewMTLSMiddleware(tc.config)
			err := mw.validateCertificate(cert)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRequireClientCert_CreatesMiddleware(t *testing.T) {
	t.Parallel()

	handler := RequireClientCert()
	require.NotNil(t, handler)
}

func TestRequireClientCertWithConfig_CreatesMiddleware(t *testing.T) {
	t.Parallel()

	handler := RequireClientCertWithConfig(&MTLSConfig{
		RequireClientCert: true,
		AllowedCNs:        []string{"test-cn"},
	})
	require.NotNil(t, handler)
}

func TestMTLSMiddleware_IsAllowedValue(t *testing.T) {
	t.Parallel()

	mw := NewMTLSMiddleware(DefaultMTLSConfig())

	tests := []struct {
		name    string
		value   string
		allowed []string
		want    bool
	}{
		{
			name:    "exact_match",
			value:   "test",
			allowed: []string{"test"},
			want:    true,
		},
		{
			name:    "case_insensitive_match",
			value:   "TEST",
			allowed: []string{"test"},
			want:    true,
		},
		{
			name:    "no_match",
			value:   "other",
			allowed: []string{"test"},
			want:    false,
		},
		{
			name:    "empty_allowed",
			value:   "test",
			allowed: []string{},
			want:    false,
		},
		{
			name:    "multiple_allowed_match",
			value:   "two",
			allowed: []string{"one", "two", "three"},
			want:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := mw.isAllowedValue(tc.value, tc.allowed)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestMTLSMiddleware_ValidateExtKeyUsage(t *testing.T) {
	t.Parallel()

	cert, _ := generateTestCertificate(t, "test", nil, nil, nil, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageEmailProtection})

	tests := []struct {
		name        string
		required    []x509.ExtKeyUsage
		expectError bool
	}{
		{
			name:        "empty_required_ok",
			required:    nil,
			expectError: false,
		},
		{
			name:        "single_required_present",
			required:    []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			expectError: false,
		},
		{
			name:        "multiple_required_all_present",
			required:    []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageEmailProtection},
			expectError: false,
		},
		{
			name:        "required_not_present",
			required:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mw := &MTLSMiddleware{
				config: &MTLSConfig{
					ValidateExtKeyUsage: tc.required,
				},
			}
			err := mw.validateExtKeyUsage(cert)

			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
