// Copyright (c) 2025 Justin Cranford

package repository

import (
	"embed"
	"fmt"
	"io/fs"
)

// MergedMigrationsFS combines template service migrations (1001-1999) with
// domain-specific migrations (2001+) into a single filesystem view.
// Domain migrations take priority (tried first), falling back to template.
//
// Implements fs.FS, fs.ReadDirFS, fs.ReadFileFS, and fs.StatFS.
type MergedMigrationsFS struct {
	templateFS embed.FS
	domainFS   embed.FS
}

// NewMergedMigrationsFS creates a merged filesystem combining:
//   - Template service migrations (1001-1999): session management, barrier,
//     realms, multi-tenancy.
//   - Domain-specific migrations (2001+): application-specific tables.
//
// Domain migrations take priority (tried first), falling back to template.
// Used by all services to apply the complete migration set as a single sequence.
//
// Parameters:
//   - domainFS: Embedded filesystem containing domain-specific migrations (2001+).
//
// Returns a *MergedMigrationsFS implementing fs.FS, fs.ReadDirFS, fs.ReadFileFS, and fs.StatFS.
func NewMergedMigrationsFS(domainFS embed.FS) *MergedMigrationsFS {
	return &MergedMigrationsFS{
		templateFS: MigrationsFS,
		domainFS:   domainFS,
	}
}

// Open implements fs.FS.
// Domain migrations are tried first (2001+), falling back to template (1001-1999).
func (m *MergedMigrationsFS) Open(name string) (fs.File, error) {
	// Try domain migrations first (2001+).
	if f, err := m.domainFS.Open(name); err == nil {
		return f, nil
	}

	// Fall back to template migrations (1001-1999).
	f, err := m.templateFS.Open(name)
	if err != nil {
		return nil, fmt.Errorf("failed to open from template: %w", err)
	}

	return f, nil
}

// ReadDir implements fs.ReadDirFS.
// Merges entries from both template (1001-1999) and domain (2001+) filesystems.
func (m *MergedMigrationsFS) ReadDir(name string) ([]fs.DirEntry, error) {
	var entries []fs.DirEntry

	// Read template migrations (1001-1999).
	if templateEntries, err := m.templateFS.ReadDir(name); err == nil {
		entries = append(entries, templateEntries...)
	}

	// Read domain migrations (2001+).
	if domainEntries, err := m.domainFS.ReadDir(name); err == nil {
		entries = append(entries, domainEntries...)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("directory not found: %s", name)
	}

	return entries, nil
}

// ReadFile implements fs.ReadFileFS.
// Domain migrations are tried first (2001+), falling back to template (1001-1999).
func (m *MergedMigrationsFS) ReadFile(name string) ([]byte, error) {
	// Try domain migrations first (2001+).
	if data, err := fs.ReadFile(m.domainFS, name); err == nil {
		return data, nil
	}

	// Fall back to template migrations (1001-1999).
	data, err := fs.ReadFile(m.templateFS, name)
	if err != nil {
		return nil, fmt.Errorf("failed to read from template: %w", err)
	}

	return data, nil
}

// Stat implements fs.StatFS.
// Domain migrations are tried first (2001+), falling back to template (1001-1999).
func (m *MergedMigrationsFS) Stat(name string) (fs.FileInfo, error) {
	// Try domain migrations first (2001+).
	if info, err := fs.Stat(m.domainFS, name); err == nil {
		return info, nil
	}

	// Fall back to template migrations (1001-1999).
	info, err := fs.Stat(m.templateFS, name)
	if err != nil {
		return nil, fmt.Errorf("failed to stat from template: %w", err)
	}

	return info, nil
}
