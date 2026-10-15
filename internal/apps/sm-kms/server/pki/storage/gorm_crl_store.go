// Copyright (c) 2025-2026 Justin Cranford.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm/clause"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	crlSequenceAllocationAttempts = 12
	crlSequenceRetryBaseDelay     = 5 * time.Millisecond
)

func (s *GORMStore) GetRevokedByIssuerKeyID(ctx context.Context, issuerKeyID string) ([]*StoredCertificate, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is required")
	}

	issuerKeyID = strings.ToLower(strings.TrimSpace(issuerKeyID))
	if issuerKeyID == "" {
		return nil, ErrInvalidIssuerKeyID
	}

	var certificates []*StoredCertificate
	if err := s.db.WithContext(ctx).
		Where("authority_key_id = ? AND status = ?", issuerKeyID, StatusRevoked).
		Order("revocation_time ASC").
		Find(&certificates).Error; err != nil {
		return nil, fmt.Errorf("failed to get revoked certificates by issuer key ID: %w", err)
	}

	return certificates, nil
}

// AllocateCRLNumber atomically allocates a positive, monotonically increasing CRL number
// for an issuer key ID. The shared database stores a signed BIGINT sequence.
func (s *GORMStore) AllocateCRLNumber(ctx context.Context, issuerKeyID string) (*big.Int, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context is required")
	}

	issuerKeyID = strings.ToLower(strings.TrimSpace(issuerKeyID))
	if issuerKeyID == "" {
		return nil, ErrInvalidIssuerKeyID
	}

	number, err := allocateCRLNumberWithRetry(
		ctx,
		s.db.Name(),
		func() (int64, error) {
			return s.allocateCRLNumberOnce(ctx, issuerKeyID)
		},
		waitForCRLSequenceRetry,
	)
	if err != nil {
		return nil, err
	}

	return big.NewInt(number), nil
}

func (s *GORMStore) allocateCRLNumberOnce(ctx context.Context, issuerKeyID string) (int64, error) {
	sequence := StoredCRLSequence{IssuerKeyID: issuerKeyID, Number: 0}

	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&sequence).Error; err != nil {
		return 0, fmt.Errorf("failed to initialize issuer CRL sequence: %w", err)
	}

	// A single UPDATE ... RETURNING statement is the atomic increment and fetch.
	// SQLite and PostgreSQL serialize concurrent updates to this issuer row.
	var number int64

	row := s.db.WithContext(ctx).Raw(
		"UPDATE "+(StoredCRLSequence{}).TableName()+" SET number = number + 1, updated_at = ? "+
			"WHERE issuer_key_id = ? AND number < ? RETURNING number",
		time.Now().UTC(),
		issuerKeyID,
		int64(^uint64(0)>>1),
	).Row()
	if err := row.Scan(&number); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("issuer CRL sequence %q is exhausted or missing", issuerKeyID)
		}

		return 0, fmt.Errorf("failed to allocate issuer CRL number: %w", err)
	}

	if number <= 0 {
		return 0, fmt.Errorf("issuer CRL sequence %q returned invalid number %d", issuerKeyID, number)
	}

	return number, nil
}

func allocateCRLNumberWithRetry(
	ctx context.Context,
	dialect string,
	allocateFn func() (int64, error),
	waitFn func(context.Context, int) error,
) (int64, error) {
	var lastErr error

	for attempt := 0; attempt < crlSequenceAllocationAttempts; attempt++ {
		number, err := allocateFn()
		if err == nil {
			return number, nil
		}

		if !retryableCRLSequenceError(err, dialect) {
			return 0, err
		}

		lastErr = err

		if attempt+1 == crlSequenceAllocationAttempts {
			break
		}

		if retryErr := waitFn(ctx, attempt); retryErr != nil {
			return 0, retryErr
		}
	}

	return 0, fmt.Errorf("failed to allocate issuer CRL number after %d attempts: %w", crlSequenceAllocationAttempts, lastErr)
}

func waitForCRLSequenceRetry(ctx context.Context, attempt int) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("failed to allocate issuer CRL number: %w", err)
	}

	delay := time.Duration(attempt+1) * crlSequenceRetryBaseDelay

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return fmt.Errorf("failed to allocate issuer CRL number: %w", ctx.Err())
	case <-timer.C:
		return nil
	}
}

func retryableCRLSequenceError(err error, dialect string) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if dialect == cryptoutilSharedMagic.DockerServicePostgres && errors.As(err, &pgErr) &&
		(pgErr.Code == cryptoutilSharedMagic.PGCodeDeadlockDetected ||
			pgErr.Code == cryptoutilSharedMagic.PGCodeSerializationFailure) {
		return true
	}

	message := strings.ToLower(err.Error())
	if dialect == cryptoutilSharedMagic.TestDatabaseSQLite {
		return strings.Contains(message, "database is locked") ||
			strings.Contains(message, "database table is locked") ||
			strings.Contains(message, "database is busy")
	}

	return dialect == cryptoutilSharedMagic.DockerServicePostgres && (strings.Contains(message, "deadlock detected") ||
		strings.Contains(message, "could not serialize access") ||
		strings.Contains(message, "serialization failure"))
}
