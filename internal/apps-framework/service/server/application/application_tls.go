// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilSharedCryptoAsn1 "cryptoutil/internal/shared/crypto/asn1"
	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilNetwork "cryptoutil/internal/shared/util/network"
)

// TLSListener holds in-memory TLS certificate subjects for both public and private servers,
// along with a shutdown function to release the underlying Basic infrastructure.
// Returned by StartTLSListener after successful initialization.
type TLSListener struct {
	PublicTLSServer  *cryptoutilSharedCryptoCertificate.Subject
	PrivateTLSServer *cryptoutilSharedCryptoCertificate.Subject
	ShutdownFunction func()
}

// StartTLSListener initializes core infrastructure (including database connectivity),
// basic services, and in-memory TLS configurations for the public and private servers.
//
// Unlike server init functions, TLS certificates are generated in memory without writing to disk,
// making this function safe to call from parallel tests.
//
// Returns an error if database connectivity fails (e.g., PostgreSQL not running).
func StartTLSListener(settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*TLSListener, error) {
	if settings == nil {
		return nil, fmt.Errorf("settings cannot be nil")
	}

	ctx := context.Background()

	// Initialize core infrastructure including database connectivity.
	// Fails for unavailable databases (e.g., PostgreSQL not running in the test environment).
	core, err := StartCore(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to start application core: %w", err)
	}

	// Initialize basic services (telemetry, unseal keys, JWK generation) for TLS cert generation.
	basic, err := StartBasic(ctx, settings)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to start basic application services: %w", err)
	}

	// Generate TLS certificate subjects in memory (no disk I/O, safe for parallel tests).
	publicSubject, privateSubject, err := GenerateTLSServerSubjectsInMemory(settings, basic)
	if err != nil {
		basic.Shutdown()
		core.Shutdown()

		return nil, fmt.Errorf("failed to generate TLS server subjects: %w", err)
	}

	return &TLSListener{
		PublicTLSServer:  publicSubject,
		PrivateTLSServer: privateSubject,
		ShutdownFunction: func() {
			basic.Shutdown()
			core.Shutdown()
		},
	}, nil
}

// GenerateTLSServerSubjectsInMemory generates TLS server certificate subjects for both public and
// private servers without writing any files to disk. Safe for parallel tests and server startup.
func GenerateTLSServerSubjectsInMemory(settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings, basic *Basic) (*cryptoutilSharedCryptoCertificate.Subject, *cryptoutilSharedCryptoCertificate.Subject, error) {
	publicTLSServerIPAddresses, err := cryptoutilSharedUtilNetwork.ParseIPAddresses(settings.TLSPublicIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse public TLS server IP addresses: %w", err)
	}

	privateTLSServerIPAddresses, err := cryptoutilSharedUtilNetwork.ParseIPAddresses(settings.TLSPrivateIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private TLS server IP addresses: %w", err)
	}

	public, err := GenerateTLSServerSubjectInMemory(basic, settings.TLSPublicDNSNames, publicTLSServerIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TLS public server certs in memory: %w", err)
	}

	private, err := GenerateTLSServerSubjectInMemory(basic, settings.TLSPrivateDNSNames, privateTLSServerIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TLS private server certs in memory: %w", err)
	}

	return public, private, nil
}

// InitTLSServerCerts initializes TLS certificates by starting Basic services,
// generating cert files to disk, then shutting down. Used by the init subcommand.
func InitTLSServerCerts(settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) error {
	ctx := context.Background()

	basic, err := StartBasic(ctx, settings)
	if err != nil {
		return fmt.Errorf("failed to initialize server application core: %w", err)
	}

	defer basic.Shutdown()

	_, _, err = GenerateTLSServerSubjectsToDisk(settings, basic)
	if err != nil {
		return fmt.Errorf("failed to generate TLS server subjects: %w", err)
	}

	return nil
}

