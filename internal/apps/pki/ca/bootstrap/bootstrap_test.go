// Copyright (c) 2025 Justin Cranford

package bootstrap

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"
	cryptoutilCAProfileSubject "cryptoutil/internal/apps/pki/ca/profile/subject"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestBootstrapper_Bootstrap_ECDSA(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	bootstrapper := NewBootstrapper(provider)

	config := &RootCAConfig{
		Name: "Test Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  10 * 365 * 24 * time.Hour, // 10 years.
		PathLenConstraint: 2,
	}

	rootCA, audit, err := bootstrapper.Bootstrap(config)
	require.NoError(t, err)
	require.NotNil(t, rootCA)
	require.NotNil(t, audit)

	// Verify root CA.
	require.Equal(t, "Test Root CA", rootCA.Name)
	require.NotNil(t, rootCA.Certificate)
	require.NotNil(t, rootCA.PrivateKey)
	require.NotNil(t, rootCA.PublicKey)
	require.NotEmpty(t, rootCA.CertificatePEM)

	// Verify certificate properties.
	cert := rootCA.Certificate
	require.True(t, cert.IsCA)
	require.Equal(t, 2, cert.MaxPathLen)
	require.Contains(t, cert.Subject.CommonName, "Test Root CA")

	// Verify audit entry.
	require.Equal(t, "root_ca_bootstrap", audit.Operation)
	require.Equal(t, "Test Root CA", audit.CAName)
	require.NotEmpty(t, audit.SerialNumber)
	require.NotEmpty(t, audit.Fingerprint)
	require.Equal(t, "ECDSA", audit.KeyAlgorithm)
}

func TestBootstrapper_Bootstrap_RSA(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	bootstrapper := NewBootstrapper(provider)

	config := &RootCAConfig{
		Name: "RSA Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:    cryptoutilCACrypto.KeyTypeRSA,
			RSABits: cryptoutilCACrypto.MinRSAKeyBits,
		},
		ValidityDuration:  5 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
	}

	rootCA, audit, err := bootstrapper.Bootstrap(config)
	require.NoError(t, err)
	require.NotNil(t, rootCA)
	require.NotNil(t, audit)

	require.Equal(t, "RSA Root CA", rootCA.Name)
	require.True(t, rootCA.Certificate.IsCA)
	require.Equal(t, 1, rootCA.Certificate.MaxPathLen)
	require.Equal(t, "RSA", audit.KeyAlgorithm)
}

func TestBootstrapper_Bootstrap_EdDSA(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	bootstrapper := NewBootstrapper(provider)

	config := &RootCAConfig{
		Name: "EdDSA Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeEdDSA,
			EdDSACurve: "Ed25519",
		},
		ValidityDuration:  3 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
	}

	rootCA, audit, err := bootstrapper.Bootstrap(config)
	require.NoError(t, err)
	require.NotNil(t, rootCA)
	require.NotNil(t, audit)

	require.Equal(t, "EdDSA Root CA", rootCA.Name)
	require.True(t, rootCA.Certificate.IsCA)
	require.True(t, rootCA.Certificate.MaxPathLenZero)
	require.Equal(t, "Ed25519", audit.KeyAlgorithm)
}

func TestBootstrapper_Bootstrap_WithSubjectProfile(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	bootstrapper := NewBootstrapper(provider)

	subjectProfile := &cryptoutilCAProfileSubject.Profile{
		Name: "enterprise-root",
		Subject: cryptoutilCAProfileSubject.DN{
			Organization:       []string{"Example Corp"},
			OrganizationalUnit: []string{"IT Security"},
			Country:            []string{"US"},
			State:              []string{"California"},
			Locality:           []string{"San Francisco"},
		},
	}

	config := &RootCAConfig{
		Name:           "Enterprise Root CA",
		SubjectProfile: subjectProfile,
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-384",
		},
		ValidityDuration:  20 * 365 * 24 * time.Hour,
		PathLenConstraint: 3,
	}

	rootCA, audit, err := bootstrapper.Bootstrap(config)
	require.NoError(t, err)
	require.NotNil(t, rootCA)

	// Verify subject DN.
	cert := rootCA.Certificate
	require.Equal(t, "Enterprise Root CA", cert.Subject.CommonName)
	require.Equal(t, []string{"Example Corp"}, cert.Subject.Organization)
	require.Equal(t, []string{"IT Security"}, cert.Subject.OrganizationalUnit)
	require.Equal(t, []string{"US"}, cert.Subject.Country)
	require.Equal(t, []string{"California"}, cert.Subject.Province)
	require.Equal(t, []string{"San Francisco"}, cert.Subject.Locality)

	require.NotEmpty(t, audit.SubjectDN)
}

