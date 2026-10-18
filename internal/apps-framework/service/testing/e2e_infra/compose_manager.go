// Copyright (c) 2025-2026 Justin Cranford.
//
//

// Package e2e_infra provides reusable helpers for E2E testing with docker compose.
// Deprecated: use cryptoutil/internal/apps-framework/service/test_orch_e2e instead.
package e2e_infra

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	http "net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	cryptoutilSharedCryptoTls "cryptoutil/internal/shared/crypto/tls"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// ComposeManager orchestrates docker compose lifecycle for E2E tests.
type ComposeManager struct {
	ComposeFile        string
	Profiles           []string
	HTTPClient         *http.Client
	healthPollInterval time.Duration
	// psOutputFn provides docker compose ps JSON output; injectable for testing.
	psOutputFn         func(ctx context.Context) ([]byte, error)
	runCommandFn       func(ctx context.Context, command string, args ...string) error
	runCommandOutputFn func(ctx context.Context, command string, args ...string) ([]byte, error)
	removeAllFn        func(path string) error
	mkdirAllFn         func(path string, perm fs.FileMode) error
	statFn             func(path string) (os.FileInfo, error)
	makeWritableFn     func(path string) error
}

// NewComposeManager creates a compose manager with TLS-enabled HTTP client.
// Profiles specifies Docker Compose profiles to activate (e.g., "dev", "postgres").
// Services with profiles: [...] in compose.yml only start when their profile is active.
func NewComposeManager(composeFile string, profiles ...string) *ComposeManager {
	cm := &ComposeManager{
		ComposeFile:        composeFile,
		Profiles:           profiles,
		HTTPClient:         cryptoutilSharedCryptoTls.NewClientForTest(),
		healthPollInterval: cryptoutilSharedMagic.DefaultE2EHealthPollInterval,
	}
	cm.psOutputFn = cm.defaultPsOutput

	return cm
}

// defaultPsOutput runs the real docker compose ps command.
func (cm *ComposeManager) defaultPsOutput(ctx context.Context) ([]byte, error) {
	output, err := cm.runCommandOutput(
		ctx, "docker", valueCompose, valueF, cm.ComposeFile, "ps", "-a", "--format", "json",
	)
	if err != nil {
		return output, fmt.Errorf("docker compose ps failed: %w", err)
	}

	return output, nil
}

// Start brings up docker compose stack.
//
// Note: Does NOT use --wait flag because Docker Compose --wait only works with containers that have
// native HEALTHCHECK instructions in their container image or Dockerfile. Many containers (like
// otel/opentelemetry-collector-contrib) don't have native healthchecks.
//
// Instead, this project uses three healthcheck strategies (implemented in docker_health.go):
//  1. Job-only healthchecks: Standalone jobs that must exit successfully (ExitCode=0)
//     Examples: healthcheck-secrets, builder-sm-kms
//  2. Service-only healthchecks: Services with native HEALTHCHECK instructions
//     Examples: cryptoutil-sqlite, cryptoutil-postgres-1, postgres, grafana-otel-lgtm
//  3. Service with healthcheck job: Services use external sidecar job for health verification
//     Example: opentelemetry-collector-contrib with healthcheck-opentelemetry-collector-contrib
//
// Use WaitForMultipleServices() or WaitForServicesHealthy() after Start() to wait for services.
func (cm *ComposeManager) Start(ctx context.Context) error {
	fmt.Println("Starting docker compose stack...")

	if err := cm.resetCertOutputDir(); err != nil {
		return fmt.Errorf("failed to prepare cert output dir: %w", err)
	}

	args := cm.buildComposeArgs("up", "-d", "--build")
	if err := cm.runCommand(ctx, "docker", args...); err != nil {
		return fmt.Errorf("failed to start docker compose: %w", err)
	}

	return nil
}

// SyncCertOutputDirFromPkiInit copies the runtime-generated cert tree from the
// pki-init container into the host deployment cert directory used by e2e tests.
func (cm *ComposeManager) SyncCertOutputDirFromPkiInit(ctx context.Context, hostCACertPath string) error {
	composeDir := filepath.Dir(cm.ComposeFile)
	hostCertsDir := filepath.Clean(filepath.Join(hostCACertPath, "..", "..", "..", ".."))
	containerName := fmt.Sprintf("%s-pki-init-1", filepath.Base(composeDir))

	if err := cm.removeAll(hostCertsDir); err != nil {
		return fmt.Errorf("remove host certs dir %s: %w", hostCertsDir, err)
	}

	if err := cm.mkdirAll(hostCertsDir, cryptoutilSharedMagic.PKIInitCertsDirMode); err != nil {
		return fmt.Errorf("mkdir host certs dir %s: %w", hostCertsDir, err)
	}

	if err := cm.runCommand(ctx, "docker", "cp", containerName+":/certs/.", hostCertsDir); err != nil {
		return fmt.Errorf("copy certs from %s to %s: %w", containerName, hostCertsDir, err)
	}

	return nil
}

