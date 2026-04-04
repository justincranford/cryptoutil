// Copyright (c) 2025 Justin Cranford
//
//

package magic

import "os"

// PKI Init constants for the pki-init Docker Compose init job.
const (
	// PSIDPKIInit is the product-service ID for the pki-init job ("pki-init").
	// Used as the pki-init subcommand name in suite and product CLI routers.
	PSIDPKIInit = "pki-init"

	// PKIInitDefaultOutputDir is the default output directory for PKI init certificates.
	PKIInitDefaultOutputDir = "/certs"

	// PKIInitRootCACertFile is the filename for the root CA certificate (raw PEM).
	// Used for BusyBox wget --ca-certificate and Go test HTTP client trust configuration.
	PKIInitRootCACertFile = "root-ca.pem"

	// PKIInitTLSConfigFile is the filename for the generated TLS config YAML.
	// Contains base64-encoded tls-static-cert-pem and tls-static-key-pem.
	PKIInitTLSConfigFile = "tls-config.yml"

	// PKIInitTLSConfigYAMLFormat is the YAML format string for the generated TLS config.
	// Parameters: 1=base64(server cert chain PEM), 2=base64(server key PEM).
	PKIInitTLSConfigYAMLFormat = "tls-public-mode: static\ntls-private-mode: static\ntls-static-cert-pem: %s\ntls-static-key-pem: %s\n"

	// PKIInitCertValidityDays is the validity period for PKI init certificates in days.
	PKIInitCertValidityDays = 365

	// PKIInitCertFileMode is the file permission mode for certificate files.
	PKIInitCertFileMode = os.FileMode(0o644)

	// PKIInitCertsDirMode is the directory permission mode for the certs directory.
	PKIInitCertsDirMode = os.FileMode(0o755)

	// PKIInitE2ECertsSubdir is the subdirectory under a deployment dir for E2E PKI certs.
	PKIInitE2ECertsSubdir = "certs"
)

// TLS signing algorithm constants for --signing-algorithm flag in pki-init functions.
// All values are FIPS 140-3 approved.
const (
	// TLSSigningAlgorithmDefault is the default TLS signing algorithm (ECDSA P-384 with SHA-384).
	TLSSigningAlgorithmDefault = "ECDSA-P384-SHA384"

	// TLSSigningAlgorithmECDSAP256SHA256 is ECDSA P-256 with SHA-256.
	TLSSigningAlgorithmECDSAP256SHA256 = "ECDSA-P256-SHA256"

	// TLSSigningAlgorithmECDSAP384SHA384 is ECDSA P-384 with SHA-384.
	TLSSigningAlgorithmECDSAP384SHA384 = "ECDSA-P384-SHA384"

	// TLSSigningAlgorithmECDSAP521SHA512 is ECDSA P-521 with SHA-512.
	TLSSigningAlgorithmECDSAP521SHA512 = "ECDSA-P521-SHA512"

	// TLSSigningAlgorithmRSA2048SHA256 is RSA-2048 with SHA-256.
	TLSSigningAlgorithmRSA2048SHA256 = "RSA-2048-SHA256"

	// TLSSigningAlgorithmRSA3072SHA256 is RSA-3072 with SHA-256.
	TLSSigningAlgorithmRSA3072SHA256 = "RSA-3072-SHA256"

	// TLSSigningAlgorithmRSA4096SHA256 is RSA-4096 with SHA-256.
	TLSSigningAlgorithmRSA4096SHA256 = "RSA-4096-SHA256"
)
