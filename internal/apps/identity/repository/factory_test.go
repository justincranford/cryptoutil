// Copyright (c) 2025-2026 Justin Cranford.
//

package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityConfig "cryptoutil/internal/apps/identity/config"
)

const testKeyRepositoryName = "key"

func newRepositoryFactoryForTest(t *testing.T) *RepositoryFactory {
	t.Helper()

	dbID, err := googleUuid.NewV7()
	require.NoError(t, err)

	factory, err := NewRepositoryFactory(context.Background(), &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  fmt.Sprintf(cryptoutilSharedMagic.SQLiteInMemoryNamedDSNFormat, dbID.String()),
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = factory.Close()
	})

	return factory
}

func TestRepositoryFactory_RepositoriesAndTransactions(t *testing.T) {
	t.Parallel()

	factory := newRepositoryFactoryForTest(t)
	ctx := context.Background()

	require.NoError(t, factory.AutoMigrate(ctx))
	require.NotNil(t, factory.DB())

	repositories := []struct {
		name       string
		repository any
	}{
		{name: "user", repository: factory.User()},
		{name: "user alias", repository: factory.UserRepository()},
		{name: "user with context", repository: factory.UserWithContext(ctx)},
		{name: "client", repository: factory.ClientRepository()},
		{name: cryptoutilSharedMagic.ParamToken, repository: factory.TokenRepository()},
		{name: "session", repository: factory.SessionRepository()},
		{name: "client profile", repository: factory.ClientProfileRepository()},
		{name: "auth flow", repository: factory.AuthFlowRepository()},
		{name: "auth profile", repository: factory.AuthProfileRepository()},
		{name: "MFA factor", repository: factory.MFAFactorRepository()},
		{name: "authorization request", repository: factory.AuthorizationRequestRepository()},
		{name: "consent decision", repository: factory.ConsentDecisionRepository()},
		{name: testKeyRepositoryName, repository: factory.KeyRepository()},
		{name: "device authorization", repository: factory.DeviceAuthorizationRepository()},
		{name: "pushed authorization request", repository: factory.PushedAuthorizationRequestRepository()},
		{name: "recovery code", repository: factory.RecoveryCodeRepository()},
		{name: "email OTP", repository: factory.EmailOTPRepository()},
		{name: "JTI replay cache", repository: factory.JTIReplayCacheRepository()},
	}

	for _, tc := range repositories {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.NotNil(t, tc.repository)
		})
	}

	require.NoError(t, factory.Transaction(ctx, func(txCtx context.Context) error {
		require.NotSame(t, factory.db, factory.getDB(txCtx))

		return nil
	}))

	transactionErr := fmt.Errorf("injected transaction error")
	err := factory.Transaction(ctx, func(txCtx context.Context) error {
		require.NotSame(t, factory.db, factory.getDB(txCtx))

		return transactionErr
	})
	require.ErrorIs(t, err, transactionErr)
	require.ErrorContains(t, err, "transaction failed")
	require.Same(t, factory.db, factory.getDB(ctx))
}

func TestRepositoryFactory_ValidationAndCloseErrors(t *testing.T) {
	t.Parallel()

	_, err := NewRepositoryFactory(context.Background(), &cryptoutilIdentityConfig.DatabaseConfig{
		Type: testUnsupportedDatabaseType,
		DSN:  "unused",
	})
	require.ErrorContains(t, err, testUnsupportedDatabaseTypeMessage)

	factory := newRepositoryFactoryForTest(t)
	require.NoError(t, factory.Close())
	require.ErrorContains(t, factory.AutoMigrate(context.Background()), "database migration failed")

	stub := NewRepositoryFactoryForTesting(nil, nil)
	require.Nil(t, stub.User())
	require.Nil(t, stub.ClientRepository())

	invalidDBFactory := &RepositoryFactory{db: &gorm.DB{Config: &gorm.Config{}}}
	require.ErrorContains(t, invalidDBFactory.Close(), "failed to get database instance")
	require.ErrorContains(t, invalidDBFactory.AutoMigrate(context.Background()), "failed to get sql.DB")

	closeErrDB := sql.OpenDB(closeErrorConnector{})
	require.NoError(t, closeErrDB.PingContext(context.Background()))
	closeErrFactory := &RepositoryFactory{db: &gorm.DB{Config: &gorm.Config{ConnPool: closeErrDB}}}
	require.ErrorContains(t, closeErrFactory.Close(), "failed to close database")
}

func TestInitializeDatabase_PostgresPingFailure(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := initializeDatabase(ctx, &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.DockerServicePostgres,
		DSN:  "host=127.0.0.1 port=1 user=test dbname=test sslmode=disable",
	})
	require.ErrorContains(t, err, "failed to connect to database")
}

func TestInitializeDatabase_SQLite(t *testing.T) {
	t.Parallel()

	db, err := initializeDatabase(context.Background(), &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.Equal(t, 5, sqlDB.Stats().MaxOpenConnections)
	require.NoError(t, sqlDB.Close())
}

func TestInitializeDatabase_SQLiteCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := initializeDatabase(ctx, &cryptoutilIdentityConfig.DatabaseConfig{
		Type: cryptoutilSharedMagic.TestDatabaseSQLite,
		DSN:  cryptoutilSharedMagic.SQLiteInMemoryDSN,
	})
	require.ErrorContains(t, err, "failed to enable WAL mode")
	require.ErrorIs(t, err, cryptoutilIdentityAppErr.ErrDatabaseConnection)
}

type closeErrorConnector struct{}

func (closeErrorConnector) Connect(context.Context) (driver.Conn, error) {
	return closeErrorConnection{}, nil
}

func (closeErrorConnector) Driver() driver.Driver {
	return closeErrorDriver{}
}

type closeErrorDriver struct{}

func (closeErrorDriver) Open(string) (driver.Conn, error) {
	return closeErrorConnection{}, nil
}

type closeErrorConnection struct{}

func (closeErrorConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is unsupported")
}

func (closeErrorConnection) Close() error {
	return errors.New("injected connection close failure")
}

func (closeErrorConnection) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are unsupported")
}

func (closeErrorConnection) Ping(context.Context) error {
	return nil
}

func (closeErrorConnection) ExecContext(context.Context, string, []driver.NamedValue) (driver.Result, error) {
	return driver.RowsAffected(0), nil
}

func (closeErrorConnection) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return nil, errors.New("injected query failure")
}
