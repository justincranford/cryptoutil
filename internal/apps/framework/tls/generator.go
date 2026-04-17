// Copyright (c) 2025 Justin Cranford
//
//

package tls

import (
	"context"
	"crypto/elliptic"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	pkcs12 "software.sslmate.com/src/go-pkcs12"

	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedPool "cryptoutil/internal/shared/pool"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// Generator holds injected dependencies for certificate generation (seam pattern).
type Generator struct {
	keyPool             *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair]
	getKeyPairFn        func() *cryptoutilSharedCryptoKeygen.KeyPair
	mkdirAllFn          func(string, os.FileMode) error
	writeFileFn         func(string, []byte, os.FileMode) error
	createCAFn          func(issuer *cryptoutilSharedCryptoCertificate.Subject, issuerKey any, name string, kp *cryptoutilSharedCryptoKeygen.KeyPair, dur time.Duration, maxPath int) (*cryptoutilSharedCryptoCertificate.Subject, error)
	createLeafFn        func(issuer *cryptoutilSharedCryptoCertificate.Subject, kp *cryptoutilSharedCryptoKeygen.KeyPair, name string, dur time.Duration, dns []string, ips []net.IP, emails []string, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage) (*cryptoutilSharedCryptoCertificate.Subject, error)
	encodePKCS12Fn      func(priv any, cert *x509.Certificate, chain []*x509.Certificate) ([]byte, error)
	encodeTrustPKCS12Fn func(certs []*x509.Certificate) ([]byte, error)
	getRealmsForPSIDFn  func(psID string) ([]string, error)
}

