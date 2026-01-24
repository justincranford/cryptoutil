// Copyright (c) 2025 Justin Cranford

package intermediate

import (
	"crypto/x509"
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilCABootstrap "cryptoutil/internal/ca/bootstrap"
	cryptoutilCACrypto "cryptoutil/internal/ca/crypto"
	cryptoutilCAProfileSubject "cryptoutil/internal/ca/profile/subject"
	cryptoutilMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func createTestRootCA(t *testing.T, provider cryptoutilCACrypto.Provider) *cryptoutilCABootstrap.RootCA {
	t.Helper()

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

	return rootCA
}

func TestProvisioner_Provision_ECDSA(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	rootCA := createTestRootCA(t, provider)

	provisioner := NewProvisioner(provider)

	config := &IntermediateCAConfig{
		Name: "Test Intermediate CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}

	intermediateCA, audit, err := provisioner.Provision(config)
	require.NoError(t, err)
	require.NotNil(t, intermediateCA)
	require.NotNil(t, audit)

	// Verify intermediate CA.
	require.Equal(t, "Test Intermediate CA", intermediateCA.Name)
	require.NotNil(t, intermediateCA.Certificate)
	require.NotNil(t, intermediateCA.PrivateKey)
	require.NotNil(t, intermediateCA.PublicKey)
	require.NotEmpty(t, intermediateCA.CertificatePEM)
	require.NotEmpty(t, intermediateCA.CertificateChainPEM)

	// Verify certificate properties.
	cert := intermediateCA.Certificate
	require.True(t, cert.IsCA)
	require.Equal(t, 1, cert.MaxPathLen)
	require.Equal(t, "Test Intermediate CA", cert.Subject.CommonName)
	require.Equal(t, "Test Root CA", cert.Issuer.CommonName)

	// Verify audit entry.
	require.Equal(t, "intermediate_ca_provision", audit.Operation)
	require.Equal(t, "Test Intermediate CA", audit.CAName)
	require.Equal(t, "Test Root CA", audit.IssuerName)
	require.NotEmpty(t, audit.SerialNumber)
	require.NotEmpty(t, audit.Fingerprint)
	require.Equal(t, "ECDSA", audit.KeyAlgorithm)
}

func TestProvisioner_Provision_RSA(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	// Create RSA root CA.
	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootConfig := &cryptoutilCABootstrap.RootCAConfig{
		Name: "RSA Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:    cryptoutilCACrypto.KeyTypeRSA,
			RSABits: cryptoutilCACrypto.MinRSAKeyBits,
		},
		ValidityDuration:  15 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
	}

	rootCA, _, err := bootstrapper.Bootstrap(rootConfig)
	require.NoError(t, err)

	provisioner := NewProvisioner(provider)

	config := &IntermediateCAConfig{
		Name: "RSA Intermediate CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:    cryptoutilCACrypto.KeyTypeRSA,
			RSABits: cryptoutilCACrypto.MinRSAKeyBits,
		},
		ValidityDuration:  5 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}

	intermediateCA, audit, err := provisioner.Provision(config)
	require.NoError(t, err)
	require.NotNil(t, intermediateCA)

	require.Equal(t, "RSA Intermediate CA", intermediateCA.Name)
	require.True(t, intermediateCA.Certificate.IsCA)
	require.True(t, intermediateCA.Certificate.MaxPathLenZero)
	require.Equal(t, "RSA", audit.KeyAlgorithm)
}

