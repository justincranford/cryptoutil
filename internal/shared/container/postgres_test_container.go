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
	dbName := fmt.Sprintf("test_%s", googleUuid.NewString())
	username := fmt.Sprintf("user_%s", googleUuid.NewString())
	password := fmt.Sprintf("pass_%s", googleUuid.NewString())

	container, err := postgres.Run(ctx,
		"postgres:18-alpine",
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
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	return container, nil
}
