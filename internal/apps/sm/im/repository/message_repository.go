// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"fmt"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilAppsSmImDomain "cryptoutil/internal/apps/sm/im/domain"
	cryptoutilAppsTemplateServiceServerRepository "cryptoutil/internal/apps/template/service/server/repository"
)

// MessageRepository handles database operations for Message entities.
type MessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new MessageRepository.
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create inserts a new message into the database.
func (r *MessageRepository) Create(ctx context.Context, message *cryptoutilAppsSmImDomain.Message) error {
	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).Create(message).Error; err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// FindByID retrieves a message by ID.
func (r *MessageRepository) FindByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilAppsSmImDomain.Message, error) {
	var message cryptoutilAppsSmImDomain.Message
	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).First(&message, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to find message: %w", err)
	}

	return &message, nil
}

// FindByRecipientID retrieves all messages for a specific recipient using JOIN with messages_recipient_jwks.
// Uses 3-table schema: messages + messages_recipient_jwks + users.
func (r *MessageRepository) FindByRecipientID(ctx context.Context, recipientID googleUuid.UUID) ([]cryptoutilAppsSmImDomain.Message, error) {
	var messages []cryptoutilAppsSmImDomain.Message

	// JOIN messages with messages_recipient_jwks to find messages for this recipient.
	// Preload Sender to populate sender information for response.
	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).
		Joins("JOIN messages_recipient_jwks ON messages.id = messages_recipient_jwks.message_id").
		Where("messages_recipient_jwks.recipient_id = ?", recipientID).
		Preload("Sender").
		Order("messages.created_at DESC").
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to find messages by recipient: %w", err)
	}

	return messages, nil
}

// MarkAsRead updates the read timestamp for a message.
func (r *MessageRepository) MarkAsRead(ctx context.Context, messageID googleUuid.UUID) error {
	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).
		Model(&cryptoutilAppsSmImDomain.Message{}).
		Where("id = ?", messageID).
		Update("read_at", gorm.Expr("CURRENT_TIMESTAMP")).Error; err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	return nil
}

// Delete removes a message from the database.
func (r *MessageRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := cryptoutilAppsTemplateServiceServerRepository.GetDB(ctx, r.db).WithContext(ctx).Delete(&cryptoutilAppsSmImDomain.Message{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}
