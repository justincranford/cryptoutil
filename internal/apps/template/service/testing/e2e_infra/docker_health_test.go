// Copyright (c) 2025 Justin Cranford

package e2e_infra

import (
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
sqliteService, exists := serviceMap["cryptoutil-sqlite"]
require.True(t, exists)
require.Equal(t, "compose-cryptoutil-sqlite-1", sqliteService["Name"])
require.Equal(t, "running", sqliteService["State"])

// Check postgres
postgresService, exists := serviceMap["postgres"]
require.True(t, exists)
require.Equal(t, "compose-postgres-1", postgresService["Name"])
require.Equal(t, "healthy", postgresService["Health"])

// Check healthcheck job
healthcheckJob, exists := serviceMap["healthcheck-secrets"]
require.True(t, exists)
require.Equal(t, "exited", healthcheckJob["State"])
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
"healthcheck-secrets": {
"Name":     "compose-healthcheck-secrets-1",
"State":    "exited",
"ExitCode": float64(0),
},
"builder-cryptoutil": {
"Name":     "compose-builder-cryptoutil-1",
"State":    "exited",
"ExitCode": float64(1), // Failed job
},
}

services := []ServiceAndJob{
{Service: "", Job: "healthcheck-secrets"},
{Service: "", Job: "builder-cryptoutil"},
}

healthStatus := determineServiceHealthStatus(serviceMap, services)

require.True(t, healthStatus["healthcheck-secrets"], "healthcheck-secrets should be healthy (ExitCode=0)")
require.False(t, healthStatus["builder-cryptoutil"], "builder-cryptoutil should be unhealthy (ExitCode=1)")
}

func TestDetermineServiceHealthStatus_ServiceOnly(t *testing.T) {
t.Parallel()

// Use case 2: Service-only (service with native healthcheck)
serviceMap := map[string]map[string]any{
"cryptoutil-sqlite": {
"Name":   "compose-cryptoutil-sqlite-1",
"State":  "running",
"Health": "healthy",
},
"postgres": {
"Name":   "compose-postgres-1",
"State":  "running",
"Health": "unhealthy",
},
"cryptoutil-postgres-1": {
"Name":  "compose-cryptoutil-postgres-1-1",
"State": "running",
// No Health field - should check State only
},
}

services := []ServiceAndJob{
{Service: "cryptoutil-sqlite", Job: ""},
{Service: "postgres", Job: ""},
{Service: "cryptoutil-postgres-1", Job: ""},
}

healthStatus := determineServiceHealthStatus(serviceMap, services)

require.True(t, healthStatus["cryptoutil-sqlite"], "cryptoutil-sqlite should be healthy")
require.False(t, healthStatus["postgres"], "postgres should be unhealthy")
require.True(t, healthStatus["cryptoutil-postgres-1"], "cryptoutil-postgres-1 should be healthy (running without health field)")
}

func TestDetermineServiceHealthStatus_ServiceWithJob(t *testing.T) {
t.Parallel()

// Use case 3: Service with healthcheck job (external job verifies service)
serviceMap := map[string]map[string]any{
"opentelemetry-collector-contrib": {
"Name":  "compose-opentelemetry-collector-contrib-1",
"State": "running",
// No native healthcheck
},
"healthcheck-opentelemetry-collector-contrib": {
"Name":     "compose-healthcheck-opentelemetry-collector-contrib-1",
"State":    "exited",
"ExitCode": float64(0),
},
}

services := []ServiceAndJob{
{Service: "opentelemetry-collector-contrib", Job: "healthcheck-opentelemetry-collector-contrib"},
}

healthStatus := determineServiceHealthStatus(serviceMap, services)

// The health status key should be the job name (job takes precedence)
require.True(t, healthStatus["healthcheck-opentelemetry-collector-contrib"], "otel-collector should be healthy via job")
}

func TestDetermineServiceHealthStatus_ServiceNotFound(t *testing.T) {
t.Parallel()

serviceMap := map[string]map[string]any{
"cryptoutil-sqlite": {
"Name":   "compose-cryptoutil-sqlite-1",
"State":  "running",
"Health": "healthy",
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
"healthcheck-secrets": {
"Name":     "compose-healthcheck-secrets-1",
"State":    "exited",
"ExitCode": float64(0),
},
"cryptoutil-sqlite": {
"Name":   "compose-cryptoutil-sqlite-1",
"State":  "running",
"Health": "healthy",
},
"opentelemetry-collector-contrib": {
"Name":  "compose-opentelemetry-collector-contrib-1",
"State": "running",
},
"healthcheck-opentelemetry-collector-contrib": {
"Name":     "compose-healthcheck-opentelemetry-collector-contrib-1",
"State":    "exited",
"ExitCode": float64(0),
},
}

services := []ServiceAndJob{
{Service: "", Job: "healthcheck-secrets"},                                                          // Use case 1
{Service: "cryptoutil-sqlite", Job: ""},                                                            // Use case 2
{Service: "opentelemetry-collector-contrib", Job: "healthcheck-opentelemetry-collector-contrib"}, // Use case 3
}

healthStatus := determineServiceHealthStatus(serviceMap, services)

require.True(t, healthStatus["healthcheck-secrets"], "standalone job should be healthy")
require.True(t, healthStatus["cryptoutil-sqlite"], "service with native healthcheck should be healthy")
require.True(t, healthStatus["healthcheck-opentelemetry-collector-contrib"], "service with external job should be healthy")
}
