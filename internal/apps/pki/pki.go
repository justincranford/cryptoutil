// Copyright (c) 2025 Justin Cranford
//
//

// Package pki implements the pki product command router.
package pki

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"io"

	cryptoutilAppsPkiCa "cryptoutil/internal/apps/pki/ca"
	cryptoutilTemplateCli "cryptoutil/internal/apps/template/service/cli"
)

const (
	usageText   = "Usage: pki <service> <subcommand> [options]\n\nAvailable services:\n  ca          Certificate Authority service\n\nUse \"pki <service> help\" for service-specific help.\nUse \"pki version\" for version information."
	versionText = "pki product (cryptoutil)"
)

// Pki implements the pki product command router.
// Supports Suite, Product, and Product-Service patterns.
//
// Call patterns:
// - Suite: cryptoutil pki ca server
// - Product: pki ca server
// - Product-Service: pki-ca server (via main.go delegation).
func Pki(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteProduct(
		cryptoutilTemplateCli.ProductConfig{
			ProductName: cryptoutilSharedMagic.PKIProductName,
			UsageText:   usageText,
			VersionText: versionText,
		},
		args, stdin, stdout, stderr,
		[]cryptoutilTemplateCli.ServiceEntry{
			{Name: "ca", Handler: cryptoutilAppsPkiCa.Ca},
		},
	)
}
