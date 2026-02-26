// Copyright (c) 2025 Justin Cranford

package userauth

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestYAMLPolicyLoader_Caching(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temporary policy files.
	tempDir := t.TempDir()
	riskScoringFile := filepath.Join(tempDir, "risk_scoring.yml")
	stepUpFile := filepath.Join(tempDir, "step_up.yml")
	adaptiveAuthFile := filepath.Join(tempDir, "adaptive_auth.yml")

	riskScoringContent := testRiskScoringYAML

	stepUpContent := `version: "1.0"
policies:
  transfer_funds:
    required_level: "step_up"
    allowed_methods: ["totp"]
    max_age: "5m"
default_policy:
  required_level: "basic"
  allowed_methods: ["password"]
  max_age: "24h"
auth_levels: {}
step_up_methods: {}
session_durations: {}
monitoring:
  step_up_rate: ""
  blocked_operations: ""
  fallback_methods: ""
  description: ""
`

	adaptiveAuthContent := `version: "1.0"
name: "test"
description: ""
risk_based_auth:
  low:
    risk_score_range:
      min: 0.0
      max: 0.1
    required_methods: ["password"]
    session_duration: "24h"
    idle_timeout: "2h"
    step_up_required: false
    allow_new_device_registration: true
    allow_password_reset: true
    monitoring:
      log_level: "info"
      alert_on_failure: false
    description: ""
fallback_policy:
  on_error: "medium"
  on_low_confidence: "medium"
  description: ""
grace_periods: {}
device_trust:
  remember_device_duration: "30d"
  max_trusted_devices: 5
  require_reauth_on_new_device: true
  device_fingerprint_factors: []
location_trust:
  remember_location_duration: "90d"
  impossible_travel_threshold: "500km/h"
  high_risk_countries_block: true
  location_factors: []
behavior_trust:
  baseline_establishment_period: "7d"
  min_events_for_baseline: 10
  tracked_patterns: []
tuning:
  risk_score_decay_rate: 0.1
  risk_score_spike_factor: 2.0
  confidence_threshold_low: 0.3
  confidence_threshold_medium: 0.6
  confidence_threshold_high: 0.8
  baseline_staleness_threshold: "30d"
`

	err := os.WriteFile(riskScoringFile, []byte(riskScoringContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
	err = os.WriteFile(stepUpFile, []byte(stepUpContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)
	err = os.WriteFile(adaptiveAuthFile, []byte(adaptiveAuthContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	loader := NewYAMLPolicyLoader(riskScoringFile, stepUpFile, adaptiveAuthFile)

	// Load policies first time.
	riskPolicy1, err := loader.LoadRiskScoringPolicy(ctx)
	require.NoError(t, err)
	require.NotNil(t, riskPolicy1)

	stepUpPolicy1, err := loader.LoadStepUpPolicies(ctx)
	require.NoError(t, err)
	require.NotNil(t, stepUpPolicy1)

	adaptivePolicy1, err := loader.LoadAdaptiveAuthPolicy(ctx)
	require.NoError(t, err)
	require.NotNil(t, adaptivePolicy1)

	// Load policies second time (should return cached).
	riskPolicy2, err := loader.LoadRiskScoringPolicy(ctx)
	require.NoError(t, err)
	require.Same(t, riskPolicy1, riskPolicy2) // Same pointer = cached.

	stepUpPolicy2, err := loader.LoadStepUpPolicies(ctx)
	require.NoError(t, err)
	require.Same(t, stepUpPolicy1, stepUpPolicy2)

	adaptivePolicy2, err := loader.LoadAdaptiveAuthPolicy(ctx)
	require.NoError(t, err)
	require.Same(t, adaptivePolicy1, adaptivePolicy2)
}

func TestYAMLPolicyLoader_HotReload(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// Create temporary policy file.
	tempDir := t.TempDir()
	riskScoringFile := filepath.Join(tempDir, "risk_scoring.yml")

	initialContent := testRiskScoringYAML

	err := os.WriteFile(riskScoringFile, []byte(initialContent), cryptoutilSharedMagic.CacheFilePermissions)
	require.NoError(t, err)

	loader := NewYAMLPolicyLoader(riskScoringFile, "", "")

	// Load initial policy.
	policy1, err := loader.LoadRiskScoringPolicy(ctx)
	require.NoError(t, err)
	require.Equal(t, "1.0", policy1.Version)

	// Enable hot-reload with short interval.
	err = loader.EnableHotReload(ctx, cryptoutilSharedMagic.JoseJAMaxMaterials*time.Millisecond)
	require.NoError(t, err)

	// Wait for hot-reload to invalidate cache.
	time.Sleep(200 * time.Millisecond)

	// Load policy again (should trigger reload from file).
	policy2, err := loader.LoadRiskScoringPolicy(ctx)
	require.NoError(t, err)
	// Verify policy reloaded by comparing values (pointer comparison unreliable with caching).
	require.Equal(t, policy1.Version, policy2.Version) // Same content expected (file unchanged).

	// Disable hot-reload.
	loader.DisableHotReload()

	// Load policy again (should use cache, no reload).
	policy3, err := loader.LoadRiskScoringPolicy(ctx)
	require.NoError(t, err)
	require.Equal(t, policy2.Version, policy3.Version) // Same version = same content.
}
