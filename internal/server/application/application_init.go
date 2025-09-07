package application

import (
	"context"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	cryptoutilConfig "cryptoutil/internal/common/config"
	"cryptoutil/internal/common/crypto/asn1"
	cryptoutilCertificate "cryptoutil/internal/common/crypto/certificate"
	cryptoutilDateTime "cryptoutil/internal/common/util/datetime"
)

func ServerInit(settings *cryptoutilConfig.Settings) error {
	ctx := context.Background()

	publicTLSServerIPAddresses, err := ParseIPAddresses(settings.TLSPublicIPAddresses)
	if err != nil {
		return fmt.Errorf("failed to parse TLS server IP addresses: %w", err)
	}

	serverApplicationBasic, err := StartServerApplicationBasic(ctx, settings)
	if err != nil {
		return fmt.Errorf("failed to initialize server application core: %w", err)
	}
	defer serverApplicationBasic.Shutdown()

	err = generateTLSCertificates(serverApplicationBasic, settings, publicTLSServerIPAddresses)
	if err != nil {
		return fmt.Errorf("failed to create TLS server certs: %w", err)
	}

	return nil
}

// TODO private TLS server cert
func generateTLSCertificates(serverApplicationBasic *ServerApplicationBasic, settings *cryptoutilConfig.Settings, publicTLSServerIPAddresses []net.IP) error {
	publicTLSServerSubjectsKeyPairs := serverApplicationBasic.JwkGenService.ECDSAP256KeyGenPool.GetMany(2)
	publicTLSServerCASubjects, err := cryptoutilCertificate.CreateCASubjects(publicTLSServerSubjectsKeyPairs[1:], "TLS Server CA", 10*365*cryptoutilDateTime.Days1)
	if err != nil {
		return fmt.Errorf("failed to create TLS server CA subjects: %w", err)
	}
	publicTLSServerEndEntitySubject, err := cryptoutilCertificate.CreateEndEntitySubject(publicTLSServerCASubjects[0], publicTLSServerSubjectsKeyPairs[0], "TLS Server", 397*cryptoutilDateTime.Days1, settings.TLSPublicDNSNames, publicTLSServerIPAddresses, nil, nil, x509.KeyUsageDigitalSignature, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth})
	if err != nil {
		return fmt.Errorf("failed to create TLS server end entity subject: %w", err)
	} else if publicTLSServerEndEntitySubject == nil {
		return fmt.Errorf("publicTLSServerEndEntitySubject is nil")
	}

	// Encode Cert Chain and Private Key as PEM
	publicTLSServerCertificateChainPEMs, err := asn1.PemEncodes(publicTLSServerEndEntitySubject.KeyMaterial.CertificateChain)
	if err != nil {
		return fmt.Errorf("failed to encode certificate chain to PEM: %w", err)
	}
	publicTLSPrivateKeyPEM, err := asn1.PemEncode(publicTLSServerEndEntitySubject.KeyMaterial.PrivateKey)
	if err != nil {
		return fmt.Errorf("failed to encode private key to PEM: %w", err)
	}

	// Write Cert Chain and Private Key PEM files to files
	for i, certPEM := range publicTLSServerCertificateChainPEMs {
		filename := fmt.Sprintf("tls_server_cert_chain_%d.pem", i)
		if err := os.WriteFile(filename, certPEM, 0600); err != nil {
			return fmt.Errorf("failed to write public TLS server certificate chain PEM file %s: %w", filename, err)
		}
	}
	if err := os.WriteFile("tls_server_private_key.pem", publicTLSPrivateKeyPEM, 0600); err != nil {
		return fmt.Errorf("failed to write public TLS server private key PEM file: %w", err)
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
