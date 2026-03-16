// Copyright (c) 2025 Justin Cranford

package issuer

import (
	"crypto/x509"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	cryptoutilCABootstrap "cryptoutil/internal/apps/pki/ca/bootstrap"
	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
	cryptoutilCAIntermediate "cryptoutil/internal/apps/pki/ca/intermediate"
	cryptoutilCAProfileCertificate "cryptoutil/internal/apps/pki/ca/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/apps/pki/ca/profile/subject"

	"github.com/stretchr/testify/require"
)

func createTestIssuingCA(t *testing.T, provider cryptoutilCACrypto.Provider) *cryptoutilCAIntermediate.IntermediateCA {
	t.Helper()

	// Create root CA.
	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootConfig := &cryptoutilCABootstrap.RootCAConfig{
		Name: "Test Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  cryptoutilSharedMagic.MaxErrorDisplay * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour,
		PathLenConstraint: 2,
	}

	rootCA, _, err := bootstrapper.Bootstrap(rootConfig)
	require.NoError(t, err)

	// Create intermediate/issuing CA.
	provisioner := cryptoutilCAIntermediate.NewProvisioner(provider)
	intermediateConfig := &cryptoutilCAIntermediate.IntermediateCAConfig{
		Name: "Test Issuing CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  cryptoutilSharedMagic.JoseJADefaultMaxMaterials * cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}

	issuingCA, _, err := provisioner.Provision(intermediateConfig)
	require.NoError(t, err)

	return issuingCA
}

func TestNewIssuer_Valid(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	caConfig := &IssuingCAConfig{
		Name:        "Test Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuer, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)
	require.NotNil(t, issuer)
}

func TestNewIssuer_InvalidConfig(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	tests := []struct {
		name    string
		config  *IssuingCAConfig
		wantErr string
	}{
		{
			name:    "nil-config",
			config:  nil,
			wantErr: "config cannot be nil",
		},
		{
			name: "empty-name",
			config: &IssuingCAConfig{
				Certificate: issuingCA.Certificate,
				PrivateKey:  issuingCA.PrivateKey,
			},
			wantErr: "CA name is required",
		},
		{
			name: "no-certificate",
			config: &IssuingCAConfig{
				Name:       "Test",
				PrivateKey: issuingCA.PrivateKey,
			},
			wantErr: "CA certificate is required",
		},
		{
			name: "no-private-key",
			config: &IssuingCAConfig{
				Name:        "Test",
				Certificate: issuingCA.Certificate,
			},
			wantErr: "CA private key is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			issuer, err := NewIssuer(provider, tc.config)
			require.Error(t, err)
			require.Nil(t, issuer)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestIssuer_Issue_TLSServer(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	caConfig := &IssuingCAConfig{
		Name:        "Test Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuer, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)

	// Generate subscriber key.
	keyPair, err := provider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilCACrypto.KeyTypeECDSA,
		ECDSACurve: "P-256",
	})
	require.NoError(t, err)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName: "www.example.com",
			DNSNames:   []string{"www.example.com", "example.com"},
		},
		PublicKey:        keyPair.PublicKey,
		ValidityDuration: cryptoutilSharedMagic.StrictCertificateMaxAgeDays * cryptoutilSharedMagic.HoursPerDay * time.Hour, // 90 days.
	}

	issued, audit, err := issuer.Issue(req)
	require.NoError(t, err)
	require.NotNil(t, issued)
	require.NotNil(t, audit)

	// Verify issued certificate.
	require.NotNil(t, issued.Certificate)
	require.NotEmpty(t, issued.CertificatePEM)
	require.NotEmpty(t, issued.ChainPEM)
	require.NotEmpty(t, issued.SerialNumber)
	require.NotEmpty(t, issued.Fingerprint)

	cert := issued.Certificate
	require.Equal(t, "www.example.com", cert.Subject.CommonName)
	require.Contains(t, cert.DNSNames, "www.example.com")
	require.Contains(t, cert.DNSNames, "example.com")
	require.False(t, cert.IsCA)

	// Verify audit entry.
	require.Equal(t, "certificate_issuance", audit.Operation)
	require.NotEmpty(t, audit.SANs)
	require.Equal(t, "ECDSA", audit.KeyAlgorithm)
}

func TestIssuer_Issue_WithIPAddresses(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	caConfig := &IssuingCAConfig{
		Name:        "Test Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuer, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)

	keyPair, err := provider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilCACrypto.KeyTypeECDSA,
		ECDSACurve: "P-256",
	})
	require.NoError(t, err)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName:  "internal-server",
			IPAddresses: []string{"192.168.1.100", "10.0.0.1"},
		},
		PublicKey:        keyPair.PublicKey,
		ValidityDuration: cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	issued, audit, err := issuer.Issue(req)
	require.NoError(t, err)
	require.NotNil(t, issued)

	cert := issued.Certificate
	require.Len(t, cert.IPAddresses, 2)
	require.Equal(t, "192.168.1.100", cert.IPAddresses[0].String())
	require.Equal(t, "10.0.0.1", cert.IPAddresses[1].String())

	// Check SANs in audit.
	require.Contains(t, audit.SANs, "IP:192.168.1.100")
	require.Contains(t, audit.SANs, "IP:10.0.0.1")
}

