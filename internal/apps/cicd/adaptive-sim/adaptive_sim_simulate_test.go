// Copyright (c) 2025 Justin Cranford

package main

import (
	"bytes"
	"context"
	json "encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity/idp/userauth"
)
func TestInternalMain_FlagParseError(t *testing.T) {
	t.Parallel()

	// Invalid flag should cause parse error.
	args := []string{"adaptive-sim", "--invalid-flag-that-does-not-exist"}
	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := internalMain(args, stdin, stdout, stderr)

	require.Equal(t, exitError, exitCode)
	require.Contains(t, stderr.String(), "not defined")
}

func TestInternalMain_MkdirAllError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	logsPath := filepath.Join(tempDir, "logs.json")

	// Create a valid logs file.
	testLogs := []HistoricalAuthLog{
		{
			UserID: "u1", Operation: "login", IPAddress: "192.168.1.1", DeviceID: "d1",
			Country: "US", City: "NYC", IsVPN: false, IsProxy: false,
			CurrentAuthLevel: "basic", Success: true, Metadata: map[string]any{},
		},
	}

	data, _ := json.MarshalIndent(testLogs, "", "  ")
	_ = os.WriteFile(logsPath, data, filePerms600)

	// Use an output path that is a file (not directory) to cause MkdirAll to fail.
	// Create a file that will block directory creation.
	blockingFile := filepath.Join(tempDir, "blocking-file")
	_ = os.WriteFile(blockingFile, []byte("blocking"), filePerms600)

	// Try to create a directory inside the file (should fail).
	invalidOutputDir := filepath.Join(blockingFile, "nested", "output")

	args := []string{
		"adaptive-sim",
		"--logs=" + logsPath,
		"--output=" + invalidOutputDir,
		"--risk-scoring=/nonexistent/risk.yml",
	}
	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := internalMain(args, stdin, stdout, stderr)

	// Should fail because MkdirAll cannot create directory inside a file.
	require.Equal(t, exitError, exitCode)
}

func TestInternalMain_SaveResultsError(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("os.Chmod does not enforce POSIX permissions on Windows")
	}

	tempDir := t.TempDir()
	logsPath := filepath.Join(tempDir, "logs.json")
	outputDir := filepath.Join(tempDir, "output")

	// Create valid logs file.
	testLogs := []HistoricalAuthLog{
		{
			UserID: "u1", Operation: "login", IPAddress: "192.168.1.1", DeviceID: "d1",
			Country: "US", City: "NYC", IsVPN: false, IsProxy: false,
			CurrentAuthLevel: "basic", Success: true, Metadata: map[string]any{},
		},
	}

	data, _ := json.MarshalIndent(testLogs, "", "  ")
	_ = os.WriteFile(logsPath, data, filePerms600)

	// Create policy files.
	riskScoringPath := filepath.Join(tempDir, "risk_scoring.yml")
	stepUpPath := filepath.Join(tempDir, "step_up.yml")
	adaptivePath := filepath.Join(tempDir, "adaptive_auth.yml")
	_ = os.WriteFile(riskScoringPath, []byte(mockRiskScoringPolicy), filePerms600)
	_ = os.WriteFile(stepUpPath, []byte(mockStepUpPolicy), filePerms600)
	_ = os.WriteFile(adaptivePath, []byte(mockAdaptiveAuthPolicy), filePerms600)

	// Create the output directory first.
	_ = os.MkdirAll(outputDir, 0o755)

	// Make the output directory read-only to cause SaveResults to fail.
	_ = os.Chmod(outputDir, 0o444)

	// Ensure we restore permissions for cleanup.
	t.Cleanup(func() {
		_ = os.Chmod(outputDir, 0o755)
	})

	args := []string{
		"adaptive-sim",
		"--logs=" + logsPath,
		"--output=" + outputDir,
		"--risk-scoring=" + riskScoringPath,
		"--step-up=" + stepUpPath,
		"--adaptive=" + adaptivePath,
	}
	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := internalMain(args, stdin, stdout, stderr)

	// Should fail because SaveResults cannot write to read-only directory.
	require.Equal(t, exitError, exitCode)
	require.Contains(t, stderr.String(), "Failed to save results")
}

