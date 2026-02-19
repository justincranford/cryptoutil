// Copyright (c) 2025 Justin Cranford
//

// Package sm implements the sm (Secrets Manager) product command router.
package sm

import (
	"io"

	cryptoutilAppsSmKms "cryptoutil/internal/apps/sm/kms"
	cryptoutilTemplateCli "cryptoutil/internal/apps/template/service/cli"
)

const (
	usageText   = "Usage: sm <service> <subcommand> [options]\n\nAvailable services:\n  kms         Key Management Service\n\nUse \"sm <service> help\" for service-specific help.\nUse \"sm version\" for version information."
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
	return cryptoutilTemplateCli.RouteProduct(
		cryptoutilTemplateCli.ProductConfig{
			ProductName: "sm",
			UsageText:   usageText,
			VersionText: versionText,
		},
		args, stdin, stdout, stderr,
		[]cryptoutilTemplateCli.ServiceEntry{
			{Name: "kms", Handler: cryptoutilAppsSmKms.Kms},
		},
	)
}
