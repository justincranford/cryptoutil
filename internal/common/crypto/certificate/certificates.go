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

type KeyMaterialDecoded struct {
	CertChain              []*x509.Certificate
	PrivateKey             crypto.PrivateKey
	PublicKey              crypto.PublicKey
	SubordinateCACertsPool *x509.CertPool
	RootCACertsPool        *x509.CertPool
}

type KeyMaterialEncoded struct {
	DERCertChain          [][]byte `json:"der_cert_chain"`
	DERPrivateKey         []byte   `json:"der_private_key"`
	DERPublicKey          []byte   `json:"der_public_key"`
	DERSubordinateCACerts [][]byte `json:"der_subordinate_ca_certs"`
	DERRootCACertsPool    [][]byte `json:"der_root_ca_certs_pool"`

	PEMCertChain          [][]byte `json:"pem_cert_chain"`
	PEMPrivateKey         []byte   `json:"pem_private_key"`
	PEMPublicKey          []byte   `json:"pem_public_key"`
	PEMSubordinateCACerts [][]byte `json:"pem_subordinate_ca_certs"`
	PEMRootCACertsPool    [][]byte `json:"pem_root_ca_certs_pool"`
}

type Subject struct {
	SubjectName        string
	IssuerName         string
	Duration           time.Duration
	KeyMaterialDecoded KeyMaterialDecoded

	// Subject type - exactly one should be set
	CASubject        *CASubject
	EndEntitySubject *EndEntitySubject
}

type CASubject struct {
	MaxPathLen int
	IsCA       bool
}

type EndEntitySubject struct {
	DNSNames       []string
	IPAddresses    []net.IP
	EmailAddresses []string
	URIs           []*url.URL
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
			KeyMaterialDecoded: KeyMaterialDecoded{
				PrivateKey:             keyPair.Private,
				PublicKey:              keyPair.Public,
				CertChain:              []*x509.Certificate{},
				RootCACertsPool:        x509.NewCertPool(),
				SubordinateCACertsPool: x509.NewCertPool(),
			},
			CASubject: &CASubject{
				MaxPathLen: numCAs - i - 1,
				IsCA:       true,
			},
		}
		previousSubject := currentSubject
		var previousCACert *x509.Certificate
		if i > 0 {
			previousSubject = subjects[i-1]
			previousCACert = previousSubject.KeyMaterialDecoded.CertChain[0]
		}

		currentCACertTemplate, err := CertificateTemplateCA(previousSubject.IssuerName, currentSubject.SubjectName, currentSubject.Duration, currentSubject.CASubject.MaxPathLen)
		if err != nil {
			return nil, fmt.Errorf("failed to create CA certificate template for %s: %w", currentSubject.SubjectName, err)
		}

		cert, _, pemBytes, err := SignCertificate(previousCACert, previousSubject.KeyMaterialDecoded.PrivateKey, currentCACertTemplate, currentSubject.KeyMaterialDecoded.PublicKey, x509.ECDSAWithSHA256)
		if err != nil {
			return nil, fmt.Errorf("failed to sign CA certificate for %s: %w", currentSubject.SubjectName, err)
		}

		currentSubject.KeyMaterialDecoded.CertChain = append([]*x509.Certificate{cert}, previousSubject.KeyMaterialDecoded.CertChain...)

		// Create DER and PEM chains locally for verification
		derChain := make([][]byte, len(currentSubject.KeyMaterialDecoded.CertChain))
		pemChain := make([][]byte, len(currentSubject.KeyMaterialDecoded.CertChain))
		for j, c := range currentSubject.KeyMaterialDecoded.CertChain {
			derChain[j] = c.Raw
			pemChain[j] = pemBytes // Use the pemBytes from SignCertificate for the first cert
		}

		currentSubject.KeyMaterialDecoded.RootCACertsPool = previousSubject.KeyMaterialDecoded.RootCACertsPool.Clone()
		currentSubject.KeyMaterialDecoded.SubordinateCACertsPool = previousSubject.KeyMaterialDecoded.SubordinateCACertsPool.Clone()
		if i == 0 {
			currentSubject.KeyMaterialDecoded.RootCACertsPool.AddCert(cert)
		} else {
			currentSubject.KeyMaterialDecoded.SubordinateCACertsPool.AddCert(cert)
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
		SubjectName: subjectName,
		IssuerName:  issuingCA.SubjectName,
		Duration:    duration,
		KeyMaterialDecoded: KeyMaterialDecoded{
			PrivateKey:             keyPair.Private,
			PublicKey:              keyPair.Public,
			CertChain:              []*x509.Certificate{},
			RootCACertsPool:        x509.NewCertPool(),
			SubordinateCACertsPool: x509.NewCertPool(),
		},
		EndEntitySubject: &EndEntitySubject{
			DNSNames:       dnsNames,
			IPAddresses:    ipAddresses,
			EmailAddresses: emailAddresses,
			URIs:           uris,
		},
	}

	endEntityCertTemplate, err := CertificateTemplateEndEntity(issuingCA.SubjectName, endEntitySubject.SubjectName, endEntitySubject.Duration, endEntitySubject.EndEntitySubject.DNSNames, endEntitySubject.EndEntitySubject.IPAddresses, endEntitySubject.EndEntitySubject.EmailAddresses, endEntitySubject.EndEntitySubject.URIs, keyUsage, extKeyUsage)
	if err != nil {
		return Subject{}, fmt.Errorf("failed to create end entity certificate template for %s: %w", subjectName, err)
	}

	cert, _, _, err := SignCertificate(issuingCA.KeyMaterialDecoded.CertChain[0], issuingCA.KeyMaterialDecoded.PrivateKey, endEntityCertTemplate, endEntitySubject.KeyMaterialDecoded.PublicKey, x509.ECDSAWithSHA256)
	if err != nil {
		return Subject{}, fmt.Errorf("failed to sign end entity certificate for %s: %w", subjectName, err)
	}

	endEntitySubject.KeyMaterialDecoded.CertChain = append([]*x509.Certificate{cert}, issuingCA.KeyMaterialDecoded.CertChain...)
	endEntitySubject.KeyMaterialDecoded.RootCACertsPool = issuingCA.KeyMaterialDecoded.RootCACertsPool.Clone()
	endEntitySubject.KeyMaterialDecoded.SubordinateCACertsPool = issuingCA.KeyMaterialDecoded.SubordinateCACertsPool.Clone()

	return endEntitySubject, nil
}

