// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"context"
	json "encoding/json"
	"strings"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

const (
	developmentStateFormat       = "{{json .State}}"
	developmentLogCommand        = "logs"
	developmentDiagnosticTimeout = 10 * time.Second
)

// Only fixed observations leave this boundary. K3s logs and Docker health
// output may contain credentials; neither is returned, logged, or persisted.
// Observations are diagnostic evidence, not authority to adopt or delete data.
func developmentBootstrapObservation(ctx context.Context, docker string, cluster gen.LocalK3sPlan,
	runner execution.Runner, snapshot developmentComputeSnapshot,
) string {
	return developmentBootstrapObservationWithLogs(ctx, docker, cluster, runner, snapshot,
		func(ctx context.Context, id string) ([]byte, error) {
			return runner.Run(ctx, docker, []string{developmentLogCommand, "--tail", "200", id}, nil)
		})
}

func developmentBootstrapObservationWithLogs(ctx context.Context, docker string, cluster gen.LocalK3sPlan,
	runner execution.Runner, snapshot developmentComputeSnapshot, logs func(context.Context, string) ([]byte, error),
) string {
	id := snapshot.Nodes[cluster.Server.Container]
	if id == "" {
		return "bootstrap observation unavailable"
	}

	actual, err := developmentOwnedNode(ctx, docker, cluster, cluster.Server, runner)
	if err != nil || actual != id {
		return "bootstrap observation identity rejected"
	}

	raw, err := runner.Run(ctx, docker, []string{
		developmentContainer, developmentInspect, developmentFormatFlag, developmentStateFormat, id,
	}, nil)
	if err != nil {
		return "bootstrap state unavailable"
	}

	var state struct {
		Running   bool
		OOMKilled bool
		Health    struct {
			Status string
			Log    []struct{ Output string }
		}
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		return "bootstrap state rejected"
	}

	if state.OOMKilled {
		return "bootstrap observed OOM termination"
	}

	if !state.Running {
		return "bootstrap observed stopped server"
	}

	if state.Health.Status == cryptoutilSharedMagic.DockerServiceHealthHealthy {
		return "bootstrap observed healthy server"
	}

	// The native runner bounds stdout; timeout bounds independent diagnostics
	// when the original launch context has already expired.
	raw, err = logs(ctx, id)
	if err != nil {
		return "bootstrap running; startup log observation unavailable"
	}

	for _, entry := range state.Health.Log {
		raw = append(raw, []byte(entry.Output)...)
	}

	return developmentStartupMarker(raw)
}

func developmentStartupMarker(raw []byte) string {
	lower := strings.ToLower(string(raw))
	for _, marker := range []struct{ text, observation string }{
		{"no space left on device", "bootstrap observed storage exhaustion marker"},
		{"out of memory", "bootstrap observed memory exhaustion marker"},
		{"certificate has expired", "bootstrap observed expired certificate marker"},
		{"certificate is not yet valid", "bootstrap observed certificate validity marker"},
		{"failed to find memory cgroup", "bootstrap observed missing memory cgroup marker"},
		{"failed to find cpu cgroup", "bootstrap observed missing CPU cgroup marker"},
		{"permission denied", "bootstrap observed permission failure marker"},
		{"connection refused", "bootstrap observed refused connection marker"},
	} {
		if strings.Contains(lower, marker.text) {
			return marker.observation
		}
	}

	return "bootstrap running; readiness unresolved"
}

func observeDevelopmentBootstrap(docker string, cluster gen.LocalK3sPlan, runner execution.Runner,
	snapshot developmentComputeSnapshot,
) string {
	ctx, cancel := context.WithTimeout(context.Background(), developmentDiagnosticTimeout)
	defer cancel()

	return developmentBootstrapObservation(ctx, docker, cluster, runner, snapshot)
}