func TestIssuer_Issue_WithEmail(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	caConfig := &IssuingCAConfig{
		Name:        "Test Issuer",
		Certificate: issuingCA.Certificate,
		PrivateKey:  issuingCA.PrivateKey,
	}

	issuer, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)

	keyPair, err := provider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilCACrypto.KeyTypeECDSA,
		ECDSACurve: "P-256",
	})
	require.NoError(t, err)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName:     "John Doe",
			EmailAddresses: []string{"john.doe@example.com"},
		},
		PublicKey:        keyPair.PublicKey,
		ValidityDuration: cryptoutilSharedMagic.TLSTestEndEntityCertValidity1Year * cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	issued, audit, err := issuer.Issue(req)
	require.NoError(t, err)
	require.NotNil(t, issued)

	cert := issued.Certificate
	require.Contains(t, cert.EmailAddresses, "john.doe@example.com")
	require.Contains(t, audit.SANs, "email:john.doe@example.com")
}

func TestIssuer_Issue_WithProfile(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	issuingCA := createTestIssuingCA(t, provider)

	subjectProfile := &cryptoutilCAProfileSubject.Profile{
		Name: "tls-server",
		Subject: cryptoutilCAProfileSubject.DN{
			Organization:       []string{"Example Corp"},
			OrganizationalUnit: []string{"Web Services"},
			Country:            []string{"US"},
		},
		SubjectAltNames: cryptoutilCAProfileSubject.SANConfig{
			DNSNames: cryptoutilCAProfileSubject.SANPatterns{
				Allowed:  true,
				MaxCount: cryptoutilSharedMagic.JoseJADefaultMaxMaterials,
			},
		},
	}

	certProfile := &cryptoutilCAProfileCertificate.Profile{
		Name: "tls-server",
		KeyUsage: cryptoutilCAProfileCertificate.KeyUsageConfig{
			DigitalSignature: true,
			KeyEncipherment:  true,
		},
		ExtendedKeyUsage: cryptoutilCAProfileCertificate.ExtKeyUsageConfig{
			ServerAuth: true,
		},
		Validity: cryptoutilCAProfileCertificate.ValidityConfig{
			Duration:    "2160h", // 90 days.
			MaxDuration: "9552h", // 398 days.
		},
	}

	caConfig := &IssuingCAConfig{
		Name:           "Test Issuer",
		Certificate:    issuingCA.Certificate,
		PrivateKey:     issuingCA.PrivateKey,
		SubjectProfile: subjectProfile,
		CertProfile:    certProfile,
	}

	issuer, err := NewIssuer(provider, caConfig)
	require.NoError(t, err)

	keyPair, err := provider.GenerateKeyPair(cryptoutilCACrypto.KeySpec{
		Type:       cryptoutilCACrypto.KeyTypeECDSA,
		ECDSACurve: "P-256",
	})
	require.NoError(t, err)

	req := &CertificateRequest{
		SubjectRequest: &cryptoutilCAProfileSubject.Request{
			CommonName: "api.example.com",
			DNSNames:   []string{"api.example.com"},
		},
		PublicKey:        keyPair.PublicKey,
		ValidityDuration: cryptoutilSharedMagic.StrictCertificateMaxAgeDays * cryptoutilSharedMagic.HoursPerDay * time.Hour,
	}

	issued, audit, err := issuer.Issue(req)
	require.NoError(t, err)
	require.NotNil(t, issued)

	cert := issued.Certificate

	// Verify profile was applied.
	require.Equal(t, []string{"Example Corp"}, cert.Subject.Organization)
	require.Equal(t, []string{"Web Services"}, cert.Subject.OrganizationalUnit)
	require.Equal(t, []string{"US"}, cert.Subject.Country)

	// Verify key usage.
	require.True(t, cert.KeyUsage&x509.KeyUsageDigitalSignature != 0)
	require.True(t, cert.KeyUsage&x509.KeyUsageKeyEncipherment != 0)
	require.Contains(t, cert.ExtKeyUsage, x509.ExtKeyUsageServerAuth)

	// Verify audit includes profile names.
	require.Equal(t, "tls-server", audit.ProfileName)
	require.Equal(t, "tls-server", audit.SubjectProfile)
}
