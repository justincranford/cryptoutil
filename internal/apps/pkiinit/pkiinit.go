// Copyright (c) 2025 Justin Cranford
//
//

// Package pkiinit provides the PKI init CLI command for generating TLS certificates.
// This command is used by Docker Compose E2E deployments to generate TLS certificate
// hierarchies into a shared volume, enabling proper TLS verification in tests.
package pkiinit

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	cryptoutilAppsFrameworkServiceConfigTlsGenerator "cryptoutil/internal/apps/framework/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// Run executes the pki-init CLI command.
// It generates a TLS certificate hierarchy and writes the certs to an output directory.
// The output directory is typically a Docker volume shared between pki-init and app services.
func Run(args []string, _ io.Reader, stdout io.Writer, stderr io.Writer) int {
	outputDir := cryptoutilSharedMagic.PKIInitDefaultOutputDir

	domains := append([]string{}, cryptoutilSharedMagic.DefaultTLSPublicDNSNames...)
	ips := append([]string{}, cryptoutilSharedMagic.DefaultTLSPublicIPAddresses...)

	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--output-dir="):
			outputDir = strings.TrimPrefix(arg, "--output-dir=")
		case strings.HasPrefix(arg, "--domain="):
			domains = append(domains, strings.TrimPrefix(arg, "--domain="))
		case strings.HasPrefix(arg, "--ip="):
			ips = append(ips, strings.TrimPrefix(arg, "--ip="))
		case arg == "--help" || arg == "-h":
			_, _ = fmt.Fprintf(stdout, "Usage: cryptoutil pki-init [--output-dir=DIR] [--domain=DNS] [--ip=IP]\n")

			return 0
		}
	}

	if err := os.MkdirAll(outputDir, cryptoutilSharedMagic.PKIInitCertsDirMode); err != nil {
		_, _ = fmt.Fprintf(stderr, "pki-init: failed to create output directory %q: %v\n", outputDir, err)

		return 1
	}

	settings, err := cryptoutilAppsFrameworkServiceConfigTlsGenerator.GenerateAutoTLSGeneratedSettings(
		domains, ips, cryptoutilSharedMagic.PKIInitCertValidityDays,
	)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "pki-init: failed to generate TLS material: %v\n", err)

		return 1
	}

	rootCACertPEM := extractRootCACert(settings.StaticCertPEM)

	rootCAPath := filepath.Join(outputDir, cryptoutilSharedMagic.PKIInitRootCACertFile)
	if err := os.WriteFile(rootCAPath, rootCACertPEM, cryptoutilSharedMagic.PKIInitCertFileMode); err != nil {
		_, _ = fmt.Fprintf(stderr, "pki-init: failed to write root CA cert to %q: %v\n", rootCAPath, err)

		return 1
	}

	certB64 := base64.StdEncoding.EncodeToString(settings.StaticCertPEM)
	keyB64 := base64.StdEncoding.EncodeToString(settings.StaticKeyPEM)
	configYAML := fmt.Sprintf(cryptoutilSharedMagic.PKIInitTLSConfigYAMLFormat, certB64, keyB64)

	configPath := filepath.Join(outputDir, cryptoutilSharedMagic.PKIInitTLSConfigFile)
	if err := os.WriteFile(configPath, []byte(configYAML), cryptoutilSharedMagic.PKIInitCertFileMode); err != nil {
		_, _ = fmt.Fprintf(stderr, "pki-init: failed to write TLS config to %q: %v\n", configPath, err)

		return 1
	}

	_, _ = fmt.Fprintf(stdout, "pki-init: certificates written to %q (root-ca.pem, tls-config.yml)\n", outputDir)

	return 0
}

// extractRootCACert extracts the last (root) certificate from a PEM chain.
func extractRootCACert(chainPEM []byte) []byte {
	var lastBlock *pem.Block

	rest := chainPEM

	for {
		var block *pem.Block

		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}

		lastBlock = block
	}

	if lastBlock == nil {
		return chainPEM
	}

	return pem.EncodeToMemory(lastBlock)
}
