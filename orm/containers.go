package orm

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startPostgresContainer(ctx context.Context, dbName, dbUsername, dbPassword string) (string, func(), error) {
	postgresContainerRequest := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       dbName,
			"POSTGRES_USER":     dbUsername,
			"POSTGRES_PASSWORD": dbPassword,
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	startedPostgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresContainerRequest,
		Started:          true,
	})
	if err != nil {
		return "", nil, fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}
	terminatePostgresContainer := func() {
		log.Printf("terminating PostgreSQL container: %v", err)
		err := startedPostgresContainer.Terminate(ctx)
		if err == nil {
			log.Printf("successfully terminated PostgreSQL container: %v", err)
		} else {
			log.Printf("failed to terminate PostgreSQL container: %v", err)
		}
	}

	host, err := startedPostgresContainer.Host(ctx)
	if err != nil {
		terminatePostgresContainer()
		return "", nil, fmt.Errorf("failed to get host: %w", err)
	}

	mappedPort, err := startedPostgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		terminatePostgresContainer()
		return "", nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, mappedPort.Port(), dbUsername, dbPassword, dbName)

	return dsn, terminatePostgresContainer, nil
}
