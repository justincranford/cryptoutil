// Copyright (c) 2025 Justin Cranford
//

// Package builder provides fluent API for constructing service applications.
package builder

import (
	"errors"
	"io/fs"
)

// MigrationMode defines how migrations are handled.
type MigrationMode string

const (
	// MigrationModeTemplateWithDomain uses template migrations (1001-1004) merged with domain migrations (2001+).
	// This is the default mode for services using the template pattern.
	MigrationModeTemplateWithDomain MigrationMode = "template_with_domain"

	// MigrationModeDomainOnly uses only domain migrations without template infrastructure.
	// Use this for services that have their own complete migration scheme.
	MigrationModeDomainOnly MigrationMode = "domain_only"
)

// ErrMigrationModeRequired is returned when migration mode is not specified.
var ErrMigrationModeRequired = errors.New("migration mode is required")

// ErrMigrationFSRequired is returned when domain FS is required but not provided.
var ErrMigrationFSRequired = errors.New("migration FS is required for this mode")

// MigrationConfig configures the migration handling for ServerBuilder.
type MigrationConfig struct {
	// Mode determines how migrations are handled.
	Mode MigrationMode

	// DomainFS contains domain-specific migrations.
	// Required for both MigrationModeTemplateWithDomain and MigrationModeDomainOnly.
	DomainFS fs.FS

	// DomainPath is the path within DomainFS (e.g., "migrations").
	DomainPath string

	// SkipTemplateMigrations allows using only domain migrations.
	// When true, template migrations (sessions, barrier, realms, tenants) are not applied.
	// ONLY set this if the service provides its own session/barrier/realm/tenant tables.
	SkipTemplateMigrations bool
}

// NewDefaultMigrationConfig creates a MigrationConfig that requires domain migrations.
// Template migrations will be applied, followed by domain migrations.
func NewDefaultMigrationConfig() *MigrationConfig {
	return &MigrationConfig{
		Mode:                   MigrationModeTemplateWithDomain,
		SkipTemplateMigrations: false,
	}
}

// NewDomainOnlyMigrationConfig creates a MigrationConfig for services with their own
// complete migration scheme. Template migrations are NOT applied.
func NewDomainOnlyMigrationConfig() *MigrationConfig {
	return &MigrationConfig{
		Mode:                   MigrationModeDomainOnly,
		SkipTemplateMigrations: true,
	}
}

// WithDomainFS sets the domain migrations filesystem.
func (c *MigrationConfig) WithDomainFS(domainFS fs.FS) *MigrationConfig {
	c.DomainFS = domainFS

	return c
}

// WithDomainPath sets the path within the domain filesystem.
func (c *MigrationConfig) WithDomainPath(domainPath string) *MigrationConfig {
	c.DomainPath = domainPath

	return c
}

// WithMode sets the migration mode.
func (c *MigrationConfig) WithMode(mode MigrationMode) *MigrationConfig {
	c.Mode = mode

	return c
}

// WithSkipTemplateMigrations sets whether to skip template migrations.
func (c *MigrationConfig) WithSkipTemplateMigrations(skip bool) *MigrationConfig {
	c.SkipTemplateMigrations = skip

	return c
}

// Validate checks that the configuration is valid.
func (c *MigrationConfig) Validate() error {
	if c == nil {
		return ErrMigrationModeRequired
	}

	if c.Mode == "" {
		return ErrMigrationModeRequired
	}

	switch c.Mode {
	case MigrationModeTemplateWithDomain:
		if c.DomainFS == nil {
			return ErrMigrationFSRequired
		}

		if c.DomainPath == "" {
			return errors.New("domain path is required when domain FS is provided")
		}
	case MigrationModeDomainOnly:
		if c.DomainFS == nil {
			return ErrMigrationFSRequired
		}

		if c.DomainPath == "" {
			return errors.New("domain path is required for domain-only mode")
		}
	default:
		return errors.New("invalid migration mode: " + string(c.Mode))
	}

	return nil
}

// RequiresTemplateMigrations returns true if template migrations should be applied.
func (c *MigrationConfig) RequiresTemplateMigrations() bool {
	return c.Mode == MigrationModeTemplateWithDomain && !c.SkipTemplateMigrations
}
