// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"bytes"
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommandSQLiteContract(t *testing.T) {
	t.Parallel()

	for _, forged := range []bool{false, true} {
		t.Run(map[bool]string{false: "explicit SQLite", true: "PG not allowed"}[forged], func(t *testing.T) {
			t.Parallel()

			configRaw := []byte(strings.ReplaceAll(validInstanceYAML, "pg", cryptoutilSharedMagic.DeploymentPoolMem))
			config, err := DecodeInstanceConfig(configRaw)
			require.NoError(t, err)

			input := commandFixture(t, config)
			for index := range input.Pods {
				input.Pods[index].PostgreSQLClientBindings = nil
			}

			if forged {
				input.Pods[0].PostgreSQLClientBindings = map[string]PostgreSQLClientBinding{"unowned-pod": {CommonName: "unowned-client", DatabaseRole: "unowned-role"}}
			}

			err = ValidateCommandInput(input, config)
			if forged {
				require.ErrorContains(t, err, "SQLite materializer")

				return
			}

			require.NoError(t, err)
			raw, err := json.Marshal(input)
			require.NoError(t, err)

			var stdout, stderr bytes.Buffer

			code := MainWithDependencies([]string{"--instance=" + InstancePath, "--input=" + InputPath}, &stdout, &stderr,
				func(key string) string {
					if key == NamespaceEnvironment {
						return input.Namespace
					}

					return input.Pods[0].PodName
				},
				func(name string) ([]byte, error) {
					if name == InstancePath {
						return configRaw, nil
					}

					if name == InputPath {
						return raw, nil
					}

					return []byte("explicit anchor"), nil
				},
				func(_ context.Context, _ []byte, _ RuntimeInputs, trust TrustInputs) error {
					require.Empty(t, trust.PostgreSQLClientPEM)

					return nil
				})
			require.Zero(t, code, stderr.String())
		})
	}
}

func TestCommandProcessMain(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	require.Zero(t, Main([]string{testCommandHelp}, nil, &stdout, &stderr))
	require.Contains(t, stdout.String(), Command)
	require.Equal(t, 1, Main(nil, nil, &stdout, &stderr))
}
