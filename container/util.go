package container

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
)

func StartContainer(ctx context.Context, containerRequest testcontainers.ContainerRequest) (testcontainers.Container, func(), error) {
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

func GetContainerHostAndMappedPort(ctx context.Context, container testcontainers.Container, port string) (string, string, error) {
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
