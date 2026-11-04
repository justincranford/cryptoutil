// Copyright (c) 2025-2026 Justin Cranford.

package materializer

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func commandFixture(t *testing.T, config InstanceConfig) CommandInput {
	t.Helper()

	input := CommandInput{Namespace: testNamespace, APIEndpoint: testAPIEndpoint, APIServerName: testAPIServerName, TrustFiles: map[string]string{}}

	for _, role := range []string{rolePublic, roleAdmin, roleHealth, roleOTLP, rolePostgres} {
		if config.Pool != cryptoutilSharedMagic.DeploymentMaterializerPoolPG && role == rolePostgres {
			continue
		}

		input.TrustFiles[role] = TrustDirectory + "/" + role + cryptoutilSharedMagic.PKIInitPublicCertFileExtension
	}

	bindings := map[string]PostgreSQLClientBinding{}

	for _, service := range config.InstanceServices {
		cn := "client-" + service.PodName
		if service.PodName == testPostgreSQLPodNameOne {
			cn = testMaterializerTestLeafCommonName
		}

		bindings[service.PodName] = PostgreSQLClientBinding{CommonName: cn, DatabaseRole: testIdentityAuthzClientDatabaseRole}
	}

	for _, service := range config.InstanceServices {
		input.Pods = append(input.Pods, CommandPod{
			PodName:                  service.PodName,
			ExpectedAdminDNS:         testExpectedAdministrativeDNS(t, config, service),
			ExpectedPublicDNSNames:   testExpectedPublicDNSNames(config, service, testNamespace),
			PostgreSQLClientBindings: bindings,
		})
	}

	return input
}

func TestMaterializerCommandRealConsumer(t *testing.T) {
	t.Parallel()
	harness := newMaterializerHarness(t, newTestPKIFixture(t))
	input := commandFixture(t, harness.fixture.config)
	raw, err := json.Marshal(input)
	require.NoError(t, err)

	var stdout, stderr bytes.Buffer

	reads := []string{}
	code := MainWithDependencies([]string{"--instance=" + InstancePath, "--input=" + InputPath}, &stdout, &stderr,
		func(key string) string {
			if key == PodNameEnvironment {
				return harness.inputs.PodName
			}

			return testNamespace
		},
		func(name string) ([]byte, error) {
			reads = append(reads, name)
			switch name {
			case InstancePath:
				return harness.configYAML, nil
			case InputPath:
				return raw, nil
			case input.TrustFiles[rolePublic]:
				return harness.trust.PublicServerPEM, nil
			case input.TrustFiles[roleAdmin]:
				return harness.trust.AdminServerPEM, nil
			case input.TrustFiles[roleHealth]:
				return harness.trust.HealthClientPEM, nil
			case input.TrustFiles[roleOTLP]:
				return harness.trust.OTLPClientPEM, nil
			default:
				return harness.trust.PostgreSQLClientPEM, nil
			}
		}, harness.materializer().Materialize)
	require.Zero(t, code, stderr.String())
	require.Empty(t, stderr.String())
	require.Len(t, reads, 7)
	require.Len(t, harness.doer.requests, 3)
	require.Equal(t, []string{cryptoutilSharedMagic.DeploymentMaterializerProjectedAPICAPath, cryptoutilSharedMagic.DeploymentMaterializerProjectedTokenPath}, harness.readPaths)
	require.FileExists(t, filepath.Join(harness.output, "public-server", cryptoutilSharedMagic.DeploymentMaterializerSecretDataCertificate))
	require.NotContains(t, stdout.String(), cryptoutilSharedMagic.StringPEMTypePKCS8PrivateKey)
}

func TestCommandInputRejections(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*CommandInput, *InstanceConfig)
	}{
		{"invalid command config", func(_ *CommandInput, c *InstanceConfig) { c.PSID = "" }},
		{"invalid command namespace", func(i *CommandInput, _ *InstanceConfig) { i.Namespace = "" }},
		{"TLS endpoint", func(i *CommandInput, _ *InstanceConfig) { i.APIEndpoint = "http://localhost" }},
		{"trust count", func(i *CommandInput, _ *InstanceConfig) { delete(i.TrustFiles, roleHealth) }},
		{"trust escape", func(i *CommandInput, _ *InstanceConfig) { i.TrustFiles[roleHealth] = TrustDirectory + "/../ca.crt" }},
		{"missing pod", func(i *CommandInput, _ *InstanceConfig) { i.Pods = i.Pods[:1] }},
		{"duplicate pod", func(i *CommandInput, _ *InstanceConfig) { i.Pods[1] = i.Pods[0] }},
		{"unowned pod", func(i *CommandInput, _ *InstanceConfig) { i.Pods[0].PodName = "not-owned" }},
		{"admin identity", func(i *CommandInput, _ *InstanceConfig) { i.Pods[0].ExpectedAdminDNS = "unowned.example" }},
		{"public identity", func(i *CommandInput, _ *InstanceConfig) { i.Pods[0].ExpectedPublicDNSNames = nil }},
		{"PG binding", func(i *CommandInput, _ *InstanceConfig) { i.Pods[0].PostgreSQLClientBindings = nil }},
		{"SQLite PG input", func(i *CommandInput, c *InstanceConfig) {
			c.Pool = cryptoutilSharedMagic.DeploymentMaterializerPoolMem

			delete(i.TrustFiles, rolePostgres)
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			config, err := DecodeInstanceConfig([]byte(validInstanceYAML))
			require.NoError(t, err)
			input := commandFixture(t, config)
			test.mutate(&input, &config)
			require.Error(t, ValidateCommandInput(input, config))
		})
	}
}

