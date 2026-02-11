// Copyright (c) 2025 Justin Cranford
//
//

// Package certificate provides X.509 certificate management utilities.
package certificate

import (
	"crypto"
	crand "crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	json "encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"time"

	cryptoutilSharedCryptoKeygen "cryptoutil/internal/shared/crypto/keygen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// KeyMaterial holds the cryptographic components for a certificate.
type KeyMaterial struct {
	CertificateChain []*x509.Certificate
	PublicKey        crypto.PublicKey
	PrivateKey       crypto.PrivateKey
}

// KeyMaterialEncoded holds the DER and PEM encodings of a KeyMaterial.
type KeyMaterialEncoded struct {
	DERCertificateChain [][]byte `json:"der_certificate_chain"`
	DERPublicKey        []byte   `json:"der_public_key"`
	DERPrivateKey       []byte   `json:"der_private_key"`

	PEMCertificateChain [][]byte `json:"pem_certificate_chain"`
	PEMPublicKey        []byte   `json:"pem_public_key"`
	PEMPrivateKey       []byte   `json:"pem_private_key"`
}

// Subject represents a certificate subject with associated key material and attributes.
type Subject struct {
	SubjectName string
	IssuerName  string
	Duration    time.Duration
	IsCA        bool

	MaxPathLen int // CA-specific fields (only valid when IsCA=true)

	DNSNames       []string   // End entity-specific fields (only valid when IsCA=false)
	IPAddresses    []net.IP   // End entity-specific fields (only valid when IsCA=false)
	EmailAddresses []string   // End entity-specific fields (only valid when IsCA=false)
	URIs           []*url.URL // End entity-specific fields (only valid when IsCA=false)

	KeyMaterial KeyMaterial
}

// CreateCASubjects creates a certificate chain from multiple CA key pairs.
func CreateCASubjects(keyPairs []*cryptoutilSharedCryptoKeygen.KeyPair, caSubjectNamePrefix string, duration time.Duration) ([]*Subject, error) {
	subjects := make([]*Subject, len(keyPairs))

	for i := len(keyPairs) - 1; i >= 0; i-- {
		subjectName := fmt.Sprintf("%s %d", caSubjectNamePrefix, len(keyPairs)-1-i)

		var err error
		if i == len(keyPairs)-1 {
			subjects[i], err = CreateCASubject(nil, nil, subjectName, keyPairs[i], duration, i)
		} else {
			subjects[i], err = CreateCASubject(subjects[i+1], subjects[i+1].KeyMaterial.PrivateKey, subjectName, keyPairs[i], duration, i)
			subjects[i+1].KeyMaterial.PrivateKey = nil // pragma: allowlist secret
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create CA subject %d: %w", len(keyPairs)-1-i, err)
		}
	}

	return subjects, nil
}

// CreateCASubject creates a single CA certificate subject with optional issuer.
func CreateCASubject(issuerSubject *Subject, issuerPrivateKey crypto.PrivateKey, subjectName string, subjectKeyPair *cryptoutilSharedCryptoKeygen.KeyPair, duration time.Duration, maxPathLen int) (*Subject, error) {
	if issuerSubject == nil && issuerPrivateKey != nil { // pragma: allowlist secret
		return nil, fmt.Errorf("issuerSubject is nil but issuerPrivateKey is not nil for CA %s", subjectName)
	} else if issuerSubject != nil && issuerPrivateKey == nil { // pragma: allowlist secret
		return nil, fmt.Errorf("issuerSubject is not nil but issuerPrivateKey is nil for CA %s", subjectName)
	} else if len(subjectName) == 0 {
		return nil, fmt.Errorf("subjectName should not be empty for CA %s", subjectName)
	} else if subjectKeyPair == nil {
		return nil, fmt.Errorf("subjectKeyPair should not be nil for CA %s", subjectName)
	} else if subjectKeyPair.Public == nil {
		return nil, fmt.Errorf("subjectKeyPair.Public should not be nil for CA %s", subjectName)
	} else if subjectKeyPair.Private == nil {
		return nil, fmt.Errorf("subjectKeyPair.Private should not be nil for CA %s", subjectName)
	} else if maxPathLen < 0 {
		return nil, fmt.Errorf("maxPathLen should not be negative for CA %s", subjectName)
	}

	var issuerName string

	var issuerCertificateChain []*x509.Certificate

	var issuerCertificate *x509.Certificate

	if issuerPrivateKey == nil || issuerSubject == nil { // pragma: allowlist secret
		issuerName = subjectName
		issuerCertificateChain = []*x509.Certificate{}
		issuerPrivateKey = subjectKeyPair.Private // pragma: allowlist secret
	} else {
		issuerName = issuerSubject.SubjectName
		issuerCertificateChain = issuerSubject.KeyMaterial.CertificateChain
		issuerCertificate = issuerSubject.KeyMaterial.CertificateChain[0]
	}

	currentSubject := Subject{
		SubjectName: subjectName,
		IssuerName:  issuerName,
		Duration:    duration,
		IsCA:        true,
		MaxPathLen:  maxPathLen,
		KeyMaterial: KeyMaterial{
			CertificateChain: []*x509.Certificate{},
			PublicKey:        subjectKeyPair.Public,
			PrivateKey:       subjectKeyPair.Private,
		},
	}

	currentCACertTemplate, err := CertificateTemplateCA(issuerName, subjectName, currentSubject.Duration, maxPathLen)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate template for %s: %w", subjectName, err)
	}

	signedCertificate, _, _, err := SignCertificate(issuerCertificate, issuerPrivateKey, currentCACertTemplate, currentSubject.KeyMaterial.PublicKey, x509.ECDSAWithSHA256)
	if err != nil {
		return nil, fmt.Errorf("failed to sign CA certificate for %s: %w", subjectName, err)
	}

	currentSubject.KeyMaterial.CertificateChain = append([]*x509.Certificate{signedCertificate}, issuerCertificateChain...)

	return &currentSubject, nil
}

