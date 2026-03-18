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

// RoleRepository provides CRUD operations for roles.
type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id googleUuid.UUID) (*Role, error)
	GetByName(ctx context.Context, tenantID googleUuid.UUID, name string) (*Role, error)
	ListByTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*Role, error)
	Delete(ctx context.Context, id googleUuid.UUID) error
}

// RoleRepositoryImpl implements RoleRepository using GORM.
type RoleRepositoryImpl struct {
	db *gorm.DB
}

// NewRoleRepository creates a new RoleRepository.
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &RoleRepositoryImpl{db: db}
}

// Create creates a new role.
func (r *RoleRepositoryImpl) Create(ctx context.Context, role *Role) error {
	return toAppErr(r.db.WithContext(ctx).Create(role).Error)
}

// GetByID retrieves a role by ID.
func (r *RoleRepositoryImpl) GetByID(ctx context.Context, id googleUuid.UUID) (*Role, error) {
	var role Role

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error; err != nil {
		return nil, toAppErr(err)
	}

	return &role, nil
}

// GetByName retrieves a role by tenant ID and name.
func (r *RoleRepositoryImpl) GetByName(ctx context.Context, tenantID googleUuid.UUID, name string) (*Role, error) {
	var role Role

	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND name = ?", tenantID, name).First(&role).Error; err != nil {
		return nil, toAppErr(err)
	}

	return &role, nil
}

// ListByTenant retrieves all roles for a tenant.
func (r *RoleRepositoryImpl) ListByTenant(ctx context.Context, tenantID googleUuid.UUID) ([]*Role, error) {
	var roles []*Role

	if err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("name ASC").Find(&roles).Error; err != nil {
		return nil, toAppErr(err)
	}

	return roles, nil
}

// Delete deletes a role by ID.
func (r *RoleRepositoryImpl) Delete(ctx context.Context, id googleUuid.UUID) error {
	return toAppErr(r.db.WithContext(ctx).Delete(&Role{}, "id = ?", id).Error)
}

// UserRoleRepository provides operations for user-role assignments.
type UserRoleRepository interface {
	Assign(ctx context.Context, userRole *UserRole) error
	Revoke(ctx context.Context, userID, roleID googleUuid.UUID) error
	ListRolesByUser(ctx context.Context, userID googleUuid.UUID) ([]*Role, error)
	ListUsersByRole(ctx context.Context, roleID googleUuid.UUID) ([]*User, error)
}

// UserRoleRepositoryImpl implements UserRoleRepository using GORM.
type UserRoleRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRoleRepository creates a new UserRoleRepository.
func NewUserRoleRepository(db *gorm.DB) UserRoleRepository {
	return &UserRoleRepositoryImpl{db: db}
}

// Assign assigns a role to a user.
func (r *UserRoleRepositoryImpl) Assign(ctx context.Context, userRole *UserRole) error {
	return toAppErr(r.db.WithContext(ctx).Create(userRole).Error)
}

// Revoke revokes a role from a user.
func (r *UserRoleRepositoryImpl) Revoke(ctx context.Context, userID, roleID googleUuid.UUID) error {
	return toAppErr(r.db.WithContext(ctx).Delete(&UserRole{}, "user_id = ? AND role_id = ?", userID, roleID).Error)
}

// ListRolesByUser retrieves all roles assigned to a user.
func (r *UserRoleRepositoryImpl) ListRolesByUser(ctx context.Context, userID googleUuid.UUID) ([]*Role, error) {
	var roles []*Role

	if err := r.db.WithContext(ctx).
		Table("roles").
		Joins("INNER JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error; err != nil {
		return nil, toAppErr(err)
	}

	return roles, nil
}

// ListUsersByRole retrieves all users assigned to a role.
func (r *UserRoleRepositoryImpl) ListUsersByRole(ctx context.Context, roleID googleUuid.UUID) ([]*User, error) {
	var users []*User

	if err := r.db.WithContext(ctx).
		Table("users").
		Joins("INNER JOIN user_roles ON user_roles.user_id = users.id").
		Where("user_roles.role_id = ?", roleID).
		Find(&users).Error; err != nil {
		return nil, toAppErr(err)
	}

	return users, nil
}

// ClientRoleRepository provides operations for client-role assignments.
type ClientRoleRepository interface {
	Assign(ctx context.Context, clientRole *ClientRole) error
	Revoke(ctx context.Context, clientID, roleID googleUuid.UUID) error
	ListRolesByClient(ctx context.Context, clientID googleUuid.UUID) ([]*Role, error)
	ListClientsByRole(ctx context.Context, roleID googleUuid.UUID) ([]*Client, error)
}

// ClientRoleRepositoryImpl implements ClientRoleRepository using GORM.
type ClientRoleRepositoryImpl struct {
	db *gorm.DB
}

