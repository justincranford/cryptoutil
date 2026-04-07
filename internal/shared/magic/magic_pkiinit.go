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

	// PKIInitCertValidityDays is the validity period for PKI init certificates in days.
	PKIInitCertValidityDays = 365

	// PKIInitCertFileMode is the file permission mode for certificate files.
	PKIInitCertFileMode = os.FileMode(0o644)

	// PKIInitCertsDirMode is the directory permission mode for the certs directory.
	PKIInitCertsDirMode = os.FileMode(0o755)
)
