// Copyright (c) 2025 Justin Cranford
//
//

package application

import (
	"context"
	"fmt"
	"net"
	"os"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsFrameworkServiceServerApplication "cryptoutil/internal/apps/framework/service/server/application"
	cryptoutilSharedCryptoAsn1 "cryptoutil/internal/shared/crypto/asn1"
	cryptoutilSharedCryptoCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilNetwork "cryptoutil/internal/shared/util/network"
)

// ServerInit initializes the server by generating TLS certificates and other required configuration.
func ServerInit(settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings) error {
	ctx := context.Background()

	basic, err := cryptoutilAppsFrameworkServiceServerApplication.StartBasic(ctx, settings)
	if err != nil {
		return fmt.Errorf("failed to initialize server application core: %w", err)
	}
	defer basic.Shutdown()

	_, _, err = generateTLSServerSubjects(settings, basic)
	if err != nil {
		return fmt.Errorf("failed to generate TLS server subjects: %w", err)
	}

	return nil
}

func generateTLSServerSubjects(settings *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings, basic *cryptoutilAppsFrameworkServiceServerApplication.Basic) (*cryptoutilSharedCryptoCertificate.Subject, *cryptoutilSharedCryptoCertificate.Subject, error) {
	publicTLSServerIPAddresses, err := cryptoutilSharedUtilNetwork.ParseIPAddresses(settings.TLSPublicIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse public TLS server IP addresses: %w", err)
	}

	privateTLSServerIPAddresses, err := cryptoutilSharedUtilNetwork.ParseIPAddresses(settings.TLSPrivateIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private TLS server IP addresses: %w", err)
	}

	public, err := generateTLSServerSubject(basic, "tls_public_server_", settings.TLSPublicDNSNames, publicTLSServerIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TLS public server certs: %w", err)
	}

	private, err := generateTLSServerSubject(basic, "tls_private_server_", settings.TLSPrivateDNSNames, privateTLSServerIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TLS private server certs: %w", err)
	}

	return public, private, nil
}

func generateTLSServerSubject(basic *cryptoutilAppsFrameworkServiceServerApplication.Basic, prefix string, publicTLSServerDNSNames []string, publicTLSServerIPAddresses []net.IP) (*cryptoutilSharedCryptoCertificate.Subject, error) {
	// Generate the TLS server subject in memory using the shared framework function.
	tlsServerEndEntitySubject, err := cryptoutilAppsFrameworkServiceServerApplication.GenerateTLSServerSubjectInMemory(basic, publicTLSServerDNSNames, publicTLSServerIPAddresses)
	if err != nil {
		return nil, fmt.Errorf("failed to generate TLS server subject: %w", err)
	}

	// Encode Certificates as PEM and write to files
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

	// Encrypt private key as PEM to write to file
	tlsPrivateKeyPEM, err := cryptoutilSharedCryptoAsn1.PEMEncode(tlsServerEndEntitySubject.KeyMaterial.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode private key as PEM: %w", err)
	}

	encryptedTLSPrivateKeyPEM, err := basic.UnsealKeysService.EncryptData(tlsPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt TLS server private key PEM: %w", err)
	}

	err = os.WriteFile(fmt.Sprintf("%sprivate_key.pem", prefix), encryptedTLSPrivateKeyPEM, cryptoutilSharedMagic.FilePermOwnerReadWriteOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to write encrypted TLS server private key PEM file: %w", err)
	}

	return tlsServerEndEntitySubject, nil
}
