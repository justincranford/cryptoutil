//go:build integration

// Copyright (c) 2025-2026 Justin Cranford.
//

package tls_test

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestGenerate_Integration_RealCrypto runs the full generator with real ECDSA P-384 key
// generation for the skeleton-template tier (smallest: 1 PS-ID × 4 variants). Verifies
// that all generated .crt files parse as valid X.509 certificates and all .key files parse
// as valid PKCS#8 private keys. Skipped under -short because real key generation is slow.
func TestGenerate_Integration_RealCrypto(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("real crypto: skipping under -short")
	}

	ctx := context.Background()

	ts, err := cryptoutilAppsFrameworkTls.ExportedProductionNewTelemetryService(ctx)
	require.NoError(t, err)
	require.NotNil(t, ts)

	t.Cleanup(ts.Shutdown)

	gen, err := cryptoutilAppsFrameworkTls.ExportedProductionNewGenerator(ctx, ts)
	require.NoError(t, err)
	require.NotNil(t, gen)

	t.Cleanup(gen.Shutdown)

	psID := cryptoutilSharedMagic.OTLPServiceSkeletonTemplate
	tmpDir := t.TempDir()

	require.NoError(t, gen.Generate(psID, tmpDir), "Generate(%q) must succeed with real crypto", psID)

	basePath := filepath.Join(tmpDir, psID)

	var crtCount, keyCount int

	require.NoError(t, filepath.WalkDir(basePath, func(p string, d fs.DirEntry, _ error) error {
		if d.IsDir() {
			return nil
		}

		data := readTestFileWithinRoot(t, basePath, p)
		require.NotEmpty(t, data, "generated file must not be empty: %s", p)

		switch filepath.Ext(p) {
		case cryptoutilSharedMagic.PKIInitPublicCertFileExtension:
			assertValidCertChain(t, p, data)

			crtCount++
		case cryptoutilSharedMagic.PKIInitPrivateKeyFileExtension:
			assertValidPKCS8Key(t, p, data)

			keyCount++
		}

		return nil
	}))

	require.Positive(t, crtCount, "expected at least one .crt file to be generated")
	require.Positive(t, keyCount, "expected at least one .key file to be generated")
}

func readTestFileWithinRoot(t *testing.T, rootPath, filePath string) []byte {
	t.Helper()

	root, err := os.OpenRoot(rootPath)

	require.NoError(t, err)
	defer func() {
		require.NoError(t, root.Close())
	}()

	relativePath, err := filepath.Rel(rootPath, filePath)
	require.NoError(t, err)

	file, err := root.Open(relativePath)
	require.NoError(t, err)

	data, err := io.ReadAll(file)
	require.NoError(t, err)
	require.NoError(t, file.Close())

	return data
}

// assertValidCertChain verifies that pemData contains at least one valid PEM-encoded
// X.509 certificate block, parseable with x509.ParseCertificate.
func assertValidCertChain(t *testing.T, path string, pemData []byte) {
	t.Helper()

	rest := pemData
	certCount := 0

	for {
		block, remaining := pem.Decode(rest)
		if block == nil {
			break
		}

		rest = remaining

		if block.Type != cryptoutilSharedMagic.StringPEMTypeCertificate {
			continue
		}

		_, err := x509.ParseCertificate(block.Bytes)
		require.NoError(t, err, "x509.ParseCertificate failed for cert in %s", path)

		certCount++
	}

	require.Positive(t, certCount, "no valid CERTIFICATE blocks found in %s", path)
}

// assertValidPKCS8Key verifies that pemData contains a valid PKCS#8-encoded private key
// (the format used by PEMEncode for all key types: RSA, ECDSA, Ed25519).
func assertValidPKCS8Key(t *testing.T, path string, pemData []byte) {
	t.Helper()

	block, _ := pem.Decode(pemData)
	require.NotNil(t, block, "no PEM block found in %s", path)
	require.Equal(t, cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey, block.Type, "expected PKCS#8 PEM type in %s", path)

	_, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	require.NoError(t, err, "x509.ParsePKCS8PrivateKey failed for %s", path)
}