func BuildTLSCertificate(endEntitySubject Subject) (tls.Certificate, *x509.CertPool, error) {
	if len(endEntitySubject.KeyMaterialDecoded.CertChain) == 0 {
		return tls.Certificate{}, nil, fmt.Errorf("certificate chain is empty")
	}
	if endEntitySubject.KeyMaterialDecoded.PrivateKey == nil {
		return tls.Certificate{}, nil, fmt.Errorf("private key is nil")
	}
	if endEntitySubject.KeyMaterialDecoded.RootCACertsPool == nil {
		return tls.Certificate{}, nil, fmt.Errorf("root CA certs pool is nil")
	}

	// Convert certificate chain to DER format for TLS
	derCertChain := make([][]byte, len(endEntitySubject.KeyMaterialDecoded.CertChain))
	for i, cert := range endEntitySubject.KeyMaterialDecoded.CertChain {
		derCertChain[i] = cert.Raw
	}

	return tls.Certificate{Certificate: derCertChain, PrivateKey: endEntitySubject.KeyMaterialDecoded.PrivateKey, Leaf: endEntitySubject.KeyMaterialDecoded.CertChain[0]}, endEntitySubject.KeyMaterialDecoded.RootCACertsPool, nil
}

func SerializeSubjects(subjects []Subject, includePrivateKey bool) ([][]byte, error) {
	if subjects == nil {
		return nil, fmt.Errorf("subjects cannot be nil")
	}

	keyMaterialJSONs := make([][]byte, len(subjects))
	for i, subject := range subjects {
		if subject.SubjectName == "" {
			return nil, fmt.Errorf("subject at index %d has empty SubjectName", i)
		} else if subject.IssuerName == "" {
			return nil, fmt.Errorf("subject at index %d has empty IssuerName", i)
		} else if subject.Duration <= 0 {
			return nil, fmt.Errorf("subject at index %d has zero or negative Duration", i)
		}
		if subject.CASubject != nil && subject.EndEntitySubject != nil {
			return nil, fmt.Errorf("subject at index %d cannot have both CASubject and EndEntitySubject populated", i)
		} else if subject.CASubject == nil && subject.EndEntitySubject == nil {
			return nil, fmt.Errorf("subject at index %d must have either CASubject or EndEntitySubject populated", i)
		} else if subject.CASubject != nil {
			if subject.CASubject.MaxPathLen < 0 {
				return nil, fmt.Errorf("subject at index %d has invalid MaxPathLen (%d), must be >= 0", i, subject.CASubject.MaxPathLen)
			}
		} else if subject.EndEntitySubject != nil {
			// Note: DNSNames, IPAddresses, EmailAddresses, and URIs can be empty
		}

		keyMaterialEncoded, err := toKeyMaterialEncoded(&subject.KeyMaterialDecoded, includePrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert KeyMaterialDecoded to JSON format for subject %d: %w", i, err)
		}
		jsonBytes, err := json.Marshal(keyMaterialEncoded)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize KeyMaterialEncoded for subject %d: %w", i, err)
		}
		keyMaterialJSONs[i] = jsonBytes
	}
	return keyMaterialJSONs, nil
}