// CreateEndEntitySubject creates an end entity certificate subject with SANs and key usage.
func CreateEndEntitySubject(issuingCASubject *Subject, keyPair *cryptoutilSharedCryptoKeygen.KeyPair, subjectName string, duration time.Duration, dnsNames []string, ipAddresses []net.IP, emailAddresses []string, uris []*url.URL, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage) (*Subject, error) {
	endEntityCertTemplate, err := CertificateTemplateEndEntity(issuingCASubject.SubjectName, subjectName, duration, dnsNames, ipAddresses, emailAddresses, uris, keyUsage, extKeyUsage)
	if err != nil {
		return nil, fmt.Errorf("failed to create end entity certificate template for %s: %w", subjectName, err)
	}

	signedCert, _, _, err := SignCertificate(issuingCASubject.KeyMaterial.CertificateChain[0], issuingCASubject.KeyMaterial.PrivateKey, endEntityCertTemplate, keyPair.Public, x509.ECDSAWithSHA256)
	if err != nil {
		return nil, fmt.Errorf("failed to sign end entity certificate for %s: %w", subjectName, err)
	}

	return &Subject{
		SubjectName:    subjectName,
		IssuerName:     issuingCASubject.SubjectName,
		Duration:       duration,
		IsCA:           false,
		DNSNames:       dnsNames,
		IPAddresses:    ipAddresses,
		EmailAddresses: emailAddresses,
		URIs:           uris,
		KeyMaterial: KeyMaterial{
			CertificateChain: append([]*x509.Certificate{signedCert}, issuingCASubject.KeyMaterial.CertificateChain...),
			PublicKey:        keyPair.Public,
			PrivateKey:       keyPair.Private,
		},
	}, nil
}