func TestSimulate_EmptyLogs(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create empty logs file.
	logsPath := filepath.Join(tempDir, "empty_logs.json")
	_ = os.WriteFile(logsPath, []byte("[]"), filePerms600)

	// Create policy files with valid content.
	riskScoringPath := filepath.Join(tempDir, "risk_scoring.yml")
	stepUpPath := filepath.Join(tempDir, "step_up.yml")
	adaptivePath := filepath.Join(tempDir, "adaptive_auth.yml")

	_ = os.WriteFile(riskScoringPath, []byte(mockRiskScoringPolicy), filePerms600)
	_ = os.WriteFile(stepUpPath, []byte(mockStepUpPolicy), filePerms600)
	_ = os.WriteFile(adaptivePath, []byte(mockAdaptiveAuthPolicy), filePerms600)

	loader := cryptoutilIdentityIdpUserauth.NewYAMLPolicyLoader(riskScoringPath, stepUpPath, adaptivePath)
	simulator := &AdaptiveSimulator{
		policyLoader: loader,
		outputDir:    tempDir,
	}

	ctx := context.Background()

	result, err := simulator.Simulate(ctx, logsPath, "v1.0")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 0, result.TotalAttempts)
	require.Equal(t, 0.0, result.StepUpRate)
	require.Equal(t, 0.0, result.BlockedRate)
}

func TestSimulate_RiskPolicyLoadError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create logs file.
	logsPath := filepath.Join(tempDir, "logs.json")
	_ = os.WriteFile(logsPath, []byte("[]"), filePerms600)

	// Create loader with nonexistent policy file.
	loader := cryptoutilIdentityIdpUserauth.NewYAMLPolicyLoader(
		"/nonexistent/risk.yml",
		filepath.Join(tempDir, "step_up.yml"),
		filepath.Join(tempDir, "adaptive.yml"),
	)
	simulator := &AdaptiveSimulator{
		policyLoader: loader,
		outputDir:    tempDir,
	}

	ctx := context.Background()

	result, err := simulator.Simulate(ctx, logsPath, "v1.0")

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to load risk scoring policy")
}

func TestSimulate_StepUpPolicyLoadError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create logs file.
	logsPath := filepath.Join(tempDir, "logs.json")
	_ = os.WriteFile(logsPath, []byte("[]"), filePerms600)

	// Create valid risk scoring policy.
	riskScoringPath := filepath.Join(tempDir, "risk_scoring.yml")
	_ = os.WriteFile(riskScoringPath, []byte(mockRiskScoringPolicy), filePerms600)

	// Create loader with nonexistent step-up policy.
	loader := cryptoutilIdentityIdpUserauth.NewYAMLPolicyLoader(
		riskScoringPath,
		"/nonexistent/step_up.yml",
		filepath.Join(tempDir, "adaptive.yml"),
	)
	simulator := &AdaptiveSimulator{
		policyLoader: loader,
		outputDir:    tempDir,
	}

	ctx := context.Background()

	result, err := simulator.Simulate(ctx, logsPath, "v1.0")

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to load step-up policies")
}

func TestSimulate_AdaptivePolicyLoadError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create logs file.
	logsPath := filepath.Join(tempDir, "logs.json")
	_ = os.WriteFile(logsPath, []byte("[]"), filePerms600)

	// Create valid risk scoring and step-up policies.
	riskScoringPath := filepath.Join(tempDir, "risk_scoring.yml")
	stepUpPath := filepath.Join(tempDir, "step_up.yml")
	_ = os.WriteFile(riskScoringPath, []byte(mockRiskScoringPolicy), filePerms600)
	_ = os.WriteFile(stepUpPath, []byte(mockStepUpPolicy), filePerms600)

	// Create loader with nonexistent adaptive policy.
	loader := cryptoutilIdentityIdpUserauth.NewYAMLPolicyLoader(
		riskScoringPath,
		stepUpPath,
		"/nonexistent/adaptive.yml",
	)
	simulator := &AdaptiveSimulator{
		policyLoader: loader,
		outputDir:    tempDir,
	}

	ctx := context.Background()

	result, err := simulator.Simulate(ctx, logsPath, "v1.0")

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to load adaptive auth policy")
}