func TestProvisioner_Provision_WithSubjectProfile(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	rootCA := createTestRootCA(t, provider)

	subjectProfile := &cryptoutilCAProfileSubject.Profile{
		Name: "enterprise-intermediate",
		Subject: cryptoutilCAProfileSubject.DN{
			Organization:       []string{"Example Corp"},
			OrganizationalUnit: []string{"PKI Operations"},
			Country:            []string{"US"},
			State:              []string{"New York"},
			Locality:           []string{"New York City"},
		},
	}

	provisioner := NewProvisioner(provider)

	config := &IntermediateCAConfig{
		Name:           "Enterprise Intermediate CA",
		SubjectProfile: subjectProfile,
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-384",
		},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}

	intermediateCA, audit, err := provisioner.Provision(config)
	require.NoError(t, err)
	require.NotNil(t, intermediateCA)

	// Verify subject DN.
	cert := intermediateCA.Certificate
	require.Equal(t, "Enterprise Intermediate CA", cert.Subject.CommonName)
	require.Equal(t, []string{"Example Corp"}, cert.Subject.Organization)
	require.Equal(t, []string{"PKI Operations"}, cert.Subject.OrganizationalUnit)
	require.Equal(t, []string{"US"}, cert.Subject.Country)
	require.Equal(t, []string{"New York"}, cert.Subject.Province)
	require.Equal(t, []string{"New York City"}, cert.Subject.Locality)

	require.NotEmpty(t, audit.SubjectDN)
}

func TestProvisioner_Provision_WithPersistence(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	rootCA := createTestRootCA(t, provider)

	provisioner := NewProvisioner(provider)

	config := &IntermediateCAConfig{
		Name: "Persisted Intermediate CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  5 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
		OutputDir:         tempDir,
	}

	intermediateCA, _, err := provisioner.Provision(config)
	require.NoError(t, err)
	require.NotNil(t, intermediateCA)

	// Verify certificate file exists.
	certPath := filepath.Join(tempDir, "Persisted Intermediate CA.crt")
	certData, err := os.ReadFile(certPath)
	require.NoError(t, err)
	require.NotEmpty(t, certData)
	require.Contains(t, string(certData), "BEGIN CERTIFICATE")

	// Verify chain file exists.
	chainPath := filepath.Join(tempDir, "Persisted Intermediate CA-chain.crt")
	chainData, err := os.ReadFile(chainPath)
	require.NoError(t, err)
	require.NotEmpty(t, chainData)
	// Chain should contain 2 certificates.
	require.Contains(t, string(chainData), "BEGIN CERTIFICATE")

	// Verify key file exists.
	keyPath := filepath.Join(tempDir, "Persisted Intermediate CA.key")
	keyData, err := os.ReadFile(keyPath)
	require.NoError(t, err)
	require.NotEmpty(t, keyData)
	require.Contains(t, string(keyData), "BEGIN "+cryptoutilMagic.StringPEMTypePKCS8PrivateKey) // pragma: allowlist secret
}