func TestDecodeCommandInputStrict(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		"", strings.Repeat(" ", cryptoutilSharedMagic.DeploymentMaterializerMaximumInstanceYAMLBytes+1), "{", `{"namespace":"a","namespace":"b"}`,
		`{"pods":[{"podName":"a","podName":"b"}]}`, `{"unknown":true}`, `{} {}`, `[]`,
	} {
		t.Run(raw[:min(len(raw), 32)], func(t *testing.T) { t.Parallel(); _, err := DecodeCommandInput([]byte(raw)); require.Error(t, err) })
	}
}

func TestMaterializerCommandBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		fail string
		want int
	}{
		{cryptoutilSharedMagic.CLIHelpCommand, []string{testCommandHelp}, "", 0},
		{"help write", []string{testCommandHelp}, testCommandWriter, 1},
		{"unknown flag", []string{"--secret=DO-NOT-LOG"}, "", 1},
		{testCommandFixedPaths, nil, "", 1},
		{"positional", []string{"--instance=" + InstancePath, "--input=" + InputPath, "extra"}, "", 1},
		{"instance read", nil, testCommandInstanceRead, 1},
		{"instance decode", nil, testCommandInstanceDecode, 1},
		{"input read", nil, testCommandInputRead, 1},
		{"input decode", nil, testCommandInputDecode, 1},
		{"input contract", nil, testCommandContract, 1},
		{"Downward namespace differs", nil, testCommandNamespace, 1},
		{"pod mismatch", nil, testCommandPod, 1},
		{"trust read", nil, testCommandTrust, 1},
		{"consumer error", nil, testCommandConsumer, 1},
		{testCommandSuccess, nil, "", 0},
		{testCommandSuccessWrite, nil, testCommandWriter, 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			config, err := DecodeInstanceConfig([]byte(validInstanceYAML))
			require.NoError(t, err)

			input := commandFixture(t, config)
			if test.fail == testCommandContract {
				input.TrustFiles = nil
			}

			raw, err := json.Marshal(input)
			require.NoError(t, err)

			args := test.args
			if args == nil && test.name != testCommandFixedPaths {
				args = []string{"--instance=" + InstancePath, "--input=" + InputPath}
			}

			var (
				stdout, stderr bytes.Buffer
				output         io.Writer = &stdout
			)
			if test.fail == testCommandWriter {
				output = commandErrorWriter{}
			}

			calls := 0
			code := MainWithDependencies(args, output, &stderr, func(key string) string {
				if key == NamespaceEnvironment {
					if test.fail == testCommandNamespace {
						return "wrong-namespace"
					}

					return testNamespace
				}

				if test.fail == testCommandPod {
					return "wrong-pod"
				}

				return input.Pods[0].PodName
			}, func(name string) ([]byte, error) {
				if name == InstancePath {
					if test.fail == testCommandInstanceRead {
						return nil, errors.New("DO-NOT-LOG")
					}

					if test.fail == testCommandInstanceDecode {
						return []byte(cryptoutilSharedMagic.DevSetupTestInvalidName), nil
					}

					return []byte(validInstanceYAML), nil
				}

				if name == InputPath {
					if test.fail == testCommandInputRead {
						return nil, errors.New("DO-NOT-LOG")
					}

					if test.fail == testCommandInputDecode {
						return []byte(cryptoutilSharedMagic.DevSetupTestInvalidName), nil
					}

					return raw, nil
				}

				if test.fail == testCommandTrust {
					return nil, errors.New("DO-NOT-LOG")
				}

				return []byte("explicit-anchor"), nil
			}, func(ctx context.Context, _ []byte, _ RuntimeInputs, _ TrustInputs) error {
				calls++
				_, bounded := ctx.Deadline()
				require.True(t, bounded)

				if test.fail == testCommandConsumer {
					return errors.New("DO-NOT-LOG")
				}

				return nil
			})
			require.Equal(t, test.want, code)
			require.NotContains(t, stderr.String(), "DO-NOT-LOG")

			if test.name == testCommandSuccess || test.name == testCommandSuccessWrite || test.fail == testCommandConsumer {
				require.Equal(t, 1, calls)
			} else {
				require.Zero(t, calls)
			}
		})
	}
}

type commandErrorWriter struct{}

func (commandErrorWriter) Write([]byte) (int, error) { return 0, errors.New("DO-NOT-LOG") }

func TestBoundedCommandRead(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name      string
		size      int
		missing   bool
		directory bool
		want      bool
	}{
		{"small", 1, false, false, false},
		{"limit", cryptoutilSharedMagic.DeploymentMaterializerMaximumInstanceYAMLBytes, false, false, false},
		{"oversized command file", cryptoutilSharedMagic.DeploymentMaterializerMaximumInstanceYAMLBytes + 1, false, false, true},
		{"missing", 0, true, false, true},
		{"directory", 0, false, true, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			name := filepath.Join(t.TempDir(), "source")
			if test.directory {
				require.NoError(t, os.Mkdir(name, 0o700))
			} else if !test.missing {
				require.NoError(t, os.WriteFile(name, bytes.Repeat([]byte("x"), test.size), 0o600))
			}

			raw, err := boundedCommandRead(name)
			if test.want {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, raw, test.size)
			}
		})
	}
}
