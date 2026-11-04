// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

type faultFileSystem struct {
	base                 FileSystem
	failOperation        string
	chmodCalls           int
	failChmodAt          int
	failLstatName        string
	failOutputLstatAt    int
	failStageLstat       bool
	symlinkPath          string
	stageSymlink         bool
	realStageSymlink     bool
	stageNonDirectory    bool
	stageEscaped         bool
	stageDirectory       string
	failRenameCheckAt    int
	outputLstatCalls     int
	outputDirectory      string
	cancelAfterMkdirTemp context.CancelFunc
	failRemove           bool
}

func TestOSReadFileAndFilesystemRootValidation(t *testing.T) {
	t.Parallel()

	fileName := filepath.Join(t.TempDir(), "projected-input")
	require.NoError(t, os.WriteFile(fileName, []byte("projected"), 0o600))
	data, err := osReadFile(fileName)
	require.NoError(t, err)
	require.Equal(t, []byte("projected"), data)

	rootDirectory := string(filepath.Separator)
	require.NoError(t, validateDirectoryChain(operatingSystemFileSystem{}, rootDirectory))
}

func (fileSystem *faultFileSystem) Lstat(name string) (os.FileInfo, error) {
	if fileSystem.failOperation == testFileOperationLstat && name == fileSystem.failLstatName {
		return nil, errors.New("injected lstat failure")
	}

	if fileSystem.symlinkPath != "" && name == fileSystem.symlinkPath {
		return syntheticFileInfo{mode: os.ModeSymlink}, nil
	}

	if name == fileSystem.stageDirectory {
		if fileSystem.failStageLstat {
			return nil, errors.New("injected staging lstat failure")
		}

		if fileSystem.stageSymlink {
			return syntheticFileInfo{mode: os.ModeSymlink}, nil
		}

		if fileSystem.stageNonDirectory {
			return syntheticFileInfo{mode: 0o600}, nil
		}
	}

	if name == fileSystem.outputDirectory {
		fileSystem.outputLstatCalls++
		if fileSystem.failOutputLstatAt > 0 && fileSystem.outputLstatCalls == fileSystem.failOutputLstatAt {
			_ = fileSystem.base.Mkdir(name, privateDirectoryMode)
		}

		if fileSystem.failRenameCheckAt > 0 && fileSystem.outputLstatCalls == fileSystem.failRenameCheckAt {
			return nil, errors.New("injected output recheck failure")
		}
	}

	return fileSystem.base.Lstat(name)
}

func (fileSystem *faultFileSystem) Mkdir(name string, mode os.FileMode) error {
	if fileSystem.failOperation == testFileOperationMkdir {
		return errors.New("injected mkdir failure")
	}

	return fileSystem.base.Mkdir(name, mode)
}

func (fileSystem *faultFileSystem) MkdirTemp(directory string, pattern string) (string, error) {
	if fileSystem.failOperation == testFileOperationMkdirTemp {
		return "", errors.New("injected mkdirtemp failure")
	}

	if fileSystem.stageEscaped {
		return filepath.Join(directory, "..", "escaped-stage"), nil
	}

	stage, err := fileSystem.base.MkdirTemp(directory, pattern)

	fileSystem.stageDirectory = stage
	if err == nil && fileSystem.realStageSymlink {
		if removeErr := fileSystem.base.RemoveAll(stage); removeErr != nil {
			return "", removeErr
		}

		if linkErr := os.Symlink(directory, stage); linkErr != nil {
			return "", linkErr
		}
	}

	if fileSystem.cancelAfterMkdirTemp != nil {
		fileSystem.cancelAfterMkdirTemp()
	}

	return stage, err
}

func (fileSystem *faultFileSystem) OpenFile(name string, flag int, mode os.FileMode) (FileHandle, error) {
	if fileSystem.failOperation == testFileOperationOpen {
		return nil, errors.New("injected open failure")
	}

	handle, err := fileSystem.base.OpenFile(name, flag, mode)
	if err != nil {
		return nil, err
	}

	return &faultFileHandle{FileHandle: handle, failOperation: fileSystem.failOperation}, nil
}

func (fileSystem *faultFileSystem) Chmod(name string, mode os.FileMode) error {
	fileSystem.chmodCalls++
	if fileSystem.failOperation == "chmod" || fileSystem.failChmodAt == fileSystem.chmodCalls {
		return errors.New("injected chmod failure")
	}

	return fileSystem.base.Chmod(name, mode)
}

