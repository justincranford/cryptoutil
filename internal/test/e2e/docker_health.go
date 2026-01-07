// Copyright (c) 2025 Justin Cranford

//go:build e2e

package test

import (
	"encoding/json"
	"fmt"
	"strings"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
)

// ServiceAndJob represents a service and its optional healthcheck job.
//
// Three use cases are supported:
//  1. Job-only (Service="", Job="job-name"): Standalone job that must exit successfully (ExitCode=0)
//  2. Service-only (Service="service-name", Job=""): Service that must be running/healthy
//  3. Service with healthcheck job (Service="service-name", Job="job-name"): Service uses external job for health verification
type ServiceAndJob struct {
	Service string // The service name to check (empty for standalone jobs)
	Job     string // Optional healthcheck job name (empty if service has native health checks)
}

// Docker compose service names for batch health checking.
//
// Three use cases are supported:
//  1. Job-only (Service="", Job="job-name"): Standalone job that must exit successfully (ExitCode=0)
//     Examples: healthcheck-secrets, builder-cryptoutil
//  2. Service-only (Service="service-name", Job=""): Service that must be running/healthy with native health checks
//     Examples: cryptoutil-sqlite, cryptoutil-postgres-1, cryptoutil-postgres-2, postgres, grafana-otel-lgtm
//  3. Service with healthcheck job (Service="service-name", Job="job-name"): Service uses external job for health verification
//     Example: opentelemetry-collector-contrib with healthcheck-opentelemetry-collector-contrib
var (
	dockerComposeServicesForHealthCheck = []ServiceAndJob{
		{Job: cryptoutilMagic.DockerJobHealthcheckSecrets},
		{Job: cryptoutilMagic.DockerJobBuilderCryptoutil},
		{Service: cryptoutilMagic.DockerServiceOtelCollector, Job: cryptoutilMagic.DockerJobHealthcheckOtelCollectorContrib},
		{Service: cryptoutilMagic.DockerServiceCryptoutilSqlite},
		{Service: cryptoutilMagic.DockerServiceCryptoutilPostgres1},
		{Service: cryptoutilMagic.DockerServiceCryptoutilPostgres2},
		{Service: cryptoutilMagic.DockerServicePostgres},
		// grafana-otel-lgtm excluded - requires --profile with-grafana (optional for E2E tests)
	}
)

// parseDockerComposePsOutput parses docker compose ps --format json output into a service name to service data map.
func parseDockerComposePsOutput(output []byte) (map[string]map[string]any, error) {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	serviceList := make([]map[string]any, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var service map[string]any
		if err := json.Unmarshal([]byte(line), &service); err != nil {
			return nil, fmt.Errorf("failed to parse Docker service JSON: %w", err)
		}

		serviceList = append(serviceList, service)
	}

	if len(serviceList) == 0 {
		return nil, fmt.Errorf("no services found in Docker compose output")
	}

	serviceMap := make(map[string]map[string]any)

	for _, service := range serviceList {
		if name, ok := service["Name"].(string); ok {
			if strings.Contains(name, "compose-") { // Example: compose-cryptoutil-sqlite-1
				parts := strings.Split(name, "-")
				if len(parts) >= cryptoutilMagic.DockerServiceNamePartsMin {
					serviceName := strings.Join(parts[1:len(parts)-1], "-") // Example: cryptoutil-sqlite
					serviceMap[serviceName] = service
				}
			}
		}
	}

	return serviceMap, nil
}

// determineServiceHealthStatus determines the health status of services from a service map.
// This function handles all three use cases:
//  1. Job-only (Service="", Job="job-name"): Checks if standalone job exited successfully (ExitCode=0)
//  2. Service-only (Service="service-name", Job=""): Checks if service is running/healthy
//  3. Service with healthcheck job (Service="service-name", Job="job-name"): Uses job to verify service health
func determineServiceHealthStatus(serviceMap map[string]map[string]any, services []ServiceAndJob) map[string]bool {
	healthStatus := make(map[string]bool)

	for _, service := range services {
		// Determine which name to check (job takes precedence if present)
		var serviceNameToCheck string

		var healthKey string

		if service.Service == "" && service.Job != "" {
			// Use case 1: Job-only (standalone job)
			serviceNameToCheck = service.Job
			healthKey = service.Job
		} else if service.Service != "" && service.Job == "" {
			// Use case 2: Service-only (native health checks)
			serviceNameToCheck = service.Service
			healthKey = service.Service
		} else if service.Service != "" && service.Job != "" {
			// Use case 3: Service with healthcheck job
			serviceNameToCheck = service.Job
			healthKey = service.Service
		} else {
			// Invalid configuration: both Service and Job are empty
			continue
		}

		serviceData, exists := serviceMap[serviceNameToCheck]
		if !exists {
			// Service/job not found
			if service.Job != "" && service.Service == "" {
				// Use case 1: Job-only not found means it completed and was cleaned up (healthy)
				healthStatus[healthKey] = true
			} else if service.Job != "" && service.Service != "" {
				// Use case 3: Healthcheck job not found means it completed successfully (healthy)
				healthStatus[healthKey] = true
			} else {
				// Use case 2: Service not found means unhealthy
				healthStatus[healthKey] = false
			}

			continue
		}

		// Check health based on the use case
		if service.Job != "" {
			// Use case 1 or 3: Check if job exited successfully
			if state, ok := serviceData["State"].(string); ok && state == cryptoutilMagic.DockerServiceStateExited {
				var exitCode int
				if exitCodeFloat, ok := serviceData["ExitCode"].(float64); ok {
					exitCode = int(exitCodeFloat)
				} else if exitCodeInt, ok := serviceData["ExitCode"].(int); ok {
					exitCode = exitCodeInt
				} else {
					healthStatus[healthKey] = false // ExitCode field not found or wrong type

					continue
				}

				healthStatus[healthKey] = exitCode == 0
			} else {
				healthStatus[healthKey] = false // Not in exited state
			}
		} else {
			// Use case 2: Regular service - check health or running state
			if health, ok := serviceData["Health"].(string); ok && health != "" {
				// Services with native health checks
				healthStatus[healthKey] = health == cryptoutilMagic.DockerServiceHealthHealthy
			} else {
				// Services without health checks: check if running
				if state, ok := serviceData["State"].(string); ok && state == cryptoutilMagic.DockerServiceStateRunning {
					healthStatus[healthKey] = true
				} else {
					healthStatus[healthKey] = false
				}
			}
		}
	}

	return healthStatus
}
