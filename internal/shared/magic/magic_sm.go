// Copyright (c) 2025 Justin Cranford

// Package magic contains magic constants for SM KMS service.
package magic

import "time"

// SM KMS service constants.
const (
	// OTLPServiceSMKMS is the OTLP service name for sm-kms telemetry.
	OTLPServiceSMKMS = "sm-kms"

	// KMSServiceID is the canonical service identifier for sm-kms.
	KMSServiceID = OTLPServiceSMKMS

	// SMProductName is the product name component of the SM product.
	SMProductName = "sm"

	// KMSServiceName is the service name component of the sm-kms service.
	KMSServiceName = "kms"

	// KMSServicePort is the default public API port for sm-kms service.
	// Port range for SM KMS: 8000-8099 (100-port block).
	KMSServicePort = 8000
)

// SM-KMS E2E Test Configuration.
const (
	// KMSE2EComposeFile is the path to the sm-kms docker compose file (relative from e2e test directory).
	// Path: internal/apps/sm/kms/e2e → ../../../../../deployments/sm-kms/compose.yml
	// Levels: e2e→kms(1)→sm(2)→apps(3)→internal(4)→root(5), then deployments/sm-kms.
	KMSE2EComposeFile = "../../../../../deployments/sm-kms/compose.yml"

	// KMSE2ESQLiteContainer is the SQLite instance service name in compose.
	KMSE2ESQLiteContainer = "sm-kms-app-sqlite-1"

	// KMSE2EPostgreSQL1Container is the PostgreSQL instance 1 service name in compose.
	KMSE2EPostgreSQL1Container = "sm-kms-app-postgres-1"

	// KMSE2EPostgreSQL2Container is the PostgreSQL instance 2 service name in compose.
	KMSE2EPostgreSQL2Container = "sm-kms-app-postgres-2"

	// KMSE2EHealthTimeout is the timeout for health checks during E2E tests.
	KMSE2EHealthTimeout = 180 * time.Second

	// KMSE2EHealthPollInterval is the interval between health check attempts.
	KMSE2EHealthPollInterval = 2 * time.Second

	// KMSE2ESQLitePublicPort is the SQLite instance public HTTPS port (SERVICE level: 8000).
	KMSE2ESQLitePublicPort = 8000

	// KMSE2EPostgreSQL1PublicPort is the PostgreSQL instance 1 public HTTPS port (SERVICE level: 8001).
	KMSE2EPostgreSQL1PublicPort = 8001

	// KMSE2EPostgreSQL2PublicPort is the PostgreSQL instance 2 public HTTPS port (SERVICE level: 8002).
	KMSE2EPostgreSQL2PublicPort = 8002

	// KMSE2EHealthEndpoint is the public health check endpoint.
	KMSE2EHealthEndpoint = "/service/api/v1/health"
)
