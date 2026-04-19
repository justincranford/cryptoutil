// Copyright (c) 2025 Justin Cranford
//

package tls_test

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps/framework/tls"
)

// TestWriteTLSConfigYAML_HappyPath verifies that writeTLSConfigYAML writes a valid
// tls-config.yml with correct content and file mode.
func TestWriteTLSConfigYAML_HappyPath(t *testing.T) {
	t.Parallel()

	var writtenPath string

	var writtenContent []byte

	var writtenMode os.FileMode

	captureWriteFile := func(path string, data []byte, mode os.FileMode) error {
		writtenPath = path
		writtenContent = data
		writtenMode = mode

		return nil
	}

	gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
		stubMkdirAll, captureWriteFile, stubCreateCA, stubCreateLeaf,
		stubGetKeyPair, stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
	)

	issuingCA := makeStubSubject(t)
	targetDir := t.TempDir()

	require.NoError(t, gen.ExportedWriteTLSConfigYAML(targetDir, issuingCA))

	// Verify the path is targetDir/tls-config.yml.
	require.Equal(t, filepath.Join(targetDir, "tls-config.yml"), writtenPath)

	// Verify file mode is 0o440 (private key embedded).
	require.Equal(t, os.FileMode(0o440), writtenMode)

	// Verify content contains the expected YAML keys.
	content := string(writtenContent)
	require.Contains(t, content, "tls-public-mode: mixed")
	require.Contains(t, content, "tls-mixed-ca-cert-pem: ")
	require.Contains(t, content, "tls-mixed-ca-key-pem: ")

	// Verify the cert and key base64 values decode to non-empty PEM data.
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "tls-mixed-ca-cert-pem: ") {
			b64 := strings.TrimPrefix(line, "tls-mixed-ca-cert-pem: ")
			decoded, err := base64.StdEncoding.DecodeString(b64)
			require.NoError(t, err, "cert PEM base64 should decode cleanly")
			require.Contains(t, string(decoded), "BEGIN CERTIFICATE")
		}

		if strings.HasPrefix(line, "tls-mixed-ca-key-pem: ") {
			b64 := strings.TrimPrefix(line, "tls-mixed-ca-key-pem: ")
			decoded, err := base64.StdEncoding.DecodeString(b64)
			require.NoError(t, err, "key PEM base64 should decode cleanly")
			require.Contains(t, string(decoded), "BEGIN PRIVATE KEY")
		}
	}
}

// TestWriteTLSConfigYAML_WriteFileError verifies that writeTLSConfigYAML propagates
// file-write errors correctly.
func TestWriteTLSConfigYAML_WriteFileError(t *testing.T) {
	t.Parallel()

	gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
		stubMkdirAll,
		func(_ string, _ []byte, _ os.FileMode) error {
			return fmt.Errorf("disk full")
		},
		stubCreateCA, stubCreateLeaf, stubGetKeyPair, stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
	)

	err := gen.ExportedWriteTLSConfigYAML(t.TempDir(), makeStubSubject(t))
	require.ErrorContains(t, err, "write tls-config.yml")
	require.ErrorContains(t, err, "disk full")
}

// TestWriteTLSConfigYAML_Generate_WritesConfigFile verifies that Generate writes
// tls-config.yml at targetDir/tls-config.yml (not inside the tier subdirectory).
func TestWriteTLSConfigYAML_Generate_WritesConfigFile(t *testing.T) {
	t.Parallel()

	var tlsConfigPath string

	captureWriteFile := func(path string, _ []byte, _ os.FileMode) error {
		if strings.HasSuffix(path, "tls-config.yml") {
			tlsConfigPath = path
		}

		return nil
	}

	gen := cryptoutilAppsFrameworkTls.ExportedNewTestGenerator(
		stubMkdirAll, captureWriteFile, stubCreateCA, stubCreateLeaf,
		stubGetKeyPair, stubEncodePKCS12, stubEncodeTrustPKCS12, stubGetRealmsForPSID,
	)

	tmpDir := t.TempDir()
	require.NoError(t, gen.Generate("skeleton-template", tmpDir))

	// tls-config.yml must be at targetDir/tls-config.yml, not inside the tier subdir.
	require.Equal(t, filepath.Join(tmpDir, "tls-config.yml"), tlsConfigPath,
		"tls-config.yml must be written at targetDir root, not inside tier subdir")
}
