// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"context"
	"time"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository provides CRUD operations for verified users.
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id googleUuid.UUID) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	ListByTenant(ctx context.Context, tenantID googleUuid.UUID, activeOnly bool) ([]*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id googleUuid.UUID) error
}

// UserRepositoryImpl implements UserRepository using GORM.
type UserRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserRepositoryImpl{db: db}
}

// Create creates a new user.
func (r *UserRepositoryImpl) Create(ctx context.Context, user *User) error {
	return toAppErr(r.db.WithContext(ctx).Create(user).Error)
}

// GetByID retrieves a user by ID.
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id googleUuid.UUID) (*User, error) {
	var user User

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, toAppErr(err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username.
func (r *UserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*User, error) {
	var user User

	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, toAppErr(err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email.
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User

	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, toAppErr(err)
	}

	return &user, nil
}

// ListByTenant retrieves all users for a tenant.
func (r *UserRepositoryImpl) ListByTenant(ctx context.Context, tenantID googleUuid.UUID, activeOnly bool) ([]*User, error) {
	var users []*User

	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if activeOnly {
		query = query.Where("active = ?", true)
	}

	if err := query.Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, toAppErr(err)
	}

	return users, nil
}

// Update updates a user.
func (r *UserRepositoryImpl) Update(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now().UTC()

	return toAppErr(r.db.WithContext(ctx).Save(user).Error)
}

// Delete deletes a user by ID.
func (r *UserRepositoryImpl) Delete(ctx context.Context, id googleUuid.UUID) error {
	return toAppErr(r.db.WithContext(ctx).Delete(&User{}, "id = ?", id).Error)
}

// ClientRepository provides CRUD operations for verified clients.
type ClientRepository interface {
	Create(ctx context.Context, client *Client) error
	GetByID(ctx context.Context, id googleUuid.UUID) (*Client, error)
	GetByClientID(ctx context.Context, clientID string) (*Client, error)
	ListByTenant(ctx context.Context, tenantID googleUuid.UUID, activeOnly bool) ([]*Client, error)
	Update(ctx context.Context, client *Client) error
	Delete(ctx context.Context, id googleUuid.UUID) error
}

// ClientRepositoryImpl implements ClientRepository using GORM.
type ClientRepositoryImpl struct {
	db *gorm.DB
}

// NewClientRepository creates a new ClientRepository.
func NewClientRepository(db *gorm.DB) ClientRepository {
	return &ClientRepositoryImpl{db: db}
}

// Create creates a new client.
func (r *ClientRepositoryImpl) Create(ctx context.Context, client *Client) error {
	return toAppErr(r.db.WithContext(ctx).Create(client).Error)
}

// GetByID retrieves a client by ID.
func (r *ClientRepositoryImpl) GetByID(ctx context.Context, id googleUuid.UUID) (*Client, error) {
	var client Client

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&client).Error; err != nil {
		return nil, toAppErr(err)
	}

	return &client, nil
}

// GetByClientID retrieves a client by client_id.
func (r *ClientRepositoryImpl) GetByClientID(ctx context.Context, clientID string) (*Client, error) {
	var client Client

	if err := r.db.WithContext(ctx).Where("client_id = ?", clientID).First(&client).Error; err != nil {
		return nil, toAppErr(err)
	}

	return &client, nil
}

// ListByTenant retrieves all clients for a tenant.
func (r *ClientRepositoryImpl) ListByTenant(ctx context.Context, tenantID googleUuid.UUID, activeOnly bool) ([]*Client, error) {
	var clients []*Client

	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if activeOnly {
		query = query.Where("active = ?", true)
	}

	if err := query.Order("created_at DESC").Find(&clients).Error; err != nil {
		return nil, toAppErr(err)
	}

	return clients, nil
}

// Update updates a client.
func (r *ClientRepositoryImpl) Update(ctx context.Context, client *Client) error {
	client.UpdatedAt = time.Now().UTC()

	return toAppErr(r.db.WithContext(ctx).Save(client).Error)
}

// Delete deletes a client by ID.
func (r *ClientRepositoryImpl) Delete(ctx context.Context, id googleUuid.UUID) error {
	return toAppErr(r.db.WithContext(ctx).Delete(&Client{}, "id = ?", id).Error)
}

// UnverifiedUserRepository provides CRUD operations for unverified users.
type UnverifiedUserRepository interface {
	Create(ctx context.Context, user *UnverifiedUser) error
	GetByID(ctx context.Context, id googleUuid.UUID) (*UnverifiedUser, error)
	GetByUsername(ctx context.Context, username string) (*UnverifiedUser, error)
	ListByTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*UnverifiedUser, error)
	Delete(ctx context.Context, id googleUuid.UUID) error
	DeleteExpired(ctx context.Context) (int64, error)
}

