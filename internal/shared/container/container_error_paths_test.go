// Copyright (c) 2025-2026 Justin Cranford.
package container

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedTelemetry "cryptoutil/internal/shared/telemetry"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	postgresDriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type connectionlessSQLConnector struct{}

type postgresConnectionStringContainer struct {
	testcontainers.Container
	endpoint string
	err      error
}

func (c *postgresConnectionStringContainer) PortEndpoint(context.Context, string, string) (string, error) {
	return c.endpoint, c.err
}

func (connectionlessSQLConnector) Connect(context.Context) (driver.Conn, error) {
	return nil, errors.New("unexpected database connection")
}

func (connectionlessSQLConnector) Driver() driver.Driver {
	return connectionlessSQLDriver{}
}

type connectionlessSQLDriver struct{}

type successfulSQLConnector struct{}

type successfulSQLConnection struct{}

type successfulSQLDriver struct{}

const containerTestSuccess = "success"

func (connectionlessSQLDriver) Open(string) (driver.Conn, error) {
	return nil, errors.New("unexpected database connection")
}

func (successfulSQLConnector) Connect(context.Context) (driver.Conn, error) {
	return successfulSQLConnection{}, nil
}

func (successfulSQLConnector) Driver() driver.Driver {
	return successfulSQLDriver{}
}

func (successfulSQLConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("unexpected statement preparation")
}

func (successfulSQLConnection) Close() error {
	return nil
}

func (successfulSQLConnection) Begin() (driver.Tx, error) {
	return nil, errors.New("unexpected transaction")
}

func (successfulSQLConnection) Ping(context.Context) error {
	return nil
}

func (successfulSQLDriver) Open(string) (driver.Conn, error) {
	return successfulSQLConnection{}, nil
}

func TestContainerEntrypoints_PropagateCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	telemetryService := setupTestTelemetry(t)

	tests := []struct {
		name string
		run  func() error
	}{
		{
			name: "start container",
			run: func() error {
				_, _, err := StartContainer(ctx, telemetryService, testcontainers.ContainerRequest{})

				return err
			},
		},
		{
			name: "start postgres",
			run: func() error {
				_, _, err := StartPostgres(ctx, telemetryService, "db", "user", "password")

				return err
			},
		},
		{
			name: "shared postgres setup",
			run: func() error {
				_, _, err := SetupSharedPostgresContainer(ctx)

				return err
			},
		},
		{
			name: "postgres test container",
			run: func() error {
				_, err := NewPostgresTestContainer(ctx)

				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Error(t, tc.run())
		})
	}
}

func TestPostgresConnectionAdapters(t *testing.T) {
	t.Parallel()

	_, err := getSQLDBFromGORM(&gorm.DB{Config: &gorm.Config{}})
	require.Error(t, err)

	gormDB, err := gorm.Open(postgresDriver.Open("host=127.0.0.1 port=1 user=test dbname=test sslmode=disable"), &gorm.Config{
		DisableAutomaticPing: true,
	})
	require.NoError(t, err)

	sqlDB, err := getSQLDBFromGORM(gormDB)
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	connectionlessDB := sql.OpenDB(connectionlessSQLConnector{})

	t.Cleanup(func() {
		require.NoError(t, connectionlessDB.Close())
	})

	require.Error(t, pingPostgresSQLDB(connectionlessDB))

	successfulDB := sql.OpenDB(successfulSQLConnector{})

	t.Cleanup(func() {
		require.NoError(t, successfulDB.Close())
	})

	require.NoError(t, pingPostgresSQLDB(successfulDB))
}

func TestPostgresContainerConnectionString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		endpointErr error
		wantErr     string
	}{
		{name: "endpoint failure", endpointErr: errors.New("endpoint lookup failed"), wantErr: "failed to get PostgreSQL container connection string"},
		{name: containerTestSuccess},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			container := &postgres.PostgresContainer{
				Container: &postgresConnectionStringContainer{
					endpoint: "127.0.0.1:54321",
					err:      tc.endpointErr,
				},
			}

			connectionString, err := postgresContainerConnectionString(context.Background(), container)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)

				return
			}

			require.NoError(t, err)
			require.Contains(t, connectionString, cryptoutilSharedMagic.IPv4Loopback)
			require.Contains(t, connectionString, "54321")
		})
	}
}

