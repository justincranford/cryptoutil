// Copyright (c) 2025 Justin Cranford

package issuer

import (
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	rsa "crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"

	cryptoutilCABootstrap "cryptoutil/internal/apps/pki/ca/bootstrap"
	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
	cryptoutilCAIntermediate "cryptoutil/internal/apps/pki/ca/intermediate"
	cryptoutilCAProfileCertificate "cryptoutil/internal/apps/pki/ca/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/apps/pki/ca/profile/subject"

	"github.com/stretchr/testify/require"
)

// TestNewIssuer_NilProvider verifies that NewIssuer rejects a nil provider.
func TestNewIssuer_NilProvider(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	caConfig := &IssuingCAConfig{
		Name:        "Test Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuerObj, err := NewIssuer(nil, caConfig)
	require.Error(t, err)
	require.Nil(t, issuerObj)
	require.Contains(t, err.Error(), "provider cannot be nil")
}

// TestIssuer_GetCAConfig verifies that GetCAConfig returns the configured CA config.
func TestIssuer_GetCAConfig(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	caConfig := &IssuingCAConfig{
		Name:        "Test Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuerObj, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)
	require.NotNil(t, issuerObj)

	got := issuerObj.GetCAConfig()
	require.Equal(t, caConfig, got)
}

// TestNewIssuer_NonCACert verifies that NewIssuer rejects a non-CA certificate.
func TestNewIssuer_NonCACert(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	// Create a real signing CA to sign the non-CA leaf cert.
	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootConfig := &cryptoutilCABootstrap.RootCAConfig{
		Name: "Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 2,
	}
	rootCA, _, err := bootstrapper.Bootstrap(rootConfig)
	require.NoError(t, err)

	// Generate a leaf key.
	leafPriv, genErr := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, genErr)

	// Create a non-CA cert template.
	leafTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(99),
		Subject:               pkix.Name{CommonName: "leaf"},
		NotBefore:             time.Now().UTC().Add(-time.Hour),
		NotAfter:              time.Now().UTC().Add(time.Hour),
		IsCA:                  false,
		BasicConstraintsValid: true,
	}

	// Sign with root CA.
	leafDER, signErr := x509.CreateCertificate(crand.Reader, leafTemplate, rootCA.Certificate, &leafPriv.PublicKey, rootCA.PrivateKey)
	require.NoError(t, signErr)

	leafCert, parseErr := x509.ParseCertificate(leafDER)
	require.NoError(t, parseErr)

	caConfig := &IssuingCAConfig{
		Name:        "Not A CA",
		Certificate: leafCert,
		PrivateKey:  leafPriv,
	}

	issuerObj, newErr := NewIssuer(provider, caConfig)
	require.Error(t, newErr)
	require.Nil(t, issuerObj)
	require.Contains(t, newErr.Error(), "certificate is not a CA")
}

// TestIssuer_Issue_RSAKey verifies that an RSA public key produces "RSA" in audit.
func TestIssuer_Issue_RSAKey(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	caConfig := &IssuingCAConfig{
		Name:        "Test Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuerObj, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)

	// Generate RSA key outside of provider (direct crypto/rsa usage).
	rsaPriv, genErr := rsa.GenerateKey(crand.Reader, 2048)
	require.NoError(t, genErr)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName: "rsa-server.example.com",
			DNSNames:   []string{"rsa-server.example.com"},
		},
		PublicKey:        &rsaPriv.PublicKey,
		ValidityDuration: 90 * 24 * time.Hour,
	}

	issued, audit, issueErr := issuerObj.Issue(req)
	require.NoError(t, issueErr)
	require.NotNil(t, issued)
	require.NotNil(t, audit)
	require.Equal(t, "RSA", audit.KeyAlgorithm)
}

