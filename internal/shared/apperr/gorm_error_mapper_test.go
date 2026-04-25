// Copyright (c) 2025 Justin Cranford
//
//

package apperr_test

import (
	"database/sql"
	"io"
	"log/slog"
	http "net/http"
	"testing"

	cryptoutilSharedApperr "cryptoutil/internal/shared/apperr"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	gsqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "modernc.org/sqlite" // CGO-free SQLite driver registration.
)

// testModel is a minimal GORM model used to trigger real SQLite constraint errors.
type testModel struct {
	ID   string `gorm:"type:text;primaryKey"`
	Name string `gorm:"type:text;not null;uniqueIndex"`
}

// discardSlogger returns a *slog.Logger that discards all output.
func discardSlogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newSQLiteDBWithUniqueConstraint opens an in-memory SQLite DB (CGO-free modernc driver)
// and auto-migrates testModel. Uses database/sql + gsqlite.Dialector so that constraint
// errors surface as *modernc.org/sqlite.Error (not mattn/go-sqlite3.Error).
func newSQLiteDBWithUniqueConstraint(t *testing.T) *gorm.DB {
	t.Helper()

	sqlDB, err := sql.Open(cryptoutilSharedMagic.TestDatabaseSQLite, "file::memory:?cache=private")
	require.NoError(t, err)

	db, err := gorm.Open(gsqlite.Dialector{Conn: sqlDB}, &gorm.Config{
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(&testModel{}))

	t.Cleanup(func() { _ = sqlDB.Close() })

	return db
}

func TestMapGormError_IsAppErr(t *testing.T) {
	t.Parallel()

	msg := "operation failed"
	err := cryptoutilSharedApperr.MapGormError(discardSlogger(), &msg, cryptoutilSharedApperr.ErrCantBeNil)

	var appErr *cryptoutilSharedApperr.Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, http.StatusBadRequest, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestMapGormError_ErrRecordNotFound(t *testing.T) {
	t.Parallel()

	msg := "record not found"
	err := cryptoutilSharedApperr.MapGormError(discardSlogger(), &msg, gorm.ErrRecordNotFound)

	var appErr *cryptoutilSharedApperr.Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, http.StatusNotFound, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestMapGormError_ErrDuplicatedKey(t *testing.T) {
	t.Parallel()

	msg := "duplicated key"
	err := cryptoutilSharedApperr.MapGormError(discardSlogger(), &msg, gorm.ErrDuplicatedKey)

	var appErr *cryptoutilSharedApperr.Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, http.StatusBadRequest, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestMapGormError_ErrForeignKeyViolated(t *testing.T) {
	t.Parallel()

	msg := "foreign key violated"
	err := cryptoutilSharedApperr.MapGormError(discardSlogger(), &msg, gorm.ErrForeignKeyViolated)

	var appErr *cryptoutilSharedApperr.Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, http.StatusBadRequest, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestMapGormError_ErrCheckConstraintViolated(t *testing.T) {
	t.Parallel()

	msg := "check constraint violated"
	err := cryptoutilSharedApperr.MapGormError(discardSlogger(), &msg, gorm.ErrCheckConstraintViolated)

	var appErr *cryptoutilSharedApperr.Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, http.StatusBadRequest, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestMapGormError_ErrInvalidData(t *testing.T) {
	t.Parallel()

	msg := "invalid data"
	err := cryptoutilSharedApperr.MapGormError(discardSlogger(), &msg, gorm.ErrInvalidData)

	var appErr *cryptoutilSharedApperr.Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, http.StatusBadRequest, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestMapGormError_ErrInvalidValueOfLength(t *testing.T) {
	t.Parallel()

	msg := "invalid value of length"
	err := cryptoutilSharedApperr.MapGormError(discardSlogger(), &msg, gorm.ErrInvalidValueOfLength)

	var appErr *cryptoutilSharedApperr.Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, http.StatusBadRequest, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestMapGormError_ErrNotImplemented(t *testing.T) {
	t.Parallel()

	msg := "not implemented"
	err := cryptoutilSharedApperr.MapGormError(discardSlogger(), &msg, gorm.ErrNotImplemented)

	var appErr *cryptoutilSharedApperr.Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, http.StatusNotImplemented, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestMapGormError_SQLiteUniqueConstraint(t *testing.T) {
	t.Parallel()

	db := newSQLiteDBWithUniqueConstraint(t)

	// Insert a row, then insert a duplicate to trigger a real SQLite unique constraint error.
	require.NoError(t, db.Create(&testModel{ID: "1", Name: "alice"}).Error)

	sqliteErr := db.Create(&testModel{ID: "2", Name: "alice"}).Error
	require.Error(t, sqliteErr, "duplicate name must trigger a SQLite error")

	msg := "sqlite unique constraint"
	err := cryptoutilSharedApperr.MapGormError(discardSlogger(), &msg, sqliteErr)

	var appErr *cryptoutilSharedApperr.Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, http.StatusBadRequest, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestMapGormError_PGCodeUniqueViolation(t *testing.T) {
	t.Parallel()

	pgErr := &pgconn.PgError{Code: cryptoutilSharedMagic.PGCodeUniqueViolation}

	msg := "pg unique violation"
	err := cryptoutilSharedApperr.MapGormError(discardSlogger(), &msg, pgErr)

	var appErr *cryptoutilSharedApperr.Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, http.StatusBadRequest, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestMapGormError_PGCodeForeignKeyViolation(t *testing.T) {
	t.Parallel()

	pgErr := &pgconn.PgError{Code: cryptoutilSharedMagic.PGCodeForeignKeyViolation}

	msg := "pg foreign key violation"
	err := cryptoutilSharedApperr.MapGormError(discardSlogger(), &msg, pgErr)

	var appErr *cryptoutilSharedApperr.Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, http.StatusBadRequest, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}

func TestMapGormError_UnknownError_HTTP500(t *testing.T) {
	t.Parallel()

	pgErr := &pgconn.PgError{Code: "99999"} // Unknown PG code.

	msg := "unknown error"
	err := cryptoutilSharedApperr.MapGormError(discardSlogger(), &msg, pgErr)

	var appErr *cryptoutilSharedApperr.Error
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, http.StatusInternalServerError, int(appErr.HTTPStatusLineAndCode.StatusLine.StatusCode))
}
