// Copyright (c) 2025 Justin Cranford

package userauth

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const testRiskScoringYAML = `version: "1.0"
risk_factors:
  location:
    weight: 1.0
    description: ""
risk_thresholds:
  low:
    min: 0.0
    max: 0.1
    auth_requirements: ["basic"]
    max_session_duration: "24h"
    description: ""
confidence_weights:
  factor_count: 0.5
  baseline_data: 0.15
  behavior_profile: 0.10
  description: ""
network_risks: {}
geographic_risks:
  high_risk_countries:
    countries: []
    score: 0.6
    description: ""
  embargoed_countries:
    countries: []
    score: 0.8
    description: ""
velocity_limits: {}
time_risks: {}
behavior_risks: {}
`

func TestYAMLPolicyLoader_LoadRiskScoringPolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		yamlContent    string
		wantVersion    string
		wantFactors    int
		wantThresholds int
		wantError      bool
		errorContains  string
	}{
		{
			name: "valid risk scoring policy",
			yamlContent: `version: "1.0"
risk_factors:
  location:
    weight: 0.25
    description: "Geographic location anomalies"
  device:
    weight: 0.20
    description: "Device fingerprint anomalies"
  time:
    weight: 0.15
    description: "Time-based anomalies"
  behavior:
    weight: 0.20
    description: "User behavior anomalies"
  network:
    weight: 0.10
    description: "VPN/proxy/Tor detection"
  velocity:
    weight: 0.10
    description: "Rapid attempt patterns"
risk_thresholds:
  low:
    min: 0.0
    max: 0.1
    auth_requirements: ["basic"]
    max_session_duration: "24h"
    description: "Low-risk context"
  medium:
    min: 0.1
    max: 0.4
    auth_requirements: ["basic", "mfa"]
    max_session_duration: "1h"
    description: "Medium-risk context"
confidence_weights:
  factor_count: 0.5
  baseline_data: 0.15
  behavior_profile: 0.10
  description: "Confidence scoring weights"
network_risks: {}
geographic_risks:
  high_risk_countries:
    countries: []
    score: 0.6
    description: ""
  embargoed_countries:
    countries: []
    score: 0.8
    description: ""
velocity_limits: {}
time_risks: {}
behavior_risks: {}
`,
			wantVersion:    "1.0",
			wantFactors:    6,
			wantThresholds: 2,
			wantError:      false,
		},
		{
			name: "invalid YAML syntax",
			yamlContent: `version: "1.0"
risk_factors:
  location:
    weight: 0.25
    - invalid syntax
`,
			wantError:     true,
			errorContains: "failed to parse",
		},
		{
			name: "weights do not sum to 1.0",
			yamlContent: `version: "1.0"
risk_factors:
  location:
    weight: 0.5
    description: ""
  device:
    weight: 0.3
    description: ""
risk_thresholds:
  low:
    min: 0.0
    max: 0.1
    auth_requirements: ["basic"]
    max_session_duration: "24h"
    description: ""
confidence_weights:
  factor_count: 0.5
  baseline_data: 0.15
  behavior_profile: 0.10
  description: ""
network_risks: {}
geographic_risks:
  high_risk_countries:
    countries: []
    score: 0.6
    description: ""
  embargoed_countries:
    countries: []
    score: 0.8
    description: ""
velocity_limits: {}
time_risks: {}
behavior_risks: {}
`,
			wantError:     true,
			errorContains: "must sum to 1.0",
		},
		{
			name: "empty risk thresholds",
			yamlContent: `version: "1.0"
risk_factors:
  location:
    weight: 1.0
    description: ""
risk_thresholds: {}
confidence_weights:
  factor_count: 0.5
  baseline_data: 0.15
  behavior_profile: 0.10
  description: ""
network_risks: {}
geographic_risks:
  high_risk_countries:
    countries: []
    score: 0.6
    description: ""
  embargoed_countries:
    countries: []
    score: 0.8
    description: ""
velocity_limits: {}
time_risks: {}
behavior_risks: {}
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
			policyFile := filepath.Join(tempDir, "risk_scoring.yml")
			err := os.WriteFile(policyFile, []byte(tc.yamlContent), 0o600)
			require.NoError(t, err)

			// Create loader.
			loader := NewYAMLPolicyLoader(policyFile, "", "")

			// Load policy.
			policy, err := loader.LoadRiskScoringPolicy(ctx)

			if tc.wantError {
				require.Error(t, err)

				if tc.errorContains != "" {
					require.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, policy)
				require.Equal(t, tc.wantVersion, policy.Version)
				require.Len(t, policy.RiskFactors, tc.wantFactors)
				require.Len(t, policy.RiskThresholds, tc.wantThresholds)

				// Verify weights sum to 1.0.
				var weightSum float64
				for _, factor := range policy.RiskFactors {
					weightSum += factor.Weight
				}

				require.InDelta(t, 1.0, weightSum, 0.001)
			}
		})
	}
}
