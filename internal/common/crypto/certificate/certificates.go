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
	cryptoutilPool "cryptoutil/internal/common/pool"
	cryptoutilDateTime "cryptoutil/internal/common/util/datetime"
)

type KeyMaterial struct {
	CertChain  []*x509.Certificate
	PrivateKey crypto.PrivateKey
	PublicKey  crypto.PublicKey
}

type KeyMaterialEncoded struct {
	DERCertChain  [][]byte `json:"der_cert_chain"`
	DERPrivateKey []byte   `json:"der_private_key"`
	DERPublicKey  []byte   `json:"der_public_key"`

	PEMCertChain  [][]byte `json:"pem_cert_chain"`
	PEMPrivateKey []byte   `json:"pem_private_key"`
	PEMPublicKey  []byte   `json:"pem_public_key"`
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

func CreateCASubjects(keygenPool *cryptoutilPool.ValueGenPool[*keygen.KeyPair], caSubjectNamePrefix string, numCAs int) ([]Subject, error) {
	if numCAs <= 0 {
		return nil, fmt.Errorf("numCAs must be greater than 0")
	}
	subjects := make([]Subject, numCAs)
	for i := range numCAs {
		keyPair := keygenPool.Get()
		if keyPair == nil {
			return nil, fmt.Errorf("keyPair should not be nil for CA %d", i)
		}
		if keyPair.Private == nil {
			return nil, fmt.Errorf("keyPair.Private should not be nil for CA %d", i)
		}
		if keyPair.Public == nil {
			return nil, fmt.Errorf("keyPair.Public should not be nil for CA %d", i)
		}

		// Determine issuer name - root CA issues itself, others are issued by previous CA
		issuerName := fmt.Sprintf("%s %d", caSubjectNamePrefix, i) // Self-signed for root CA
		if i > 0 {
			issuerName = fmt.Sprintf("%s %d", caSubjectNamePrefix, i-1) // Previous CA for intermediate CAs
		}

		currentSubject := Subject{
			SubjectName: fmt.Sprintf("%s %d", caSubjectNamePrefix, i),
			IssuerName:  issuerName,
			Duration:    10 * 365 * cryptoutilDateTime.Days1,
			IsCA:        true,
			MaxPathLen:  numCAs - i - 1,
			KeyMaterial: KeyMaterial{
				PrivateKey: keyPair.Private,
				PublicKey:  keyPair.Public,
				CertChain:  []*x509.Certificate{},
			},
		}
		previousSubject := currentSubject
		var previousCACert *x509.Certificate
		if i > 0 {
			previousSubject = subjects[i-1]
			previousCACert = previousSubject.KeyMaterial.CertChain[0]
		}

		currentCACertTemplate, err := CertificateTemplateCA(previousSubject.IssuerName, currentSubject.SubjectName, currentSubject.Duration, currentSubject.MaxPathLen)
		if err != nil {
			return nil, fmt.Errorf("failed to create CA certificate template for %s: %w", currentSubject.SubjectName, err)
		}

		cert, _, pemBytes, err := SignCertificate(previousCACert, previousSubject.KeyMaterial.PrivateKey, currentCACertTemplate, currentSubject.KeyMaterial.PublicKey, x509.ECDSAWithSHA256)
		if err != nil {
			return nil, fmt.Errorf("failed to sign CA certificate for %s: %w", currentSubject.SubjectName, err)
		}

		currentSubject.KeyMaterial.CertChain = append([]*x509.Certificate{cert}, previousSubject.KeyMaterial.CertChain...)

		// Create DER and PEM chains locally for verification
		derChain := make([][]byte, len(currentSubject.KeyMaterial.CertChain))
		pemChain := make([][]byte, len(currentSubject.KeyMaterial.CertChain))
		for j, c := range currentSubject.KeyMaterial.CertChain {
			derChain[j] = c.Raw
			pemChain[j] = pemBytes // Use the pemBytes from SignCertificate for the first cert
		}

		subjects[i] = currentSubject
	}
	return subjects, nil
}

func CreateEndEntitySubject(keygenPool *cryptoutilPool.ValueGenPool[*keygen.KeyPair], subjectName string, duration time.Duration, dnsNames []string, ipAddresses []net.IP, emailAddresses []string, uris []*url.URL, keyUsage x509.KeyUsage, extKeyUsage []x509.ExtKeyUsage, caSubjects []Subject) (Subject, error) {
	keyPair := keygenPool.Get()
	if keyPair == nil {
		return Subject{}, fmt.Errorf("keyPair should not be nil")
	}
	if keyPair.Private == nil {
		return Subject{}, fmt.Errorf("keyPair.Private should not be nil")
	}
	if keyPair.Public == nil {
		return Subject{}, fmt.Errorf("keyPair.Public should not be nil")
	}

	// The issuing CA is the last one in the chain (leaf CA)
	if len(caSubjects) == 0 {
		return Subject{}, fmt.Errorf("caSubjects should not be empty")
	}
	issuingCA := caSubjects[len(caSubjects)-1]
	if issuingCA.SubjectName == "" {
		return Subject{}, fmt.Errorf("issuingCA.SubjectName should not be empty")
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
		return Subject{}, fmt.Errorf("failed to create end entity certificate template for %s: %w", subjectName, err)
	}

	signedCert, _, _, err := SignCertificate(issuingCA.KeyMaterial.CertChain[0], issuingCA.KeyMaterial.PrivateKey, endEntityCertTemplate, endEntitySubject.KeyMaterial.PublicKey, x509.ECDSAWithSHA256)
	if err != nil {
		return Subject{}, fmt.Errorf("failed to sign end entity certificate for %s: %w", subjectName, err)
	}

	endEntitySubject.KeyMaterial.CertChain = append([]*x509.Certificate{signedCert}, issuingCA.KeyMaterial.CertChain...)

	return endEntitySubject, nil
}

func BuildTLSCertificate(endEntitySubject Subject) (tls.Certificate, *x509.CertPool, *x509.CertPool, error) {
	if len(endEntitySubject.KeyMaterial.CertChain) == 0 {
		return tls.Certificate{}, nil, nil, fmt.Errorf("certificate chain is empty")
	} else if endEntitySubject.KeyMaterial.PrivateKey == nil {
		return tls.Certificate{}, nil, nil, fmt.Errorf("private key is nil")
	}

	// Convert certificate chain to DER format for TLS
	derCertChain := make([][]byte, len(endEntitySubject.KeyMaterial.CertChain))
	for i, cert := range endEntitySubject.KeyMaterial.CertChain {
		derCertChain[i] = cert.Raw
	}

	// Construct root CA pool from the last certificate in the chain
	rootCACertsPool := x509.NewCertPool()
	if len(endEntitySubject.KeyMaterial.CertChain) > 0 {
		rootCert := endEntitySubject.KeyMaterial.CertChain[len(endEntitySubject.KeyMaterial.CertChain)-1]
		rootCACertsPool.AddCert(rootCert)
	}

	// Construct intermediate certificate pool from certificates between leaf and root
	intermediateCertsPool := x509.NewCertPool()
	for j := 1; j < len(endEntitySubject.KeyMaterial.CertChain)-1; j++ {
		intermediateCertsPool.AddCert(endEntitySubject.KeyMaterial.CertChain[j])
	}

	return tls.Certificate{Certificate: derCertChain, PrivateKey: endEntitySubject.KeyMaterial.PrivateKey, Leaf: endEntitySubject.KeyMaterial.CertChain[0]}, rootCACertsPool, intermediateCertsPool, nil
}

func SerializeSubjects(subjects []Subject, includePrivateKey bool) ([][]byte, error) {
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

func DeserializeSubjects(keyMaterialEncodedBytesList [][]byte) ([]Subject, error) {
	subjects := make([]Subject, len(keyMaterialEncodedBytesList))
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
		cert := keyMaterial.CertChain[0]
		subject := Subject{
			KeyMaterial: *keyMaterial,
			SubjectName: cert.Subject.CommonName,
			IssuerName:  cert.Issuer.CommonName,
			Duration:    cert.NotAfter.Sub(cert.NotBefore),
			IsCA:        cert.IsCA,
		}
		if cert.IsCA {
			subject.MaxPathLen = cert.MaxPathLen
		} else {
			subject.DNSNames = cert.DNSNames
			subject.IPAddresses = cert.IPAddresses
			subject.EmailAddresses = cert.EmailAddresses
			subject.URIs = cert.URIs
		}
		subjects[i] = subject
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
	for i, cert := range keyMaterial.CertChain {
		if cert == nil {
			return nil, fmt.Errorf("certificate at index %d in chain cannot be nil", i)
		}
	}

	var err error
	keyMaterialEncoded := &KeyMaterialEncoded{}
	keyMaterialEncoded.DERCertChain = make([][]byte, len(keyMaterial.CertChain))
	keyMaterialEncoded.PEMCertChain = make([][]byte, len(keyMaterial.CertChain))
	for i, cert := range keyMaterial.CertChain {
		keyMaterialEncoded.DERCertChain[i] = cert.Raw
		keyMaterialEncoded.PEMCertChain[i] = pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		})
	}

	if includePrivateKey && keyMaterial.PrivateKey != nil {
		keyMaterialEncoded.DERPrivateKey, err = x509.MarshalPKCS8PrivateKey(keyMaterial.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal private key to DER: %w", err)
		}
		keyMaterialEncoded.PEMPrivateKey = pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: keyMaterialEncoded.DERPrivateKey,
		})
	}

	keyMaterialEncoded.DERPublicKey, err = x509.MarshalPKIXPublicKey(keyMaterial.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key to DER: %w", err)
	}
	keyMaterialEncoded.PEMPublicKey = pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: keyMaterialEncoded.DERPublicKey,
	})

	return keyMaterialEncoded, nil
}

func toKeyMaterial(keyMaterialEncoded *KeyMaterialEncoded) (*KeyMaterial, error) {
	if keyMaterialEncoded == nil {
		return nil, fmt.Errorf("keyMaterialEncoded cannot be nil")
	} else if len(keyMaterialEncoded.DERCertChain) == 0 {
		return nil, fmt.Errorf("DER certificate chain cannot be empty")
	} else if len(keyMaterialEncoded.DERPublicKey) == 0 {
		return nil, fmt.Errorf("DER public key cannot be empty")
	}
	for i, derBytes := range keyMaterialEncoded.DERCertChain {
		if len(derBytes) == 0 {
			return nil, fmt.Errorf("DER certificate at index %d in chain cannot be empty", i)
		}
	}

	keyMaterial := &KeyMaterial{}
	var err error

	keyMaterial.CertChain = make([]*x509.Certificate, len(keyMaterialEncoded.DERCertChain))
	for i, derBytes := range keyMaterialEncoded.DERCertChain {
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
