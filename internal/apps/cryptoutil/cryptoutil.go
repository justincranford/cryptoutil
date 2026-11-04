// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package cryptoutil provides command-line interface for cryptoutil suite operations.
package cryptoutil

import (
	"io"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilAppsFrameworkSuiteCli "cryptoutil/internal/apps-framework/suite/cli"
	cryptoutilAppsFrameworkTls "cryptoutil/internal/apps-framework/tls"
	cryptoutilDeploymentDatabase "cryptoutil/internal/apps-tools/deploymentrun/database"
	cryptoutilDeploymentMaterializer "cryptoutil/internal/apps-tools/deploymentrun/materializer"
	cryptoutilAppsIdentity "cryptoutil/internal/apps/identity"
	cryptoutilAppsSkeleton "cryptoutil/internal/apps/skeleton"
	cryptoutilAppsSm "cryptoutil/internal/apps/sm"
)

const (
	suiteUsageText = "Usage: cryptoutil <product> [service] [options]\n\nAvailable products:\n" +
		"  identity    Identity product (OAuth 2.1, OIDC 1.0)\n" +
		"  skeleton    Skeleton product (service template demonstration)\n" +
		"  sm          Secrets Manager product (KMS)\n" +
		"  pki-init    PKI Init (generate TLS certificates)\n" +
		"  pki-materialize    Verify and materialize explicitly bound pod PKI\n" +
		"  pki-verify    Read-only verification of host-prepublished PKI\n" +
		"  db-lifecycle    Explicit database initialization, migrations and replication\n" +
		"\nUse \"cryptoutil <product> help\" for product-specific help."
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
			{Name: cryptoutilSharedMagic.SkeletonProductName, Handler: cryptoutilAppsSkeleton.Skeleton},
			{Name: cryptoutilSharedMagic.SMProductName, Handler: cryptoutilAppsSm.Sm},
			{Name: cryptoutilSharedMagic.PSIDPKIInit, Handler: cryptoutilAppsFrameworkTls.Init},
			{Name: cryptoutilDeploymentMaterializer.Command, Handler: cryptoutilDeploymentMaterializer.Main},
			{Name: cryptoutilDeploymentMaterializer.VerifyCommand, Handler: cryptoutilDeploymentMaterializer.VerifyMain},
			{Name: cryptoutilDeploymentDatabase.Command, Handler: cryptoutilDeploymentDatabase.Main},
		},
	)
}
