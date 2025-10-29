// Package magic provides commonly used magic numbers and values as named constants.
// This file contains Docker-related constants.
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
	// DockerServiceCryptoutilSqlite - Cryptoutil SQLite service name.
	DockerServiceCryptoutilSqlite = "cryptoutil-sqlite"
	// DockerServiceCryptoutilPostgres1 - Cryptoutil PostgreSQL 1 service name.
	DockerServiceCryptoutilPostgres1 = "cryptoutil-postgres-1"
	// DockerServiceCryptoutilPostgres2 - Cryptoutil PostgreSQL 2 service name.
	DockerServiceCryptoutilPostgres2 = "cryptoutil-postgres-2"
	// DockerServicePostgres - PostgreSQL service name.
	DockerServicePostgres = "postgres"
	// DockerServiceGrafanaOtelLgtm - Grafana OTEL LGTM service name.
	DockerServiceGrafanaOtelLgtm = "grafana-otel-lgtm"
	// DockerServiceOtelCollector - OpenTelemetry collector service name.
	DockerServiceOtelCollector = "opentelemetry-collector-contrib"
	// DockerJobOtelCollectorHealthcheck - OpenTelemetry collector healthcheck job name.
	DockerJobOtelCollectorHealthcheck = "opentelemetry-collector-contrib-healthcheck"
)

// Docker-related magic numbers.
const (
	// DockerServiceNamePartsMin - Minimum number of parts in a Docker service name.
	DockerServiceNamePartsMin = 3
	// DockerHTTPClientTimeoutSeconds - HTTP client timeout for Docker operations in seconds.
	DockerHTTPClientTimeoutSeconds = 5
)

// Docker Compose relative file paths from project root.
const (
	// DockerComposeRelativeFilePathWindows - Docker Compose relative file path for Windows from project root.
	DockerComposeRelativeFilePathWindows = ".\\deployments\\compose\\compose.yml"
	// DockerComposeRelativeFilePathLinux - Docker Compose relative file path for Linux/Mac from project root.
	DockerComposeRelativeFilePathLinux = "./deployments/compose/compose.yml"
)
