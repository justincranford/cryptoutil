// Copyright (c) 2025 Justin Cranford
//
// Shared test helpers for cipher-im e2e tests.

package e2e_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	// dockerComposeTimeout is the timeout for docker compose commands.
	dockerComposeTimeout = 5 * time.Minute

	// testHTTPClientTimeout is the timeout for HTTP client requests.
	testHTTPClientTimeout = 10 * time.Second
)

// createHTTPSClient creates HTTP client with TLS verification disabled for testing.
func createHTTPSClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // Test environment only
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   testHTTPClientTimeout,
	}
}

// runDockerCompose executes docker compose command from cmd/cipher-im directory.
func runDockerCompose(args ...string) error {
	composeDir := filepath.Join("..", "..", "..", "..", "..", "..", "cmd", "cipher-im")

	ctx, cancel := context.WithTimeout(context.Background(), dockerComposeTimeout)
	defer cancel()

	cmdArgs := append([]string{"compose"}, args...)
	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)
	cmd.Dir = composeDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose %v failed: %w", args, err)
	}

	return nil
}
