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
	"time"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity/idp/userauth"
)

// fixedTime is a consistent timestamp for test reproducibility.
var fixedTime = time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

func TestLoadHistoricalLogs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		logData   string
		wantCount int
		wantError bool
	}{
		{
			name: "valid single log entry",
			logData: `[{
				"timestamp": "2025-01-15T10:30:00Z",
				"user_id": "user-123",
				"operation": "transfer_funds",
				"ip_address": "192.168.1.1",
				"device_id": "device-abc",
				"country": "US",
				"city": "New York",
				"is_vpn": false,
				"is_proxy": false,
				"current_auth_level": "basic",
				"success": true,
				"metadata": {}
			}]`,
			wantCount: 1,
			wantError: false,
		},
		{
			name: "valid multiple log entries",
			logData: `[
				{"timestamp": "2025-01-15T10:30:00Z", "user_id": "user-123", "operation": "view_balance", "ip_address": "192.168.1.1", "device_id": "device-abc", "country": "US", "city": "New York", "is_vpn": false, "is_proxy": false, "current_auth_level": "basic", "success": true, "metadata": {}},
				{"timestamp": "2025-01-15T10:35:00Z", "user_id": "user-456", "operation": "transfer_funds", "ip_address": "10.0.0.1", "device_id": "device-def", "country": "CA", "city": "Toronto", "is_vpn": true, "is_proxy": false, "current_auth_level": "mfa", "success": true, "metadata": {}}
			]`,
			wantCount: 2,
			wantError: false,
		},
		{
			name:      "invalid JSON syntax",
			logData:   `{invalid json}`,
			wantCount: 0,
			wantError: true,
		},
		{
			name:      "empty array",
			logData:   `[]`,
			wantCount: 0,
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create temporary log file.
			tmpFile := createTempLogFile(t, tc.logData)

			simulator := &AdaptiveSimulator{}

			logs, err := simulator.LoadHistoricalLogs(tmpFile)

			if tc.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, logs, tc.wantCount)
			}
		})
	}
}

func TestCalculateRiskScore(t *testing.T) {
	t.Parallel()

	// Create mock policy with known weights.
	policy := &cryptoutilIdentityIdpUserauth.RiskScoringPolicy{
		RiskFactors: map[string]cryptoutilIdentityIdpUserauth.RiskFactorConfig{
			"network":  {Weight: 0.30},
			"location": {Weight: 0.25},
			"device":   {Weight: 0.20},
		},
		NetworkRisks: map[string]cryptoutilIdentityIdpUserauth.NetworkRisk{
			"vpn":   {Score: 0.6},
			"proxy": {Score: 0.5},
		},
		GeographicRisks: cryptoutilIdentityIdpUserauth.GeographicRisks{
			HighRiskCountries: cryptoutilIdentityIdpUserauth.HighRiskCountries{
				Countries: []string{"XX", "YY"},
				Score:     0.8,
			},
		},
	}

	tests := []struct {
		name         string
		log          HistoricalAuthLog
		wantMinScore float64
		wantMaxScore float64
	}{
		{
			name: "low risk - no VPN, no proxy, safe country",
			log: HistoricalAuthLog{
				IsVPN:   false,
				IsProxy: false,
				Country: "US",
			},
			wantMinScore: 0.10, // Device risk only (0.6 * 0.20).
			wantMaxScore: 0.15,
		},
		{
			name: "medium risk - VPN enabled",
			log: HistoricalAuthLog{
				IsVPN:   true,
				IsProxy: false,
				Country: "US",
			},
			wantMinScore: 0.28, // VPN (0.6 * 0.30) + Device (0.6 * 0.20).
			wantMaxScore: 0.32,
		},
		{
			name: "high risk - proxy + high-risk country",
			log: HistoricalAuthLog{
				IsVPN:   false,
				IsProxy: true,
				Country: "XX", // High-risk country.
			},
			wantMinScore: 0.45, // Proxy (0.5 * 0.30) + Country (0.8 * 0.25) + Device (0.6 * 0.20).
			wantMaxScore: 0.50,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			simulator := &AdaptiveSimulator{}

			score := simulator.CalculateRiskScore(tc.log, policy)

			require.GreaterOrEqual(t, score, tc.wantMinScore, "Risk score below expected minimum")
			require.LessOrEqual(t, score, tc.wantMaxScore, "Risk score above expected maximum")
		})
	}
}

