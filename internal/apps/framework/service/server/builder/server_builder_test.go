// Copyright (c) 2025 Justin Cranford
//
//

package builder

import (
	"context"
	"io/fs"
	"testing"
	"testing/fstest"

	cryptoutilAppsFrameworkServiceConfig "cryptoutil/internal/apps/framework/service/config"
	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps/framework/service/server"

	"github.com/stretchr/testify/require"
)

// TestNewServerBuilder_Success tests successful server builder creation.
func TestNewServerBuilder_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	builder := NewServerBuilder(ctx, settings)

	require.NotNil(t, builder)
	require.Nil(t, builder.err)
	require.Equal(t, ctx, builder.ctx)
	require.Equal(t, settings, builder.config)
}

// TestNewServerBuilder_ValidationErrors tests error handling for invalid inputs using table-driven pattern.
func TestNewServerBuilder_ValidationErrors(t *testing.T) {
	t.Parallel()

	settings := getMinimalSettings()

	tests := []struct {
		name          string
		ctx           context.Context
		config        *cryptoutilAppsFrameworkServiceConfig.ServiceFrameworkServerSettings
		wantErrSubstr string
	}{
		{
			name:          "nil context",
			ctx:           nil,
			config:        settings,
			wantErrSubstr: "context cannot be nil",
		},
		{
			name:          "nil config",
			ctx:           context.Background(),
			config:        nil,
			wantErrSubstr: "config cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := NewServerBuilder(tt.ctx, tt.config)

			require.NotNil(t, builder)
			require.Error(t, builder.err)
			require.Contains(t, builder.err.Error(), tt.wantErrSubstr)
		})
	}
}

// TestWithDomainMigrations_Success tests successful domain migration registration.
func TestWithDomainMigrations_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	// Create test migration filesystem.
	migrationFS := fstest.MapFS{
		"migrations/2001_test.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE test (id TEXT);"),
		},
		"migrations/2001_test.down.sql": &fstest.MapFile{
			Data: []byte("DROP TABLE test;"),
		},
	}

	builder := NewServerBuilder(ctx, settings).
		WithDomainMigrations(migrationFS, "migrations")

	require.NotNil(t, builder)
	require.Nil(t, builder.err)
	require.NotNil(t, builder.migrationFS)
	require.Equal(t, "migrations", builder.migrationsPath)
}

// TestWithDomainMigrations_ValidationErrors tests error handling using table-driven pattern.
func TestWithDomainMigrations_ValidationErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	migrationFS := fstest.MapFS{
		"migrations/2001_test.up.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE test (id TEXT);"),
		},
	}

	tests := []struct {
		name          string
		fs            fs.FS
		path          string
		wantErrSubstr string
	}{
		{
			name:          "nil filesystem",
			fs:            nil,
			path:          "migrations",
			wantErrSubstr: "migration FS cannot be nil",
		},
		{
			name:          "empty path",
			fs:            migrationFS,
			path:          "",
			wantErrSubstr: "migrations path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := NewServerBuilder(ctx, settings).
				WithDomainMigrations(tt.fs, tt.path)

			require.NotNil(t, builder)
			require.Error(t, builder.err)
			require.Contains(t, builder.err.Error(), tt.wantErrSubstr)
		})
	}
}

// TestWithPublicRouteRegistration_Success tests successful route registration function.
func TestWithPublicRouteRegistration_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	registerFunc := func(_ *cryptoutilAppsFrameworkServiceServer.PublicServerBase, _ *ServiceResources) error {
		return nil
	}

	builder := NewServerBuilder(ctx, settings).
		WithPublicRouteRegistration(registerFunc)

	require.NotNil(t, builder)
	require.Nil(t, builder.err)
	require.NotNil(t, builder.publicRouteRegister)
}

// TestWithPublicRouteRegistration_NilFunc tests error handling for nil registration function.
func TestWithPublicRouteRegistration_NilFunc(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	settings := getMinimalSettings()

	builder := NewServerBuilder(ctx, settings).
		WithPublicRouteRegistration(nil)

	require.NotNil(t, builder)
	require.Error(t, builder.err)
	require.Contains(t, builder.err.Error(), "route registration function cannot be nil")
}

// TestErrorAccumulation tests that fluent chain short-circuits on first error.
func TestErrorAccumulation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name             string
		buildFn          func() *ServerBuilder
		wantErrSubstr    string
		notWantErrSubstr string
	}{
		{
			name: "domain migrations after nil config",
			buildFn: func() *ServerBuilder {
				return NewServerBuilder(ctx, nil).
					WithDomainMigrations(fstest.MapFS{}, "migrations")
			},
			wantErrSubstr:    "config cannot be nil",
			notWantErrSubstr: "migration",
		},
		{
			name: "route registration after nil config",
			buildFn: func() *ServerBuilder {
				registerFunc := func(_ *cryptoutilAppsFrameworkServiceServer.PublicServerBase, _ *ServiceResources) error {
					return nil
				}

				return NewServerBuilder(ctx, nil).
					WithPublicRouteRegistration(registerFunc)
			},
			wantErrSubstr:    "config cannot be nil",
			notWantErrSubstr: "route registration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := tt.buildFn()

			require.NotNil(t, builder)
			require.Error(t, builder.err)
			require.Contains(t, builder.err.Error(), tt.wantErrSubstr)
			require.NotContains(t, builder.err.Error(), tt.notWantErrSubstr)
		})
	}
}
