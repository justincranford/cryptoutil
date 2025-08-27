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

// KeyMaterialEncoded contains serializable DER/PEM representations of KeyMaterialDecoded
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
	SubjectName string
	IssuerName  string
	Duration    time.Duration
	KeyMaterial KeyMaterialDecoded

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

// SerializeSubjects serializes a slice of Subject to JSON bytes
func SerializeSubjects(subjects []Subject, includePrivateKey bool) ([][]byte, error) {
	keyMaterialJSONs := make([][]byte, len(subjects))

	for i, subject := range subjects {
		// Convert KeyMaterialDecoded to JSON format
		keyMaterialJSON, err := subject.KeyMaterial.ToJSON(includePrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert KeyMaterialDecoded to JSON format for subject %d: %w", i, err)
		}

		// Serialize the KeyMaterialEncoded to bytes
		jsonBytes, err := json.Marshal(keyMaterialJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize KeyMaterialEncoded for subject %d: %w", i, err)
		}

		keyMaterialJSONs[i] = jsonBytes
	}

	return keyMaterialJSONs, nil
}

// DeserializeSubjects deserializes JSON bytes to a slice of KeyMaterialDecoded
// Note: This only returns the KeyMaterialDecoded parts since subject metadata (name, duration, maxPathLen)
// is not included in the serialized data. To rebuild full Subject, caller must provide metadata separately.
func DeserializeSubjects(keyMaterialJSONBytes [][]byte) ([]KeyMaterialDecoded, error) {
	keyMaterials := make([]KeyMaterialDecoded, len(keyMaterialJSONBytes))
	for i, jsonBytes := range keyMaterialJSONBytes {
		var keyMaterialJSON KeyMaterialEncoded
		err := json.Unmarshal(jsonBytes, &keyMaterialJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize KeyMaterialEncoded for item %d: %w", i, err)
		}

		keyMaterial, err := keyMaterialJSON.ToKeyMaterialDecoded()
		if err != nil {
			return nil, fmt.Errorf("failed to convert KeyMaterialEncoded to KeyMaterialDecoded for item %d: %w", i, err)
		}

		keyMaterials[i] = *keyMaterial
	}

	return keyMaterials, nil
}

func SerializeKeyMaterial(keyMaterial *KeyMaterialDecoded, includePrivateKey bool) ([]byte, error) {
	// Convert KeyMaterialDecoded to JSON format
	keyMaterialJSON, err := keyMaterial.ToJSON(includePrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to convert KeyMaterialDecoded to JSON format: %w", err)
	}

	data, err := json.Marshal(keyMaterialJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize KeyMaterialDecoded: %w", err)
	}
	return data, nil
}

func DeserializeKeyMaterial(data []byte) (*KeyMaterialDecoded, error) {
	var keyMaterialJSON KeyMaterialEncoded
	err := json.Unmarshal(data, &keyMaterialJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize KeyMaterialEncoded: %w", err)
	}

	// Convert JSON format back to KeyMaterialDecoded
	keyMaterial, err := keyMaterialJSON.ToKeyMaterialDecoded()
	if err != nil {
		return nil, fmt.Errorf("failed to convert JSON format to KeyMaterialDecoded: %w", err)
	}

	return keyMaterial, nil
}

// ToJSON converts KeyMaterialDecoded to KeyMaterialEncoded with serializable representations
func (km *KeyMaterialDecoded) ToJSON(includePrivateKey bool) (*KeyMaterialEncoded, error) {
	result := &KeyMaterialEncoded{}
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

// ToKeyMaterialDecoded converts KeyMaterialEncoded back to KeyMaterialDecoded with crypto objects
func (kmj *KeyMaterialEncoded) ToKeyMaterialDecoded() (*KeyMaterialDecoded, error) {
	result := &KeyMaterialDecoded{}
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

// CreateCASubjects creates a chain of CA subjects with the specified prefix and count
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
			KeyMaterial: KeyMaterialDecoded{
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
			previousCACert = previousSubject.KeyMaterial.CertChain[0]
		}

		currentCACertTemplate, err := CertificateTemplateCA(previousSubject.IssuerName, currentSubject.SubjectName, currentSubject.Duration, currentSubject.CASubject.MaxPathLen)
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

		currentSubject.KeyMaterial.RootCACertsPool = previousSubject.KeyMaterial.RootCACertsPool.Clone()
		currentSubject.KeyMaterial.SubordinateCACertsPool = previousSubject.KeyMaterial.SubordinateCACertsPool.Clone()
		if i == 0 {
			currentSubject.KeyMaterial.RootCACertsPool.AddCert(cert)
		} else {
			currentSubject.KeyMaterial.SubordinateCACertsPool.AddCert(cert)
		}

		subjects[i] = currentSubject
	}
	return subjects, nil
}

// CreateEndEntitySubject creates an end entity subject with the specified parameters
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
		KeyMaterial: KeyMaterialDecoded{
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

	cert, _, _, err := SignCertificate(issuingCA.KeyMaterial.CertChain[0], issuingCA.KeyMaterial.PrivateKey, endEntityCertTemplate, endEntitySubject.KeyMaterial.PublicKey, x509.ECDSAWithSHA256)
	if err != nil {
		return Subject{}, fmt.Errorf("failed to sign end entity certificate for %s: %w", subjectName, err)
	}

	endEntitySubject.KeyMaterial.CertChain = append([]*x509.Certificate{cert}, issuingCA.KeyMaterial.CertChain...)
	endEntitySubject.KeyMaterial.RootCACertsPool = issuingCA.KeyMaterial.RootCACertsPool.Clone()
	endEntitySubject.KeyMaterial.SubordinateCACertsPool = issuingCA.KeyMaterial.SubordinateCACertsPool.Clone()

	return endEntitySubject, nil
}

// BuildTLSCertificate converts a Subject to a tls.Certificate suitable for TLS connections
func BuildTLSCertificate(endEntitySubject Subject) (tls.Certificate, *x509.CertPool, error) {
	if len(endEntitySubject.KeyMaterial.CertChain) == 0 {
		return tls.Certificate{}, nil, fmt.Errorf("certificate chain is empty")
	}
	if endEntitySubject.KeyMaterial.PrivateKey == nil {
		return tls.Certificate{}, nil, fmt.Errorf("private key is nil")
	}
	if endEntitySubject.KeyMaterial.RootCACertsPool == nil {
		return tls.Certificate{}, nil, fmt.Errorf("root CA certs pool is nil")
	}

	// Convert certificate chain to DER format for TLS
	derCertChain := make([][]byte, len(endEntitySubject.KeyMaterial.CertChain))
	for i, cert := range endEntitySubject.KeyMaterial.CertChain {
		derCertChain[i] = cert.Raw
	}

	return tls.Certificate{Certificate: derCertChain, PrivateKey: endEntitySubject.KeyMaterial.PrivateKey, Leaf: endEntitySubject.KeyMaterial.CertChain[0]}, endEntitySubject.KeyMaterial.RootCACertsPool, nil
}
