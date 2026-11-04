// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testLifecycleOperatorFile  = "operator.json"
	testLifecycleValuesPreview = "values preview"
	testLifecycleWriteCase     = cryptoutilSharedMagic.ScopeWrite
)

func TestLifecycleCLIPreviewWriteCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                       string
		args                       []string
		want                       int
		failRead, failWrite, stale bool
	}{
		{name: testLifecycleValuesPreview},
		{name: "secret file plan", args: []string{cliFormatFlag, lifecycleSecretPlanFormat}},
		{name: "revision preview", args: []string{cliFormatFlag, lifecycleRevisionFormat}},
		{name: testLifecycleWriteCase, args: []string{cliWriteFlag}},
		{name: "check", args: []string{cliCheckFlag}},
		{name: "stale check", args: []string{cliCheckFlag}, stale: true, want: 1},
		{name: "write failure", args: []string{cliWriteFlag}, failWrite: true, want: 1},
		{name: "read failure", failRead: true, want: 1},
		{name: "unknown format", args: []string{cliFormatFlag, "not-a-lifecycle-format"}, want: 2},
		{name: "forbid writing file plan", args: []string{cliFormatFlag, lifecycleSecretPlanFormat, cliWriteFlag}, want: 2},
		{name: "conflicting flags", args: []string{cliWriteFlag, cliCheckFlag}, want: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
			inputJSON, err := json.Marshal(input)
			require.NoError(t, err)
			modelJSON, err := json.Marshal(model)
			require.NoError(t, err)
			values, err := BuildLifecycleChartValues(model, inventory, input)
			require.NoError(t, err)
			valuesJSON, err := marshalCanonicalJSON(values)
			require.NoError(t, err)

			writes := 0
			read := func(file string) ([]byte, error) {
				if test.failRead {
					return nil, errors.New("read failed")
				}

				switch file {
				case testLifecycleOperatorFile:
					return inputJSON, nil
				case defaultPlannerPaths().model:
					return modelJSON, nil
				case lifecycleValuesSchemaPath:
					return os.ReadFile(filepath.Join("..", "..", "..", lifecycleValuesSchemaPath))
				default:
					if test.stale {
						return nil, errors.New("missing")
					}

					require.Equal(t, filepath.Join(
						"deployments", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, "generated", model.Run, inventory.Releases[0].Helm,
						"lifecycle-"+model.Revision+".json",
					), file)

					return valuesJSON, nil
				}
			}
			write := func(_ string, data []byte, mode os.FileMode) error {
				writes++

				require.Equal(t, valuesJSON, data)
				require.Equal(t, os.FileMode(0o600), mode)

				if test.failWrite {
					return errors.New("write failed")
				}

				return nil
			}

			args := append([]string{testLifecycleSubcommand, testLifecycleInputFlag, testLifecycleOperatorFile, cliSelectFlag, testServiceSMKMSMemSelector}, test.args...)

			var stdout, stderr bytes.Buffer

			code := Main(args, &stdout, &stderr, read, write)
			require.Equal(t, test.want, code, stderr.String())
			require.NotContains(t, stdout.String(), "fixture-material")

			if test.name == testLifecycleValuesPreview {
				require.Equal(t, valuesJSON, stdout.Bytes())
			}

			if test.name == testLifecycleWriteCase || test.failWrite {
				require.Equal(t, 1, writes)
			} else {
				require.Zero(t, writes)
			}
		})
	}
}
