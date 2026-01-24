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

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAsn1 "cryptoutil/internal/shared/crypto/asn1"
	cryptoutilCertificate "cryptoutil/internal/shared/crypto/certificate"
	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilDateTime "cryptoutil/internal/shared/util/datetime"
	cryptoutilSharedUtilNetwork "cryptoutil/internal/shared/util/network"
)

const (
	// TLS certificate validity and helper constants.
	tlsCACertValidityYears   = cryptoutilMagic.TLSDefaultValidityCACertYears // years for CA certificates
	tlsEndEntityValidityDays = 396                                           // days for server end-entity certificate (reduced from 397 to allow for randomization)
	tlsServerKeyPairsNeeded  = 2                                             // number of keypairs requested for server TLS

	// File mode for written PEM files.
	defaultPEMFileMode = cryptoutilMagic.FilePermOwnerReadWriteOnly
)

// ServerInit initializes the server by generating TLS certificates and other required configuration.
func ServerInit(settings *cryptoutilConfig.ServiceTemplateServerSettings) error {
	ctx := context.Background()

	serverApplicationBasic, err := StartServerApplicationBasic(ctx, settings)
	if err != nil {
		return fmt.Errorf("failed to initialize server application core: %w", err)
	}
	defer serverApplicationBasic.Shutdown()

	_, _, err = generateTLSServerSubjects(settings, serverApplicationBasic)
	if err != nil {
		return fmt.Errorf("failed to run new function: %w", err)
	}

	return nil
}

func generateTLSServerSubjects(settings *cryptoutilConfig.ServiceTemplateServerSettings, serverApplicationBasic *ServerApplicationBasic) (*cryptoutilCertificate.Subject, *cryptoutilCertificate.Subject, error) {
	publicTLSServerIPAddresses, err := cryptoutilSharedUtilNetwork.ParseIPAddresses(settings.TLSPublicIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse public TLS server IP addresses: %w", err)
	}

	privateTLSServerIPAddresses, err := cryptoutilSharedUtilNetwork.ParseIPAddresses(settings.TLSPrivateIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse private TLS server IP addresses: %w", err)
	}

	public, err := generateTLSServerSubject(serverApplicationBasic, "tls_public_server_", settings.TLSPublicDNSNames, publicTLSServerIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TLS public server certs: %w", err)
	}

	private, err := generateTLSServerSubject(serverApplicationBasic, "tls_private_server_", settings.TLSPrivateDNSNames, privateTLSServerIPAddresses)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create TLS private server certs: %w", err)
	}

	return public, private, nil
}

func generateTLSServerSubject(serverApplicationBasic *ServerApplicationBasic, prefix string, publicTLSServerDNSNames []string, publicTLSServerIPAddresses []net.IP) (*cryptoutilCertificate.Subject, error) {
	tlsServerSubjectsKeyPairs := serverApplicationBasic.JWKGenService.ECDSAP256KeyGenPool.GetMany(tlsServerKeyPairsNeeded)

	tlsServerCASubjects, err := cryptoutilCertificate.CreateCASubjects(tlsServerSubjectsKeyPairs[1:], "TLS Server CA", tlsCACertValidityYears*365*cryptoutilDateTime.Days1)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS server CA subjects: %w", err)
	}

	tlsServerEndEntitySubject, err := cryptoutilCertificate.CreateEndEntitySubject(tlsServerCASubjects[0], tlsServerSubjectsKeyPairs[0], "TLS Server", tlsEndEntityValidityDays*cryptoutilDateTime.Days1, publicTLSServerDNSNames, publicTLSServerIPAddresses, nil, nil, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth})
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS server end entity subject: %w", err)
	}

	// Encode Certificates as PEM and write to files
	tlsServerCertificateChainPEMs, err := cryptoutilAsn1.PEMEncodes(tlsServerEndEntitySubject.KeyMaterial.CertificateChain)
	if err != nil {
		return nil, fmt.Errorf("failed to encode certificate chain as PEM: %w", err)
	}

	for i, certPEM := range tlsServerCertificateChainPEMs {
		filename := fmt.Sprintf("%scertificate_%d.pem", prefix, i)
		if err := os.WriteFile(filename, certPEM, defaultPEMFileMode); err != nil {
			return nil, fmt.Errorf("failed to write TLS server certificate PEM file %s: %w", filename, err)
		}
	}

	// Encrypt private key as PEM to write to file
	tlsPrivateKeyPEM, err := cryptoutilAsn1.PEMEncode(tlsServerEndEntitySubject.KeyMaterial.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode private key as PEM: %w", err)
	}

	encryptedTLSPrivateKeyPEM, err := serverApplicationBasic.UnsealKeysService.EncryptData(tlsPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt TLS server private key PEM: %w", err)
	}

	err = os.WriteFile(fmt.Sprintf("%sprivate_key.pem", prefix), encryptedTLSPrivateKeyPEM, defaultPEMFileMode)
	if err != nil {
		return nil, fmt.Errorf("failed to write encrypted TLS server private key PEM file: %w", err)
	}

	return tlsServerEndEntitySubject, nil
}
