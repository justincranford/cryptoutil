// Copyright (c) 2025-2026 Justin Cranford.
package apps_ps_id_template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// findProjectRoot traverses up from the current directory to locate go.mod.
func findProjectRoot() (string, error) {
	dir, _ := os.Getwd()

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

// copyManifest copies the real PS-ID template directory into a synthetic root directory.
func copyManifest(t *testing.T, realRoot, tmpDir string) {
	t.Helper()

	templateDir := filepath.Join(realRoot, "api", "cryptosuite-registry", "templates", "internal", "apps", cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID)
	destDir := filepath.Join(tmpDir, "api", "cryptosuite-registry", "templates", "internal", "apps", cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID)

	require.NoError(t, filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, relErr := filepath.Rel(templateDir, path)
		if relErr != nil {
			return relErr
		}

		targetPath := filepath.Join(destDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		return os.WriteFile(targetPath, data, cryptoutilSharedMagic.CacheFilePermissions)
	}))
}

// createFullPSIDRoot creates all required files for all PS-IDs according to the manifest +
// per-PS-ID exclusion maps (matching the production exclusions).
func createFullPSIDRoot(t *testing.T, realRoot, tmpDir string) {
	t.Helper()

	copyManifest(t, realRoot, tmpDir)

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		serverDir := filepath.Join(psDir, "server")

		require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		// Canonical template-enforced root/client files.
		realServiceRootPath := filepath.Join(realRoot, "internal", "apps", ps.PSID, ps.Service+".go")
		realServiceRootContent, err := os.ReadFile(realServiceRootPath)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+".go"), realServiceRootContent, cryptoutilSharedMagic.CacheFilePermissions))

		canonicalFiles := []string{
			ps.Service + "_usage.go",
			ps.Service + "_cli_test.go",
			"testmain_test.go",
			"README.md",
			filepath.Join("client", "client.go"),
			filepath.Join("server", ps.Service+"_port_conflict_test.go"),
		}

		for _, rel := range canonicalFiles {
			srcPath := filepath.Join(realRoot, "internal", "apps", ps.PSID, rel)
			srcData, readErr := os.ReadFile(srcPath)
			require.NoError(t, readErr)

			dstPath := filepath.Join(psDir, rel)
			require.NoError(t, os.MkdirAll(filepath.Dir(dstPath), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			require.NoError(t, os.WriteFile(dstPath, srcData, cryptoutilSharedMagic.CacheFilePermissions))
		}

		// Required server file: server.go (no exclusions).
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		// swagger.go, swagger_test.go, and port_conflict_test are excluded, skip.
		if !knownServerFileExclusions["__SERVICE___lifecycle_test.go"][ps.PSID] {
			require.NoError(t, os.WriteFile(
				filepath.Join(serverDir, ps.Service+"_lifecycle_test.go"),
				[]byte("package server\n"),
				cryptoutilSharedMagic.CacheFilePermissions,
			))
		}

		// Required server subdirs: apis, model, repository (respecting production exclusions).
		for _, dir := range []string{"apis", "model", "repository"} {
			if !knownServerDirExclusions[dir][ps.PSID] {
				require.NoError(t, os.MkdirAll(filepath.Join(serverDir, dir), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			}
		}

		require.NoError(t, os.MkdirAll(filepath.Join(psDir, "client"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(psDir, "e2e"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, "e2e", "testmain_e2e_test.go"), []byte("package e2e_test\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, "e2e", ps.Service+"_e2e_test.go"), []byte("package e2e_test\n"), cryptoutilSharedMagic.CacheFilePermissions))

		// Required server config files (respecting production exclusions).
		if !knownServerConfigFileExclusions["config.go"][ps.PSID] {
			configDir := filepath.Join(serverDir, "config")
			require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
			require.NoError(t, os.WriteFile(filepath.Join(configDir, "config_test.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}

		if !knownServerConfigFileExclusions["config_test_helper.go"][ps.PSID] {
			configDir := filepath.Join(serverDir, "config")
			require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			require.NoError(t, os.WriteFile(filepath.Join(configDir, "config_test_helper.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}

		// Required server repository files and dirs (respecting production exclusions).
		if !knownServerRepositoryFileExclusions["migrations.go"][ps.PSID] {
			repoDir := filepath.Join(serverDir, "repository")
			require.NoError(t, os.MkdirAll(filepath.Join(repoDir, "migrations"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			require.NoError(t, os.WriteFile(filepath.Join(repoDir, "migrations.go"), []byte("package repository\n"), cryptoutilSharedMagic.CacheFilePermissions))
		}
	}
}

// TestCheck_RealWorkspace verifies the linter passes against the actual workspace.
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

// Sequential: uses os.Chdir (global process state, cannot run in parallel).
func TestCheck_Integration(t *testing.T) {
	root, err := findProjectRoot()
	if err != nil {
		t.Skip("Skipping integration test - cannot find project root (no go.mod)")
	}

	origDir, getErr := os.Getwd()
	require.NoError(t, getErr)

	require.NoError(t, os.Chdir(root))

	defer func() {
		require.NoError(t, os.Chdir(origDir))
	}()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = Check(logger)
	require.NoError(t, err)
}

// TestCheckInDir_NoManifest exercises the "manifest not found" error path.
func TestCheckInDir_NoManifest(t *testing.T) {
	t.Parallel()

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, t.TempDir())
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read PS-ID MANIFEST.yaml")
}

// TestCheckInDir_InvalidManifest exercises the YAML parse error path.
func TestCheckInDir_InvalidManifest(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	manifestDir := filepath.Join(tmpDir, "api", "cryptosuite-registry", "templates", "internal", "apps", cryptoutilSharedMagic.CICDTemplateExpansionKeyPSID)

	require.NoError(t, os.MkdirAll(manifestDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
	require.NoError(t, os.WriteFile(filepath.Join(manifestDir, "MANIFEST.yaml"), []byte(":\tinvalid::yaml{"), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err := CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to parse PS-ID MANIFEST.yaml")
}

// TestCheckInDir_NoAppsDir exercises the "internal/apps not found" error path.
func TestCheckInDir_NoAppsDir(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	// Use a tmpDir that has the MANIFEST (borrowed from real root) but no internal/apps/.
	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "internal/apps directory not found")
}

// TestCheckInDir_WithExclusions_AllPass verifies the linter passes when all non-excluded
// PS-IDs have their required files.
func TestCheckInDir_WithExclusions_AllPass(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	createFullPSIDRoot(t, root, tmpDir)

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, tmpDir)
	require.NoError(t, err)
}

// TestCheckInDir_WithExclusions_ServiceRootTemplateMismatch verifies template-content enforcement
// for canonical root templates.
func TestCheckInDir_WithExclusions_ServiceRootTemplateMismatch(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	createFullPSIDRoot(t, root, tmpDir)

	servicePath := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.OTLPServicePKICA, "ca_cli_test.go")
	serviceContent, readErr := os.ReadFile(servicePath)
	require.NoError(t, readErr)

	require.NoError(t, os.WriteFile(servicePath, append(serviceContent, []byte("\n// mismatch\n")...), cryptoutilSharedMagic.CacheFilePermissions))

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = CheckInDir(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not match canonical template")
}

// TestCheckInDir_NoExclusions_MissingRootFile exercises the root-file violation path.
func TestCheckInDir_NoExclusions_MissingRootFile(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	// Create server dirs for all PS-IDs but omit all root files.
	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		require.NoError(t, os.MkdirAll(
			filepath.Join(tmpDir, "internal", "apps", ps.PSID, "server"),
			cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute,
		))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = ExportedCheckInDirNoExclusions(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required root file")
}

// TestCheckInDir_NoExclusions_MissingRequiredDir exercises the required-dir violation path.
func TestCheckInDir_NoExclusions_MissingRequiredDir(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	// Create PS-ID dirs with root files but no server/ dir.
	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		require.NoError(t, os.MkdirAll(psDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+".go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_usage.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_cli_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = ExportedCheckInDirNoExclusions(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required directory")
}

// TestCheckInDir_NoExclusions_MissingServerFile exercises the server-file violation path.
func TestCheckInDir_NoExclusions_MissingServerFile(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	// Create full root files and server/ dir but no server files.
	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		serverDir := filepath.Join(psDir, "server")
		require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+".go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_usage.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_cli_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		// No server files created.
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = ExportedCheckInDirNoExclusions(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required server file")
}

// TestCheckInDir_NoExclusions_AllValid verifies no violations when all required files present.
func TestCheckInDir_NoExclusions_AllValid(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		serverDir := filepath.Join(psDir, "server")
		configDir := filepath.Join(serverDir, "config")
		repoDir := filepath.Join(serverDir, "repository")

		require.NoError(t, os.MkdirAll(filepath.Join(serverDir, "apis"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(serverDir, "model"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(psDir, "client"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(psDir, "e2e"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, "e2e", "testmain_e2e_test.go"), []byte("package e2e_test\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, "e2e", ps.Service+"_e2e_test.go"), []byte("package e2e_test\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(repoDir, "migrations"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		// All required root files.
		realServiceRootPath := filepath.Join(root, "internal", "apps", ps.PSID, ps.Service+".go")
		realServiceRootContent, readErr := os.ReadFile(realServiceRootPath)
		require.NoError(t, readErr)
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+".go"), realServiceRootContent, cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_usage.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_cli_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))

		// All required server files.
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "swagger.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "swagger_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, ps.Service+"_lifecycle_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, ps.Service+"_port_conflict_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))

		// All required server config files.
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config_test.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config_test_helper.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))

		// All required server repository files.
		require.NoError(t, os.WriteFile(filepath.Join(repoDir, "migrations.go"), []byte("package repository\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = ExportedCheckInDirNoExclusions(logger, tmpDir)
	require.NoError(t, err)
}

// TestCheckInDir_NoExclusions_MissingServerDir exercises the server subdirectory violation path.
func TestCheckInDir_NoExclusions_MissingServerDir(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	// Create root files + server/ + required server files + config + repository files, but omit server/apis/.
	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		serverDir := filepath.Join(psDir, "server")
		configDir := filepath.Join(serverDir, "config")
		repoDir := filepath.Join(serverDir, "repository")

		require.NoError(t, os.MkdirAll(filepath.Join(serverDir, "model"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(repoDir, "migrations"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+".go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_usage.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_cli_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config_test.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config_test_helper.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(repoDir, "migrations.go"), []byte("package repository\n"), cryptoutilSharedMagic.CacheFilePermissions))
		// No server/apis/ directory created.
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = ExportedCheckInDirNoExclusions(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required server subdirectory")
}

// TestCheckInDir_NoExclusions_MissingServerConfigFile exercises the server config file violation path.
func TestCheckInDir_NoExclusions_MissingServerConfigFile(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	// Create everything except server/config/config_test_helper.go.
	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		serverDir := filepath.Join(psDir, "server")
		configDir := filepath.Join(serverDir, "config")
		repoDir := filepath.Join(serverDir, "repository")

		require.NoError(t, os.MkdirAll(filepath.Join(serverDir, "apis"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(serverDir, "model"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(repoDir, "migrations"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+".go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_usage.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_cli_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config_test.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		// config_test_helper.go intentionally omitted.
		require.NoError(t, os.WriteFile(filepath.Join(repoDir, "migrations.go"), []byte("package repository\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = ExportedCheckInDirNoExclusions(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required server config file")
}

// TestCheckInDir_NoExclusions_MissingServerRepositoryFile exercises the server repository file violation path.
func TestCheckInDir_NoExclusions_MissingServerRepositoryFile(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	// Create everything except server/repository/migrations.go.
	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		serverDir := filepath.Join(psDir, "server")
		configDir := filepath.Join(serverDir, "config")
		repoDir := filepath.Join(serverDir, "repository")

		require.NoError(t, os.MkdirAll(filepath.Join(serverDir, "apis"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(serverDir, "model"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(repoDir, "migrations"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+".go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_usage.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_cli_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config_test.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config_test_helper.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		// migrations.go intentionally omitted.
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = ExportedCheckInDirNoExclusions(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required server repository file")
}

// TestCheckInDir_NoExclusions_MissingServerRepositoryDir exercises the server repository subdirectory violation path.
func TestCheckInDir_NoExclusions_MissingServerRepositoryDir(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	// Create everything except server/repository/migrations/ subdirectory.
	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		serverDir := filepath.Join(psDir, "server")
		configDir := filepath.Join(serverDir, "config")
		repoDir := filepath.Join(serverDir, "repository")

		require.NoError(t, os.MkdirAll(filepath.Join(serverDir, "apis"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(serverDir, "model"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(repoDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		// migrations/ subdirectory intentionally omitted.

		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+".go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_usage.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_cli_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config_test.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config_test_helper.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(repoDir, "migrations.go"), []byte("package repository\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = ExportedCheckInDirNoExclusions(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required server repository subdirectory")
}

// TestCheckInDir_NoExclusions_MissingE2EFile exercises the e2e file violation path.
// Only fires when the e2e/ directory exists but required files are absent.
func TestCheckInDir_NoExclusions_MissingE2EFile(t *testing.T) {
	t.Parallel()

	root, err := findProjectRoot()
	if err != nil {
		t.Skip("cannot find project root")
	}

	tmpDir := t.TempDir()
	copyManifest(t, root, tmpDir)

	// Create a full valid PS-ID structure but add an e2e/ dir without required files for one PS-ID.
	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		serverDir := filepath.Join(psDir, "server")
		configDir := filepath.Join(serverDir, "config")
		repoDir := filepath.Join(serverDir, "repository")

		require.NoError(t, os.MkdirAll(filepath.Join(serverDir, "apis"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(serverDir, "model"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		require.NoError(t, os.MkdirAll(filepath.Join(repoDir, "migrations"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		// Create e2e/ dir but intentionally omit required e2e files.
		require.NoError(t, os.MkdirAll(filepath.Join(psDir, "e2e"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+".go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_usage.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_cli_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(serverDir, "testmain_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config_test.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config_test_helper.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(repoDir, "migrations.go"), []byte("package repository\n"), cryptoutilSharedMagic.CacheFilePermissions))
	}

	logger := cryptoutilCmdCicdCommon.NewLogger("test")
	err = ExportedCheckInDirNoExclusions(logger, tmpDir)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required e2e file")
}
