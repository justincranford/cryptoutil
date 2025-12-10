// Copyright (c) 2025 Iwan van der Kleijn
// SPDX-License-Identifier: MIT

package mfa

import (
	"context"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
	cryptoutilMagic "cryptoutil/internal/identity/magic"
	cryptoutilIdentityRepository "cryptoutil/internal/identity/repository"
)

// RecoveryCodeService manages recovery code operations.
type RecoveryCodeService struct {
	repo cryptoutilIdentityRepository.RecoveryCodeRepository
}

// NewRecoveryCodeService creates a new recovery code service.
func NewRecoveryCodeService(repo cryptoutilIdentityRepository.RecoveryCodeRepository) *RecoveryCodeService {
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
	expiresAt := time.Now().UTC().Add(cryptoutilMagic.DefaultRecoveryCodeLifetime)

	for i, plaintext := range plaintextCodes {
		// Hash code with bcrypt (cost 10 = default).
		hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
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

		// Compare plaintext with bcrypt hash.
		if err := bcrypt.CompareHashAndPassword([]byte(code.CodeHash), []byte(plaintext)); err == nil {
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
