// Copyright (c) 2025-2026 Justin Cranford.
package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	testCaseNilTelemetryService = "nil telemetryService"
	testCaseNilGormDB           = "nil gormDB"
	testCaseNilJWKGenService    = "nil jwkGenService"
)

func TestNewOrmRepository_NilChecks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		telemetry       any
		gormDB          any
		jwkGen          any
		wantErrContains string
	}{
		{name: testCaseNilTelemetryService, telemetry: nil, wantErrContains: "telemetryService must be non-nil"},
		{name: testCaseNilGormDB, telemetry: testTelemetryService, gormDB: nil, wantErrContains: "gormDB must be non-nil"},
		{name: testCaseNilJWKGenService, telemetry: testTelemetryService, gormDB: testGormDB, jwkGen: nil, wantErrContains: "jwkGenService must be non-nil"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var (
				repo *OrmRepository
				err  error
			)

			switch tc.name {
			case testCaseNilTelemetryService:
				repo, err = NewOrmRepository(testCtx, nil, testGormDB, testJWKGenService, false)
			case testCaseNilGormDB:
				repo, err = NewOrmRepository(testCtx, testTelemetryService, nil, testJWKGenService, false)
			case testCaseNilJWKGenService:
				repo, err = NewOrmRepository(testCtx, testTelemetryService, testGormDB, nil, false)
			}

			require.Error(t, err)
			require.Contains(t, err.Error(), tc.wantErrContains)
			require.Nil(t, repo)
		})
	}
}

func TestNewOrmRepository_Success(t *testing.T) {
	t.Parallel()

	repo, err := NewOrmRepository(testCtx, testTelemetryService, testGormDB, testJWKGenService, true)
	require.NoError(t, err)
	require.NotNil(t, repo)
	require.NotNil(t, repo.GormDB())
	repo.Shutdown()
}

func TestOrmRepository_Shutdown_NoOp(t *testing.T) {
	t.Parallel()

	require.NotPanics(t, func() { testOrmRepository.Shutdown() })
	require.NotPanics(t, func() { testOrmRepository.Shutdown() })
}

func TestOrmRepository_GormDB(t *testing.T) {
	t.Parallel()

	require.NotNil(t, testOrmRepository.GormDB())
}

func TestOrmRepository_SetVerboseMode(t *testing.T) {
	t.Parallel()

	repo, err := NewOrmRepository(testCtx, testTelemetryService, testGormDB, testJWKGenService, false)
	require.NoError(t, err)
	require.NotNil(t, repo)

	require.NotPanics(t, func() { repo.SetVerboseMode(true) })
	require.NotPanics(t, func() { repo.SetVerboseMode(false) })
}

func TestOrmRepository_HealthCheck_NilGormDB(t *testing.T) {
	t.Parallel()

	repo := &OrmRepository{}
	result, err := repo.HealthCheck(testCtx)

	require.Error(t, err)
	require.Contains(t, err.Error(), "database connection not initialized")
	require.Equal(t, cryptoutilSharedMagic.StringError, result[cryptoutilSharedMagic.StringStatus])
}

func TestOrmRepository_HealthCheck_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		gormDB  func(*testing.T) *gorm.DB
		wantErr string
	}{
		{
			name: "GORM SQL handle error",
			gormDB: func(*testing.T) *gorm.DB {
				return &gorm.DB{Config: &gorm.Config{}}
			},
			wantErr: "failed to get sql.DB from GORM",
		},
		{
			name: "database ping error",
			gormDB: func(t *testing.T) *gorm.DB {
				t.Helper()

				sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, cryptoutilSharedMagic.SQLiteMemoryPlaceholder)
				require.NoError(t, err)
				require.NoError(t, sqlDB.Close())

				return &gorm.DB{
					Config: &gorm.Config{ConnPool: sqlDB},
					Statement: &gorm.Statement{
						ConnPool: sqlDB,
					},
				}
			},
			wantErr: "database ping failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := &OrmRepository{gormDB: tc.gormDB(t)}
			result, err := repo.HealthCheck(testCtx)

			require.ErrorContains(t, err, tc.wantErr)
			require.Equal(t, cryptoutilSharedMagic.StringError, result[cryptoutilSharedMagic.StringStatus])
		})
	}
}

func TestOrmRepository_HealthCheck_Success(t *testing.T) {
	t.Parallel()

	result, err := testOrmRepository.HealthCheck(testCtx)

	require.NoError(t, err)
	require.Equal(t, cryptoutilSharedMagic.StringStatusOK, result[cryptoutilSharedMagic.StringStatus])
	require.NotEmpty(t, result["db_type"])
}

func TestOrmRepository_WithTransaction_Success(t *testing.T) {
	t.Parallel()

	var called bool

	err := testOrmRepository.WithTransaction(testCtx, AutoCommit, func(tx *OrmTransaction) error {
		called = true

		require.NotNil(t, tx)
		require.NotNil(t, tx.GormTx())

		return nil
	})

	require.NoError(t, err)
	require.True(t, called)
}

func TestOrmRepository_WithTransaction_FnError(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, AutoCommit, func(_ *OrmTransaction) error {
		return fmt.Errorf("injected error")
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to execute transaction")
}

func TestOrmRepository_WithTransaction_ReadOnly(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, ReadOnly, func(tx *OrmTransaction) error {
		require.NotNil(t, tx.GormTx())
		require.Equal(t, ReadOnly, *tx.Mode())

		return nil
	})

	require.NoError(t, err)
}

