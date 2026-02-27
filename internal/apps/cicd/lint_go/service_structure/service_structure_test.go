// Copyright (c) 2025 Justin Cranford

package service_structure

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps/cicd/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testProductSkeleton = "skeleton"
	testServiceTemplate = "template"
)

// findProjectRoot finds the project root by looking for go.mod.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}

		dir = parent
	}
}

func TestCheck_RealWorkspace(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, root)
	require.NoError(t, err)
}

func TestCheckInDir_AllValid(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create all 8 services with required files.
	for _, svc := range knownServices {
		serviceDir := filepath.Join(tmpDir, "internal", "apps", svc.Product, svc.Service)
		require.NoError(t, os.MkdirAll(filepath.Join(serviceDir, "server", "config"), 0o755))

		required := svc.Required
		if required == nil {
			required = defaultRequiredFiles
		}

		for _, tmpl := range required {
			relPath := filepath.Join(serviceDir, replaceService(tmpl, svc.Service))
			require.NoError(t, os.MkdirAll(filepath.Dir(relPath), 0o755))
			require.NoError(t, os.WriteFile(relPath, []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

func TestCheckInDir_MissingServiceDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "internal", "apps"), 0o755))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "service directory missing")
}

func TestCheckInDir_MissingEntryFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	serviceDir := filepath.Join(tmpDir, "internal", "apps", testProductSkeleton, testServiceTemplate)
	require.NoError(t, os.MkdirAll(filepath.Join(serviceDir, "server", "config"), 0o755))

	// Create all required files except the entry file.
	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, "template_usage.go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, "server", "server.go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, "server", "config", "config.go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))

	// Create all other services with all files to isolate the test.
	for _, svc := range knownServices {
		if svc.Product == testProductSkeleton && svc.Service == testServiceTemplate {
			continue
		}

		svcDir := filepath.Join(tmpDir, "internal", "apps", svc.Product, svc.Service)

		required := svc.Required
		if required == nil {
			required = defaultRequiredFiles
		}

		for _, tmpl := range required {
			relPath := filepath.Join(svcDir, replaceService(tmpl, svc.Service))
			require.NoError(t, os.MkdirAll(filepath.Dir(relPath), 0o755))
			require.NoError(t, os.WriteFile(relPath, []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required file template.go")
}

func TestCheckInDir_MissingUsageFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	serviceDir := filepath.Join(tmpDir, "internal", "apps", testProductSkeleton, testServiceTemplate)
	require.NoError(t, os.MkdirAll(filepath.Join(serviceDir, "server", "config"), 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, "template.go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, "server", "server.go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, "server", "config", "config.go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))

	for _, svc := range knownServices {
		if svc.Product == testProductSkeleton && svc.Service == testServiceTemplate {
			continue
		}

		svcDir := filepath.Join(tmpDir, "internal", "apps", svc.Product, svc.Service)

		required := svc.Required
		if required == nil {
			required = defaultRequiredFiles
		}

		for _, tmpl := range required {
			relPath := filepath.Join(svcDir, replaceService(tmpl, svc.Service))
			require.NoError(t, os.MkdirAll(filepath.Dir(relPath), 0o755))
			require.NoError(t, os.WriteFile(relPath, []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required file template_usage.go")
}

func TestCheckInDir_MissingServerGo(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	serviceDir := filepath.Join(tmpDir, "internal", "apps", testProductSkeleton, testServiceTemplate)
	require.NoError(t, os.MkdirAll(filepath.Join(serviceDir, "server", "config"), 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, "template.go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, "template_usage.go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, "server", "config", "config.go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))

	for _, svc := range knownServices {
		if svc.Product == testProductSkeleton && svc.Service == testServiceTemplate {
			continue
		}

		svcDir := filepath.Join(tmpDir, "internal", "apps", svc.Product, svc.Service)

		required := svc.Required
		if required == nil {
			required = defaultRequiredFiles
		}

		for _, tmpl := range required {
			relPath := filepath.Join(svcDir, replaceService(tmpl, svc.Service))
			require.NoError(t, os.MkdirAll(filepath.Dir(relPath), 0o755))
			require.NoError(t, os.WriteFile(relPath, []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required file server/server.go")
}

func TestCheckInDir_MissingConfigGo(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	serviceDir := filepath.Join(tmpDir, "internal", "apps", testProductSkeleton, testServiceTemplate)
	require.NoError(t, os.MkdirAll(filepath.Join(serviceDir, "server"), 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, "template.go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, "template_usage.go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(serviceDir, "server", "server.go"), []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))

	for _, svc := range knownServices {
		if svc.Product == testProductSkeleton && svc.Service == testServiceTemplate {
			continue
		}

		svcDir := filepath.Join(tmpDir, "internal", "apps", svc.Product, svc.Service)

		required := svc.Required
		if required == nil {
			required = defaultRequiredFiles
		}

		for _, tmpl := range required {
			relPath := filepath.Join(svcDir, replaceService(tmpl, svc.Service))
			require.NoError(t, os.MkdirAll(filepath.Dir(relPath), 0o755))
			require.NoError(t, os.WriteFile(relPath, []byte("package x\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required file server/config/config.go")
}

func TestCheckInDir_NoAppsDir(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "internal/apps directory not found")
}

// replaceService replaces {SERVICE} placeholder in file templates.
func replaceService(tmpl, service string) string {
	return filepath.FromSlash(filepath.Clean(
		filepath.Join(filepath.SplitList(
			filepath.ToSlash(
				strings.ReplaceAll(tmpl, "{SERVICE}", service),
			),
		)...),
	))
}

// Sequential: uses os.Chdir (global process state).
func TestCheck_FromProjectRoot(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping - cannot find project root")
	}

	origDir, wdErr := os.Getwd()
	require.NoError(t, wdErr)

	require.NoError(t, os.Chdir(root))

	t.Cleanup(func() {
		require.NoError(t, os.Chdir(origDir))
	})

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.NoError(t, err)
}
