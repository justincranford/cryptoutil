package certificate

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"time"
)

func CertificateTemplateCA(signatureAlgorithm x509.SignatureAlgorithm, issuerName string, subjectName string, duration time.Duration, maxPathLen int) (*x509.Certificate, error) {
	serialNumber, err := GenerateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number for TLS root CA: %w", err)
	}
	notBefore, notAfter, err := randomizedNotBeforeNotAfterCA(time.Now().UTC(), duration, 1*time.Minute, 120*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate validity period for TLS root CA: %w", err)
	}
	return &x509.Certificate{
		SignatureAlgorithm:    signatureAlgorithm,
		Issuer:                pkix.Name{CommonName: issuerName},
		Subject:               pkix.Name{CommonName: subjectName},
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageTimeStamping, x509.ExtKeyUsageOCSPSigning},
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            maxPathLen,
		MaxPathLenZero:        maxPathLen == 0,
	}, nil
}

func CertificateTemplateTLSServer(signatureAlgorithm x509.SignatureAlgorithm, issuerName string, subjectName string, duration time.Duration, dnsNames []string, ipAddresses []net.IP, emailAddresses []string, uris []*url.URL) (*x509.Certificate, error) {
	serialNumber, err := GenerateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number for TLS server: %w", err)
	}
	notBefore, notAfter, err := randomizedNotBeforeNotAfterEndEntity(time.Now().UTC(), duration, 1*time.Minute, 120*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate validity period for TLS server: %w", err)
	}
	template := &x509.Certificate{
		SignatureAlgorithm: signatureAlgorithm,
		Issuer:             pkix.Name{CommonName: issuerName},
		Subject:            pkix.Name{CommonName: subjectName},
		SerialNumber:       serialNumber,
		NotBefore:          notBefore,
		NotAfter:           notAfter,
		KeyUsage:           x509.KeyUsageDigitalSignature,
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:           dnsNames,
		EmailAddresses:     emailAddresses,
		IPAddresses:        ipAddresses,
		URIs:               uris,
	}
	return template, nil
}

func CertificateTemplateTLSClient(signatureAlgorithm x509.SignatureAlgorithm, issuerName string, subjectName string, duration time.Duration, dnsNames []string, ipAddresses []net.IP, emailAddresses []string, uris []*url.URL) (*x509.Certificate, error) {
	serialNumber, err := GenerateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number for TLS client: %w", err)
	}
	notBefore, notAfter, err := randomizedNotBeforeNotAfterEndEntity(time.Now().UTC(), duration, 1*time.Minute, 120*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate validity period for TLS client: %w", err)
	}
	template := &x509.Certificate{
		SignatureAlgorithm: signatureAlgorithm,
		Issuer:             pkix.Name{CommonName: issuerName},
		Subject:            pkix.Name{CommonName: subjectName},
		SerialNumber:       serialNumber,
		NotBefore:          notBefore,
		NotAfter:           notAfter,
		KeyUsage:           x509.KeyUsageDigitalSignature,
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		DNSNames:           dnsNames,
		EmailAddresses:     emailAddresses,
		IPAddresses:        ipAddresses,
		URIs:               uris,
	}
	return template, nil
}

func SignCertificate(issuerCert *x509.Certificate, issuerPrivateKey crypto.Signer, subjectCert *x509.Certificate, subjectPublicKey crypto.PublicKey) (*x509.Certificate, []byte, []byte, error) {
	var err error
	var certificateDer []byte
	if issuerCert == nil {
		certificateDer, err = x509.CreateCertificate(rand.Reader, subjectCert, subjectCert, subjectPublicKey, issuerPrivateKey)
	} else {
		certificateDer, err = x509.CreateCertificate(rand.Reader, subjectCert, issuerCert, subjectPublicKey, issuerPrivateKey)
	}
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}
	certificate, err := x509.ParseCertificate(certificateDer)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}
	certificatePemBlock := &pem.Block{Type: "CERTIFICATE", Bytes: certificateDer}
	certificatePem := pem.EncodeToMemory(certificatePemBlock)

	return certificate, certificateDer, certificatePem, nil
}