func TestOrmRepository_WithTransaction_ErrorPaths(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected transaction failure")
	tests := []struct {
		name       string
		repository *OrmRepository
		fn         func(*OrmTransaction) error
		wantErr    string
		wantPanic  bool
	}{
		{
			name: "begin failure",
			repository: &OrmRepository{
				telemetryService: testTelemetryService,
				jwkGenService:    testJWKGenService,
				gormDB:           newOrmErrorGormDB(injectedErr),
			},
			fn:      func(*OrmTransaction) error { return nil },
			wantErr: "failed to begin transaction",
		},
		{
			name:       "transaction function error triggers rollback",
			repository: testOrmRepository,
			fn:         func(*OrmTransaction) error { return injectedErr },
			wantErr:    "failed to execute transaction",
		},
		{
			name:       "commit error triggers rollback attempt",
			repository: testOrmRepository,
			fn: func(tx *OrmTransaction) error {
				tx.state.gormTx = newOrmErrorGormDB(injectedErr)

				return nil
			},
			wantErr: "failed to commit transaction",
		},
		{
			name:       "panic is rethrown after rollback",
			repository: testOrmRepository,
			fn: func(*OrmTransaction) error {
				panic(injectedErr)
			},
			wantPanic: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.wantPanic {
				require.PanicsWithValue(t, injectedErr, func() {
					_ = tc.repository.WithTransaction(testCtx, ReadWrite, tc.fn)
				})

				return
			}

			err := tc.repository.WithTransaction(testCtx, ReadWrite, tc.fn)

			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

type ormFailingTransaction struct {
	*sql.DB
	err error
}

func newOrmErrorGormDB(err error) *gorm.DB {
	pool := &ormFailingTransaction{err: err}
	config := &gorm.Config{ConnPool: pool}
	db := &gorm.DB{Config: config}
	db.Statement = &gorm.Statement{DB: db, ConnPool: pool}

	return db
}

func (tx *ormFailingTransaction) BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error) {
	return nil, tx.err
}

func (tx *ormFailingTransaction) Commit() error {
	return tx.err
}

func (tx *ormFailingTransaction) Rollback() error {
	return tx.err
}

func TestOrmTransaction_NilStateAccessors(t *testing.T) {
	t.Parallel()

	tx := &OrmTransaction{}

	require.Nil(t, tx.GormTx())
	require.Nil(t, tx.ID())
	require.Nil(t, tx.Context())
	require.Nil(t, tx.Mode())
}

func TestOrmTransaction_CommitNilState(t *testing.T) {
	t.Parallel()

	tx := &OrmTransaction{ormRepository: testOrmRepository}
	err := tx.Commit()

	require.Error(t, err)
	require.Contains(t, err.Error(), "can't commit because transaction not active")
}

func TestOrmTransaction_RollbackNilState(t *testing.T) {
	t.Parallel()

	tx := &OrmTransaction{ormRepository: testOrmRepository}
	err := tx.Rollback()

	require.Error(t, err)
	require.Contains(t, err.Error(), "can't rollback because transaction not active")
}

func TestOrmTransaction_CommitAutoCommit(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, AutoCommit, func(tx *OrmTransaction) error {
		commitErr := tx.Commit()
		require.Error(t, commitErr)
		require.Contains(t, commitErr.Error(), "can't commit because transaction is autocommit")

		return nil
	})
	require.NoError(t, err)
}

func TestOrmTransaction_RollbackAutoCommit(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, AutoCommit, func(tx *OrmTransaction) error {
		rollbackErr := tx.Rollback()
		require.Error(t, rollbackErr)
		require.Contains(t, rollbackErr.Error(), "can't rollback because transaction is autocommit")

		return nil
	})
	require.NoError(t, err)
}

func TestOrmTransaction_BeginAlreadyStarted(t *testing.T) {
	t.Parallel()

	tx := &OrmTransaction{ormRepository: testOrmRepository}

	err := tx.Begin(testCtx, ReadWrite)
	require.NoError(t, err)
	require.NotNil(t, tx.ID())
	require.NotNil(t, tx.Context())
	require.NotNil(t, tx.Mode())

	err = tx.Begin(testCtx, ReadWrite)
	require.Error(t, err)
	require.Contains(t, err.Error(), "transaction already started")

	_ = tx.Rollback()
}

func TestOrmTransaction_VerboseLogging(t *testing.T) {
	t.Parallel()

	repo := &OrmRepository{
		telemetryService: testTelemetryService,
		jwkGenService:    testJWKGenService,
		gormDB:           testGormDB,
		verboseMode:      true,
	}
	tx := &OrmTransaction{ormRepository: repo}

	require.NoError(t, tx.Begin(testCtx, AutoCommit))
	require.ErrorContains(t, tx.Commit(), "can't commit because transaction is autocommit")
	require.ErrorContains(t, tx.Rollback(), "can't rollback because transaction is autocommit")
}

func TestOrmTransaction_ReadWriteLifecycle(t *testing.T) {
	t.Parallel()

	err := testOrmRepository.WithTransaction(testCtx, ReadWrite, func(tx *OrmTransaction) error {
		require.NotNil(t, tx.GormTx())
		require.Equal(t, ReadWrite, *tx.Mode())
		require.NotNil(t, tx.ID())
		require.NotNil(t, tx.Context())

		return nil
	})
	require.NoError(t, err)
}

func TestNewOrmTransactionWithRepository_NotNil(t *testing.T) {
	t.Parallel()

	tx := NewOrmTransactionWithRepository(testOrmRepository)
	require.NotNil(t, tx)
	require.Nil(t, tx.GormTx())
}
