// Copyright (c) 2025 Justin Cranford

// Package middleware provides HTTP middleware for the CA server.
package middleware

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"

	cryptoutilCAMagic "cryptoutil/internal/apps/pki/ca/magic"

	fiber "github.com/gofiber/fiber/v2"
)

// MTLSConfig configures mTLS authentication middleware.
type MTLSConfig struct {
	// RequireClientCert requires client certificate for all protected routes.
	RequireClientCert bool

	// AllowedCNs restricts allowed Common Names (empty = allow all).
	AllowedCNs []string

	// AllowedOUs restricts allowed Organizational Units (empty = allow all).
	AllowedOUs []string

	// AllowedDNSSANs restricts allowed DNS Subject Alternative Names (empty = allow all).
	AllowedDNSSANs []string

	// AllowedEmailSANs restricts allowed Email Subject Alternative Names (empty = allow all).
	AllowedEmailSANs []string

	// TrustedCAs for additional certificate validation beyond TLS handshake.
	TrustedCAs *x509.CertPool

	// ValidateExtKeyUsage requires specific extended key usages.
	ValidateExtKeyUsage []x509.ExtKeyUsage

	// AllowExpired allows expired certificates (for testing only).
	AllowExpired bool
}

// DefaultMTLSConfig returns default mTLS configuration.
func DefaultMTLSConfig() *MTLSConfig {
	return &MTLSConfig{
		RequireClientCert:   true,
		AllowedCNs:          nil,
		AllowedOUs:          nil,
		AllowedDNSSANs:      nil,
		AllowedEmailSANs:    nil,
		TrustedCAs:          nil,
		ValidateExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		AllowExpired:        false,
	}
}

// MTLSMiddleware provides mTLS authentication.
type MTLSMiddleware struct {
	config        *MTLSConfig
	getTLSStateFn func(c *fiber.Ctx) *tls.ConnectionState
}

// NewMTLSMiddleware creates a new mTLS middleware.
func NewMTLSMiddleware(config *MTLSConfig) *MTLSMiddleware {
	if config == nil {
		config = DefaultMTLSConfig()
	}

	return &MTLSMiddleware{
		config: config,
		getTLSStateFn: func(c *fiber.Ctx) *tls.ConnectionState {
			return c.Context().TLSConnectionState()
		},
	}
}

// ClientCertInfo contains validated client certificate information.
type ClientCertInfo struct {
	// CommonName from the certificate subject.
	CommonName string

	// OrganizationalUnits from the certificate subject.
	OrganizationalUnits []string

	// DNSNames from Subject Alternative Names.
	DNSNames []string

	// EmailAddresses from Subject Alternative Names.
	EmailAddresses []string

	// SerialNumber as hex string.
	SerialNumber string

	// Certificate is the raw x509 certificate.
	Certificate *x509.Certificate
}

// ClientCertContextKey is the context key for client certificate info.
type ClientCertContextKey struct{}

// Handler returns the Fiber middleware handler.
func (m *MTLSMiddleware) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get TLS connection state.
		tlsState := m.getTLSStateFn(c)
		if tlsState == nil {
			if m.config.RequireClientCert {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "unauthorized",
					"message": "TLS connection required",
				})
			}

			return c.Next()
		}

		// Check for peer certificates.
		if len(tlsState.PeerCertificates) == 0 {
			if m.config.RequireClientCert {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error":   "unauthorized",
					"message": "client certificate required",
				})
			}

			return c.Next()
		}

		// Get client certificate (first in chain).
		clientCert := tlsState.PeerCertificates[0]

		// Validate certificate.
		if err := m.validateCertificate(clientCert); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "forbidden",
				"message": fmt.Sprintf("certificate validation failed: %v", err),
			})
		}

		// Build client cert info.
		certInfo := &ClientCertInfo{
			CommonName:          clientCert.Subject.CommonName,
			OrganizationalUnits: clientCert.Subject.OrganizationalUnit,
			DNSNames:            clientCert.DNSNames,
			EmailAddresses:      clientCert.EmailAddresses,
			SerialNumber:        clientCert.SerialNumber.Text(cryptoutilCAMagic.HexBase),
			Certificate:         clientCert,
		}

		// Store in context.
		c.Locals(ClientCertContextKey{}, certInfo)

		return c.Next()
	}
}

// GetClientCertInfo retrieves client certificate info from context.
func GetClientCertInfo(c *fiber.Ctx) *ClientCertInfo {
	if info, ok := c.Locals(ClientCertContextKey{}).(*ClientCertInfo); ok {
		return info
	}

	return nil
}

// validateCertificate validates the client certificate against configuration.
func (m *MTLSMiddleware) validateCertificate(cert *x509.Certificate) error {
	// Validate CN if configured.
	if len(m.config.AllowedCNs) > 0 {
		if !m.isAllowedValue(cert.Subject.CommonName, m.config.AllowedCNs) {
			return fmt.Errorf("CN '%s' not in allowed list", cert.Subject.CommonName)
		}
	}

	// Validate OU if configured.
	if len(m.config.AllowedOUs) > 0 {
		allowed := false

		for _, ou := range cert.Subject.OrganizationalUnit {
			if m.isAllowedValue(ou, m.config.AllowedOUs) {
				allowed = true

				break
			}
		}

		if !allowed {
			return errors.New("organizational unit not in allowed list")
		}
	}

	// Validate DNS SANs if configured.
	if len(m.config.AllowedDNSSANs) > 0 {
		allowed := false

		for _, san := range cert.DNSNames {
			if m.isAllowedValue(san, m.config.AllowedDNSSANs) {
				allowed = true

				break
			}
		}

		if !allowed {
			return errors.New("DNS SAN not in allowed list")
		}
	}

	// Validate Email SANs if configured.
	if len(m.config.AllowedEmailSANs) > 0 {
		allowed := false

		for _, email := range cert.EmailAddresses {
			if m.isAllowedValue(email, m.config.AllowedEmailSANs) {
				allowed = true

				break
			}
		}

		if !allowed {
			return errors.New("email SAN not in allowed list")
		}
	}

	// Validate extended key usage if configured.
	if len(m.config.ValidateExtKeyUsage) > 0 {
		if err := m.validateExtKeyUsage(cert); err != nil {
			return err
		}
	}

	return nil
}

// validateExtKeyUsage checks that certificate has required extended key usages.
func (m *MTLSMiddleware) validateExtKeyUsage(cert *x509.Certificate) error {
	for _, required := range m.config.ValidateExtKeyUsage {
		found := false

		for _, eku := range cert.ExtKeyUsage {
			if eku == required {
				found = true

				break
			}
		}

		if !found {
			return fmt.Errorf("missing required extended key usage: %v", required)
		}
	}

	return nil
}

// isAllowedValue checks if value is in allowed list (case-insensitive).
func (m *MTLSMiddleware) isAllowedValue(value string, allowed []string) bool {
	for _, a := range allowed {
		if strings.EqualFold(value, a) {
			return true
		}
	}

	return false
}

// RequireClientCert creates middleware that requires a valid client certificate.
func RequireClientCert() fiber.Handler {
	mw := NewMTLSMiddleware(DefaultMTLSConfig())

	return mw.Handler()
}

// RequireClientCertWithConfig creates middleware with custom configuration.
func RequireClientCertWithConfig(config *MTLSConfig) fiber.Handler {
	mw := NewMTLSMiddleware(config)

	return mw.Handler()
}
