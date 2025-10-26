//go:build e2e

package test

import (
	"encoding/json"
	"fmt"
	"strings"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// ServiceAndJob represents a service and its optional healthcheck job.
type ServiceAndJob struct {
	Service string // The service name to check
	Job     string // Optional healthcheck job name (empty if no healthcheck job)
}

// Docker compose service names for batch health checking.
var (
	dockerComposeServicesForHealthCheck = []ServiceAndJob{
		{Service: cryptoutilMagic.DockerServiceCryptoutilSqlite},
		{Service: cryptoutilMagic.DockerServiceCryptoutilPostgres1},
		{Service: cryptoutilMagic.DockerServiceCryptoutilPostgres2},
		{Service: cryptoutilMagic.DockerServicePostgres},
		{Service: cryptoutilMagic.DockerServiceGrafanaOtelLgtm},
		{Service: cryptoutilMagic.DockerServiceOtelCollector, Job: cryptoutilMagic.DockerJobOtelCollectorHealthcheck},
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
// This function contains the shared logic for determining if services are healthy.
// It handles services with and without healthcheck jobs.
func determineServiceHealthStatus(serviceMap map[string]map[string]any, services []ServiceAndJob) map[string]bool {
	healthStatus := make(map[string]bool)

	for _, service := range services {
		var serviceNameToCheck string
		if service.Job == "" {
			serviceNameToCheck = service.Service // Check the service directly
		} else {
			serviceNameToCheck = service.Job // Use the healthcheck job to determine service health
		}

		serviceData, exists := serviceMap[serviceNameToCheck]
		if !exists { // Service/job not found
			if service.Job == "" { // For regular services: not found means unhealthy
				healthStatus[service.Service] = false
			} else {
				healthStatus[service.Service] = true // For healthcheck jobs: not found means it completed successfully and was cleaned up
			}

			continue
		}

		if service.Job == "" {
			// Regular service - check health or running state
			if health, ok := serviceData["Health"].(string); ok && health != "" {
				// Services with health checks
				healthStatus[service.Service] = health == cryptoutilMagic.DockerServiceHealthHealthy
			} else {
				// Services without health checks: check if running
				if state, ok := serviceData["State"].(string); ok && state == cryptoutilMagic.DockerServiceStateRunning {
					healthStatus[service.Service] = true
				} else {
					healthStatus[service.Service] = false
				}
			}
		} else {
			// This is a healthcheck job - check if it exited successfully
			if state, ok := serviceData["State"].(string); ok && state == cryptoutilMagic.DockerServiceStateExited {
				var exitCode int
				if exitCodeFloat, ok := serviceData["ExitCode"].(float64); ok { // Handle float64 ExitCode
					exitCode = int(exitCodeFloat)
				} else if exitCodeInt, ok := serviceData["ExitCode"].(int); ok { // Handle int ExitCode
					exitCode = exitCodeInt
				} else {
					healthStatus[service.Service] = false // ExitCode field not found or wrong type

					continue
				}

				healthStatus[service.Service] = exitCode == 0
			} else {
				healthStatus[service.Service] = false // Not in exited state
			}
		}
	}

	return healthStatus
}
