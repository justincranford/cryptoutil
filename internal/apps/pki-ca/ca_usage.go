// Copyright (c) 2025 Justin Cranford

// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
package ca

import (
	"fmt"

	cryptoutilUsage "cryptoutil/internal/apps/framework/service/usage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// CAUsageMain is the main usage message for the pki ca command.
	CAUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
		"Certificate Authority",
	)

	// CAUsageServer is the usage message for the server subcommand.
	CAUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
		"Certificate Authority",
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.PKICAServiceID, cryptoutilSharedMagic.PKICAServiceID),
	)

	// CAUsageClient is the usage message for the client subcommand.
	CAUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
		"Certificate Authority",
	)

	// CAUsageInit is the usage message for the init subcommand.
	CAUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
		"Certificate Authority",
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.PKICAServiceID, cryptoutilSharedMagic.PKICAServiceID),
	)

	// CAUsageHealth is the usage message for the health subcommand.
	CAUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.PKICAServicePort),
	)

	// CAUsageLivez is the usage message for the livez subcommand.
	CAUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
	)

	// CAUsageReadyz is the usage message for the readyz subcommand.
	CAUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
	)

	// CAUsageShutdown is the usage message for the shutdown subcommand.
	CAUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.PKIProductName,
		cryptoutilSharedMagic.PKICAServiceName,
	)
)
