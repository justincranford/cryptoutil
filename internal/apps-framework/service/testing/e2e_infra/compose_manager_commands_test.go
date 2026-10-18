// Copyright (c) 2025-2026 Justin Cranford.

package e2e_infra

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComposeManagerDirectOperations_WrapErrors(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected direct operation failure")
	tests := []struct {
		name      string
		operation func() error
		wantErr   string
	}{
		{
			name: "command run",
			operation: func() error {
				return runCommandWithFn(context.Background(), "docker", nil, func(*exec.Cmd) error {
					return injectedErr
				})
			},
			wantErr: "run command",
		},
		{
			name: "command output",
			operation: func() error {
				_, err := runCommandOutputWithFn(context.Background(), "docker", nil, func(*exec.Cmd) ([]byte, error) {
					return []byte("output"), injectedErr
				})

				return err
			},
			wantErr: "capture output",
		},
		{
			name: testRemovePath,
			operation: func() error {
				return removeAllWithFn("path", func(string) error { return injectedErr })
			},
			wantErr: testRemovePath,
		},
		{
			name: testCreateDirectory,
			operation: func() error {
				return mkdirAllWithFn("path", cryptoutilSharedMagic.PKIInitCertsDirMode, func(string, os.FileMode) error {
					return injectedErr
				})
			},
			wantErr: testCreateDirectory,
		},
		{
			name: testStatPath,
			operation: func() error {
				_, err := statWithFn("path", func(string) (os.FileInfo, error) {
					return nil, injectedErr
				})

				return err
			},
			wantErr: testStatPath,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.operation()

			require.ErrorIs(t, err, injectedErr)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestComposeManagerStartAndStop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		subcommand string
		runErr     error
		wantErr    string
	}{
		{name: "start succeeds", subcommand: testComposeUp},
		{name: "start command fails", subcommand: testComposeUp, runErr: errors.New("injected start failure"), wantErr: "failed to start docker compose"},
		{name: "stop succeeds", subcommand: testComposeDown},
		{name: "stop command fails", subcommand: testComposeDown, runErr: errors.New("injected stop failure"), wantErr: "failed to stop docker compose"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			composeDir := t.TempDir()
			composeFile := filepath.Join(composeDir, cryptoutilSharedMagic.COMPOSE_YML)
			require.NoError(t, os.WriteFile(composeFile, []byte("services: {}"), cryptoutilSharedMagic.CacheFilePermissions))

			cm := NewComposeManager(composeFile, "test-profile")

			var (
				command string
				gotArgs []string
			)

			cm.runCommandFn = func(_ context.Context, gotCommand string, args ...string) error {
				command = gotCommand

				gotArgs = append([]string(nil), args...)

				return tc.runErr
			}

			var err error
			if tc.subcommand == testComposeUp {
				err = cm.Start(context.Background())
			} else {
				err = cm.Stop(context.Background())
			}

			if tc.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tc.wantErr)
				require.ErrorContains(t, err, tc.runErr.Error())
			}

			require.Equal(t, "docker", command)

			expectedSubcommandArgs := []string{valueCompose, valueF, composeFile, testProfileFlag, "test-profile"}
			if tc.subcommand == "up" {
				expectedSubcommandArgs = append(expectedSubcommandArgs, testComposeUp, "-d", "--build")
			} else {
				expectedSubcommandArgs = append(expectedSubcommandArgs, testComposeDown, "-v")
			}

			require.Equal(t, expectedSubcommandArgs, gotArgs)
		})
	}
}

func TestComposeManagerStartCertPreparationError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	composeFile := filepath.Join(tempDir, cryptoutilSharedMagic.COMPOSE_YML)
	require.NoError(t, os.WriteFile(composeFile, []byte("services: {}"), cryptoutilSharedMagic.CacheFilePermissions))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "certs"), []byte("not a directory"), cryptoutilSharedMagic.CacheFilePermissions))

	cm := NewComposeManager(composeFile)
	commandCalled := false
	cm.runCommandFn = func(context.Context, string, ...string) error {
		commandCalled = true

		return nil
	}

	err := cm.Start(context.Background())
	require.ErrorContains(t, err, "failed to prepare cert output dir")
	require.ErrorContains(t, err, "open root")
	require.False(t, commandCalled)
}

