// Copyright (c) 2025 Justin Cranford
//
//

// Package tls provides TLS configuration utilities for creating secure server and client
// configurations. This package enforces security best practices including TLS 1.3 only,
// full certificate validation, and proper CA chain handling.
package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
)

var configBuildTLSCertificateFn = cryptoutilSharedCryptoCertificate.BuildTLSCertificate

// MinTLSVersion is the minimum TLS version allowed (TLS 1.3 only per Session 4 Q5).
const MinTLSVersion = tls.VersionTLS13

// Config holds the TLS configuration for a server or client.
type Config struct {
	Certificate         *tls.Certificate
	RootCAsPool         *x509.CertPool
	IntermediateCAsPool *x509.CertPool
	TLSConfig           *tls.Config
}

// ServerConfigOptions holds options for creating a server TLS configuration.
type ServerConfigOptions struct {
	// Subject contains the certificate chain and key material.
	Subject *cryptoutilSharedCryptoCertificate.Subject

	// ClientAuth specifies the client authentication mode.
	// Use tls.NoClientCert for server-only TLS, tls.RequireAndVerifyClientCert for mTLS.
	ClientAuth tls.ClientAuthType

	// ClientCAs is the pool of root CAs to verify client certificates (for mTLS).
	// If nil and ClientAuth requires verification, RootCAsPool from Subject will be used.
	ClientCAs *x509.CertPool

	// CipherSuites is an optional list of allowed cipher suites.
	// If empty, Go's default TLS 1.3 cipher suites will be used.
	CipherSuites []uint16
}

// ClientConfigOptions holds options for creating a client TLS configuration.
type ClientConfigOptions struct {
	// ClientSubject contains the client certificate chain and key material (for mTLS).
	// If nil, no client certificate will be presented.
	ClientSubject *cryptoutilSharedCryptoCertificate.Subject

	// RootCAs is the pool of root CAs to verify server certificates.
	RootCAs *x509.CertPool

	// ServerName is the expected server name for verification.
	// If empty, the server name will be extracted from the connection.
	ServerName string

	// SkipVerify disables server certificate verification.
	// CRITICAL: This should NEVER be true in production (Session 4 Q4).
	SkipVerify bool
}

// NewServerConfig creates a TLS configuration for a server.
// This enforces TLS 1.3 only and full certificate validation (Session 4 Q4, Q5).
func NewServerConfig(opts *ServerConfigOptions) (*Config, error) {
	if opts == nil {
		return nil, fmt.Errorf("options cannot be nil")
	} else if opts.Subject == nil {
		return nil, fmt.Errorf("subject cannot be nil")
	}

	tlsCert, rootCAsPool, intermediateCAsPool, err := configBuildTLSCertificateFn(opts.Subject)
	if err != nil {
		return nil, fmt.Errorf("failed to build TLS certificate: %w", err)
	}

	// Determine client CAs for mTLS
	clientCAs := opts.ClientCAs
	if clientCAs == nil && (opts.ClientAuth == tls.RequireAndVerifyClientCert ||
		opts.ClientAuth == tls.VerifyClientCertIfGiven ||
		opts.ClientAuth == tls.RequireAnyClientCert) {
		clientCAs = rootCAsPool
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*tlsCert},
		MinVersion:   MinTLSVersion,
		ClientAuth:   opts.ClientAuth,
		ClientCAs:    clientCAs,
	}

	// Only set cipher suites if explicitly provided
	// For TLS 1.3, Go manages cipher suites automatically.
	if len(opts.CipherSuites) > 0 {
		tlsConfig.CipherSuites = opts.CipherSuites
	}

	return &Config{
		Certificate:         tlsCert,
		RootCAsPool:         rootCAsPool,
		IntermediateCAsPool: intermediateCAsPool,
		TLSConfig:           tlsConfig,
	}, nil
}

// NewClientConfig creates a TLS configuration for a client.
// This enforces TLS 1.3 only and full certificate validation by default (Session 4 Q4, Q5).
func NewClientConfig(opts *ClientConfigOptions) (*Config, error) {
	if opts == nil {
		return nil, fmt.Errorf("options cannot be nil")
	}

	tlsConfig := &tls.Config{
		MinVersion:         MinTLSVersion,
		RootCAs:            opts.RootCAs,
		ServerName:         opts.ServerName,
		InsecureSkipVerify: opts.SkipVerify, //nolint:gosec // Only for explicit test scenarios
	}

	var tlsCert *tls.Certificate

	var rootCAsPool *x509.CertPool

	var intermediateCAsPool *x509.CertPool

	// Add client certificate if provided (for mTLS).
	if opts.ClientSubject != nil {
		var err error

		tlsCert, rootCAsPool, intermediateCAsPool, err = configBuildTLSCertificateFn(opts.ClientSubject)
		if err != nil {
			return nil, fmt.Errorf("failed to build client TLS certificate: %w", err)
		}

		tlsConfig.Certificates = []tls.Certificate{*tlsCert}
	}

	return &Config{
		Certificate:         tlsCert,
		RootCAsPool:         rootCAsPool,
		IntermediateCAsPool: intermediateCAsPool,
		TLSConfig:           tlsConfig,
	}, nil
}

// ValidateConfig validates that a TLS configuration meets security requirements.
// This ensures TLS 1.3 is enforced and proper certificate validation is configured.
func ValidateConfig(config *tls.Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	} else if config.MinVersion < MinTLSVersion {
		return fmt.Errorf("minimum TLS version must be %d (TLS 1.3), got %d", MinTLSVersion, config.MinVersion)
	} else if config.InsecureSkipVerify {
		return fmt.Errorf("InsecureSkipVerify must be false for production use")
	}

	return nil
}
