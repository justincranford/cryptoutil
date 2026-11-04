// Copyright (c) 2025-2026 Justin Cranford.

package deploymentgen

import (
	"bytes"
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	json "encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testDevelopmentRun               = "p2sm1103"
	testDevelopmentFreshSmoke        = "fresh development smoke"
	testDevelopmentHelp              = cryptoutilSharedMagic.CLIHelpFlag
	testDevelopmentCollectorMismatch = "incompatible collector projection"
	testDevelopmentInvalidRun        = "invalid owned run"
	testDevelopmentBadRun            = "development-bad-run"
)

func TestDevelopmentRuntimeCLIReadOnlyGuards(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name                                          string
		flags                                         []string
		failRead, invalidRead                         string
		failImage, badImage, collision, ports, output bool
		want                                          int
	}{
		{name: testDevelopmentFreshSmoke},
		{name: "invalid flags", flags: []string{cliUnknownFlag}, want: 2},
		{name: "development planning help", flags: []string{testDevelopmentHelp}},
		{name: "missing image", flags: []string{"-image="}, want: 2},
		{name: "missing development model", failRead: defaultPlannerPaths().model, want: 1},
		{name: "invalid development model", invalidRead: defaultPlannerPaths().model, want: 1},
		{name: testDevelopmentCollectorMismatch, want: 1},
		{name: "missing lock", failRead: filepath.Join(filepath.Dir(defaultPlannerPaths().model), k8sToolchainLockFileName), want: 1},
		{name: "invalid development lock", invalidRead: filepath.Join(filepath.Dir(defaultPlannerPaths().model), k8sToolchainLockFileName), want: 1},
		{name: "image transport rejected", failImage: true, want: 1},
		{name: "image provenance rejected", badImage: true, want: 1},
		{name: testDevelopmentInvalidRun, flags: []string{"-run=" + testDevelopmentBadRun}, want: 1},
		{name: "missing infrastructure", failRead: infrastructureChartDefaultsPath, want: 1},
		{name: "invalid infrastructure", invalidRead: infrastructureChartDefaultsPath, want: 1},
		{name: "owned resource collision", collision: true, want: 1},
		{name: "occupied host port", ports: true, want: 1},
		{name: "full matrix insufficient host", flags: []string{"-select="}, want: 1},
		{name: "failed output", output: true, want: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			model, _ := fixture(t)
			modelData, err := json.Marshal(model)
			require.NoError(t, err)
			defaults, err := os.ReadFile(filepath.Join("..", "..", "..", infrastructureChartDefaultsPath))
			require.NoError(t, err)

			read := func(name string) ([]byte, error) {
				if name == test.failRead {
					return nil, errors.New("PRIVATE")
				}

				if name == test.invalidRead {
					return []byte("{"), nil
				}

				if name == defaultPlannerPaths().model {
					if test.name == testDevelopmentCollectorMismatch {
						model.Infrastructure.Collector = 3

						return json.Marshal(model)
					}

					return modelData, nil
				}

				if name == infrastructureChartDefaultsPath {
					return defaults, nil
				}

				return toolchainLockFixture(t), nil
			}
			runner := func(ctx context.Context, command string, args ...string) ([]byte, error) {
				require.Equal(t, "docker", command)
				require.NotContains(t, args, "create")
				require.NotContains(t, args, cryptoutilSharedMagic.WorkflowSubCmdRun)
				require.NotContains(t, args, "rm")
				require.NotContains(t, args, "up")

				_, bounded := ctx.Deadline()
				require.True(t, bounded)

				if args[0] == "image" {
					if test.failImage {
						return nil, errors.New("PRIVATE")
					}

					if test.badImage {
						return []byte("{}"), nil
					}

					labelRun := testDevelopmentRun
					if test.name == testDevelopmentInvalidRun {
						labelRun = testDevelopmentBadRun
					}

					return json.Marshal(map[string]any{"digests": []string{"test/suite@sha256:" + strings.Repeat("a", 64)}, "os": cryptoutilSharedMagic.MockRuntimeGoOS, "architecture": cryptoutilSharedMagic.MockRuntimeGoArch, cryptoutilSharedMagic.WorkflowSubCmdRun: labelRun, "owner": "phase2-runtime"})
				}

				if args[0] == "info" {
					return []byte("14 8326418432"), nil
				}

				if test.collision && args[0] == "container" {
					return []byte(`{"Names":"cu-p2sm1103-cluster-server-0"}`), nil
				}

				return nil, nil
			}
			listen := func(string, string) (net.Listener, error) {
				if test.ports {
					return nil, errors.New("port occupied")
				}

				return &bootstrapTestListener{close: func() error { return nil }}, nil
			}

			var (
				stdout, stderr bytes.Buffer
				writer         io.Writer = &stdout
			)
			if test.output {
				writer = bootstrapCLIErrorWriter{}
			}

			args := append([]string{cliRunFlag, testDevelopmentRun, "-image", "test/suite:local", "-reserve-cpu-milli", "2000", "-reserve-memory-bytes", "2147483648"}, test.flags...)
			code := developmentRuntimePlanMain(args, writer, &stderr, read, runner, listen)
			require.Equal(t, test.want, code, stderr.String())
			require.NotContains(t, stdout.String()+stderr.String(), "PRIVATE")

			if test.name == testDevelopmentFreshSmoke {
				require.Contains(t, stdout.String(), "NOT runtime/material certification")
			}
		})
	}
}
