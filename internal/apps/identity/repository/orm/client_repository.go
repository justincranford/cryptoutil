// Copyright (c) 2025 Justin Cranford
//
//

package orm

import (
	"context"
	crand "crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilIdentityAppErr "cryptoutil/internal/apps/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/apps/identity/domain"
	cryptoutilSharedCryptoHash "cryptoutil/internal/shared/crypto/hash"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ClientRepositoryGORM implements the ClientRepository interface using GORM.
type ClientRepositoryGORM struct {
	db *gorm.DB
}

// NewClientRepository creates a new ClientRepositoryGORM.
func NewClientRepository(db *gorm.DB) *ClientRepositoryGORM {
	return &ClientRepositoryGORM{db: db}
}

// Create creates a new client with an initial secret (version 1).
func (r *ClientRepositoryGORM) Create(ctx context.Context, client *cryptoutilIdentityDomain.Client) error {
	err := getDB(ctx, r.db).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Create client record.
		if err := tx.Create(client).Error; err != nil {
			return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to create client: %w", err))
		}

		// 2. Use provided client secret hash if available.
		// Tests typically pre-hash secrets before calling Create().
		secretHash := client.ClientSecret
		if secretHash == "" {
			// Generate and hash new secret if none provided.
			initialSecret, err := generateRandomSecret(cryptoutilSharedMagic.SecretGenerationDefaultByteLength)
			if err != nil {
				return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrKeyGenerationFailed, fmt.Errorf("failed to generate initial secret: %w", err))
			}

			secretHash, err = cryptoutilSharedCryptoHash.HashLowEntropyNonDeterministic(initialSecret)
			if err != nil {
				return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrPasswordHashFailed, fmt.Errorf("failed to hash initial secret: %w", err))
			}
		}

		// 4. Store ClientSecretVersion (version 1, active, no expiration).
		now := time.Now().UTC()

		version := &cryptoutilIdentityDomain.ClientSecretVersion{
			ID:         googleUuid.New(),
			ClientID:   client.ID,
			Version:    1,
			SecretHash: secretHash,
			Status:     cryptoutilIdentityDomain.SecretStatusActive,
			CreatedAt:  now,
			ExpiresAt:  nil,
		}
		if err := tx.Create(version).Error; err != nil {
			return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to create initial secret version: %w", err))
		}

		// 5. Create KeyRotationEvent audit log.
		oldVersion := 0
		newVersion := 1
		success := true

		event := &cryptoutilIdentityDomain.KeyRotationEvent{
			ID:            googleUuid.New(),
			EventType:     "secret_created",
			KeyType:       "client_secret",
			KeyID:         client.ID.String(),
			Timestamp:     now,
			Initiator:     "system",
			OldKeyVersion: &oldVersion,
			NewKeyVersion: &newVersion,
			Reason:        "Initial client creation",
			Success:       &success,
		}
		if err := tx.Create(event).Error; err != nil {
			return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to create audit event: %w", err))
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
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

// GetByID retrieves a client by ID.
func (r *ClientRepositoryGORM) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.Client, error) {
	var client cryptoutilIdentityDomain.Client
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrClientNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get client by ID: %w", err))
	}

	return &client, nil
}

// GetByClientID retrieves a client by OAuth client_id.
func (r *ClientRepositoryGORM) GetByClientID(ctx context.Context, clientID string) (*cryptoutilIdentityDomain.Client, error) {
	var client cryptoutilIdentityDomain.Client
	// Enable debug mode to see SQL queries.
	if err := getDB(ctx, r.db).Debug().WithContext(ctx).Where("client_id = ? AND deleted_at IS NULL", clientID).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrClientNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get client by client_id: %w", err))
	}

	return &client, nil
}

// GetAll retrieves all clients (for secret migration).
func (r *ClientRepositoryGORM) GetAll(ctx context.Context) ([]*cryptoutilIdentityDomain.Client, error) {
	var clients []*cryptoutilIdentityDomain.Client
	if err := getDB(ctx, r.db).WithContext(ctx).Where("deleted_at IS NULL").Find(&clients).Error; err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get all clients: %w", err))
	}

	return clients, nil
}

// Update updates an existing client.
func (r *ClientRepositoryGORM) Update(ctx context.Context, client *cryptoutilIdentityDomain.Client) error {
	client.UpdatedAt = time.Now().UTC()
	if err := getDB(ctx, r.db).WithContext(ctx).Save(client).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to update client: %w", err))
	}

	return nil
}

// Delete deletes a client by ID (soft delete).
func (r *ClientRepositoryGORM) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Where("id = ?", id).Delete(&cryptoutilIdentityDomain.Client{}).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to delete client: %w", err))
	}

	return nil
}

// List lists clients with pagination.
func (r *ClientRepositoryGORM) List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.Client, error) {
	var clients []*cryptoutilIdentityDomain.Client
	if err := getDB(ctx, r.db).WithContext(ctx).Where("deleted_at IS NULL").Offset(offset).Limit(limit).Find(&clients).Error; err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to list clients: %w", err))
	}

	return clients, nil
}

// Count returns the total number of clients.
func (r *ClientRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := getDB(ctx, r.db).WithContext(ctx).Model(&cryptoutilIdentityDomain.Client{}).Where("deleted_at IS NULL").Count(&count).Error; err != nil {
		return 0, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to count clients: %w", err))
	}

	return count, nil
}

// RotateSecret rotates client secret and archives old secret in history.
func (r *ClientRepositoryGORM) RotateSecret(_ context.Context, clientID googleUuid.UUID, newSecretHash string, rotatedBy string, reason string) error {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Fetch current client.
		var client cryptoutilIdentityDomain.Client
		if err := tx.Where("id = ? AND deleted_at IS NULL", clientID).First(&client).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return cryptoutilIdentityAppErr.ErrClientNotFound
			}

			return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to find client: %w", err))
		}

		// 2. Archive old secret to history.
		historyID, err := googleUuid.NewV7()
		if err != nil {
			return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to generate history ID: %w", err))
		}

		history := cryptoutilIdentityDomain.ClientSecretHistory{
			ID:         historyID,
			ClientID:   clientID,
			SecretHash: client.ClientSecret,
			RotatedAt:  time.Now().UTC(),
			RotatedBy:  rotatedBy,
			Reason:     reason,
			ExpiresAt:  nil,
			CreatedAt:  time.Now().UTC(),
			UpdatedAt:  time.Now().UTC(),
		}
		if err := tx.Create(&history).Error; err != nil {
			return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to create secret history: %w", err))
		}

		// 3. Update client with new secret hash.
		if err := tx.Model(&client).Update("client_secret", newSecretHash).Error; err != nil {
			return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to update client secret: %w", err))
		}

		return nil
	})
	if err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to rotate client secret: %w", err))
	}

	return nil
}

// GetSecretHistory retrieves secret rotation history for a client.
func (r *ClientRepositoryGORM) GetSecretHistory(ctx context.Context, clientID googleUuid.UUID) ([]cryptoutilIdentityDomain.ClientSecretHistory, error) {
	var history []cryptoutilIdentityDomain.ClientSecretHistory
	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("client_id = ?", clientID).
		Order("rotated_at DESC, id DESC").
		Find(&history).Error; err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get secret history: %w", err))
	}

	return history, nil
}
