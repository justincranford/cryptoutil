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

	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedPool "cryptoutil/internal/shared/pool"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"
)

// Generator holds injected dependencies for certificate generation (seam pattern).
type Generator struct {
	keyPool      *cryptoutilSharedPool.ValueGenPool[*cryptoutilSharedCryptoKeygen.KeyPair]
	getKeyPairFn func() *cryptoutilSharedCryptoKeygen.KeyPair
	mkdirAllFn   func(string, os.FileMode) error
	writeFileFn  func(string, []byte, os.FileMode) error
	createCAFn   func(issuer *cryptoutilSharedCryptoCertificate.Subject, issuerKey any, name string, kp *cryptoutilSharedCryptoKeygen.KeyPair, dur time.Duration, maxPath int) (*cryptoutilSharedCryptoCertificate.Subject, error)
	createLeafFn func(issuer *cryptoutilSharedCryptoCertificate.Subject, kp *cryptoutilSharedCryptoKeygen.KeyPair, name string, dur time.Duration, dns []string, ips []net.IP, emails []string, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage) (*cryptoutilSharedCryptoCertificate.Subject, error)
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
	}, nil
}

// Shutdown stops the key generation pool.
func (g *Generator) Shutdown() {
	cryptoutilSharedPool.CancelNotNil(g.keyPool)
}

