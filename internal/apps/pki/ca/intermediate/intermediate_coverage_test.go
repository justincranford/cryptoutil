// Copyright (c) 2025 Justin Cranford

package intermediate

import (
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilCABootstrap "cryptoutil/internal/apps/pki/ca/bootstrap"
	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"

	"github.com/stretchr/testify/require"
)

// ─── validateConfig: missing branches ────────────────────────────────────────

func TestProvisioner_Provision_NonCAIssuer(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	provisioner := NewProvisioner(provider)

	// Create a non-CA leaf certificate to use as issuer.
	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	leafTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(42),
		Subject:               pkix.Name{CommonName: "leaf"},
		NotBefore:             time.Now().UTC().Add(-time.Hour),
		NotAfter:              time.Now().UTC().Add(time.Hour),
		IsCA:                  false,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(crand.Reader, leafTemplate, leafTemplate, &key.PublicKey, key)
	require.NoError(t, err)

	leafCert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	config := &IntermediateCAConfig{
		Name: "Test Intermediate",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: leafCert, // non-CA issuer.
		IssuerPrivateKey:  key,
	}

	_, _, err = provisioner.Provision(config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "issuer certificate is not a CA")
}

func TestProvisioner_Provision_IntermediatePathLenTooLong(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	// Create root CA with path length 1.
	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootConfig := &cryptoutilCABootstrap.RootCAConfig{
		Name: "Path-1 Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1, // MaxPathLen=1.
	}

	rootCA, _, err := bootstrapper.Bootstrap(rootConfig)
	require.NoError(t, err)

	provisioner := NewProvisioner(provider)

	// Try to create intermediate with PathLenConstraint=1 (must be < 1, i.e., only 0 is valid).
	config := &IntermediateCAConfig{
		Name: "Too Long Intermediate",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeECDSA,
			ECDSACurve: "P-256",
		},
		ValidityDuration:  5 * 365 * 24 * time.Hour,
		PathLenConstraint: 1, // == issuer MaxPathLen, so should fail.
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}

	_, _, err = provisioner.Provision(config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "intermediate path length")
}

// ─── keyAlgorithmName: Ed25519 and default branches ──────────────────────────

func TestKeyAlgorithmName_Ed25519(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	// Create Ed25519 root CA.
	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootConfig := &cryptoutilCABootstrap.RootCAConfig{
		Name: "Ed25519 Root CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeEdDSA,
			EdDSACurve: "Ed25519",
		},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
	}

	rootCA, _, err := bootstrapper.Bootstrap(rootConfig)
	require.NoError(t, err)

	provisioner := NewProvisioner(provider)

	// Create Ed25519 intermediate signed by Ed25519 root.
	config := &IntermediateCAConfig{
		Name: "Ed25519 Intermediate CA",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type:       cryptoutilCACrypto.KeyTypeEdDSA,
			EdDSACurve: "Ed25519",
		},
		ValidityDuration:  5 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}

	intermediateCA, audit, err := provisioner.Provision(config)
	require.NoError(t, err)
	require.NotNil(t, intermediateCA)
	require.NotNil(t, audit)
	require.Equal(t, "Ed25519", audit.KeyAlgorithm)

	// Verify it's actually Ed25519.
	_, ok := intermediateCA.PublicKey.(ed25519.PublicKey)
	require.True(t, ok, "expected Ed25519 public key")
}

func TestKeyAlgorithmName_Unknown(t *testing.T) {
	t.Parallel()

	// Call keyAlgorithmName directly with an unknown key type.
	result := keyAlgorithmName("unknown-key-type")
	require.Equal(t, "Unknown", result)
}

