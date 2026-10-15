// Copyright (c) 2025-2026 Justin Cranford.
package tls_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestGenerate_PublishedFilePermissions(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
		os.MkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
		stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
	)

	require.NoError(t, gen.Generate(cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, outputDir))

	var (
		certificateCount   int
		privateKeyCount    int
		privatePKCS12Count int
		publicPKCS12Count  int
	)

	err := filepath.WalkDir(filepath.Join(outputDir, cryptoutilSharedMagic.OTLPServiceSkeletonTemplate), func(path string, entry os.DirEntry, walkErr error) error {
		require.NoError(t, walkErr)

		if entry.IsDir() {
			return nil
		}

		expectedMode := cryptoutilSharedMagic.PKIInitPublicCertFileMode

		switch filepath.Ext(path) {
		case cryptoutilSharedMagic.PKIInitPublicCertFileExtension:
			certificateCount++
		case cryptoutilSharedMagic.PKIInitPrivateKeyFileExtension:
			expectedMode = cryptoutilSharedMagic.PKIInitPrivateKeyFileMode
			privateKeyCount++
		case cryptoutilSharedMagic.PKIInitPKCS12FileExtension:
			if filepath.Base(filepath.Dir(path)) == "truststore" {
				publicPKCS12Count++
			} else {
				expectedMode = cryptoutilSharedMagic.PKIInitPrivateKeyFileMode
				privatePKCS12Count++
			}
		default:
			return nil
		}

		assertPublishedPermissions(t, path, expectedMode)

		return nil
	})
	require.NoError(t, err)
	require.Positive(t, certificateCount)
	require.Positive(t, privateKeyCount)
	require.Positive(t, privatePKCS12Count)
	require.Positive(t, publicPKCS12Count)

	assertPublishedPermissions(t, filepath.Join(outputDir, "tls-config.yml"), cryptoutilSharedMagic.PKIInitPrivateKeyFileMode)
	assertPublishedPermissions(t, filepath.Join(outputDir, cryptoutilSharedMagic.PKIInitAdminCABundleFilename), cryptoutilSharedMagic.PKIInitPublicCertFileMode)
}

func TestGenerate_RejectsOverlyPermissiveStagedPrivateKey(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()
	writeFile := func(path string, data []byte, mode os.FileMode) error {
		if filepath.Ext(path) == cryptoutilSharedMagic.PKIInitPrivateKeyFileExtension {
			mode |= 0o222
		}

		return stubWriteFile(path, data, mode)
	}

	gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
		os.MkdirAll, writeFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
		stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
	)

	err := gen.Generate(cryptoutilSharedMagic.OTLPServiceSkeletonTemplate, outputDir)
	require.ErrorContains(t, err, "permissions")
	require.NoDirExists(t, filepath.Join(outputDir, cryptoutilSharedMagic.OTLPServiceSkeletonTemplate))
	require.NoFileExists(t, filepath.Join(outputDir, "tls-config.yml"))
	require.NoFileExists(t, filepath.Join(outputDir, cryptoutilSharedMagic.PKIInitAdminCABundleFilename))
}

func assertPublishedPermissions(t *testing.T, path string, expected os.FileMode) {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err)

	permissions := info.Mode().Perm()
	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		require.Zero(t, permissions&0o222, "published file %q must be read-only on Windows", path)

		return
	}

	require.NotZero(t, permissions&0o400, "published file %q must be owner-readable", path)
	require.Zero(t, permissions&^expected.Perm(), "published file %q has excessive permissions", path)
}
