//go:build ignore

// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
package __SERVICE__

import (
	"fmt"

	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// __USAGE_PREFIX__UsageMain is the main usage message for the __PS_ID__ command.
	__USAGE_PREFIX__UsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
		"__SERVICE_DISPLAY_NAME__",
	)

	// __USAGE_PREFIX__UsageServer is the usage message for the server subcommand.
	__USAGE_PREFIX__UsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
		"__SERVICE_DISPLAY_NAME__",
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.__SERVICE_ID_CONST__, cryptoutilSharedMagic.__SERVICE_ID_CONST__),
	)

	// __USAGE_PREFIX__UsageClient is the usage message for the client subcommand.
	__USAGE_PREFIX__UsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
		"__SERVICE_DISPLAY_NAME__",
	)

	// __USAGE_PREFIX__UsageInit is the usage message for the init subcommand.
	__USAGE_PREFIX__UsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
		"__SERVICE_DISPLAY_NAME__",
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.__SERVICE_ID_CONST__, cryptoutilSharedMagic.__SERVICE_ID_CONST__),
	)

	// __USAGE_PREFIX__UsageHealth is the usage message for the health subcommand.
	__USAGE_PREFIX__UsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
		fmt.Sprintf("%d", cryptoutilSharedMagic.__SERVICE_PORT_CONST__),
	)

	// __USAGE_PREFIX__UsageLivez is the usage message for the livez subcommand.
	__USAGE_PREFIX__UsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
	)

	// __USAGE_PREFIX__UsageReadyz is the usage message for the readyz subcommand.
	__USAGE_PREFIX__UsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
	)

	// __USAGE_PREFIX__UsageShutdown is the usage message for the shutdown subcommand.
	__USAGE_PREFIX__UsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.__PRODUCT_NAME_CONST__,
		cryptoutilSharedMagic.__SERVICE_NAME_CONST__,
	)
)
