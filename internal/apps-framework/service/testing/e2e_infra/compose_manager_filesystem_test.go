// Copyright (c) 2025-2026 Justin Cranford.

package e2e_infra

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	currentDirectory    = "."
	nestedDirectory     = "nested"
	testGOOSLinux       = cryptoutilSharedMagic.MockRuntimeGoOS
	testGOOSWindows     = cryptoutilSharedMagic.OSNameWindows
	testErrChmodFile    = "chmod file"
	testReadOnlyPEMFile = "readonly.pem"
)

type testWritableRoot struct {
	chmodErr   error
	chmodCalls *[]string
	chmodModes *[]fs.FileMode
}

type testDirEntryInfoError struct {
	isDir   bool
	infoErr error
}

func (entry testDirEntryInfoError) Name() string {
	return currentDirectory
}

func (entry testDirEntryInfoError) IsDir() bool {
	return entry.isDir
}

func (entry testDirEntryInfoError) Type() fs.FileMode {
	if entry.isDir {
		return fs.ModeDir
	}

	return 0
}

func (entry testDirEntryInfoError) Info() (fs.FileInfo, error) {
	return nil, entry.infoErr
}

func (root testWritableRoot) Chmod(name string, mode fs.FileMode) error {
	if root.chmodCalls != nil {
		*root.chmodCalls = append(*root.chmodCalls, name)
	}

	if root.chmodModes != nil {
		*root.chmodModes = append(*root.chmodModes, mode)
	}

	return root.chmodErr
}

type testFileInfoMode struct {
	fs.FileInfo
	mode fs.FileMode
}

func (info testFileInfoMode) Mode() fs.FileMode {
	return info.mode
}

func (testWritableRoot) Close() error {
	return nil
}

func TestComposeManagerFilesystemErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*ComposeManager, string) error
		run     func(*ComposeManager) error
		wantErr string
	}{
		{
			name: "remove certificates error",
			setup: func(cm *ComposeManager, _ string) error {
				cm.removeAllFn = func(string) error {
					return errors.New("injected remove error")
				}

				return nil
			},
			run: func(cm *ComposeManager) error {
				return cm.SyncCertOutputDirFromPkiInit(
					context.Background(),
					filepath.Join(filepath.Dir(cm.ComposeFile), "a", "b", "c", "d", "ca.pem"),
				)
			},
			wantErr: "remove host certs dir",
		},
		{
			name: "create certificates directory error",
			setup: func(cm *ComposeManager, _ string) error {
				cm.mkdirAllFn = func(string, fs.FileMode) error {
					return errors.New("injected mkdir error")
				}

				return nil
			},
			run: func(cm *ComposeManager) error {
				return cm.SyncCertOutputDirFromPkiInit(
					context.Background(),
					filepath.Join(filepath.Dir(cm.ComposeFile), "a", "b", "c", "d", "ca.pem"),
				)
			},
			wantErr: "mkdir host certs dir",
		},
		{
			name: "stat certificates directory error",
			setup: func(cm *ComposeManager, _ string) error {
				cm.statFn = func(string) (os.FileInfo, error) {
					return nil, errors.New("injected stat error")
				}

				return nil
			},
			run: func(cm *ComposeManager) error {
				return cm.resetCertOutputDir()
			},
			wantErr: "stat certs dir",
		},
		{
			name: "make certificates directory writable error",
			setup: func(cm *ComposeManager, composeDir string) error {
				info, err := os.Stat(composeDir)
				if err != nil {
					return err
				}

				cm.statFn = func(string) (os.FileInfo, error) {
					return info, nil
				}
				cm.makeWritableFn = func(string) error {
					return errors.New("injected chmod error")
				}

				return nil
			},
			run: func(cm *ComposeManager) error {
				return cm.resetCertOutputDir()
			},
			wantErr: "make writable certs dir",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			composeDir := t.TempDir()
			composeFile := filepath.Join(composeDir, cryptoutilSharedMagic.COMPOSE_YML)
			require.NoError(t, os.WriteFile(composeFile, []byte("services: {}"), cryptoutilSharedMagic.CacheFilePermissions))

			cm := NewComposeManager(composeFile)
			require.NoError(t, tc.setup(cm, composeDir))
			err := tc.run(cm)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestMakeWritableRecursiveWithOpsSkipsMatchingDirectoryPermissions(t *testing.T) {
	t.Parallel()

	rootPath := t.TempDir()
	nestedPath := filepath.Join(rootPath, nestedDirectory)
	require.NoError(t, os.Mkdir(nestedPath, cryptoutilSharedMagic.CICDTempDirPermissions))

	dirInfo, err := os.Stat(nestedPath)
	require.NoError(t, err)

	rootEntry := fs.FileInfoToDirEntry(testFileInfoMode{
		FileInfo: dirInfo,
		mode:     fs.ModeDir | cryptoutilSharedMagic.CICDTempDirPermissions,
	})

	chmodCalls := []string{}
	root := testWritableRoot{
		chmodErr:   errors.New("root chmod must not be called"),
		chmodCalls: &chmodCalls,
	}
	err = makeWritableRecursiveWithOps(
		rootPath,
		func(string) (writableRoot, error) {
			return root, nil
		},
		func(_ string, walkFn fs.WalkDirFunc) error {
			return walkFn(nestedPath, rootEntry, nil)
		},
		func(_, _ string) (string, error) {
			return nestedDirectory, nil
		},
		runtime.GOOS,
	)

	require.NoError(t, err)
	require.Empty(t, chmodCalls)
}

func TestMakeWritableRecursiveWithOpsSkipsOwnerWritableFiles(t *testing.T) {
	t.Parallel()

	rootPath := t.TempDir()
	filePath := filepath.Join(rootPath, "writable.pem")
	require.NoError(t, os.WriteFile(filePath, []byte("certificate"), cryptoutilSharedMagic.FilePermissions))

	fileInfo, err := os.Stat(filePath)
	require.NoError(t, err)

	fileEntry := fs.FileInfoToDirEntry(testFileInfoMode{
		FileInfo: fileInfo,
		mode:     cryptoutilSharedMagic.FilePermissions,
	})
	chmodCalls := []string{}
	root := testWritableRoot{
		chmodErr:   errors.New("writable file chmod must not be called"),
		chmodCalls: &chmodCalls,
	}
	err = makeWritableRecursiveWithOps(
		rootPath,
		func(string) (writableRoot, error) {
			return root, nil
		},
		func(_ string, walkFn fs.WalkDirFunc) error {
			return walkFn(filePath, fileEntry, nil)
		},
		func(string, string) (string, error) {
			return "writable.pem", nil
		},
		runtime.GOOS,
	)

	require.NoError(t, err)
	require.Empty(t, chmodCalls)
}

func TestMakeWritableRecursiveWithOpsFilePermissionsByOS(t *testing.T) {
	t.Parallel()

	rootPath := t.TempDir()
	filePath := filepath.Join(rootPath, testReadOnlyPEMFile)
	require.NoError(t, os.WriteFile(filePath, []byte("certificate"), cryptoutilSharedMagic.CacheFilePermissions))

	fileInfo, err := os.Stat(filePath)
	require.NoError(t, err)

	fileEntry := fs.FileInfoToDirEntry(testFileInfoMode{
		FileInfo: fileInfo,
		mode:     fileInfo.Mode() &^ cryptoutilSharedMagic.CICDFileOwnerWritePermission,
	})
	injectedErr := errors.New("injected chmod error")

	tests := []struct {
		name          string
		goos          string
		chmodErr      error
		wantChmodCall bool
		wantErr       string
	}{
		{
			name:     "unix skips read-only file chmod",
			goos:     testGOOSLinux,
			chmodErr: injectedErr,
		},
		{
			name:          "windows clears read-only file permission",
			goos:          testGOOSWindows,
			wantChmodCall: true,
		},
		{
			name:          "windows reports chmod error",
			goos:          testGOOSWindows,
			chmodErr:      injectedErr,
			wantChmodCall: true,
			wantErr:       testErrChmodFile,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			chmodCalls := []string{}
			chmodModes := []fs.FileMode{}
			root := testWritableRoot{
				chmodErr:   tc.chmodErr,
				chmodCalls: &chmodCalls,
				chmodModes: &chmodModes,
			}
			err := makeWritableRecursiveWithOps(
				rootPath,
				func(string) (writableRoot, error) {
					return root, nil
				},
				func(_ string, walkFn fs.WalkDirFunc) error {
					return walkFn(filePath, fileEntry, nil)
				},
				func(string, string) (string, error) {
					return testReadOnlyPEMFile, nil
				},
				tc.goos,
			)

			if tc.wantErr != "" {
				require.ErrorContains(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tc.wantChmodCall {
				require.Equal(t, []string{testReadOnlyPEMFile}, chmodCalls)
				require.Equal(t, []fs.FileMode{cryptoutilSharedMagic.CacheFilePermissions}, chmodModes)
			} else {
				require.Empty(t, chmodCalls)
				require.Empty(t, chmodModes)
			}
		})
	}
}

func TestMakeWritableRecursiveWithOpsErrors(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "leaf.pem")
	require.NoError(t, os.WriteFile(filePath, []byte("certificate"), cryptoutilSharedMagic.CacheFilePermissions))

	dirInfo, err := os.Stat(tempDir)
	require.NoError(t, err)
	fileInfo, err := os.Stat(filePath)
	require.NoError(t, err)

	dirEntry := fs.FileInfoToDirEntry(testFileInfoMode{
		FileInfo: dirInfo,
		mode:     fs.ModeDir,
	})
	fileEntry := fs.FileInfoToDirEntry(testFileInfoMode{
		FileInfo: fileInfo,
		mode:     fileInfo.Mode() &^ cryptoutilSharedMagic.CICDFileOwnerWritePermission,
	})
	injectedErr := errors.New("injected filesystem error")

	tests := []struct {
		name     string
		walkPath string
		entry    fs.DirEntry
		walkErr  error
		relErr   error
		chmodErr error
		goos     string
		wantErr  string
	}{
		{
			name:     "walk error",
			walkPath: tempDir,
			walkErr:  injectedErr,
			wantErr:  "walk and normalize writable perms",
		},
		{
			name:     "relative path error",
			walkPath: tempDir,
			entry:    dirEntry,
			relErr:   injectedErr,
			wantErr:  "rel path",
		},
		{
			name:     "directory info error",
			walkPath: tempDir,
			entry:    testDirEntryInfoError{isDir: true, infoErr: injectedErr},
			wantErr:  "stat dir",
		},
		{
			name:     "directory chmod error",
			walkPath: tempDir,
			entry:    dirEntry,
			chmodErr: injectedErr,
			wantErr:  "chmod dir",
		},
		{
			name:     "file info error",
			walkPath: filePath,
			entry:    testDirEntryInfoError{infoErr: injectedErr},
			goos:     testGOOSWindows,
			wantErr:  "stat file",
		},
		{
			name:     "file chmod error",
			walkPath: filePath,
			entry:    fileEntry,
			chmodErr: injectedErr,
			goos:     testGOOSWindows,
			wantErr:  testErrChmodFile,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := testWritableRoot{chmodErr: tc.chmodErr}
			err := makeWritableRecursiveWithOps(
				tempDir,
				func(string) (writableRoot, error) {
					return root, nil
				},
				func(_ string, walkFn fs.WalkDirFunc) error {
					return walkFn(tc.walkPath, tc.entry, tc.walkErr)
				},
				func(string, string) (string, error) {
					if tc.relErr != nil {
						return "", tc.relErr
					}

					return currentDirectory, nil
				},
				func() string {
					if tc.goos != "" {
						return tc.goos
					}

					return runtime.GOOS
				}(),
			)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
