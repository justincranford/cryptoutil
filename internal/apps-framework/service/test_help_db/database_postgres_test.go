// Copyright (c) 2025-2026 Justin Cranford.

package test_help_db

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	postgresContainerModule "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/gorm"
)

const (
	testDBErrorPostgresContainer = "postgres container"
	testDBErrorStartPostgres     = "failed to start PostgreSQL container"
	testUnusedPostgresURL        = "postgres://unused"
)

func TestSafeNewPostgresTestContainer_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		factoryFn        func(context.Context) (*postgresContainerModule.PostgresContainer, error)
		partialContainer bool
		cleanupFails     bool
		wantTerminated   bool
		wantErr          string
		wantCleanupErr   string
	}{
		{
			name: "factory error",
			factoryFn: func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
				return nil, errors.New("factory error")
			},
			wantErr: testDBErrorPostgresContainer,
		},
		{
			name:             "factory error with partial container terminates it",
			partialContainer: true,
			wantTerminated:   true,
			wantErr:          testDBErrorPostgresContainer,
		},
		{
			name:             "factory and termination errors are joined",
			partialContainer: true,
			cleanupFails:     true,
			wantTerminated:   true,
			wantErr:          testDBErrorPostgresContainer,
			wantCleanupErr:   "failed to terminate PostgreSQL container after startup error",
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

			partial := &postgresContainerModule.PostgresContainer{}

			factoryFn := tc.factoryFn
			if tc.partialContainer {
				factoryFn = func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
					return partial, errors.New("factory error")
				}
			}

			containerTerminated := false
			terminateFn := func(context.Context, *postgresContainerModule.PostgresContainer) error {
				containerTerminated = true

				if tc.cleanupFails {
					return errors.New("termination failure")
				}

				return nil
			}

			container, err := safeNewPostgresTestContainer(context.Background(), factoryFn, terminateFn)

			require.Equal(t, tc.wantTerminated, containerTerminated)

			if tc.wantErr != "" {
				require.Nil(t, container)
				require.Error(t, err)
				require.ErrorContains(t, err, tc.wantErr)

				if tc.wantCleanupErr != "" {
					require.ErrorContains(t, err, tc.wantCleanupErr)
				}

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
			wantErr: testDBErrorStartPostgres,
		},
		{
			name: "partial container factory error terminates container",
			configure: func(deps *dbDeps, terminated *bool) {
				container := &postgresContainerModule.PostgresContainer{}
				deps.newPostgresContainerFactoryFn = func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
					return container, errors.New("factory failure")
				}
				deps.containerTerminateFn = func(context.Context, *postgresContainerModule.PostgresContainer) error {
					*terminated = true

					return nil
				}
			},
			wantErr:                 testDBErrorStartPostgres,
			wantContainerTerminated: true,
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
			name: "connection string and termination errors are joined",
			configure: func(deps *dbDeps, _ *bool) {
				deps.containerConnectionStringFn = func(context.Context, *postgresContainerModule.PostgresContainer) (string, error) {
					return "", errors.New("connection string failure")
				}
				deps.containerTerminateFn = func(context.Context, *postgresContainerModule.PostgresContainer) error {
					return errors.New("termination failure")
				}
			},
			wantErr: "failed to terminate PostgreSQL container after setup failure",
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
			name: "nil GORM database terminates container",
			configure: func(deps *dbDeps, terminated *bool) {
				deps.openGormPostgresFn = func(string) (*gorm.DB, error) {
					return nil, nil
				}
				deps.containerTerminateFn = func(context.Context, *postgresContainerModule.PostgresContainer) error {
					*terminated = true

					return nil
				}
			},
			wantErr:                 "GORM PostgreSQL opener returned nil database",
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
		{
			name: "cleanup reports nil sql database",
			configure: func(deps *dbDeps, _ *bool) {
				deps.getSQLDBFn = func(*gorm.DB) (*sql.DB, error) {
					return nil, nil
				}
			},
			wantCleanupErrors: []string{"GORM PostgreSQL database returned nil sql.DB"},
		},
		{
			name: "cleanup reports sql database lookup error",
			configure: func(deps *dbDeps, _ *bool) {
				deps.getSQLDBFn = func(*gorm.DB) (*sql.DB, error) {
					return nil, errors.New("database lookup failure")
				}
			},
			wantCleanupErrors: []string{"failed to get sql.DB for cleanup"},
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

type testPostgresContainer struct {
	testcontainers.Container
	endpoint        string
	endpointErr     error
	terminateErr    error
	terminateCalled bool
	terminateCtxErr error
}

func (c *testPostgresContainer) PortEndpoint(context.Context, string, string) (string, error) {
	return c.endpoint, c.endpointErr
}

func (c *testPostgresContainer) Terminate(ctx context.Context, _ ...testcontainers.TerminateOption) error {
	c.terminateCalled = true
	c.terminateCtxErr = ctx.Err()

	return c.terminateErr
}

func TestDefaultPostgresSeamFunctions(t *testing.T) {
	t.Parallel()

	t.Run("default PostgreSQL GORM opener wraps invalid DSN", func(t *testing.T) {
		t.Parallel()

		db, err := defaultOpenGormPostgres("not a postgres URL")
		require.NotNil(t, db)
		require.ErrorContains(t, err, "gorm.Open postgres")
	})

	t.Run("connection string success and error", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name      string
			container *testPostgresContainer
			wantErr   string
		}{
			{
				name:      testDBCaseSuccess,
				container: &testPostgresContainer{endpoint: "127.0.0.1:5432"},
			},
			{
				name:      "endpoint error",
				container: &testPostgresContainer{endpointErr: errors.New("endpoint failure")},
				wantErr:   "ConnectionString",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				container := &postgresContainerModule.PostgresContainer{Container: tc.container}

				connStr, err := defaultContainerConnectionString(context.Background(), container)
				if tc.wantErr != "" {
					require.Empty(t, connStr)
					require.ErrorContains(t, err, tc.wantErr)

					return
				}

				require.NoError(t, err)
				require.Contains(t, connStr, "127.0.0.1:5432")
				require.Contains(t, connStr, "sslmode=disable")
			})
		}
	})

	t.Run("container termination success and error", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name        string
			err         error
			ctxCanceled bool
			wantErr     string
		}{
			{name: testDBCaseSuccess, ctxCanceled: true},
			{name: cryptoutilSharedMagic.StringError, err: errors.New("terminate failure"), wantErr: "terminate"},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				stub := &testPostgresContainer{terminateErr: tc.err}
				ctx := context.Background()

				if tc.ctxCanceled {
					var cancel context.CancelFunc

					ctx, cancel = context.WithCancel(ctx)
					cancel()
				}

				err := defaultContainerTerminate(ctx, &postgresContainerModule.PostgresContainer{Container: stub})
				if tc.wantErr != "" {
					require.ErrorContains(t, err, tc.wantErr)
				} else {
					require.NoError(t, err)
				}

				require.True(t, stub.terminateCalled)

				if tc.ctxCanceled {
					require.NoError(t, stub.terminateCtxErr)
				}
			})
		}
	})
}

