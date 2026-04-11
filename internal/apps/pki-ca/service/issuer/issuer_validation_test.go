// Copyright (c) 2025 Justin Cranford

package issuer

import (
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilCABootstrap "cryptoutil/internal/apps/pki-ca/bootstrap"
	cryptoutilCACrypto "cryptoutil/internal/apps/pki-ca/crypto"
	cryptoutilCAIntermediate "cryptoutil/internal/apps/pki-ca/intermediate"
	cryptoutilCAProfileCertificate "cryptoutil/internal/apps/pki-ca/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/apps/pki-ca/profile/subject"

	"github.com/stretchr/testify/require"
)

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
		ValidityDuration: cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour, // 365 days > 30 day max
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
		ValidityDuration: cryptoutilSharedMagic.StrictCertificateMaxAgeDays * cryptoutilSharedMagic.HoursPerDay * time.Hour,
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
		ValidityDuration: cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * cryptoutilSharedMagic.HoursPerDay * time.Hour,
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
			RSABits: cryptoutilSharedMagic.DefaultMetricsBatchSize,
		},
		ValidityDuration:  cryptoutilSharedMagic.JoseJADefaultMaxMaterials * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour,
		PathLenConstraint: 1,
	}
	rootCA, _, err := bootstrapper.Bootstrap(rootConfig)
	require.NoError(t, err)

	provisioner := cryptoutilCAIntermediate.NewProvisioner(provider)
	intermediateConfig := &cryptoutilCAIntermediate.IntermediateCAConfig{
		Name: "RSA Issuing CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:    cryptoutilCACrypto.KeyTypeRSA,
			RSABits: cryptoutilSharedMagic.DefaultMetricsBatchSize,
		},
		ValidityDuration:  cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour,
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
		ValidityDuration: cryptoutilSharedMagic.StrictCertificateMaxAgeDays * cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	issued, audit, issueErr := issuerObj.Issue(req)
	require.NoError(t, issueErr)
	require.NotNil(t, issued)
	require.NotNil(t, audit)
}
