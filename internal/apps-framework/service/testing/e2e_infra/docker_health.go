// Copyright (c) 2025-2026 Justin Cranford.
// Package e2e_infra provides reusable helpers for E2E testing with docker compose.
package e2e_infra

import (
	"context"
	json "encoding/json"
	"fmt"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ServiceAndJob represents a service and its optional healthcheck job.
//
// Three use cases are supported:
//  1. Job-only (Service="", Job="job-name"): Standalone job that must exit successfully (ExitCode=0)
//     Examples: healthcheck-secrets, builder-sm-kms
//  2. Service-only (Service="service-name", Job=""): Service that must be running/healthy with native health checks
//     Examples: cryptoutil-sqlite, cryptoutil-postgres-1, cryptoutil-postgres-2, postgres, grafana-otel-lgtm
//  3. Service with healthcheck job (Service="service-name", Job="job-name"): Service uses external job for health verification
//     Example: opentelemetry-collector-contrib with healthcheck-opentelemetry-collector-contrib
type ServiceAndJob struct {
	Service string // The service name to check (empty for standalone jobs)
	Job     string // Optional healthcheck job name (empty if service has native health checks)
}

// parseDockerComposePsOutput parses docker compose ps --format json output into a service name to service data map.
func parseDockerComposePsOutput(output []byte) (map[string]map[string]any, error) {
	trimmedOutput := strings.TrimSpace(string(output))
	if trimmedOutput == "" {
		return nil, fmt.Errorf("no services found in Docker compose output")
	}

	var (
		parsed      any
		serviceList []map[string]any
	)

	if err := json.Unmarshal([]byte(trimmedOutput), &parsed); err == nil {
		switch containers := parsed.(type) {
		case []any:
			for _, container := range containers {
				service, ok := container.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("docker compose JSON contains a malformed container entry")
				}

				serviceList = append(serviceList, service)
			}
		case map[string]any:
			serviceList = append(serviceList, containers)
		default:
			return nil, fmt.Errorf("docker compose JSON must contain a container object or array")
		}
	} else {
		for _, line := range strings.Split(trimmedOutput, "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}

			var service map[string]any
			if err := json.Unmarshal([]byte(line), &service); err != nil {
				return nil, fmt.Errorf("failed to parse Docker service JSON: %w", err)
			}

			if service == nil {
				return nil, fmt.Errorf("docker compose JSON contains a malformed container entry")
			}

			serviceList = append(serviceList, service)
		}
	}

	if len(serviceList) == 0 {
		return nil, fmt.Errorf("no services found in Docker compose output")
	}

	serviceMap := make(map[string]map[string]any)

	for _, service := range serviceList {
		if serviceName, ok := service[valueService].(string); ok && serviceName != "" {
			serviceMap[serviceName] = service

			continue
		}

		if name, ok := service[valueName].(string); ok && strings.HasPrefix(name, "compose-") {
			parts := strings.Split(name, "-")
			if len(parts) >= cryptoutilSharedMagic.DockerServiceNamePartsMin {
				serviceName := strings.Join(parts[1:len(parts)-1], "-")
				serviceMap[serviceName] = service
			}
		}
	}

	if len(serviceMap) == 0 {
		return nil, fmt.Errorf("no services found in Docker compose output")
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

		if service.Job != "" {
			serviceNameToCheck = service.Job
			healthKey = service.Job
		} else {
			serviceNameToCheck = service.Service
			healthKey = service.Service
		}

		serviceData, exists := serviceMap[serviceNameToCheck]
		if !exists {
			healthStatus[healthKey] = false

			continue
		}

		// Check health based on service type
		if service.Job != "" {
			// Use case 1 & 3: Job is present - Check job's exit code
			// (Covers both "Job-only" and "Service with healthcheck job")
			exitCode, hasExitCode := serviceData[valueExitCode].(float64)
			healthStatus[healthKey] = hasExitCode && exitCode == 0
		} else {
			// Use case 2: Service-only - Check service's running/healthy status
			state, hasState := serviceData[valueState].(string)
			health, hasHealth := serviceData[valueHealth].(string)

			var isHealthy bool
			if hasHealth {
				// If Health field present, check if "healthy"
				isHealthy = health == cryptoutilSharedMagic.DockerServiceHealthHealthy
			} else {
				// If no Health field, check if State is "running"
				isHealthy = hasState && state == cryptoutilSharedMagic.DockerServiceStateRunning
			}

			healthStatus[healthKey] = isHealthy
		}
	}

	return healthStatus
}

// WaitForServicesHealthy waits for all specified services to become healthy.
// Supports all three healthcheck use cases (job-only, service-only, service-with-job).
// Returns error if any service fails to become healthy within the polling mechanism.
func (cm *ComposeManager) WaitForServicesHealthy(ctx context.Context, services []ServiceAndJob) error {
	fmt.Println("[WaitForServicesHealthy] Starting batch health check...")

	// Run docker compose ps to get service status.
	output, err := cm.psOutputFn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get docker compose ps output: %w (output: %s)", err, string(output))
	}

	// Parse service data
	serviceMap, err := parseDockerComposePsOutput(output)
	if err != nil {
		return fmt.Errorf("failed to parse docker compose ps output: %w", err)
	}

	// Determine health status for all services
	healthStatus := determineServiceHealthStatus(serviceMap, services)

	// Check if all services are healthy
	unhealthyServices := make([]string, 0)

	for serviceName, isHealthy := range healthStatus {
		if !isHealthy {
			unhealthyServices = append(unhealthyServices, serviceName)
		}
	}

	if len(unhealthyServices) > 0 {
		return fmt.Errorf("unhealthy services: %s", strings.Join(unhealthyServices, ", "))
	}

	fmt.Println("[WaitForServicesHealthy] All services are healthy")

	return nil
}