func TestProvisioner_Provision_InvalidKeySpec(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	// Create root CA for issuer.
	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootCA, _, err := bootstrapper.Bootstrap(&cryptoutilCABootstrap.RootCAConfig{
		Name:              "Root CA for Invalid KeySpec Test",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
	})
	require.NoError(t, err)

	provisioner := NewProvisioner(provider)

	// Use an unsupported key type that will fail in GenerateKeyPair.
	config := &IntermediateCAConfig{
		Name: "Invalid KeySpec Intermediate",
		KeySpec: cryptoutilCACrypto.KeySpec{
			Type: "UNSUPPORTED_KEY_TYPE",
		},
		ValidityDuration:  5 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	}

	_, _, err = provisioner.Provision(config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate key pair")
}

func TestPersistMaterials_MkdirError(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	p := NewProvisioner(provider)
	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootCA, _, err := bootstrapper.Bootstrap(&cryptoutilCABootstrap.RootCAConfig{
		Name:              "Root CA for Persist Test",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
	})
	require.NoError(t, err)

	intCA, _, err := p.Provision(&IntermediateCAConfig{
		Name:              "Persist Error Intermediate",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  5 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	})
	require.NoError(t, err)

	// Create a regular file at the output path to cause MkdirAll to fail.
	tmpDir := t.TempDir()
	conflictPath := filepath.Join(tmpDir, "conflictfile")

	err = os.WriteFile(conflictPath, []byte("block"), 0o600)
	require.NoError(t, err)

	config := &IntermediateCAConfig{
		Name:      "Persist Error Intermediate",
		OutputDir: filepath.Join(conflictPath, "subdir"), // Can't mkdir inside a file.
	}

	err = p.persistMaterials(config, intCA)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create output directory")
}

func TestPersistMaterials_CertWriteError(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	p := NewProvisioner(provider)

	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootCA, _, err := bootstrapper.Bootstrap(&cryptoutilCABootstrap.RootCAConfig{
		Name:              "Root CA for Cert Write Test",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
	})
	require.NoError(t, err)

	intCA, _, err := p.Provision(&IntermediateCAConfig{
		Name:              "Cert Write Error Intermediate",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  5 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	})
	require.NoError(t, err)

	// Create a directory where the cert file should be, so WriteFile fails.
	tmpDir := t.TempDir()
	certBlockPath := filepath.Join(tmpDir, "Cert Write Error Intermediate.crt")

	err = os.Mkdir(certBlockPath, 0o755) // Create dir in place of file.
	require.NoError(t, err)

	config := &IntermediateCAConfig{
		Name:      "Cert Write Error Intermediate",
		OutputDir: tmpDir,
	}

	err = p.persistMaterials(config, intCA)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write certificate")
}

func TestPersistMaterials_ChainWriteError(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	p := NewProvisioner(provider)

	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootCA, _, err := bootstrapper.Bootstrap(&cryptoutilCABootstrap.RootCAConfig{
		Name:              "Root CA for Chain Write Test",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
	})
	require.NoError(t, err)

	intCA, _, err := p.Provision(&IntermediateCAConfig{
		Name:              "Chain Write Error CA",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  5 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	})
	require.NoError(t, err)

	// Create a directory where the chain file should be, so WriteFile fails.
	tmpDir := t.TempDir()
	chainBlockPath := filepath.Join(tmpDir, "Chain Write Error CA-chain.crt")

	err = os.Mkdir(chainBlockPath, 0o755)
	require.NoError(t, err)

	config := &IntermediateCAConfig{
		Name:      "Chain Write Error CA",
		OutputDir: tmpDir,
	}

	err = p.persistMaterials(config, intCA)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write certificate chain")
}

func TestPersistMaterials_KeyWriteError(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	p := NewProvisioner(provider)

	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootCA, _, err := bootstrapper.Bootstrap(&cryptoutilCABootstrap.RootCAConfig{
		Name:              "Root CA for Key Write Test",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
	})
	require.NoError(t, err)

	intCA, _, err := p.Provision(&IntermediateCAConfig{
		Name:              "Key Write Error CA",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  5 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
	})
	require.NoError(t, err)

	// Create a directory where the key file should be, so WriteFile fails.
	tmpDir := t.TempDir()
	keyBlockPath := filepath.Join(tmpDir, "Key Write Error CA.key")

	err = os.Mkdir(keyBlockPath, 0o755)
	require.NoError(t, err)

	config := &IntermediateCAConfig{
		Name:      "Key Write Error CA",
		OutputDir: tmpDir,
	}

	err = p.persistMaterials(config, intCA)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write private key")
}

// TestProvisioner_Provision_SignatureAlgorithmError triggers the GetSignatureAlgorithm
// error branch by providing an issuer certificate with an unsupported public key type.
func TestProvisioner_Provision_SignatureAlgorithmError(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	provisioner := NewProvisioner(provider)

	// Generate a real ECDSA key pair so validateConfig passes.
	realKey, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	// Build a fake issuer certificate with an unsupported public key type
	// (a plain string instead of *ecdsa.PublicKey / *rsa.PublicKey / ed25519.PublicKey).
	fakeIssuer := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Fake Issuer"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		MaxPathLen:            1,
		PublicKey:             "unsupported-key-type", // Triggers unknown type in GetSignatureAlgorithm.
	}

	_, _, err = provisioner.Provision(&IntermediateCAConfig{
		Name:              "Test",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  5 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: fakeIssuer,
		IssuerPrivateKey:  realKey, // Passes validateConfig but signing won't be reached.
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to select signature algorithm")
}

// TestProvisioner_Provision_PersistMaterialsError triggers the persistMaterials
// error in Provision by setting OutputDir to an existing regular file.
func TestProvisioner_Provision_PersistMaterialsError(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	bootstrapper := cryptoutilCABootstrap.NewBootstrapper(provider)
	rootCA, _, err := bootstrapper.Bootstrap(&cryptoutilCABootstrap.RootCAConfig{
		Name:              "Root CA for Persist Fail Test",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
	})
	require.NoError(t, err)

	// Create a regular file where OutputDir should be.
	tmpDir := t.TempDir()
	conflictFile := filepath.Join(tmpDir, "conflict")
	require.NoError(t, os.WriteFile(conflictFile, []byte("x"), 0o600))

	provisioner := NewProvisioner(provider)
	_, _, err = provisioner.Provision(&IntermediateCAConfig{
		Name:              "Persist Fail Intermediate",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  5 * 365 * 24 * time.Hour,
		PathLenConstraint: 0,
		IssuerCertificate: rootCA.Certificate,
		IssuerPrivateKey:  rootCA.PrivateKey,
		OutputDir:         filepath.Join(conflictFile, "subdir"), // Conflict: file treated as dir.
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to persist materials")
}