// NewGenerator creates a Generator with production dependencies.
func NewGenerator(ctx context.Context, telemetryService *cryptoutilSharedTelemetry.TelemetryService) (*Generator, error) {
	switch {
	case ctx == nil:
		return nil, fmt.Errorf("context must be non-nil")
	case telemetryService == nil:
		return nil, fmt.Errorf("telemetry service must be non-nil")
	}

	pool, err := cryptoutilSharedPool.NewValueGenPool(cryptoutilSharedPool.NewValueGenPoolConfig(
		ctx, telemetryService, "pki-init ECDSA-P384",
		cryptoutilSharedMagic.DefaultPoolConfigECDSAP384.NumWorkers,
		cryptoutilSharedMagic.DefaultPoolConfigECDSAP384.MaxSize,
		cryptoutilSharedMagic.MaxPoolLifetimeValues,
		cryptoutilSharedMagic.MaxPoolLifetimeDuration,
		cryptoutilSharedCryptoKeygen.GenerateECDSAKeyPairFunction(elliptic.P384()),
		false,
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create ECDSA P-384 key pool: %w", err)
	}

	return &Generator{
		keyPool:      pool,
		getKeyPairFn: pool.Get,
		mkdirAllFn:   os.MkdirAll,
		writeFileFn:  os.WriteFile,
		createCAFn: func(issuer *cryptoutilSharedCryptoCertificate.Subject, issuerKey any, name string, kp *cryptoutilSharedCryptoKeygen.KeyPair, dur time.Duration, maxPath int) (*cryptoutilSharedCryptoCertificate.Subject, error) {
			return cryptoutilSharedCryptoCertificate.CreateCASubject(issuer, issuerKey, name, kp, dur, maxPath)
		},
		createLeafFn: func(issuer *cryptoutilSharedCryptoCertificate.Subject, kp *cryptoutilSharedCryptoKeygen.KeyPair, name string, dur time.Duration, dns []string, ips []net.IP, emails []string, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage) (*cryptoutilSharedCryptoCertificate.Subject, error) {
			return cryptoutilSharedCryptoCertificate.CreateEndEntitySubject(issuer, kp, name, dur, dns, ips, emails, nil, keyUsage, extKeyUsage)
		},
		encodePKCS12Fn: func(priv any, cert *x509.Certificate, chain []*x509.Certificate) ([]byte, error) {
			return pkcs12.Modern.Encode(priv, cert, chain, "")
		},
		encodeTrustPKCS12Fn: func(certs []*x509.Certificate) ([]byte, error) {
			return pkcs12.Modern.EncodeTrustStore(certs, "")
		},
		getRealmsForPSIDFn: func(psID string) ([]string, error) {
			return readRealmsForPSID(defaultRegistryPath, psID)
		},
	}, nil
}

// Shutdown stops the key generation pool.
func (g *Generator) Shutdown() {
	cryptoutilSharedPool.CancelNotNil(g.keyPool)
}

// Generate creates the full /certs directory structure for the given tier ID
// per the specification in docs/tls-structure.md. All output is written under
// filepath.Join(targetDir, tierID)/.
func (g *Generator) Generate(tierID, targetDir string) error {
	_, psIDs, err := ResolveTier(tierID)
	if err != nil {
		return fmt.Errorf("failed to resolve tier %q: %w", tierID, err)
	}

	basePath := filepath.Join(targetDir, tierID)

	if err := g.validateTargetDir(basePath); err != nil {
		return fmt.Errorf("failed to validate target directory %q: %w", basePath, err)
	}

	shared, err := g.generateSharedCAs(basePath)
	if err != nil {
		return fmt.Errorf("failed to generate shared CAs: %w", err)
	}

	errCh := make(chan error, len(psIDs))

	var wg sync.WaitGroup

	for _, psID := range psIDs {
		wg.Add(1)

		go func(id string) {
			defer wg.Done()

			if genErr := g.generatePSIDCerts(basePath, id, shared); genErr != nil {
				errCh <- fmt.Errorf("PS-ID %s: %w", id, genErr)
			}
		}(psID)
	}

	wg.Wait()
	close(errCh)

	for e := range errCh {
		if err == nil {
			err = e
		}
	}

	return err
}

// sharedCAs holds the issuing CA subjects that are shared across all PS-IDs.
// These are used to sign leaf certificates in the per-PS-ID generation phase.
type sharedCAs struct {
	globalServerIssuing   *cryptoutilSharedCryptoCertificate.Subject // Cat 1: signs Cats 2, 3
	grafanaClientIssuing  *cryptoutilSharedCryptoCertificate.Subject // Cat 8 grafana: signs Cat 9 grafana leaves
	otelClientIssuing     *cryptoutilSharedCryptoCertificate.Subject // Cat 8 otel: signs Cat 9 otel leaves
	postgresServerIssuing *cryptoutilSharedCryptoCertificate.Subject // Cat 10: signs Cat 11
	postgresClientIssuing *cryptoutilSharedCryptoCertificate.Subject // Cat 12: signs Cats 13, 14
}

// generateSharedCAs generates the globally shared certificate infrastructure:
// Categories 1, 2, 8, 9 (admin+infra), 10, 11, 12, 13.
// Returns the shared CA subjects needed to sign per-PS-ID leaf certs.
//
//nolint:cyclop // each block mirrors one TLS category; splitting would obscure the 1:1 spec mapping.
func (g *Generator) generateSharedCAs(basePath string) (*sharedCAs, error) {
	var err error

	shared := &sharedCAs{}

	// --- Category 1: Global HTTPS Server CAs (4 dirs) ---
	shared.globalServerIssuing, err = g.generateCAChain(basePath, "public-https-server", "ca")
	if err != nil {
		return nil, fmt.Errorf("cat1 global server CA: %w", err)
	}

	// --- Category 2: Grafana LGTM + OTel Collector Server Certs (2 dirs) ---
	grafanaDNS := []string{cryptoutilSharedMagic.DockerServiceGrafanaOtelLgtm, cryptoutilSharedMagic.HostnameLocalhost}

	if err := g.generateServerLeafDir(basePath, "public-https-server-entity-"+cryptoutilSharedMagic.DockerServiceGrafanaOtelLgtm,
		shared.globalServerIssuing, grafanaDNS, defaultIPs()); err != nil {
		return nil, fmt.Errorf("cat2 grafana server leaf: %w", err)
	}

	otelDNS := []string{cryptoutilSharedMagic.IME2EOtelCollectorContainer, cryptoutilSharedMagic.HostnameLocalhost}

	if err := g.generateServerLeafDir(basePath, "public-https-server-entity-"+cryptoutilSharedMagic.PKIInitOtelCollectorContrib,
		shared.globalServerIssuing, otelDNS, defaultIPs()); err != nil {
		return nil, fmt.Errorf("cat2 otel server leaf: %w", err)
	}

	// --- Category 8: Grafana LGTM + OTel Collector Client CAs (8 dirs) ---
	shared.grafanaClientIssuing, err = g.generateCAChain(basePath, cryptoutilSharedMagic.DockerServiceGrafanaOtelLgtm+"-https-client", "ca")
	if err != nil {
		return nil, fmt.Errorf("cat8 grafana client CA: %w", err)
	}

	shared.otelClientIssuing, err = g.generateCAChain(basePath, cryptoutilSharedMagic.PKIInitOtelCollectorContrib+"-https-client", "ca")
	if err != nil {
		return nil, fmt.Errorf("cat8 otel client CA: %w", err)
	}

	// --- Category 9 (admin): Admin client certs for Grafana + OTel (2 dirs) ---
	grafanaAdminDir := cryptoutilSharedMagic.DockerServiceGrafanaOtelLgtm + "-https-client-entity-admin"

	if err := g.generateClientLeafDir(basePath, grafanaAdminDir, shared.grafanaClientIssuing); err != nil {
		return nil, fmt.Errorf("cat9 grafana admin client: %w", err)
	}

	otelAdminDir := cryptoutilSharedMagic.PKIInitOtelCollectorContrib + "-https-client-entity-admin"

	if err := g.generateClientLeafDir(basePath, otelAdminDir, shared.otelClientIssuing); err != nil {
		return nil, fmt.Errorf("cat9 otel admin client: %w", err)
	}

	// --- Category 9 (infra): OTel-to-Grafana service-to-service forwarding client certs (2 dirs) ---
	grafanaInfraDir := cryptoutilSharedMagic.DockerServiceGrafanaOtelLgtm + "-https-client-entity-" + cryptoutilSharedMagic.PKIInitEntityInfra

	if err := g.generateClientLeafDir(basePath, grafanaInfraDir, shared.grafanaClientIssuing); err != nil {
		return nil, fmt.Errorf("cat9 grafana infra client: %w", err)
	}

	otelInfraDir := cryptoutilSharedMagic.PKIInitOtelCollectorContrib + "-https-client-entity-" + cryptoutilSharedMagic.PKIInitEntityInfra

	if err := g.generateClientLeafDir(basePath, otelInfraDir, shared.otelClientIssuing); err != nil {
		return nil, fmt.Errorf("cat9 otel infra client: %w", err)
	}

	// --- Category 10: PostgreSQL Server CAs (4 dirs) ---
	shared.postgresServerIssuing, err = g.generateCAChain(basePath, "postgres-tls-server", "ca")
	if err != nil {
		return nil, fmt.Errorf("cat10 postgres server CA: %w", err)
	}

	// --- Category 11: PostgreSQL Server Certs (2 dirs) ---
	leaderDNS := []string{cryptoutilSharedMagic.PKIInitPostgresLeaderService, cryptoutilSharedMagic.HostnameLocalhost}

	if err := g.generateServerLeafDir(basePath, "postgres-tls-server-entity-"+cryptoutilSharedMagic.PKIInitPostgresLeader,
		shared.postgresServerIssuing, leaderDNS, defaultIPs()); err != nil {
		return nil, fmt.Errorf("cat11 postgres leader server leaf: %w", err)
	}

	followerDNS := []string{cryptoutilSharedMagic.PKIInitPostgresFollowerService, cryptoutilSharedMagic.HostnameLocalhost}

	if err := g.generateServerLeafDir(basePath, "postgres-tls-server-entity-"+cryptoutilSharedMagic.PKIInitPostgresFollower,
		shared.postgresServerIssuing, followerDNS, defaultIPs()); err != nil {
		return nil, fmt.Errorf("cat11 postgres follower server leaf: %w", err)
	}

	// --- Category 12: PostgreSQL Client CAs (4 dirs) ---
	shared.postgresClientIssuing, err = g.generateCAChain(basePath, "postgres-tls-client", "ca")
	if err != nil {
		return nil, fmt.Errorf("cat12 postgres client CA: %w", err)
	}

	// --- Category 13: PostgreSQL Replication Client Certs (2 dirs) ---
	if err := g.generateClientLeafDir(basePath, "postgres-tls-client-entity-"+cryptoutilSharedMagic.PKIInitPostgresLeader+"-replication",
		shared.postgresClientIssuing); err != nil {
		return nil, fmt.Errorf("cat13 postgres leader replication client: %w", err)
	}

	if err := g.generateClientLeafDir(basePath, "postgres-tls-client-entity-"+cryptoutilSharedMagic.PKIInitPostgresFollower+"-replication",
		shared.postgresClientIssuing); err != nil {
		return nil, fmt.Errorf("cat13 postgres follower replication client: %w", err)
	}

	return shared, nil
}

// generatePSIDCerts generates all per-PS-ID certificate directories:
// Categories 3, 4, 5, 6, 7, 9 (per-PS-ID), 14.
//
//nolint:cyclop // each block mirrors one TLS category; splitting would obscure the 1:1 spec mapping.
func (g *Generator) generatePSIDCerts(basePath, psID string, shared *sharedCAs) error {
	// --- Category 3: PS-ID App Server Certs (4 dirs) ---
	for _, suffix := range PKIInitAppInstanceSuffixes() {
		dirName := "public-https-server-entity-" + psID + "-" + suffix
		dns := []string{psID + "-app-" + suffix, cryptoutilSharedMagic.HostnameLocalhost}

		if err := g.generateServerLeafDir(basePath, dirName, shared.globalServerIssuing, dns, defaultIPs()); err != nil {
			return fmt.Errorf("cat3 app server leaf %s: %w", suffix, err)
		}
	}

	// --- Categories 4 + 5: PS-ID HTTPS Client CAs and Leaf Certs (12 + 12 dirs) ---
	realms, err := g.getRealmsForPSIDFn(psID)
	if err != nil {
		return fmt.Errorf("cat4/5 read realms for %s: %w", psID, err)
	}

	for _, domain := range PKIInitClientPKIDomains() {
		// Category 4: One CA chain per PKI domain (4 dirs each = 12 total).
		clientIssuing, caErr := g.generateCAChain(basePath, "public-https-client", "ca-"+psID+"-"+domain)
		if caErr != nil {
			return fmt.Errorf("cat4 client CA domain=%s: %w", domain, caErr)
		}

		// Category 5: Client leaf certs per user type × realm (per domain).
		for _, userType := range PKIInitUserTypes() {
			for _, realm := range realms {
				dirName := "public-https-client-entity-" + psID + "-" + domain + "-" + userType + "-" + realm
				if leafErr := g.generateClientLeafDir(basePath, dirName, clientIssuing); leafErr != nil {
					return fmt.Errorf("cat5 client leaf user=%s realm=%s domain=%s: %w", userType, realm, domain, leafErr)
				}
			}
		}
	}

	// --- Categories 6 + 7: Private Admin mTLS CAs and Leaf Certs (16 + 4 dirs) ---
	for _, suffix := range PKIInitAdminInstanceSuffixes() {
		// Category 6: One admin CA chain per instance (4 dirs each = 16 total).
		adminIssuing, caErr := g.generateCAChain(basePath, "private-https-mutual", "ca-"+psID+"-"+suffix)
		if caErr != nil {
			return fmt.Errorf("cat6 admin CA suffix=%s: %w", suffix, caErr)
		}

		// Category 7: One mTLS leaf per instance.
		mutualDir := "private-https-mutual-entity-" + psID + "-" + suffix
		dns := []string{psID + "-app-" + suffix, cryptoutilSharedMagic.HostnameLocalhost}

		if leafErr := g.generateMutualLeafDir(basePath, mutualDir, adminIssuing, dns, defaultIPs()); leafErr != nil {
			return fmt.Errorf("cat7 admin mutual leaf suffix=%s: %w", suffix, leafErr)
		}
	}

	// --- Category 9 (per-PS-ID): Grafana + OTel client certs per PS-ID × instance (8 dirs) ---
	for _, suffix := range PKIInitAppInstanceSuffixes() {
		grafanaPSIDDir := cryptoutilSharedMagic.DockerServiceGrafanaOtelLgtm + "-https-client-entity-" + psID + "-" + suffix

		if err := g.generateClientLeafDir(basePath, grafanaPSIDDir, shared.grafanaClientIssuing); err != nil {
			return fmt.Errorf("cat9 grafana psid client %s-%s: %w", psID, suffix, err)
		}

		otelPSIDDir := cryptoutilSharedMagic.PKIInitOtelCollectorContrib + "-https-client-entity-" + psID + "-" + suffix

		if err := g.generateClientLeafDir(basePath, otelPSIDDir, shared.otelClientIssuing); err != nil {
			return fmt.Errorf("cat9 otel psid client %s-%s: %w", psID, suffix, err)
		}
	}

	// --- Category 14: PS-ID PostgreSQL App Client Certs (4 dirs) ---
	for _, role := range []string{cryptoutilSharedMagic.PKIInitPostgresLeader, cryptoutilSharedMagic.PKIInitPostgresFollower} {
		for _, suffix := range PKIInitPostgresAppInstanceSuffixes() {
			dirName := "postgres-tls-client-entity-" + role + "-" + psID + "-" + suffix

			if err := g.generateClientLeafDir(basePath, dirName, shared.postgresClientIssuing); err != nil {
				return fmt.Errorf("cat14 postgres app client %s-%s: %w", role, suffix, err)
			}
		}
	}

	return nil
}

// defaultIPs returns the default IP SANs for development/E2E certificates.
func defaultIPs() []net.IP {
	return []net.IP{
		net.ParseIP(cryptoutilSharedMagic.IPv4Loopback),
		net.ParseIP(cryptoutilSharedMagic.IPv6Loopback),
	}
}