func TestBootstrapper_Bootstrap_WithPersistence(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	bootstrapper := NewBootstrapper(provider)

	config := &RootCAConfig{
		Name: "Persisted Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  1 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
		OutputDir:         tempDir,
	}

	rootCA, _, err := bootstrapper.Bootstrap(config)
	require.NoError(t, err)
	require.NotNil(t, rootCA)

	// Verify certificate file exists.
	certPath := filepath.Join(tempDir, "Persisted Root CA.crt")
	certData, err := os.ReadFile(certPath)
	require.NoError(t, err)
	require.NotEmpty(t, certData)
	require.Contains(t, string(certData), "BEGIN CERTIFICATE")

	// Verify key file exists.
	keyPath := filepath.Join(tempDir, "Persisted Root CA.key")
	keyData, err := os.ReadFile(keyPath)
	require.NoError(t, err)
	require.NotEmpty(t, keyData)
	require.Contains(t, string(keyData), "BEGIN "+cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey) // pragma: allowlist secret

	// Verify key file permissions (should be restrictive).
	keyInfo, err := os.Stat(keyPath)
	require.NoError(t, err)
	require.NotNil(t, keyInfo)
}

func TestBootstrapper_Bootstrap_InvalidConfig(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	bootstrapper := NewBootstrapper(provider)

	tests := []struct {
		name    string
		config  *RootCAConfig
		wantErr string
	}{
		{
			name:    "nil-config",
			config:  nil,
			wantErr: "config cannot be nil",
		},
		{
			name: "empty-name",
			config: &RootCAConfig{
				KeySpec: cryptoutilCACrypto.KeySpec{
					Type:       cryptoutilCACrypto.KeyTypeECDSA,
					ECDSACurve: "P-256",
				},
				ValidityDuration: 1 * 365 * 24 * time.Hour,
			},
			wantErr: "CA name is required",
		},
		{
			name: "zero-validity",
			config: &RootCAConfig{
				Name: "Test CA",
				KeySpec: cryptoutilCACrypto.KeySpec{
					Type:       cryptoutilCACrypto.KeyTypeECDSA,
					ECDSACurve: "P-256",
				},
				ValidityDuration: 0,
			},
			wantErr: "validity duration must be positive",
		},
		{
			name: "negative-path-len",
			config: &RootCAConfig{
				Name: "Test CA",
				KeySpec: cryptoutilCACrypto.KeySpec{
					Type:       cryptoutilCACrypto.KeyTypeECDSA,
					ECDSACurve: "P-256",
				},
				ValidityDuration:  1 * 365 * 24 * time.Hour,
				PathLenConstraint: -1,
			},
			wantErr: "path length constraint cannot be negative",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rootCA, audit, err := bootstrapper.Bootstrap(tc.config)
			require.Error(t, err)
			require.Nil(t, rootCA)
			require.Nil(t, audit)
			require.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestBootstrapper_Bootstrap_CertificateValidity(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	bootstrapper := NewBootstrapper(provider)

	validityDuration := 5 * 365 * 24 * time.Hour

	config := &RootCAConfig{
		Name: "Validity Test CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  validityDuration,
		PathLenConstraint: 0,
	}

	beforeBoot := time.Now().UTC()
	rootCA, _, err := bootstrapper.Bootstrap(config)
	require.NoError(t, err)

	afterBoot := time.Now().UTC()

	cert := rootCA.Certificate

	// NotBefore should be slightly before boot time (backdated).
	require.True(t, cert.NotBefore.Before(beforeBoot.Add(time.Second)))
	require.True(t, cert.NotBefore.After(beforeBoot.Add(-2*time.Minute)))

	// NotAfter should be approximately validityDuration after boot.
	expectedNotAfter := beforeBoot.Add(validityDuration)
	require.True(t, cert.NotAfter.After(expectedNotAfter.Add(-time.Minute)))
	require.True(t, cert.NotAfter.Before(afterBoot.Add(validityDuration).Add(time.Minute)))
}
