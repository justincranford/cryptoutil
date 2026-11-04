// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"time"

	gen "cryptoutil/internal/apps-tools/deploymentgen"
	execution "cryptoutil/internal/apps-tools/deploymentrun/lifecycle"
)

// Docker emits the container's stderr on its own stderr. Capture both streams
// into the same bounded writer (exec serializes equal stdout/stderr writers).
// Unlike the JSON runner, this boundary must not discard diagnostic stderr.
func nativeDevelopmentBootstrapObservation(docker string, cluster gen.LocalK3sPlan,
	runner execution.Runner, snapshot developmentComputeSnapshot,
) string {
	ctx, cancel := context.WithTimeout(context.Background(), developmentDiagnosticTimeout)
	defer cancel()

	return developmentBootstrapObservationWithLogs(ctx, docker, cluster, runner, snapshot,
		func(ctx context.Context, id string) ([]byte, error) {
			return readNativeDevelopmentLogs(ctx, docker, id, func(command *exec.Cmd) error { return command.Run() })
		})
}

const developmentDiagnosticBytes = 1 << 20

// Do not embed bytes.Buffer: its promoted ReadFrom would bypass Write's bound.
type developmentDiagnosticBuffer struct{ buffer bytes.Buffer }

func (b *developmentDiagnosticBuffer) Write(raw []byte) (int, error) {
	if len(raw) > developmentDiagnosticBytes-b.buffer.Len() {
		return 0, errors.New("bootstrap diagnostic memory bound exceeded")
	}

	count, _ := b.buffer.Write(raw) // bytes.Buffer.Write cannot fail.

	return count, nil
}

func readNativeDevelopmentLogs(ctx context.Context, docker, id string, run func(*exec.Cmd) error) ([]byte, error) {
	command := exec.CommandContext(ctx, docker, developmentLogCommand, "--tail", "200", id)
	buffer := &developmentDiagnosticBuffer{}
	command.Stdout, command.Stderr = buffer, buffer

	command.WaitDelay = time.Second
	if err := run(command); err != nil || ctx.Err() != nil {
		return nil, errors.New("bootstrap native diagnostic unavailable")
	}

	return buffer.buffer.Bytes(), nil
}
