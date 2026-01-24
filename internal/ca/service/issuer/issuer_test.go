// Copyright (c) 2025 Justin Cranford

package issuer

import (
	"crypto/x509"
	"testing"
	"time"

	cryptoutilCABootstrap "cryptoutil/internal/ca/bootstrap"
	cryptoutilCACrypto "cryptoutil/internal/ca/crypto"
	cryptoutilCAIntermediate "cryptoutil/internal/ca/intermediate"
	cryptoutilCAProfileCertificate "cryptoutil/internal/ca/profile/certificate"
	cryptoutilCAProfileSubject "cryptoutil/internal/ca/profile/subject"

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
		ValidityDuration:  20 * 365 * 24 * time.Hour,
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
		ValidityDuration:  10 * 365 * 24 * time.Hour,
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
		ValidityDuration: 90 * 24 * time.Hour, // 90 days.
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
		ValidityDuration: 30 * 24 * time.Hour,
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
		ValidityDuration: 365 * 24 * time.Hour,
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
				MaxCount: 10,
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
		ValidityDuration: 90 * 24 * time.Hour,
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

func TestIssuer_Issue_InvalidRequest(t *testing.T) {
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

	tests := []struct {
		name    string
		req     *CertificateRequest
		wantErr string
	}{
		{
			name:    "nil-request",
			req:     nil,
			wantErr: "request cannot be nil",
		},
		{
			name: "no-public-key",
			req: &CertificateRequest{
				SubjectRequest: &cryptoutilCAProfileSubject.Request{
					CommonName: "test",
				},
				ValidityDuration: 24 * time.Hour,
			},
			wantErr: "public key is required",
		},
		{
			name: "zero-validity",
			req: &CertificateRequest{
				SubjectRequest: &cryptoutilCAProfileSubject.Request{
					CommonName: "test",
				},
				PublicKey:        keyPair.PublicKey,
				ValidityDuration: 0,
			},
			wantErr: "validity duration must be positive",
		},
		{
			name: "no-subject-request",
			req: &CertificateRequest{
				PublicKey:        keyPair.PublicKey,
				ValidityDuration: 24 * time.Hour,
			},
			wantErr: "subject request is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			issued, audit, err := issuer.Issue(tc.req)
			require.Error(t, err)
			require.Nil(t, issued)
			require.Nil(t, audit)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestIssuer_Issue_ValidityTruncation(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	// Create short-lived root.
	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootConfig := &cryptoutilCABootstrap.RootCAConfig{
		Name: "Short Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  180 * 24 * time.Hour, // 180 days.
		PathLenConstraint: 1,
	}

	rootCA, _, err := bootstrapper.Bootstrap(rootConfig)
	require.NoError(t, err)

	// Create issuing CA with short validity.
	provisioner := cryptoutilCAIntermediate.NewProvisioner(provider)
	intermediateConfig := &cryptoutilCAIntermediate.IntermediateCAConfig{
		Name: "Short Issuing CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  90 * 24 * time.Hour, // 90 days.
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}

	issuingCA, _, err := provisioner.Provision(intermediateConfig)
	require.NoError(t, err)

	caConfig := &IssuingCAConfig{
		Name:        "Short Issuer",
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
			CommonName: "long-validity-request",
		},
		PublicKey:        keyPair.PublicKey,
		ValidityDuration: 365 * 24 * time.Hour, // Request 1 year.
	}

	issued, _, err := issuer.Issue(req)
	require.NoError(t, err)
	require.NotNil(t, issued)

	// Certificate should not outlive issuing CA.
	require.True(t, !issued.Certificate.NotAfter.After(issuingCA.Certificate.NotAfter))
}

func TestIssuer_Issue_ChainVerification(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	// Create full chain: root -> intermediate -> end-entity.
	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootConfig := &cryptoutilCABootstrap.RootCAConfig{
		Name: "Chain Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  20 * 365 * 24 * time.Hour,
		PathLenConstraint: 2,
	}

	rootCA, _, err := bootstrapper.Bootstrap(rootConfig)
	require.NoError(t, err)

	provisioner := cryptoutilCAIntermediate.NewProvisioner(provider)
	intermediateConfig := &cryptoutilCAIntermediate.IntermediateCAConfig{
		Name: "Chain Issuing CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}

	issuingCA, _, err := provisioner.Provision(intermediateConfig)
	require.NoError(t, err)

	caConfig := &IssuingCAConfig{
		Name:        "Chain Issuer",
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
			CommonName: "chain-test.example.com",
			DNSNames:   []string{"chain-test.example.com"},
		},
		PublicKey:        keyPair.PublicKey,
		ValidityDuration: 90 * 24 * time.Hour,
	}

	issued, _, err := issuer.Issue(req)
	require.NoError(t, err)

	// Verify full chain.
	rootPool := x509.NewCertPool()
	rootPool.AddCert(rootCA.Certificate)

	intermediatePool := x509.NewCertPool()
	intermediatePool.AddCert(issuingCA.Certificate)

	opts := x509.VerifyOptions{
		Roots:         rootPool,
		Intermediates: intermediatePool,
		CurrentTime:   time.Now(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	chains, err := issued.Certificate.Verify(opts)
	require.NoError(t, err)
	require.NotEmpty(t, chains)

	// Chain should be: end-entity -> issuing CA -> root CA.
	require.Len(t, chains[0], 3)
}
