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
)

type KeyMaterial struct {
	CertChain              []*x509.Certificate `json:"-"`
	PrivateKey             crypto.PrivateKey   `json:"-"`
	PublicKey              crypto.PublicKey    `json:"-"`
	SubordinateCACertsPool *x509.CertPool      `json:"-"`
	RootCACertsPool        *x509.CertPool      `json:"-"`

	DERCertChain          [][]byte `json:"der_cert_chain"`
	DERPrivateKey         []byte   `json:"der_private_key,omitempty"`
	DERPublicKey          []byte   `json:"der_public_key"`
	DERSubordinateCACerts [][]byte `json:"der_subordinate_ca_certs"`
	DERRootCACertsPool    [][]byte `json:"der_root_ca_certs_pool"`

	PEMCertChain          [][]byte `json:"pem_cert_chain"`
	PEMPrivateKey         []byte   `json:"pem_private_key,omitempty"`
	PEMPublicKey          []byte   `json:"pem_public_key"`
	PEMSubordinateCACerts [][]byte `json:"pem_subordinate_ca_certs"`
	PEMRootCACertsPool    [][]byte `json:"pem_root_ca_certs_pool"`
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
			DERChain:    subject.KeyMaterial.DERCertChain,
			PEMChain:    subject.KeyMaterial.PEMCertChain,
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

func SerializeKeyMaterial(keyMaterial *KeyMaterial, includePrivateKey bool) ([]byte, error) {
	var data []byte
	var err error

	if includePrivateKey {
		// No copy needed, serialize directly
		data, err = json.Marshal(keyMaterial)
	} else {
		// Create a copy to avoid modifying the original, then clear private key fields
		keyMaterialWithoutPrivateKey := *keyMaterial
		keyMaterialWithoutPrivateKey.DERPrivateKey = nil
		keyMaterialWithoutPrivateKey.PEMPrivateKey = nil
		data, err = json.Marshal(keyMaterialWithoutPrivateKey)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to serialize KeyMaterial: %w", err)
	}
	return data, nil
}

func DeserializeKeyMaterial(data []byte) (*KeyMaterial, error) {
	var keyMaterial KeyMaterial
	err := json.Unmarshal(data, &keyMaterial)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize KeyMaterial: %w", err)
	}
	return &keyMaterial, nil
}

func (km *KeyMaterial) PopulateSerializableFields() error {
	var err error

	// Serialize private key if present
	if km.PrivateKey != nil {
		km.DERPrivateKey, err = x509.MarshalPKCS8PrivateKey(km.PrivateKey)
		if err != nil {
			return fmt.Errorf("failed to marshal private key to DER: %w", err)
		}

		km.PEMPrivateKey = pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: km.DERPrivateKey,
		})
	}

	// Serialize public key if present
	if km.PublicKey != nil {
		km.DERPublicKey, err = x509.MarshalPKIXPublicKey(km.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to marshal public key to DER: %w", err)
		}

		km.PEMPublicKey = pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: km.DERPublicKey,
		})
	}

	// Serialize subordinate CA certs pool if present
	// Note: x509.CertPool doesn't expose the certificates directly,
	// so these fields need to be populated manually when certificates are added to the pool

	// Serialize root CA certs pool if present
	// Note: Same limitation as above - these fields need to be populated manually

	return nil
}

// ReconstructCryptoObjects reconstructs the crypto objects from the serializable DER/PEM data
func (km *KeyMaterial) ReconstructCryptoObjects() error {
	var err error

	// Reconstruct private key from DER if present
	if len(km.DERPrivateKey) > 0 {
		km.PrivateKey, err = x509.ParsePKCS8PrivateKey(km.DERPrivateKey)
		if err != nil {
			return fmt.Errorf("failed to parse private key from DER: %w", err)
		}
	}

	// Reconstruct public key from DER if present
	if len(km.DERPublicKey) > 0 {
		km.PublicKey, err = x509.ParsePKIXPublicKey(km.DERPublicKey)
		if err != nil {
			return fmt.Errorf("failed to parse public key from DER: %w", err)
		}
	}

	// Reconstruct cert chain from DER chain
	if len(km.DERCertChain) > 0 {
		km.CertChain = make([]*x509.Certificate, len(km.DERCertChain))
		for i, derBytes := range km.DERCertChain {
			km.CertChain[i], err = x509.ParseCertificate(derBytes)
			if err != nil {
				return fmt.Errorf("failed to parse certificate %d from DER: %w", i, err)
			}
		}
	}

	// Reconstruct subordinate CA certs pool from DER
	if len(km.DERSubordinateCACerts) > 0 {
		km.SubordinateCACertsPool = x509.NewCertPool()
		for i, derBytes := range km.DERSubordinateCACerts {
			cert, err := x509.ParseCertificate(derBytes)
			if err != nil {
				return fmt.Errorf("failed to parse subordinate CA certificate %d from DER: %w", i, err)
			}
			km.SubordinateCACertsPool.AddCert(cert)
		}
	}

	// Reconstruct root CA certs pool from DER
	if len(km.DERRootCACertsPool) > 0 {
		km.RootCACertsPool = x509.NewCertPool()
		for i, derBytes := range km.DERRootCACertsPool {
			cert, err := x509.ParseCertificate(derBytes)
			if err != nil {
				return fmt.Errorf("failed to parse root CA certificate %d from DER: %w", i, err)
			}
			km.RootCACertsPool.AddCert(cert)
		}
	}

	return nil
}

// AddToSubordinateCACertsPool adds a certificate to the subordinate CA certs pool
// and updates the serializable DER/PEM representations
func (km *KeyMaterial) AddToSubordinateCACertsPool(cert *x509.Certificate) {
	if km.SubordinateCACertsPool == nil {
		km.SubordinateCACertsPool = x509.NewCertPool()
	}
	km.SubordinateCACertsPool.AddCert(cert)

	// Update serializable representations
	km.DERSubordinateCACerts = append(km.DERSubordinateCACerts, cert.Raw)
	km.PEMSubordinateCACerts = append(km.PEMSubordinateCACerts, pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}))
}

// AddToRootCACertsPool adds a certificate to the root CA certs pool
// and updates the serializable DER/PEM representations
func (km *KeyMaterial) AddToRootCACertsPool(cert *x509.Certificate) {
	if km.RootCACertsPool == nil {
		km.RootCACertsPool = x509.NewCertPool()
	}
	km.RootCACertsPool.AddCert(cert)

	// Update serializable representations
	km.DERRootCACertsPool = append(km.DERRootCACertsPool, cert.Raw)
	km.PEMRootCACertsPool = append(km.PEMRootCACertsPool, pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}))
}
