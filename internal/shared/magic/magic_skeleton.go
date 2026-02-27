// Copyright (c) 2025 Justin Cranford

package magic

import "time"

// Skeleton-Template Service Configuration.
const (
	// OTLPServiceSkeletonTemplate is the OTLP service name for skeleton-template telemetry.
	OTLPServiceSkeletonTemplate = "skeleton-template"

	// SkeletonTemplateServiceID is the canonical service identifier for skeleton-template.
	SkeletonTemplateServiceID = OTLPServiceSkeletonTemplate

	// SkeletonProductName is the product name component of the Skeleton product.
	SkeletonProductName = "skeleton"

	// SkeletonTemplateServiceName is the service name component of the skeleton-template service.
	SkeletonTemplateServiceName = "template"

	// SkeletonTemplateServicePort is the default public API port for skeleton-template service.
	SkeletonTemplateServicePort = 8900

	// SkeletonTemplateAdminPort is the admin API port (same for all services).
	SkeletonTemplateAdminPort = 9090

	// SkeletonTemplatePostgresPort is the host PostgreSQL port for skeleton-template.
	SkeletonTemplatePostgresPort = 54329
)

// Skeleton-Template E2E Test Configuration.
const (
	// SkeletonTemplateE2EComposeFile is the path to the skeleton docker compose file (relative from e2e test directory).
	// Path: internal/apps/skeleton/template/e2e → ../../../../../deployments/skeleton/compose.yml.
	SkeletonTemplateE2EComposeFile = "../../../../../deployments/skeleton/compose.yml"

	// SkeletonTemplateE2ESQLiteContainer is the SQLite instance service name in compose.
	SkeletonTemplateE2ESQLiteContainer = "skeleton-template-app-sqlite-1"

	// SkeletonTemplateE2EPostgreSQL1Container is the PostgreSQL instance 1 service name in compose.
	SkeletonTemplateE2EPostgreSQL1Container = "skeleton-template-app-postgres-1"

	// SkeletonTemplateE2EPostgreSQL2Container is the PostgreSQL instance 2 service name in compose.
	SkeletonTemplateE2EPostgreSQL2Container = "skeleton-template-app-postgres-2"

	// SkeletonTemplateE2EHealthTimeout is the timeout for health checks during E2E tests.
	SkeletonTemplateE2EHealthTimeout = 180 * time.Second

	// SkeletonTemplateE2EHealthPollInterval is the interval between health check attempts.
	SkeletonTemplateE2EHealthPollInterval = 2 * time.Second

	// SkeletonTemplateE2ESQLitePublicPort is the SQLite instance public HTTPS port (PRODUCT level: 18900).
	SkeletonTemplateE2ESQLitePublicPort = 18900

	// SkeletonTemplateE2EPostgreSQL1PublicPort is the PostgreSQL instance 1 public HTTPS port (PRODUCT level: 18901).
	SkeletonTemplateE2EPostgreSQL1PublicPort = 18901

	// SkeletonTemplateE2EPostgreSQL2PublicPort is the PostgreSQL instance 2 public HTTPS port (PRODUCT level: 18902).
	SkeletonTemplateE2EPostgreSQL2PublicPort = 18902

	// SkeletonTemplateE2EGrafanaPort is the Grafana port for E2E tests.
	SkeletonTemplateE2EGrafanaPort = 3000

	// SkeletonTemplateE2EOtelCollectorGRPCPort is the OpenTelemetry collector gRPC port for E2E tests.
	SkeletonTemplateE2EOtelCollectorGRPCPort = 4317

	// SkeletonTemplateE2EOtelCollectorHTTPPort is the OpenTelemetry collector HTTP port for E2E tests.
	SkeletonTemplateE2EOtelCollectorHTTPPort = 4318

	// SkeletonTemplateE2EHealthEndpoint is the public health check endpoint.
	SkeletonTemplateE2EHealthEndpoint = "/service/api/v1/health"
)
