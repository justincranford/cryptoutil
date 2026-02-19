// Copyright (c) 2025 Justin Cranford

package main

import (
	"bytes"
	json "encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity/idp/userauth"
)
func TestInternalMain_MissingLogsFlag(t *testing.T) {
	t.Parallel()

	args := []string{"adaptive-sim"}
	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := internalMain(args, stdin, stdout, stderr)

	require.Equal(t, exitError, exitCode)
	require.Contains(t, stderr.String(), "Error: --logs flag is required")
	require.Contains(t, stderr.String(), "Usage: adaptive-sim")
}

func TestInternalMain_InvalidLogsFile(t *testing.T) {
	t.Parallel()

	args := []string{"adaptive-sim", "--logs=nonexistent.json"}
	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := internalMain(args, stdin, stdout, stderr)

	require.Equal(t, exitError, exitCode)
	require.Contains(t, stderr.String(), "Simulation failed")
}

func TestInternalMain_HappyPath(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	logsPath := filepath.Join(tempDir, "logs.json")
	outputDir := filepath.Join(tempDir, "output")

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

	require.Equal(t, exitSuccess, exitCode, "stderr: %s", stderr.String())
	require.Contains(t, stdout.String(), "Adaptive Authentication Policy Simulation")
	require.Contains(t, stdout.String(), "Total Attempts: 1")
}

func TestPrintSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		result          *SimulationResult
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "basic summary with all sections",
			result: &SimulationResult{
				PolicyVersion:     "v1.0.0",
				SimulationTime:    fixedTime,
				TotalAttempts:     100,
				AllowedOperations: 80,
				StepUpRequired:    15,
				BlockedOperations: 5,
				StepUpRate:        0.15,
				BlockedRate:       0.05,
				RiskDistribution: map[string]int{
					"low":      50,
					"medium":   30,
					"high":     15,
					"critical": 5,
				},
				Recommendations: []string{
					"Consider enabling MFA for high-risk operations",
					"Review blocked requests for false positives",
				},
			},
			wantContains: []string{
				"Adaptive Authentication Policy Simulation",
				"Policy Version: v1.0.0",
				"Total Attempts: 100",
				"=== Decisions ===",
				"Allowed: 80",
				"Step-Up Required: 15",
				"Blocked: 5",
				"=== Risk Distribution ===",
				"=== Recommendations ===",
				"1. Consider enabling MFA for high-risk operations",
				"2. Review blocked requests for false positives",
			},
		},
		{
			name: "summary with zero step-ups and blocks",
			result: &SimulationResult{
				PolicyVersion:     "v2.0.0",
				SimulationTime:    fixedTime,
				TotalAttempts:     50,
				AllowedOperations: 50,
				StepUpRequired:    0,
				BlockedOperations: 0,
				StepUpRate:        0.0,
				BlockedRate:       0.0,
				RiskDistribution: map[string]int{
					"low": 50,
				},
				Recommendations: []string{"No policy adjustments needed"},
			},
			wantContains: []string{
				"Total Attempts: 50",
				"Allowed: 50",
				"Step-Up Required: 0",
				"Blocked: 0",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			simulator := &AdaptiveSimulator{}
			stdout := &bytes.Buffer{}

			simulator.PrintSummary(tc.result, stdout)

			output := stdout.String()
			for _, want := range tc.wantContains {
				require.Contains(t, output, want)
			}

			for _, notWant := range tc.wantNotContains {
				require.NotContains(t, output, notWant)
			}
		})
	}
}

func TestDetermineRiskLevel_DefaultReturn(t *testing.T) {
	t.Parallel()

	// Create policy with gaps in thresholds (no threshold covers score 0.45).
	policy := &cryptoutilIdentityIdpUserauth.RiskScoringPolicy{
		RiskThresholds: map[string]cryptoutilIdentityIdpUserauth.RiskThreshold{
			"low":      {Min: 0.0, Max: 0.2},
			"high":     {Min: 0.6, Max: 0.8},
			"critical": {Min: 0.8, Max: 1.0},
		},
	}

	simulator := &AdaptiveSimulator{}

	// Score 0.45 falls between low (0.0-0.2) and high (0.6-0.8), no threshold matches.
	level := simulator.DetermineRiskLevel(0.45, policy)

	// Default return should be "medium".
	require.Equal(t, "medium", level)
}

