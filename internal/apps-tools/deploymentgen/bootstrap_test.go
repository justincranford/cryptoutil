// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

func TestDecodeToolchainLock(t *testing.T) {
	t.Parallel()

	valid := toolchainLockFixture(t)
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{name: "valid pinned toolchain", data: valid},
		{name: "malformed JSON", data: []byte("{"), want: "toolchain lock"},
		{name: "unknown toolchain field", data: []byte(strings.Replace(string(valid), `"platform":`, `"unknown": true, "platform":`, 1)), want: testUnknownJSONField},
		{name: "duplicate field", data: []byte(strings.Replace(string(valid), `"schemaVersion": 1`, `"schemaVersion": 1, "schemaVersion": 1`, 1)), want: testDuplicateJSONKey},
		{name: "trailing JSON document", data: []byte(string(valid) + "{}"), want: testExactlyOneJSONDocument},
		{name: "unsupported schema", data: []byte(strings.Replace(string(valid), `"schemaVersion": 1`, `"schemaVersion": 2`, 1)), want: testInvalidSchemaVersion},
		{name: "invalid platform", data: []byte(strings.Replace(string(valid), `"linux/amd64"`, `"windows/amd64"`, 1)), want: testInvalidSchemaVersion},
		{name: "invalid K3s version", data: []byte(strings.Replace(string(valid), `"v1.37.1-k3s1"`, `"latest"`, 1)), want: testInvalidSchemaVersion},
		{name: "invalid K3s manifest digest", data: []byte(strings.Replace(string(valid), "8d298cb81ca94e875d31e804e5aa10b56355d1d7d95d7e1ef8a8d9239f15ada0", "bad", 1)), want: "invalid sha256 digest"},
		{name: "unpinned infrastructure image", data: []byte(strings.Replace(string(valid), "postgres:18@sha256:", "postgres:18:", 1)), want: testDigestPinned},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := DecodeToolchainLock(test.data)
			if test.want == "" {
				require.NoError(t, err)

				return
			}

			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestBuildLocalK3sPlan(t *testing.T) {
	t.Parallel()

	model, _ := fixture(t)
	lock, err := DecodeToolchainLock(toolchainLockFixture(t))
	require.NoError(t, err)

	plan, err := BuildLocalK3sPlan(model, lock)
	require.NoError(t, err)
	require.Equal(t, "cu-review-cluster", plan.Project)
	require.Equal(t, "cu-review-nodes", plan.Network)
	require.Equal(t, lock.Kubernetes.K3S.Image, plan.Image)
	require.Equal(t, "cu-review-cluster-server-0", plan.Server.Container)
	require.Len(t, plan.Agents, 2)
	require.Equal(t, "cu-review-cluster-agent-0", plan.Agents[0].Container)
	require.Equal(t, "cu-review-cluster-agent-1", plan.Agents[1].Container)
	require.Equal(t, "./secrets/review/k3s-join-token", plan.TokenSecretFile)
	require.Equal(t, 6445, plan.APIPort)
	require.Equal(t, 8443, plan.GatewayPort)
}

func TestBuildLocalK3sPlanErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edit func(*Model, *ToolchainLock)
		want string
	}{
		{
			name: "invalid run identity",
			edit: func(model *Model, _ *ToolchainLock) { model.Run = "INVALID" },
			want: "run identity",
		},
		{
			name: testNonLoopbackAPIHostCaseName,
			edit: func(model *Model, _ *ToolchainLock) { model.Bootstrap.APIHost = testNonLoopbackIPv4 },
			want: testLoopbackIPv4Error,
		},
		{
			name: "unpinned k3s image",
			edit: func(_ *Model, lock *ToolchainLock) { lock.Kubernetes.K3S.Image = "rancher/k3s:latest" },
			want: testDigestPinned,
		},
		{
			name: "invalid lock schema",
			edit: func(_ *Model, lock *ToolchainLock) { lock.SchemaVersion = 2 },
			want: testInvalidSchemaVersion,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, _ := fixture(t)
			lock, err := DecodeToolchainLock(toolchainLockFixture(t))
			require.NoError(t, err)
			test.edit(&model, &lock)

			_, err = BuildLocalK3sPlan(model, lock)
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestRenderLocalK3sCompose(t *testing.T) {
	t.Parallel()

	model, _ := fixture(t)
	lock, err := DecodeToolchainLock(toolchainLockFixture(t))
	require.NoError(t, err)
	plan, err := BuildLocalK3sPlan(model, lock)
	require.NoError(t, err)
	source, err := os.ReadFile(filepath.Join("..", "..", "..", localK3sComposeTemplate))
	require.NoError(t, err)

	first, err := RenderLocalK3sCompose(source, plan)
	require.NoError(t, err)
	second, err := RenderLocalK3sCompose(source, plan)
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.NotContains(t, string(first), "\r")
	require.NotContains(t, string(first), cryptoutilSharedMagic.DockerServicePostgres)
	require.NotContains(t, string(first), "otel")
	require.NotContains(t, string(first), "cryptoutil app")
	require.Contains(t, string(first), `"127.0.0.1"`)
	require.Contains(t, string(first), `"6445"`)
	require.Contains(t, string(first), `"8443"`)
	require.Contains(t, string(first), "cu-review-cluster-agent-0")
	require.Contains(t, string(first), "30443")
	require.Contains(t, string(first), plan.TokenSecretFile)
	require.Contains(t, string(first), cryptoutilSharedMagic.K3SBootstrapTokenFileArgument)
	require.Contains(t, string(first), cryptoutilSharedMagic.K3SBootstrapAgentServerArgument)
	require.Contains(t, string(first), cryptoutilSharedMagic.K3SBootstrapServerEncryptionArgument)
	require.Contains(t, string(first), cryptoutilSharedMagic.K3SBootstrapServerEncryptionProviderArgument)
}

func TestRenderLocalK3sComposeErrors(t *testing.T) {
	t.Parallel()

	model, _ := fixture(t)
	lock, err := DecodeToolchainLock(toolchainLockFixture(t))
	require.NoError(t, err)
	plan, err := BuildLocalK3sPlan(model, lock)
	require.NoError(t, err)
	source, err := os.ReadFile(filepath.Join("..", "..", "..", localK3sComposeTemplate))
	require.NoError(t, err)

	tests := []struct {
		name   string
		source []byte
		want   string
	}{
		{name: "empty source", source: nil, want: "template is empty"},
		{name: "template parse error", source: []byte("{{ if }}"), want: "parse local K3s"},
		{name: "missing template key", source: []byte("{{ .Missing }}\n"), want: "render local K3s"},
		{name: "missing final LF", source: []byte(cryptoutilSharedMagic.DevSetupTestInvalidName), want: testRenderedLFError},
		{
			name: "CRLF forbidden", source: []byte(cryptoutilSharedMagic.DevSetupTestInvalidName + "\r\n"),
			want: testRenderedLFError,
		},
		{name: "invalid rendered YAML", source: []byte("invalid: [\n"), want: "parse rendered local K3s Compose YAML"},
		{
			name:   "invalid server port mapping",
			source: []byte(strings.Replace(string(source), "target: 6443", "target: 6444", 1)),
			want:   "API/gateway ports",
		},
		{
			name:   "invalid Compose project identity",
			source: []byte(strings.Replace(string(source), `printf "%q" .Project`, `"wrong"`, 1)),
			want:   "project or node count",
		},
		{
			name:   "invalid rendered node identity",
			source: []byte(strings.Replace(string(source), `printf "%q" .Server.Container`, `"wrong"`, 1)),
			want:   "does not match plan",
		},
		{
			name: "token environment forbidden",
			source: []byte(strings.Replace(
				string(source),
				"    command:\n      - server",
				"    environment:\n      "+cryptoutilSharedMagic.K3SBootstrapTokenEnvironment+": insecure\n    command:\n      - server",
				1,
			)),
			want: "must not expose the join token through its environment",
		},
		{
			name: "agent URL environment forbidden",
			source: []byte(strings.Replace(
				string(source),
				"    command:\n      - agent",
				"    environment:\n      "+cryptoutilSharedMagic.K3SBootstrapURLEnvironment+": "+cryptoutilSharedMagic.K3SBootstrapServerURL+"\n    command:\n      - agent",
				1,
			)),
			want: "agent URL must use the CLI argument",
		},
		{
			name: "missing join-token secret declaration",
			source: []byte(strings.Replace(
				string(source),
				"\nsecrets:\n  # Create this run-owned file with owner-only permissions; never commit its contents.\n  "+
					cryptoutilSharedMagic.K3SBootstrapTokenSecretName+
					":\n    file: {{ printf \"%q\" .TokenSecretFile }}\n",
				"\n",
				1,
			)),
			want: testBootstrapProtectedRelativeJoinTokenFileError,
		},
		{
			name: "absolute join-token secret source",
			source: []byte(strings.Replace(
				string(source), `{{ printf "%q" .TokenSecretFile }}`, `"/tmp/k3s-join-token"`, 1,
			)),
			want: testBootstrapProtectedRelativeJoinTokenFileError,
		},
		{
			name: "traversal join-token secret source",
			source: []byte(strings.Replace(
				string(source), `{{ printf "%q" .TokenSecretFile }}`, `"./secrets/../k3s-join-token"`, 1,
			)),
			want: testBootstrapProtectedRelativeJoinTokenFileError,
		},
		{
			name:   "join-token secret mounted at wrong path",
			source: []byte(strings.Replace(string(source), "target: "+cryptoutilSharedMagic.K3SBootstrapTokenSecretTarget, "target: /tmp/k3s-join-token", 1)),
			want:   "must mount the join-token Docker Secret read-only at the protected path",
		},
		{
			name:   "server token file flag missing",
			source: []byte(strings.Replace(string(source), "      - "+cryptoutilSharedMagic.K3SBootstrapTokenFileArgument+"\n", "", 1)),
			want:   "rendered server must use the protected token-file CLI argument",
		},
		{
			name:   "server token argument forbidden",
			source: []byte(strings.Replace(string(source), cryptoutilSharedMagic.K3SBootstrapTokenFileArgument, cryptoutilSharedMagic.K3SBootstrapTokenArgument, 1)),
			want:   testBootstrapNoTokenInCommandError,
		},
		{
			name:   "server token assignment forbidden",
			source: []byte(strings.Replace(string(source), cryptoutilSharedMagic.K3SBootstrapTokenFileArgument, cryptoutilSharedMagic.K3SBootstrapTokenAssignmentPrefix, 1)),
			want:   testBootstrapNoTokenInCommandError,
		},
		{
			name:   "server Secret encryption missing",
			source: []byte(strings.Replace(string(source), "      - "+cryptoutilSharedMagic.K3SBootstrapServerEncryptionArgument+"\n", "", 1)),
			want:   testBootstrapSecretEncryptionError,
		},
		{
			name:   "server Secret encryption provider changed",
			source: []byte(strings.Replace(string(source), cryptoutilSharedMagic.K3SBootstrapServerEncryptionProviderArgument, "--secrets-encryption-provider=secretbox", 1)),
			want:   testBootstrapSecretEncryptionError,
		},
		{
			name:   "agent server URL CLI argument missing",
			source: []byte(strings.Replace(string(source), "      - "+cryptoutilSharedMagic.K3SBootstrapAgentServerArgument+"\n", "", 1)),
			want:   testBootstrapAgentArgumentsError,
		},
		{
			name:   "agent token file CLI argument missing",
			source: []byte(strings.Replace(string(source), "      - "+cryptoutilSharedMagic.K3SBootstrapTokenFileArgument+"\n    depends_on:", "    depends_on:", 1)),
			want:   testBootstrapAgentArgumentsError,
		},
		{
			name: "agent publishes invalid server command",
			source: []byte(strings.Replace(
				string(source),
				"      - agent\n      - "+cryptoutilSharedMagic.K3SBootstrapAgentServerArgument,
				"      - server\n      - "+cryptoutilSharedMagic.K3SBootstrapAgentServerArgument,
				1,
			)),
			want: "rendered agent must be a non-published K3s agent",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := RenderLocalK3sCompose(test.source, plan)
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestCheckBootstrapPorts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		listenError  int
		closeError   bool
		nilListener  bool
		listenAbsent bool
		invalidHost  bool
		want         string
	}{
		{name: "bind and release both endpoints"},
		{name: "second endpoint collision", listenError: 2, want: "host endpoint unavailable"},
		{name: "listener close error", closeError: true, want: "close bootstrap port probe"},
		{name: "nil listener result", nilListener: true, want: "nil listener"},
		{name: "missing listen dependency", listenAbsent: true, want: "listener is required"},
		{name: testInvalidBootstrapCaseName, invalidHost: true, want: testLoopbackIPv4Error},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			model, _ := fixture(t)

			bootstrap := model.Bootstrap
			if test.invalidHost {
				bootstrap.APIHost = testNonLoopbackIPv4
			}

			var calls, closes int

			listen := func(network, address string) (net.Listener, error) {
				calls++

				require.Equal(t, "tcp", network)
				require.Contains(t, []string{"127.0.0.1:6445", "127.0.0.1:8443"}, address)

				if calls == test.listenError {
					return nil, errors.New("injected address collision")
				}

				if test.nilListener {
					return nil, nil
				}

				return &bootstrapTestListener{
					close: func() error {
						closes++

						if test.closeError {
							return errors.New("injected close error")
						}

						return nil
					},
				}, nil
			}

			listenFn := listen
			if test.listenAbsent {
				listenFn = nil
			}

			err := CheckBootstrapPorts(bootstrap, listenFn)
			if test.want != "" {
				require.ErrorContains(t, err, test.want)

				return
			}

			require.NoError(t, err)
			require.Equal(t, 2, calls)
			require.Equal(t, 2, closes)
		})
	}
}

func toolchainLockFixture(t *testing.T) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("..", "..", "..", "api", "cryptosuite-registry", cryptoutilSharedMagic.KubernetesSuiteImageDeployment, k8sToolchainLockFileName))
	require.NoError(t, err)

	return data
}

type bootstrapTestListener struct {
	close func() error
}

func (listener *bootstrapTestListener) Accept() (net.Conn, error) {
	return nil, errors.New("not implemented")
}

func (listener *bootstrapTestListener) Close() error {
	return listener.close()
}

func (listener *bootstrapTestListener) Addr() net.Addr {
	return &net.TCPAddr{}
}