// BuildTLSCertificate converts an end entity Subject into a tls.Certificate with root and intermediate pools.
func BuildTLSCertificate(endEntitySubject *Subject) (*tls.Certificate, *x509.CertPool, *x509.CertPool, error) {
	if len(endEntitySubject.KeyMaterial.CertificateChain) == 0 {
		return nil, nil, nil, fmt.Errorf("certificate chain is empty")
	} else if endEntitySubject.KeyMaterial.PrivateKey == nil { // pragma: allowlist secret
		return nil, nil, nil, fmt.Errorf("private key is nil")
	}

	derCertChain := make([][]byte, len(endEntitySubject.KeyMaterial.CertificateChain))
	for i, certificate := range endEntitySubject.KeyMaterial.CertificateChain {
		derCertChain[i] = certificate.Raw
	}

	rootCACertsPool := x509.NewCertPool()

	if len(endEntitySubject.KeyMaterial.CertificateChain) > 0 {
		rootCert := endEntitySubject.KeyMaterial.CertificateChain[len(endEntitySubject.KeyMaterial.CertificateChain)-1]
		rootCACertsPool.AddCert(rootCert)
	}

	intermediateCertsPool := x509.NewCertPool()
	for j := 1; j < len(endEntitySubject.KeyMaterial.CertificateChain)-1; j++ {
		intermediateCertsPool.AddCert(endEntitySubject.KeyMaterial.CertificateChain[j])
	}

	return &tls.Certificate{Certificate: derCertChain, PrivateKey: endEntitySubject.KeyMaterial.PrivateKey, Leaf: endEntitySubject.KeyMaterial.CertificateChain[0]}, rootCACertsPool, intermediateCertsPool, nil
}

// CertificateTemplateCA creates an x509 certificate template for a CA with specified path length.
func CertificateTemplateCA(issuerName, subjectName string, duration time.Duration, maxPathLen int) (*x509.Certificate, error) {
	serialNumber, err := GenerateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number for TLS root CA: %w", err)
	}

	notBefore, notAfter, err := randomizedNotBeforeNotAfterCA(time.Now().UTC(), duration, 1*time.Minute, cryptoutilSharedMagic.CertificateRandomizationNotBeforeMinutes*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate validity period for TLS root CA: %w", err)
	}

	return &x509.Certificate{
		Issuer:                pkix.Name{CommonName: issuerName},
		Subject:               pkix.Name{CommonName: subjectName},
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            maxPathLen,
		MaxPathLenZero:        maxPathLen == 0,
	}, nil
}

// CertificateTemplateEndEntity creates an x509 certificate template for an end entity with SANs and key usage.
func CertificateTemplateEndEntity(issuerName, subjectName string, duration time.Duration, dnsNames []string, ipAddresses []net.IP, emailAddresses []string, uris []*url.URL, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage) (*x509.Certificate, error) {
	serialNumber, err := GenerateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number for TLS server: %w", err)
	}

	notBefore, notAfter, err := randomizedNotBeforeNotAfterEndEntity(time.Now().UTC(), duration, 1*time.Minute, cryptoutilSharedMagic.CertificateRandomizationNotBeforeMinutes*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate certificate validity period for TLS server: %w", err)
	}

	template := &x509.Certificate{
		Issuer:         pkix.Name{CommonName: issuerName},
		Subject:        pkix.Name{CommonName: subjectName},
		SerialNumber:   serialNumber,
		NotBefore:      notBefore,
		NotAfter:       notAfter,
		KeyUsage:       keyUsage,
		ExtKeyUsage:    extKeyUsage,
		DNSNames:       dnsNames,
		EmailAddresses: emailAddresses,
		IPAddresses:    ipAddresses,
		URIs:           uris,
	}

	return template, nil
}

// SignCertificate signs a subject certificate using an issuer certificate and private key.
func SignCertificate(issuerCertificate *x509.Certificate, issuerPrivateKey crypto.PrivateKey, subjectCertificate *x509.Certificate, subjectPublicKey crypto.PublicKey, signatureAlgorithm x509.SignatureAlgorithm) (*x509.Certificate, []byte, []byte, error) {
	_, ok := issuerPrivateKey.(crypto.Signer)
	if !ok {
		return nil, nil, nil, fmt.Errorf("issuer private key is not a crypto.Signer")
	}

	subjectCertificate.SignatureAlgorithm = signatureAlgorithm

	var err error

	var certificateDER []byte
	if issuerCertificate == nil {
		certificateDER, err = x509.CreateCertificate(crand.Reader, subjectCertificate, subjectCertificate, subjectPublicKey, issuerPrivateKey)
	} else {
		certificateDER, err = x509.CreateCertificate(crand.Reader, subjectCertificate, issuerCertificate, subjectPublicKey, issuerPrivateKey)
	}

	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	certificate, err := x509.ParseCertificate(certificateDER)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return certificate, certificateDER, toCertificatePEM(certificateDER), nil
}

