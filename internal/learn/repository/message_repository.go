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

// MessageRepository handles database operations for Message entities.
type MessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new MessageRepository.
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create inserts a new message with receivers into the database.
func (r *MessageRepository) Create(ctx context.Context, message *domain.Message) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Create(message).Error; err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	return nil
}

// FindByID retrieves a message by ID with its receivers.
func (r *MessageRepository) FindByID(ctx context.Context, id googleUuid.UUID) (*domain.Message, error) {
	var message domain.Message
	if err := getDB(ctx, r.db).WithContext(ctx).Preload("Receivers").First(&message, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("failed to find message: %w", err)
	}

	return &message, nil
}

// FindByReceiverID retrieves all messages for a specific receiver.
func (r *MessageRepository) FindByReceiverID(ctx context.Context, receiverID googleUuid.UUID) ([]domain.Message, error) {
	var messages []domain.Message

	if err := getDB(ctx, r.db).WithContext(ctx).
		Joins("JOIN message_receivers ON messages.id = message_receivers.message_id").
		Where("message_receivers.receiver_id = ?", receiverID).
		Preload("Receivers").
		Find(&messages).Error; err != nil {
		return nil, fmt.Errorf("failed to find messages by receiver: %w", err)
	}

	return messages, nil
}

// MarkAsReceived updates the received timestamp for a message receiver.
func (r *MessageRepository) MarkAsReceived(ctx context.Context, messageID, receiverID googleUuid.UUID) error {
	if err := getDB(ctx, r.db).WithContext(ctx).
		Model(&domain.MessageReceiver{}).
		Where("message_id = ? AND receiver_id = ?", messageID, receiverID).
		Update("received_at", gorm.Expr("CURRENT_TIMESTAMP")).Error; err != nil {
		return fmt.Errorf("failed to mark message as received: %w", err)
	}

	return nil
}

// Delete removes a message and its receivers from the database.
func (r *MessageRepository) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := getDB(ctx, r.db).WithContext(ctx).Delete(&domain.Message{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}
