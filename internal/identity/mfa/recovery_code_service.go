// Copyright (c) 2025 Justin Cranford

package mfa

import (
	"context"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilIdentityMagic "cryptoutil/internal/identity/magic"
	cryptoutilSharedCryptoPassword "cryptoutil/internal/shared/crypto/password"
)

// RecoveryCodeRepository defines minimal repository interface.
type RecoveryCodeRepository interface {
	Create(ctx context.Context, code *cryptoutilIdentityDomain.RecoveryCode) error
	CreateBatch(ctx context.Context, codes []*cryptoutilIdentityDomain.RecoveryCode) error
	GetByUserID(ctx context.Context, userID googleUuid.UUID) ([]*cryptoutilIdentityDomain.RecoveryCode, error)
	Update(ctx context.Context, code *cryptoutilIdentityDomain.RecoveryCode) error
	DeleteByUserID(ctx context.Context, userID googleUuid.UUID) error
	CountUnused(ctx context.Context, userID googleUuid.UUID) (int64, error)
}

// RecoveryCodeService manages recovery code operations.
type RecoveryCodeService struct {
	repo RecoveryCodeRepository
}

// NewRecoveryCodeService creates a new recovery code service.
func NewRecoveryCodeService(repo RecoveryCodeRepository) *RecoveryCodeService {
	return &RecoveryCodeService{repo: repo}
}

// GenerateForUser generates a new batch of recovery codes for a user.
// Returns plaintext codes (shown once to user) and stores hashed versions.
func (s *RecoveryCodeService) GenerateForUser(ctx context.Context, userID googleUuid.UUID, count int) ([]string, error) {
	// Generate plaintext codes.
	plaintextCodes, err := GenerateRecoveryCodes(count)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recovery codes: %w", err)
	}

	// Hash codes and create domain models.
	codes := make([]*cryptoutilIdentityDomain.RecoveryCode, count)
	expiresAt := time.Now().UTC().Add(cryptoutilIdentityMagic.DefaultRecoveryCodeLifetime)

	for i, plaintext := range plaintextCodes {
		// Hash code with PBKDF2 (FIPS-compliant).
		hash, err := cryptoutilSharedCryptoPassword.HashPassword(plaintext)
		if err != nil {
			return nil, fmt.Errorf("failed to hash recovery code: %w", err)
		}

		codes[i] = &cryptoutilIdentityDomain.RecoveryCode{
			ID:        googleUuid.New(),
			UserID:    userID,
			CodeHash:  string(hash),
			Used:      false,
			UsedAt:    nil,
			CreatedAt: time.Now().UTC(),
			ExpiresAt: expiresAt,
		}
	}

	// Store hashed codes in database.
	if err := s.repo.CreateBatch(ctx, codes); err != nil {
		return nil, fmt.Errorf("failed to store recovery codes: %w", err)
	}

	return plaintextCodes, nil
}

// Verify checks if a recovery code is valid and marks it as used.
func (s *RecoveryCodeService) Verify(ctx context.Context, userID googleUuid.UUID, plaintext string) error {
	// Get all codes for user.
	codes, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get recovery codes: %w", err)
	}

	// Find matching unused, unexpired code.
	for _, code := range codes {
		if code.IsUsed() || code.IsExpired() {
			continue
		}

		// Compare plaintext with hash (PBKDF2-HMAC-SHA256, FIPS-compliant).
		match, _, err := cryptoutilSharedCryptoPassword.VerifyPassword(plaintext, code.CodeHash)
		if err == nil && match {
			// Code matches - mark as used.
			code.MarkAsUsed()

			if err := s.repo.Update(ctx, code); err != nil {
				return fmt.Errorf("failed to mark recovery code as used: %w", err)
			}

			return nil
		}
	}

	// No matching code found.
	return cryptoutilIdentityAppErr.ErrRecoveryCodeNotFound
}

// RegenerateForUser deletes old codes and generates new batch.
func (s *RecoveryCodeService) RegenerateForUser(ctx context.Context, userID googleUuid.UUID, count int) ([]string, error) {
	// Delete all existing codes for user.
	if err := s.repo.DeleteByUserID(ctx, userID); err != nil {
		return nil, fmt.Errorf("failed to delete old recovery codes: %w", err)
	}

	// Generate new codes.
	return s.GenerateForUser(ctx, userID, count)
}

// GetRemainingCount returns count of unused, unexpired codes for a user.
func (s *RecoveryCodeService) GetRemainingCount(ctx context.Context, userID googleUuid.UUID) (int64, error) {
	count, err := s.repo.CountUnused(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count unused recovery codes: %w", err)
	}

	return count, nil
}
