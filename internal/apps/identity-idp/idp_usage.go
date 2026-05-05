// Copyright (c) 2025-2026 Justin Cranford.
// Health endpoints exposed by this service (referenced in BuildUsage* output):
//   - /service/api/v1/health  (service-to-service health check)
//   - /browser/api/v1/health  (browser health check)
//   - /admin/api/v1/livez     (liveness probe)
//   - /admin/api/v1/readyz    (readiness probe)
//   - /admin/api/v1/shutdown  (graceful shutdown trigger)
package idp

import (
	"fmt"

	cryptoutilUsage "cryptoutil/internal/apps-framework/service/usage"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

var (
	// IDPUsageMain is the main usage message for the identity-idp command.
	IDPUsageMain = cryptoutilUsage.BuildUsageMain(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
		cryptoutilSharedMagic.IDPDisplayName,
	)

	// IDPUsageServer is the usage message for the server subcommand.
	IDPUsageServer = cryptoutilUsage.BuildUsageServer(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
		cryptoutilSharedMagic.IDPDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityIDPServiceID, cryptoutilSharedMagic.IdentityIDPServiceID),
	)

	// IDPUsageClient is the usage message for the client subcommand.
	IDPUsageClient = cryptoutilUsage.BuildUsageClient(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
		cryptoutilSharedMagic.IDPDisplayName,
	)

	// IDPUsageInit is the usage message for the init subcommand.
	IDPUsageInit = cryptoutilUsage.BuildUsageInit(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
		cryptoutilSharedMagic.IDPDisplayName,
		fmt.Sprintf("configs/%s/%s-framework.yml", cryptoutilSharedMagic.IdentityIDPServiceID, cryptoutilSharedMagic.IdentityIDPServiceID),
	)

	// IDPUsageHealth is the usage message for the health subcommand.
	IDPUsageHealth = cryptoutilUsage.BuildUsageHealth(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
		fmt.Sprintf("%d", cryptoutilSharedMagic.IdentityIDPServicePort),
	)

	// IDPUsageLivez is the usage message for the livez subcommand.
	IDPUsageLivez = cryptoutilUsage.BuildUsageLivez(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
	)

	// IDPUsageReadyz is the usage message for the readyz subcommand.
	IDPUsageReadyz = cryptoutilUsage.BuildUsageReadyz(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
	)

	// IDPUsageShutdown is the usage message for the shutdown subcommand.
	IDPUsageShutdown = cryptoutilUsage.BuildUsageShutdown(
		cryptoutilSharedMagic.IdentityProductName,
		cryptoutilSharedMagic.IDPServiceName,
	)
)
