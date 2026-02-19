// Copyright (c) 2025 Justin Cranford
//
//

// Package cipher implements the cipher product command router.
package cipher

import (
	"io"

	cryptoutilAppsCipherIm "cryptoutil/internal/apps/cipher/im"
	cryptoutilTemplateCli "cryptoutil/internal/apps/template/service/cli"
)

const (
	usageText   = "Usage: cipher <service> <subcommand> [options]\n\nAvailable services:\n  im          Instant messaging service\n\nUse \"cipher <service> help\" for service-specific help.\nUse \"cipher version\" for version information."
	versionText = "cipher product (cryptoutil)"
)

// Cipher implements the cipher product command router.
// Supports Suite, Product, and Product-Service patterns.
//
// Call patterns:
// - Suite: cryptoutil cipher im server
// - Product: cipher im server
// - Product-Service: cipher-im server (via main.go delegation).
func Cipher(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteProduct(
		cryptoutilTemplateCli.ProductConfig{
			ProductName: "cipher",
			UsageText:   usageText,
			VersionText: versionText,
		},
		args, stdin, stdout, stderr,
		[]cryptoutilTemplateCli.ServiceEntry{
			{Name: "im", Handler: cryptoutilAppsCipherIm.Im},
		},
	)
}
