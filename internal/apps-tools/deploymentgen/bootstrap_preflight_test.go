// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testBootstrapAPIEndpoint       = "127.0.0.1:6445"
	testBootstrapGatewayEndpoint   = "127.0.0.1:8443"
	testBootstrapInsufficientHost  = "exceed Docker host capacity"
	testBootstrapContainerExists   = "container name already exists"
	testBootstrapUnrelatedName     = "unrelated"
	testBootstrapMissingDependency = "context, runner, and listener are required"
	testBootstrapCanceled          = "preflight canceled"
	testBootstrapDecodeOwnership   = "decode bootstrap container ownership"
	testBootstrapMissingName       = "ownership record has no name"
	testBootstrapProjectLabel      = "project"
	testBootstrapProjectPrefix     = "prefix"
	testBootstrapNamesKey          = "Names"
	testBootstrapNameKey           = "Name"
	testBootstrapLabelsKey         = "Labels"
	testBootstrapAgentMarker       = "agent"
	testBootstrapServerMarker      = cryptoutilSharedMagic.CMD_SERVER
)

func TestCheckLocalK3sPreflight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fullMatrix bool
		edit       func(*BootstrapPreflightInput)
		failAt     int
		outputAt   int
		output     string
		failListen int
		closeError bool
		wantErr    string
		wantCalls  int
		wantListen int
		wantCloses int
	}{
		{name: "selected release admitted", wantCalls: 4, wantListen: 2, wantCloses: 2},
		{name: "full matrix rejected without downsizing", fullMatrix: true, wantErr: testBootstrapInsufficientHost, wantCalls: 4},
		{name: "run mismatch", edit: func(input *BootstrapPreflightInput) { input.Inventory.Run = testDifferentIdentity }, wantErr: "run identities differ"},
		{name: "invalid plan", edit: func(input *BootstrapPreflightInput) { input.Model.Run = "bad-run" }, wantErr: "run identity must"},
		{name: "invalid estimate", edit: func(input *BootstrapPreflightInput) { input.RuntimeValues = nil }, wantErr: "cover every selected release"},
		{name: "missing reserves", edit: func(input *BootstrapPreflightInput) { input.Reserve = ResourceCapacity{} }, wantErr: "explicit positive host and K3s"},
		{name: "container daemon failure", failAt: 1, wantErr: "read bootstrap container ownership", wantCalls: 1},
		{name: "network daemon failure", failAt: 2, wantErr: "read bootstrap network ownership", wantCalls: 2},
		{name: "volume daemon failure", failAt: testThree, wantErr: "read bootstrap volume ownership", wantCalls: testThree},
		{name: "host daemon failure", failAt: 4, wantErr: "read Docker host capacity: injected", wantCalls: 4},
		{name: "invalid host measurement", outputAt: 4, output: cryptoutilSharedMagic.DevSetupTestInvalidName, wantErr: "CPU and memory values", wantCalls: 4},
		{name: "occupied API port", failListen: 1, wantErr: "host endpoint unavailable 127.0.0.1:6445", wantCalls: 4, wantListen: 1},
		{name: "occupied gateway releases API", failListen: 2, wantErr: "host endpoint unavailable 127.0.0.1:8443", wantCalls: 4, wantListen: 2, wantCloses: 1},
		{name: "close errors reject admission", closeError: true, wantErr: "close bootstrap port probe: injected preflight close", wantCalls: 4, wantListen: 2, wantCloses: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			selector := "s:sm-kms:pg"
			if test.fullMatrix {
				selector = ""
			}

			input, _ := bootstrapPreflightFixture(t, selector)
			if test.edit != nil {
				test.edit(&input)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			var calls, listens, closes int

			runner := func(gotCtx context.Context, command string, args ...string) ([]byte, error) {
				calls++

				require.Same(t, ctx, gotCtx)
				require.Equal(t, "docker", command)

				expected := [][]string{
					{bootstrapContainerKind, bootstrapListCommand, bootstrapAllContainersFlag, dockerOutputFormatFlag, bootstrapJSONFormat},
					{bootstrapNetworkKind, bootstrapListCommand, dockerOutputFormatFlag, bootstrapJSONFormat},
					{bootstrapVolumeKind, bootstrapListCommand, dockerOutputFormatFlag, bootstrapJSONFormat},
					{dockerInfoCommand, dockerOutputFormatFlag, dockerHostFormat},
				}
				require.LessOrEqual(t, calls, len(expected))
				require.Equal(t, expected[calls-1], args)

				if calls == test.failAt {
					return nil, errors.New("injected daemon error")
				}

				if calls == test.outputAt {
					return []byte(test.output), nil
				}

				if calls == 4 {
					return []byte(testBootstrapHost), nil
				}

				if calls == 1 {
					return []byte("{\"Names\":\"unrelated\",\"Labels\":\"com.docker.compose.project=other\"}\n"), nil
				}

				return []byte("{\"Name\":\"unrelated\",\"Labels\":\"\"}\n"), nil
			}
			listen := func(network, address string) (net.Listener, error) {
				listens++

				require.Equal(t, "tcp", network)
				require.Contains(t, []string{testBootstrapAPIEndpoint, testBootstrapGatewayEndpoint}, address)

				if listens == test.failListen {
					return nil, errors.New("injected port collision")
				}

				return &bootstrapTestListener{close: func() error {
					closes++

					if test.closeError {
						return errors.New("injected preflight close")
					}

					return nil
				}}, nil
			}

			estimate, err := CheckLocalK3sPreflight(ctx, input, runner, listen)

			require.Equal(t, test.wantCalls, calls)
			require.Equal(t, test.wantListen, listens)
			require.Equal(t, test.wantCloses, closes)

			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
				require.Equal(t, CapacityEstimate{}, estimate)

				return
			}

			require.NoError(t, err)
			require.Equal(t, 2, estimate.ApplicationPods)
			require.Equal(t, 5, estimate.InfrastructurePods)
		})
	}
}

