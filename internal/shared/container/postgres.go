// Copyright (c) 2025 Justin Cranford
//
//

package container

import (
	"context"
	"fmt"

	cryptoutilMagic "cryptoutil/internal/shared/magic"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func StartPostgres(ctx context.Context, telemetryService *cryptoutilTelemetry.TelemetryService, dbName, dbUsername, dbPassword string) (string, func(), error) {
	postgresContainerRequest := testcontainers.ContainerRequest{
		Image:        "postgres:18",
		ExposedPorts: []string{"5432/tcp"},
		Env:          map[string]string{"POSTGRES_DB": dbName, "POSTGRES_USER": dbUsername, "POSTGRES_PASSWORD": dbPassword},
		// WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(postgresContainerStartupTimeout),
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithStartupTimeout(cryptoutilMagic.DBPostgresContainerStartupTimeout),
	}

	container, terminateContainer, err := StartContainer(ctx, telemetryService, postgresContainerRequest)
	if err != nil {
		telemetryService.Slogger.Error("failed to start postgres container", "error", err)

		return "", nil, fmt.Errorf("failed to start sqlite container: %w", err)
	}

	containerHost, containerMappedPort, err := GetContainerHostAndMappedPort(ctx, telemetryService, container, "5432")
	if err != nil {
		telemetryService.Slogger.Error("failed to get postgres container host and mapped port", "error", err)
		terminateContainer()

		return "", nil, fmt.Errorf("failed to get postgres container host and mapped port: %w", err)
	}

	databaseURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUsername, dbPassword, containerHost, containerMappedPort, dbName)

	return databaseURL, terminateContainer, nil
}
