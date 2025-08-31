package certificate

import (
	"crypto"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"time"

	"cryptoutil/internal/common/crypto/keygen"
	cryptoutilDateTime "cryptoutil/internal/common/util/datetime"
)

type KeyMaterial struct {
	CertChain  []*x509.Certificate
	PrivateKey crypto.PrivateKey
	PublicKey  crypto.PublicKey
}

type KeyMaterialEncoded struct {
	DERCertificateChain [][]byte `json:"der_certificate_chain"`
	DERPrivateKey       []byte   `json:"der_private_key"`
	DERPublicKey        []byte   `json:"der_public_key"`

	PEMCertificateChain [][]byte `json:"pem_certificate_chain"`
	PEMPrivateKey       []byte   `json:"pem_private_key"`
	PEMPublicKey        []byte   `json:"pem_public_key"`
}

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

func CreateCASubjects(keyPairs []*keygen.KeyPair, caSubjectNamePrefix string) ([]*Subject, error) {
	subjects := make([]*Subject, len(keyPairs))
	for i := range len(keyPairs) {
		var err error
		if i == 0 {
			subjects[i], err = CreateCASubject(nil, keyPairs[i].Private, fmt.Sprintf("%s %d", caSubjectNamePrefix, i), keyPairs[i], len(keyPairs)-i-1)
		} else {
			subjects[i], err = CreateCASubject(subjects[i-1], subjects[i-1].KeyMaterial.PrivateKey, fmt.Sprintf("%s %d", caSubjectNamePrefix, i), keyPairs[i], len(keyPairs)-i-1)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create CA subject %d: %w", i, err)
		}
	}
	return subjects, nil
}

func CreateCASubject(issuerSubject *Subject, issuerPrivateKey crypto.PrivateKey, subjectName string, subjectKeyPair *keygen.KeyPair, maxPathLen int) (*Subject, error) {
	if issuerPrivateKey == nil {
		return nil, fmt.Errorf("issuerPrivateKey should not be nil for CA %s", subjectName)
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
	var issuerCert *x509.Certificate
	var issuerCertificateChain []*x509.Certificate
	if issuerSubject == nil {
		issuerName = subjectName
		issuerCert = nil
		issuerCertificateChain = []*x509.Certificate{}
	} else {
		issuerName = issuerSubject.SubjectName
		issuerCert = issuerSubject.KeyMaterial.CertChain[0]
		issuerCertificateChain = issuerSubject.KeyMaterial.CertChain[1:]
	}
	currentSubject := Subject{
		SubjectName: subjectName,
		IssuerName:  issuerName,
		Duration:    10 * 365 * cryptoutilDateTime.Days1,
		IsCA:        true,
		MaxPathLen:  maxPathLen,
		KeyMaterial: KeyMaterial{
			PrivateKey: subjectKeyPair.Private,
			PublicKey:  subjectKeyPair.Public,
			CertChain:  []*x509.Certificate{},
		},
	}

	currentCACertTemplate, err := CertificateTemplateCA(issuerName, subjectName, currentSubject.Duration, maxPathLen)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate template for %s: %w", subjectName, err)
	}

	signedCertificate, _, _, err := SignCertificate(issuerCert, issuerPrivateKey, currentCACertTemplate, currentSubject.KeyMaterial.PublicKey, x509.ECDSAWithSHA256)
	if err != nil {
		return nil, fmt.Errorf("failed to sign CA certificate for %s: %w", subjectName, err)
	}

	currentSubject.KeyMaterial.CertChain = append([]*x509.Certificate{signedCertificate}, issuerCertificateChain...)
	return &currentSubject, nil
}

