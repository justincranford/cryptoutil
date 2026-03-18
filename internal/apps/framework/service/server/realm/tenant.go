// Copyright (c) 2025 Justin Cranford
//
//

// Package realm provides tenant isolation for KMS.
package realm

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gorm.io/gorm"
)

// TenantIsolationMode defines how tenant data is isolated.
type TenantIsolationMode string

const (
	// TenantIsolationSchema uses separate database schemas per tenant.
	TenantIsolationSchema TenantIsolationMode = "schema"

	// TenantIsolationRow uses row-level filtering with tenant_id column.
	TenantIsolationRow TenantIsolationMode = "row"

	// TenantIsolationDatabase uses separate databases per tenant.
	TenantIsolationDatabase TenantIsolationMode = "database"
)

// TenantConfig defines tenant configuration.
type TenantConfig struct {
	// ID is the unique tenant identifier (UUIDv4).
	ID string `yaml:"id" json:"id"`

	// Name is the human-readable tenant name.
	Name string `yaml:"name" json:"name"`

	// Description is the optional tenant description.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// RealmID is the realm this tenant belongs to.
	RealmID string `yaml:"realm_id" json:"realm_id"`

	// IsolationMode determines how tenant data is isolated.
	IsolationMode TenantIsolationMode `yaml:"isolation_mode" json:"isolation_mode"`

	// SchemaName is the database schema name (for schema isolation).
	SchemaName string `yaml:"schema_name,omitempty" json:"schema_name,omitempty"`

	// Enabled indicates if the tenant is active.
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// TenantManager manages tenant isolation and context.
type TenantManager struct {
	db             *gorm.DB
	isolationMode  TenantIsolationMode
	tenants        map[string]*TenantConfig
	schemaCreated  map[string]bool
	mu             sync.RWMutex
	defaultRealmID string
}

// TenantManagerConfig configures the tenant manager.
type TenantManagerConfig struct {
	// IsolationMode is the default isolation mode for tenants.
	IsolationMode TenantIsolationMode `yaml:"isolation_mode" json:"isolation_mode"`

	// DefaultRealmID is the default realm for new tenants.
	DefaultRealmID string `yaml:"default_realm_id" json:"default_realm_id"`

	// AutoCreateSchema enables automatic schema creation for new tenants.
	AutoCreateSchema bool `yaml:"auto_create_schema" json:"auto_create_schema"`
}

// NewTenantManager creates a new tenant manager.
func NewTenantManager(db *gorm.DB, config *TenantManagerConfig) (*TenantManager, error) {
	if db == nil {
		return nil, errors.New("database connection is required")
	}

	if config == nil {
		config = &TenantManagerConfig{
			IsolationMode: TenantIsolationRow,
		}
	}

	return &TenantManager{
		db:             db,
		isolationMode:  config.IsolationMode,
		tenants:        make(map[string]*TenantConfig),
		schemaCreated:  make(map[string]bool),
		defaultRealmID: config.DefaultRealmID,
	}, nil
}

// RegisterTenant registers a new tenant.
func (m *TenantManager) RegisterTenant(ctx context.Context, tenant *TenantConfig) error {
	if tenant == nil {
		return errors.New("tenant cannot be nil")
	}

	if tenant.ID == "" {
		return errors.New("tenant ID is required")
	}

	if tenant.Name == "" {
		return errors.New("tenant name is required")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate.
	if _, exists := m.tenants[tenant.ID]; exists {
		return fmt.Errorf("tenant %s already exists", tenant.ID)
	}

	// Set defaults.
	if tenant.IsolationMode == "" {
		tenant.IsolationMode = m.isolationMode
	}

	if tenant.RealmID == "" {
		tenant.RealmID = m.defaultRealmID
	}

	// Create schema if using schema isolation.
	if tenant.IsolationMode == TenantIsolationSchema {
		if err := m.createTenantSchema(ctx, tenant); err != nil {
			return fmt.Errorf("failed to create tenant schema: %w", err)
		}
	}

	m.tenants[tenant.ID] = tenant

	return nil
}

// GetTenant retrieves a tenant by ID.
func (m *TenantManager) GetTenant(tenantID string) (*TenantConfig, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenant, ok := m.tenants[tenantID]

	return tenant, ok
}

// ListTenants returns all registered tenants.
func (m *TenantManager) ListTenants() []TenantConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenants := make([]TenantConfig, 0, len(m.tenants))
	for _, tenant := range m.tenants {
		tenants = append(tenants, *tenant)
	}

	return tenants
}

// DeleteTenant removes a tenant.
func (m *TenantManager) DeleteTenant(ctx context.Context, tenantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, exists := m.tenants[tenantID]
	if !exists {
		return fmt.Errorf("tenant %s not found", tenantID)
	}

	// Drop schema if using schema isolation.
	if tenant.IsolationMode == TenantIsolationSchema && tenant.SchemaName != "" {
		if err := m.dropTenantSchema(ctx, tenant); err != nil {
			return fmt.Errorf("failed to drop tenant schema: %w", err)
		}
	}

	delete(m.tenants, tenantID)

	return nil
}

// WithTenant returns a GORM DB scoped to the specified tenant.
func (m *TenantManager) WithTenant(ctx context.Context, tenantID string) (*gorm.DB, error) {
	m.mu.RLock()
	tenant, exists := m.tenants[tenantID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tenant %s not found", tenantID)
	}

	if !tenant.Enabled {
		return nil, fmt.Errorf("tenant %s is disabled", tenantID)
	}

	db := m.db.WithContext(ctx)

	switch tenant.IsolationMode {
	case TenantIsolationSchema:
		// Use schema search path (PostgreSQL specific).
		if tenant.SchemaName != "" {
			db = db.Exec(fmt.Sprintf("SET search_path TO %s", tenant.SchemaName))
		}
	case TenantIsolationRow:
		// Add tenant_id filter using GORM scopes.
		db = db.Scopes(func(tx *gorm.DB) *gorm.DB {
			return tx.Where("tenant_id = ?", tenantID)
		})
	case TenantIsolationDatabase:
		// Separate database would require new connection - not implemented.
		return nil, errors.New("database isolation mode not implemented")
	}

	return db, nil
}

// createTenantSchema creates a database schema for the tenant.
func (m *TenantManager) createTenantSchema(ctx context.Context, tenant *TenantConfig) error {
	if tenant.SchemaName == "" {
		tenant.SchemaName = sanitizeSchemaName("tenant_" + tenant.ID)
	}

	// Check if already created.
	if m.schemaCreated[tenant.ID] {
		return nil
	}

	// Create schema (PostgreSQL specific).
	sql := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", tenant.SchemaName)
	if err := m.db.WithContext(ctx).Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to create schema %s: %w", tenant.SchemaName, err)
	}

	m.schemaCreated[tenant.ID] = true

	return nil
}