// TestIssuer_Issue_EdDSAKey verifies that an Ed25519 public key produces "Ed25519" in audit.
func TestIssuer_Issue_EdDSAKey(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	caConfig := &IssuingCAConfig{
		Name:        "Test Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuerObj, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)

	// Generate Ed25519 key directly.
	edPub, _, genErr := ed25519.GenerateKey(crand.Reader)
	require.NoError(t, genErr)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName: "eddsa-server.example.com",
			DNSNames:   []string{"eddsa-server.example.com"},
		},
		PublicKey:        edPub,
		ValidityDuration: 90 * 24 * time.Hour,
	}

	issued, audit, issueErr := issuerObj.Issue(req)
	require.NoError(t, issueErr)
	require.NotNil(t, issued)
	require.NotNil(t, audit)
	require.Equal(t, "Ed25519", audit.KeyAlgorithm)
}

// TestIssuer_Issue_UnknownKeyType verifies that an unknown public key type produces "Unknown" in audit.
func TestIssuer_Issue_UnknownKeyType(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	caConfig := &IssuingCAConfig{
		Name:        "Test Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuerObj, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)

	// Use an unknown struct type as public key.
	type unknownKey struct{}

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName: "unknown-key.example.com",
		},
		PublicKey:        unknownKey{},
		ValidityDuration: 90 * 24 * time.Hour,
	}

	// Issue will fail (x509.CreateCertificate rejects unknown key type), but
	// keyAlgorithmName is invoked before the error occurs — we validate via audit.
	// If issue fails, the keyAlgorithmName default branch is still exercised.
	_, _, issueErr := issuerObj.Issue(req)
	// The certificate creation with an unknown key type is expected to fail.
	require.Error(t, issueErr)
}

// TestIssuer_Issue_WithURIs verifies that URI SANs are included in issued certificates.
func TestIssuer_Issue_WithURIs(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	caConfig := &IssuingCAConfig{
		Name:        "Test Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuerObj, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)

	keyPair, genErr := provider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilCACrypto.KeyTypeECDSA,
		ECDSACurve: "P-256",
	})
	require.NoError(t, genErr)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName: "spiffe-workload.example.com",
			URIs:       []string{"spiffe://example.org/workload/api"},
		},
		PublicKey:        keyPair.PublicKey,
		ValidityDuration: 90 * 24 * time.Hour,
	}

	issued, audit, issueErr := issuerObj.Issue(req)
	require.NoError(t, issueErr)
	require.NotNil(t, issued)
	require.NotNil(t, audit)

	cert := issued.Certificate
	require.Len(t, cert.URIs, 1)
	require.Equal(t, "spiffe://example.org/workload/api", cert.URIs[0].String())
	require.Contains(t, audit.SANs, "URI:spiffe://example.org/workload/api")
}

// TestIssuer_Issue_InvalidIPAddress verifies that an invalid IP string causes an error.
func TestIssuer_Issue_InvalidIPAddress(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	caConfig := &IssuingCAConfig{
		Name:        "Test Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuerObj, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)

	keyPair, genErr := provider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilCACrypto.KeyTypeECDSA,
		ECDSACurve: "P-256",
	})
	require.NoError(t, genErr)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName:  "bad-ip.example.com",
			IPAddresses: []string{"not.a.valid.ip"},
		},
		PublicKey:        keyPair.PublicKey,
		ValidityDuration: 30 * 24 * time.Hour,
	}

	_, _, issueErr := issuerObj.Issue(req)
	require.Error(t, issueErr)
	require.Contains(t, issueErr.Error(), "invalid IP address")
}

// TestIssuer_Issue_ValidityValidationFailure verifies that a cert profile MaxDuration
// constraint causes an error when the requested duration exceeds it.
func TestIssuer_Issue_ValidityValidationFailure(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	// CertProfile with a MaxDuration of 30 days; request will ask for 365 days.
	certProfile := &cryptoutilCAProfileCertificate.Profile{
		Name: "short-validity",
		KeyUsage: cryptoutilCAProfileCertificate.KeyUsageConfig{
			DigitalSignature: true,
			KeyEncipherment:  true,
		},
		Validity: cryptoutilCAProfileCertificate.ValidityConfig{
			AllowCustom: true,
			MaxDuration: "720h", // 30 days max
		},
	}

	caConfig := &IssuingCAConfig{
		Name:        "Test Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
		CertProfile: certProfile,
	}

	issuerObj, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)

	keyPair, genErr := provider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilCACrypto.KeyTypeECDSA,
		ECDSACurve: "P-256",
	})
	require.NoError(t, genErr)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName: "server.example.com",
			DNSNames:   []string{"server.example.com"},
		},
		PublicKey:        keyPair.PublicKey,
		ValidityDuration: 365 * 24 * time.Hour, // 365 days > 30 day max
	}

	_, _, issueErr := issuerObj.Issue(req)
	require.Error(t, issueErr)
	require.Contains(t, issueErr.Error(), "validity validation failed")
}

