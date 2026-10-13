// Copyright (c) 2025-2026 Justin Cranford.
//
//

package cicd_workflow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type workflowDependencies struct {
	commandContextFn func(context.Context, string, ...string) *exec.Cmd
	setupCmdPipesFn  func(*exec.Cmd) (io.ReadCloser, io.ReadCloser, error)
	makeDirectoryFn  func(string, os.FileMode) error
	openFileFn       func(string, int, os.FileMode) (*os.File, error)
	closeFileFn      func(*os.File) error
	readFileFn       func(string) ([]byte, error)
}

func newWorkflowDependencies() workflowDependencies {
	return workflowDependencies{
		commandContextFn: exec.CommandContext,
		setupCmdPipesFn:  setupCommandPipes,
		makeDirectoryFn:  os.MkdirAll,
		openFileFn:       os.OpenFile,
		closeFileFn:      func(file *os.File) error { return file.Close() },
		readFileFn:       os.ReadFile,
	}
}

func setupCommandPipes(cmd *exec.Cmd) (io.ReadCloser, io.ReadCloser, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		closeErr := stdout.Close()
		if stdoutWriter, ok := cmd.Stdout.(*os.File); ok {
			closeErr = errors.Join(closeErr, stdoutWriter.Close())
		}

		if closeErr != nil {
			return nil, nil, errors.Join(fmt.Errorf("create stderr pipe: %w", err), fmt.Errorf("close stdout pipe: %w", closeErr))
		}

		return nil, nil, fmt.Errorf("create stderr pipe: %w", err)
	}

	return stdout, stderr, nil
}