func CreateEndEntitySubject(keyPair *keygen.KeyPair, subjectName string, duration time.Duration, dnsNames []string, ipAddresses []net.IP, emailAddresses []string, uris []*url.URL, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage, caSubjects []*Subject) (*Subject, error) {
	// The issuing CA is the last one in the chain (leaf CA)
	if len(caSubjects) == 0 {
		return nil, fmt.Errorf("caSubjects should not be empty")
	}
	issuingCA := caSubjects[len(caSubjects)-1]
	if issuingCA.SubjectName == "" {
		return nil, fmt.Errorf("issuingCA.SubjectName should not be empty")
	}

	endEntitySubject := Subject{
		SubjectName:    subjectName,
		IssuerName:     issuingCA.SubjectName,
		Duration:       duration,
		IsCA:           false,
		DNSNames:       dnsNames,
		IPAddresses:    ipAddresses,
		EmailAddresses: emailAddresses,
		URIs:           uris,
		KeyMaterial: KeyMaterial{
			PrivateKey: keyPair.Private,
			PublicKey:  keyPair.Public,
			CertChain:  []*x509.Certificate{},
		},
	}

	endEntityCertTemplate, err := CertificateTemplateEndEntity(issuingCA.SubjectName, endEntitySubject.SubjectName, endEntitySubject.Duration, endEntitySubject.DNSNames, endEntitySubject.IPAddresses, endEntitySubject.EmailAddresses, endEntitySubject.URIs, keyUsage, extKeyUsage)
	if err != nil {
		return nil, fmt.Errorf("failed to create end entity certificate template for %s: %w", subjectName, err)
	}

	signedCert, _, _, err := SignCertificate(issuingCA.KeyMaterial.CertChain[0], issuingCA.KeyMaterial.PrivateKey, endEntityCertTemplate, endEntitySubject.KeyMaterial.PublicKey, x509.ECDSAWithSHA256)
	if err != nil {
		return nil, fmt.Errorf("failed to sign end entity certificate for %s: %w", subjectName, err)
	}

	endEntitySubject.KeyMaterial.CertChain = append([]*x509.Certificate{signedCert}, issuingCA.KeyMaterial.CertChain...)

	return &endEntitySubject, nil
}

func BuildTLSCertificate(endEntitySubject *Subject) (tls.Certificate, *x509.CertPool, *x509.CertPool, error) {
	if len(endEntitySubject.KeyMaterial.CertChain) == 0 {
		return tls.Certificate{}, nil, nil, fmt.Errorf("certificate chain is empty")
	} else if endEntitySubject.KeyMaterial.PrivateKey == nil {
		return tls.Certificate{}, nil, nil, fmt.Errorf("private key is nil")
	}
	derCertChain := make([][]byte, len(endEntitySubject.KeyMaterial.CertChain))
	for i, certificate := range endEntitySubject.KeyMaterial.CertChain {
		derCertChain[i] = certificate.Raw
	}
	rootCACertsPool := x509.NewCertPool()
	if len(endEntitySubject.KeyMaterial.CertChain) > 0 {
		rootCert := endEntitySubject.KeyMaterial.CertChain[len(endEntitySubject.KeyMaterial.CertChain)-1]
		rootCACertsPool.AddCert(rootCert)
	}
	intermediateCertsPool := x509.NewCertPool()
	for j := 1; j < len(endEntitySubject.KeyMaterial.CertChain)-1; j++ {
		intermediateCertsPool.AddCert(endEntitySubject.KeyMaterial.CertChain[j])
	}

	return tls.Certificate{Certificate: derCertChain, PrivateKey: endEntitySubject.KeyMaterial.PrivateKey, Leaf: endEntitySubject.KeyMaterial.CertChain[0]}, rootCACertsPool, intermediateCertsPool, nil
}

func CertificateTemplateCA(issuerName string, subjectName string, duration time.Duration, maxPathLen int) (*x509.Certificate, error) {
	serialNumber, err := GenerateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number for TLS root CA: %w", err)
	}
	notBefore, notAfter, err := randomizedNotBeforeNotAfterCA(time.Now().UTC(), duration, 1*time.Minute, 120*time.Minute)
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

func CertificateTemplateEndEntity(issuerName string, subjectName string, duration time.Duration, dnsNames []string, ipAddresses []net.IP, emailAddresses []string, uris []*url.URL, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage) (*x509.Certificate, error) {
	serialNumber, err := GenerateSerialNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial number for TLS server: %w", err)
	}
	notBefore, notAfter, err := randomizedNotBeforeNotAfterEndEntity(time.Now().UTC(), duration, 1*time.Minute, 120*time.Minute)
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

