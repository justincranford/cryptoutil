package container

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// NewPostgresTestContainer creates and starts a new PostgreSQL test container.
func NewPostgresTestContainer(ctx context.Context) (*postgres.PostgresContainer, error) {
	return newPostgresTestContainer(ctx, postgres.Run)
}

func newPostgresTestContainer(
	ctx context.Context,
	runPostgresFn func(context.Context, string, ...testcontainers.ContainerCustomizer) (*postgres.PostgresContainer, error),
) (*postgres.PostgresContainer, error) {
	dbName := fmt.Sprintf("test_%s", googleUuid.NewString())
	username := fmt.Sprintf("user_%s", googleUuid.NewString())
	password := fmt.Sprintf("pass_%s", googleUuid.NewString())

	container, err := runPostgresFn(ctx,
		"postgres:latest",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(username),
		postgres.WithPassword(password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(cryptoutilSharedMagic.DBPostgresContainerStartupTimeout),
		),
	)
	if err != nil {
		startErr := fmt.Errorf("failed to start container: %w", err)
		if container != nil {
			startErr = cleanupTestContainerAfterError(ctx, container, startErr, "failed to clean up container after startup failure")
		}

		return nil, startErr
	}

	return container, nil
}
