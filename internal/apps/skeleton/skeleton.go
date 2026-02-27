// Copyright (c) 2025 Justin Cranford
//

// Package skeleton implements the skeleton product command router.
package skeleton

import (
cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
"io"

cryptoutilAppsSkeletonTemplate "cryptoutil/internal/apps/skeleton/template"
cryptoutilTemplateCli "cryptoutil/internal/apps/template/service/cli"
)

const (
usageText   = "Usage: skeleton <service> <subcommand> [options]\n\nAvailable services:\n  template    Skeleton Template service\n\nUse \"skeleton <service> help\" for service-specific help.\nUse \"skeleton version\" for version information."
versionText = "skeleton product (cryptoutil)"
)

// Skeleton implements the skeleton product command router.
// Supports Suite, Product, and Product-Service patterns.
//
// Call patterns:
// - Suite: cryptoutil skeleton template server
// - Product: skeleton template server
// - Product-Service: skeleton-template server (via main.go delegation).
func Skeleton(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
return cryptoutilTemplateCli.RouteProduct(
cryptoutilTemplateCli.ProductConfig{
ProductName: cryptoutilSharedMagic.SkeletonProductName,
UsageText:   usageText,
VersionText: versionText,
},
args, stdin, stdout, stderr,
[]cryptoutilTemplateCli.ServiceEntry{
{Name: cryptoutilSharedMagic.SkeletonTemplateServiceName, Handler: cryptoutilAppsSkeletonTemplate.Template},
},
)
}