// UnverifiedUserRepositoryImpl implements UnverifiedUserRepository using GORM.
type UnverifiedUserRepositoryImpl struct {
	db *gorm.DB
}

// NewUnverifiedUserRepository creates a new UnverifiedUserRepository.
func NewUnverifiedUserRepository(db *gorm.DB) UnverifiedUserRepository {
	return &UnverifiedUserRepositoryImpl{db: db}
}

// Create creates a new unverified user.
func (r *UnverifiedUserRepositoryImpl) Create(ctx context.Context, user *UnverifiedUser) error {
	return toAppErr(r.db.WithContext(ctx).Create(user).Error)
}

// GetByID retrieves an unverified user by ID.
func (r *UnverifiedUserRepositoryImpl) GetByID(ctx context.Context, id googleUuid.UUID) (*UnverifiedUser, error) {
	var user UnverifiedUser

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		return nil, toAppErr(err)
	}

	return &user, nil
}

// GetByUsername retrieves an unverified user by username.
func (r *UnverifiedUserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*UnverifiedUser, error) {
	var user UnverifiedUser

	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		return nil, toAppErr(err)
	}

	return &user, nil
}

// ListByTenant retrieves all unverified users for a tenant.
func (r *UnverifiedUserRepositoryImpl) ListByTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*UnverifiedUser, error) {
	var users []*UnverifiedUser

	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, toAppErr(err)
	}

	return users, nil
}

// Delete deletes an unverified user by ID.
func (r *UnverifiedUserRepositoryImpl) Delete(ctx context.Context, id googleUuid.UUID) error {
	return toAppErr(r.db.WithContext(ctx).Delete(&UnverifiedUser{}, "id = ?", id).Error)
}

// DeleteExpired deletes all expired unverified users.
func (r *UnverifiedUserRepositoryImpl) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at < ?", time.Now().UTC()).Delete(&UnverifiedUser{})

	if result.Error != nil {
		return 0, toAppErr(result.Error)
	}

	return result.RowsAffected, nil
}

// UnverifiedClientRepository provides CRUD operations for unverified clients.
type UnverifiedClientRepository interface {
	Create(ctx context.Context, client *UnverifiedClient) error
	GetByID(ctx context.Context, id googleUuid.UUID) (*UnverifiedClient, error)
	GetByClientID(ctx context.Context, clientID string) (*UnverifiedClient, error)
	ListByTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*UnverifiedClient, error)
	Delete(ctx context.Context, id googleUuid.UUID) error
	DeleteExpired(ctx context.Context) (int64, error)
}

// UnverifiedClientRepositoryImpl implements UnverifiedClientRepository using GORM.
type UnverifiedClientRepositoryImpl struct {
	db *gorm.DB
}

// NewUnverifiedClientRepository creates a new UnverifiedClientRepository.
func NewUnverifiedClientRepository(db *gorm.DB) UnverifiedClientRepository {
	return &UnverifiedClientRepositoryImpl{db: db}
}

// Create creates a new unverified client.
func (r *UnverifiedClientRepositoryImpl) Create(ctx context.Context, client *UnverifiedClient) error {
	return toAppErr(r.db.WithContext(ctx).Create(client).Error)
}

// GetByID retrieves an unverified client by ID.
func (r *UnverifiedClientRepositoryImpl) GetByID(ctx context.Context, id googleUuid.UUID) (*UnverifiedClient, error) {
	var client UnverifiedClient

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&client).Error; err != nil {
		return nil, toAppErr(err)
	}

	return &client, nil
}

// GetByClientID retrieves an unverified client by client_id.
func (r *UnverifiedClientRepositoryImpl) GetByClientID(ctx context.Context, clientID string) (*UnverifiedClient, error) {
	var client UnverifiedClient

	if err := r.db.WithContext(ctx).Where("client_id = ?", clientID).First(&client).Error; err != nil {
		return nil, toAppErr(err)
	}

	return &client, nil
}

// ListByTenant retrieves all unverified clients for a tenant.
func (r *UnverifiedClientRepositoryImpl) ListByTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*UnverifiedClient, error) {
	var clients []*UnverifiedClient

	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&clients).Error; err != nil {
		return nil, toAppErr(err)
	}

	return clients, nil
}

// Delete deletes an unverified client by ID.
func (r *UnverifiedClientRepositoryImpl) Delete(ctx context.Context, id googleUuid.UUID) error {
	return toAppErr(r.db.WithContext(ctx).Delete(&UnverifiedClient{}, "id = ?", id).Error)
}

// DeleteExpired deletes all expired unverified clients.
func (r *UnverifiedClientRepositoryImpl) DeleteExpired(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at < ?", time.Now().UTC()).Delete(&UnverifiedClient{})

	if result.Error != nil {
		return 0, toAppErr(result.Error)
	}

	return result.RowsAffected, nil
}
