// Copyright (c) 2025 Justin Cranford

package userauth

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYAMLPolicyLoader_LoadStepUpPolicies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		yamlContent   string
		wantVersion   string
		wantPolicies  int
		wantError     bool
		errorContains string
	}{
		{
			name: "valid step-up policies",
			yamlContent: `version: "1.0"
policies:
  transfer_funds:
    operation_pattern: "transfer_funds"
    required_level: "step_up"
    allowed_methods: ["sms_otp", "totp", "webauthn"]
    max_age: "5m"
    description: "High-value financial transfer"
  change_password:
    operation_pattern: "change_password"
    required_level: "step_up"
    allowed_methods: ["email_otp", "totp"]
    max_age: "2m"
    description: "Password change operation"
default_policy:
  required_level: "basic"
  allowed_methods: ["password"]
  max_age: "24h"
auth_levels:
  none: 0
  basic: 1
  mfa: 2
  step_up: 3
  strong_mfa: 4
step_up_methods: {}
session_durations: {}
monitoring:
  step_up_rate: "15%"
  blocked_operations: "5%"
  fallback_methods: "20%"
  description: ""
`,
			wantVersion:  "1.0",
			wantPolicies: 2,
			wantError:    false,
		},
		{
			name: "invalid YAML syntax",
			yamlContent: `version: "1.0"
policies:
  - invalid list format
`,
			wantError:     true,
			errorContains: "failed to parse",
		},
		{
			name: "empty default policy required_level",
			yamlContent: `version: "1.0"
policies:
  transfer_funds:
    operation_pattern: "transfer_funds"
    required_level: "step_up"
    allowed_methods: ["sms_otp"]
    max_age: "5m"
default_policy:
  required_level: ""
  allowed_methods: []
  max_age: "24h"
auth_levels: {}
step_up_methods: {}
session_durations: {}
monitoring:
  step_up_rate: ""
  blocked_operations: ""
  fallback_methods: ""
  description: ""
`,
			wantError:     true,
			errorContains: "cannot be empty",
		},
		{
			name: "empty policies map",
			yamlContent: `version: "1.0"
policies: {}
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
`,
			wantError:     true,
			errorContains: "cannot be empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			// Create temporary file.
			tempDir := t.TempDir()
			policyFile := filepath.Join(tempDir, "step_up.yml")
			err := os.WriteFile(policyFile, []byte(tc.yamlContent), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			// Create loader.
			loader := NewYAMLPolicyLoader("", policyFile, "")

			// Load policies.
			policies, err := loader.LoadStepUpPolicies(ctx)

			if tc.wantError {
				require.Error(t, err)

				if tc.errorContains != "" {
					require.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, policies)
				require.Equal(t, tc.wantVersion, policies.Version)
				require.Len(t, policies.Policies, tc.wantPolicies)
				require.NotEmpty(t, policies.DefaultPolicy.RequiredLevel)
			}
		})
	}
}

func TestYAMLPolicyLoader_LoadAdaptiveAuthPolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		yamlContent    string
		wantVersion    string
		wantRiskLevels int
		wantError      bool
		errorContains  string
	}{
		{
			name: "valid adaptive auth policy",
			yamlContent: `version: "1.0"
name: "default-adaptive-auth"
description: "Default adaptive authentication policy"
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
    description: "Low-risk context"
  medium:
    risk_score_range:
      min: 0.1
      max: 0.4
    required_methods: ["password", "mfa"]
    session_duration: "1h"
    idle_timeout: "15m"
    step_up_required: true
    allow_new_device_registration: true
    allow_password_reset: true
    monitoring:
      log_level: "warn"
      alert_on_failure: false
    description: "Medium-risk context"
fallback_policy:
  on_error: "medium"
  on_low_confidence: "medium"
  description: "Conservative fallback"
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
`,
			wantVersion:    "1.0",
			wantRiskLevels: 2,
			wantError:      false,
		},
		{
			name: "invalid YAML syntax",
			yamlContent: `version: "1.0"
risk_based_auth:
  - invalid list
`,
			wantError:     true,
			errorContains: "failed to parse",
		},
		{
			name: "empty risk_based_auth",
			yamlContent: `version: "1.0"
name: "test"
description: ""
risk_based_auth: {}
fallback_policy:
  on_error: "medium"
  on_low_confidence: "medium"
  description: ""
grace_periods: {}
device_trust:
  remember_device_duration: ""
  max_trusted_devices: 0
  require_reauth_on_new_device: false
  device_fingerprint_factors: []
location_trust:
  remember_location_duration: ""
  impossible_travel_threshold: ""
  high_risk_countries_block: false
  location_factors: []
behavior_trust:
  baseline_establishment_period: ""
  min_events_for_baseline: 0
  tracked_patterns: []
tuning:
  risk_score_decay_rate: 0.0
  risk_score_spike_factor: 0.0
  confidence_threshold_low: 0.0
  confidence_threshold_medium: 0.0
  confidence_threshold_high: 0.0
  baseline_staleness_threshold: ""
`,
			wantError:     true,
			errorContains: "cannot be empty",
		},
		{
			name: "empty fallback on_error",
			yamlContent: `version: "1.0"
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
  on_error: ""
  on_low_confidence: ""
  description: ""
grace_periods: {}
device_trust:
  remember_device_duration: ""
  max_trusted_devices: 0
  require_reauth_on_new_device: false
  device_fingerprint_factors: []
location_trust:
  remember_location_duration: ""
  impossible_travel_threshold: ""
  high_risk_countries_block: false
  location_factors: []
behavior_trust:
  baseline_establishment_period: ""
  min_events_for_baseline: 0
  tracked_patterns: []
tuning:
  risk_score_decay_rate: 0.0
  risk_score_spike_factor: 0.0
  confidence_threshold_low: 0.0
  confidence_threshold_medium: 0.0
  confidence_threshold_high: 0.0
  baseline_staleness_threshold: ""
`,
			wantError:     true,
			errorContains: "cannot be empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			// Create temporary file.
			tempDir := t.TempDir()
			policyFile := filepath.Join(tempDir, "adaptive_auth.yml")
			err := os.WriteFile(policyFile, []byte(tc.yamlContent), cryptoutilSharedMagic.CacheFilePermissions)
			require.NoError(t, err)

			// Create loader.
			loader := NewYAMLPolicyLoader("", "", policyFile)

			// Load policy.
			policy, err := loader.LoadAdaptiveAuthPolicy(ctx)

			if tc.wantError {
				require.Error(t, err)

				if tc.errorContains != "" {
					require.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, policy)
				require.Equal(t, tc.wantVersion, policy.Version)
				require.Len(t, policy.RiskBasedAuth, tc.wantRiskLevels)
				require.NotEmpty(t, policy.FallbackPolicy.OnError)
			}
		})
	}
}
