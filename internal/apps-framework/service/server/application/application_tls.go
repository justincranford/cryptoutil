// Copyright (c) 2025-2026 Justin Cranford.
//
//

package application

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps-framework/service/config"
	cryptoutilSharedCryptoAsn1 "cryptoutil/internal/shared/crypto/asn1"
	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
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
	return startTLSListener(settings, startTLSListenerOperations{
		startCoreFn:   StartCore,
		startBasicFn:  StartBasic,
		generateTLSFn: GenerateTLSServerSubjectsInMemory,
	})
}

type startTLSListenerOperations struct {
	startCoreFn   func(context.Context, *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*Core, error)
	startBasicFn  func(context.Context, *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) (*Basic, error)
	generateTLSFn func(*cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings, *Basic) (*cryptoutilSharedCryptoCertificate.Subject, *cryptoutilSharedCryptoCertificate.Subject, error)
}

func startTLSListener(
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	operations startTLSListenerOperations,
) (*TLSListener, error) {
	if settings == nil {
		return nil, errors.New(valueSettingsCannotBeNil)
	}

	ctx := context.Background()

	// Initialize core infrastructure including database connectivity.
	// Fails for unavailable databases (e.g., PostgreSQL not running in the test environment).
	core, err := operations.startCoreFn(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to start application core: %w", err)
	}

	// Initialize basic services (telemetry, unseal keys, JWK generation) for TLS cert generation.
	basic, err := operations.startBasicFn(ctx, settings)
	if err != nil {
		core.Shutdown()

		return nil, fmt.Errorf("failed to start basic application services: %w", err)
	}

	// Generate TLS certificate subjects in memory (no disk I/O, safe for parallel tests).
	publicSubject, privateSubject, err := operations.generateTLSFn(settings, basic)
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
	return generateTLSServerSubjectsInMemory(settings, basic, func(basic *Basic, dnsNames []string, ipAddresses []net.IP) (*cryptoutilSharedCryptoCertificate.Subject, error) {
		return GenerateTLSServerSubjectInMemory(basic, dnsNames, ipAddresses)
	})
}

func generateTLSServerSubjectsInMemory(
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	basic *Basic,
	generateSubjectFn func(*Basic, []string, []net.IP) (*cryptoutilSharedCryptoCertificate.Subject, error),
) (*cryptoutilSharedCryptoCertificate.Subject, *cryptoutilSharedCryptoCertificate.Subject, error) {
	publicTLSServerIPAddresses, err := cryptoutilSharedUtilNetwork.ParseIPAddresses(settings.TLSPublicIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse public TLS server IP addresses: %w", err)
	}

	privateTLSServerIPAddresses, err := cryptoutilSharedUtilNetwork.ParseIPAddresses(settings.TLSPrivateIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private TLS server IP addresses: %w", err)
	}

	public, err := generateSubjectFn(basic, settings.TLSPublicDNSNames, publicTLSServerIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TLS public server certs in memory: %w", err)
	}

	private, err := generateSubjectFn(basic, settings.TLSPrivateDNSNames, privateTLSServerIPAddresses)
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
	return generateTLSServerSubjectsToDisk(settings, basic, GenerateTLSServerSubjectToDisk)
}

func generateTLSServerSubjectsToDisk(
	settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings,
	basic *Basic,
	generateSubjectFn func(*Basic, string, []string, []net.IP) (*cryptoutilSharedCryptoCertificate.Subject, error),
) (*cryptoutilSharedCryptoCertificate.Subject, *cryptoutilSharedCryptoCertificate.Subject, error) {
	publicTLSServerIPAddresses, err := cryptoutilSharedUtilNetwork.ParseIPAddresses(settings.TLSPublicIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse public TLS server IP addresses: %w", err)
	}

	privateTLSServerIPAddresses, err := cryptoutilSharedUtilNetwork.ParseIPAddresses(settings.TLSPrivateIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private TLS server IP addresses: %w", err)
	}

	public, err := generateSubjectFn(basic, "tls_public_server_", settings.TLSPublicDNSNames, publicTLSServerIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TLS public server certs: %w", err)
	}

	private, err := generateSubjectFn(basic, "tls_private_server_", settings.TLSPrivateDNSNames, privateTLSServerIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TLS private server certs: %w", err)
	}

	return public, private, nil
}