func (cm *ComposeManager) runCommand(ctx context.Context, command string, args ...string) error {
	if cm.runCommandFn != nil {
		return cm.runCommandFn(ctx, command, args...)
	}

	return runCommandWithFn(ctx, command, args, (*exec.Cmd).Run)
}

func runCommandWithFn(
	ctx context.Context,
	command string,
	args []string,
	runFn func(*exec.Cmd) error,
) error {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := runFn(cmd); err != nil {
		return fmt.Errorf("run command %q: %w", command, err)
	}

	return nil
}

func (cm *ComposeManager) runCommandOutput(ctx context.Context, command string, args ...string) ([]byte, error) {
	if cm.runCommandOutputFn != nil {
		return cm.runCommandOutputFn(ctx, command, args...)
	}

	return runCommandOutputWithFn(ctx, command, args, (*exec.Cmd).CombinedOutput)
}

func runCommandOutputWithFn(
	ctx context.Context,
	command string,
	args []string,
	runFn func(*exec.Cmd) ([]byte, error),
) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	output, err := runFn(cmd)
	if err != nil {
		return output, fmt.Errorf("run command %q and capture output: %w", command, err)
	}

	return output, nil
}

func (cm *ComposeManager) removeAll(path string) error {
	if cm.removeAllFn != nil {
		return cm.removeAllFn(path)
	}

	return removeAllWithFn(path, os.RemoveAll)
}

func removeAllWithFn(path string, removeAllFn func(string) error) error {
	if err := removeAllFn(path); err != nil {
		return fmt.Errorf("remove path %q: %w", path, err)
	}

	return nil
}

func (cm *ComposeManager) mkdirAll(path string, perm fs.FileMode) error {
	if cm.mkdirAllFn != nil {
		return cm.mkdirAllFn(path, perm)
	}

	return mkdirAllWithFn(path, perm, os.MkdirAll)
}

func mkdirAllWithFn(path string, perm fs.FileMode, mkdirAllFn func(string, fs.FileMode) error) error {
	if err := mkdirAllFn(path, perm); err != nil {
		return fmt.Errorf("create directory %q: %w", path, err)
	}

	return nil
}

func (cm *ComposeManager) stat(path string) (os.FileInfo, error) {
	if cm.statFn != nil {
		return cm.statFn(path)
	}

	return statWithFn(path, os.Stat)
}

func statWithFn(path string, statFn func(string) (os.FileInfo, error)) (os.FileInfo, error) {
	info, err := statFn(path)
	if err != nil {
		return nil, fmt.Errorf("stat path %q: %w", path, err)
	}

	return info, nil
}

func (cm *ComposeManager) makeWritable(path string) error {
	if cm.makeWritableFn != nil {
		return cm.makeWritableFn(path)
	}

	return makeWritableRecursive(path)
}

// resetCertOutputDir makes the compose cert output directory writable.
// On Windows, read-only attributes on files (for example issuing-ca.pem)
// can prevent pki-init from regenerating certs.
func (cm *ComposeManager) resetCertOutputDir() error {
	composeDir := filepath.Dir(cm.ComposeFile)
	certsDir := filepath.Join(composeDir, "certs")

	if _, err := cm.stat(certsDir); errors.Is(err, fs.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("stat certs dir %s: %w", certsDir, err)
	}

	if err := cm.makeWritable(certsDir); err != nil {
		return fmt.Errorf("make writable certs dir %s: %w", certsDir, err)
	}

	return nil
}

// makeWritableRecursive normalizes directory permissions and clears read-only
// file permissions on Windows so pki-init can rewrite cert output.
func makeWritableRecursive(path string) error {
	return makeWritableRecursiveWithOps(
		path,
		func(rootPath string) (writableRoot, error) {
			return os.OpenRoot(rootPath)
		},
		filepath.WalkDir,
		filepath.Rel,
		runtime.GOOS,
	)
}

type writableRoot interface {
	Chmod(name string, mode fs.FileMode) error
	Close() error
}

