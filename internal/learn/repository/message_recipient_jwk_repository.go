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
)

// MessageRecipientJWKRepository handles database operations for MessageRecipientJWK entities.
type MessageRecipientJWKRepository struct {
	db *gorm.DB
}

// NewMessageRecipientJWKRepository creates a new MessageRecipientJWKRepository.
func NewMessageRecipientJWKRepository(db *gorm.DB) *MessageRecipientJWKRepository {
	return &MessageRecipientJWKRepository{db: db}
}

// Create inserts a new message recipient JWK into the database.
func (r *MessageRecipientJWKRepository) Create(ctx context.Context, jwk *domain.MessageRecipientJWK) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(jwk).Error; err != nil {
		return fmt.Errorf("failed to create message recipient JWK: %w", err)
	}

	return nil
}

// FindByRecipientAndMessage retrieves a message recipient JWK by recipient ID and message ID.
func (r *MessageRecipientJWKRepository) FindByRecipientAndMessage(ctx context.Context, recipientID, messageID googleUuid.UUID) (*domain.MessageRecipientJWK, error) {
	var jwk domain.MessageRecipientJWK

	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("recipient_id = ? AND message_id = ?", recipientID, messageID).
		First(&jwk).Error; err != nil {
		return nil, fmt.Errorf("failed to find message recipient JWK: %w", err)
	}

	return &jwk, nil
}

// FindByMessageID retrieves all recipient JWKs for a specific message.
func (r *MessageRecipientJWKRepository) FindByMessageID(ctx context.Context, messageID googleUuid.UUID) ([]domain.MessageRecipientJWK, error) {
	var jwks []domain.MessageRecipientJWK

	if err := getDB(ctx, r.db).WithContext(ctx).
		Where("message_id = ?", messageID).
		Find(&jwks).Error; err != nil {
		return nil, fmt.Errorf("failed to find message recipient JWKs: %w", err)
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
