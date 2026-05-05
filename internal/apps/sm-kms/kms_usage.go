// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
package kms

import (
	"fmt"

	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// KMSUsageMain is the main usage message for the sm-kms command.
	KMSUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
		cryptoutilSharedMagic.KMSDisplayName,
	)

	// KMSUsageServer is the usage message for the server subcommand.
	KMSUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
		cryptoutilSharedMagic.KMSDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.KMSServiceID, cryptoutilSharedMagic.KMSServiceID),
	)

	// KMSUsageClient is the usage message for the client subcommand.
	KMSUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
		cryptoutilSharedMagic.KMSDisplayName,
	)

	// KMSUsageInit is the usage message for the init subcommand.
	KMSUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
		cryptoutilSharedMagic.KMSDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.KMSServiceID, cryptoutilSharedMagic.KMSServiceID),
	)

	// KMSUsageHealth is the usage message for the health subcommand.
	KMSUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.KMSServicePort),
	)

	// KMSUsageLivez is the usage message for the livez subcommand.
	KMSUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
	)

	// KMSUsageReadyz is the usage message for the readyz subcommand.
	KMSUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
	)

	// KMSUsageShutdown is the usage message for the shutdown subcommand.
	KMSUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.SMProductName,
		cryptoutilSharedMagic.KMSServiceName,
	)
)
