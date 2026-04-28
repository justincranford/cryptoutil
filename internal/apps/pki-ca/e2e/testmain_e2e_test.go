//go:build e2e

// Copyright (c) 2025 Justin Cranford

package e2e_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

func TestMain(m *testing.M) {
	rootDir, err := projectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to locate project root: %v\n", err)
		os.Exit(1)
	}

	composeFile := filepath.Join(rootDir, "deployments", cryptoutilSharedMagic.OTLPServicePKICA, "compose.yml")
	_ = startCompose(composeFile)
	code := m.Run()
	_ = stopCompose(composeFile)

	os.Exit(code)
}

func projectRoot() (string, error) {
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("runtime.Caller failed")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filePath), "..", "..", "..", "..")), nil
}

func startCompose(composeFile string) error {
	cmd := exec.Command("docker", "compose", "-f", composeFile, "up", "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func stopCompose(composeFile string) error {
	cmd := exec.Command("docker", "compose", "-f", composeFile, "down", "-v")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