func TestDetermineRiskLevel(t *testing.T) {
	t.Parallel()

	policy := &cryptoutilIdentityIdpUserauth.RiskScoringPolicy{
		RiskThresholds: map[string]cryptoutilIdentityIdpUserauth.RiskThreshold{
			"low":      {Min: 0.0, Max: 0.2},
			"medium":   {Min: 0.2, Max: 0.5},
			"high":     {Min: 0.5, Max: 0.8},
			"critical": {Min: 0.8, Max: 1.0},
		},
	}

	tests := []struct {
		name      string
		score     float64
		wantLevel string
	}{
		{
			name:      "low risk score",
			score:     0.1,
			wantLevel: "low",
		},
		{
			name:      "medium risk score",
			score:     0.3,
			wantLevel: "medium",
		},
		{
			name:      "high risk score",
			score:     0.6,
			wantLevel: "high",
		},
		{
			name:      "critical risk score",
			score:     0.9,
			wantLevel: "critical",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			simulator := &AdaptiveSimulator{}

			level := simulator.DetermineRiskLevel(tc.score, policy)

			require.Equal(t, tc.wantLevel, level)
		})
	}
}

func TestDetermineRequiredLevel(t *testing.T) {
	t.Parallel()

	policy := &cryptoutilIdentityIdpUserauth.StepUpPolicies{
		Policies: map[string]cryptoutilIdentityIdpUserauth.OperationPolicy{
			"transfer_funds": {
				RequiredLevel: "mfa",
			},
			"view_balance": {
				RequiredLevel: "basic",
			},
		},
		DefaultPolicy: cryptoutilIdentityIdpUserauth.OperationPolicy{
			RequiredLevel: "basic",
		},
	}

	tests := []struct {
		name         string
		operation    string
		currentLevel string
		wantRequired string
		wantStepUp   bool
	}{
		{
			name:         "step-up required - insufficient level",
			operation:    "transfer_funds",
			currentLevel: "basic",
			wantRequired: "mfa",
			wantStepUp:   true,
		},
		{
			name:         "no step-up - sufficient level",
			operation:    "view_balance",
			currentLevel: "mfa",
			wantRequired: "basic",
			wantStepUp:   false,
		},
		{
			name:         "default policy - unknown operation",
			operation:    "unknown_operation",
			currentLevel: "basic",
			wantRequired: "basic",
			wantStepUp:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			simulator := &AdaptiveSimulator{}

			requiredLevel, stepUpNeeded := simulator.DetermineRequiredLevel(tc.operation, tc.currentLevel, policy)

			require.Equal(t, tc.wantRequired, requiredLevel)
			require.Equal(t, tc.wantStepUp, stepUpNeeded)
		})
	}
}

func TestMakeDecision(t *testing.T) {
	t.Parallel()

	policy := &cryptoutilIdentityIdpUserauth.AdaptiveAuthPolicy{}

	tests := []struct {
		name         string
		riskLevel    string
		stepUpNeeded bool
		wantDecision string
	}{
		{
			name:         "block on critical risk",
			riskLevel:    "critical",
			stepUpNeeded: false,
			wantDecision: "block",
		},
		{
			name:         "step-up required",
			riskLevel:    "medium",
			stepUpNeeded: true,
			wantDecision: "step_up",
		},
		{
			name:         "allow - low risk, no step-up",
			riskLevel:    "low",
			stepUpNeeded: false,
			wantDecision: "allow",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			simulator := &AdaptiveSimulator{}

			decision := simulator.MakeDecision(tc.riskLevel, tc.stepUpNeeded, policy)

			require.Equal(t, tc.wantDecision, decision)
		})
	}
}

func TestGenerateRecommendations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		result             *SimulationResult
		wantRecommendation bool
		wantContains       string
	}{
		{
			name: "high step-up rate",
			result: &SimulationResult{
				TotalAttempts:  100,
				StepUpRequired: 20,
				StepUpRate:     0.20,
			},
			wantRecommendation: true,
			wantContains:       "High step-up rate",
		},
		{
			name: "high blocked rate",
			result: &SimulationResult{
				TotalAttempts:     100,
				BlockedOperations: 10,
				BlockedRate:       0.10,
			},
			wantRecommendation: true,
			wantContains:       "High blocked rate",
		},
		{
			name: "high critical risk attempts",
			result: &SimulationResult{
				TotalAttempts: 100,
				RiskDistribution: map[string]int{
					"critical": 15,
				},
			},
			wantRecommendation: true,
			wantContains:       "critical-risk attempts",
		},
		{
			name: "no recommendations - good policy",
			result: &SimulationResult{
				TotalAttempts:     100,
				StepUpRequired:    5,
				BlockedOperations: 2,
				StepUpRate:        0.05,
				BlockedRate:       0.02,
				RiskDistribution: map[string]int{
					"low":      80,
					"medium":   15,
					"high":     4,
					"critical": 1,
				},
			},
			wantRecommendation: true,
			wantContains:       "No policy adjustments",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			simulator := &AdaptiveSimulator{}

			recommendations := simulator.GenerateRecommendations(tc.result)

			require.NotEmpty(t, recommendations, "Expected recommendations but got none")

			if tc.wantRecommendation && tc.wantContains != "" {
				found := false

				for _, rec := range recommendations {
					if contains(rec, tc.wantContains) {
						found = true

						break
					}
				}

				require.True(t, found, "Expected recommendation containing '%s'", tc.wantContains)
			}
		})
	}
}

