// Copyright (c) 2025 Justin Cranford
//
//

package magic

import "os"

// PKI Init constants for the pki-init Docker Compose init job.
const (
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