func TestStartContainer_InjectedGenericContainer(t *testing.T) {
	t.Parallel()

	telemetryService := setupTestTelemetry(t)
	injectedErr := errors.New("injected startup error")
	tests := []struct {
		name         string
		container    testcontainers.Container
		startErr     error
		terminateErr error
		wantErr      bool
	}{
		{name: "startup failure", startErr: injectedErr, wantErr: true},
		{name: "successful start and termination", container: &mockContainer{terminateFn: func(context.Context, ...testcontainers.TerminateOption) error { return nil }}},
		{name: "termination failure is logged", container: &mockContainer{terminateFn: func(context.Context, ...testcontainers.TerminateOption) error { return injectedErr }}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotContainer, cleanup, err := startContainer(context.Background(), telemetryService, testcontainers.ContainerRequest{}, func(context.Context, testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
				return tc.container, tc.startErr
			})

			if tc.wantErr {
				require.ErrorIs(t, err, injectedErr)
				require.Nil(t, gotContainer)
				require.Nil(t, cleanup)

				return
			}

			require.NoError(t, err)
			require.Same(t, tc.container, gotContainer)
			require.NotNil(t, cleanup)
			cleanup()
		})
	}
}

func TestStartPostgres_InjectedContainerFunctions(t *testing.T) {
	t.Parallel()

	telemetryService := setupTestTelemetry(t)
	injectedErr := errors.New("injected postgres error")
	tests := []struct {
		name          string
		startErr      error
		getErr        error
		wantErr       string
		wantTerminate bool
	}{
		{name: "start failure", startErr: injectedErr, wantErr: "failed to start postgres container"},
		{name: "host lookup failure", getErr: injectedErr, wantErr: "failed to get postgres container host and mapped port", wantTerminate: true},
		{name: containerTestSuccess, wantTerminate: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			terminateCount := 0
			container := &mockContainer{
				terminateFn: func(context.Context, ...testcontainers.TerminateOption) error {
					terminateCount++

					return nil
				},
			}
			startFn := func(context.Context, *cryptoutilSharedTelemetry.TelemetryService, testcontainers.ContainerRequest) (testcontainers.Container, func(), error) {
				if tc.startErr != nil {
					return nil, nil, tc.startErr
				}

				return container, func() { terminateCount++ }, nil
			}
			hostFn := func(context.Context, *cryptoutilSharedTelemetry.TelemetryService, testcontainers.Container, string) (string, string, error) {
				if tc.getErr != nil {
					return "", "", tc.getErr
				}

				return cryptoutilSharedMagic.IPv4Loopback, "54321", nil
			}

			dsn, terminate, err := startPostgres(context.Background(), telemetryService, "db", "user", "password", startFn, hostFn)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, dsn)
				require.NotNil(t, terminate)
				terminate()
			}

			if tc.wantTerminate {
				require.Equal(t, 1, terminateCount)
			} else {
				require.Zero(t, terminateCount)
			}
		})
	}
}

