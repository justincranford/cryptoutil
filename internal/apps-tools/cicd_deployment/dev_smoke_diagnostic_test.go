// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	json "encoding/json"
	"errors"
	"strings"
	"testing"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

type developmentDiagnosticRunner struct {
	owner *developmentCleanupRunner
	state string
	logs  string
	fail  string
	calls int
}

const (
	testDiagnosticState   = cryptoutilSharedMagic.ParamState
	testDiagnosticPrivate = "PRIVATE"
	testDiagnosticRunning = `{"Running":true}`
)

func (r *developmentDiagnosticRunner) Run(ctx context.Context, binary string, args []string, stdin []byte) ([]byte, error) {
	if args[0] == developmentLogCommand {
		r.calls++
		if r.fail == developmentLogCommand {
			return nil, errors.New(testDiagnosticPrivate)
		}

		return []byte(r.logs), nil
	}

	if args[len(args)-2] == developmentStateFormat {
		if r.fail == testDiagnosticState {
			return nil, errors.New(testDiagnosticPrivate)
		}

		return []byte(r.state), nil
	}

	return r.owner.Run(ctx, binary, args, stdin)
}

func TestDevelopmentBootstrapObservationRedacted(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, state, logs, failure, want string
		missing                          bool
	}{
		{name: "missing captured identity", missing: true, want: "bootstrap observation unavailable"},
		{name: "foreign captured owner", failure: developmentInspect, want: "bootstrap observation identity rejected"},
		{name: "state failure", failure: testDiagnosticState, want: "bootstrap state unavailable"},
		{name: "malformed state response", state: testDiagnosticPrivate, want: "bootstrap state rejected"},
		{name: "oom", state: `{"Running":true,"OOMKilled":true}`, want: "bootstrap observed OOM termination"},
		{name: "stopped", state: `{}`, want: "bootstrap observed stopped server"},
		{name: cryptoutilSharedMagic.DockerServiceHealthHealthy, state: `{"Running":true,"Health":{"Status":"healthy"}}`, want: "bootstrap observed healthy server"},
		{name: "log failure", state: testDiagnosticRunning, failure: developmentLogCommand, want: "bootstrap running; startup log observation unavailable"},
		{name: "unclassified startup", state: testDiagnosticRunning, logs: "PRIVATE token K10-secret", want: "bootstrap running; readiness unresolved"},
		{name: "health marker", state: `{"Running":true,"Health":{"Log":[{"Output":"PRIVATE permission denied K10-secret"}]}}`, want: "bootstrap observed permission failure marker"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cluster := gen.LocalK3sPlan{
				Run: gen.DevelopmentSmokeRun, Project: testDevCleanupProject, Image: testDevCleanupImage,
				Server: gen.LocalK3sNode{Container: "diagnostic-server", Volume: "diagnostic-volume"},
			}

			owner := &developmentCleanupRunner{t: t, cluster: cluster}
			if tc.failure == developmentInspect {
				owner.failure = tc.failure
			}

			runner := &developmentDiagnosticRunner{owner: owner, state: tc.state, logs: tc.logs, fail: tc.failure}

			snapshot := developmentComputeSnapshot{Nodes: map[string]string{cluster.Server.Container: "owned-id-" + cluster.Server.Container}}
			if tc.missing {
				snapshot.Nodes = nil
			}

			got := observeDevelopmentBootstrap("docker", cluster, runner, snapshot)
			require.Equal(t, tc.want, got)
			require.NotContains(t, got, testDiagnosticPrivate)
			require.NotContains(t, got, "K10-secret")
			require.Zero(t, owner.deletes)
		})
	}
}

func TestDevelopmentStartupMarkers(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ marker, want string }{
		{"no space left on device", "storage exhaustion"},
		{"out of memory", "memory exhaustion"},
		{"certificate has expired", "expired certificate"},
		{"certificate is not yet valid", "certificate validity"},
		{"failed to find memory cgroup", "missing memory cgroup"},
		{"failed to find cpu cgroup", "missing CPU cgroup"},
		{"permission denied", "permission failure"},
		{"connection refused", "refused connection"},
	} {
		t.Run(tc.marker, func(t *testing.T) {
			t.Parallel()

			raw, err := json.Marshal(map[string]string{cryptoutilSharedMagic.StringMessage: strings.ToUpper(tc.marker), cryptoutilSharedMagic.ParamToken: testDiagnosticPrivate})
			require.NoError(t, err)
			require.Equal(t, "bootstrap observed "+tc.want+" marker", developmentStartupMarker(raw))
		})
	}
}
