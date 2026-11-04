// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

const (
	testLifecycleIdentityChanged = "identity or type changed"
	testLifecycleUnsupportedACE  = "unsupported"
)

type lifecycleRootFault struct {
	*os.Root
	lstat func(string) (fs.FileInfo, error)
	open  func(string) (*os.File, error)
	files fs.FS
}

func (root lifecycleRootFault) Lstat(name string) (fs.FileInfo, error) {
	if root.lstat != nil {
		return root.lstat(name)
	}

	return root.Root.Lstat(name)
}

func (root lifecycleRootFault) Open(name string) (*os.File, error) {
	if root.open != nil {
		return root.open(name)
	}

	return root.Root.Open(name)
}

func (root lifecycleRootFault) FS() fs.FS {
	if root.files != nil {
		return root.files
	}

	return root.Root.FS()
}

func TestLifecyclePhysicalHandleAndDirectoryFailures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*testing.T, *lifecycleRootFault)
		want string
	}{
		{"open unavailable", func(_ *testing.T, root *lifecycleRootFault) {
			root.open = func(string) (*os.File, error) { return nil, errors.New("injected open") }
		}, "open protected"},
		{"stat unavailable", func(t *testing.T, root *lifecycleRootFault) {
			root.open = func(name string) (*os.File, error) {
				file, err := root.Root.Open(name)
				require.NoError(t, err)
				require.NoError(t, file.Close())

				return file, nil
			}
		}, "inspect protected"},
		{"name replaced after open", func(t *testing.T, root *lifecycleRootFault) {
			calls := 0
			root.lstat = func(name string) (fs.FileInfo, error) {
				calls++
				if calls == 2 {
					return root.Root.Lstat("other-file")
				}

				return root.Root.Lstat(name)
			}
		}, testLifecycleIdentityChanged},
		{"name vanished after open", func(_ *testing.T, root *lifecycleRootFault) {
			calls := 0
			root.lstat = func(name string) (fs.FileInfo, error) {
				calls++
				if calls == 2 {
					return nil, errors.New("injected vanished name")
				}

				return root.Root.Lstat(name)
			}
		}, testLifecycleIdentityChanged},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			directory := t.TempDir()
			protectLifecycleTestPath(t, directory, "")

			for _, name := range []string{testLifecycleMaterialFile, "other-file"} {
				file := filepath.Join(directory, name)
				require.NoError(t, os.WriteFile(file, []byte(testLifecyclePhysicalMaterial), 0o600))
				protectLifecycleTestPath(t, file, "")
			}

			root, err := os.OpenRoot(directory)
			require.NoError(t, err)

			defer func() { require.NoError(t, root.Close()) }()

			fault := lifecycleRootFault{Root: root}
			test.edit(t, &fault)
			_, err = (protectedLifecycleFS{root: fault}).openChecked(testLifecycleMaterialFile, false)
			require.ErrorContains(t, err, test.want)
		})
	}

	directory := t.TempDir()
	protectLifecycleTestPath(t, directory, "")
	root, err := os.OpenRoot(directory)
	require.NoError(t, err)

	defer func() { require.NoError(t, root.Close()) }()

	source := protectedLifecycleFS{root: lifecycleRootFault{Root: root, files: fstest.MapFS{}}}
	_, err = source.ReadDir("../invalid")
	require.Error(t, err)
	_, err = source.ReadDir("absent")
	require.Error(t, err)

	source.root = lifecycleRootFault{Root: root, files: lifecycleFaultFS{files: fstest.MapFS{}, fault: testLifecycleDirectoryFault}}
	_, err = source.ReadDir(".")
	require.ErrorContains(t, err, "read protected directory")
	file, err := root.Open(".")
	require.NoError(t, err)

	view := protectedLifecycleFile{File: file}
	info, err := view.Stat()
	require.NoError(t, err)
	require.Equal(t, fs.ModeDir|lifecyclePrivateDirectoryMode, info.Mode())
	require.NoError(t, file.Close())

	_, err = view.Stat()
	require.Error(t, err)
	require.Error(t, validatePhysicalLifecycleHandle(file, info, true))

	model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
	model.Revision = testLifecycleOldRevision
	_, err = ImportPhysicalLifecycleSecrets(model, inventory, input, directory)
	require.Error(t, err)

	var classified protectedLifecycleImportError

	require.ErrorAs(t, err, &classified)
	require.NotNil(t, classified.Unwrap())
}

func TestLifecycleWindowsNativeGrantEnumeration(t *testing.T) {
	t.Parallel()

	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	require.NoError(t, err)

	tests := []struct {
		name, sddl string
		count      int
		want       string
	}{
		{"no DACL", "O:SY", 0, "conversion"},
		{"incomplete entry", "D:P(A;;FA;;;SY", 1, "incomplete"},
		{"unsupported callback", "D:P(XA;;FA;;;SY)", 1, testLifecycleUnsupportedACE},
		{"unsupported object GUID", "D:P(A;;FA;guid;;SY)", 1, testLifecycleUnsupportedACE},
		{"invalid principal", "D:P(A;;FA;;;not-a-sid)", 1, "principal"},
		{"count mismatch", "D:P", 1, "enumeration"},
		{"SACL excluded", "D:P(A;;FA;;;SY)S:(AU;SA;FR;;;WD)", 1, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateLifecycleWindowsGrants(test.sddl, user.User.Sid, test.count)
			if test.want != "" {
				require.ErrorContains(t, err, test.want)

				return
			}

			require.NoError(t, err)
		})
	}
}