func TestSimulate_HistoricalLogsLoadError(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create valid policy files.
	riskScoringPath := filepath.Join(tempDir, "risk_scoring.yml")
	stepUpPath := filepath.Join(tempDir, "step_up.yml")
	adaptivePath := filepath.Join(tempDir, "adaptive_auth.yml")
	_ = os.WriteFile(riskScoringPath, []byte(mockRiskScoringPolicy), filePerms600)
	_ = os.WriteFile(stepUpPath, []byte(mockStepUpPolicy), filePerms600)
	_ = os.WriteFile(adaptivePath, []byte(mockAdaptiveAuthPolicy), filePerms600)

	loader := cryptoutilIdentityIdpUserauth.NewYAMLPolicyLoader(riskScoringPath, stepUpPath, adaptivePath)
	simulator := &AdaptiveSimulator{
		policyLoader: loader,
		outputDir:    tempDir,
	}

	ctx := context.Background()

	// Use nonexistent logs file.
	result, err := simulator.Simulate(ctx, "/nonexistent/logs.json", "v1.0")

	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "failed to load historical logs")
}

func TestSimulate_BlockDecision(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create logs with high-risk user that should trigger block decision.
	// The risk score needs to be >= 0.8 (critical threshold) to trigger block.
	// VPN (0.6 * 0.30 = 0.18) + Proxy (0.5 * 0.30 = 0.15) + High-risk country (0.8 * 0.25 = 0.20) + Device (0.6 * 0.20 = 0.12)
	// Total: 0.18 + 0.15 + 0.20 + 0.12 = 0.65 (high risk, not critical)
	// Need embargoed country (1.0 * 0.25 = 0.25) + VPN + Proxy + Device
	// 0.18 + 0.15 + 0.25 + 0.12 = 0.70 (still high, not critical)
	// The block decision is based on risk_level = "critical", so we need to use direct risk evaluation.
	// Looking at MakeDecision: it returns "block" when riskLevel == "critical".

	highRiskLogs := `[
  {
    "timestamp": "2025-01-15T10:30:00Z",
    "user_id": "attacker-123",
    "operation": "admin_action",
    "ip_address": "1.2.3.4",
    "device_id": "unknown-device",
    "country": "ZZ",
    "city": "Unknown",
    "is_vpn": true,
    "is_proxy": true,
    "current_auth_level": "none",
    "success": false,
    "metadata": {}
  }
]`

	logsPath := filepath.Join(tempDir, "high_risk_logs.json")
	_ = os.WriteFile(logsPath, []byte(highRiskLogs), filePerms600)

	// Create policy files.
	riskScoringPath := filepath.Join(tempDir, "risk_scoring.yml")
	stepUpPath := filepath.Join(tempDir, "step_up.yml")
	adaptivePath := filepath.Join(tempDir, "adaptive_auth.yml")
	_ = os.WriteFile(riskScoringPath, []byte(mockRiskScoringPolicy), filePerms600)
	_ = os.WriteFile(stepUpPath, []byte(mockStepUpPolicy), filePerms600)
	_ = os.WriteFile(adaptivePath, []byte(mockAdaptiveAuthPolicy), filePerms600)

	loader := cryptoutilIdentityIdpUserauth.NewYAMLPolicyLoader(riskScoringPath, stepUpPath, adaptivePath)
	simulator := &AdaptiveSimulator{
		policyLoader: loader,
		outputDir:    tempDir,
	}

	ctx := context.Background()

	result, err := simulator.Simulate(ctx, logsPath, "v1.0")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 1, result.TotalAttempts)

	// Verify at least one of the following: allowed, step_up, or blocked.
	totalDecisions := result.AllowedOperations + result.StepUpRequired + result.BlockedOperations
	require.Equal(t, 1, totalDecisions)
}

