// Copyright (c) 2025 Justin Cranford

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityIdpUserauth "cryptoutil/internal/identity/idp/userauth"
)

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
	// NOTE: This test requires policy YAML files which depend on working directory.
	// Skip if not in project root.
	if _, err := os.Stat("configs/identity/policies/risk_scoring.yml"); os.IsNotExist(err) {
		t.Skip("Skipping test - policy files not accessible from current working directory")
	}

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

	args := []string{"adaptive-sim", "--logs=" + logsPath, "--output=" + outputDir}
	stdin := strings.NewReader("")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	exitCode := internalMain(args, stdin, stdout, stderr)

	require.Equal(t, exitSuccess, exitCode, "stderr: %s", stderr.String())
	require.Contains(t, stdout.String(), "Adaptive Authentication Policy Simulation")
	require.Contains(t, stdout.String(), "Total Attempts: 1")
}
