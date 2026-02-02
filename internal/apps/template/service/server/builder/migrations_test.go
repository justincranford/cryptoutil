// Copyright (c) 2025 Justin Cranford
//

package builder

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

func TestNewDefaultMigrationConfig(t *testing.T) {
	t.Parallel()

	config := NewDefaultMigrationConfig()

	require.NotNil(t, config)
	require.Equal(t, MigrationModeTemplateWithDomain, config.Mode)
	require.False(t, config.SkipTemplateMigrations)
	require.Nil(t, config.DomainFS)
	require.Empty(t, config.DomainPath)
}

func TestNewDomainOnlyMigrationConfig(t *testing.T) {
	t.Parallel()

	config := NewDomainOnlyMigrationConfig()

	require.NotNil(t, config)
	require.Equal(t, MigrationModeDomainOnly, config.Mode)
	require.True(t, config.SkipTemplateMigrations)
	require.Nil(t, config.DomainFS)
	require.Empty(t, config.DomainPath)
}

func TestNewDisabledMigrationConfig(t *testing.T) {
	t.Parallel()

	config := NewDisabledMigrationConfig()

	require.NotNil(t, config)
	require.Equal(t, MigrationModeDisabled, config.Mode)
	require.True(t, config.SkipTemplateMigrations)
	require.Nil(t, config.DomainFS)
	require.Empty(t, config.DomainPath)
}

func TestMigrationConfig_Validate_TemplateWithDomain_Success(t *testing.T) {
	t.Parallel()

	mockFS := fstest.MapFS{
		"migrations/0001_init.up.sql": &fstest.MapFile{Data: []byte("CREATE TABLE test;")},
	}

	config := NewDefaultMigrationConfig().
		WithDomainFS(mockFS).
		WithDomainPath("migrations")

	err := config.Validate()
	require.NoError(t, err)
}

func TestMigrationConfig_Validate_TemplateWithDomain_MissingFS(t *testing.T) {
	t.Parallel()

	config := NewDefaultMigrationConfig().
		WithDomainPath("migrations")

	err := config.Validate()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrMigrationFSRequired)
}

func TestMigrationConfig_Validate_TemplateWithDomain_MissingPath(t *testing.T) {
	t.Parallel()

	mockFS := fstest.MapFS{
		"migrations/0001_init.up.sql": &fstest.MapFile{Data: []byte("CREATE TABLE test;")},
	}

	config := NewDefaultMigrationConfig().
		WithDomainFS(mockFS)

	err := config.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "domain path is required")
}

func TestMigrationConfig_Validate_DomainOnly_Success(t *testing.T) {
	t.Parallel()

	mockFS := fstest.MapFS{
		"migrations/0001_init.up.sql": &fstest.MapFile{Data: []byte("CREATE TABLE test;")},
	}

	config := NewDomainOnlyMigrationConfig().
		WithDomainFS(mockFS).
		WithDomainPath("migrations")

	err := config.Validate()
	require.NoError(t, err)
}

func TestMigrationConfig_Validate_DomainOnly_MissingFS(t *testing.T) {
	t.Parallel()

	config := NewDomainOnlyMigrationConfig().
		WithDomainPath("migrations")

	err := config.Validate()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrMigrationFSRequired)
}

func TestMigrationConfig_Validate_DomainOnly_MissingPath(t *testing.T) {
	t.Parallel()

	mockFS := fstest.MapFS{
		"migrations/0001_init.up.sql": &fstest.MapFile{Data: []byte("CREATE TABLE test;")},
	}

	config := NewDomainOnlyMigrationConfig().
		WithDomainFS(mockFS)

	err := config.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "domain path is required")
}

func TestMigrationConfig_Validate_Disabled_Success(t *testing.T) {
	t.Parallel()

	config := NewDisabledMigrationConfig()

	err := config.Validate()
	require.NoError(t, err)
}