// NewClientRoleRepository creates a new ClientRoleRepository.
func NewClientRoleRepository(db *gorm.DB) ClientRoleRepository {
	return &ClientRoleRepositoryImpl{db: db}
}

// Assign assigns a role to a client.
func (r *ClientRoleRepositoryImpl) Assign(ctx context.Context, clientRole *ClientRole) error {
	return toAppErr(r.db.WithContext(ctx).Create(clientRole).Error)
}

// Revoke revokes a role from a client.
func (r *ClientRoleRepositoryImpl) Revoke(ctx context.Context, clientID, roleID googleUuid.UUID) error {
	return toAppErr(r.db.WithContext(ctx).Delete(&ClientRole{}, "client_id = ? AND role_id = ?", clientID, roleID).Error)
}

// ListRolesByClient retrieves all roles assigned to a client.
func (r *ClientRoleRepositoryImpl) ListRolesByClient(ctx context.Context, clientID googleUuid.UUID) ([]*Role, error) {
	var roles []*Role

	if err := r.db.WithContext(ctx).
		Table("roles").
		Joins("INNER JOIN client_roles ON client_roles.role_id = roles.id").
		Where("client_roles.client_id = ?", clientID).
		Find(&roles).Error; err != nil {
		return nil, toAppErr(err)
	}

	return roles, nil
}

// ListClientsByRole retrieves all clients assigned to a role.
func (r *ClientRoleRepositoryImpl) ListClientsByRole(ctx context.Context, roleID googleUuid.UUID) ([]*Client, error) {
	var clients []*Client

	if err := r.db.WithContext(ctx).
		Table("clients").
		Joins("INNER JOIN client_roles ON client_roles.client_id = clients.id").
		Where("client_roles.role_id = ?", roleID).
		Find(&clients).Error; err != nil {
		return nil, toAppErr(err)
	}

	return clients, nil
}

// TenantRealmRepository provides CRUD operations for tenant realms.
type TenantRealmRepository interface {
	Create(ctx context.Context, realm *TenantRealm) error
	GetByID(ctx context.Context, id googleUuid.UUID) (*TenantRealm, error)
	GetByRealmID(ctx context.Context, tenantID, realmID googleUuid.UUID) (*TenantRealm, error)
	ListByTenant(ctx context.Context, tenantID googleUuid.UUID, activeOnly bool) ([]*TenantRealm, error)
	Update(ctx context.Context, realm *TenantRealm) error
	Delete(ctx context.Context, id googleUuid.UUID) error
}

// TenantRealmRepositoryImpl implements TenantRealmRepository using GORM.
type TenantRealmRepositoryImpl struct {
	db *gorm.DB
}

// NewTenantRealmRepository creates a new TenantRealmRepository.
func NewTenantRealmRepository(db *gorm.DB) TenantRealmRepository {
	return &TenantRealmRepositoryImpl{db: db}
}

// Create creates a new tenant realm.
func (r *TenantRealmRepositoryImpl) Create(ctx context.Context, realm *TenantRealm) error {
	return toAppErr(r.db.WithContext(ctx).Create(realm).Error)
}

// GetByID retrieves a tenant realm by ID.
func (r *TenantRealmRepositoryImpl) GetByID(ctx context.Context, id googleUuid.UUID) (*TenantRealm, error) {
	var realm TenantRealm

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&realm).Error; err != nil {
		return nil, toAppErr(err)
	}

	return &realm, nil
}

// GetByRealmID retrieves a tenant realm by tenant ID and realm ID.
func (r *TenantRealmRepositoryImpl) GetByRealmID(ctx context.Context, tenantID, realmID googleUuid.UUID) (*TenantRealm, error) {
	var realm TenantRealm

	if err := r.db.WithContext(ctx).Where("tenant_id = ? AND realm_id = ?", tenantID, realmID).First(&realm).Error; err != nil {
		return nil, toAppErr(err)
	}

	return &realm, nil
}

// ListByTenant retrieves all realms for a tenant.
func (r *TenantRealmRepositoryImpl) ListByTenant(ctx context.Context, tenantID googleUuid.UUID, activeOnly bool) ([]*TenantRealm, error) {
	var realms []*TenantRealm

	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if activeOnly {
		query = query.Where("active = ?", true)
	}

	if err := query.Order("created_at DESC").Find(&realms).Error; err != nil {
		return nil, toAppErr(err)
	}

	return realms, nil
}

// Update updates a tenant realm.
func (r *TenantRealmRepositoryImpl) Update(ctx context.Context, realm *TenantRealm) error {
	realm.UpdatedAt = time.Now().UTC()

	return toAppErr(r.db.WithContext(ctx).Save(realm).Error)
}

// Delete deletes a tenant realm by ID.
func (r *TenantRealmRepositoryImpl) Delete(ctx context.Context, id googleUuid.UUID) error {
	return toAppErr(r.db.WithContext(ctx).Delete(&TenantRealm{}, "id = ?", id).Error)
}
