// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"bytes"
	json "encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestMainRejectsUnknownFlags(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer

	code := Main([]string{"-unknown"}, strings.NewReader(""), &stdout, &stderr)

	require.Equal(t, 2, code)
	require.Contains(t, stderr.String(), "flag provided")
}

func TestWriteDeploymentFileCreatesGeneratedDirectories(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "generated", "release", "values.yaml")
	data, err := json.Marshal(map[string]string{cryptoutilSharedMagic.WorkflowSubCmdRun: "fixture"})
	require.NoError(t, err)

	data = append(data, '\n')

	require.NoError(t, writeDeploymentFile(path, data, os.FileMode(cryptoutilSharedMagic.CICDOutputFilePermissions)))

	actual, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, data, actual)
}

func TestWriteDeploymentFileInjectedErrors(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "generated", "release", "values.yaml")
	data := []byte("fixture\n")
	injected := errors.New("injected filesystem failure")

	tests := []struct {
		name           string
		mkdirErr       error
		writeErr       error
		wantErr        string
		wantWriteCalls int
	}{
		{name: "mkdir failure", mkdirErr: injected, wantErr: "create generated output directory", wantWriteCalls: 0},
		{name: "write failure", writeErr: injected, wantErr: "write generated output", wantWriteCalls: 1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			mkdirCalls := 0
			writeCalls := 0
			mkdirAll := func(dir string, mode os.FileMode) error {
				mkdirCalls++

				require.Equal(t, filepath.Dir(path), dir)
				require.Equal(t, os.FileMode(cryptoutilSharedMagic.FilePermOwnerReadWriteExecuteGroupReadExecute), mode)

				return test.mkdirErr
			}
			writeFile := func(actualPath string, actualData []byte, mode os.FileMode) error {
				writeCalls++

				require.Equal(t, path, actualPath)
				require.Equal(t, data, actualData)
				require.Equal(t, os.FileMode(cryptoutilSharedMagic.CICDOutputFilePermissions), mode)

				return test.writeErr
			}

			err := writeDeploymentFileWithIO(
				path, data, os.FileMode(cryptoutilSharedMagic.CICDOutputFilePermissions), mkdirAll, writeFile,
			)

			require.ErrorContains(t, err, test.wantErr)
			require.ErrorIs(t, err, injected)
			require.Equal(t, 1, mkdirCalls)
			require.Equal(t, test.wantWriteCalls, writeCalls)
		})
	}
}
