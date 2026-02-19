// Copyright (c) 2025 Justin Cranford

package main

import (
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
