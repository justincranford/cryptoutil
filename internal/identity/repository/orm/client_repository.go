package orm

import (
	"context"
	"errors"
	"fmt"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilIdentityAppErr "cryptoutil/internal/identity/apperr"
	cryptoutilIdentityDomain "cryptoutil/internal/identity/domain"
)

// ClientRepositoryGORM implements the ClientRepository interface using GORM.
type ClientRepositoryGORM struct {
	db *gorm.DB
}

// NewClientRepository creates a new ClientRepositoryGORM.
func NewClientRepository(db *gorm.DB) *ClientRepositoryGORM {
	return &ClientRepositoryGORM{db: db}
}

// Create creates a new client.
func (r *ClientRepositoryGORM) Create(ctx context.Context, client *cryptoutilIdentityDomain.Client) error {
	if err := r.db.WithContext(ctx).Create(client).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to create client: %w", err))
	}

	return nil
}

// GetByID retrieves a client by ID.
func (r *ClientRepositoryGORM) GetByID(ctx context.Context, id googleUuid.UUID) (*cryptoutilIdentityDomain.Client, error) {
	var client cryptoutilIdentityDomain.Client
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&client).Error; err != nil {
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
	if err := r.db.WithContext(ctx).Where("client_id = ? AND deleted_at IS NULL", clientID).First(&client).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, cryptoutilIdentityAppErr.ErrClientNotFound
		}

		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to get client by client_id: %w", err))
	}

	return &client, nil
}

// Update updates an existing client.
func (r *ClientRepositoryGORM) Update(ctx context.Context, client *cryptoutilIdentityDomain.Client) error {
	client.UpdatedAt = time.Now()
	if err := r.db.WithContext(ctx).Save(client).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to update client: %w", err))
	}

	return nil
}

// Delete deletes a client by ID (soft delete).
func (r *ClientRepositoryGORM) Delete(ctx context.Context, id googleUuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&cryptoutilIdentityDomain.Client{}).Error; err != nil {
		return cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to delete client: %w", err))
	}

	return nil
}

// List lists clients with pagination.
func (r *ClientRepositoryGORM) List(ctx context.Context, offset, limit int) ([]*cryptoutilIdentityDomain.Client, error) {
	var clients []*cryptoutilIdentityDomain.Client
	if err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Offset(offset).Limit(limit).Find(&clients).Error; err != nil {
		return nil, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to list clients: %w", err))
	}

	return clients, nil
}

// Count returns the total number of clients.
func (r *ClientRepositoryGORM) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&cryptoutilIdentityDomain.Client{}).Where("deleted_at IS NULL").Count(&count).Error; err != nil {
		return 0, cryptoutilIdentityAppErr.WrapError(cryptoutilIdentityAppErr.ErrDatabaseQuery, fmt.Errorf("failed to count clients: %w", err))
	}

	return count, nil
}