func (fileSystem *faultFileSystem) Rename(oldPath string, newPath string) error {
	if fileSystem.failOperation == testFileOperationRename {
		return errors.New("injected rename failure")
	}

	return fileSystem.base.Rename(oldPath, newPath)
}

func (fileSystem *faultFileSystem) RemoveAll(name string) error {
	if fileSystem.failRemove {
		return errors.New("injected remove failure")
	}

	return fileSystem.base.RemoveAll(name)
}

type faultFileHandle struct {
	FileHandle
	failOperation string
}

func (file *faultFileHandle) Write(data []byte) (int, error) {
	switch file.failOperation {
	case cryptoutilSharedMagic.ScopeWrite:
		return 0, errors.New("injected write failure")
	case testFileOperationShortWrite:
		return len(data) - 1, nil
	}

	return file.FileHandle.Write(data)
}

func (file *faultFileHandle) Sync() error {
	if file.failOperation == testFileOperationSync {
		return errors.New("injected sync failure")
	}

	return file.FileHandle.Sync()
}

func (file *faultFileHandle) Close() error {
	closeErr := file.FileHandle.Close()
	if file.failOperation == testFileOperationClose {
		return errors.Join(closeErr, errors.New("injected close failure"))
	}

	return closeErr
}

func TestPublishFilesAtomically(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		files       map[string]materializedFile
		wantErr     string
		verifyFiles bool
	}{
		{
			name: "role files are restrictive and complete",
			files: map[string]materializedFile{
				cryptoutilSharedMagic.DeploymentMaterializerHealthClientCertificateFile: {data: []byte("certificate")},
				cryptoutilSharedMagic.DeploymentMaterializerHealthClientPrivateKeyFile:  {data: []byte("private key")},
				cryptoutilSharedMagic.DeploymentMaterializerPublicServerCertificateFile: {data: []byte("public certificate")},
				cryptoutilSharedMagic.DeploymentMaterializerPublicServerPrivateKeyFile:  {data: []byte("public private key")},
			},
			verifyFiles: true,
		},
		{name: "empty output rejected", wantErr: "no validated output files"},
		{name: "unsafe relative output rejected", files: map[string]materializedFile{"../escape": {data: []byte("x")}}, wantErr: testFilePathInvalidError},
		{name: "absolute output file rejected", files: map[string]materializedFile{"/escape": {data: []byte("x")}}, wantErr: testFilePathInvalidError},
		{name: "backslash output file rejected", files: map[string]materializedFile{`health-client\tls.key`: {data: []byte("x")}}, wantErr: testFilePathInvalidError},
		{name: "repeated separator rejected", files: map[string]materializedFile{"health-client//tls.key": {data: []byte("x")}}, wantErr: testFilePathInvalidError},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			parent := t.TempDir()
			outputDirectory := filepath.Join(parent, "materialized")

			err := publishFiles(context.Background(), operatingSystemFileSystem{}, outputDirectory, testCase.files)
			if testCase.wantErr != "" {
				require.ErrorContains(t, err, testCase.wantErr)
				require.NoDirExists(t, outputDirectory)

				return
			}

			require.NoError(t, err)
			require.DirExists(t, outputDirectory)

			if testCase.verifyFiles {
				for name, expected := range testCase.files {
					path := filepath.Join(outputDirectory, filepath.FromSlash(name))
					content, readErr := os.ReadFile(path)
					require.NoError(t, readErr)
					require.Equal(t, expected.data, content)

					info, statErr := os.Stat(path)
					require.NoError(t, statErr)
					requirePlatformFileMode(t, os.FileMode(privateFileMode), info.Mode().Perm())
					requirePlatformFileMode(
						t,
						os.FileMode(privateDirectoryMode),
						mustStat(t, filepath.Dir(path)).Mode().Perm(),
					)
				}

				parentEntries, readErr := os.ReadDir(parent)
				require.NoError(t, readErr)
				require.Len(t, parentEntries, 1)
				require.Equal(t, "materialized", parentEntries[0].Name())
			}
		})
	}
}