func SignCertificate(issuerCert *x509.Certificate, issuerPrivateKey crypto.PrivateKey, subjectCert *x509.Certificate, subjectPublicKey crypto.PublicKey, signatureAlgorithm x509.SignatureAlgorithm) (*x509.Certificate, []byte, []byte, error) {
	_, ok := issuerPrivateKey.(crypto.Signer)
	if !ok {
		return nil, nil, nil, fmt.Errorf("issuer private key is not a crypto.Signer")
	}
	subjectCert.SignatureAlgorithm = signatureAlgorithm
	var err error
	var certificateDER []byte
	if issuerCert == nil {
		certificateDER, err = x509.CreateCertificate(rand.Reader, subjectCert, subjectCert, subjectPublicKey, issuerPrivateKey)
	} else {
		certificateDER, err = x509.CreateCertificate(rand.Reader, subjectCert, issuerCert, subjectPublicKey, issuerPrivateKey)
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

		keyMaterialEncoded, err := toKeyMaterialEncoded(&subject.KeyMaterial, includePrivateKey)
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

func DeserializeSubjects(keyMaterialEncodedBytesList [][]byte) ([]*Subject, error) {
	subjects := make([]*Subject, len(keyMaterialEncodedBytesList))
	for i, keyMaterialEncodedBytes := range keyMaterialEncodedBytesList {
		var keyMaterialEncoded KeyMaterialEncoded
		err := json.Unmarshal(keyMaterialEncodedBytes, &keyMaterialEncoded)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize KeyMaterialEncoded for item %d: %w", i, err)
		}
		keyMaterial, err := toKeyMaterial(&keyMaterialEncoded)
		if err != nil {
			return nil, fmt.Errorf("failed to convert KeyMaterialEncoded to KeyMaterial for item %d: %w", i, err)
		}
		if len(keyMaterial.CertChain) == 0 {
			return nil, fmt.Errorf("CertChain is empty for item %d", i)
		} else if keyMaterial.PublicKey == nil {
			return nil, fmt.Errorf("PublicKey is nil for item %d", i)
		}
		certificate := keyMaterial.CertChain[0]
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

func toKeyMaterialEncoded(keyMaterial *KeyMaterial, includePrivateKey bool) (*KeyMaterialEncoded, error) {
	if keyMaterial == nil {
		return nil, fmt.Errorf("keyMaterial cannot be nil")
	} else if len(keyMaterial.CertChain) == 0 {
		return nil, fmt.Errorf("certificate chain cannot be empty")
	} else if keyMaterial.PublicKey == nil {
		return nil, fmt.Errorf("PublicKey cannot be nil")
	}
	for i, certificate := range keyMaterial.CertChain {
		if certificate == nil {
			return nil, fmt.Errorf("certificate at index %d in chain cannot be nil", i)
		}
	}

	var err error
	keyMaterialEncoded := &KeyMaterialEncoded{}
	keyMaterialEncoded.DERCertificateChain = make([][]byte, len(keyMaterial.CertChain))
	keyMaterialEncoded.PEMCertificateChain = make([][]byte, len(keyMaterial.CertChain))
	for i, certificate := range keyMaterial.CertChain {
		keyMaterialEncoded.DERCertificateChain[i] = certificate.Raw
		keyMaterialEncoded.PEMCertificateChain[i] = toCertificatePEM(certificate.Raw)
	}

	if includePrivateKey && keyMaterial.PrivateKey != nil {
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

func toKeyMaterial(keyMaterialEncoded *KeyMaterialEncoded) (*KeyMaterial, error) {
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

	keyMaterial.CertChain = make([]*x509.Certificate, len(keyMaterialEncoded.DERCertificateChain))
	for i, derBytes := range keyMaterialEncoded.DERCertificateChain {
		keyMaterial.CertChain[i], err = x509.ParseCertificate(derBytes)
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
