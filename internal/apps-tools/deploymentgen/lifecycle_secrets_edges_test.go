// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"errors"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
)

const (
	testLifecycleDirectoryFault = "directory"
	testLifecycleOpenFault      = "open"
	testLifecycleStatFault      = "stat"
	testLifecycleReadFault      = cryptoutilSharedMagic.ScopeRead
	testLifecycleCloseFault     = "close"
	testLifecycleGrowthFault    = "growth"
	testLifecycleShrinkFault    = "shrink"
	testLifecycleProtectedFile  = "protected-material"
)

type lifecycleFaultFS struct {
	files fstest.MapFS
	fault string
}

func (source lifecycleFaultFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if source.fault == testLifecycleDirectoryFault {
		return nil, errors.New("injected directory failure")
	}

	return fs.ReadDir(source.files, name)
}

func (source lifecycleFaultFS) Open(name string) (fs.File, error) {
	if source.fault == testLifecycleOpenFault {
		return nil, errors.New("injected open failure")
	}

	handle, err := source.files.Open(name)
	if err != nil {
		return nil, err
	}

	return lifecycleFaultFile{File: handle, fault: source.fault}, nil
}

type lifecycleFaultFile struct {
	fs.File
	fault string
}

func (file lifecycleFaultFile) Stat() (fs.FileInfo, error) {
	if file.fault == testLifecycleStatFault {
		return nil, errors.New("injected stat failure")
	}

	info, err := file.File.Stat()
	if err != nil {
		return nil, err
	}

	if file.fault == testLifecycleGrowthFault || file.fault == testLifecycleShrinkFault {
		return lifecycleFalseSize{FileInfo: info}, nil
	}

	return info, nil
}

type lifecycleFalseSize struct{ fs.FileInfo }

func (lifecycleFalseSize) Size() int64 { return 1 }

func (file lifecycleFaultFile) Read(data []byte) (int, error) {
	if file.fault == testLifecycleReadFault {
		return 0, errors.New("injected read failure")
	}

	return file.File.Read(data)
}

func (file lifecycleFaultFile) Close() error {
	if err := file.File.Close(); err != nil {
		return err
	}

	if file.fault == testLifecycleCloseFault {
		return errors.New("injected close failure")
	}

	return nil
}

func TestLifecycleProtectedFileErrorBoundaries(t *testing.T) {
	t.Parallel()

	for _, fault := range []string{testLifecycleDirectoryFault, testLifecycleOpenFault, testLifecycleStatFault, testLifecycleReadFault, testLifecycleCloseFault, testLifecycleGrowthFault, testLifecycleShrinkFault} {
		t.Run(fault, func(t *testing.T) {
			t.Parallel()

			data := []byte("material")
			if fault == testLifecycleGrowthFault {
				data = []byte(strings.Repeat(testLifecycleFieldX, lifecycleMaxSecretBytes+1))
			}

			if fault == testLifecycleShrinkFault {
				data = nil
			}

			source := lifecycleFaultFS{files: fstest.MapFS{testLifecycleProtectedFile: &fstest.MapFile{Data: data, Mode: 0o600}}, fault: fault}
			got, err := readProtectedLifecycleFile(source, testLifecycleProtectedFile)
			require.Error(t, err)
			require.Nil(t, got)
		})
	}
}

func TestLifecycleSecretAggregateLimit(t *testing.T) {
	t.Parallel()

	model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSMemSelector)

	files := lifecycleSecretFiles(input)
	for _, file := range input.Secrets[0].Files {
		files[file].Data = []byte(strings.Repeat(testLifecycleFieldX, lifecycleMaxSecretBytes/2))
	}

	_, err := ImportLifecycleSecrets(model, inventory, input, files)
	require.ErrorContains(t, err, "exceeds 1 MiB")
}
