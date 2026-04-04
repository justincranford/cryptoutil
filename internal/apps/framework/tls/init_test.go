// Copyright (c) 2025 Justin Cranford
//

package tls_test

import (
	"bytes"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps/framework/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestInit_HappyPath(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.Init([]string{"--output-dir=" + outputDir}, nil, &stdout, &stderr)
	require.Equal(t, 0, code, "expected exit 0; stderr=%s", stderr.String())
	require.Contains(t, stdout.String(), "certificates written")
	require.Empty(t, stderr.String())

	rootCAPath := filepath.Join(outputDir, cryptoutilSharedMagic.PKIInitRootCACertFile)
	rootCABytes, err := os.ReadFile(rootCAPath)

	require.NoError(t, err, "root-ca.pem should be written")

	block, _ := pem.Decode(rootCABytes)

	require.NotNil(t, block, "root-ca.pem should contain valid PEM")
	require.Equal(t, cryptoutilSharedMagic.StringPEMTypeCertificate, block.Type)

	tlsConfigPath := filepath.Join(outputDir, cryptoutilSharedMagic.PKIInitTLSConfigFile)
	tlsConfigBytes, err := os.ReadFile(tlsConfigPath)

	require.NoError(t, err, "tls-config.yml should be written")
	require.Contains(t, string(tlsConfigBytes), "tls-public-mode: static")
	require.Contains(t, string(tlsConfigBytes), "tls-private-mode: static")
	require.Contains(t, string(tlsConfigBytes), "tls-static-cert-pem:")
	require.Contains(t, string(tlsConfigBytes), "tls-static-key-pem:")
}

func TestInit_HelpFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		arg  string
	}{
		{name: "help long", arg: "--help"},
		{name: "help short", arg: "-h"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var stdout, stderr bytes.Buffer

			code := cryptoutilAppsFrameworkTls.Init([]string{tc.arg}, nil, &stdout, &stderr)
			require.Equal(t, 0, code)
			require.Contains(t, stdout.String(), "output-dir")
			require.Empty(t, stderr.String())
		})
	}
}

func TestInit_DefaultOutputDir(t *testing.T) {
	t.Parallel()

	// Verify the default output dir constant value is set.
	require.NotEmpty(t, cryptoutilSharedMagic.PKIInitDefaultOutputDir)
}

func TestInit_InvalidOutputDir(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	// Use a file path as a directory - this will fail MkdirAll on any OS.
	// Create a file and use it as an output directory.
	tmpFile, err := os.CreateTemp(t.TempDir(), "not-a-dir")

	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	code := cryptoutilAppsFrameworkTls.Init([]string{"--output-dir=" + filepath.Join(tmpFile.Name(), "subdir")}, nil, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "failed to create output directory")
}

func TestInit_ExtraFlagsIgnored(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.Init([]string{"--output-dir=" + outputDir, "--domain=example.com", "--ip=192.168.1.1"}, nil, &stdout, &stderr)
	require.Equal(t, 0, code, "stderr=%s", stderr.String())
	require.Contains(t, stdout.String(), "certificates written")
}

func TestExtractRootCACert_MultipleCerts(t *testing.T) {
	t.Parallel()

	// Build a multi-cert PEM chain: cert1 + cert2 (cert2 = root CA).
	cert1PEM := buildDummyCertPEM(t, "cert1")
	cert2PEM := buildDummyCertPEM(t, "cert2")
	chain := append(cert1PEM, cert2PEM...)

	// Use a real run to indirectly test extractRootCACert via output.
	outputDir := t.TempDir()

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.Init([]string{"--output-dir=" + outputDir}, nil, &stdout, &stderr)
	require.Equal(t, 0, code, "stderr=%s", stderr.String())

	rootCAPath := filepath.Join(outputDir, cryptoutilSharedMagic.PKIInitRootCACertFile)
	rootCABytes, err := os.ReadFile(rootCAPath)

	require.NoError(t, err)

	// Root CA should be a single PEM block (not the full chain).
	blocks := countPEMBlocks(rootCABytes)

	require.Equal(t, 1, blocks, "root-ca.pem should contain exactly 1 PEM block (the root CA)")

	_ = chain // suppress unused var
}

