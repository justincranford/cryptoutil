// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
package ja

import (
	"fmt"

	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// JAUsageMain is the main usage message for the jose-ja command.
	JAUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
		"JWK Authority",
	)

	// JAUsageServer is the usage message for the server subcommand.
	JAUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
		"JWK Authority",
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.JoseJAServiceID, cryptoutilSharedMagic.JoseJAServiceID),
	)

	// JAUsageClient is the usage message for the client subcommand.
	JAUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
		"JWK Authority",
	)

	// JAUsageInit is the usage message for the init subcommand.
	JAUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
		"JWK Authority",
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.JoseJAServiceID, cryptoutilSharedMagic.JoseJAServiceID),
	)

	// JAUsageHealth is the usage message for the health subcommand.
	JAUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.JoseJAServicePort),
	)

	// JAUsageLivez is the usage message for the livez subcommand.
	JAUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
	)

	// JAUsageReadyz is the usage message for the readyz subcommand.
	JAUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
	)

	// JAUsageShutdown is the usage message for the shutdown subcommand.
	JAUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.JoseProductName,
		cryptoutilSharedMagic.JoseJAServiceName,
	)
)
