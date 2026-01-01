// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	"cryptoutil/internal/learn/domain"
	cryptoutilBarrier "cryptoutil/internal/template/server/barrier"
)

// MessageRecipientJWKRepository handles database operations for MessageRecipientJWK entities.
type MessageRecipientJWKRepository struct {
	db             *gorm.DB
	barrierService *cryptoutilBarrier.BarrierService
}

// NewMessageRecipientJWKRepository creates a new MessageRecipientJWKRepository.
func NewMessageRecipientJWKRepository(db *gorm.DB, barrierService *cryptoutilBarrier.BarrierService) *MessageRecipientJWKRepository {
	return &MessageRecipientJWKRepository{
		db:             db,
		barrierService: barrierService,
	}
}

// Create inserts a new message recipient JWK into the database.
// The JWK is encrypted with barrier encryption before storing.
func (r *MessageRecipientJWKRepository) Create(ctx context.Context, jwk *domain.MessageRecipientJWK) error {
	// Encrypt the JWK with barrier encryption (adds second layer of protection)
	encryptedJWK, err := r.barrierService.EncryptContentWithContext(ctx, []byte(jwk.JWK))
	if err != nil {
		return fmt.Errorf("failed to encrypt JWK with barrier: %w", err)
	}

	// Create a copy with encrypted JWK for storage
	encryptedEntity := *jwk
	encryptedEntity.JWK = string(encryptedJWK)

	if err := getDB(ctx, r.db).WithContext(ctx).Create(&encryptedEntity).Error; err != nil {
		return fmt.Errorf("failed to create message recipient JWK: %w", err)
	}

	return nil
}

// FindByRecipientAndMessage retrieves a message recipient JWK by recipient ID and message ID.
// The JWK is decrypted with barrier encryption after retrieval.
func (r *MessageRecipientJWKRepository) FindByRecipientAndMessage(ctx context.Context, recipientID, messageID googleUuid.UUID) (*domain.MessageRecipientJWK, error) {
	var jwk domain.MessageRecipientJWK

	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("recipient_id = ? AND message_id = ?", recipientID, messageID).
		First(&jwk).Error; err != nil {
		return nil, fmt.Errorf("failed to find message recipient JWK: %w", err)
	}

	// Decrypt the JWK with barrier encryption
	decryptedJWK, err := r.barrierService.DecryptContentWithContext(ctx, []byte(jwk.JWK))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt JWK with barrier: %w", err)
	}

	jwk.JWK = string(decryptedJWK)

	return &jwk, nil
}

// FindByMessageID retrieves all recipient JWKs for a specific message.
// The JWKs are decrypted with barrier encryption after retrieval.
func (r *MessageRecipientJWKRepository) FindByMessageID(ctx context.Context, messageID googleUuid.UUID) ([]domain.MessageRecipientJWK, error) {
	var jwks []domain.MessageRecipientJWK

	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("message_id = ?", messageID).
		Find(&jwks).Error; err != nil {
		return nil, fmt.Errorf("failed to find message recipient JWKs: %w", err)
	}

	// Decrypt each JWK with barrier encryption
	for i := range jwks {
		decryptedJWK, err := r.barrierService.DecryptContentWithContext(ctx, []byte(jwks[i].JWK))
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt JWK with barrier: %w", err)
		}
		jwks[i].JWK = string(decryptedJWK)
	}

	return jwks, nil
}

// Delete removes a message recipient JWK from the database.
func (r *MessageRecipientJWKRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Delete(&domain.MessageRecipientJWK{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete message recipient JWK: %w", err)
	}

	return nil
}

// DeleteByMessageID removes all recipient JWKs for a specific message.
func (r *MessageRecipientJWKRepository) DeleteByMessageID(ctx context.Context, messageID googleUuid.UUID) error {
	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("message_id = ?", messageID).
		Delete(&domain.MessageRecipientJWK{}).Error; err != nil {
		return fmt.Errorf("failed to delete message recipient JWKs: %w", err)
	}

	return nil
}