func TestNewPostgresTestContainerTestHooks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		ctxNil       bool
		cleanupError bool
		wantFatal    bool
		wantError    bool
	}{
		{name: "setup error", ctxNil: true, wantFatal: true},
		{name: testDBCaseSuccess},
		{name: testDBCaseCleanupError, cleanupError: true, wantError: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			deps := defaultDBDeps()
			deps.newPostgresContainerFactoryFn = func(context.Context) (*postgresContainerModule.PostgresContainer, error) {
				return &postgresContainerModule.PostgresContainer{}, nil
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
			deps.closeSQLDBFn = func(*sql.DB) error {
				if tc.cleanupError {
					return errors.New("close failure")
				}

				return nil
			}
			deps.containerTerminateFn = func(context.Context, *postgresContainerModule.PostgresContainer) error {
				return nil
			}

			var (
				cleanupCallback func()
				fatalMessage    string
				errorMessage    string
			)

			ctx := context.Background()
			if tc.ctxNil {
				ctx = nil
			}

			db := newPostgresTestContainerWithTestHooks(
				ctx,
				t,
				deps,
				func(format string, args ...any) { fatalMessage = format },
				func(fn func()) { cleanupCallback = fn },
				func(format string, args ...any) { errorMessage = format },
			)

			if tc.wantFatal {
				require.Nil(t, db)
				require.Contains(t, fatalMessage, "failed to create PostgreSQL test container")

				return
			}

			require.NotNil(t, db)
			require.NotNil(t, cleanupCallback)
			cleanupCallback()

			if tc.wantError {
				require.Contains(t, errorMessage, "failed to clean up PostgreSQL test container")
			} else {
				require.Empty(t, errorMessage)
			}
		})
	}
}