func TestSetupSharedPostgresContainer_InjectedFunctions(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected postgres setup error")
	tests := []struct {
		name           string
		passwordErr    error
		runErr         error
		connection     string
		connectionErr  error
		wantErr        string
		wantConnection string
	}{
		{name: "password failure", passwordErr: injectedErr, wantErr: "failed to generate database password"},
		{name: "container start failure", runErr: injectedErr, wantErr: "failed to start PostgreSQL container"},
		{name: containerTestConnectionStringFailure, connectionErr: injectedErr, wantErr: "failed to get connection string"},
		{name: "adds ssl mode", connection: "postgres://db", wantConnection: "postgres://db?sslmode=disable"},
		{name: "appends ssl mode", connection: "postgres://db?connect_timeout=1", wantConnection: "postgres://db?connect_timeout=1&sslmode=disable"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runCount := 0
			connectionCount := 0
			container, connection, err := setupSharedPostgresContainer(
				context.Background(),
				func() (string, error) { return newContainerTestPassword(t), tc.passwordErr },
				func(context.Context, string, ...testcontainers.ContainerCustomizer) (*postgres.PostgresContainer, error) {
					runCount++

					return nil, tc.runErr
				},
				func(context.Context, *postgres.PostgresContainer) (string, error) {
					connectionCount++

					return tc.connection, tc.connectionErr
				},
			)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantConnection, connection)
				require.Nil(t, container)
			}

			if tc.passwordErr != nil {
				require.Zero(t, runCount)
				require.Zero(t, connectionCount)
			} else if tc.runErr != nil {
				require.Equal(t, 1, runCount)
				require.Zero(t, connectionCount)
			} else {
				require.Equal(t, 1, runCount)
				require.Equal(t, 1, connectionCount)
			}
		})
	}
}

func TestNewPostgresTestContainer_InjectedRunner(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected postgres startup error")
	tests := []struct {
		name    string
		runErr  error
		wantErr bool
	}{
		{name: "runner error", runErr: injectedErr, wantErr: true},
		{name: "runner success"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			container, err := newPostgresTestContainer(context.Background(), func(context.Context, string, ...testcontainers.ContainerCustomizer) (*postgres.PostgresContainer, error) {
				return nil, tc.runErr
			})
			if tc.wantErr {
				require.ErrorIs(t, err, injectedErr)
				require.Nil(t, container)

				return
			}

			require.NoError(t, err)
			require.Nil(t, container)
			require.Nil(t, container)
		})
	}
}

func TestVerifyPostgresConnection_InjectedFunctions(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected verification error")
	tests := []struct {
		name      string
		openErr   error
		sqlDBErr  error
		pingErr   error
		wantError string
	}{
		{name: "open error", openErr: injectedErr, wantError: "failed to connect to PostgreSQL"},
		{name: "SQL DB error", sqlDBErr: injectedErr, wantError: "failed to get SQL database instance"},
		{name: "ping error", pingErr: injectedErr, wantError: "failed to ping PostgreSQL"},
		{name: "success"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := verifyPostgresConnectionWithFns(
				"test-dsn",
				func(string) (*gorm.DB, error) {
					if tc.openErr != nil {
						return nil, tc.openErr
					}

					return &gorm.DB{}, nil
				},
				func(*gorm.DB) (*sql.DB, error) {
					if tc.sqlDBErr != nil {
						return nil, tc.sqlDBErr
					}

					return sql.OpenDB(connectionlessSQLConnector{}), nil
				},
				func(*sql.DB) error {
					return tc.pingErr
				},
			)

			if tc.wantError != "" {
				require.ErrorContains(t, err, tc.wantError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStartContainer_InvalidRequestError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	telemetryService := setupTestTelemetry(t)

	_, cleanup, err := StartContainer(ctx, telemetryService, testcontainers.ContainerRequest{})
	require.Error(t, err)
	require.Nil(t, cleanup)
}

func TestStartPostgres_CanceledContextError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	telemetryService := setupTestTelemetry(t)

	_, cleanup, err := StartPostgres(ctx, telemetryService, "db", "user", "password")
	require.ErrorContains(t, err, "failed to start postgres container")
	require.Nil(t, cleanup)
}

func TestPostgresSetupHelpers_UsePostgresImage(t *testing.T) {
	t.Parallel()

	imageSeen := ""
	container, err := newPostgresTestContainer(context.Background(), func(_ context.Context, image string, _ ...testcontainers.ContainerCustomizer) (*postgres.PostgresContainer, error) {
		imageSeen = image

		return nil, fmt.Errorf("injected runner stop")
	})
	require.ErrorContains(t, err, "injected runner stop")
	require.Nil(t, container)
	require.Equal(t, "postgres:latest", imageSeen)
}