func TestDetermineRiskLevel_EmptyThresholds(t *testing.T) {
	t.Parallel()

	// Create policy with empty thresholds.
	policy := &cryptoutilIdentityIdpUserauth.RiskScoringPolicy{
		RiskThresholds: map[string]cryptoutilIdentityIdpUserauth.RiskThreshold{},
	}

	simulator := &AdaptiveSimulator{}

	// Any score should return default "medium" when no thresholds defined.
	level := simulator.DetermineRiskLevel(0.5, policy)

	require.Equal(t, "medium", level)
}

func TestAuthLevelToInt_UnknownLevel(t *testing.T) {
	t.Parallel()

	simulator := &AdaptiveSimulator{}

	tests := []struct {
		name      string
		level     string
		wantValue int
	}{
		{
			name:      "unknown level returns default basic (1)",
			level:     "unknown_level",
			wantValue: 1,
		},
		{
			name:      "empty string returns default basic (1)",
			level:     "",
			wantValue: 1,
		},
		{
			name:      "random string returns default basic (1)",
			level:     "xyz123",
			wantValue: 1,
		},
		{
			name:      "known level - none",
			level:     "none",
			wantValue: 0,
		},
		{
			name:      "known level - basic",
			level:     "basic",
			wantValue: 1,
		},
		{
			name:      "known level - mfa",
			level:     "mfa",
			wantValue: 2,
		},
		{
			name:      "known level - step_up",
			level:     "step_up",
			wantValue: 3,
		},
		{
			name:      "known level - block",
			level:     "block",
			wantValue: 4,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := simulator.AuthLevelToInt(tc.level)
			require.Equal(t, tc.wantValue, result)
		})
	}
}

func TestSaveResults_OutputDirectoryCreation(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	outputDir := filepath.Join(tempDir, "nested", "output", "dir")

	// Directory doesn't exist yet.
	require.NoDirExists(t, outputDir)

	// Create directory before saving.
	err := os.MkdirAll(outputDir, 0o755)
	require.NoError(t, err)

	simulator := &AdaptiveSimulator{outputDir: outputDir}
	result := &SimulationResult{
		PolicyVersion:  "v1.0.0",
		SimulationTime: fixedTime,
		TotalAttempts:  10,
	}
	stdout := &bytes.Buffer{}

	err = simulator.SaveResults(result, stdout)
	require.NoError(t, err)
	require.Contains(t, stdout.String(), "Simulation results saved to:")
}

func TestLoadHistoricalLogs_NonexistentFile(t *testing.T) {
	t.Parallel()

	simulator := &AdaptiveSimulator{}

	_, err := simulator.LoadHistoricalLogs("/nonexistent/path/logs.json")

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read logs file")
}

func TestInternalMain_InvalidOutputDir(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	logsPath := filepath.Join(tempDir, "logs.json")

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

	// Use invalid output directory (file path instead of directory).
	invalidOutputDir := logsPath // This is a file, not a directory.

	args := []string{
		"adaptive-sim",
		"--logs=" + logsPath,
		"--output=" + invalidOutputDir,
		"--risk-scoring=" + riskScoringPath,
		"--step-up=" + stepUpPath,
		"--adaptive=" + adaptivePath,
	}
	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := internalMain(args, stdin, stdout, stderr)

	// Should fail because output directory is a file.
	require.Equal(t, exitError, exitCode)
	require.Contains(t, stderr.String(), "Failed to create output directory")
}

func TestMakeDecision_HighRiskWithStepUp(t *testing.T) {
	t.Parallel()

	simulator := &AdaptiveSimulator{}
	policy := &cryptoutilIdentityIdpUserauth.AdaptiveAuthPolicy{}

	// High risk with step-up should result in step_up decision.
	decision := simulator.MakeDecision("high", true, policy)

	require.Equal(t, "step_up", decision)
}

func TestMakeDecision_HighRiskNoStepUp(t *testing.T) {
	t.Parallel()

	simulator := &AdaptiveSimulator{}
	policy := &cryptoutilIdentityIdpUserauth.AdaptiveAuthPolicy{}

	// High risk without step-up should allow.
	decision := simulator.MakeDecision("high", false, policy)

	require.Equal(t, "allow", decision)
}

func TestSaveResults_WriteError(t *testing.T) {
	t.Parallel()

	// Use a path that cannot be written to (a file path where parent doesn't exist).
	simulator := &AdaptiveSimulator{outputDir: "/nonexistent/path/that/does/not/exist"}
	result := &SimulationResult{
		PolicyVersion:  "v1.0.0",
		SimulationTime: fixedTime,
		TotalAttempts:  10,
	}
	stdout := &bytes.Buffer{}

	err := simulator.SaveResults(result, stdout)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to write results")
}
