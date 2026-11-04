// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

const (
	DevelopmentSmokeRun      = "p2sm1103"
	DevelopmentSmokeSelector = "s:sm-kms:mem"
	developmentServerUsage   = cryptoutilSharedMagic.CMD_SERVER
	developmentClientUsage   = "client"
	developmentCollectorTLS  = "collector-server-tls"
	developmentOwnerLabel    = "cu.io/owner"
	developmentDirectoryMode = 0o700
	developmentCollectorName = "otel-collector"
	developmentSecretOwner   = "external-secret-import"
)