func requirePlatformFileMode(t *testing.T, expected os.FileMode, actual os.FileMode) {
	t.Helper()

	if runtime.GOOS == cryptoutilSharedMagic.OSNameWindows {
		expected &= cryptoutilSharedMagic.FilePermOwnerReadWriteOnly
		actual &= cryptoutilSharedMagic.FilePermOwnerReadWriteOnly
	}

	require.Equal(t, expected, actual)
}

func TestPublishFilesInjectedFailuresCleanup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		fault        func(*faultFileSystem)
		wantErr      string
		wantLeftover bool
	}{
		{name: testInitialContextCancellationCaseName, fault: func(_ *faultFileSystem) {}, wantErr: testCanceledError, wantLeftover: false},
		{
			name: "output destination lstat failure",
			fault: func(fileSystem *faultFileSystem) {
				fileSystem.failOperation = testFileOperationLstat
				fileSystem.failLstatName = fileSystem.outputDirectory
			},
			wantErr: "inspect materializer output directory",
		},
		{name: "staging creation failure", fault: func(fileSystem *faultFileSystem) { fileSystem.failOperation = testFileOperationMkdirTemp }, wantErr: "create materializer staging"},
		{name: "escaped staging directory rejected", fault: func(fileSystem *faultFileSystem) { fileSystem.stageEscaped = true }, wantErr: "escaped its output parent"},
		{name: "staging inspection failure", fault: func(fileSystem *faultFileSystem) { fileSystem.failStageLstat = true }, wantErr: "inspect materializer staging"},
		{name: "staging symlink rejected", fault: func(fileSystem *faultFileSystem) { fileSystem.stageSymlink = true }, wantErr: "staging path is a symlink"},
		{name: "staging non-directory rejected", fault: func(fileSystem *faultFileSystem) { fileSystem.stageNonDirectory = true }, wantErr: "staging path is a symlink or non-directory"},
		{name: "staging chmod failure", fault: func(fileSystem *faultFileSystem) { fileSystem.failChmodAt = 1 }, wantErr: "secure materializer staging"},
		{name: "role mkdir failure", fault: func(fileSystem *faultFileSystem) { fileSystem.failOperation = testFileOperationMkdir }, wantErr: "create materializer role"},
		{name: "role chmod failure", fault: func(fileSystem *faultFileSystem) { fileSystem.failChmodAt = 2 }, wantErr: "secure materializer role"},
		{name: "output open failure", fault: func(fileSystem *faultFileSystem) { fileSystem.failOperation = testFileOperationOpen }, wantErr: testCreateMaterializerOutputFileError},
		{name: "output write failure", fault: func(fileSystem *faultFileSystem) { fileSystem.failOperation = cryptoutilSharedMagic.ScopeWrite }, wantErr: testWriteMaterializerOutputError},
		{name: "short output write", fault: func(fileSystem *faultFileSystem) { fileSystem.failOperation = testFileOperationShortWrite }, wantErr: "short write"},
		{name: "output sync failure", fault: func(fileSystem *faultFileSystem) { fileSystem.failOperation = testFileOperationSync }, wantErr: testWriteMaterializerOutputError},
		{name: "output close failure", fault: func(fileSystem *faultFileSystem) { fileSystem.failOperation = "close" }, wantErr: "close materializer output"},
		{name: "output chmod failure", fault: func(fileSystem *faultFileSystem) { fileSystem.failChmodAt = testThirdOccurrence }, wantErr: "secure materializer output"},
		{name: "atomic rename failure", fault: func(fileSystem *faultFileSystem) { fileSystem.failOperation = testFileOperationRename }, wantErr: "atomically publish"},
		{
			name: "cleanup failure is reported",
			fault: func(fileSystem *faultFileSystem) {
				fileSystem.failOperation = testFileOperationOpen
				fileSystem.failRemove = true
			},
			wantErr:      "remove materializer staging directory",
			wantLeftover: true,
		},
		{
			name: testTargetCreatedDuringPublishRejectedCaseName,
			fault: func(fileSystem *faultFileSystem) {
				fileSystem.failOutputLstatAt = 2
			},
			wantErr: "reused during publication",
		},
		{
			name: "target recheck error is reported",
			fault: func(fileSystem *faultFileSystem) {
				fileSystem.failRenameCheckAt = 2
			},
			wantErr: "recheck materializer output directory",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			parent := t.TempDir()
			output := filepath.Join(parent, "materialized")
			fileSystem := &faultFileSystem{base: operatingSystemFileSystem{}, outputDirectory: output}
			testCase.fault(fileSystem)

			ctx := context.Background()

			var cancel context.CancelFunc
			if testCase.name == testInitialContextCancellationCaseName {
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			if testCase.name == "canceled after staging directory" {
				ctx, cancel = context.WithCancel(ctx)
				fileSystem.cancelAfterMkdirTemp = cancel
			}

			err := publishFiles(ctx, fileSystem, output, oneTestFile())
			require.ErrorContains(t, err, testCase.wantErr)

			entries, readErr := os.ReadDir(parent)
			require.NoError(t, readErr)

			if testCase.wantLeftover {
				require.NotEmpty(t, entries)
			} else if testCase.name != testTargetCreatedDuringPublishRejectedCaseName {
				require.Empty(t, entries)
			}
		})
	}
}

