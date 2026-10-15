// Copyright (c) 2025-2026 Justin Cranford.
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gorm.io/gorm/clause"
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

	for attempt := 0; attempt < crlSequenceAllocationAttempts; attempt++ {
		sequence := StoredCRLSequence{IssuerKeyID: issuerKeyID, Number: 0}

		err := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&sequence).Error
		if err != nil {
			if retryableCRLSequenceError(err, s.db.Name()) {
				if retryErr := waitForCRLSequenceRetry(ctx, attempt); retryErr != nil {
					return nil, retryErr
				}

				continue
			}

			return nil, fmt.Errorf("failed to initialize issuer CRL sequence: %w", err)
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
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("issuer CRL sequence %q is exhausted or missing", issuerKeyID)
			}

			if retryableCRLSequenceError(err, s.db.Name()) {
				if retryErr := waitForCRLSequenceRetry(ctx, attempt); retryErr != nil {
					return nil, retryErr
				}

				continue
			}

			return nil, fmt.Errorf("failed to allocate issuer CRL number: %w", err)
		}

		if number <= 0 {
			return nil, fmt.Errorf("issuer CRL sequence %q returned invalid number %d", issuerKeyID, number)
		}

		return big.NewInt(number), nil
	}

	return nil, fmt.Errorf("failed to allocate issuer CRL number after %d attempts: database remained busy", crlSequenceAllocationAttempts)
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

	message := strings.ToLower(err.Error())
	if dialect == "sqlite" {
		return strings.Contains(message, "database is locked") ||
			strings.Contains(message, "database table is locked") ||
			strings.Contains(message, "database is busy")
	}

	return strings.Contains(message, "deadlock detected") ||
		strings.Contains(message, "could not serialize access") ||
		strings.Contains(message, "serialization failure")
}