func TestMigrationConfig_Validate_EmptyMode(t *testing.T) {
	t.Parallel()

	config := &MigrationConfig{
		Mode: "",
	}

	err := config.Validate()
	require.Error(t, err)
	require.ErrorIs(t, err, ErrMigrationModeRequired)
}

func TestMigrationConfig_Validate_InvalidMode(t *testing.T) {
	t.Parallel()

	config := &MigrationConfig{
		Mode: "invalid-mode",
	}

	err := config.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid migration mode")
}

func TestMigrationConfig_WithDomainFS(t *testing.T) {
	t.Parallel()

	mockFS := fstest.MapFS{
		"migrations/0001_init.up.sql": &fstest.MapFile{Data: []byte("CREATE TABLE test;")},
	}

	config := NewDefaultMigrationConfig()
	require.Nil(t, config.DomainFS)

	result := config.WithDomainFS(mockFS)

	require.Same(t, config, result)
	require.NotNil(t, config.DomainFS)
}

func TestMigrationConfig_WithDomainPath(t *testing.T) {
	t.Parallel()

	config := NewDefaultMigrationConfig()
	require.Empty(t, config.DomainPath)

	result := config.WithDomainPath("migrations")

	require.Same(t, config, result)
	require.Equal(t, "migrations", config.DomainPath)
}

func TestMigrationConfig_WithMode(t *testing.T) {
	t.Parallel()

	config := NewDefaultMigrationConfig()
	require.Equal(t, MigrationModeTemplateWithDomain, config.Mode)

	result := config.WithMode(MigrationModeDomainOnly)

	require.Same(t, config, result)
	require.Equal(t, MigrationModeDomainOnly, config.Mode)
}

func TestMigrationConfig_WithSkipTemplateMigrations(t *testing.T) {
	t.Parallel()

	config := NewDefaultMigrationConfig()
	require.False(t, config.SkipTemplateMigrations)

	result := config.WithSkipTemplateMigrations(true)

	require.Same(t, config, result)
	require.True(t, config.SkipTemplateMigrations)

	result = config.WithSkipTemplateMigrations(false)

	require.Same(t, config, result)
	require.False(t, config.SkipTemplateMigrations)
}

func TestMigrationConfig_IsEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   *MigrationConfig
		expected bool
	}{
		{
			name:     "template with domain is enabled",
			config:   NewDefaultMigrationConfig(),
			expected: true,
		},
		{
			name:     "domain only is enabled",
			config:   NewDomainOnlyMigrationConfig(),
			expected: true,
		},
		{
			name:     "disabled mode is not enabled",
			config:   NewDisabledMigrationConfig(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expected, tt.config.IsEnabled())
		})
	}
}

func TestMigrationConfig_RequiresTemplateMigrations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   *MigrationConfig
		expected bool
	}{
		{
			name:     "template with domain requires template migrations",
			config:   NewDefaultMigrationConfig(),
			expected: true,
		},
		{
			name:     "template with domain and skip flag does not require template migrations",
			config:   NewDefaultMigrationConfig().WithSkipTemplateMigrations(true),
			expected: false,
		},
		{
			name:     "domain only does not require template migrations",
			config:   NewDomainOnlyMigrationConfig(),
			expected: false,
		},
		{
			name:     "disabled does not require template migrations",
			config:   NewDisabledMigrationConfig(),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expected, tt.config.RequiresTemplateMigrations())
		})
	}
}

func TestMigrationConfig_FluentChaining(t *testing.T) {
	t.Parallel()

	mockFS := fstest.MapFS{
		"migrations/0001_init.up.sql": &fstest.MapFile{Data: []byte("CREATE TABLE test;")},
	}

	config := NewDefaultMigrationConfig().
		WithMode(MigrationModeDomainOnly).
		WithDomainFS(mockFS).
		WithDomainPath("migrations").
		WithSkipTemplateMigrations(true)

	require.Equal(t, MigrationModeDomainOnly, config.Mode)
	require.NotNil(t, config.DomainFS)
	require.Equal(t, "migrations", config.DomainPath)
	require.True(t, config.SkipTemplateMigrations)
}