func TestPublishFilesCancellationAfterStaging(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	output := filepath.Join(parent, "materialized")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fileSystem := &faultFileSystem{
		base: operatingSystemFileSystem{}, outputDirectory: output, cancelAfterMkdirTemp: cancel,
	}
	err := publishFiles(ctx, fileSystem, output, oneTestFile())
	require.ErrorContains(t, err, "canceled while staging")

	entries, readErr := os.ReadDir(parent)
	require.NoError(t, readErr)
	require.Empty(t, entries)
}

func TestPublishFilesCancellationBeforeAtomicRename(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	output := filepath.Join(parent, "materialized")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fileSystem := &cancelOnCloseFileSystem{
		FileSystem: operatingSystemFileSystem{}, cancel: cancel,
	}
	err := publishFiles(ctx, fileSystem, output, oneTestFile())
	require.ErrorContains(t, err, "canceled before atomic publication")

	entries, readErr := os.ReadDir(parent)
	require.NoError(t, readErr)
	require.Empty(t, entries)
}

func TestValidateDirectoryChainLstatError(t *testing.T) {
	t.Parallel()

	parent := t.TempDir()
	fileSystem := &faultFileSystem{
		base: operatingSystemFileSystem{}, failOperation: testFileOperationLstat,
		failLstatName: filepath.Join(parent, "child"),
	}
	err := validateDirectoryChain(fileSystem, filepath.Join(parent, "child"))
	require.ErrorContains(t, err, "inspect materializer output parent")
	require.ErrorContains(t, err, "injected lstat failure")
}

func oneTestFile() map[string]materializedFile {
	return map[string]materializedFile{cryptoutilSharedMagic.DeploymentMaterializerHealthClientPrivateKeyFile: {data: []byte("secret")}}
}

func mustStat(t *testing.T, name string) os.FileInfo {
	t.Helper()

	info, err := os.Stat(name)
	require.NoError(t, err)

	return info
}

func TestOutputModesDoNotIncludeGroupWrite(t *testing.T) {
	t.Parallel()

	require.Equal(t, os.FileMode(0o750), os.FileMode(privateDirectoryMode))
	require.Equal(t, os.FileMode(0o440), os.FileMode(privateFileMode))
	require.True(t, strings.HasPrefix(defaultOutputDirectory, "/run/cryptoutil/pki"))
}

type cancelOnCloseFileSystem struct {
	FileSystem
	cancel context.CancelFunc
}

func (fileSystem *cancelOnCloseFileSystem) OpenFile(name string, flag int, mode os.FileMode) (FileHandle, error) {
	handle, err := fileSystem.FileSystem.OpenFile(name, flag, mode)
	if err != nil {
		return nil, err
	}

	return &cancelOnCloseFile{FileHandle: handle, cancel: fileSystem.cancel}, nil
}

func (fileSystem *cancelOnCloseFile) Close() error {
	err := fileSystem.FileHandle.Close()
	fileSystem.cancel()

	return err
}

type cancelOnCloseFile struct {
	FileHandle
	cancel context.CancelFunc
}

type syntheticFileInfo struct {
	mode os.FileMode
}

func (info syntheticFileInfo) Name() string       { return "synthetic" }
func (info syntheticFileInfo) Size() int64        { return 0 }
func (info syntheticFileInfo) Mode() os.FileMode  { return info.mode }
func (info syntheticFileInfo) ModTime() time.Time { return time.Time{} }
func (info syntheticFileInfo) IsDir() bool        { return false }
func (info syntheticFileInfo) Sys() any           { return nil }
