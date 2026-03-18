// Copyright (c) 2025 Justin Cranford

package e2e_infra

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDockerComposePsOutput_ValidJSON(t *testing.T) {
	t.Parallel()

	input := `{"Name":"compose-cryptoutil-sqlite-1","State":"running","Health":""}
{"Name":"compose-postgres-1","State":"running","Health":"healthy"}
{"Name":"compose-healthcheck-secrets-1","State":"exited","ExitCode":0}`

	serviceMap, err := parseDockerComposePsOutput([]byte(input))
	require.NoError(t, err)
	require.Len(t, serviceMap, 3)

	// Check cryptoutil-sqlite
	sqliteService, exists := serviceMap[cryptoutilSharedMagic.DockerServiceCryptoutilSqlite]
	require.True(t, exists)
	require.Equal(t, "compose-cryptoutil-sqlite-1", sqliteService["Name"])
	require.Equal(t, cryptoutilSharedMagic.DockerServiceStateRunning, sqliteService["State"])

	// Check postgres
	postgresService, exists := serviceMap[cryptoutilSharedMagic.DockerServicePostgres]
	require.True(t, exists)
	require.Equal(t, "compose-postgres-1", postgresService["Name"])
	require.Equal(t, cryptoutilSharedMagic.DockerServiceHealthHealthy, postgresService["Health"])

	// Check healthcheck job
	healthcheckJob, exists := serviceMap[cryptoutilSharedMagic.DockerJobHealthcheckSecrets]
	require.True(t, exists)
	require.Equal(t, cryptoutilSharedMagic.DockerServiceStateExited, healthcheckJob["State"])
}

func TestParseDockerComposePsOutput_EmptyInput(t *testing.T) {
	t.Parallel()

	input := ""

	_, err := parseDockerComposePsOutput([]byte(input))
	require.Error(t, err)
	require.Contains(t, err.Error(), "no services found")
}

func TestParseDockerComposePsOutput_InvalidJSON(t *testing.T) {
	t.Parallel()

	input := `{"Name":"compose-cryptoutil-sqlite-1","State":"running"
invalid json line`

	_, err := parseDockerComposePsOutput([]byte(input))
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse Docker service JSON")
}

func TestDetermineServiceHealthStatus_JobOnly(t *testing.T) {
	t.Parallel()

	// Use case 1: Job-only (standalone job that must exit successfully)
	serviceMap := map[string]map[string]any{
		cryptoutilSharedMagic.DockerJobHealthcheckSecrets: {
			"Name":     "compose-healthcheck-secrets-1",
			"State":    cryptoutilSharedMagic.DockerServiceStateExited,
			"ExitCode": float64(0),
		},
		cryptoutilSharedMagic.DockerJobBuilderCryptoutil: {
			"Name":     "compose-builder-cryptoutil-1",
			"State":    cryptoutilSharedMagic.DockerServiceStateExited,
			"ExitCode": float64(1), // Failed job
		},
	}

	services := []ServiceAndJob{
		{Service: "", Job: cryptoutilSharedMagic.DockerJobHealthcheckSecrets},
		{Service: "", Job: cryptoutilSharedMagic.DockerJobBuilderCryptoutil},
	}

	healthStatus := determineServiceHealthStatus(serviceMap, services)

	require.True(t, healthStatus[cryptoutilSharedMagic.DockerJobHealthcheckSecrets], "healthcheck-secrets should be healthy (ExitCode=0)")
	require.False(t, healthStatus[cryptoutilSharedMagic.DockerJobBuilderCryptoutil], "builder-cryptoutil should be unhealthy (ExitCode=1)")
}

func TestDetermineServiceHealthStatus_ServiceOnly(t *testing.T) {
	t.Parallel()

	// Use case 2: Service-only (service with native healthcheck)
	serviceMap := map[string]map[string]any{
		cryptoutilSharedMagic.DockerServiceCryptoutilSqlite: {
			"Name":   "compose-cryptoutil-sqlite-1",
			"State":  cryptoutilSharedMagic.DockerServiceStateRunning,
			"Health": cryptoutilSharedMagic.DockerServiceHealthHealthy,
		},
		cryptoutilSharedMagic.DockerServicePostgres: {
			"Name":   "compose-postgres-1",
			"State":  cryptoutilSharedMagic.DockerServiceStateRunning,
			"Health": "unhealthy",
		},
		cryptoutilSharedMagic.DockerServiceCryptoutilPostgres1: {
			"Name":  "compose-cryptoutil-postgres-1-1",
			"State": cryptoutilSharedMagic.DockerServiceStateRunning,
			// No Health field - should check State only
		},
	}

	services := []ServiceAndJob{
		{Service: cryptoutilSharedMagic.DockerServiceCryptoutilSqlite, Job: ""},
		{Service: cryptoutilSharedMagic.DockerServicePostgres, Job: ""},
		{Service: cryptoutilSharedMagic.DockerServiceCryptoutilPostgres1, Job: ""},
	}

	healthStatus := determineServiceHealthStatus(serviceMap, services)

	require.True(t, healthStatus[cryptoutilSharedMagic.DockerServiceCryptoutilSqlite], "cryptoutil-sqlite should be healthy")
	require.False(t, healthStatus[cryptoutilSharedMagic.DockerServicePostgres], "postgres should be unhealthy")
	require.True(t, healthStatus[cryptoutilSharedMagic.DockerServiceCryptoutilPostgres1], "cryptoutil-postgres-1 should be healthy (running without health field)")
}