func TestMigrationModeConstants(t *testing.T) {
	t.Parallel()

	// Verify constant values don't change unexpectedly.
	require.Equal(t, MigrationMode("template_with_domain"), MigrationModeTemplateWithDomain)
	require.Equal(t, MigrationMode("domain_only"), MigrationModeDomainOnly)
	require.Equal(t, MigrationMode("disabled"), MigrationModeDisabled)
}

func TestErrMigrationModeRequired(t *testing.T) {
	t.Parallel()

	require.NotNil(t, ErrMigrationModeRequired)
	require.Equal(t, "migration mode is required", ErrMigrationModeRequired.Error())
}

func TestErrMigrationFSRequired(t *testing.T) {
	t.Parallel()

	require.NotNil(t, ErrMigrationFSRequired)
	require.Equal(t, "migration FS is required for this mode", ErrMigrationFSRequired.Error())
}

func TestWithMigrationConfig_NilConfig(t *testing.T) {
	t.Parallel()

	// ServerBuilder should accept nil config without error.
	builder := &ServerBuilder{}
	result := builder.WithMigrationConfig(nil)

	require.Same(t, builder, result)
	require.NoError(t, builder.err)
	require.Nil(t, builder.migrationConfig)
}

func TestWithMigrationConfig_ValidConfig(t *testing.T) {
	t.Parallel()

	mockFS := fstest.MapFS{
		"migrations/0001_init.up.sql": &fstest.MapFile{Data: []byte("CREATE TABLE test;")},
	}

	tests := []struct {
		name   string
		config *MigrationConfig
	}{
		{
			name: "template with domain",
			config: NewDefaultMigrationConfig().
				WithDomainFS(mockFS).
				WithDomainPath("migrations"),
		},
		{
			name: "domain only",
			config: NewDomainOnlyMigrationConfig().
				WithDomainFS(mockFS).
				WithDomainPath("migrations"),
		},
		{
			name:   "disabled",
			config: NewDisabledMigrationConfig(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := &ServerBuilder{}
			result := builder.WithMigrationConfig(tt.config)

			require.Same(t, builder, result)
			require.NoError(t, builder.err)
			require.Same(t, tt.config, builder.migrationConfig)
		})
	}
}

func TestWithMigrationConfig_InvalidConfig(t *testing.T) {
	t.Parallel()

	config := &MigrationConfig{
		Mode: "invalid-mode",
	}

	builder := &ServerBuilder{}
	result := builder.WithMigrationConfig(config)

	require.Same(t, builder, result)
	require.Error(t, builder.err)
	require.Contains(t, builder.err.Error(), "invalid migration config")
}

func TestWithMigrationConfig_SetsLegacyFields(t *testing.T) {
	t.Parallel()

	mockFS := fstest.MapFS{
		"migrations/0001_init.up.sql": &fstest.MapFile{Data: []byte("CREATE TABLE test;")},
	}

	config := NewDefaultMigrationConfig().
		WithDomainFS(mockFS).
		WithDomainPath("migrations")

	builder := &ServerBuilder{}
	result := builder.WithMigrationConfig(config)

	require.Same(t, builder, result)
	require.NoError(t, builder.err)
	require.NotNil(t, builder.migrationFS)
	require.Equal(t, "migrations", builder.migrationsPath)
}

func TestWithMigrationConfig_ErrorAccumulation(t *testing.T) {
	t.Parallel()

	mockFS := fstest.MapFS{
		"migrations/0001_init.up.sql": &fstest.MapFile{Data: []byte("CREATE TABLE test;")},
	}

	// Builder with existing error should not process new config.
	builder := &ServerBuilder{err: ErrMigrationModeRequired}
	result := builder.WithMigrationConfig(NewDefaultMigrationConfig().
		WithDomainFS(mockFS).
		WithDomainPath("migrations"))

	require.Same(t, builder, result)
	require.ErrorIs(t, builder.err, ErrMigrationModeRequired)
	require.Nil(t, builder.migrationConfig)
}
