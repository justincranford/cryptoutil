//go:build e2e

package test

import (
	"encoding/json"
	"fmt"
	"strings"

	cryptoutilMagic "cryptoutil/internal/common/magic"
)

// ServiceNameAndJob represents a service and its optional healthcheck job.
type ServiceNameAndJob struct {
	Name           string // The service name to check
	HealthcheckJob string // Optional healthcheck job name (empty if no healthcheck job)
}

// Docker compose service names for batch health checking.
var (
	dockerComposeServicesForHealthCheck = []ServiceNameAndJob{
		{Name: cryptoutilMagic.DockerServiceCryptoutilSqlite},
		{Name: cryptoutilMagic.DockerServiceCryptoutilPostgres1},
		{Name: cryptoutilMagic.DockerServiceCryptoutilPostgres2},
		{Name: cryptoutilMagic.DockerServicePostgres},
		{Name: cryptoutilMagic.DockerServiceGrafanaOtelLgtm},
		{Name: cryptoutilMagic.DockerServiceOtelCollector, HealthcheckJob: cryptoutilMagic.DockerJobOtelCollectorHealthcheck},
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
			if strings.Contains(name, "compose-") {
				parts := strings.Split(name, "-")
				if len(parts) >= cryptoutilMagic.DockerServiceNamePartsMin {
					serviceName := strings.Join(parts[1:len(parts)-1], "-")
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
func determineServiceHealthStatus(serviceMap map[string]map[string]any, services []ServiceNameAndJob) map[string]bool {
	healthStatus := make(map[string]bool)

	for _, service := range services {
		var serviceNameToCheck string
		if service.HealthcheckJob != "" {
			// Use the healthcheck job to determine service health
			serviceNameToCheck = service.HealthcheckJob
		} else {
			// Check the service directly
			serviceNameToCheck = service.Name
		}

		serviceData, exists := serviceMap[serviceNameToCheck]
		if !exists {
			// Service/job not found
			if service.HealthcheckJob != "" {
				// For healthcheck jobs: not found means it completed successfully and was cleaned up
				healthStatus[service.Name] = true
			} else {
				// For regular services: not found means unhealthy
				healthStatus[service.Name] = false
			}

			continue
		}

		if service.HealthcheckJob != "" {
			// This is a healthcheck job - check if it exited successfully
			if state, ok := serviceData["State"].(string); ok && state == cryptoutilMagic.DockerServiceStateExited {
				// Handle both int and float64 types for ExitCode
				var exitCode int
				if exitCodeFloat, ok := serviceData["ExitCode"].(float64); ok {
					exitCode = int(exitCodeFloat)
				} else if exitCodeInt, ok := serviceData["ExitCode"].(int); ok {
					exitCode = exitCodeInt
				} else {
					// ExitCode field not found or wrong type
					healthStatus[service.Name] = false

					continue
				}

				healthStatus[service.Name] = exitCode == 0
			} else {
				// Not in exited state
				healthStatus[service.Name] = false
			}
		} else {
			// Regular service - check health or running state
			if health, ok := serviceData["Health"].(string); ok && health != "" {
				// Services with health checks
				healthStatus[service.Name] = health == cryptoutilMagic.DockerServiceHealthHealthy
			} else {
				// Services without health checks: check if running
				if state, ok := serviceData["State"].(string); ok && state == cryptoutilMagic.DockerServiceStateRunning {
					healthStatus[service.Name] = true
				} else {
					healthStatus[service.Name] = false
				}
			}
		}
	}

	return healthStatus
}
