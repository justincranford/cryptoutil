// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
package im

import (
	"fmt"

	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// IMUsageMain is the main usage message for the sm-im command.
	IMUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.IMProductName,
		cryptoutilSharedMagic.IMServiceName,
		"Instant Messenger",
	)

	// IMUsageServer is the usage message for the server subcommand.
	IMUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.IMProductName,
		cryptoutilSharedMagic.IMServiceName,
		"Instant Messenger",
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IMServiceID, cryptoutilSharedMagic.IMServiceID),
	)

	// IMUsageClient is the usage message for the client subcommand.
	IMUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.IMProductName,
		cryptoutilSharedMagic.IMServiceName,
		"Instant Messenger",
	)

	// IMUsageInit is the usage message for the init subcommand.
	IMUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.IMProductName,
		cryptoutilSharedMagic.IMServiceName,
		"Instant Messenger",
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IMServiceID, cryptoutilSharedMagic.IMServiceID),
	)

	// IMUsageHealth is the usage message for the health subcommand.
	IMUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.IMProductName,
		cryptoutilSharedMagic.IMServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.IMServicePort),
	)

	// IMUsageLivez is the usage message for the livez subcommand.
	IMUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.IMProductName,
		cryptoutilSharedMagic.IMServiceName,
	)

	// IMUsageReadyz is the usage message for the readyz subcommand.
	IMUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.IMProductName,
		cryptoutilSharedMagic.IMServiceName,
	)

	// IMUsageShutdown is the usage message for the shutdown subcommand.
	IMUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.IMProductName,
		cryptoutilSharedMagic.IMServiceName,
	)
)
