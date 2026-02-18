// Copyright (c) 2025 Justin Cranford
//
//

package magic

// Docker service states.
const (
	// DockerServiceStateRunning - Docker service running state.
	DockerServiceStateRunning = "running"
	// DockerServiceStateExited - Docker service exited state.
	DockerServiceStateExited = "exited"
	// DockerServiceHealthHealthy - Docker service healthy status.
	DockerServiceHealthHealthy = "healthy"
)

// Docker service names.
const (
	// DockerJobHealthcheckOtelCollectorContrib - Healthcheck for OpenTelemetry Collector Contrib job name.
	DockerJobHealthcheckOtelCollectorContrib = "healthcheck-opentelemetry-collector-contrib"
	// DockerJobHealthcheckSecrets - Healthcheck for secrets job name.
	DockerJobHealthcheckSecrets = "healthcheck-secrets"

	// DockerJobBuilderCryptoutil - Builder for Cryptoutil job name.
	DockerJobBuilderCryptoutil = "builder-cryptoutil"

	// DockerServicePostgres - PostgreSQL service name.
	DockerServicePostgres = "postgres"

	// DockerServiceGrafanaOtelLgtm - Grafana OTEL LGTM service name.
	DockerServiceGrafanaOtelLgtm = "grafana-otel-lgtm"
	// DockerServiceOtelCollector - OpenTelemetry collector service name.
	DockerServiceOtelCollector = "opentelemetry-collector-contrib"

	// DockerServiceCryptoutilSqlite - Cryptoutil SQLite service name.
	DockerServiceCryptoutilSqlite = "cryptoutil-sqlite"
	// DockerServiceCryptoutilPostgres1 - Cryptoutil PostgreSQL 1 service name.
	DockerServiceCryptoutilPostgres1 = "cryptoutil-postgres-1"
	// DockerServiceCryptoutilPostgres2 - Cryptoutil PostgreSQL 2 service name.
	DockerServiceCryptoutilPostgres2 = "cryptoutil-postgres-2"
)

// Docker-related magic numbers.
const (
	// DockerServiceNamePartsMin - Minimum number of parts in a Docker service name.
	DockerServiceNamePartsMin = 3
	// DockerHTTPClientTimeoutSeconds - HTTP client timeout for Docker operations in seconds.
	DockerHTTPClientTimeoutSeconds = 5
)

// Docker Compose relative file paths from project root.
// NOTE: Legacy E2E compose was archived during PRODUCT-SERVICE restructuring.
// See deployments/archived/compose-legacy/ for the original monolithic compose.
const (
	// DockerComposeRelativeFilePathWindows - Docker Compose relative file path for Windows from project root (archived).
	DockerComposeRelativeFilePathWindows = ".\\deployments\\archived\\compose-legacy\\compose.yml"
	// DockerComposeRelativeFilePathLinux - Docker Compose relative file path for Linux/Mac from project root (archived).
	DockerComposeRelativeFilePathLinux = "./deployments/archived/compose-legacy/compose.yml"
)
