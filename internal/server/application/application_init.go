package application

import (
	"context"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilAsn1 "cryptoutil/internal/common/crypto/asn1"
	cryptoutilCertificate "cryptoutil/internal/common/crypto/certificate"
	cryptoutilDateTime "cryptoutil/internal/common/util/datetime"
)

func ServerInit(settings *cryptoutilConfig.Settings) error {
	ctx := context.Background()

	publicTLSServerIPAddresses, err := ParseIPAddresses(settings.TLSPublicIPAddresses)
	if err != nil {
		return fmt.Errorf("failed to parse public TLS server IP addresses: %w", err)
	}
	privateTLSServerIPAddresses, err := ParseIPAddresses(settings.TLSPrivateIPAddresses)
	if err != nil {
		return fmt.Errorf("failed to parse private TLS server IP addresses: %w", err)
	}

	serverApplicationBasic, err := StartServerApplicationBasic(ctx, settings)
	if err != nil {
		return fmt.Errorf("failed to initialize server application core: %w", err)
	}
	defer serverApplicationBasic.Shutdown()

	err = generateTLSServerCertificates(serverApplicationBasic, "tls_public_server_", settings.TLSPublicDNSNames, publicTLSServerIPAddresses)
	if err != nil {
		return fmt.Errorf("failed to create TLS public server certs: %w", err)
	}

	err = generateTLSServerCertificates(serverApplicationBasic, "tls_private_server_", settings.TLSPrivateDNSNames, privateTLSServerIPAddresses)
	if err != nil {
		return fmt.Errorf("failed to create TLS private server certs: %w", err)
	}

	return nil
}

func generateTLSServerCertificates(serverApplicationBasic *ServerApplicationBasic, prefix string, publicTLSServerDNSNames []string, publicTLSServerIPAddresses []net.IP) error {
	tlsServerSubjectsKeyPairs := serverApplicationBasic.JwkGenService.ECDSAP256KeyGenPool.GetMany(2)

	tlsServerCASubjects, err := cryptoutilCertificate.CreateCASubjects(tlsServerSubjectsKeyPairs[1:], "TLS Server CA", 10*365*cryptoutilDateTime.Days1)
	if err != nil {
		return fmt.Errorf("failed to create TLS server CA subjects: %w", err)
	}

	tlsServerEndEntitySubject, err := cryptoutilCertificate.CreateEndEntitySubject(tlsServerCASubjects[0], tlsServerSubjectsKeyPairs[0], "TLS Server", 397*cryptoutilDateTime.Days1, publicTLSServerDNSNames, publicTLSServerIPAddresses, nil, nil, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth})
	if err != nil {
		return fmt.Errorf("failed to create TLS server end entity subject: %w", err)
	}

	// Encode Certificates as PEM and write to files
	tlsServerCertificateChainPEMs, err := cryptoutilAsn1.PemEncodes(tlsServerEndEntitySubject.KeyMaterial.CertificateChain)
	if err != nil {
		return fmt.Errorf("failed to encode certificate chain as PEM: %w", err)
	}
	for i, certPEM := range tlsServerCertificateChainPEMs {
		filename := fmt.Sprintf("%scertificate_%d.pem", prefix, i)
		if err := os.WriteFile(filename, certPEM, 0600); err != nil {
			return fmt.Errorf("failed to write TLS server certificate PEM file %s: %w", filename, err)
		}
	}

	// Encrypt private key as PEM to write to file
	tlsPrivateKeyPEM, err := cryptoutilAsn1.PemEncode(tlsServerEndEntitySubject.KeyMaterial.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to encode private key as PEM: %w", err)
	}
	encryptedTLSPrivateKeyPEM, err := serverApplicationBasic.UnsealKeysService.EncryptData(tlsPrivateKeyPEM)
	if err != nil {
		return fmt.Errorf("failed to encrypt TLS server private key PEM: %w", err)
	}
	err = os.WriteFile(fmt.Sprintf("%sprivate_key.pem", prefix), encryptedTLSPrivateKeyPEM, 0600)
	if err != nil {
		return fmt.Errorf("failed to write encrypted TLS server private key PEM file: %w", err)
	}
	return nil
}

func ParseIPAddresses(ipAddresses []string) ([]net.IP, error) {
	var parsedIPs []net.IP
	for _, ip := range ipAddresses {
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			return nil, fmt.Errorf("failed to parse IP address: %s", ip)
		}
		parsedIPs = append(parsedIPs, parsedIP)
	}
	return parsedIPs, nil
}
