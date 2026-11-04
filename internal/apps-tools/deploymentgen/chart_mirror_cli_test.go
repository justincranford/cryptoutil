// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	json "encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestGeneratorCLIChartMirrorModes(t *testing.T) {
	t.Parallel()

	model, _ := fixture(t)
	modelJSON, err := json.Marshal(model)
	require.NoError(t, err)

	repository := os.DirFS(filepath.Join("..", "..", ".."))
	artifacts, err := canonicalChartArtifacts(repository)
	require.NoError(t, err)

	artifactData := make(map[string][]byte, len(artifacts))

	artifactPaths := make([]string, len(artifacts))
	for index, artifact := range artifacts {
		artifactData[artifact.path] = artifact.data
		artifactPaths[index] = artifact.path
	}

	manifest, err := marshalCanonicalJSON(artifactPaths)
	require.NoError(t, err)

	tests := []struct {
		name        string
		args        []string
		wantCode    int
		stale       bool
		extra       bool
		writeFail   bool
		stdoutFail  bool
		wantMessage string
	}{
		{
			name: testPreviewCaseName,
			args: []string{cliFormatFlag, chartTemplatesFormat},
		},
		{
			name:        "write all charts",
			args:        []string{cliWriteFlag, cliFormatFlag, chartTemplatesFormat},
			wantMessage: cliGeneratedMessage,
		},
		{
			name:        testCheckCaseName,
			args:        []string{cliCheckFlag, cliFormatFlag, chartTemplatesFormat},
			wantMessage: cliPassMessage,
		},
		{
			name:  "stale chart copy",
			args:  []string{cliCheckFlag, cliFormatFlag, chartTemplatesFormat},
			stale: true, wantCode: 1, wantMessage: "missing or stale",
		},
		{
			name:        "unexpected chart copy",
			args:        []string{cliCheckFlag, cliFormatFlag, chartTemplatesFormat},
			extra:       true,
			wantCode:    1,
			wantMessage: unexpectedChartMirrorFileMessage,
		},
		{
			name:      "write chart copy failure",
			args:      []string{cliWriteFlag, cliFormatFlag, chartTemplatesFormat},
			writeFail: true, wantCode: 1, wantMessage: cliWriteArtifactMessage,
		},
		{
			name:       "chart status write failure",
			args:       []string{cliWriteFlag, cliFormatFlag, chartTemplatesFormat},
			stdoutFail: true, wantCode: 1, wantMessage: cliWriteStatusMessage,
		},
		{
			name:     "partial chart write prohibited",
			args:     []string{cliWriteFlag, cliFormatFlag, chartTemplatesFormat, cliSelectFlag, testServiceSMKMSPGSelector},
			wantCode: 2, wantMessage: cliInvalidFlagsMessage,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			source := chartMirrorTestFS(artifacts, test.extra)
			readFile := func(path string) ([]byte, error) {
				if filepath.Base(path) == k8sModelFileName {
					return modelJSON, nil
				}

				data, ok := artifactData[path]
				if !ok {
					return nil, fmt.Errorf("unexpected path: %s", path)
				}

				if test.stale {
					return append(append([]byte(nil), data...), ' '), nil
				}

				return data, nil
			}

			writes := 0
			writeFile := func(path string, data []byte, mode os.FileMode) error {
				writes++

				require.Equal(t, artifactData[path], data)
				require.Equal(t, os.FileMode(cryptoutilSharedMagic.CICDOutputFilePermissions), mode)

				if test.writeFail {
					return errors.New("injected chart output failure")
				}

				return nil
			}

			var (
				stdout bytes.Buffer
				stderr bytes.Buffer
			)

			var stdoutWriter interface{ Write([]byte) (int, error) } = &stdout
			if test.stdoutFail {
				stdoutWriter = failingWriter{}
			}

			code := internalMainWithJSONAndFS(
				test.args, stdoutWriter, &stderr, readFile, writeFile, marshalCanonicalJSON, source,
			)
			require.Equal(t, test.wantCode, code, stderr.String())

			if containsFlag(test.args, cliWriteFlag) && test.wantCode != 2 {
				wantWrites := len(artifacts)
				if test.writeFail {
					wantWrites = 1
				}

				require.Equal(t, wantWrites, writes)
			} else {
				require.Zero(t, writes)
			}

			if test.name == testPreviewCaseName {
				require.Equal(t, manifest, stdout.Bytes())
			}

			if test.wantMessage != "" {
				require.Contains(t, stdout.String()+stderr.String(), test.wantMessage)
			}
		})
	}
}

func chartMirrorTestFS(artifacts []generatedArtifact, includeExtra bool) fstest.MapFS {
	source := make(fstest.MapFS, len(artifacts)*2+1)
	for _, artifact := range artifacts {
		suffix := strings.TrimPrefix(artifact.path, "deployments/k8s/charts/")
		chart, relativePath, _ := strings.Cut(suffix, "/")
		canonicalPath := path.Join(canonicalChartSourceRoot, chart, relativePath)
		source[canonicalPath] = &fstest.MapFile{Data: artifact.data}
		source[artifact.path] = &fstest.MapFile{Data: artifact.data}
	}

	if includeExtra {
		source["deployments/k8s/charts/cryptoutil-runtime/templates/stale.yaml"] = &fstest.MapFile{Data: []byte("stale\n")}
	}

	return source
}

func containsFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}

	return false
}
