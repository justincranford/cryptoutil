// Copyright (c) 2025 Justin Cranford
//
//

// Package tls provides TLS certificate initialization for the framework.
// This package is used by Docker Compose E2E deployments to generate TLS
// certificate hierarchies into a shared volume, enabling proper TLS
// verification in tests and deployments.
package tls

import (
	"bytes"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"

	cryptoutilAppsFrameworkServiceConfigTlsGenerator "cryptoutil/internal/apps/framework/service/config/tls_generator"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// validSigningAlgorithms is the set of FIPS 140-3 approved signing algorithm flag values.
var validSigningAlgorithms = map[string]bool{
	cryptoutilSharedMagic.TLSSigningAlgorithmECDSAP256SHA256: true,
	cryptoutilSharedMagic.TLSSigningAlgorithmECDSAP384SHA384: true,
	cryptoutilSharedMagic.TLSSigningAlgorithmECDSAP521SHA512: true,
	cryptoutilSharedMagic.TLSSigningAlgorithmRSA2048SHA256:   true,
	cryptoutilSharedMagic.TLSSigningAlgorithmRSA3072SHA256:   true,
	cryptoutilSharedMagic.TLSSigningAlgorithmRSA4096SHA256:   true,
}

// Init executes the pki-init CLI command using pflag for argument parsing.
// It generates a TLS certificate hierarchy and writes the certs to an output directory.
// The output directory is typically a Docker volume shared between pki-init and app services.
func Init(args []string, _ io.Reader, stdout io.Writer, stderr io.Writer) int {
	return initWithID(cryptoutilSharedMagic.PSIDPKIInit, nil, args, stdout, stderr)
}

// InitForSuite executes the init subcommand for a named suite.
// Uses pflag for argument parsing. The suiteID is added as an extra DNS SAN so the
// generated certificate covers Docker Compose service discovery at the suite level.
func InitForSuite(suiteID string, args []string, stdout, stderr io.Writer) int {
	return initWithID(suiteID+"-init", []string{suiteID}, args, stdout, stderr)
}

// InitForProduct executes the init subcommand for a named product.
// Uses pflag for argument parsing. The productID is added as an extra DNS SAN so the
// generated certificate covers Docker Compose service discovery at the product level.
func InitForProduct(productID string, args []string, stdout, stderr io.Writer) int {
	return initWithID(productID+"-init", []string{productID}, args, stdout, stderr)
}

// InitForService executes the init subcommand for a named PS-ID service.
// Uses pflag for argument parsing. The serviceID is added as an extra DNS
// SAN so the generated certificate covers Docker Compose service discovery.
func InitForService(serviceID string, args []string, stdout, stderr io.Writer) int {
	return initWithID(serviceID+"-init", []string{serviceID}, args, stdout, stderr)
}

// initWithID is the unified pflag-based implementation for all tier Init functions.
// The name parameter is used as the pflag FlagSet name for error messages.
// The extraDNSDefaults are DNS SANs added before any --domain flags (e.g., serviceID).
func initWithID(name string, extraDNSDefaults []string, args []string, stdout, stderr io.Writer) int {
	var flagOutput bytes.Buffer

	fs := pflag.NewFlagSet(name, pflag.ContinueOnError)
	fs.SetOutput(&flagOutput)

	outputDir := fs.String("output-dir", cryptoutilSharedMagic.PKIInitDefaultOutputDir, "Output directory for generated certificates")
	signingAlgorithm := fs.String("signing-algorithm", cryptoutilSharedMagic.TLSSigningAlgorithmDefault, "TLS signing algorithm (FIPS-approved: ECDSA-P256-SHA256, ECDSA-P384-SHA384, ECDSA-P521-SHA512, RSA-2048-SHA256, RSA-3072-SHA256, RSA-4096-SHA256)")

	var extraDomains []string

	var extraIPs []string

	fs.StringArrayVar(&extraDomains, "domain", nil, "Additional DNS SAN (may be repeated)")
	fs.StringArrayVar(&extraIPs, "ip", nil, "Additional IP SAN (may be repeated)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, pflag.ErrHelp) {
			_, _ = fmt.Fprint(stdout, flagOutput.String())

			return 0
		}

		_, _ = fmt.Fprintf(stderr, "%s: %v\n", name, err)

		return 1
	}

	if !validSigningAlgorithms[*signingAlgorithm] {
		_, _ = fmt.Fprintf(stderr, "pki-init: invalid --signing-algorithm %q; valid values: ECDSA-P256-SHA256, ECDSA-P384-SHA384, ECDSA-P521-SHA512, RSA-2048-SHA256, RSA-3072-SHA256, RSA-4096-SHA256\n", *signingAlgorithm)

		return 1
	}

	domains := append([]string{}, cryptoutilSharedMagic.DefaultTLSPublicDNSNames...)
	domains = append(domains, extraDNSDefaults...)
	domains = append(domains, extraDomains...)

	ips := append([]string{}, cryptoutilSharedMagic.DefaultTLSPublicIPAddresses...)
	ips = append(ips, extraIPs...)

	return runInit(*outputDir, domains, ips, stdout, stderr)
}

// runInit contains the shared cert generation and file writing logic.
func runInit(outputDir string, domains, ips []string, stdout, stderr io.Writer) int {
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
