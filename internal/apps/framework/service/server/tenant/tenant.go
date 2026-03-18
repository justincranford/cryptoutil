// Copyright (c) 2025 Justin Cranford
//
//

// Package tenant provides schema-per-tenant database isolation.
// PostgreSQL: CREATE SCHEMA tenant_xxx; CREATE TABLE tenant_xxx.users
// SQLite: ATTACH 'tenant_xxx.db' AS tenant_xxx; CREATE TABLE tenant_xxx.users
package tenant

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"gorm.io/gorm"
)

// SchemaPrefix is the prefix used for tenant schema names.
const SchemaPrefix = "tenant_"

// DBType represents the supported database types.
type DBType string

// Supported database types.
const (
	DBTypeSQLite   DBType = "sqlite"
	DBTypePostgres DBType = "postgres"
)

// SchemaManager handles schema-per-tenant operations for both PostgreSQL and SQLite.
type SchemaManager struct {
	db     *gorm.DB
	sqlDB  *sql.DB
	dbType DBType
}

// NewSchemaManager creates a new SchemaManager instance.
func NewSchemaManager(db *gorm.DB, dbType DBType) (*SchemaManager, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	return &SchemaManager{
		db:     db,
		sqlDB:  sqlDB,
		dbType: dbType,
	}, nil
}

// SchemaName returns the schema name for a given tenant ID.
func SchemaName(tenantID string) string {
	// Sanitize tenant ID to only allow alphanumeric and hyphens.
	sanitized := sanitizeTenantID(tenantID)

	return SchemaPrefix + sanitized
}

// sanitizeTenantID removes all characters except alphanumeric and hyphens.
// Also replaces hyphens with underscores for SQL compatibility.
func sanitizeTenantID(tenantID string) string {
	var result strings.Builder

	result.Grow(len(tenantID))

	for _, r := range tenantID {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		} else if r == '-' {
			result.WriteRune('_')
		}
	}

	return result.String()
}

// CreateSchema creates a schema for the given tenant ID.
func (sm *SchemaManager) CreateSchema(ctx context.Context, tenantID string) error {
	schemaName := SchemaName(tenantID)

	switch sm.dbType {
	case DBTypePostgres:
		return sm.createPostgresSchema(ctx, schemaName)
	case DBTypeSQLite:
		return sm.createSQLiteSchema(ctx, schemaName)
	default:
		return fmt.Errorf("unsupported database type: %s", sm.dbType)
	}
}

// DropSchema drops the schema for the given tenant ID.
func (sm *SchemaManager) DropSchema(ctx context.Context, tenantID string) error {
	schemaName := SchemaName(tenantID)

	switch sm.dbType {
	case DBTypePostgres:
		return sm.dropPostgresSchema(ctx, schemaName)
	case DBTypeSQLite:
		return sm.dropSQLiteSchema(ctx, schemaName)
	default:
		return fmt.Errorf("unsupported database type: %s", sm.dbType)
	}
}

// SchemaExists checks if a schema exists for the given tenant ID.
func (sm *SchemaManager) SchemaExists(ctx context.Context, tenantID string) (bool, error) {
	schemaName := SchemaName(tenantID)

	switch sm.dbType {
	case DBTypePostgres:
		return sm.postgresSchemaExists(ctx, schemaName)
	case DBTypeSQLite:
		return sm.sqliteSchemaExists(ctx, schemaName)
	default:
		return false, fmt.Errorf("unsupported database type: %s", sm.dbType)
	}
}

// ListSchemas returns all tenant schemas.
func (sm *SchemaManager) ListSchemas(ctx context.Context) ([]string, error) {
	switch sm.dbType {
	case DBTypePostgres:
		return sm.listPostgresSchemas(ctx)
	case DBTypeSQLite:
		return sm.listSQLiteSchemas(ctx)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", sm.dbType)
	}
}

// GetScopedDB returns a GORM DB scoped to a specific tenant's schema.
func (sm *SchemaManager) GetScopedDB(tenantID string) *gorm.DB {
	schemaName := SchemaName(tenantID)

	switch sm.dbType {
	case DBTypePostgres:
		// PostgreSQL: Use search_path to scope queries.
		return sm.db.Session(&gorm.Session{}).Exec(fmt.Sprintf("SET search_path TO %s, public", schemaName))
	case DBTypeSQLite:
		// SQLite: Tables are prefixed with schema name (attached database).
		// GORM queries need to use full table names: schema.table.
		return sm.db.Table(schemaName + ".")
	default:
		return sm.db
	}
}

// PostgreSQL schema operations.

func (sm *SchemaManager) createPostgresSchema(ctx context.Context, schemaName string) error {
	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)
	if _, err := sm.sqlDB.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to create PostgreSQL schema %s: %w", schemaName, err)
	}

	return nil
}

