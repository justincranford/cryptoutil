// Copyright (c) 2025-2026 Justin Cranford.
package e2e_infra

import (
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const testNoServicesFound = "no services found"

func TestParseDockerComposePsOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		wantServices   []string
		wantErrContain string
	}{
		{
			name: "newline-delimited JSON with name fallback",
			input: `{"Name":"compose-cryptoutil-sqlite-1","State":"running","Health":""}
{"Name":"compose-postgres-1","State":"running","Health":"healthy"}
{"Name":"compose-healthcheck-secrets-1","State":"exited","ExitCode":0}`,
			wantServices: []string{
				cryptoutilSharedMagic.DockerServiceCryptoutilSqlite,
				cryptoutilSharedMagic.DockerServicePostgres,
				cryptoutilSharedMagic.DockerJobHealthcheckSecrets,
			},
		},
		{
			name:  "JSON array with service names",
			input: `[{"Service":"cryptoutil-sqlite","State":"running"},{"Service":"postgres","State":"running","Health":"healthy"}]`,
			wantServices: []string{
				cryptoutilSharedMagic.DockerServiceCryptoutilSqlite,
				cryptoutilSharedMagic.DockerServicePostgres,
			},
		},
		{
			name:         "single JSON object with service name",
			input:        `{"Service":"healthcheck-secrets","State":"exited","ExitCode":0}`,
			wantServices: []string{cryptoutilSharedMagic.DockerJobHealthcheckSecrets},
		},
		{
			name:           "empty input",
			input:          "",
			wantErrContain: testNoServicesFound,
		},
		{
			name:           "empty JSON array",
			input:          `[]`,
			wantErrContain: testNoServicesFound,
		},
		{
			name: "malformed newline-delimited JSON",
			input: `{"Name":"compose-cryptoutil-sqlite-1","State":"running"
invalid json line`,
			wantErrContain: "failed to parse Docker service JSON",
		},
		{
			name:           "newline-delimited null entry",
			input:          "null\n{\"Service\":\"cryptoutil-sqlite\"}",
			wantErrContain: testMalformedContainerEntry,
		},
		{
			name:         "newline-delimited output skips blank lines",
			input:        "{\"Service\":\"cryptoutil-sqlite\"}\n\n{\"Service\":\"postgres\"}",
			wantServices: []string{cryptoutilSharedMagic.DockerServiceCryptoutilSqlite, cryptoutilSharedMagic.DockerServicePostgres},
		},
		{
			name:           "malformed array entry",
			input:          `[null]`,
			wantErrContain: "malformed container entry",
		},
		{
			name:           "unsupported top-level JSON value",
			input:          `true`,
			wantErrContain: "must contain a container object or array",
		},
		{
			name:           "object without a recognizable service",
			input:          `{"State":"running"}`,
			wantErrContain: testNoServicesFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			serviceMap, err := parseDockerComposePsOutput([]byte(tc.input))
			if tc.wantErrContain != "" {
				require.ErrorContains(t, err, tc.wantErrContain)

				return
			}

			require.NoError(t, err)
			require.Len(t, serviceMap, len(tc.wantServices))

			for _, serviceName := range tc.wantServices {
				require.Contains(t, serviceMap, serviceName)
			}
		})
	}
}

func TestDetermineServiceHealthStatus_JobOnly(t *testing.T) {
	t.Parallel()

	builderJobName := "builder-" + cryptoutilSharedMagic.OTLPServiceSMKMS

	// Use case 1: Job-only (standalone job that must exit successfully)
	serviceMap := map[string]map[string]any{
		cryptoutilSharedMagic.DockerJobHealthcheckSecrets: {
			valueName:     testComposeHealthcheckSecrets1,
			valueState:    cryptoutilSharedMagic.DockerServiceStateExited,
			valueExitCode: float64(0),
		},
		builderJobName: {
			valueName:     "compose-builder-sm-kms-1",
			valueState:    cryptoutilSharedMagic.DockerServiceStateExited,
			valueExitCode: float64(1), // Failed job
		},
	}

	services := []ServiceAndJob{
		{Service: "", Job: cryptoutilSharedMagic.DockerJobHealthcheckSecrets},
		{Service: "", Job: builderJobName},
	}

	healthStatus := determineServiceHealthStatus(serviceMap, services)

	require.True(t, healthStatus[cryptoutilSharedMagic.DockerJobHealthcheckSecrets], "healthcheck-secrets should be healthy (ExitCode=0)")
	require.False(t, healthStatus[builderJobName], "builder-sm-kms should be unhealthy (ExitCode=1)")
}

func TestDetermineServiceHealthStatus_ServiceOnly(t *testing.T) {
	t.Parallel()

	// Use case 2: Service-only (service with native healthcheck)
	serviceMap := map[string]map[string]any{
		cryptoutilSharedMagic.DockerServiceCryptoutilSqlite: {
			valueName:   testComposeCryptoutilSqlite1,
			valueState:  cryptoutilSharedMagic.DockerServiceStateRunning,
			valueHealth: cryptoutilSharedMagic.DockerServiceHealthHealthy,
		},
		cryptoutilSharedMagic.DockerServicePostgres: {
			valueName:   "compose-postgres-1",
			valueState:  cryptoutilSharedMagic.DockerServiceStateRunning,
			valueHealth: "unhealthy",
		},
		cryptoutilSharedMagic.DockerServiceCryptoutilPostgres1: {
			valueName:  "compose-cryptoutil-postgres-1-1",
			valueState: cryptoutilSharedMagic.DockerServiceStateRunning,
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
			valueName:  testComposeOpentelemetryCollectorContrib1,
			valueState: cryptoutilSharedMagic.DockerServiceStateRunning,
			// No native healthcheck
		},
		cryptoutilSharedMagic.DockerJobHealthcheckOtelCollectorContrib: {
			valueName:     testComposeHealthcheckOpentelemetryCollectorContrib1,
			valueState:    cryptoutilSharedMagic.DockerServiceStateExited,
			valueExitCode: float64(0),
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
			valueName:   testComposeCryptoutilSqlite1,
			valueState:  cryptoutilSharedMagic.DockerServiceStateRunning,
			valueHealth: cryptoutilSharedMagic.DockerServiceHealthHealthy,
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
			valueName:     testComposeHealthcheckSecrets1,
			valueState:    cryptoutilSharedMagic.DockerServiceStateExited,
			valueExitCode: float64(0),
		},
		cryptoutilSharedMagic.DockerServiceCryptoutilSqlite: {
			valueName:   testComposeCryptoutilSqlite1,
			valueState:  cryptoutilSharedMagic.DockerServiceStateRunning,
			valueHealth: cryptoutilSharedMagic.DockerServiceHealthHealthy,
		},
		cryptoutilSharedMagic.DockerServiceOtelCollector: {
			valueName:  testComposeOpentelemetryCollectorContrib1,
			valueState: cryptoutilSharedMagic.DockerServiceStateRunning,
		},
		cryptoutilSharedMagic.DockerJobHealthcheckOtelCollectorContrib: {
			valueName:     testComposeHealthcheckOpentelemetryCollectorContrib1,
			valueState:    cryptoutilSharedMagic.DockerServiceStateExited,
			valueExitCode: float64(0),
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
