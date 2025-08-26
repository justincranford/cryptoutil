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
	CertChain              []*x509.Certificate
	PrivateKey             crypto.PrivateKey
	PublicKey              crypto.PublicKey
	SubordinateCACertsPool *x509.CertPool
	RootCACertsPool        *x509.CertPool
}

// KeyMaterialJSON contains serializable DER/PEM representations of KeyMaterial
type KeyMaterialJSON struct {
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
		// Convert KeyMaterial to JSON format
		keyMaterialJSON, err := subject.KeyMaterial.ToJSON(false) // Don't include private keys
		if err != nil {
			return nil, fmt.Errorf("failed to convert KeyMaterial to JSON format: %w", err)
		}

		serializableSubjects[i] = SerializableCASubject{
			SubjectName: subject.SubjectName,
			Duration:    subject.Duration,
			MaxPathLen:  subject.MaxPathLen,
			DERChain:    keyMaterialJSON.DERCertChain,
			PEMChain:    keyMaterialJSON.PEMCertChain,
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
	// Convert KeyMaterial to JSON format
	keyMaterialJSON, err := keyMaterial.ToJSON(includePrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert KeyMaterial to JSON format: %w", err)
	}

	data, err := json.Marshal(keyMaterialJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize KeyMaterial: %w", err)
	}
	return data, nil
}

func DeserializeKeyMaterial(data []byte) (*KeyMaterial, error) {
	var keyMaterialJSON KeyMaterialJSON
	err := json.Unmarshal(data, &keyMaterialJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize KeyMaterialJSON: %w", err)
	}

	// Convert JSON format back to KeyMaterial
	keyMaterial, err := keyMaterialJSON.ToKeyMaterial()
	if err != nil {
		return nil, fmt.Errorf("failed to convert JSON format to KeyMaterial: %w", err)
	}

	return keyMaterial, nil
}

// ToJSON converts KeyMaterial to KeyMaterialJSON with serializable representations
func (km *KeyMaterial) ToJSON(includePrivateKey bool) (*KeyMaterialJSON, error) {
	result := &KeyMaterialJSON{}
	var err error

	// Serialize private key if present and requested
	if includePrivateKey && km.PrivateKey != nil {
		result.DERPrivateKey, err = x509.MarshalPKCS8PrivateKey(km.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal private key to DER: %w", err)
		}

		result.PEMPrivateKey = pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: result.DERPrivateKey,
		})
	}

	// Serialize public key if present
	if km.PublicKey != nil {
		result.DERPublicKey, err = x509.MarshalPKIXPublicKey(km.PublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal public key to DER: %w", err)
		}

		result.PEMPublicKey = pem.EncodeToMemory(&pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: result.DERPublicKey,
		})
	}

	// Serialize certificate chain
	if len(km.CertChain) > 0 {
		result.DERCertChain = make([][]byte, len(km.CertChain))
		result.PEMCertChain = make([][]byte, len(km.CertChain))
		for i, cert := range km.CertChain {
			result.DERCertChain[i] = cert.Raw
			result.PEMCertChain[i] = pem.EncodeToMemory(&pem.Block{
				Type:  "CERTIFICATE",
				Bytes: cert.Raw,
			})
		}
	}

	// Convert subordinate CA certs pool to slices
	if km.SubordinateCACertsPool != nil {
		// Note: x509.CertPool doesn't expose certificates directly,
		// so we need to track them separately during construction
		result.DERSubordinateCACerts = [][]byte{}
		result.PEMSubordinateCACerts = [][]byte{}
	}

	// Convert root CA certs pool to slices
	if km.RootCACertsPool != nil {
		// Note: x509.CertPool doesn't expose certificates directly,
		// so we need to track them separately during construction
		result.DERRootCACertsPool = [][]byte{}
		result.PEMRootCACertsPool = [][]byte{}
	}

	return result, nil
}

// ToKeyMaterial converts KeyMaterialJSON back to KeyMaterial with crypto objects
func (kmj *KeyMaterialJSON) ToKeyMaterial() (*KeyMaterial, error) {
	result := &KeyMaterial{}
	var err error

	// Reconstruct private key from DER if present
	if len(kmj.DERPrivateKey) > 0 {
		result.PrivateKey, err = x509.ParsePKCS8PrivateKey(kmj.DERPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key from DER: %w", err)
		}
	}

	// Reconstruct public key from DER if present
	if len(kmj.DERPublicKey) > 0 {
		result.PublicKey, err = x509.ParsePKIXPublicKey(kmj.DERPublicKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key from DER: %w", err)
		}
	}

	// Reconstruct cert chain from DER chain
	if len(kmj.DERCertChain) > 0 {
		result.CertChain = make([]*x509.Certificate, len(kmj.DERCertChain))
		for i, derBytes := range kmj.DERCertChain {
			result.CertChain[i], err = x509.ParseCertificate(derBytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate %d from DER: %w", i, err)
			}
		}
	}

	// Reconstruct subordinate CA certs pool from DER
	if len(kmj.DERSubordinateCACerts) > 0 {
		result.SubordinateCACertsPool = x509.NewCertPool()
		for i, derBytes := range kmj.DERSubordinateCACerts {
			cert, err := x509.ParseCertificate(derBytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse subordinate CA certificate %d from DER: %w", i, err)
			}
			result.SubordinateCACertsPool.AddCert(cert)
		}
	}

	// Reconstruct root CA certs pool from DER
	if len(kmj.DERRootCACertsPool) > 0 {
		result.RootCACertsPool = x509.NewCertPool()
		for i, derBytes := range kmj.DERRootCACertsPool {
			cert, err := x509.ParseCertificate(derBytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse root CA certificate %d from DER: %w", i, err)
			}
			result.RootCACertsPool.AddCert(cert)
		}
	}

	return result, nil
}
