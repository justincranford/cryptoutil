// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
package authz

import (
	"fmt"

	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// AUTHZUsageMain is the main usage message for the identity-authz command.
	AUTHZUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
		cryptoutilSharedMagic.AuthzDisplayName,
	)

	// AUTHZUsageServer is the usage message for the server subcommand.
	AUTHZUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
		cryptoutilSharedMagic.AuthzDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityAuthzServiceID, cryptoutilSharedMagic.IdentityAuthzServiceID),
	)

	// AUTHZUsageClient is the usage message for the client subcommand.
	AUTHZUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
		cryptoutilSharedMagic.AuthzDisplayName,
	)

	// AUTHZUsageInit is the usage message for the init subcommand.
	AUTHZUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
		cryptoutilSharedMagic.AuthzDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityAuthzServiceID, cryptoutilSharedMagic.IdentityAuthzServiceID),
	)

	// AUTHZUsageHealth is the usage message for the health subcommand.
	AUTHZUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.IdentityAuthzServicePort),
	)

	// AUTHZUsageLivez is the usage message for the livez subcommand.
	AUTHZUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
	)

	// AUTHZUsageReadyz is the usage message for the readyz subcommand.
	AUTHZUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
	)

	// AUTHZUsageShutdown is the usage message for the shutdown subcommand.
	AUTHZUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.AuthzServiceName,
	)
)
