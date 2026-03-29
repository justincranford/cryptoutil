// Copyright (c) 2025 Justin Cranford
//

// Package sm implements the sm (Secrets Manager) product command router.
package sm

import (
	"io"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	cryptoutilAppsFrameworkProductCli "cryptoutil/internal/apps/framework/product/cli"
	cryptoutilAppsSmIm "cryptoutil/internal/apps/sm-im"
	cryptoutilAppsSmKms "cryptoutil/internal/apps/sm-kms"
)

const (
	usageText   = "Usage: sm <service> <subcommand> [options]\n\nAvailable services:\n  kms         Key Management Service\n  im          Instant messaging service\n\nUse \"sm <service> help\" for service-specific help.\nUse \"sm version\" for version information."
	versionText = "sm product (cryptoutil)"
)

// Sm implements the sm (Secrets Manager) product command router.
// Supports Suite, Product, and Product-Service patterns.
//
// Call patterns:
// - Suite: cryptoutil sm kms server
// - Product: sm kms server
// - Product-Service: sm-kms server (via main.go delegation).
func Sm(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilAppsFrameworkProductCli.RouteProduct(
		cryptoutilAppsFrameworkProductCli.ProductConfig{
			ProductName: cryptoutilSharedMagic.SMProductName,
			UsageText:   usageText,
			VersionText: versionText,
		},
		args, stdin, stdout, stderr,
		[]cryptoutilAppsFrameworkProductCli.ServiceEntry{
			{Name: cryptoutilSharedMagic.KMSServiceName, Handler: cryptoutilAppsSmKms.Kms},
			{Name: cryptoutilSharedMagic.IMServiceName, Handler: cryptoutilAppsSmIm.Im},
		},
	)
}
