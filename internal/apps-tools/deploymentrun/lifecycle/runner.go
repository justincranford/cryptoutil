// Copyright (c) 2025-2026 Justin Cranford.

package lifecycle

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

const maxOutputBytes = 8 << 20

// Runner executes an argument vector, with optional in-memory stdin. It must
// never log input/output or return diagnostics containing private material.
type Runner interface {
	Run(context.Context, string, []string, []byte) ([]byte, error)
}

// ExecRunner reuses exec.CommandContext rather than introducing a Kubernetes
// SDK. Stdout is bounded in memory and stderr is discarded, not logged.
type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, executable string, args []string, input []byte) ([]byte, error) {
	if ctx == nil {
		return nil, errors.New("runner requires context")
	}

	command := exec.CommandContext(ctx, executable, args...)
	command.Stdin = bytes.NewReader(input)
	output := &boundedOutput{}
	command.Stdout = output

	command.WaitDelay = time.Second
	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("lifecycle process: %w", ctx.Err())
		}

		return nil, errors.New("lifecycle process failed")
	}

	return output.Bytes(), nil
}

type boundedOutput struct{ buffer bytes.Buffer }

func (b *boundedOutput) Bytes() []byte { return b.buffer.Bytes() }

func (b *boundedOutput) Len() int { return b.buffer.Len() }

func (b *boundedOutput) Write(p []byte) (int, error) {
	if len(p) > maxOutputBytes-b.Len() {
		return 0, errors.New("lifecycle process output exceeds bound")
	}

	n, _ := b.buffer.Write(p) // bytes.Buffer.Write always returns a nil error.

	return n, nil
}