// Generate creates the full /certs directory structure for the given tier ID
// per the specification in docs/tls-structure.md.
func (g *Generator) Generate(tierID, targetDir string) error {
	_, psIDs, err := ResolveTier(tierID)
	if err != nil {
		return fmt.Errorf("failed to resolve tier %q: %w", tierID, err)
	}

	if err := g.validateTargetDir(targetDir); err != nil {
		return fmt.Errorf("failed to validate target directory %q: %w", targetDir, err)
	}

	validity := cryptoutilSharedMagic.PKIInitCertValidityDays * cryptoutilSharedMagic.HoursPerDay * time.Hour

	shared, err := g.generateSharedDomains(targetDir, validity)
	if err != nil {
		return fmt.Errorf("failed to generate shared domains: %w", err)
	}

	errCh := make(chan error, len(psIDs))

	var wg sync.WaitGroup

	for _, psID := range psIDs {
		wg.Add(1)

		go func(id string) {
			defer wg.Done()

			if genErr := g.generatePSIDDomains(targetDir, id, shared, validity); genErr != nil {
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

// sharedDomains holds the shared CA chains used across all PS-IDs.
type sharedDomains struct {
	publicServerCA     *cryptoutilSharedCryptoCertificate.Subject // ALL-app-public-server.
	dbLeaderClientCA   *cryptoutilSharedCryptoCertificate.Subject // ALL-db-postgresql-leader-private-client.
	dbFollowerClientCA *cryptoutilSharedCryptoCertificate.Subject // ALL-db-postgresql-follower-private-client.
	otelClientCA       *cryptoutilSharedCryptoCertificate.Subject // ALL-telemetry-otel-private-client.
}

//nolint:cyclop // shared domains mirror tls-structure.md spec sections; splitting would obscure the 1:1 mapping.
func (g *Generator) generateSharedDomains(targetDir string, validity time.Duration) (*sharedDomains, error) {
	sd := &sharedDomains{}

	var err error

	// --- ALL-app-public-server/ (CA chain for all public HTTPS servers) ---
	publicServerDir := filepath.Join(targetDir, "ALL-app-public-server")

	sd.publicServerCA, err = g.generateCAChain(publicServerDir, "ALL-app-public-server", validity)
	if err != nil {
		return nil, fmt.Errorf("public server CA: %w", err)
	}

	// --- ALL-telemetry-grafana-lgtm-public-server/ (Grafana LGTM public server leaf, issued by ALL-app-public-server) ---
	grafanaPublicServerDir := filepath.Join(targetDir, "ALL-telemetry-grafana-lgtm-public-server")

	if err := g.generateServerLeaf(grafanaPublicServerDir, "ALL-telemetry-grafana-lgtm-public-server", sd.publicServerCA, validity,
		[]string{cryptoutilSharedMagic.DockerServiceGrafanaOtelLgtm, cryptoutilSharedMagic.HostnameLocalhost}, defaultIPs()); err != nil {
		return nil, fmt.Errorf("grafana LGTM public server leaf: %w", err)
	}

	// --- ALL-telemetry-grafana-lgtm-public-client/ (CA chain + admin leaf) ---
	grafanaPublicClientDir := filepath.Join(targetDir, "ALL-telemetry-grafana-lgtm-public-client")

	grafanaPublicClientCA, err := g.generateCAChain(grafanaPublicClientDir, "ALL-telemetry-grafana-lgtm-public-client", validity)
	if err != nil {
		return nil, fmt.Errorf("grafana LGTM public client CA: %w", err)
	}

	if err := g.generateClientLeaf(grafanaPublicClientDir,
		"ALL-telemetry-grafana-lgtm-public-client-admin", "ALL-telemetry-grafana-lgtm-public-client-admin",
		grafanaPublicClientCA, validity); err != nil {
		return nil, fmt.Errorf("grafana LGTM public client admin leaf: %w", err)
	}

	// --- ALL-db-postgres-private-server/ (CA chain + leader + follower server leaves) ---
	// Note: top-level dir uses "postgres" per spec; CA/leaf names use "postgresql".
	dbServerDir := filepath.Join(targetDir, "ALL-db-postgres-private-server")

	dbServerCA, err := g.generateCAChain(dbServerDir, "ALL-db-postgresql-private-server", validity)
	if err != nil {
		return nil, fmt.Errorf("DB server CA: %w", err)
	}

	if err := g.generateServerLeaf(dbServerDir, "ALL-db-postgresql-leader-private-server", dbServerCA, validity,
		[]string{"postgres-leader", cryptoutilSharedMagic.HostnameLocalhost}, defaultIPs()); err != nil {
		return nil, fmt.Errorf("DB leader server leaf: %w", err)
	}

	if err := g.generateServerLeaf(dbServerDir, "ALL-db-postgresql-follower-private-server", dbServerCA, validity,
		[]string{"postgres-follower", cryptoutilSharedMagic.HostnameLocalhost}, defaultIPs()); err != nil {
		return nil, fmt.Errorf("DB follower server leaf: %w", err)
	}

	// --- ALL-db-postgresql-leader-private-client/ (CA chain only) ---
	leaderClientDir := filepath.Join(targetDir, "ALL-db-postgresql-leader-private-client")

	sd.dbLeaderClientCA, err = g.generateCAChain(leaderClientDir, "ALL-db-postgresql-leader-private-client", validity)
	if err != nil {
		return nil, fmt.Errorf("DB leader client CA: %w", err)
	}

	// --- ALL-db-postgresql-leader-private-client-follower/ (follower replication leaf) ---
	leaderClientFollowerDir := filepath.Join(targetDir, "ALL-db-postgresql-leader-private-client-follower")

	if err := g.generateClientLeaf(leaderClientFollowerDir,
		"ALL-db-postgresql-leader-private-client-follower", "ALL-db-postgresql-leader-private-client-follower",
		sd.dbLeaderClientCA, validity); err != nil {
		return nil, fmt.Errorf("DB leader client follower leaf: %w", err)
	}

	// --- ALL-db-postgresql-follower-private-client/ (CA chain only) ---
	followerClientDir := filepath.Join(targetDir, "ALL-db-postgresql-follower-private-client")

	sd.dbFollowerClientCA, err = g.generateCAChain(followerClientDir, "ALL-db-postgresql-follower-private-client", validity)
	if err != nil {
		return nil, fmt.Errorf("DB follower client CA: %w", err)
	}

	// --- ALL-db-postgresql-follower-private-client-leader/ (issuing CA copy + leader replication leaf) ---
	followerClientLeaderDir := filepath.Join(targetDir, "ALL-db-postgresql-follower-private-client-leader")

	if err := g.writeIssuerCert(followerClientLeaderDir,
		"ALL-db-postgresql-follower-private-client-leader-ca-issuing", sd.dbFollowerClientCA); err != nil {
		return nil, fmt.Errorf("DB follower client leader CA issuing copy: %w", err)
	}

	if err := g.generateClientLeaf(followerClientLeaderDir,
		"ALL-db-postgresql-follower-private-client-leader", "ALL-db-postgresql-follower-private-client-leader",
		sd.dbFollowerClientCA, validity); err != nil {
		return nil, fmt.Errorf("DB follower client leader leaf: %w", err)
	}

	// --- ALL-telemetry-otel-private-server/ (CA chain + receiver leaf) ---
	otelServerDir := filepath.Join(targetDir, "ALL-telemetry-otel-private-server")

	otelServerCA, err := g.generateCAChain(otelServerDir, "ALL-telemetry-otel-private-server", validity)
	if err != nil {
		return nil, fmt.Errorf("OTel server CA: %w", err)
	}

	if err := g.generateServerLeaf(otelServerDir, "ALL-telemetry-otel-receiver-private-server", otelServerCA, validity,
		[]string{cryptoutilSharedMagic.IME2EOtelCollectorContainer, cryptoutilSharedMagic.HostnameLocalhost}, defaultIPs()); err != nil {
		return nil, fmt.Errorf("OTel receiver server leaf: %w", err)
	}

	// --- ALL-telemetry-grafana-private-server/ (CA chain + Grafana LGTM private server leaf) ---
	grafanaPrivateServerDir := filepath.Join(targetDir, "ALL-telemetry-grafana-private-server")

	grafanaPrivateServerCA, err := g.generateCAChain(grafanaPrivateServerDir, "ALL-telemetry-grafana-private-server", validity)
	if err != nil {
		return nil, fmt.Errorf("grafana private server CA: %w", err)
	}

	if err := g.generateServerLeaf(grafanaPrivateServerDir, "ALL-telemetry-grafana-lgtm-private-server", grafanaPrivateServerCA, validity,
		[]string{cryptoutilSharedMagic.DockerServiceGrafanaOtelLgtm, cryptoutilSharedMagic.HostnameLocalhost}, defaultIPs()); err != nil {
		return nil, fmt.Errorf("grafana LGTM private server leaf: %w", err)
	}

	// --- ALL-telemetry-otel-private-client/ (CA chain only) ---
	otelClientDir := filepath.Join(targetDir, "ALL-telemetry-otel-private-client")

	sd.otelClientCA, err = g.generateCAChain(otelClientDir, "ALL-telemetry-otel-private-client", validity)
	if err != nil {
		return nil, fmt.Errorf("OTel client CA: %w", err)
	}

	// --- ALL-telemetry-grafana-private-client/ (CA chain only) ---
	grafanaPrivateClientDir := filepath.Join(targetDir, "ALL-telemetry-grafana-private-client")

	grafanaPrivateClientCA, err := g.generateCAChain(grafanaPrivateClientDir, "ALL-telemetry-grafana-private-client", validity)
	if err != nil {
		return nil, fmt.Errorf("grafana private client CA: %w", err)
	}

	// --- ALL-telemetry-otel-grafana-private-client/ (OTel->Grafana client leaf) ---
	otelGrafanaClientDir := filepath.Join(targetDir, "ALL-telemetry-otel-grafana-private-client")

	if err := g.generateClientLeaf(otelGrafanaClientDir,
		"ALL-telemetry-otel-grafana-private-client", "ALL-telemetry-otel-grafana-private-client",
		grafanaPrivateClientCA, validity); err != nil {
		return nil, fmt.Errorf("OTel->Grafana private client leaf: %w", err)
	}

	return sd, nil
}

func (g *Generator) generatePSIDDomains(targetDir, psID string, sd *sharedDomains, validity time.Duration) error {
	instances := AppInstances(psID)
	domains := PKIDomains(psID)
	realms := ClientRealms()

	// --- {PS-ID}-app-public-server/ (4 server leaves, issued by shared public server CA) ---
	publicServerDir := filepath.Join(targetDir, psID+"-app-public-server")

	for _, inst := range instances {
		leafName := inst + "-public-server"
		dns := []string{inst, cryptoutilSharedMagic.HostnameLocalhost}

		if err := g.generateServerLeaf(publicServerDir, leafName, sd.publicServerCA, validity, dns, defaultIPs()); err != nil {
			return fmt.Errorf("public server leaf %s: %w", leafName, err)
		}
	}

	// --- {PS-ID}-app-public-client/ (3 PKI domains × (CA chain + 4 realm leaves)) ---
	publicClientDir := filepath.Join(targetDir, psID+"-app-public-client")

	for _, dom := range domains {
		caName := dom + "-public-client"

		clientCA, err := g.generateCAChain(publicClientDir, caName, validity)
		if err != nil {
			return fmt.Errorf("public client CA %s: %w", caName, err)
		}

		for _, realm := range realms {
			leafName := dom + "-public-client-" + realm

			if err := g.generateClientLeaf(publicClientDir, leafName, leafName, clientCA, validity); err != nil {
				return fmt.Errorf("public client leaf %s: %w", leafName, err)
			}
		}
	}

	// --- {PS-ID}-app-private-mutual/ (4 instances × (CA chain + mTLS leaf)) ---
	privateMutualDir := filepath.Join(targetDir, psID+"-app-private-mutual")

	for _, inst := range instances {
		caName := inst + "-private-mutual"

		mutualCA, err := g.generateCAChain(privateMutualDir, caName, validity)
		if err != nil {
			return fmt.Errorf("private mutual CA %s: %w", caName, err)
		}

		leafName := inst + "-private-mutual-ALL"
		dns := []string{inst, cryptoutilSharedMagic.HostnameLocalhost}

		if err := g.generateMutualLeaf(privateMutualDir, leafName, mutualCA, validity, dns, defaultIPs()); err != nil {
			return fmt.Errorf("private mutual leaf %s: %w", leafName, err)
		}
	}

	// --- {PS-ID}-app-postgresql-ALL-leader-private-client/ (2 client leaves) ---
	leaderClientDir := filepath.Join(targetDir, psID+"-app-postgresql-ALL-leader-private-client")

	for _, suffix := range []string{"postgresql-1", "postgresql-2"} {
		leafName := psID + "-app-" + suffix + "-leader-private-client"

		if err := g.generateClientLeaf(leaderClientDir, leafName, leafName, sd.dbLeaderClientCA, validity); err != nil {
			return fmt.Errorf("PG leader client leaf %s: %w", leafName, err)
		}
	}

	// --- {PS-ID}-app-postgresql-ALL-follower-private-client/ (2 client leaves) ---
	followerClientDir := filepath.Join(targetDir, psID+"-app-postgresql-ALL-follower-private-client")

	for _, suffix := range []string{"postgresql-1", "postgresql-2"} {
		leafName := psID + "-app-" + suffix + "-follower-private-client"

		if err := g.generateClientLeaf(followerClientDir, leafName, leafName, sd.dbFollowerClientCA, validity); err != nil {
			return fmt.Errorf("PG follower client leaf %s: %w", leafName, err)
		}
	}

	// --- {PS-ID}-app-ALL-otel-private-client/ (4 client leaves) ---
	otelClientDir := filepath.Join(targetDir, psID+"-app-ALL-otel-private-client")

	for _, inst := range instances {
		leafName := inst + "-otel-private-client"

		if err := g.generateClientLeaf(otelClientDir, leafName, leafName, sd.otelClientCA, validity); err != nil {
			return fmt.Errorf("OTel client leaf %s: %w", leafName, err)
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
