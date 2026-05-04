// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
package template

import (
	"fmt"

	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// TemplateUsageMain is the main usage message for the skeleton-template command.
	TemplateUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
		"Skeleton Template",
	)

	// TemplateUsageServer is the usage message for the server subcommand.
	TemplateUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
		"Skeleton Template",
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.SkeletonTemplateServiceID, cryptoutilSharedMagic.SkeletonTemplateServiceID),
	)

	// TemplateUsageClient is the usage message for the client subcommand.
	TemplateUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
		"Skeleton Template",
	)

	// TemplateUsageInit is the usage message for the init subcommand.
	TemplateUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
		"Skeleton Template",
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.SkeletonTemplateServiceID, cryptoutilSharedMagic.SkeletonTemplateServiceID),
	)

	// TemplateUsageHealth is the usage message for the health subcommand.
	TemplateUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.SkeletonTemplateServicePort),
	)

	// TemplateUsageLivez is the usage message for the livez subcommand.
	TemplateUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
	)

	// TemplateUsageReadyz is the usage message for the readyz subcommand.
	TemplateUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
	)

	// TemplateUsageShutdown is the usage message for the shutdown subcommand.
	TemplateUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.SkeletonProductName,
		cryptoutilSharedMagic.SkeletonTemplateServiceName,
	)
)
