// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
package rs

import (
	"fmt"

	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// RSUsageMain is the main usage message for the identity-rs command.
	RSUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
		cryptoutilSharedMagic.RSDisplayName,
	)

	// RSUsageServer is the usage message for the server subcommand.
	RSUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
		cryptoutilSharedMagic.RSDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityRSServiceID, cryptoutilSharedMagic.IdentityRSServiceID),
	)

	// RSUsageClient is the usage message for the client subcommand.
	RSUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
		cryptoutilSharedMagic.RSDisplayName,
	)

	// RSUsageInit is the usage message for the init subcommand.
	RSUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
		cryptoutilSharedMagic.RSDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityRSServiceID, cryptoutilSharedMagic.IdentityRSServiceID),
	)

	// RSUsageHealth is the usage message for the health subcommand.
	RSUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.IdentityRSServicePort),
	)

	// RSUsageLivez is the usage message for the livez subcommand.
	RSUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
	)

	// RSUsageReadyz is the usage message for the readyz subcommand.
	RSUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
	)

	// RSUsageShutdown is the usage message for the shutdown subcommand.
	RSUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RSServiceName,
	)
)
