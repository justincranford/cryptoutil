// Copyright (c) 2025 Justin Cranford
//
//

// Package demo provides demo-specific utilities including CA generation.
// This package is for demo purposes only and should not be used in production.
package demo

import (
	"crypto/x509"
	"fmt"
	"sync"

	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
)

// DemoCA holds a pre-generated CA chain for demo purposes.
// This is a singleton that is lazily initialized on first access.
type DemoCA struct {
	Chain *cryptoutilSharedCryptoTls.CAChain
}

// DefaultDemoCAOptions are the options used for the demo CA.
// Uses FQDN style with "cryptoutil.demo.local" prefix.
var DefaultDemoCAOptions = &cryptoutilSharedCryptoTls.CAChainOptions{
	ChainLength:      cryptoutilSharedCryptoTls.DefaultCAChainLength,
	CommonNamePrefix: "cryptoutil.demo.local",
	CNStyle:          cryptoutilSharedCryptoTls.CNStyleFQDN,
	Duration:         cryptoutilSharedCryptoTls.DefaultCADuration,
	Curve:            cryptoutilSharedCryptoTls.DefaultECCurve,
}

var (
	demoCAInstance *DemoCA
	demoCAOnce     sync.Once
	demoCAErr      error
)

// GetDemoCA returns the singleton demo CA instance.
// The CA is lazily created on first access.
func GetDemoCA() (*DemoCA, error) {
	demoCAOnce.Do(func() {
		chain, err := cryptoutilSharedCryptoTls.CreateCAChain(DefaultDemoCAOptions)
		if err != nil {
			demoCAErr = fmt.Errorf("failed to create demo CA chain: %w", err)

			return
		}

		demoCAInstance = &DemoCA{
			Chain: chain,
		}
	})

	return demoCAInstance, demoCAErr
}

// CreateDemoCA creates a new demo CA with the default options.
// Unlike GetDemoCA, this creates a fresh CA each time.
func CreateDemoCA() (*DemoCA, error) {
	chain, err := cryptoutilSharedCryptoTls.CreateCAChain(DefaultDemoCAOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create demo CA chain: %w", err)
	}

	return &DemoCA{
		Chain: chain,
	}, nil
}

// CreateDemoCAWithOptions creates a new demo CA with custom options.
func CreateDemoCAWithOptions(opts *cryptoutilSharedCryptoTls.CAChainOptions) (*DemoCA, error) {
	if opts == nil {
		opts = DefaultDemoCAOptions
	}

	chain, err := cryptoutilSharedCryptoTls.CreateCAChain(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create demo CA chain: %w", err)
	}

	return &DemoCA{
		Chain: chain,
	}, nil
}

// CreateServerCertificate creates a TLS server certificate for the demo.
func (d *DemoCA) CreateServerCertificate(serverName string) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	if serverName == "" {
		return nil, fmt.Errorf("server name cannot be empty")
	}

	opts := cryptoutilSharedCryptoTls.ServerEndEntityOptions(
		serverName,
		[]string{serverName},
		nil, // No IP addresses
	)

	subject, err := d.Chain.CreateEndEntity(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create server certificate: %w", err)
	}

	return subject, nil
}

// CreateClientCertificate creates a TLS client certificate for the demo.
func (d *DemoCA) CreateClientCertificate(clientName string) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	if clientName == "" {
		return nil, fmt.Errorf("client name cannot be empty")
	}

	opts := cryptoutilSharedCryptoTls.ClientEndEntityOptions(clientName)

	subject, err := d.Chain.CreateEndEntity(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create client certificate: %w", err)
	}

	return subject, nil
}

// RootCAsPool returns the root CAs pool for certificate validation.
func (d *DemoCA) RootCAsPool() *x509.CertPool {
	return d.Chain.RootCAsPool()
}

// IntermediateCAsPool returns the intermediate CAs pool.
func (d *DemoCA) IntermediateCAsPool() *x509.CertPool {
	return d.Chain.IntermediateCAsPool()
}
