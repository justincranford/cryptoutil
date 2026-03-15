// Copyright (c) 2025 Justin Cranford
//
//

// Package rotation provides secret rotation services for identity components.
package rotation

import (
	"context"
	crand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedCryptoDigests "cryptoutil/internal/shared/crypto/digests"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ErrNoActiveSecretVersion indicates no active secret version exists for a client.
var ErrNoActiveSecretVersion = errors.New("no active secret version found")

// SecretRotationService handles client secret rotation operations.
type SecretRotationService struct {
	db               *gorm.DB
	generateSecretFn func(int) (string, error)
	hashSecretFn     func(string) (string, error)
}

// NewSecretRotationService creates a new secret rotation service.
func NewSecretRotationService(db *gorm.DB) *SecretRotationService {
	return &SecretRotationService{
		db:               db,
		generateSecretFn: generateRandomSecret,
		hashSecretFn:     cryptoutilSharedCryptoHash.HashLowEntropyNonDeterministic,
	}
}

// RotateClientSecretResult contains the results of a secret rotation operation.
type RotateClientSecretResult struct {
	OldVersion         int
	NewVersion         int
	NewSecretPlaintext string
	GracePeriodEnd     time.Time
	EventID            googleUuid.UUID
}

// RotateClientSecret rotates a client secret with grace period support.
func (s *SecretRotationService) RotateClientSecret(
	ctx context.Context,
	clientID googleUuid.UUID,
	gracePeriodDuration time.Duration,
	initiator string,
	reason string,
) (*RotateClientSecretResult, error) {
	var result RotateClientSecretResult

	// Generate new secret.
	newSecretPlaintext, err := s.generateSecretFn(cryptoutilSharedMagic.SecretGenerationDefaultByteLength)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new secret: %w", err)
	}

	// Hash new secret.
	newSecretHash, err := s.hashSecretFn(newSecretPlaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to hash new secret: %w", err)
	} // Execute rotation in transaction.

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get current active secret version.
		var currentVersion cryptoutilIdentityDomain.ClientSecretVersion

		err := tx.Where("client_id = ? AND status = ?", clientID, cryptoutilIdentityDomain.SecretStatusActive).
			Order("version DESC").
			First(&currentVersion).Error

		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to query current version: %w", err)
		}

		// Determine new version number.
		newVersionNum := 1
		if err == nil {
			// Found existing version.
			newVersionNum = currentVersion.Version + 1
			result.OldVersion = currentVersion.Version

			// Mark old version as expired (after grace period).
			expiresAt := time.Now().UTC().Add(gracePeriodDuration)
			currentVersion.ExpiresAt = &expiresAt

			if updateErr := tx.Save(&currentVersion).Error; updateErr != nil {
				return fmt.Errorf("failed to update old version expiration: %w", updateErr)
			}
		}

		// Create new secret version.
		newVersion := &cryptoutilIdentityDomain.ClientSecretVersion{
			ClientID:   clientID,
			Version:    newVersionNum,
			SecretHash: newSecretHash,
			Status:     cryptoutilIdentityDomain.SecretStatusActive,
		}

		if createErr := tx.Create(newVersion).Error; createErr != nil {
			return fmt.Errorf("failed to create new version: %w", createErr)
		}

		result.NewVersion = newVersionNum
		result.NewSecretPlaintext = newSecretPlaintext
		result.GracePeriodEnd = time.Now().UTC().Add(gracePeriodDuration)

		// Create rotation event.
		oldVersionPtr := &result.OldVersion
		newVersionPtr := &result.NewVersion

		if result.OldVersion == 0 {
			oldVersionPtr = nil
		}

		gracePeriodStr := gracePeriodDuration.String()

		event := &cryptoutilIdentityDomain.KeyRotationEvent{
			EventType:     cryptoutilIdentityDomain.EventTypeRotation,
			KeyType:       cryptoutilIdentityDomain.KeyTypeClientSecret,
			KeyID:         clientID.String(),
			Initiator:     initiator,
			OldKeyVersion: oldVersionPtr,
			NewKeyVersion: newVersionPtr,
			GracePeriod:   &gracePeriodStr,
			Reason:        reason,
			Success:       &[]bool{true}[0],
		}

		if eventErr := tx.Create(event).Error; eventErr != nil {
			return fmt.Errorf("failed to create rotation event: %w", eventErr)
		}

		result.EventID = event.ID

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("rotation transaction failed: %w", err)
	}

	return &result, nil
}