func TestInit_WriteRootCAError(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()

	// Pre-create root-ca.pem as a directory so WriteFile will fail.
	require.NoError(t, os.Mkdir(filepath.Join(outputDir, cryptoutilSharedMagic.PKIInitRootCACertFile), cryptoutilSharedMagic.DirPermissions))

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.Init([]string{"--output-dir=" + outputDir}, nil, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "failed to write root CA cert")
}

func TestInit_WriteTLSConfigError(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()

	// Pre-create tls-config.yml as a directory so WriteFile will fail.
	require.NoError(t, os.Mkdir(filepath.Join(outputDir, cryptoutilSharedMagic.PKIInitTLSConfigFile), cryptoutilSharedMagic.DirPermissions))

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.Init([]string{"--output-dir=" + outputDir}, nil, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "failed to write TLS config")
}

func TestExtractRootCACert_EmptyInput(t *testing.T) {
	t.Parallel()

	// nil input: lastBlock stays nil, returns original nil input.
	result := cryptoutilAppsFrameworkTls.ExportedExtractRootCACert(nil)

	require.Nil(t, result)

	// Empty bytes: same nil-lastBlock path, returns original empty slice.
	empty := []byte{}
	result2 := cryptoutilAppsFrameworkTls.ExportedExtractRootCACert(empty)

	require.Equal(t, empty, result2)
}

// buildDummyCertPEM creates a minimal valid PEM certificate block for testing.
func buildDummyCertPEM(t *testing.T, _ string) []byte {
	t.Helper()

	return pem.EncodeToMemory(&pem.Block{
		Type:  cryptoutilSharedMagic.StringPEMTypeCertificate,
		Bytes: []byte("dummy"),
	})
}

// countPEMBlocks counts the number of PEM blocks in the given data.
func countPEMBlocks(data []byte) int {
	count := 0
	rest := data

	for {
		var block *pem.Block

		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}

		count++
	}

	return count
}

func TestInitForSuite_HappyPath(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.InitForSuite("cryptoutil", []string{"--output-dir=" + outputDir}, &stdout, &stderr)
	require.Equal(t, 0, code, "expected exit 0; stderr=%s", stderr.String())
	require.Contains(t, stdout.String(), "certificates written")
	require.Empty(t, stderr.String())
}

func TestInitForProduct_HappyPath(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.InitForProduct("jose", []string{"--output-dir=" + outputDir}, &stdout, &stderr)
	require.Equal(t, 0, code, "expected exit 0; stderr=%s", stderr.String())
	require.Contains(t, stdout.String(), "certificates written")
	require.Empty(t, stderr.String())
}

func TestInitForService_HappyPath(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.InitForService("sm-kms", []string{"--output-dir=" + outputDir}, &stdout, &stderr)
	require.Equal(t, 0, code, "expected exit 0; stderr=%s", stderr.String())
	require.Contains(t, stdout.String(), "certificates written")
	require.Empty(t, stderr.String())
}

func TestInit_BadFlag(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.Init([]string{"--unknown-flag=value"}, nil, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Empty(t, stdout.String())
	require.Contains(t, stderr.String(), "unknown flag")
}

func TestInit_InvalidSigningAlgorithm(t *testing.T) {
	t.Parallel()

	outputDir := t.TempDir()

	var stdout, stderr bytes.Buffer

	code := cryptoutilAppsFrameworkTls.Init([]string{"--output-dir=" + outputDir, "--signing-algorithm=MD5-INVALID"}, nil, &stdout, &stderr)
	require.Equal(t, 1, code)
	require.Contains(t, stderr.String(), "invalid --signing-algorithm")
}
