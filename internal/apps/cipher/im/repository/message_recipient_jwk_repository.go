// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	"cryptoutil/internal/apps/cipher/im/domain"
	cryptoutilBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// MessageRecipientJWKRepository handles database operations for MessageRecipientJWK entities.
type MessageRecipientJWKRepository struct {
	db             *gorm.DB
	barrierService *cryptoutilBarrier.Service
}

// NewMessageRecipientJWKRepository creates a new MessageRecipientJWKRepository.
func NewMessageRecipientJWKRepository(db *gorm.DB, barrierService *cryptoutilBarrier.Service) *MessageRecipientJWKRepository {
	return &MessageRecipientJWKRepository{
		db:             db,
		barrierService: barrierService,
	}
}

// Create inserts a new message recipient JWK into the database.
// The JWK is already encrypted by the handler layer.
func (r *MessageRecipientJWKRepository) Create(ctx context.Context, messageRecipientJWK *domain.MessageRecipientJWK) error {
	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).Create(messageRecipientJWK).Error; err != nil {
		return fmt.Errorf("failed to create message recipient JWK: %w", err)
	}

	return nil
}

// FindByRecipientAndMessage retrieves a message recipient JWK by recipient ID and message ID.
// The JWK is already encrypted; decryption is handled by the handler layer.
func (r *MessageRecipientJWKRepository) FindByRecipientAndMessage(ctx context.Context, recipientID, messageID googleUuid.UUID) (*domain.MessageRecipientJWK, error) {
	var messageRecipientJWK domain.MessageRecipientJWK

	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).
		Where("recipient_id = ? AND message_id = ?", recipientID, messageID).
		First(&messageRecipientJWK).Error; err != nil {
		return nil, fmt.Errorf("failed to find message recipient JWK: %w", err)
	}

	return &messageRecipientJWK, nil
}

// FindByMessageID retrieves all recipient JWKs for a specific message.
// The JWKs are already encrypted; decryption is handled by the handler layer.
func (r *MessageRecipientJWKRepository) FindByMessageID(ctx context.Context, messageID googleUuid.UUID) ([]domain.MessageRecipientJWK, error) {
	var messageRecipientJWKs []domain.MessageRecipientJWK

	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).
		Where("message_id = ?", messageID).
		Find(&messageRecipientJWKs).Error; err != nil {
		return nil, fmt.Errorf("failed to find message recipient JWKs: %w", err)
	}

	return messageRecipientJWKs, nil
}

// Delete removes a message recipient JWK from the database.
func (r *MessageRecipientJWKRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).Delete(&domain.MessageRecipientJWK{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete message recipient JWK: %w", err)
	}

	return nil
}

// DeleteByMessageID removes all recipient JWKs for a specific message.
func (r *MessageRecipientJWKRepository) DeleteByMessageID(ctx context.Context, messageID googleUuid.UUID) error {
	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).
		Where("message_id = ?", messageID).
		Delete(&domain.MessageRecipientJWK{}).Error; err != nil {
		return fmt.Errorf("failed to delete message recipient JWKs: %w", err)
	}

	return nil
}