// GenerateTLSServerSubjectToDisk generates a single TLS server certificate subject in memory,
// then writes the certificate chain PEMs and encrypted private key PEM to disk.
// prefix is used to name the output files (e.g., "tls_public_server_").
func GenerateTLSServerSubjectToDisk(basic *Basic, prefix string, dnsNames []string, ipAddresses []net.IP) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	return generateTLSServerSubjectToDisk(basic, prefix, dnsNames, ipAddresses, tlsDiskOperations{
		generateSubjectFn: GenerateTLSServerSubjectInMemory,
		pemEncodesFn:      cryptoutilSharedCryptoAsn1.PEMEncodes,
		writeFileFn:       os.WriteFile,
		pemEncodeFn:       cryptoutilSharedCryptoAsn1.PEMEncode,
		encryptFn:         basic.UnsealKeysService.EncryptData,
	})
}

type tlsDiskOperations struct {
	generateSubjectFn func(*Basic, []string, []net.IP) (*cryptoutilSharedCryptoCertificate.Subject, error)
	pemEncodesFn      func(any) ([][]byte, error)
	writeFileFn       func(string, []byte, os.FileMode) error
	pemEncodeFn       func(any) ([]byte, error)
	encryptFn         func([]byte) ([]byte, error)
}

func generateTLSServerSubjectToDisk(
	basic *Basic,
	prefix string,
	dnsNames []string,
	ipAddresses []net.IP,
	operations tlsDiskOperations,
) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	tlsServerEndEntitySubject, err := operations.generateSubjectFn(basic, dnsNames, ipAddresses)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS server subject: %w", err)
	}

	// Encode certificates as PEM and write to files.
	tlsServerCertificateChainPEMs, err := operations.pemEncodesFn(tlsServerEndEntitySubject.KeyMaterial.CertificateChain)
	if err != nil {
		return nil, fmt.Errorf("failed to encode certificate chain as PEM: %w", err)
	}

	for i, certPEM := range tlsServerCertificateChainPEMs {
		filename := fmt.Sprintf("%scertificate_%d.pem", prefix, i)
		if err := operations.writeFileFn(filename, certPEM, cryptoutilSharedMagic.FilePermOwnerReadWriteOnly); err != nil {
			return nil, fmt.Errorf("failed to write TLS server certificate PEM file %s: %w", filename, err)
		}
	}

	// Encrypt private key as PEM and write to file.
	tlsPrivateKeyPEM, err := operations.pemEncodeFn(tlsServerEndEntitySubject.KeyMaterial.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode private key as PEM: %w", err)
	}

	encryptedTLSPrivateKeyPEM, err := operations.encryptFn(tlsPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt TLS server private key PEM: %w", err)
	}

	if err := operations.writeFileFn(fmt.Sprintf("%sprivate_key.pem", prefix), encryptedTLSPrivateKeyPEM, cryptoutilSharedMagic.FilePermOwnerReadWriteOnly); err != nil {
		return nil, fmt.Errorf("failed to write encrypted TLS server private key PEM file: %w", err)
	}

	return tlsServerEndEntitySubject, nil
}

// GenerateTLSServerSubjectInMemory generates a single TLS server certificate subject in memory.
// No files are written to disk, making it safe for concurrent parallel tests and server startup.
func GenerateTLSServerSubjectInMemory(basic *Basic, dnsNames []string, ipAddresses []net.IP) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	return generateTLSServerSubjectInMemory(basic, dnsNames, ipAddresses, tlsSubjectOperations{
		createCASubjectsFn:       cryptoutilSharedCryptoCertificate.CreateCASubjects,
		createEndEntitySubjectFn: cryptoutilSharedCryptoCertificate.CreateEndEntitySubject,
	})
}

type tlsSubjectOperations struct {
	createCASubjectsFn       func([]*cryptoutilSharedCryptoKeygen.KeyPair, string, time.Duration) ([]*cryptoutilSharedCryptoCertificate.Subject, error)
	createEndEntitySubjectFn func(
		*cryptoutilSharedCryptoCertificate.Subject,
		*cryptoutilSharedCryptoKeygen.KeyPair,
		string,
		time.Duration,
		[]string,
		[]net.IP,
		[]string,
		[]*url.URL,
		x509.KeyUsage,
		[]x509.ExtKeyUsage,
	) (*cryptoutilSharedCryptoCertificate.Subject, error)
}

func generateTLSServerSubjectInMemory(
	basic *Basic,
	dnsNames []string,
	ipAddresses []net.IP,
	operations tlsSubjectOperations,
) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	tlsServerSubjectsKeyPairs := basic.JWKGenService.ECDSAP256KeyGenPool.GetMany(cryptoutilSharedMagic.TLSServerKeyPairsNeeded)

	tlsServerCASubjects, err := operations.createCASubjectsFn(tlsServerSubjectsKeyPairs[1:], "TLS Server CA", cryptoutilSharedMagic.TLSDefaultValidityCACertYears*cryptoutilSharedMagic.Days365)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS server CA subjects: %w", err)
	}

	tlsServerEndEntitySubject, err := operations.createEndEntitySubjectFn(
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
