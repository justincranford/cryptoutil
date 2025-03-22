package orm

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const postgresContainerStartupTimeout = 30 * time.Second

func startPostgresContainer(ctx context.Context, dbName, dbUsername, dbPassword string) (string, func(), error) {
	postgresContainerRequest := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env:          map[string]string{"POSTGRES_DB": dbName, "POSTGRES_USER": dbUsername, "POSTGRES_PASSWORD": dbPassword},
		// WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(postgresContainerStartupTimeout),
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithStartupTimeout(postgresContainerStartupTimeout),
	}

	container, terminateContainer, err := startContainer(ctx, postgresContainerRequest)
	if err != nil {
		return "", nil, fmt.Errorf("failed to start sqlite container: %w", err)
	}

	containerHost, containerMappedPort, err := getContainerHostAndPort(ctx, container, "5432")
	if err != nil {
		terminateContainer()
		return "", nil, fmt.Errorf("failed to get sqlite container host and port: %w", err)
	}

	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUsername, dbPassword, containerHost, containerMappedPort, dbName)

	return databaseUrl, terminateContainer, nil
}

func startContainer(ctx context.Context, containerRequest testcontainers.ContainerRequest) (testcontainers.Container, func(), error) {
	startedContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: containerRequest,
		Started:          true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to start container: %w", err)
	}

	terminateContainer := func() {
		log.Printf("terminating container")
		err := startedContainer.Terminate(ctx)
		if err == nil {
			log.Printf("successfully terminated container")
		} else {
			log.Printf("failed to terminate container: %v", err)
		}
	}

	return startedContainer, terminateContainer, nil
}

func getContainerHostAndPort(ctx context.Context, container testcontainers.Container, port string) (string, string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return "", "", fmt.Errorf("failed to get container host: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, nat.Port(port))
	if err != nil {
		return "", "", fmt.Errorf("failed to get container mapped port: %w", err)
	}

	return host, mappedPort.Port(), nil
}
