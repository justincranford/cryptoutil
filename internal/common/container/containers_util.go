package container

import (
	"context"
	"fmt"

	cryptoutilTelemetry "cryptoutil/internal/common/telemetry"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
)

func StartContainer(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, containerRequest testcontainers.ContainerRequest) (testcontainers.Container, func(), error) {
	telemetryService.Slogger.Debug("starting container")

	startedContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: containerRequest,
		Started:          true,
	})
	if err != nil {
		telemetryService.Slogger.Error("failed to start container", "error", err)

		return nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	terminateContainer := func() {
		telemetryService.Slogger.Debug("terminating container")

		err := startedContainer.Terminate(ctx)

		if err == nil {
			telemetryService.Slogger.Debug("successfully terminated container")
		} else {
			telemetryService.Slogger.Error("failed to terminate container")
		}
	}

	return startedContainer, terminateContainer, nil
}

func GetContainerHostAndMappedPort(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, container testcontainers.Container, port string) (string, string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		telemetryService.Slogger.Error("failed to get container host", "error", err)

		return "", "", fmt.Errorf("failed to get container host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(port))
	if err != nil {
		telemetryService.Slogger.Error("failed to get container mapped port", "error", err)

		return "", "", fmt.Errorf("failed to get container mapped port: %w", err)
	}

	return host, mappedPort.Port(), nil
}