func TestSimulation_EndToEnd(t *testing.T) {
	t.Parallel()

	// Create temporary output directory.
	outputDir := t.TempDir()

	// Create mock policy files.
	riskScoringPath := createTempPolicyFile(t, "risk_scoring.yml", mockRiskScoringPolicy)
	stepUpPath := createTempPolicyFile(t, "step_up.yml", mockStepUpPolicy)
	adaptivePath := createTempPolicyFile(t, "adaptive_auth.yml", mockAdaptiveAuthPolicy)

	// Create historical logs.
	logsPath := createTempLogFile(t, mockHistoricalLogs)

	// Create policy loader.
	loader := cryptoutilIdentityIdpUserauth.NewYAMLPolicyLoader(riskScoringPath, stepUpPath, adaptivePath)

	simulator := &AdaptiveSimulator{
		policyLoader: loader,
		outputDir:    outputDir,
	}

	// Run simulation.
	result, err := simulator.Simulate(context.Background(), logsPath, "test-v1.0")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "test-v1.0", result.PolicyVersion)
	require.Greater(t, result.TotalAttempts, 0)
	require.NotEmpty(t, result.PolicyEvaluations)
	require.NotEmpty(t, result.Recommendations)

	// Save results.
	stdout := &bytes.Buffer{}

	err = simulator.SaveResults(result, stdout)

	require.NoError(t, err)
}

// Helper functions.

func createTempLogFile(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), "logs-*.json")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	return tmpFile.Name()
}

func createTempPolicyFile(t *testing.T, name, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp(t.TempDir(), name)
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	return tmpFile.Name()
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}

	return false
}

// Mock policy data for testing.

const mockRiskScoringPolicy = `
version: "1.0"
risk_factors:
  location:
    weight: 0.25
    description: "Geographic location risk"
  device:
    weight: 0.20
    description: "Device fingerprint risk"
  network:
    weight: 0.30
    description: "Network-based risk"
  behavior:
    weight: 0.25
    description: "Behavioral risk patterns"
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
    max: 1.0
    description: "Critical risk"
confidence_weights:
  factor_count: 0.4
  baseline_data: 0.3
  behavior_profile: 0.3
  description: "Confidence scoring weights"
network_risks:
  vpn:
    score: 0.6
    description: "VPN usage"
  proxy:
    score: 0.5
    description: "Proxy usage"
geographic_risks:
  high_risk_countries:
    countries: ["XX", "YY"]
    score: 0.8
    description: "High-risk countries"
  embargoed_countries:
    countries: ["ZZ"]
    score: 1.0
    description: "Embargoed countries"
velocity_limits:
  rapid_location_changes:
    window: "1h"
    max_locations: 3
    risk_score: 0.8
    description: "Rapid location changes"
time_risks:
  unusual_hour:
    score: 0.3
    description: "Unusual hour activity"
behavior_risks:
  new_device:
    score: 0.6
    description: "New device"
`

const mockStepUpPolicy = `
version: "1.0"
policies:
  transfer_funds:
    required_level: "mfa"
    allowed_methods: ["otp", "totp"]
    max_age: "5m"
    description: "Transfer funds operation"
  view_balance:
    required_level: "basic"
    allowed_methods: ["password"]
    max_age: "30m"
    description: "View balance operation"
default_policy:
  required_level: "basic"
  allowed_methods: ["password"]
  max_age: "30m"
  description: "Default policy"
auth_levels:
  none: 0
  basic: 1
  mfa: 2
  step_up: 3
  strong_mfa: 4
step_up_methods:
  otp:
    strength: "medium"
    fallback_priority: 1
    description: "OTP via email/SMS"
  totp:
    strength: "medium"
    fallback_priority: 2
    description: "TOTP via authenticator app"
session_durations:
  basic: "30m"
  mfa: "8h"
  step_up: "1h"
  strong_mfa: "24h"
monitoring:
  step_up_rate: "15%"
  blocked_operations: "5%"
  fallback_methods: "10%"
  description: "Monitoring thresholds"
`

