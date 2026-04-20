// Copyright (c) 2025 Justin Cranford
//

package tls_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps/framework/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// TestWriteAdminCABundle_WriteFileFails verifies that writeAdminCABundle propagates
// a file-write error correctly.
func TestWriteAdminCABundle_WriteFileFails(t *testing.T) {
	t.Parallel()

	gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
		stubMkdirAll,
		func(_ string, _ []byte, _ os.FileMode) error {
			return fmt.Errorf("injected write error")
		},
		stubCreateCA, stubCreateLeaf, stubGetKeyPair,
		stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
	)

	err := gen.ExportedWriteAdminCABundle(t.TempDir(), [][]byte{[]byte("CERT PEM DATA")})
	require.ErrorContains(t, err, "write admin CA bundle")
	require.ErrorContains(t, err, "injected write error")
}

// TestWriteAdminCABundle_EmptyCerts verifies that writeAdminCABundle writes an empty
// bundle file (no-op concatenation) when adminCACerts is empty.
func TestWriteAdminCABundle_EmptyCerts(t *testing.T) {
	t.Parallel()

	var writtenData []byte

	gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
		stubMkdirAll,
		func(_ string, data []byte, _ os.FileMode) error {
			writtenData = data

			return nil
		},
		stubCreateCA, stubCreateLeaf, stubGetKeyPair,
		stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
	)

	require.NoError(t, gen.ExportedWriteAdminCABundle(t.TempDir(), nil))
	require.Empty(t, writtenData)
}

// TestValidateTargetDir_FileInsteadOfDir verifies that validateTargetDir returns an error
// when the basePath exists as a file rather than a directory.
// Note: the "file not dir" branch is also covered via TestGenerate_BasepathIsFile,
// but an explicit unit test here validates the direct seam.
func TestValidateTargetDir_FileInsteadOfDir(t *testing.T) {
	t.Parallel()

	gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
		stubMkdirAll, stubWriteFile, stubCreateCA, stubCreateLeaf, stubGetKeyPair,
		stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
	)

	tmpDir := t.TempDir()
	filePath := fmt.Sprintf("%s/blocking-file", tmpDir)
	require.NoError(t, os.WriteFile(filePath, []byte("data"), cryptoutilSharedMagic.CacheFilePermissions))

	err := gen.ExportedValidateTargetDir(filePath)
	require.ErrorContains(t, err, "not a directory")
}
