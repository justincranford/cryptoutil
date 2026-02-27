// Copyright (c) 2025 Justin Cranford

package middleware

import (
	"crypto/tls"
	"crypto/x509"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// TestHandler_TLSState_NoPeerCerts tests the Handler when TLS state exists but has no peer certificates.
func TestHandler_TLSState_NoPeerCerts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		requireClientCert  bool
		expectedStatusCode int
	}{
		{
			name:               "required_no_peer_certs_returns_401",
			requireClientCert:  true,
			expectedStatusCode: fiber.StatusUnauthorized,
		},
		{
			name:               "optional_no_peer_certs_returns_200",
			requireClientCert:  false,
			expectedStatusCode: fiber.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mw := NewMTLSMiddleware(&MTLSConfig{
				RequireClientCert: tc.requireClientCert,
			})

			// Inject TLS state with no peer certificates.
			mw.getTLSStateFn = func(_ *fiber.Ctx) *tls.ConnectionState {
				return &tls.ConnectionState{
					PeerCertificates: nil,
				}
			}

			app := fiber.New(fiber.Config{DisableStartupMessage: true})
			app.Use(mw.Handler())
			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendStatus(fiber.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)

			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			require.Equal(t, tc.expectedStatusCode, resp.StatusCode)

			err = resp.Body.Close()
			require.NoError(t, err)
		})
	}
}

// TestHandler_TLSState_ValidCert tests the Handler with a valid peer certificate.
func TestHandler_TLSState_ValidCert(t *testing.T) {
	t.Parallel()

	cert, _ := generateTestCertificate(t, "test-client", []string{"Engineering"}, []string{"test.example.com"}, []string{"test@example.com"}, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})

	mw := NewMTLSMiddleware(&MTLSConfig{
		RequireClientCert: true,
	})

	// Inject TLS state with a valid peer certificate.
	mw.getTLSStateFn = func(_ *fiber.Ctx) *tls.ConnectionState {
		return &tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{cert},
		}
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(mw.Handler())
	app.Get("/test", func(c *fiber.Ctx) error {
		info := GetClientCertInfo(c)
		require.NotNil(t, info)
		require.Equal(t, "test-client", info.CommonName)
		require.Equal(t, []string{"Engineering"}, info.OrganizationalUnits)
		require.Equal(t, []string{"test.example.com"}, info.DNSNames)
		require.Equal(t, []string{"test@example.com"}, info.EmailAddresses)
		require.NotEmpty(t, info.SerialNumber)
		require.Equal(t, cert, info.Certificate)

		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// TestHandler_TLSState_ValidationFailure tests the Handler when certificate validation fails.
func TestHandler_TLSState_ValidationFailure(t *testing.T) {
	t.Parallel()

	// Create a certificate with CN "bad-client" that won't match allowed CNs.
	cert, _ := generateTestCertificate(t, "bad-client", nil, nil, nil, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})

	mw := NewMTLSMiddleware(&MTLSConfig{
		RequireClientCert: true,
		AllowedCNs:        []string{"good-client"},
	})

	// Inject TLS state with a certificate that fails validation.
	mw.getTLSStateFn = func(_ *fiber.Ctx) *tls.ConnectionState {
		return &tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{cert},
		}
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(mw.Handler())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}

// TestGetClientCertInfo_WithInfo tests GetClientCertInfo when cert info is set in locals.
func TestGetClientCertInfo_WithInfo(t *testing.T) {
	t.Parallel()

	expectedInfo := &ClientCertInfo{
		CommonName:          "test-cn",
		OrganizationalUnits: []string{"test-ou"},
		DNSNames:            []string{"test.example.com"},
		EmailAddresses:      []string{"test@example.com"},
		SerialNumber:        "abc123",
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Get("/test", func(c *fiber.Ctx) error {
		c.Locals(ClientCertContextKey{}, expectedInfo)

		info := GetClientCertInfo(c)
		require.NotNil(t, info)
		require.Equal(t, expectedInfo.CommonName, info.CommonName)
		require.Equal(t, expectedInfo.SerialNumber, info.SerialNumber)

		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	require.Equal(t, fiber.StatusOK, resp.StatusCode)

	err = resp.Body.Close()
	require.NoError(t, err)
}