func TestBootstrapDockerOwnership(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		at        int
		resource  map[string]string
		raw       string
		wantErr   string
		wantCalls int
	}{
		{name: "empty host", wantCalls: testThree},
		{name: "stopped server collision", at: 1, resource: map[string]string{testBootstrapNamesKey: testBootstrapServerMarker}, wantErr: testBootstrapContainerExists, wantCalls: 1},
		{name: "agent collision", at: 1, resource: map[string]string{testBootstrapNamesKey: testBootstrapAgentMarker}, wantErr: testBootstrapContainerExists, wantCalls: 1},
		{name: "network collision", at: 2, resource: map[string]string{testBootstrapNameKey: bootstrapNetworkKind}, wantErr: "network name already exists", wantCalls: 2},
		{name: "volume collision", at: testThree, resource: map[string]string{testBootstrapNameKey: bootstrapVolumeKind}, wantErr: "volume name already exists", wantCalls: testThree},
		{name: "renamed project container", at: 1, resource: map[string]string{testBootstrapNamesKey: testBootstrapUnrelatedName, testBootstrapLabelsKey: testBootstrapProjectLabel}, wantErr: "project already owns container", wantCalls: 1},
		{name: "renamed project network", at: 2, resource: map[string]string{testBootstrapNameKey: testBootstrapUnrelatedName, testBootstrapLabelsKey: testBootstrapProjectLabel}, wantErr: "project already owns network", wantCalls: 2},
		{name: "renamed project volume", at: testThree, resource: map[string]string{testBootstrapNameKey: testBootstrapUnrelatedName, testBootstrapLabelsKey: testBootstrapProjectLabel}, wantErr: "project already owns volume", wantCalls: testThree},
		{name: "malformed ownership output", at: 1, raw: "{", wantErr: testBootstrapDecodeOwnership, wantCalls: 1},
		{name: "missing name", at: 2, raw: "{}", wantErr: testBootstrapMissingName, wantCalls: 2},
		{name: "null record", at: testThree, raw: testNullDocument, wantErr: testBootstrapMissingName, wantCalls: testThree},
		{name: "trailing malformed record", at: 1, raw: "{\"Names\":\"unrelated\"}\n!", wantErr: testBootstrapDecodeOwnership, wantCalls: 1},
		{name: "project prefix is unrelated", at: 1, resource: map[string]string{testBootstrapNamesKey: testBootstrapUnrelatedName, testBootstrapLabelsKey: testBootstrapProjectPrefix}, wantCalls: testThree},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, plan := bootstrapPreflightFixture(t, "s:sm-kms:pg")

			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			calls := 0
			run := func(context.Context, string, ...string) ([]byte, error) {
				calls++
				if calls != test.at {
					return nil, nil
				}

				if test.raw != "" {
					return []byte(test.raw), nil
				}

				for key, value := range test.resource {
					switch value {
					case testBootstrapServerMarker:
						test.resource[key] = plan.Server.Container
					case testBootstrapAgentMarker:
						test.resource[key] = plan.Agents[0].Container
					case bootstrapNetworkKind:
						test.resource[key] = plan.Network
					case bootstrapVolumeKind:
						test.resource[key] = plan.Server.Volume
					case testBootstrapProjectLabel:
						test.resource[key] = "other=value,com.docker.compose.project=" + plan.Project + ",last=value"
					case testBootstrapProjectPrefix:
						test.resource[key] = "com.docker.compose.project=" + plan.Project + "-other"
					}
				}

				return json.Marshal(test.resource)
			}

			err := checkBootstrapDockerOwnership(ctx, plan, run)

			require.Equal(t, test.wantCalls, calls)

			if test.wantErr != "" {
				require.ErrorContains(t, err, test.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func bootstrapPreflightFixture(t *testing.T, selector string) (BootstrapPreflightInput, LocalK3sPlan) {
	t.Helper()

	model, inventory, lifecycle := lifecycleDeployFixture(t, selector)
	lock, err := DecodeToolchainLock(toolchainLockFixture(t))
	require.NoError(t, err)
	runtimeValues, infrastructureValues := capacityProjections(t, model, RegisteredScopes(), inventory)
	plan, err := BuildLocalK3sPlan(model, lock)
	require.NoError(t, err)

	return BootstrapPreflightInput{
		Model: model, Lock: lock, Inventory: inventory, RuntimeValues: runtimeValues,
		InfrastructureValues: infrastructureValues, Reserve: ResourceCapacity{CPUMilli: 1_000, MemoryBytes: 1 << 30},
		Lifecycle: &lifecycle,
	}, plan
}

func TestCheckLocalK3sPreflightContextAndDependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		nilContext bool
		noDeadline bool
		nilRunner  bool
		nilListen  bool
		cancelAt   int
		wantErr    string
		wantCalls  int
		wantCloses int
	}{
		{name: "missing context", nilContext: true, wantErr: testBootstrapMissingDependency},
		{name: "absent command dependency", nilRunner: true, wantErr: testBootstrapMissingDependency},
		{name: "missing listener", nilListen: true, wantErr: testBootstrapMissingDependency},
		{name: "unbounded context", noDeadline: true, wantErr: "requires a context deadline"},
		{name: "already canceled", cancelAt: -1, wantErr: testBootstrapCanceled},
		{name: "canceled during ownership", cancelAt: 1, wantErr: "ownership probe canceled", wantCalls: 1},
		{name: "canceled after measurement", cancelAt: 4, wantErr: testBootstrapCanceled, wantCalls: 4},
		{name: "canceled during ports", cancelAt: 5, wantErr: testBootstrapCanceled, wantCalls: 4, wantCloses: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			input, _ := bootstrapPreflightFixture(t, "s:sm-kms:pg")

			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			if test.cancelAt == -1 {
				cancel()
			}

			if test.nilContext {
				ctx = nil
			} else if test.noDeadline {
				ctx = context.Background()
			}

			calls, closes := 0, 0

			var runner DockerCommandRunner = func(context.Context, string, ...string) ([]byte, error) {
				calls++
				if calls == test.cancelAt {
					cancel()
				}

				if calls == 4 {
					return []byte(testBootstrapHost), nil
				}

				return nil, nil
			}

			listen := func(string, string) (net.Listener, error) {
				if test.cancelAt == 5 {
					cancel()
				}

				return &bootstrapTestListener{close: func() error {
					closes++

					return nil
				}}, nil
			}

			if test.nilRunner {
				runner = nil
			}

			if test.nilListen {
				listen = nil
			}

			estimate, err := CheckLocalK3sPreflight(ctx, input, runner, listen)
			require.ErrorContains(t, err, test.wantErr)
			require.Equal(t, CapacityEstimate{}, estimate)
			require.Equal(t, test.wantCalls, calls)
			require.Equal(t, test.wantCloses, closes)
		})
	}
}