func TestComposeManagerSyncCertOutputDirFromPkiInit(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	composeDir := filepath.Join(tempDir, "deployments", "test-service")
	require.NoError(t, os.MkdirAll(composeDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	composeFile := filepath.Join(composeDir, cryptoutilSharedMagic.COMPOSE_YML)
	require.NoError(t, os.WriteFile(composeFile, []byte("services: {}"), cryptoutilSharedMagic.CacheFilePermissions))

	hostCACertPath := filepath.Join(tempDir, "deep", "one", "two", "three", "four", "ca.pem")
	hostCertsDir := filepath.Clean(filepath.Join(hostCACertPath, "..", "..", "..", ".."))
	require.NoError(t, os.MkdirAll(hostCertsDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(filepath.Join(hostCertsDir, "stale.pem"), []byte("stale"), cryptoutilSharedMagic.CacheFilePermissions))

	cm := NewComposeManager(composeFile)

	var (
		command string
		gotArgs []string
	)

	cm.runCommandFn = func(_ context.Context, gotCommand string, args ...string) error {
		command = gotCommand

		gotArgs = append([]string(nil), args...)

		return nil
	}

	require.NoError(t, cm.SyncCertOutputDirFromPkiInit(context.Background(), hostCACertPath))
	require.Equal(t, "docker", command)
	require.Equal(t, []string{"cp", "test-service-pki-init-1:/certs/.", hostCertsDir}, gotArgs)
	_, err := os.Stat(filepath.Join(hostCertsDir, "stale.pem"))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestComposeManagerSyncCertOutputDirFromPkiInitMakesHostCertTreeWritableBeforeRemoval(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	composeDir := filepath.Join(tempDir, "deployments", "test-service")
	require.NoError(t, os.MkdirAll(composeDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	composeFile := filepath.Join(composeDir, cryptoutilSharedMagic.COMPOSE_YML)
	require.NoError(t, os.WriteFile(composeFile, []byte("services: {}"), cryptoutilSharedMagic.CacheFilePermissions))

	hostCACertPath := filepath.Join(tempDir, "deep", "one", "two", "three", "four", "ca.pem")
	hostCertsDir := filepath.Clean(filepath.Join(hostCACertPath, "..", "..", "..", ".."))
	adminDir := filepath.Join(hostCertsDir, "admin")
	adminKeyPath := filepath.Join(adminDir, "admin.key")
	require.NoError(t, os.MkdirAll(adminDir, cryptoutilSharedMagic.CICDTempDirPermissions))
	require.NoError(t, os.WriteFile(adminKeyPath, []byte("private key"), cryptoutilSharedMagic.CacheFilePermissions))
	t.Cleanup(func() {
		if err := makeWritableRecursive(hostCertsDir); err != nil {
			assert.Failf(t, cryptoutilSharedMagic.CICDTestErrorfFailureMessage, "make temporary host cert tree writable for cleanup: %v", err)
		}
	})

	readOnlyDirPermissions := fs.FileMode(cryptoutilSharedMagic.CICDTempDirPermissions &^ cryptoutilSharedMagic.CICDFileOwnerWritePermission)
	readOnlyFilePermissions := fs.FileMode(cryptoutilSharedMagic.CacheFilePermissions &^ cryptoutilSharedMagic.CICDFileOwnerWritePermission)
	require.NoError(t, os.Chmod(adminKeyPath, readOnlyFilePermissions))
	require.NoError(t, os.Chmod(adminDir, readOnlyDirPermissions))
	require.NoError(t, os.Chmod(hostCertsDir, readOnlyDirPermissions))

	cm := NewComposeManager(composeFile)
	writablePrepared := false
	cm.makeWritableFn = func(path string) error {
		require.Equal(t, hostCertsDir, path)

		if err := makeWritableRecursive(path); err != nil {
			return err
		}

		for _, directory := range []string{hostCertsDir, adminDir} {
			info, err := os.Stat(directory)
			require.NoError(t, err)
			require.NotZero(t, info.Mode().Perm()&cryptoutilSharedMagic.CICDFileOwnerWritePermission)
		}

		keyInfo, err := os.Stat(adminKeyPath)
		require.NoError(t, err)

		if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
			require.NotZero(t, keyInfo.Mode().Perm()&cryptoutilSharedMagic.CICDFileOwnerWritePermission)
		} else {
			require.Equal(t, readOnlyFilePermissions, keyInfo.Mode().Perm())
		}

		writablePrepared = true

		return nil
	}
	cm.removeAllFn = func(path string) error {
		require.Equal(t, hostCertsDir, path)
		require.True(t, writablePrepared, "host cert tree must be made writable before removal")

		return os.RemoveAll(path)
	}
	cm.runCommandFn = func(_ context.Context, command string, args ...string) error {
		require.Equal(t, "docker", command)
		require.Equal(t, []string{"cp", "test-service-pki-init-1:/certs/.", hostCertsDir}, args)

		return nil
	}

	require.NoError(t, cm.SyncCertOutputDirFromPkiInit(context.Background(), hostCACertPath))

	_, err := os.Stat(adminKeyPath)
	require.ErrorIs(t, err, os.ErrNotExist)
	require.DirExists(t, hostCertsDir)
}

func TestComposeManagerSyncCertOutputDirFromPkiInitCopyError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	composeDir := filepath.Join(tempDir, "deployments", "copy-error")
	require.NoError(t, os.MkdirAll(composeDir, cryptoutilSharedMagic.CICDTempDirPermissions))

	cm := NewComposeManager(filepath.Join(composeDir, cryptoutilSharedMagic.COMPOSE_YML))
	cm.runCommandFn = func(_ context.Context, _ string, args ...string) error {
		require.Equal(t, "cp", args[0])

		return errors.New("injected copy failure")
	}

	err := cm.SyncCertOutputDirFromPkiInit(
		context.Background(),
		filepath.Join(tempDir, "deep", "one", "two", "three", "four", "ca.pem"),
	)
	require.ErrorContains(t, err, "copy certs from copy-error-pki-init-1")
	require.ErrorContains(t, err, "injected copy failure")
}

func TestComposeManagerSyncCertOutputDirFromPkiInitMkdirError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	cm := NewComposeManager(filepath.Join(tempDir, cryptoutilSharedMagic.COMPOSE_YML))
	cm.mkdirAllFn = func(_ string, _ fs.FileMode) error {
		return errors.New("injected mkdir failure")
	}

	hostCACertPath := filepath.Join(tempDir, "certs", "one", "two", "three", "four", "ca.pem")
	err := cm.SyncCertOutputDirFromPkiInit(context.Background(), hostCACertPath)
	require.ErrorContains(t, err, "mkdir host certs dir")
	require.ErrorContains(t, err, "injected mkdir failure")
}

func TestComposeManagerDefaultPsOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		output     []byte
		commandErr error
		wantErr    string
	}{
		{name: "output succeeds", output: []byte("[]")},
		{name: "command error retains output", output: []byte("docker diagnostic"), commandErr: errors.New("injected ps failure"), wantErr: "docker compose ps failed"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cm := NewComposeManager("compose-test.yml")

			var (
				gotCommand string
				gotArgs    []string
			)

			cm.runCommandOutputFn = func(_ context.Context, command string, args ...string) ([]byte, error) {
				gotCommand = command

				gotArgs = append([]string(nil), args...)

				return tc.output, tc.commandErr
			}

			output, err := cm.defaultPsOutput(context.Background())
			if tc.wantErr == "" {
				require.NoError(t, err)
				require.Equal(t, tc.output, output)
			} else {
				require.ErrorContains(t, err, tc.wantErr)
				require.ErrorContains(t, err, tc.commandErr.Error())
				require.Equal(t, tc.output, output)
			}

			require.Equal(t, "docker", gotCommand)
			require.Equal(t, []string{valueCompose, valueF, "compose-test.yml", "ps", "-a", "--format", "json"}, gotArgs)
		})
	}
}

func TestComposeManagerRunCommandOutput_DefaultRunner(t *testing.T) {
	t.Parallel()

	output, err := (&ComposeManager{}).runCommandOutput(context.Background(), "go", cryptoutilSharedMagic.CLIVersionCommand)
	require.NoError(t, err)
	require.Contains(t, string(output), cryptoutilSharedMagic.DevSetupGoCheckCmd)
}
