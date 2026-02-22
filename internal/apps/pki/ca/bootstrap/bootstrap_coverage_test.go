// Copyright (c) 2025 Justin Cranford

package bootstrap

import (
	ecdsa "crypto/ecdsa"
	"crypto/ed25519"
	crand "crypto/rand"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	cryptoutilCACrypto "cryptoutil/internal/apps/pki/ca/crypto"

	"github.com/stretchr/testify/require"
)

// TestBootstrapper_Bootstrap_InvalidKeySpec triggers the GenerateKeyPair error branch.
func TestBootstrapper_Bootstrap_InvalidKeySpec(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	bootstrapper := NewBootstrapper(provider)

	_, _, err := bootstrapper.Bootstrap(&RootCAConfig{
		Name:              "Invalid KeySpec Root",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: "UNSUPPORTED_KEY_TYPE"},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate key pair")
}

// TestBootstrapper_Bootstrap_PersistMaterialsError triggers the persistMaterials
// error in Bootstrap by setting OutputDir to a file-blocked path.
func TestBootstrapper_Bootstrap_PersistMaterialsError(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	bootstrapper := NewBootstrapper(provider)

	tmpDir := t.TempDir()
	conflictFile := filepath.Join(tmpDir, "conflict")
	require.NoError(t, os.WriteFile(conflictFile, []byte("x"), 0o600))

	_, _, err := bootstrapper.Bootstrap(&RootCAConfig{
		Name:              "Persist Fail Root",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
		OutputDir:         filepath.Join(conflictFile, "subdir"),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to persist materials")
}

// TestKeyAlgorithmName_Ed25519_Bootstrap tests the Ed25519 branch of keyAlgorithmName.
func TestKeyAlgorithmName_Ed25519_Bootstrap(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	bootstrapper := NewBootstrapper(provider)

	rootCA, audit, err := bootstrapper.Bootstrap(&RootCAConfig{
		Name:              "Ed25519 Root CA",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeEdDSA, EdDSACurve: "Ed25519"},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, rootCA)
	require.NotNil(t, audit)
	require.Equal(t, "Ed25519", audit.KeyAlgorithm)

	_, isEd25519 := rootCA.PublicKey.(ed25519.PublicKey)
	require.True(t, isEd25519)
}

// TestKeyAlgorithmName_Unknown_Bootstrap tests the Unknown/default branch of keyAlgorithmName.
func TestKeyAlgorithmName_Unknown_Bootstrap(t *testing.T) {
	t.Parallel()

	result := keyAlgorithmName("not-a-key")
	require.Equal(t, "Unknown", result)
}

// TestPersistMaterials_Bootstrap_MkdirError tests MkdirAll failure in persistMaterials.
func TestPersistMaterials_Bootstrap_MkdirError(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	b := NewBootstrapper(provider)

	rootCA, _, err := b.Bootstrap(&RootCAConfig{
		Name:              "Root CA Mkdir Error",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
	})
	require.NoError(t, err)

	tmpDir := t.TempDir()
	conflictPath := filepath.Join(tmpDir, "conflict")
	require.NoError(t, os.WriteFile(conflictPath, []byte("x"), 0o600))

	config := &RootCAConfig{
		Name:      "Root CA Mkdir Error",
		OutputDir: filepath.Join(conflictPath, "subdir"),
	}

	err = b.persistMaterials(config, rootCA)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create output directory")
}

// TestPersistMaterials_Bootstrap_CertWriteError tests cert write failure in persistMaterials.
func TestPersistMaterials_Bootstrap_CertWriteError(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	b := NewBootstrapper(provider)

	rootCA, _, err := b.Bootstrap(&RootCAConfig{
		Name:              "Root CA Cert Error",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
	})
	require.NoError(t, err)

	tmpDir := t.TempDir()
	certBlockPath := filepath.Join(tmpDir, "Root CA Cert Error.crt")
	require.NoError(t, os.Mkdir(certBlockPath, 0o755))

	config := &RootCAConfig{
		Name:      "Root CA Cert Error",
		OutputDir: tmpDir,
	}

	err = b.persistMaterials(config, rootCA)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write certificate")
}

// TestPersistMaterials_Bootstrap_KeyWriteError tests key write failure in persistMaterials.
func TestPersistMaterials_Bootstrap_KeyWriteError(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()
	b := NewBootstrapper(provider)

	rootCA, _, err := b.Bootstrap(&RootCAConfig{
		Name:              "Key Write Error Root",
		KeySpec:           cryptoutilCACrypto.KeySpec{Type: cryptoutilCACrypto.KeyTypeECDSA, ECDSACurve: "P-256"},
		ValidityDuration:  10 * 365 * 24 * time.Hour,
		PathLenConstraint: 1,
	})
	require.NoError(t, err)

	tmpDir := t.TempDir()
	keyBlockPath := filepath.Join(tmpDir, "Key Write Error Root.key")
	require.NoError(t, os.Mkdir(keyBlockPath, 0o755))

	config := &RootCAConfig{
		Name:      "Key Write Error Root",
		OutputDir: tmpDir,
	}

	err = b.persistMaterials(config, rootCA)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write private key")
}

// TestBootstrapper_Bootstrap_SignatureAlgorithmError triggers GetSignatureAlgorithm error.
// This is done via the provider mock: we use a crafted key type that isn't supported.
// We bypass the normal flow by directly patching the fakeKey into the bootstrap path…
// Actually we need to test it via a custom call. Since Bootstrap uses keyPair.PublicKey
// from GenerateKeyPair – which always returns a real key – we test the similar error
// path by calling b.provider.GetSignatureAlgorithm with a bad key type directly.
func TestBootstrapper_GetSignatureAlgorithm_Unsupported(t *testing.T) {
	t.Parallel()

	provider := cryptoutilCACrypto.NewSoftwareProvider()

	_, err := provider.GetSignatureAlgorithm("not-a-real-key")
	require.Error(t, err)
}

// var prevents unused import error for pkix and big.
var (
	_ = pkix.Name{}
	_ = big.NewInt
	_ *ecdsa.PublicKey
	_ = crand.Reader
)
