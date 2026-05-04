// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
package spa

import (
	"fmt"

	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// SPAUsageMain is the main usage message for the identity-spa command.
	SPAUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
		"Single Page App",
	)

	// SPAUsageServer is the usage message for the server subcommand.
	SPAUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
		"Single Page App",
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentitySPAServiceID, cryptoutilSharedMagic.IdentitySPAServiceID),
	)

	// SPAUsageClient is the usage message for the client subcommand.
	SPAUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
		"Single Page App",
	)

	// SPAUsageInit is the usage message for the init subcommand.
	SPAUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
		"Single Page App",
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentitySPAServiceID, cryptoutilSharedMagic.IdentitySPAServiceID),
	)

	// SPAUsageHealth is the usage message for the health subcommand.
	SPAUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.IdentitySPAServicePort),
	)

	// SPAUsageLivez is the usage message for the livez subcommand.
	SPAUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
	)

	// SPAUsageReadyz is the usage message for the readyz subcommand.
	SPAUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
	)

	// SPAUsageShutdown is the usage message for the shutdown subcommand.
	SPAUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.SPAServiceName,
	)
)
