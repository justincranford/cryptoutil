// Copyright (c) 2025-2026 Justin Cranford.

package test_help_db

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	postgresContainerModule "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/gorm"
)

const testUnusedPostgresURL = "postgres://unused"

func TestSafeNewPostgresTestContainer_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		factoryFn func(context.Context) (*postgresContainerModule.PostgresContainer, error)
		wantErr   string
	}{
		{
			name: "factory error",
			factoryFn: func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
				return nil, errors.New("factory error")
			},
			wantErr: "postgres container",
		},
		{
			name: "factory panic",
			factoryFn: func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
				panic("factory panic")
			},
			wantErr: testDBErrorDockerUnavailablePanic,
		},
		{
			name: testDBCaseSuccess,
			factoryFn: func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
				return &postgresContainerModule.PostgresContainer{}, nil
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			container, err := safeNewPostgresTestContainer(context.Background(), tc.factoryFn)
			if tc.wantErr != "" {
				require.Nil(t, container)
				require.Error(t, err)
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, container)
		})
	}
}

func TestNewPostgresTestContainerWithDeps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                    string
		ctxNil                  bool
		configure               func(*dbDeps, *bool)
		wantErr                 string
		wantCleanupErrors       []string
		wantContainerTerminated bool
	}{
		{
			name:    "nil context",
			ctxNil:  true,
			wantErr: "context cannot be nil",
		},
		{
			name: "container factory error",
			configure: func(deps *dbDeps, _ *bool) {
				deps.newPostgresContainerFactoryFn = func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
					return nil, errors.New("factory failure")
				}
			},
			wantErr: "failed to start PostgreSQL container",
		},
		{
			name: "container factory panic",
			configure: func(deps *dbDeps, _ *bool) {
				deps.newPostgresContainerFactoryFn = func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
					panic("factory panic")
				}
			},
			wantErr: testDBErrorDockerUnavailablePanic,
		},
		{
			name: "nil container",
			configure: func(deps *dbDeps, _ *bool) {
				deps.newPostgresContainerFactoryFn = func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
					return nil, nil
				}
			},
			wantErr: "returned nil container",
		},
		{
			name: "connection string error terminates container",
			configure: func(deps *dbDeps, terminated *bool) {
				deps.containerConnectionStringFn = func(context.Context, *postgresContainerModule.PostgresContainer) (string, error) {
					return "", errors.New("connection string failure")
				}
				deps.containerTerminateFn = func(context.Context, *postgresContainerModule.PostgresContainer) error {
					*terminated = true

					return nil
				}
			},
			wantErr:                 "failed to get PostgreSQL connection string",
			wantContainerTerminated: true,
		},
		{
			name: "GORM open error terminates container",
			configure: func(deps *dbDeps, terminated *bool) {
				deps.openGormPostgresFn = func(string) (*gorm.DB, error) {
					return nil, errors.New("GORM open failure")
				}
				deps.containerTerminateFn = func(context.Context, *postgresContainerModule.PostgresContainer) error {
					*terminated = true

					return nil
				}
			},
			wantErr:                 "failed to open GORM PostgreSQL DB",
			wantContainerTerminated: true,
		},
		{
			name: testDBCaseSuccess,
		},
		{
			name: "cleanup reports database and container errors",
			configure: func(deps *dbDeps, _ *bool) {
				deps.getSQLDBFn = func(*gorm.DB) (*sql.DB, error) {
					return &sql.DB{}, nil
				}
				deps.closeSQLDBFn = func(*sql.DB) error {
					return errors.New("close failure")
				}
				deps.containerTerminateFn = func(context.Context, *postgresContainerModule.PostgresContainer) error {
					return errors.New("terminate failure")
				}
			},
			wantCleanupErrors: []string{"failed to close PostgreSQL DB", "failed to terminate PostgreSQL container"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			deps := defaultDBDeps()
			container := &postgresContainerModule.PostgresContainer{}
			deps.newPostgresContainerFactoryFn = func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
				return container, nil
			}
			deps.containerConnectionStringFn = func(context.Context, *postgresContainerModule.PostgresContainer) (string, error) {
				return testUnusedPostgresURL, nil
			}
			deps.openGormPostgresFn = func(string) (*gorm.DB, error) {
				return &gorm.DB{}, nil
			}
			deps.getSQLDBFn = func(*gorm.DB) (*sql.DB, error) {
				return &sql.DB{}, nil
			}
			deps.closeSQLDBFn = func(*sql.DB) error { return nil }
			deps.containerTerminateFn = func(context.Context, *postgresContainerModule.PostgresContainer) error { return nil }

			containerTerminated := false
			if tc.configure != nil {
				tc.configure(&deps, &containerTerminated)
			}

			ctx := context.Background()
			if tc.ctxNil {
				ctx = nil
			}

			db, cleanup, err := newPostgresTestContainerWithDeps(ctx, deps)
			if tc.wantErr != "" {
				require.Nil(t, db)
				require.Nil(t, cleanup)
				require.ErrorContains(t, err, tc.wantErr)
				require.Equal(t, tc.wantContainerTerminated, containerTerminated)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, db)
			require.NotNil(t, cleanup)

			cleanupErr := cleanup()
			for _, wantErr := range tc.wantCleanupErrors {
				require.ErrorContains(t, cleanupErr, wantErr)
			}

			if len(tc.wantCleanupErrors) == 0 {
				require.NoError(t, cleanupErr)
			}
		})
	}
}
