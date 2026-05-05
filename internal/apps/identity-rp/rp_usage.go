// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
package rp

import (
	"fmt"

	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// RPUsageMain is the main usage message for the identity-rp command.
	RPUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
		cryptoutilSharedMagic.RPDisplayName,
	)

	// RPUsageServer is the usage message for the server subcommand.
	RPUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
		cryptoutilSharedMagic.RPDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityRPServiceID, cryptoutilSharedMagic.IdentityRPServiceID),
	)

	// RPUsageClient is the usage message for the client subcommand.
	RPUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
		cryptoutilSharedMagic.RPDisplayName,
	)

	// RPUsageInit is the usage message for the init subcommand.
	RPUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
		cryptoutilSharedMagic.RPDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityRPServiceID, cryptoutilSharedMagic.IdentityRPServiceID),
	)

	// RPUsageHealth is the usage message for the health subcommand.
	RPUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.IdentityRPServicePort),
	)

	// RPUsageLivez is the usage message for the livez subcommand.
	RPUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
	)

	// RPUsageReadyz is the usage message for the readyz subcommand.
	RPUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
	)

	// RPUsageShutdown is the usage message for the shutdown subcommand.
	RPUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.RPServiceName,
	)
)