func DeserializeSubjects(keyMaterialEncodedBytesList [][]byte) ([]Subject, error) {
	subjects := make([]Subject, len(keyMaterialEncodedBytesList))
	for i, keyMaterialEncodedBytes := range keyMaterialEncodedBytesList {
		var keyMaterialEncoded KeyMaterialEncoded
		err := json.Unmarshal(keyMaterialEncodedBytes, &keyMaterialEncoded)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize KeyMaterialEncoded for item %d: %w", i, err)
		}
		keyMaterialDecoded, err := toKeyMaterialDecoded(&keyMaterialEncoded)
		if err != nil {
			return nil, fmt.Errorf("failed to convert KeyMaterialEncoded to KeyMaterialDecoded for item %d: %w", i, err)
		}
		if len(keyMaterialDecoded.CertChain) == 0 {
			return nil, fmt.Errorf("CertChain is empty for item %d", i)
		} else if keyMaterialDecoded.PublicKey == nil {
			return nil, fmt.Errorf("PublicKey is nil for item %d", i)
		}
		cert := keyMaterialDecoded.CertChain[0]
		subject := Subject{
			KeyMaterialDecoded: *keyMaterialDecoded,
			SubjectName:        cert.Subject.CommonName,
			IssuerName:         cert.Issuer.CommonName,
			Duration:           cert.NotAfter.Sub(cert.NotBefore),
		}
		if cert.IsCA {
			subject.CASubject = &CASubject{
				MaxPathLen: cert.MaxPathLen,
				IsCA:       cert.IsCA,
			}
		} else {
			subject.EndEntitySubject = &EndEntitySubject{
				DNSNames:       cert.DNSNames,
				IPAddresses:    cert.IPAddresses,
				EmailAddresses: cert.EmailAddresses,
				URIs:           cert.URIs,
			}
		}
		subjects[i] = subject
	}
	return subjects, nil
}

