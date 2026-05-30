// Copyright (c) 2025-2026 Justin Cranford.
package apps_ps_id_template

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilFitnessRegistry "cryptoutil/internal/apps-tools/cicd_lint/lint_fitness/registry"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestCanonicalTemplateFilesUseActualCanonicalPaths(t *testing.T) {
	t.Parallel()

	for _, spec := range canonicalTemplateFiles {
		if strings.HasPrefix(spec.templatePath, "server/") || strings.HasPrefix(spec.templatePath, "client/") {
			continue
		}

		switch spec.templatePath {
		case "__SERVICE__.go.tmpl":
			require.Equal(t, "__SERVICE__.go", spec.actualPath)
		case "__SERVICE___test.go.tmpl":
			require.Equal(t, "__SERVICE___test.go", spec.actualPath)
		default:
			t.Fatalf("unexpected root template path: %s", spec.templatePath)
		}
	}
}

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

	require.NoError(t, os.MkdirAll(destDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

	srcRoot, err := os.OpenRoot(templateDir)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, srcRoot.Close())
	}()

	dstRoot, err := os.OpenRoot(destDir)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, dstRoot.Close())
	}()

	require.NoError(t, fs.WalkDir(srcRoot.FS(), ".", func(relPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			if relPath == "." {
				return nil
			}

			return dstRoot.MkdirAll(relPath, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute)
		}

		data, readErr := srcRoot.ReadFile(relPath)
		if readErr != nil {
			return readErr
		}

		return dstRoot.WriteFile(relPath, data, cryptoutilSharedMagic.CacheFilePermissions)
	}))
}