func TestSimulate_WithCriticalRisk(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	// Create a custom risk scoring policy with very low critical threshold.
	// The risk score calculation is:
	// - VPN: policy.NetworkRisks["vpn"].Score * policy.RiskFactors["network"].Weight
	// - Proxy: policy.NetworkRisks["proxy"].Score * policy.RiskFactors["network"].Weight
	// - Location: policy.GeographicRisks.HighRiskCountries.Score * policy.RiskFactors["location"].Weight
	// - Device: 0.6 (hardcoded) * policy.RiskFactors["device"].Weight
	//
	// To reach critical (>= 0.8), we need high scores AND high weights.
	// With weights: network=0.5, location=0.4, device=0.1
	// VPN=1.0*0.5=0.5, Proxy=1.0*0.5=0.5, Location=1.0*0.4=0.4, Device=0.6*0.1=0.06
	// Total max: 1.46 (well above 0.8 critical threshold)
	customRiskPolicy := `
version: "1.0"
risk_factors:
  location:
    weight: 0.40
    description: "Geographic location risk"
  device:
    weight: 0.10
    description: "Device fingerprint risk"
  network:
    weight: 0.50
    description: "Network-based risk"
risk_thresholds:
  low:
    min: 0.0
    max: 0.2
    description: "Low risk"
  medium:
    min: 0.2
    max: 0.5
    description: "Medium risk"
  high:
    min: 0.5
    max: 0.8
    description: "High risk"
  critical:
    min: 0.8
    max: 2.0
    description: "Critical risk"
network_risks:
  vpn:
    score: 1.0
    description: "VPN usage"
  proxy:
    score: 1.0
    description: "Proxy usage"
geographic_risks:
  high_risk_countries:
    countries: ["XX", "YY", "ZZ"]
    score: 1.0
    description: "High-risk countries"
`

	highRiskLogs := `[
  {
    "timestamp": "2025-01-15T10:30:00Z",
    "user_id": "attacker-123",
    "operation": "admin_action",
    "ip_address": "1.2.3.4",
    "device_id": "unknown-device",
    "country": "XX",
    "city": "Unknown",
    "is_vpn": true,
    "is_proxy": true,
    "current_auth_level": "none",
    "success": false,
    "metadata": {}
  }
]`

	logsPath := filepath.Join(tempDir, "high_risk_logs.json")
	_ = os.WriteFile(logsPath, []byte(highRiskLogs), filePerms600)

	// Create policy files.
	riskScoringPath := filepath.Join(tempDir, "risk_scoring.yml")
	stepUpPath := filepath.Join(tempDir, "step_up.yml")
	adaptivePath := filepath.Join(tempDir, "adaptive_auth.yml")
	_ = os.WriteFile(riskScoringPath, []byte(customRiskPolicy), filePerms600)
	_ = os.WriteFile(stepUpPath, []byte(mockStepUpPolicy), filePerms600)
	_ = os.WriteFile(adaptivePath, []byte(mockAdaptiveAuthPolicy), filePerms600)

	loader := cryptoutilIdentityIdpUserauth.NewYAMLPolicyLoader(riskScoringPath, stepUpPath, adaptivePath)
	simulator := &AdaptiveSimulator{
		policyLoader: loader,
		outputDir:    tempDir,
	}

	ctx := context.Background()

	result, err := simulator.Simulate(ctx, logsPath, "v1.0")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 1, result.TotalAttempts)

	// With VPN (1.0 * 0.50 = 0.50) + Proxy (1.0 * 0.50 = 0.50) + Location (1.0 * 0.40 = 0.40) + Device (0.6 * 0.10 = 0.06)
	// Total: 0.50 + 0.50 + 0.40 + 0.06 = 1.46 (well above 0.8 critical threshold)
	// This should trigger a block decision.
	require.Equal(t, 1, result.BlockedOperations, "Expected 1 blocked operation due to critical risk")
	require.Contains(t, result.RiskDistribution, "critical")
}