func toKeyMaterialEncoded(keyMaterialDecoded *KeyMaterialDecoded, includePrivateKey bool) (*KeyMaterialEncoded, error) {
	if keyMaterialDecoded == nil {
		return nil, fmt.Errorf("keyMaterialDecoded cannot be nil")
	} else if len(keyMaterialDecoded.CertChain) == 0 {
		return nil, fmt.Errorf("certificate chain cannot be empty")
	} else if keyMaterialDecoded.PublicKey == nil {
		return nil, fmt.Errorf("PublicKey cannot be nil")
	} else if keyMaterialDecoded.SubordinateCACertsPool == nil {
		return nil, fmt.Errorf("SubordinateCACertsPool cannot be nil")
	} else if keyMaterialDecoded.RootCACertsPool == nil {
		return nil, fmt.Errorf("RootCACertsPool cannot be nil")
	}
	for i, cert := range keyMaterialDecoded.CertChain {
		if cert == nil {
			return nil, fmt.Errorf("certificate at index %d in chain cannot be nil", i)
		}
	}

	var err error
	keyMaterialEncoded := &KeyMaterialEncoded{}
	keyMaterialEncoded.DERCertChain = make([][]byte, len(keyMaterialDecoded.CertChain))
	keyMaterialEncoded.PEMCertChain = make([][]byte, len(keyMaterialDecoded.CertChain))
	for i, cert := range keyMaterialDecoded.CertChain {
		keyMaterialEncoded.DERCertChain[i] = cert.Raw
		keyMaterialEncoded.PEMCertChain[i] = pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		})
	}

	if includePrivateKey && keyMaterialDecoded.PrivateKey != nil {
		keyMaterialEncoded.DERPrivateKey, err = x509.MarshalPKCS8PrivateKey(keyMaterialDecoded.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal private key to DER: %w", err)
		}
		keyMaterialEncoded.PEMPrivateKey = pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: keyMaterialEncoded.DERPrivateKey,
		})
	}

	keyMaterialEncoded.DERPublicKey, err = x509.MarshalPKIXPublicKey(keyMaterialDecoded.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key to DER: %w", err)
	}
	keyMaterialEncoded.PEMPublicKey = pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: keyMaterialEncoded.DERPublicKey,
	})

	keyMaterialEncoded.DERSubordinateCACerts = [][]byte{}
	keyMaterialEncoded.PEMSubordinateCACerts = [][]byte{}
	keyMaterialEncoded.DERRootCACertsPool = [][]byte{}
	keyMaterialEncoded.PEMRootCACertsPool = [][]byte{}

	for i := 1; i < len(keyMaterialDecoded.CertChain)-1; i++ {
		cert := keyMaterialDecoded.CertChain[i]
		if cert.IsCA {
			keyMaterialEncoded.DERSubordinateCACerts = append(keyMaterialEncoded.DERSubordinateCACerts, cert.Raw)
			keyMaterialEncoded.PEMSubordinateCACerts = append(keyMaterialEncoded.PEMSubordinateCACerts, pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: cert.Raw,
			}))
		}
	}

	rootCA := keyMaterialDecoded.CertChain[len(keyMaterialDecoded.CertChain)-1]
	if rootCA.IsCA {
		keyMaterialEncoded.DERRootCACertsPool = append(keyMaterialEncoded.DERRootCACertsPool, rootCA.Raw)
		keyMaterialEncoded.PEMRootCACertsPool = append(keyMaterialEncoded.PEMRootCACertsPool, pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: rootCA.Raw,
		}))
	}

	return keyMaterialEncoded, nil
}

func toKeyMaterialDecoded(keyMaterialEncoded *KeyMaterialEncoded) (*KeyMaterialDecoded, error) {
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

	keyMaterialDecoded := &KeyMaterialDecoded{}
	var err error

	keyMaterialDecoded.CertChain = make([]*x509.Certificate, len(keyMaterialEncoded.DERCertChain))
	for i, derBytes := range keyMaterialEncoded.DERCertChain {
		keyMaterialDecoded.CertChain[i], err = x509.ParseCertificate(derBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate %d from DER: %w", i, err)
		}
	}

	keyMaterialDecoded.PublicKey, err = x509.ParsePKIXPublicKey(keyMaterialEncoded.DERPublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key from DER: %w", err)
	}

	if len(keyMaterialEncoded.DERPrivateKey) > 0 {
		keyMaterialDecoded.PrivateKey, err = x509.ParsePKCS8PrivateKey(keyMaterialEncoded.DERPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key from DER: %w", err)
		}
	}

	keyMaterialDecoded.SubordinateCACertsPool = x509.NewCertPool()
	for i, derBytes := range keyMaterialEncoded.DERSubordinateCACerts {
		if len(derBytes) == 0 {
			return nil, fmt.Errorf("DER subordinate CA certificate at index %d cannot be empty", i)
		}
		cert, err := x509.ParseCertificate(derBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse subordinate CA certificate %d from DER: %w", i, err)
		}
		keyMaterialDecoded.SubordinateCACertsPool.AddCert(cert)
	}

	keyMaterialDecoded.RootCACertsPool = x509.NewCertPool()
	for i, derBytes := range keyMaterialEncoded.DERRootCACertsPool {
		if len(derBytes) == 0 {
			return nil, fmt.Errorf("DER root CA certificate at index %d cannot be empty", i)
		}
		cert, err := x509.ParseCertificate(derBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse root CA certificate %d from DER: %w", i, err)
		}
		keyMaterialDecoded.RootCACertsPool.AddCert(cert)
	}

	return keyMaterialDecoded, nil
}