func makeWritableRecursiveWithOps(
	path string,
	openRootFn func(path string) (writableRoot, error),
	walkDirFn func(root string, fn fs.WalkDirFunc) error,
	relFn func(basepath, targpath string) (string, error),
	goos string,
) error {
	root, err := openRootFn(path)
	if err != nil {
		return fmt.Errorf("open root %s: %w", path, err)
	}

	defer func() {
		_ = root.Close()
	}()

	if err := walkDirFn(path, func(walkPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk %s: %w", walkPath, walkErr)
		}

		relPath, relErr := relFn(path, walkPath)
		if relErr != nil {
			return fmt.Errorf("rel path %s: %w", walkPath, relErr)
		}

		if d.IsDir() {
			info, infoErr := d.Info()
			if infoErr != nil {
				return fmt.Errorf("stat dir %s: %w", walkPath, infoErr)
			}

			if info.Mode().Perm() == cryptoutilSharedMagic.CICDTempDirPermissions {
				return nil
			}

			if err := root.Chmod(relPath, cryptoutilSharedMagic.CICDTempDirPermissions); err != nil {
				return fmt.Errorf("chmod dir %s: %w", walkPath, err)
			}

			return nil
		}

		if goos != "windows" {
			return nil
		}

		info, infoErr := d.Info()
		if infoErr != nil {
			return fmt.Errorf("stat file %s: %w", walkPath, infoErr)
		}

		if info.Mode().Perm()&cryptoutilSharedMagic.CICDFileOwnerWritePermission != 0 {
			return nil
		}

		if err := root.Chmod(relPath, cryptoutilSharedMagic.CacheFilePermissions); err != nil {
			return fmt.Errorf("chmod file %s: %w", walkPath, err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("walk and normalize writable perms for %s: %w", path, err)
	}

	return nil
}

// Stop tears down docker compose stack.
func (cm *ComposeManager) Stop(ctx context.Context) error {
	fmt.Println("Stopping docker compose stack...")

	args := cm.buildComposeArgs("down", "-v")
	if err := cm.runCommand(ctx, "docker", args...); err != nil {
		return fmt.Errorf("failed to stop docker compose: %w", err)
	}

	return nil
}

// buildComposeArgs builds the docker compose command arguments including profiles.
func (cm *ComposeManager) buildComposeArgs(subcommand ...string) []string {
	args := make([]string, 0, 3+2*len(cm.Profiles)+len(subcommand))
	args = append(args, valueCompose, valueF, cm.ComposeFile)

	for _, profile := range cm.Profiles {
		args = append(args, "--profile", profile)
	}

	args = append(args, subcommand...)

	return args
}

// BuildDockerExecArgs builds the docker exec arguments to run a command inside a container.
// Returns arguments suitable for passing to exec.Command("docker", args...).
func (cm *ComposeManager) BuildDockerExecArgs(service string, command ...string) []string {
	execArgs := cm.buildComposeArgs(append([]string{valueExec, service}, command...)...)

	return execArgs
}

// WaitForHealth polls an health endpoint until healthy or timeout.
// Supports both admin endpoints (/admin/api/v1/livez) and public endpoints (/health).
func (cm *ComposeManager) WaitForHealth(healthURL string, timeout time.Duration) error {
	ctx := context.Background()

	timeoutCh := time.After(timeout)

	pollInterval := cm.healthPollInterval
	if pollInterval == 0 {
		pollInterval = cryptoutilSharedMagic.DefaultE2EHealthPollInterval
	} else if pollInterval < 0 {
		return fmt.Errorf("health poll interval must be positive: %s", pollInterval)
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	fmt.Printf("[WaitForHealth] Starting health check for %s (timeout: %v)\n", healthURL, timeout)

	attempts := 0

	for {
		select {
		case <-timeoutCh:
			fmt.Printf("[WaitForHealth] TIMEOUT for %s after %d attempts\n", healthURL, attempts)

			return fmt.Errorf("health check timeout after %v (%d attempts)", timeout, attempts)
		default:
		}

		attempts++

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
		if err != nil {
			fmt.Printf("[WaitForHealth] Attempt %d for %s: request creation error: %v\n", attempts, healthURL, err)
		} else {
			resp, err := cm.HTTPClient.Do(req)
			if err != nil {
				fmt.Printf("[WaitForHealth] Attempt %d for %s: connection error: %v\n", attempts, healthURL, err)
			} else {
				_ = resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					fmt.Printf("[WaitForHealth] SUCCESS for %s after %d attempts\n", healthURL, attempts)

					return nil
				}

				fmt.Printf("[WaitForHealth] Attempt %d for %s: HTTP %d\n", attempts, healthURL, resp.StatusCode)
			}
		}

		select {
		case <-timeoutCh:
			fmt.Printf("[WaitForHealth] TIMEOUT for %s after %d attempts\n", healthURL, attempts)

			return fmt.Errorf("health check timeout after %v (%d attempts)", timeout, attempts)
		case <-ticker.C:
		}
	}
}

// WaitForMultipleServices waits for multiple services to be healthy concurrently.
// Returns error if any service fails health check within timeout.
func (cm *ComposeManager) WaitForMultipleServices(services map[string]string, timeout time.Duration) error {
	type result struct {
		name string
		err  error
	}

	resultsCh := make(chan result, len(services))

	// Start health checks for all services concurrently.
	for name, healthURL := range services {
		go func(serviceName, url string) {
			err := cm.WaitForHealth(url, timeout)
			resultsCh <- result{name: serviceName, err: err}
		}(name, healthURL)
	}

	// Collect results.
	for i := 0; i < len(services); i++ {
		res := <-resultsCh
		if res.err != nil {
			return fmt.Errorf("service %s health check failed: %w", res.name, res.err)
		}
	}

	return nil
}