func (sm *SchemaManager) dropPostgresSchema(ctx context.Context, schemaName string) error {
	query := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)
	if _, err := sm.sqlDB.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to drop PostgreSQL schema %s: %w", schemaName, err)
	}

	return nil
}

func (sm *SchemaManager) postgresSchemaExists(ctx context.Context, schemaName string) (bool, error) {
	query := "SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = $1)"

	var exists bool
	if err := sm.sqlDB.QueryRowContext(ctx, query, schemaName).Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check PostgreSQL schema existence: %w", err)
	}

	return exists, nil
}

func (sm *SchemaManager) listPostgresSchemas(ctx context.Context) ([]string, error) {
	query := fmt.Sprintf("SELECT schema_name FROM information_schema.schemata WHERE schema_name LIKE '%s%%'", SchemaPrefix)

	rows, err := sm.sqlDB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list PostgreSQL schemas: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var schemas []string

	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, fmt.Errorf("failed to scan schema name: %w", err)
		}

		schemas = append(schemas, schemaName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schema rows: %w", err)
	}

	return schemas, nil
}

// SQLite schema operations (using attached databases).

func (sm *SchemaManager) createSQLiteSchema(ctx context.Context, schemaName string) error {
	// SQLite uses ATTACH DATABASE to create isolated schemas.
	// For in-memory databases, use file::memory: with unique name.
	dbPath := fmt.Sprintf("file:%s?mode=memory&cache=shared", schemaName)
	query := fmt.Sprintf("ATTACH DATABASE '%s' AS %s", dbPath, schemaName)

	if _, err := sm.sqlDB.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to attach SQLite database %s: %w", schemaName, err)
	}

	return nil
}

func (sm *SchemaManager) dropSQLiteSchema(ctx context.Context, schemaName string) error {
	query := fmt.Sprintf("DETACH DATABASE %s", schemaName)
	if _, err := sm.sqlDB.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("failed to detach SQLite database %s: %w", schemaName, err)
	}

	return nil
}

func (sm *SchemaManager) sqliteSchemaExists(ctx context.Context, schemaName string) (bool, error) {
	query := "SELECT name FROM pragma_database_list WHERE name = ?"

	var name string

	err := sm.sqlDB.QueryRowContext(ctx, query, schemaName).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to check SQLite schema existence: %w", err)
	}

	return true, nil
}

func (sm *SchemaManager) listSQLiteSchemas(ctx context.Context) ([]string, error) {
	query := "SELECT name FROM pragma_database_list WHERE name LIKE ? AND name != 'main' AND name != 'temp'"

	rows, err := sm.sqlDB.QueryContext(ctx, query, SchemaPrefix+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to list SQLite schemas: %w", err)
	}

	defer func() { _ = rows.Close() }()

	var schemas []string

	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, fmt.Errorf("failed to scan schema name: %w", err)
		}

		schemas = append(schemas, schemaName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schema rows: %w", err)
	}

	return schemas, nil
}

// TenantContext holds tenant-scoped context values.
type TenantContext struct {
	TenantID   string
	SchemaName string
}

// tenantContextKey is used to store tenant context in context.Context.
type tenantContextKey struct{}

// WithTenant returns a context with tenant information.
func WithTenant(ctx context.Context, tenantID string) context.Context {
	tc := &TenantContext{
		TenantID:   tenantID,
		SchemaName: SchemaName(tenantID),
	}

	return context.WithValue(ctx, tenantContextKey{}, tc)
}

// GetTenant retrieves tenant context from context.Context.
func GetTenant(ctx context.Context) *TenantContext {
	if tc, ok := ctx.Value(tenantContextKey{}).(*TenantContext); ok {
		return tc
	}

	return nil
}

// IsValidTenantID validates a tenant ID format.
// Must be a valid UUID format (36 characters with hyphens).
func IsValidTenantID(tenantID string) bool {
	if len(tenantID) != cryptoutilSharedMagic.UUIDStringLength {
		return false
	}

	// Expected positions for hyphens: 8, 13, 18, 23.
	hyphenPositions := []int{cryptoutilSharedMagic.IMMinPasswordLength, 13, 18, 23}
	for _, pos := range hyphenPositions {
		if tenantID[pos] != '-' {
			return false
		}
	}

	// Validate hex characters at non-hyphen positions.
	for i, r := range tenantID {
		if i == cryptoutilSharedMagic.IMMinPasswordLength || i == 13 || i == 18 || i == 23 {
			continue
		}

		if !isHexChar(r) {
			return false
		}
	}

	return true
}

// isHexChar checks if a rune is a valid hexadecimal character.
func isHexChar(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}
