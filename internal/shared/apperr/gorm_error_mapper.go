// Copyright (c) 2025 Justin Cranford
//
//

package apperr

import (
	"errors"
	"fmt"
	"log/slog"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gorm.io/gorm"

	"github.com/jackc/pgx/v5/pgconn"
	"modernc.org/sqlite"
)

// MapGormError maps a database error to an HTTP application error.
// Priority: custom app errors → GORM sentinel errors → SQLite errors → PostgreSQL errors → HTTP 500.
func MapGormError(slogger *slog.Logger, msg *string, err error) error {
	slogger.Error(*msg, cryptoutilSharedMagic.StringError, err)

	switch {
	case IsAppErr(err):
		return NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
	case errors.Is(err, gorm.ErrRecordNotFound):
		return NewHTTP404NotFound(msg, fmt.Errorf("%s: %w", *msg, err))
	case errors.Is(err, gorm.ErrDuplicatedKey),
		errors.Is(err, gorm.ErrForeignKeyViolated),
		errors.Is(err, gorm.ErrCheckConstraintViolated),
		errors.Is(err, gorm.ErrInvalidData),
		errors.Is(err, gorm.ErrInvalidValueOfLength):
		return NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
	case errors.Is(err, gorm.ErrNotImplemented):
		return NewHTTP501StatusLineAndCodeNotImplemented(msg, fmt.Errorf("%s: %w", *msg, err))
	}

	// SQLite-specific errors.
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		switch sqliteErr.Code() {
		case cryptoutilSharedMagic.SQLiteErrUniqueConstraint,
			cryptoutilSharedMagic.SQLiteErrForeignKey,
			cryptoutilSharedMagic.SQLiteErrCheckConstraint:
			return NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
		}
	}

	// PostgreSQL-specific errors.
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case cryptoutilSharedMagic.PGCodeUniqueViolation,
			cryptoutilSharedMagic.PGCodeForeignKeyViolation:
			return NewHTTP400BadRequest(msg, fmt.Errorf("%s: %w", *msg, err))
		}
	}

	return NewHTTP500InternalServerError(msg, fmt.Errorf("%s: %w", *msg, err))
}
