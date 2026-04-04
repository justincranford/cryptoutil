// Copyright (c) 2025 Justin Cranford
//
//

// Package cryptoutil provides command-line interface for cryptoutil suite operations.
package cryptoutil

import (
	"io"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilAppsFrameworkSuiteCli "cryptoutil/internal/apps/framework/suite/cli"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps/framework/tls"
	cryptoutilAppsIdentity "cryptoutil/internal/apps/identity"
	cryptoutilAppsJose "cryptoutil/internal/apps/jose"
	cryptoutilAppsPki "cryptoutil/internal/apps/pki"
	cryptoutilAppsSkeleton "cryptoutil/internal/apps/skeleton"
	cryptoutilAppsSm "cryptoutil/internal/apps/sm"
)

const (
	suiteUsageText = "Usage: cryptoutil <product> [service] [options]\n\nAvailable products:\n  identity    Identity product (OAuth 2.1, OIDC 1.0)\n  jose        JOSE product (JWK/JWS/JWE/JWT operations)\n  pki         PKI product (X.509 certificates, CA)\n  skeleton    Skeleton product (service template demonstration)\n  sm          Secrets Manager product (KMS, IM)\n  pki-init    PKI Init (generate TLS certificates for Docker Compose E2E deployments)\n\nUse \"cryptoutil <product> help\" for product-specific help."
)

// Suite runs the cryptoutil suite command-line interface.
// This is the entry point for the suite-level CLI.
func Suite(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var suiteArgs []string
	if len(args) > 0 {
		suiteArgs = args[1:]
	}

	return cryptoutilAppsFrameworkSuiteCli.RouteSuite(
		cryptoutilAppsFrameworkSuiteCli.SuiteConfig{UsageText: suiteUsageText},
		suiteArgs, stdin, stdout, stderr,
		[]cryptoutilAppsFrameworkSuiteCli.ProductEntry{
			{Name: cryptoutilSharedMagic.IdentityProductName, Handler: cryptoutilAppsIdentity.Identity},
			{Name: cryptoutilSharedMagic.JoseProductName, Handler: cryptoutilAppsJose.Jose},
			{Name: cryptoutilSharedMagic.PKIProductName, Handler: cryptoutilAppsPki.Pki},
			{Name: cryptoutilSharedMagic.SkeletonProductName, Handler: cryptoutilAppsSkeleton.Skeleton},
			{Name: cryptoutilSharedMagic.SMProductName, Handler: cryptoutilAppsSm.Sm},
			{Name: cryptoutilSharedMagic.PSIDPKIInit, Handler: cryptoutilAppsFrameworkTls.Init},
		},
	)
}