// GenerateTLSServerSubjectsToDisk generates TLS server certificate subjects for both public and
// private servers and writes the certificate chain PEMs and encrypted private key PEMs to disk.
// prefix is used to name files (e.g., "tls_public_server_" or "tls_private_server_").
func GenerateTLSServerSubjectsToDisk(settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings, basic *Basic) (*cryptoutilSharedCryptoCertificate.Subject, *cryptoutilSharedCryptoCertificate.Subject, error) {
	publicTLSServerIPAddresses, err := cryptoutilSharedUtilNetwork.ParseIPAddresses(settings.TLSPublicIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse public TLS server IP addresses: %w", err)
	}

	privateTLSServerIPAddresses, err := cryptoutilSharedUtilNetwork.ParseIPAddresses(settings.TLSPrivateIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private TLS server IP addresses: %w", err)
	}

	public, err := GenerateTLSServerSubjectToDisk(basic, "tls_public_server_", settings.TLSPublicDNSNames, publicTLSServerIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TLS public server certs: %w", err)
	}

	private, err := GenerateTLSServerSubjectToDisk(basic, "tls_private_server_", settings.TLSPrivateDNSNames, privateTLSServerIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TLS private server certs: %w", err)
	}

	return public, private, nil
}

// GenerateTLSServerSubjectToDisk generates a single TLS server certificate subject in memory,
// then writes the certificate chain PEMs and encrypted private key PEM to disk.
// prefix is used to name the output files (e.g., "tls_public_server_").
func GenerateTLSServerSubjectToDisk(basic *Basic, prefix string, dnsNames []string, ipAddresses []net.IP) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	tlsServerEndEntitySubject, err := GenerateTLSServerSubjectInMemory(basic, dnsNames, ipAddresses)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS server subject: %w", err)
	}

	// Encode certificates as PEM and write to files.
	tlsServerCertificateChainPEMs, err := cryptoutilSharedCryptoAsn1.PEMEncodes(tlsServerEndEntitySubject.KeyMaterial.CertificateChain)
	if err != nil {
		return nil, fmt.Errorf("failed to encode certificate chain as PEM: %w", err)
	}

	for i, certPEM := range tlsServerCertificateChainPEMs {
		filename := fmt.Sprintf("%scertificate_%d.pem", prefix, i)
		if err := os.WriteFile(filename, certPEM, cryptoutilSharedMagic.FilePermOwnerReadWriteOnly); err != nil {
			return nil, fmt.Errorf("failed to write TLS server certificate PEM file %s: %w", filename, err)
		}
	}

	// Encrypt private key as PEM and write to file.
	tlsPrivateKeyPEM, err := cryptoutilSharedCryptoAsn1.PEMEncode(tlsServerEndEntitySubject.KeyMaterial.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode private key as PEM: %w", err)
	}

	encryptedTLSPrivateKeyPEM, err := basic.UnsealKeysService.EncryptData(tlsPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt TLS server private key PEM: %w", err)
	}

	if err := os.WriteFile(fmt.Sprintf("%sprivate_key.pem", prefix), encryptedTLSPrivateKeyPEM, cryptoutilSharedMagic.FilePermOwnerReadWriteOnly); err != nil {
		return nil, fmt.Errorf("failed to write encrypted TLS server private key PEM file: %w", err)
	}

	return tlsServerEndEntitySubject, nil
}

// GenerateTLSServerSubjectInMemory generates a single TLS server certificate subject in memory.
// No files are written to disk, making it safe for concurrent parallel tests and server startup.
func GenerateTLSServerSubjectInMemory(basic *Basic, dnsNames []string, ipAddresses []net.IP) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	tlsServerSubjectsKeyPairs := basic.JWKGenService.ECDSAP256KeyGenPool.GetMany(cryptoutilSharedMagic.TLSServerKeyPairsNeeded)

	tlsServerCASubjects, err := cryptoutilSharedCryptoCertificate.CreateCASubjects(tlsServerSubjectsKeyPairs[1:], "TLS Server CA", cryptoutilSharedMagic.TLSDefaultValidityCACertYears*cryptoutilSharedMagic.Days365)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS server CA subjects: %w", err)
	}

	tlsServerEndEntitySubject, err := cryptoutilSharedCryptoCertificate.CreateEndEntitySubject(
		tlsServerCASubjects[0],
		tlsServerSubjectsKeyPairs[0],
		"TLS Server",
		cryptoutilSharedMagic.TLSDefaultValidityEndEntityDaysWithRandomizationBuffer*cryptoutilSharedMagic.Days1,
		dnsNames,
		ipAddresses,
		nil,
		nil,
		x509.KeyUsageDigitalSignature,
		[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS server end entity subject: %w", err)
	}

	return tlsServerEndEntitySubject, nil
}
