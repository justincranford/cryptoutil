package container

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const postgresContainerStartupTimeout = 30 * time.Second

func StartPostgres(ctx context.Context, dbName, dbUsername, dbPassword string) (string, func(), error) {
	postgresContainerRequest := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env:          map[string]string{"POSTGRES_DB": dbName, "POSTGRES_USER": dbUsername, "POSTGRES_PASSWORD": dbPassword},
		// WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(postgresContainerStartupTimeout),
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithStartupTimeout(postgresContainerStartupTimeout),
	}

	container, terminateContainer, err := StartContainer(ctx, postgresContainerRequest)
	if err != nil {
		return "", nil, fmt.Errorf("failed to start sqlite container: %w", err)
	}

	containerHost, containerMappedPort, err := GetContainerHostAndMappedPort(ctx, container, "5432")
	if err != nil {
		terminateContainer()
		return "", nil, fmt.Errorf("failed to get sqlite container host and port: %w", err)
	}

	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUsername, dbPassword, containerHost, containerMappedPort, dbName)

	return databaseUrl, terminateContainer, nil
}
