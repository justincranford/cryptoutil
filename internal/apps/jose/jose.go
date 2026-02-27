// Copyright (c) 2025 Justin Cranford
//

// Package jose implements the jose product command router.
package jose

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"io"

	cryptoutilAppsJoseJa "cryptoutil/internal/apps/jose/ja"
	cryptoutilTemplateCli "cryptoutil/internal/apps/template/service/cli"
)

const (
	usageText   = "Usage: jose <service> <subcommand> [options]\n\nAvailable services:\n  ja          JWK Authority service\n\nUse \"jose <service> help\" for service-specific help.\nUse \"jose version\" for version information."
	versionText = "jose product (cryptoutil)"
)

// Jose implements the jose product command router.
// Supports Suite, Product, and Product-Service patterns.
//
// Call patterns:
// - Suite: cryptoutil jose ja server
// - Product: jose ja server
// - Product-Service: jose-ja server (via main.go delegation).
func Jose(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteProduct(
		cryptoutilTemplateCli.ProductConfig{
			ProductName: cryptoutilSharedMagic.JoseProductName,
			UsageText:   usageText,
			VersionText: versionText,
		},
		args, stdin, stdout, stderr,
		[]cryptoutilTemplateCli.ServiceEntry{
			{Name: cryptoutilSharedMagic.JoseJAServiceName, Handler: cryptoutilAppsJoseJa.Ja},
		},
	)
}
