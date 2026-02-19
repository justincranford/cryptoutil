// Copyright (c) 2025 Justin Cranford

package main

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilIdentityIdpUserauth "cryptoutil/internal/apps/identity/idp/userauth"
)
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
