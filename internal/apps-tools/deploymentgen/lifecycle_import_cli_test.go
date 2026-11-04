// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	json "encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const (
	testLifecycleMissingRootCase = "missing root"
	testLifecycleSecretRootFlag  = "-secret-root"
	testLifecyclePrivateFixture  = "private fixture must never appear in diagnostics"
)

func TestLifecycleCLIProtectedImportBoundary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		args                   []string
		failImport, failOutput bool
		want                   int
		calls                  int
	}{
		{name: "memory import only", calls: 1},
		{name: "sanitized import failure", failImport: true, want: 1, calls: 1},
		{name: "sanitized output failure", failOutput: true, want: 1, calls: 1},
		{name: testLifecycleMissingRootCase, args: []string{testLifecycleInputFlag, testLifecycleOperatorFile, cliFormatFlag, lifecycleImportCheckFormat}, want: 2},
		{name: "no artifact persistence", args: []string{cliWriteFlag}, want: 2},
		{name: "no artifact checking", args: []string{cliCheckFlag}, want: 2},
		{name: "root forbidden for values", args: []string{cliFormatFlag, lifecycleValuesFormat}, want: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, inventory, input := lifecycleDeployFixture(t, testServiceSMKMSMemSelector)
			inputJSON, err := json.Marshal(input)
			require.NoError(t, err)
			modelJSON, err := json.Marshal(model)
			require.NoError(t, err)

			read := func(file string) ([]byte, error) {
				switch file {
				case testLifecycleOperatorFile:
					return inputJSON, nil
				case defaultPlannerPaths().model:
					return modelJSON, nil
				default:
					require.FailNowf(t, cryptoutilSharedMagic.CICDTestFatalfFailureMessage, "import-check must not read charts or material via ordinary file reader: %s", file)

					return nil, nil
				}
			}
			calls := 0
			importer := func(gotModel Model, gotInventory Inventory, gotInput LifecycleInput, root string) ([]LifecycleSecret, error) {
				calls++

				require.Equal(t, model, gotModel)
				require.Equal(t, inventory, gotInventory)
				require.Equal(t, input, gotInput)
				require.Equal(t, "protected-root", root)

				if test.failImport {
					return nil, errors.New(testLifecyclePrivateFixture)
				}

				return []LifecycleSecret{{Data: map[string][]byte{"private-key": []byte(testLifecyclePrivateFixture)}}}, nil
			}

			args := append([]string{
				testLifecycleInputFlag, testLifecycleOperatorFile, cliSelectFlag, testServiceSMKMSMemSelector,
				cliFormatFlag, lifecycleImportCheckFormat, testLifecycleSecretRootFlag, "protected-root",
			}, test.args...)
			if test.name == testLifecycleMissingRootCase {
				args = test.args
			}

			var stdout, stderr bytes.Buffer

			var output io.Writer = &stdout
			if test.failOutput {
				output = failingWriter{}
			}

			code := lifecycleMainWithImporter(args, output, &stderr, read, func(string, []byte, os.FileMode) error {
				t.Fatal("protected import must not persist Secret data")

				return nil
			}, importer)
			require.Equal(t, test.want, code, stderr.String())
			require.Equal(t, test.calls, calls)
			require.NotContains(t, stdout.String()+stderr.String(), testLifecyclePrivateFixture)

			if test.want == 0 {
				require.Contains(t, stdout.String(), "validated 1 protected lifecycle Secrets in memory")
			}
		})
	}
}