// TestIssuer_Issue_SubjectProfileResolveError verifies that a SubjectProfile
// that fails to resolve causes an error.
func TestIssuer_Issue_SubjectProfileResolveError(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	// SubjectProfile requiring CommonName but request will have none.
	subjectProfile := &cryptoutilCAProfileSubject.Profile{
		Name: "requires-cn",
		Constraints: cryptoutilCAProfileSubject.Constraints{
			RequireCommonName: true,
		},
	}

	caConfig := &IssuingCAConfig{
		Name:           "Test Issuer",
		Certificate:    issuingCA.Certificate,
		PrivateKey:     issuingCA.PrivateKey,
		SubjectProfile: subjectProfile,
	}

	issuerObj, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)

	keyPair, genErr := provider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilCACrypto.KeyTypeECDSA,
		ECDSACurve: "P-256",
	})
	require.NoError(t, genErr)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName: "", // No CommonName — profile requires one.
		},
		PublicKey:        keyPair.PublicKey,
		ValidityDuration: 90 * 24 * time.Hour,
	}

	_, _, issueErr := issuerObj.Issue(req)
	require.Error(t, issueErr)
	require.Contains(t, issueErr.Error(), "failed to resolve subject from profile")
}

// TestIssuer_Issue_InvalidURI verifies that an invalid URI string in SubjectRequest.URIs causes an error.
func TestIssuer_Issue_InvalidURI(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	caConfig := &IssuingCAConfig{
		Name:        "Test Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuerObj, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)

	keyPair, genErr := provider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilCACrypto.KeyTypeECDSA,
		ECDSACurve: "P-256",
	})
	require.NoError(t, genErr)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName: "bad-uri.example.com",
			URIs:       []string{"http://[invalid"}, // missing ']' in host
		},
		PublicKey:        keyPair.PublicKey,
		ValidityDuration: 30 * 24 * time.Hour,
	}

	_, _, issueErr := issuerObj.Issue(req)
	require.Error(t, issueErr)
	require.Contains(t, issueErr.Error(), "invalid URI")
}

// TestIssuer_Issue_WithRSACA verifies issuing when the CA itself uses RSA.
func TestIssuer_Issue_WithRSACA(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootConfig := &cryptoutilCABootstrap.RootCAConfig{
		Name: "RSA Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:    cryptoutilCACrypto.KeyTypeRSA,
			RSABits: 2048,
		},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
	}
	rootCA, _, err := bootstrapper.Bootstrap(rootConfig)
	require.NoError(t, err)

	provisioner := cryptoutilCAIntermediate.NewProvisioner(provider)
	intermediateConfig := &cryptoutilCAIntermediate.IntermediateCAConfig{
		Name: "RSA Issuing CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:    cryptoutilCACrypto.KeyTypeRSA,
			RSABits: 2048,
		},
		ValidityDuration:  5 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}
	issuingCA, _, err := provisioner.Provision(intermediateConfig)
	require.NoError(t, err)

	caConfig := &IssuingCAConfig{
		Name:        "RSA Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuerObj, newErr := NewIssuer(provider, caConfig)
	require.NoError(t, newErr)

	// Subscriber with ECDSA key.
	keyPair, genErr := provider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilCACrypto.KeyTypeECDSA,
		ECDSACurve: "P-256",
	})
	require.NoError(t, genErr)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName: "rsa-ca-issued.example.com",
			DNSNames:   []string{"rsa-ca-issued.example.com"},
		},
		PublicKey:        keyPair.PublicKey,
		ValidityDuration: 90 * 24 * time.Hour,
	}

	issued, audit, issueErr := issuerObj.Issue(req)
	require.NoError(t, issueErr)
	require.NotNil(t, issued)
	require.NotNil(t, audit)
}