func TestProvisioner_Provision_InvalidConfig(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	rootCA := createTestRootCA(t, provider)

	provisioner := NewProvisioner(provider)

	tests := []struct {
		name    string
		config  *IntermediateCAConfig
		wantErr string
	}{
		{
			name:    "nil-config",
			config:  nil,
			wantErr: "config cannot be nil",
		},
		{
			name: "empty-name",
			config: &IntermediateCAConfig{
				KeySpec: cryptoutilCACrypto.KeySpec{
					Type:       cryptoutilCACrypto.KeyTypeECDSA,
					ECDSACurve: "P-256",
				},
				ValidityDuration:  1 * 365 * 24 * time.Hour,
				IssuerCertificate: rootCA.Certificate,
				IssuerPrivateKey:  rootCA.PrivateKey,
			},
			wantErr: "CA name is required",
		},
		{
			name: "zero-validity",
			config: &IntermediateCAConfig{
				Name: "Test CA",
				KeySpec: cryptoutilCACrypto.KeySpec{
					Type:       cryptoutilCACrypto.KeyTypeECDSA,
					ECDSACurve: "P-256",
				},
				ValidityDuration:  0,
				IssuerCertificate: rootCA.Certificate,
				IssuerPrivateKey:  rootCA.PrivateKey,
			},
			wantErr: "validity duration must be positive",
		},
		{
			name: "negative-path-len",
			config: &IntermediateCAConfig{
				Name: "Test CA",
				KeySpec: cryptoutilCACrypto.KeySpec{
					Type:       cryptoutilCACrypto.KeyTypeECDSA,
					ECDSACurve: "P-256",
				},
				ValidityDuration:  1 * 365 * 24 * time.Hour,
				PathLenConstraint: -1,
				IssuerCertificate: rootCA.Certificate,
				IssuerPrivateKey:  rootCA.PrivateKey,
			},
			wantErr: "path length constraint cannot be negative",
		},
		{
			name: "no-issuer-cert",
			config: &IntermediateCAConfig{
				Name: "Test CA",
				KeySpec: cryptoutilCACrypto.KeySpec{
					Type:       cryptoutilCACrypto.KeyTypeECDSA,
					ECDSACurve: "P-256",
				},
				ValidityDuration: 1 * 365 * 24 * time.Hour,
				IssuerPrivateKey: rootCA.PrivateKey,
			},
			wantErr: "issuer certificate is required",
		},
		{
			name: "no-issuer-key",
			config: &IntermediateCAConfig{
				Name: "Test CA",
				KeySpec: cryptoutilCACrypto.KeySpec{
					Type:       cryptoutilCACrypto.KeyTypeECDSA,
					ECDSACurve: "P-256",
				},
				ValidityDuration:  1 * 365 * 24 * time.Hour,
				IssuerCertificate: rootCA.Certificate,
			},
			wantErr: "issuer private key is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			intermediateCA, audit, err := provisioner.Provision(tc.config)
			require.Error(t, err)
			require.Nil(t, intermediateCA)
			require.Nil(t, audit)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestProvisioner_Provision_PathLenEnforcement(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	// Create root with path length 0.
	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootConfig := &cryptoutilCABootstrap.RootCAConfig{
		Name: "Zero Path Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
	}

	rootCA, _, err := bootstrapper.Bootstrap(rootConfig)
	require.NoError(t, err)

	provisioner := NewProvisioner(provider)

	config := &IntermediateCAConfig{
		Name: "Should Fail Intermediate",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  5 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}

	intermediateCA, _, err := provisioner.Provision(config)
	require.Error(t, err)
	require.Nil(t, intermediateCA)
	require.Contains(t, err.Error(), "path length 0, cannot sign subordinate CAs")
}

func TestProvisioner_Provision_ValidityTruncation(t *testing.T) {
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
		ValidityDuration:  1 * 365 * 24 * time.Hour, // 1 year.
		PathLenConstraint: 1,
	}

	rootCA, _, err := bootstrapper.Bootstrap(rootConfig)
	require.NoError(t, err)

	provisioner := NewProvisioner(provider)

	config := &IntermediateCAConfig{
		Name: "Truncated Intermediate",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  10 * 365 * 24 * time.Hour, // Request 10 years.
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}

	intermediateCA, _, err := provisioner.Provision(config)
	require.NoError(t, err)
	require.NotNil(t, intermediateCA)

	// Intermediate should not outlive root.
	require.True(t, !intermediateCA.Certificate.NotAfter.After(rootCA.Certificate.NotAfter))
}

func TestProvisioner_Provision_ChainVerification(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	rootCA := createTestRootCA(t, provider)

	provisioner := NewProvisioner(provider)

	config := &IntermediateCAConfig{
		Name: "Chain Test Intermediate",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  5 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}

	intermediateCA, _, err := provisioner.Provision(config)
	require.NoError(t, err)

	// Verify chain is valid.
	rootPool := x509.NewCertPool()
	rootPool.AddCert(rootCA.Certificate)

	opts := x509.VerifyOptions{
		Roots:       rootPool,
		CurrentTime: time.Now(),
	}

	chains, err := intermediateCA.Certificate.Verify(opts)
	require.NoError(t, err)
	require.NotEmpty(t, chains)
}
