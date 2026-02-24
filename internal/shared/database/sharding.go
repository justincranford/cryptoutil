// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

package database

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// schemaNamePattern validates schema names to prevent SQL injection.
// Only allows alphanumeric characters, underscores, and hyphens.
var schemaNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ShardStrategy defines how tenants are mapped to database shards.
type ShardStrategy int

const (
	// StrategyRowLevel uses tenant_id column in each table.
	StrategyRowLevel ShardStrategy = iota

	// StrategySchemaLevel uses separate PostgreSQL schemas per tenant.
	StrategySchemaLevel

	// StrategyDatabaseLevel uses separate databases per tenant.
	StrategyDatabaseLevel
)

// String returns the string representation of the shard strategy.
func (s ShardStrategy) String() string {
	switch s {
	case StrategyRowLevel:
		return "row-level"
	case StrategySchemaLevel:
		return "schema-level"
	case StrategyDatabaseLevel:
		return "database-level"
	default:
		return "unknown"
	}
}

// ShardConfig configures database sharding behavior.
type ShardConfig struct {
	Strategy        ShardStrategy // Sharding strategy to use.
	SchemaPrefix    string        // Prefix for tenant schemas (e.g., "tenant_").
	DefaultSchema   string        // Default schema for system tables (e.g., "public").
	EnableMigration bool          // Whether to auto-create schemas.
}

// DefaultShardConfig returns a default shard configuration.
func DefaultShardConfig() *ShardConfig {
	return &ShardConfig{
		Strategy:        StrategyRowLevel,
		SchemaPrefix:    "tenant_",
		DefaultSchema:   "public",
		EnableMigration: true,
	}
}

// ShardManager manages database connections and routing for multi-tenancy.
type ShardManager struct {
	config *ShardConfig
	baseDB *gorm.DB

	// schemaCache caches GORM sessions per tenant schema.
	schemaCache sync.Map
}

// NewShardManager creates a new shard manager.
func NewShardManager(baseDB *gorm.DB, config *ShardConfig) *ShardManager {
	if config == nil {
		config = DefaultShardConfig()
	}

	return &ShardManager{
		config: config,
		baseDB: baseDB,
	}
}

// GetDB returns a GORM DB instance scoped to the tenant in context.
func (sm *ShardManager) GetDB(ctx context.Context) (*gorm.DB, error) {
	tc, err := RequireTenantContext(ctx)
	if err != nil {
		return nil, err
	}

	switch sm.config.Strategy {
	case StrategyRowLevel:
		return sm.getRowLevelDB(ctx, tc)
	case StrategySchemaLevel:
		return sm.getSchemaLevelDB(ctx, tc)
	case StrategyDatabaseLevel:
		return nil, fmt.Errorf("database-level sharding not yet implemented")
	default:
		return nil, fmt.Errorf("unknown shard strategy: %d", sm.config.Strategy)
	}
}

// getRowLevelDB returns a GORM DB with tenant context for row-level isolation.
func (sm *ShardManager) getRowLevelDB(ctx context.Context, tc *TenantContext) (*gorm.DB, error) {
	// Row-level: Return base DB with context (queries must filter by tenant_id).
	return sm.baseDB.WithContext(ctx), nil
}

// getSchemaLevelDB returns a GORM DB with the search_path set to tenant schema.
func (sm *ShardManager) getSchemaLevelDB(ctx context.Context, tc *TenantContext) (*gorm.DB, error) {
	schemaName := sm.config.SchemaPrefix + tc.TenantID.String()

	if err := validateSchemaName(schemaName); err != nil {
		return nil, err
	}

	// Check cache first.
	if cached, ok := sm.schemaCache.Load(schemaName); ok {
		cachedDB, assertOk := cached.(*gorm.DB)
		if !assertOk {
			return nil, fmt.Errorf("invalid cached DB type for schema %s", schemaName)
		}

		return cachedDB.WithContext(ctx), nil
	}

	// Create schema if needed.
	if sm.config.EnableMigration {
		if err := sm.ensureSchema(schemaName); err != nil {
			return nil, fmt.Errorf("failed to create schema %s: %w", schemaName, err)
		}
	}

	// Create a session with the schema search_path.
	db := sm.baseDB.Session(&gorm.Session{NewDB: true})
	if err := db.Exec(fmt.Sprintf("SET search_path TO \"%s\"", schemaName)).Error; err != nil {
		return nil, fmt.Errorf("failed to set search_path: %w", err)
	}

	// Cache the session.
	sm.schemaCache.Store(schemaName, db)

	return db.WithContext(ctx), nil
}

// ensureSchema creates the schema if it doesn't exist.
func (sm *ShardManager) ensureSchema(schemaName string) error {
	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS \"%s\"", schemaName)

	return sm.baseDB.Exec(query).Error
}

// GetTenantSchemaName returns the schema name for a tenant.
func (sm *ShardManager) GetTenantSchemaName(tenantID googleUuid.UUID) string {
	return sm.config.SchemaPrefix + tenantID.String()
}

// DropTenantSchema removes a tenant's schema (use with caution).
func (sm *ShardManager) DropTenantSchema(tenantID googleUuid.UUID) error {
	schemaName := sm.GetTenantSchemaName(tenantID)

	if err := validateSchemaName(schemaName); err != nil {
		return err
	}

	sm.schemaCache.Delete(schemaName)

	query := fmt.Sprintf("DROP SCHEMA IF EXISTS \"%s\" CASCADE", schemaName)

	return sm.baseDB.Exec(query).Error
}

// validateSchemaName ensures the schema name contains only safe characters.
func validateSchemaName(name string) error {
	if !schemaNamePattern.MatchString(name) {
		return fmt.Errorf("invalid schema name %q: must match %s", name, schemaNamePattern.String())
	}

	return nil
}
