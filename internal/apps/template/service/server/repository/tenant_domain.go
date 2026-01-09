// Copyright (c) 2025 Justin Cranford
//
//

package repository

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// Tenant represents a tenant in a multi-tenant system.
// Each tenant has isolated users, clients, and realm configurations.
type Tenant struct {
	ID          googleUuid.UUID `gorm:"type:text;primaryKey"`
	Name        string          `gorm:"type:text;not null;uniqueIndex"`
	Description string          `gorm:"type:text"`
	Active      bool            `gorm:"not null;default:true;index"`
	CreatedAt   time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

// TableName returns the database table name for Tenant.
func (Tenant) TableName() string {
	return "tenants"
}

// User represents a verified user associated with a tenant.
type User struct {
	ID           googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID     googleUuid.UUID `gorm:"type:text;not null;index"`
	Username     string          `gorm:"type:text;not null;uniqueIndex"`
	PasswordHash string          `gorm:"type:text;not null"`
	Email        string          `gorm:"type:text;uniqueIndex"`
	Active       bool            `gorm:"not null;default:true;index"`
	CreatedAt    time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// Relationship.
	Tenant *Tenant `gorm:"foreignKey:TenantID;references:ID"`
}

// TableName returns the database table name for User.
func (User) TableName() string {
	return "users"
}

// Client represents a verified non-browser client associated with a tenant.
type Client struct {
	ID           googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID     googleUuid.UUID `gorm:"type:text;not null;index"`
	ClientID     string          `gorm:"type:text;not null;uniqueIndex"`
	ClientSecret string          `gorm:"type:text;not null"`
	Name         string          `gorm:"type:text"`
	Active       bool            `gorm:"not null;default:true;index"`
	CreatedAt    time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// Relationship.
	Tenant *Tenant `gorm:"foreignKey:TenantID;references:ID"`
}

// TableName returns the database table name for Client.
func (Client) TableName() string {
	return "clients"
}

// UnverifiedUser represents a user awaiting admin verification.
// Auto-expires after ExpiresAt timestamp.
type UnverifiedUser struct {
	ID           googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID     googleUuid.UUID `gorm:"type:text;not null;index"`
	Username     string          `gorm:"type:text;not null;uniqueIndex"`
	PasswordHash string          `gorm:"type:text;not null"`
	Email        string          `gorm:"type:text;uniqueIndex"`
	CreatedAt    time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`
	ExpiresAt    time.Time       `gorm:"not null;index"` // Auto-expire timestamp.

	// Relationship.
	Tenant *Tenant `gorm:"foreignKey:TenantID;references:ID"`
}

// TableName returns the database table name for UnverifiedUser.
func (UnverifiedUser) TableName() string {
	return "unverified_users"
}

// IsExpired checks if the unverified user has expired.
func (u *UnverifiedUser) IsExpired() bool {
	return time.Now().After(u.ExpiresAt)
}

// UnverifiedClient represents a client awaiting admin verification.
// Auto-expires after ExpiresAt timestamp.
type UnverifiedClient struct {
	ID           googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID     googleUuid.UUID `gorm:"type:text;not null;index"`
	ClientID     string          `gorm:"type:text;not null;uniqueIndex"`
	ClientSecret string          `gorm:"type:text;not null"`
	Name         string          `gorm:"type:text"`
	CreatedAt    time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`
	ExpiresAt    time.Time       `gorm:"not null;index"` // Auto-expire timestamp.

	// Relationship.
	Tenant *Tenant `gorm:"foreignKey:TenantID;references:ID"`
}

// TableName returns the database table name for UnverifiedClient.
func (UnverifiedClient) TableName() string {
	return "unverified_clients"
}

// IsExpired checks if the unverified client has expired.
func (c *UnverifiedClient) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// Role represents a role that can be assigned to users or clients.
type Role struct {
	ID          googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID    googleUuid.UUID `gorm:"type:text;not null;index"`
	Name        string          `gorm:"type:text;not null"`
	Description string          `gorm:"type:text"`
	CreatedAt   time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// Relationship.
	Tenant *Tenant `gorm:"foreignKey:TenantID;references:ID"`
}

// TableName returns the database table name for Role.
func (Role) TableName() string {
	return "roles"
}

// UserRole represents the many-to-many relationship between users and roles.
type UserRole struct {
	UserID    googleUuid.UUID `gorm:"type:text;primaryKey"`
	RoleID    googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID  googleUuid.UUID `gorm:"type:text;not null;index"`
	CreatedAt time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// Relationships.
	User   *User   `gorm:"foreignKey:UserID;references:ID"`
	Role   *Role   `gorm:"foreignKey:RoleID;references:ID"`
	Tenant *Tenant `gorm:"foreignKey:TenantID;references:ID"`
}

// TableName returns the database table name for UserRole.
func (UserRole) TableName() string {
	return "user_roles"
}

// ClientRole represents the many-to-many relationship between clients and roles.
type ClientRole struct {
	ClientID  googleUuid.UUID `gorm:"type:text;primaryKey"`
	RoleID    googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID  googleUuid.UUID `gorm:"type:text;not null;index"`
	CreatedAt time.Time       `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// Relationships.
	Client *Client `gorm:"foreignKey:ClientID;references:ID"`
	Role   *Role   `gorm:"foreignKey:RoleID;references:ID"`
	Tenant *Tenant `gorm:"foreignKey:TenantID;references:ID"`
}

// TableName returns the database table name for ClientRole.
func (ClientRole) TableName() string {
	return "client_roles"
}

// TenantRealm represents a realm configuration for a tenant.
// Realms define authentication methods and security policies.
type TenantRealm struct {
	ID       googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID googleUuid.UUID `gorm:"type:text;not null;index"`
	RealmID  googleUuid.UUID `gorm:"type:text;not null"`             // Unique realm identifier per tenant.
	Type     string          `gorm:"type:text;not null"`             // Realm type: username_password, ldap, oauth2.
	Config   string          `gorm:"type:text"`                      // JSON configuration for realm.
	Active   bool            `gorm:"not null;default:true;index"`    // Active/inactive realm.
	Source   string          `gorm:"type:text;not null;default:db"` // Source: db or file.
	CreatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`

	// Relationship.
	Tenant *Tenant `gorm:"foreignKey:TenantID;references:ID"`
}

// TableName returns the database table name for TenantRealm.
func (TenantRealm) TableName() string {
	return "tenant_realms"
}