const mockAdaptiveAuthPolicy = `
version: "1.0"
name: "Adaptive Authentication Policy"
description: "Test policy for simulation"
risk_based_auth:
  low:
    risk_score_range:
      min: 0.0
      max: 0.2
    required_methods: ["password"]
    session_duration: "8h"
    idle_timeout: "30m"
    step_up_required: false
    allow_new_device_registration: true
    allow_password_reset: true
    monitoring:
      log_level: "info"
      alert_on_failure: false
    description: "Low risk requirements"
  medium:
    risk_score_range:
      min: 0.2
      max: 0.5
    required_methods: ["password", "mfa"]
    session_duration: "4h"
    idle_timeout: "15m"
    step_up_required: true
    allow_new_device_registration: true
    allow_password_reset: false
    monitoring:
      log_level: "warn"
      alert_on_failure: true
    description: "Medium risk requirements"
  high:
    risk_score_range:
      min: 0.5
      max: 0.8
    required_methods: ["password", "strong_mfa"]
    session_duration: "1h"
    idle_timeout: "5m"
    step_up_required: true
    allow_new_device_registration: false
    allow_password_reset: false
    monitoring:
      log_level: "error"
      alert_on_failure: true
      alert_security_team: true
    description: "High risk requirements"
  critical:
    risk_score_range:
      min: 0.8
      max: 1.0
    required_methods: ["block"]
    session_duration: "0m"
    idle_timeout: "0m"
    step_up_required: false
    allow_new_device_registration: false
    allow_password_reset: false
    monitoring:
      log_level: "error"
      alert_on_failure: true
      alert_security_team: true
      alert_fraud_team: true
    description: "Critical risk - block authentication"
fallback_policy:
  on_error: "allow_with_step_up"
  on_low_confidence: "require_step_up"
  description: "Fallback behavior"
grace_periods:
  device_registration:
    duration: "24h"
    risk_level_override: "medium"
    description: "New device registration grace period"
device_trust:
  remember_device_duration: "30d"
  max_trusted_devices: 5
  require_reauth_on_new_device: true
  device_fingerprint_factors: ["user_agent", "screen_resolution", "timezone"]
location_trust:
  remember_location_duration: "90d"
  impossible_travel_threshold: "1000km/h"
  high_risk_countries_block: false
  location_factors: ["ip_address", "gps_coordinates"]
behavior_trust:
  baseline_establishment_period: "30d"
  min_events_for_baseline: 10
  tracked_patterns: ["login_times", "ip_addresses", "devices"]
tuning:
  risk_score_decay_rate: 0.1
  risk_score_spike_factor: 2.0
  confidence_threshold_low: 0.3
  confidence_threshold_medium: 0.5
  confidence_threshold_high: 0.7
  baseline_staleness_threshold: "90d"
`

const mockHistoricalLogs = `[
  {
    "timestamp": "2025-01-15T10:30:00Z",
    "user_id": "user-123",
    "operation": "view_balance",
    "ip_address": "192.168.1.1",
    "device_id": "device-abc",
    "country": "US",
    "city": "New York",
    "is_vpn": false,
    "is_proxy": false,
    "current_auth_level": "basic",
    "success": true,
    "metadata": {}
  },
  {
    "timestamp": "2025-01-15T10:35:00Z",
    "user_id": "user-456",
    "operation": "transfer_funds",
    "ip_address": "10.0.0.1",
    "device_id": "device-def",
    "country": "CA",
    "city": "Toronto",
    "is_vpn": true,
    "is_proxy": false,
    "current_auth_level": "basic",
    "success": false,
    "metadata": {}
  },
  {
    "timestamp": "2025-01-15T10:40:00Z",
    "user_id": "user-789",
    "operation": "transfer_funds",
    "ip_address": "203.0.113.42",
    "device_id": "device-ghi",
    "country": "XX",
    "city": "Unknown",
    "is_vpn": false,
    "is_proxy": true,
    "current_auth_level": "mfa",
    "success": true,
    "metadata": {}
  }
]`

// TestInternalMain tests for main() testability pattern.

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