// dropTenantSchema drops a database schema for the tenant.
func (m *TenantManager) dropTenantSchema(ctx context.Context, tenant *TenantConfig) error {
	if tenant.SchemaName == "" {
		return nil
	}

	// Drop schema (PostgreSQL specific).
	sql := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", tenant.SchemaName)
	if err := m.db.WithContext(ctx).Exec(sql).Error; err != nil {
		return fmt.Errorf("failed to drop schema %s: %w", tenant.SchemaName, err)
	}

	delete(m.schemaCreated, tenant.ID)

	return nil
}

// sanitizeSchemaName sanitizes a schema name to be SQL-safe.
func sanitizeSchemaName(name string) string {
	// Replace non-alphanumeric characters with underscores.
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}

		return '_'
	}, name)

	// Ensure it starts with a letter.
	if len(safe) > 0 && safe[0] >= '0' && safe[0] <= '9' {
		safe = "t_" + safe
	}

	// Limit length.
	const maxSchemaNameLength = 63
	if len(safe) > maxSchemaNameLength {
		safe = safe[:maxSchemaNameLength]
	}

	return strings.ToLower(safe)
}

// TenantContext holds tenant information in context.
type TenantContext struct {
	TenantID   string
	TenantName string
	RealmID    string
}

// tenantContextKey is the context key for tenant information.
type tenantContextKey struct{}

// ContextWithTenant adds tenant information to the context.
func ContextWithTenant(ctx context.Context, tenant *TenantContext) context.Context {
	return context.WithValue(ctx, tenantContextKey{}, tenant)
}

// TenantFromContext retrieves tenant information from the context.
func TenantFromContext(ctx context.Context) (*TenantContext, bool) {
	tenant, ok := ctx.Value(tenantContextKey{}).(*TenantContext)

	return tenant, ok
}

// ValidateTenantID validates a tenant ID format.
func ValidateTenantID(tenantID string) error {
	if tenantID == "" {
		return errors.New("tenant ID cannot be empty")
	}

	// UUIDs are 36 characters with hyphens.
	if len(tenantID) != cryptoutilSharedMagic.UUIDStringLength {
		return fmt.Errorf("tenant ID must be a valid UUID (36 characters), got %d", len(tenantID))
	}

	return nil
}