// SerializeSubjects converts Subject structs to JSON-encoded byte slices with optional private key inclusion.
func SerializeSubjects(subjects []*Subject, includePrivateKey bool) ([][]byte, error) {
	if subjects == nil {
		return nil, fmt.Errorf("subjects cannot be nil")
	}

	keyMaterialEncodedsBytes := make([][]byte, len(subjects))

	for i, subject := range subjects {
		if subject.SubjectName == "" {
			return nil, fmt.Errorf("subject at index %d has empty SubjectName", i)
		} else if subject.IssuerName == "" {
			return nil, fmt.Errorf("subject at index %d has empty IssuerName", i)
		} else if subject.Duration <= 0 {
			return nil, fmt.Errorf("subject at index %d has zero or negative Duration", i)
		}

		if subject.IsCA && (len(subject.DNSNames) > 0 || len(subject.IPAddresses) > 0 || len(subject.EmailAddresses) > 0 || len(subject.URIs) > 0) {
			return nil, fmt.Errorf("subject at index %d is a CA but has end-entity fields (DNSNames, IPAddresses, EmailAddresses, or URIs) populated", i)
		} else if !subject.IsCA && subject.MaxPathLen > 0 {
			return nil, fmt.Errorf("subject at index %d is not a CA but has MaxPathLen populated", i)
		} else if subject.IsCA {
			if subject.MaxPathLen < 0 {
				return nil, fmt.Errorf("subject at index %d has invalid MaxPathLen (%d), must be >= 0", i, subject.MaxPathLen)
			}
		}

		keyMaterialEncoded, err := serializeKeyMaterial(&subject.KeyMaterial, includePrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert KeyMaterial to JSON format for subject %d: %w", i, err)
		}

		keyMaterialEncodedBytes, err := json.Marshal(keyMaterialEncoded)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize KeyMaterialEncoded for subject %d: %w", i, err)
		}

		keyMaterialEncodedsBytes[i] = keyMaterialEncodedBytes
	}

	return keyMaterialEncodedsBytes, nil
}

// DeserializeSubjects reconstructs Subject structs from JSON-encoded byte slices.
func DeserializeSubjects(keyMaterialEncodedBytesList [][]byte) ([]*Subject, error) {
	subjects := make([]*Subject, len(keyMaterialEncodedBytesList))

	for i, keyMaterialEncodedBytes := range keyMaterialEncodedBytesList {
		var keyMaterialEncoded KeyMaterialEncoded

		err := json.Unmarshal(keyMaterialEncodedBytes, &keyMaterialEncoded)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize KeyMaterialEncoded for item %d: %w", i, err)
		}

		keyMaterial, err := deserializeKeyMaterial(&keyMaterialEncoded)
		if err != nil {
			return nil, fmt.Errorf("failed to convert KeyMaterialEncoded to KeyMaterial for item %d: %w", i, err)
		}

		if len(keyMaterial.CertificateChain) == 0 {
			return nil, fmt.Errorf("certChain is empty for item %d", i)
		} else if keyMaterial.PublicKey == nil {
			return nil, fmt.Errorf("publicKey is nil for item %d", i)
		}

		certificate := keyMaterial.CertificateChain[0]
		subject := Subject{
			KeyMaterial: *keyMaterial,
			SubjectName: certificate.Subject.CommonName,
			IssuerName:  certificate.Issuer.CommonName,
			Duration:    certificate.NotAfter.Sub(certificate.NotBefore),
			IsCA:        certificate.IsCA,
		}

		if certificate.IsCA {
			subject.MaxPathLen = certificate.MaxPathLen
		} else {
			subject.DNSNames = certificate.DNSNames
			subject.IPAddresses = certificate.IPAddresses
			subject.EmailAddresses = certificate.EmailAddresses
			subject.URIs = certificate.URIs
		}

		subjects[i] = &subject
	}

	return subjects, nil
}

