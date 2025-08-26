package certificate

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"time"

	cryptoutilKeyGen "cryptoutil/internal/common/crypto/keygen"
)

type KeyMaterial struct {
	KeyPair                *cryptoutilKeyGen.KeyPair
	CertChain              []*x509.Certificate
	DERChain               [][]byte
	PEMChain               [][]byte
	SubordinateCACertsPool *x509.CertPool
	RootCACertsPool        *x509.CertPool
}

type Subject struct {
	SubjectName string
	Duration    time.Duration
	KeyMaterial KeyMaterial
}

type CASubject struct {
	Subject    // Embedded Subject struct
	MaxPathLen int
}

type EndEntitySubject struct {
	Subject        // Embedded Subject struct
	DNSNames       []string
	IPAddresses    []net.IP
	EmailAddresses []string
	URIs           []*url.URL
}

// SerializableCASubject represents a CASubject with only serializable fields
type SerializableCASubject struct {
	SubjectName string        `json:"subject_name"`
	Duration    time.Duration `json:"duration"`
	MaxPathLen  int           `json:"max_path_len"`
	DERChain    [][]byte      `json:"der_chain"`
	PEMChain    [][]byte      `json:"pem_chain"`
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

// SerializeCASubjects serializes a slice of CASubject to JSON bytes
// Note: This function excludes private keys and cert pools for security reasons
func SerializeCASubjects(caSubjects []CASubject) ([]byte, error) {
	serializableSubjects := make([]SerializableCASubject, len(caSubjects))
	for i, subject := range caSubjects {
		serializableSubjects[i] = SerializableCASubject{
			SubjectName: subject.SubjectName,
			Duration:    subject.Duration,
			MaxPathLen:  subject.MaxPathLen,
			DERChain:    subject.KeyMaterial.DERChain,
			PEMChain:    subject.KeyMaterial.PEMChain,
		}
	}

	data, err := json.Marshal(serializableSubjects)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize CA subjects: %w", err)
	}
	return data, nil
}

// DeserializeCASubjects deserializes JSON bytes to a slice of SerializableCASubject
// Note: This returns SerializableCASubject structs, not full CASubject structs,
// as private keys and cert pools cannot be safely serialized/deserialized
func DeserializeCASubjects(data []byte) ([]SerializableCASubject, error) {
	var serializableSubjects []SerializableCASubject
	err := json.Unmarshal(data, &serializableSubjects)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize CA subjects: %w", err)
	}
	return serializableSubjects, nil
}