// GetActiveSecretVersion retrieves the current active secret version for a client.
func (s *SecretRotationService) GetActiveSecretVersion(
	ctx context.Context,
	clientID googleUuid.UUID,
) (*cryptoutilIdentityDomain.ClientSecretVersion, error) {
	var version cryptoutilIdentityDomain.ClientSecretVersion

	err := s.db.WithContext(ctx).
		Where("client_id = ? AND status = ?", clientID, cryptoutilIdentityDomain.SecretStatusActive).
		Order("version DESC").
		First(&version).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoActiveSecretVersion
		}

		return nil, fmt.Errorf("failed to query active version: %w", err)
	}

	return &version, nil
}

// ValidateSecretDuringGracePeriod validates a client secret against all active versions.
func (s *SecretRotationService) ValidateSecretDuringGracePeriod(
	ctx context.Context,
	clientID googleUuid.UUID,
	secretPlaintext string,
) (bool, int, error) {
	// Get all active versions (including expired but within grace period).
	activeVersions, err := s.GetActiveSecretVersions(ctx, clientID)
	if err != nil {
		return false, 0, fmt.Errorf("failed to query active versions: %w", err)
	}

	// Validate against all active versions.
	now := time.Now().UTC()

	for i := range activeVersions {
		if activeVersions[i].IsValid(now) {
			if compareSecret(secretPlaintext, activeVersions[i].SecretHash) {
				return true, activeVersions[i].Version, nil
			}
		}
	}

	return false, 0, nil
}

// GetActiveSecretVersions returns all active secret versions for grace period validation.
func (s *SecretRotationService) GetActiveSecretVersions(
	ctx context.Context,
	clientID googleUuid.UUID,
) ([]*cryptoutilIdentityDomain.ClientSecretVersion, error) {
	var versions []*cryptoutilIdentityDomain.ClientSecretVersion

	err := s.db.WithContext(ctx).
		Where("client_id = ? AND status = ?", clientID, cryptoutilIdentityDomain.SecretStatusActive).
		Order("version DESC").
		Find(&versions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query active secrets: %w", err)
	}

	return versions, nil
}

// RevokeSecretVersion revokes a specific secret version.
func (s *SecretRotationService) RevokeSecretVersion(
	ctx context.Context,
	clientID googleUuid.UUID,
	version int,
	revokerID string,
	reason string,
) error {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find and revoke the version.
		var secretVersion cryptoutilIdentityDomain.ClientSecretVersion

		err := tx.Where("client_id = ? AND version = ?", clientID, version).
			First(&secretVersion).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("secret version %d not found", version)
			}

			return fmt.Errorf("failed to query secret version: %w", err)
		}

		secretVersion.MarkRevoked(revokerID)

		if updateErr := tx.Save(&secretVersion).Error; updateErr != nil {
			return fmt.Errorf("failed to revoke secret version: %w", updateErr)
		}

		// Create revocation event.
		versionPtr := &version
		success := true

		event := &cryptoutilIdentityDomain.KeyRotationEvent{
			EventType:     cryptoutilIdentityDomain.EventTypeRevocation,
			KeyType:       cryptoutilIdentityDomain.KeyTypeClientSecret,
			KeyID:         clientID.String(),
			Initiator:     revokerID,
			OldKeyVersion: versionPtr,
			Reason:        reason,
			Success:       &success,
		}

		if eventErr := tx.Create(event).Error; eventErr != nil {
			return fmt.Errorf("failed to create revocation event: %w", eventErr)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to revoke secret version: %w", err)
	}

	return nil
}

// generateRandomSecret generates a cryptographically secure random secret.
func generateRandomSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := crand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}

// compareSecret compares a plaintext secret against a stored hash using PBKDF2.
func compareSecret(plaintext, hash string) bool {
	match, err := cryptoutilSharedCryptoDigests.VerifySecret(hash, plaintext)

	return err == nil && match
}