func TestDetermineServiceHealthStatus_ServiceWithJob(t *testing.T) {
	t.Parallel()

	// Use case 3: Service with healthcheck job (external job verifies service)
	serviceMap := map[string]map[string]any{
		cryptoutilSharedMagic.DockerServiceOtelCollector: {
			"Name":  "compose-opentelemetry-collector-contrib-1",
			"State": cryptoutilSharedMagic.DockerServiceStateRunning,
			// No native healthcheck
		},
		cryptoutilSharedMagic.DockerJobHealthcheckOtelCollectorContrib: {
			"Name":     "compose-healthcheck-opentelemetry-collector-contrib-1",
			"State":    cryptoutilSharedMagic.DockerServiceStateExited,
			"ExitCode": float64(0),
		},
	}

	services := []ServiceAndJob{
		{Service: cryptoutilSharedMagic.DockerServiceOtelCollector, Job: cryptoutilSharedMagic.DockerJobHealthcheckOtelCollectorContrib},
	}

	healthStatus := determineServiceHealthStatus(serviceMap, services)

	// The health status key should be the job name (job takes precedence)
	require.True(t, healthStatus[cryptoutilSharedMagic.DockerJobHealthcheckOtelCollectorContrib], "otel-collector should be healthy via job")
}

func TestDetermineServiceHealthStatus_ServiceNotFound(t *testing.T) {
	t.Parallel()

	serviceMap := map[string]map[string]any{
		cryptoutilSharedMagic.DockerServiceCryptoutilSqlite: {
			"Name":   "compose-cryptoutil-sqlite-1",
			"State":  cryptoutilSharedMagic.DockerServiceStateRunning,
			"Health": cryptoutilSharedMagic.DockerServiceHealthHealthy,
		},
	}

	services := []ServiceAndJob{
		{Service: "nonexistent-service", Job: ""},
	}

	healthStatus := determineServiceHealthStatus(serviceMap, services)

	require.False(t, healthStatus["nonexistent-service"], "nonexistent service should be unhealthy")
}

func TestDetermineServiceHealthStatus_MixedUseCases(t *testing.T) {
	t.Parallel()

	// Test all three use cases together
	serviceMap := map[string]map[string]any{
		cryptoutilSharedMagic.DockerJobHealthcheckSecrets: {
			"Name":     "compose-healthcheck-secrets-1",
			"State":    cryptoutilSharedMagic.DockerServiceStateExited,
			"ExitCode": float64(0),
		},
		cryptoutilSharedMagic.DockerServiceCryptoutilSqlite: {
			"Name":   "compose-cryptoutil-sqlite-1",
			"State":  cryptoutilSharedMagic.DockerServiceStateRunning,
			"Health": cryptoutilSharedMagic.DockerServiceHealthHealthy,
		},
		cryptoutilSharedMagic.DockerServiceOtelCollector: {
			"Name":  "compose-opentelemetry-collector-contrib-1",
			"State": cryptoutilSharedMagic.DockerServiceStateRunning,
		},
		cryptoutilSharedMagic.DockerJobHealthcheckOtelCollectorContrib: {
			"Name":     "compose-healthcheck-opentelemetry-collector-contrib-1",
			"State":    cryptoutilSharedMagic.DockerServiceStateExited,
			"ExitCode": float64(0),
		},
	}

	services := []ServiceAndJob{
		{Service: "", Job: cryptoutilSharedMagic.DockerJobHealthcheckSecrets},                                                            // Use case 1
		{Service: cryptoutilSharedMagic.DockerServiceCryptoutilSqlite, Job: ""},                                                          // Use case 2
		{Service: cryptoutilSharedMagic.DockerServiceOtelCollector, Job: cryptoutilSharedMagic.DockerJobHealthcheckOtelCollectorContrib}, // Use case 3
	}

	healthStatus := determineServiceHealthStatus(serviceMap, services)

	require.True(t, healthStatus[cryptoutilSharedMagic.DockerJobHealthcheckSecrets], "standalone job should be healthy")
	require.True(t, healthStatus[cryptoutilSharedMagic.DockerServiceCryptoutilSqlite], "service with native healthcheck should be healthy")
	require.True(t, healthStatus[cryptoutilSharedMagic.DockerJobHealthcheckOtelCollectorContrib], "service with external job should be healthy")
}
