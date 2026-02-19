// Copyright (c) 2025 Justin Cranford
//

// Package identity implements the identity product command router.
package identity

import (
	"io"

	cryptoutilAppsIdentityAuthz "cryptoutil/internal/apps/identity/authz"
	cryptoutilAppsIdentityIdp "cryptoutil/internal/apps/identity/idp"
	cryptoutilAppsIdentityRp "cryptoutil/internal/apps/identity/rp"
	cryptoutilAppsIdentityRs "cryptoutil/internal/apps/identity/rs"
	cryptoutilAppsIdentitySpa "cryptoutil/internal/apps/identity/spa"
	cryptoutilTemplateCli "cryptoutil/internal/apps/template/service/cli"
)

const (
	usageText   = "Usage: identity <service> <subcommand> [options]\n\nAvailable services:\n  authz       OAuth 2.1 Authorization Server\n  idp         OIDC 1.0 Identity Provider\n  rp          OAuth 2.1 Relying Party (BFF reference implementation)\n  rs          OAuth 2.1 Resource Server\n  spa         Single Page Application (SPA reference implementation)\n\nUse \"identity <service> help\" for service-specific help.\nUse \"identity version\" for version information."
	versionText = "identity product (cryptoutil)"
)

// Identity implements the identity product command router.
// Supports Suite, Product, and Product-Service patterns.
//
// Call patterns:
// - Suite: cryptoutil identity authz server
// - Product: identity authz server
// - Product-Service: identity-authz server (via main.go delegation).
func Identity(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	return cryptoutilTemplateCli.RouteProduct(
		cryptoutilTemplateCli.ProductConfig{
			ProductName: "identity",
			UsageText:   usageText,
			VersionText: versionText,
		},
		args, stdin, stdout, stderr,
		[]cryptoutilTemplateCli.ServiceEntry{
			{Name: "authz", Handler: cryptoutilAppsIdentityAuthz.Authz},
			{Name: "idp", Handler: cryptoutilAppsIdentityIdp.Idp},
			{Name: "rp", Handler: cryptoutilAppsIdentityRp.Rp},
			{Name: "rs", Handler: cryptoutilAppsIdentityRs.Rs},
			{Name: "spa", Handler: cryptoutilAppsIdentitySpa.Spa},
		},
	)
}