func serializeKeyMaterial(keyMaterial *KeyMaterial, includePrivateKey bool) (*KeyMaterialEncoded, error) {
	if keyMaterial == nil {
		return nil, fmt.Errorf("keyMaterial cannot be nil")
	} else if len(keyMaterial.CertificateChain) == 0 {
		return nil, fmt.Errorf("certificate chain cannot be empty")
	} else if keyMaterial.PublicKey == nil {
		return nil, fmt.Errorf("PublicKey cannot be nil")
	}

	for i, certificate := range keyMaterial.CertificateChain {
		if certificate == nil {
			return nil, fmt.Errorf("certificate %d in chain cannot be nil", i)
		}
	}

	var err error

	keyMaterialEncoded := &KeyMaterialEncoded{}
	keyMaterialEncoded.DERCertificateChain = make([][]byte, len(keyMaterial.CertificateChain))
	keyMaterialEncoded.PEMCertificateChain = make([][]byte, len(keyMaterial.CertificateChain))

	for i, certificate := range keyMaterial.CertificateChain {
		keyMaterialEncoded.DERCertificateChain[i] = certificate.Raw
		keyMaterialEncoded.PEMCertificateChain[i] = toCertificatePEM(certificate.Raw)
	}

	if includePrivateKey && keyMaterial.PrivateKey != nil { // pragma: allowlist secret
		keyMaterialEncoded.DERPrivateKey, err = x509.MarshalPKCS8PrivateKey(keyMaterial.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal private key to DER: %w", err)
		}

		keyMaterialEncoded.PEMPrivateKey = toPrivateKeyPEM(keyMaterialEncoded.DERPrivateKey)
	}

	keyMaterialEncoded.DERPublicKey, err = x509.MarshalPKIXPublicKey(keyMaterial.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key to DER: %w", err)
	}

	keyMaterialEncoded.PEMPublicKey = toPublicKeyPEM(keyMaterialEncoded.DERPublicKey)

	return keyMaterialEncoded, nil
}

func deserializeKeyMaterial(keyMaterialEncoded *KeyMaterialEncoded) (*KeyMaterial, error) {
	if keyMaterialEncoded == nil {
		return nil, fmt.Errorf("keyMaterialEncoded cannot be nil")
	} else if len(keyMaterialEncoded.DERCertificateChain) == 0 {
		return nil, fmt.Errorf("DER certificate chain cannot be empty")
	} else if len(keyMaterialEncoded.DERPublicKey) == 0 {
		return nil, fmt.Errorf("DER public key cannot be empty")
	}

	for i, derBytes := range keyMaterialEncoded.DERCertificateChain {
		if len(derBytes) == 0 {
			return nil, fmt.Errorf("DER certificate at index %d in chain cannot be empty", i)
		}
	}

	keyMaterial := &KeyMaterial{}

	var err error

	keyMaterial.CertificateChain = make([]*x509.Certificate, len(keyMaterialEncoded.DERCertificateChain))
	for i, derBytes := range keyMaterialEncoded.DERCertificateChain {
		keyMaterial.CertificateChain[i], err = x509.ParseCertificate(derBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate %d from DER: %w", i, err)
		}
	}

	keyMaterial.PublicKey, err = x509.ParsePKIXPublicKey(keyMaterialEncoded.DERPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key from DER: %w", err)
	}

	if len(keyMaterialEncoded.DERPrivateKey) > 0 {
		keyMaterial.PrivateKey, err = x509.ParsePKCS8PrivateKey(keyMaterialEncoded.DERPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key from DER: %w", err)
		}
	}

	return keyMaterial, nil
}

func toCertificatePEM(certificateDER []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificateDER})
}

func toPrivateKeyPEM(privateKeyDER []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateKeyDER})
}

func toPublicKeyPEM(publicKeyDER []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicKeyDER})
}