// createFullPSIDRoot creates all required files for all PS-IDs according to the manifest +
// per-PS-ID exclusion maps (matching the production exclusions).
func createFullPSIDRoot(t *testing.T, realRoot, tmpDir string) {
	t.Helper()

	copyManifest(t, realRoot, tmpDir)

	for _, ps := range cryptoutilFitnessRegistry.AllProductServices() {
		psDir := filepath.Join(tmpDir, "internal", "apps", ps.PSID)
		require.NoError(t, os.MkdirAll(psDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

		psRoot, err := os.OpenRoot(psDir)
		require.NoError(t, err)

		func() {
			defer func() {
				require.NoError(t, psRoot.Close())
			}()

			serverDir := filepath.Join(psDir, "server")
			serverRelDir := "server"

			require.NoError(t, os.MkdirAll(serverDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))

			// Canonical template-enforced root/client files.
			realServiceRootPath := filepath.Join(realRoot, "internal", "apps", ps.PSID, ps.Service+".go")
			realServiceRootContent, err := os.ReadFile(realServiceRootPath)
			require.NoError(t, err)
			require.NoError(t, psRoot.WriteFile(ps.Service+".go", realServiceRootContent, cryptoutilSharedMagic.CacheFilePermissions))

			canonicalFiles := []string{
				ps.Service + "_test.go",
				filepath.Join("client", "client.go"),
				filepath.Join("server", ps.Service+"_port_conflict_test.go"),
			}

			for _, rel := range canonicalFiles {
				srcPath := filepath.Join(realRoot, "internal", "apps", ps.PSID, rel)
				srcData, readErr := os.ReadFile(srcPath)
				require.NoError(t, readErr)

				require.NoError(t, psRoot.MkdirAll(filepath.Dir(rel), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, psRoot.WriteFile(rel, srcData, cryptoutilSharedMagic.CacheFilePermissions))
			}

			// Required server file: server.go (no exclusions).
			require.NoError(t, psRoot.WriteFile(filepath.Join(serverRelDir, "server.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
			require.NoError(t, psRoot.WriteFile(filepath.Join(serverRelDir, "testmain_test.go"), []byte("package server\n"), cryptoutilSharedMagic.CacheFilePermissions))
			// swagger.go, swagger_test.go, and port_conflict_test are excluded, skip.
			if !knownServerFileExclusions["__SERVICE___lifecycle_test.go"][ps.PSID] {
				require.NoError(t, psRoot.WriteFile(
					filepath.Join(serverRelDir, ps.Service+"_lifecycle_test.go"),
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

			require.NoError(t, psRoot.MkdirAll("client", cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			require.NoError(t, psRoot.MkdirAll("e2e", cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
			require.NoError(t, psRoot.WriteFile(filepath.Join("e2e", "testmain_e2e_test.go"), []byte("package e2e_test\n"), cryptoutilSharedMagic.CacheFilePermissions))
			require.NoError(t, psRoot.WriteFile(filepath.Join("e2e", ps.Service+"_e2e_test.go"), []byte("package e2e_test\n"), cryptoutilSharedMagic.CacheFilePermissions))

			// Required server config files (respecting production exclusions).
			if !knownServerConfigFileExclusions["config.go"][ps.PSID] {
				configDir := filepath.Join(serverDir, "config")
				require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, psRoot.WriteFile(filepath.Join(serverRelDir, "config", "config.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
				require.NoError(t, psRoot.WriteFile(filepath.Join(serverRelDir, "config", "config_test.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
			}

			if !knownServerConfigFileExclusions["config_test_helper.go"][ps.PSID] {
				configDir := filepath.Join(serverDir, "config")
				require.NoError(t, os.MkdirAll(configDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, psRoot.WriteFile(filepath.Join(serverRelDir, "config", "config_test_helper.go"), []byte("package config\n"), cryptoutilSharedMagic.CacheFilePermissions))
			}

			// Required server repository files and dirs (respecting production exclusions).
			if !knownServerRepositoryFileExclusions["migrations.go"][ps.PSID] {
				repoDir := filepath.Join(serverDir, "repository")
				require.NoError(t, os.MkdirAll(filepath.Join(repoDir, "migrations"), cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
				require.NoError(t, psRoot.WriteFile(filepath.Join(serverRelDir, "repository", "migrations.go"), []byte("package repository\n"), cryptoutilSharedMagic.CacheFilePermissions))
			}
		}()
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

	servicePath := filepath.Join(tmpDir, "internal", "apps", cryptoutilSharedMagic.OTLPServicePKICA, "ca_test.go")
	serviceContent, readErr := os.ReadFile(servicePath)
	require.NoError(t, readErr)

	serviceRoot, err := os.OpenRoot(filepath.Dir(servicePath))
	require.NoError(t, err)

	defer func() {
		require.NoError(t, serviceRoot.Close())
	}()

	require.NoError(t, serviceRoot.WriteFile(filepath.Base(servicePath), append(serviceContent, []byte("\n// mismatch\n")...), cryptoutilSharedMagic.CacheFilePermissions))

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
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
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
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
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
		require.NoError(t, os.MkdirAll(psDir, cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupOtherReadExecute))
		psRoot, openErr := os.OpenRoot(psDir)
		require.NoError(t, openErr)

		defer func() {
			require.NoError(t, psRoot.Close())
		}()

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
		require.NoError(t, psRoot.WriteFile(ps.Service+".go", realServiceRootContent, cryptoutilSharedMagic.CacheFilePermissions))
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))

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

// TestCheckInDir_NoExclusions_MissingServerDir exercises the required directory violation path.
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
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
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
	require.Contains(t, err.Error(), "missing required directory: client")
}

// TestCheckInDir_NoExclusions_MissingServerConfigFile exercises the required directory violation path.
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
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
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
	require.Contains(t, err.Error(), "missing required directory: client")
}

// TestCheckInDir_NoExclusions_MissingServerRepositoryFile exercises the required directory violation path.
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
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
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
	require.Contains(t, err.Error(), "missing required directory: client")
}

// TestCheckInDir_NoExclusions_MissingServerRepositoryDir exercises the required directory violation path.
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
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
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
	require.Contains(t, err.Error(), "missing required directory: client")
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
		require.NoError(t, os.WriteFile(filepath.Join(psDir, ps.Service+"_test.go"), []byte("package main\n"), cryptoutilSharedMagic.CacheFilePermissions))
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
	require.Contains(t, err.Error(), "missing required directory: client")
}
