//go:build !integration

package builder

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilAppsTemplateServiceConfigTlsGenerator "cryptoutil/internal/apps/template/service/config/tls_generator"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerApplication "cryptoutil/internal/apps/template/service/server/application"
	cryptoutilAppsTemplateServiceServerListener "cryptoutil/internal/apps/template/service/server/listener"
)

// TestBuild_AdminServerCreateError tests Build when NewAdminHTTPServer fails.
// Covers server_builder_build.go L48-50 (admin server error path).
// Cannot use t.Parallel() because it modifies the package-level injectable var.
func TestBuild_AdminServerCreateError(t *testing.T) {
	original := newAdminHTTPServerFn
	newAdminHTTPServerFn = func(
		_ context.Context,
		_ *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings,
		_ *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings,
	) (*cryptoutilAppsTemplateServiceServerListener.AdminServer, error) {
		return nil, fmt.Errorf("mock admin server error")
	}

	defer func() { newAdminHTTPServerFn = original }()

	cfg := getMinimalSettings()
	builder := NewServerBuilder(context.Background(), cfg)

	resources, err := builder.Build()
	require.Nil(t, resources)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create admin server")
	require.Contains(t, err.Error(), "mock admin server error")
}

// TestBuild_StartCoreError tests Build when StartCore fails.
// Covers server_builder_build.go L55-57 (start core error path).
// Cannot use t.Parallel() because it modifies the package-level injectable var.
func TestBuild_StartCoreError(t *testing.T) {
	original := startCoreFn
	startCoreFn = func(
		_ context.Context,
		_ *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings,
	) (*cryptoutilAppsTemplateServiceServerApplication.Core, error) {
		return nil, fmt.Errorf("mock start core error")
	}

	defer func() { startCoreFn = original }()

	cfg := getMinimalSettings()
	builder := NewServerBuilder(context.Background(), cfg)

	resources, err := builder.Build()
	require.Nil(t, resources)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to start application core")
	require.Contains(t, err.Error(), "mock start core error")
}

// TestBuild_InitServicesError tests Build when InitializeServicesOnCore fails.
// Covers server_builder_build.go L88-92 (services init error path).
// Cannot use t.Parallel() because it modifies the package-level injectable var.
func TestBuild_InitServicesError(t *testing.T) {
	original := initServicesOnCoreFn
	initServicesOnCoreFn = func(
		_ context.Context,
		_ *cryptoutilAppsTemplateServiceServerApplication.Core,
		_ *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings,
	) (*cryptoutilAppsTemplateServiceServerApplication.CoreWithServices, error) {
		return nil, fmt.Errorf("mock services init error")
	}

	defer func() { initServicesOnCoreFn = original }()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second)
	defer cancel()

	cfg := getMinimalSettings()
	builder := NewServerBuilder(ctx, cfg)

	resources, err := builder.Build()
	require.Nil(t, resources)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to initialize services on core")
	require.Contains(t, err.Error(), "mock services init error")
}

// TestBuild_GenerateTLSMaterialError tests Build when GenerateTLSMaterial fails for public server.
// Covers server_builder_build.go L114-118 (TLS material generation error).
// Cannot use t.Parallel() because it modifies the package-level injectable var.
func TestBuild_GenerateTLSMaterialError(t *testing.T) {
	original := generateTLSMaterialFn
	generateTLSMaterialFn = func(
		_ *cryptoutilAppsTemplateServiceConfigTlsGenerator.TLSGeneratedSettings,
	) (*cryptoutilAppsTemplateServiceConfig.TLSMaterial, error) {
		return nil, fmt.Errorf("mock TLS material error")
	}

	defer func() { generateTLSMaterialFn = original }()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second)
	defer cancel()

	cfg := getMinimalSettings()
	builder := NewServerBuilder(ctx, cfg)

	resources, err := builder.Build()
	require.Nil(t, resources)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to generate public TLS material")
	require.Contains(t, err.Error(), "mock TLS material error")
}

// TestBuild_PublicServerBaseError tests Build when NewPublicServerBase fails.
// Covers server_builder_build.go L126-130 (public server base creation error).
// Cannot use t.Parallel() because it modifies the package-level injectable var.
func TestBuild_PublicServerBaseError(t *testing.T) {
	original := newPublicServerBaseFn
	newPublicServerBaseFn = func(
		_ *cryptoutilAppsTemplateServiceServer.PublicServerConfig,
	) (*cryptoutilAppsTemplateServiceServer.PublicServerBase, error) {
		return nil, fmt.Errorf("mock public server base error")
	}

	defer func() { newPublicServerBaseFn = original }()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second)
	defer cancel()

	cfg := getMinimalSettings()
	builder := NewServerBuilder(ctx, cfg)

	resources, err := builder.Build()
	require.Nil(t, resources)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create public server base")
	require.Contains(t, err.Error(), "mock public server base error")
}

// TestBuild_NewApplicationError tests Build when NewApplication fails.
// Covers server_builder_build.go L202-206 (application creation error).
// Cannot use t.Parallel() because it modifies the package-level injectable var.
func TestBuild_NewApplicationError(t *testing.T) {
	original := newApplicationFn
	newApplicationFn = func(
		_ context.Context,
		_ cryptoutilAppsTemplateServiceServer.IPublicServer,
		_ cryptoutilAppsTemplateServiceServer.IAdminServer,
	) (*cryptoutilAppsTemplateServiceServer.Application, error) {
		return nil, fmt.Errorf("mock application error")
	}

	defer func() { newApplicationFn = original }()

	ctx, cancel := context.WithTimeout(context.Background(), cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days*time.Second)
	defer cancel()

	cfg := getMinimalSettings()
	builder := NewServerBuilder(ctx, cfg)

	resources, err := builder.Build()
	require.Nil(t, resources)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to create application")
	require.Contains(t, err.Error(), "mock application error")
}

// TestBuild_GormDBError tests Build when applicationCore.DB.DB() fails.
// Covers server_builder_build.go L73-77 (sql.DB from GORM error path).
// Cannot use t.Parallel() because it modifies the package-level injectable var.
// Sequential: modifies package-level injectable function variable.
func TestBuild_GormDBError(t *testing.T) {
	original := startCoreFn
	startCoreFn = func(
		_ context.Context,
		_ *cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings,
	) (*cryptoutilAppsTemplateServiceServerApplication.Core, error) {
		// Return core with fake GORM DB that has non-sql.DB ConnPool.
		// Calling DB.DB() on this returns gorm.ErrInvalidDB.
		fakeDB := &gorm.DB{
			Config: &gorm.Config{
				ConnPool: &fakeConnPool{},
			},
		}

		return &cryptoutilAppsTemplateServiceServerApplication.Core{
			DB:                  fakeDB,
			Basic:               &cryptoutilAppsTemplateServiceServerApplication.Basic{},
			ShutdownDBContainer: func() {},
		}, nil
	}

	defer func() { startCoreFn = original }()

	cfg := getMinimalSettings()
	builder := NewServerBuilder(context.Background(), cfg)

	resources, err := builder.Build()
	require.Nil(t, resources)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get sql.DB from GORM")
}
