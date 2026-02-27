// Copyright (c) 2025 Justin Cranford

package userauth

import (
	"context"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewTelemetryRecorder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	recorder, err := NewTelemetryRecorder(ctx)

	require.NoError(t, err)
	require.NotNil(t, recorder)
	require.NotNil(t, recorder.meter)
	require.NotNil(t, recorder.riskScoreHistogram)
	require.NotNil(t, recorder.riskLevelCounter)
	require.NotNil(t, recorder.confidenceScoreHistogram)
	require.NotNil(t, recorder.stepUpTriggeredCounter)
	require.NotNil(t, recorder.stepUpMethodCounter)
	require.NotNil(t, recorder.stepUpSuccessCounter)
	require.NotNil(t, recorder.stepUpFailureCounter)
	require.NotNil(t, recorder.policyEvaluationDuration)
	require.NotNil(t, recorder.policyLoadDuration)
	require.NotNil(t, recorder.policyReloadCounter)
	require.NotNil(t, recorder.blockedOperationsCounter)
	require.NotNil(t, recorder.allowedOperationsCounter)
	require.NotNil(t, recorder.riskAssessmentErrorCounter)
	require.NotNil(t, recorder.policyLoadErrorCounter)
}

func TestRecordRiskScore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	recorder, err := NewTelemetryRecorder(ctx)
	require.NoError(t, err)

	tests := []struct {
		name       string
		score      float64
		level      RiskLevel
		confidence float64
	}{
		{
			name:       "low risk",
			score:      cryptoutilSharedMagic.Tolerance10Percent,
			level:      RiskLevelLow,
			confidence: cryptoutilSharedMagic.RiskScoreVeryHigh,
		},
		{
			name:       "medium risk",
			score:      0.3,
			level:      RiskLevelMedium,
			confidence: cryptoutilSharedMagic.RiskScoreHigh,
		},
		{
			name:       "high risk",
			score:      cryptoutilSharedMagic.RiskScoreHigh,
			level:      RiskLevelHigh,
			confidence: cryptoutilSharedMagic.RiskScoreExtreme,
		},
		{
			name:       "critical risk",
			score:      cryptoutilSharedMagic.RiskScoreExtreme,
			level:      RiskLevelCritical,
			confidence: 0.95,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Should not panic.
			require.NotPanics(t, func() {
				recorder.RecordRiskScore(ctx, tc.score, tc.level, tc.confidence)
			})
		})
	}
}

func TestRecordStepUpTriggered(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	recorder, err := NewTelemetryRecorder(ctx)
	require.NoError(t, err)

	tests := []struct {
		name          string
		operation     string
		currentLevel  string
		requiredLevel string
	}{
		{
			name:          "transfer requires step-up from basic to mfa",
			operation:     "transfer_funds",
			currentLevel:  "basic",
			requiredLevel: cryptoutilSharedMagic.AMRMultiFactor,
		},
		{
			name:          "sensitive op requires strong mfa",
			operation:     "change_password",
			currentLevel:  cryptoutilSharedMagic.AMRMultiFactor,
			requiredLevel: "strong_mfa",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotPanics(t, func() {
				recorder.RecordStepUpTriggered(ctx, tc.operation, tc.currentLevel, tc.requiredLevel)
			})
		})
	}
}

func TestRecordStepUpMethod(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	recorder, err := NewTelemetryRecorder(ctx)
	require.NoError(t, err)

	tests := []struct {
		name    string
		method  string
		success bool
	}{
		{
			name:    "successful OTP",
			method:  cryptoutilSharedMagic.AMRTOTP,
			success: true,
		},
		{
			name:    "failed TOTP",
			method:  cryptoutilSharedMagic.MFATypeTOTP,
			success: false,
		},
		{
			name:    "successful WebAuthn",
			method:  cryptoutilSharedMagic.MFATypeWebAuthn,
			success: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotPanics(t, func() {
				recorder.RecordStepUpMethod(ctx, tc.method, tc.success)
			})
		})
	}
}

func TestRecordPolicyEvaluation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	recorder, err := NewTelemetryRecorder(ctx)
	require.NoError(t, err)

	tests := []struct {
		name      string
		operation string
		duration  time.Duration
	}{
		{
			name:      "fast evaluation",
			operation: "view_balance",
			duration:  cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Millisecond,
		},
		{
			name:      "slow evaluation",
			operation: "transfer_funds",
			duration:  cryptoutilSharedMagic.JoseJAMaxMaterials * time.Millisecond,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotPanics(t, func() {
				recorder.RecordPolicyEvaluation(ctx, tc.operation, tc.duration)
			})
		})
	}
}

func TestRecordPolicyLoad(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	recorder, err := NewTelemetryRecorder(ctx)
	require.NoError(t, err)

	tests := []struct {
		name       string
		policyType string
		duration   time.Duration
		success    bool
	}{
		{
			name:       "successful risk scoring policy load",
			policyType: "risk_scoring",
			duration:   cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond,
			success:    true,
		},
		{
			name:       "failed step-up policy load",
			policyType: "step_up",
			duration:   cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days * time.Millisecond,
			success:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotPanics(t, func() {
				recorder.RecordPolicyLoad(ctx, tc.policyType, tc.duration, tc.success)
			})
		})
	}
}

func TestRecordPolicyReload(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	recorder, err := NewTelemetryRecorder(ctx)
	require.NoError(t, err)

	tests := []struct {
		name       string
		policyType string
	}{
		{
			name:       "reload risk scoring policy",
			policyType: "risk_scoring",
		},
		{
			name:       "reload step-up policy",
			policyType: "step_up",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotPanics(t, func() {
				recorder.RecordPolicyReload(ctx, tc.policyType)
			})
		})
	}
}

func TestRecordOperationDecision(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	recorder, err := NewTelemetryRecorder(ctx)
	require.NoError(t, err)

	tests := []struct {
		name      string
		operation string
		riskLevel RiskLevel
		blocked   bool
	}{
		{
			name:      "allow low risk operation",
			operation: "view_balance",
			riskLevel: RiskLevelLow,
			blocked:   false,
		},
		{
			name:      "block high risk operation",
			operation: "transfer_funds",
			riskLevel: RiskLevelHigh,
			blocked:   true,
		},
		{
			name:      "block critical risk operation",
			operation: "change_password",
			riskLevel: RiskLevelCritical,
			blocked:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotPanics(t, func() {
				recorder.RecordOperationDecision(ctx, tc.operation, tc.riskLevel, tc.blocked)
			})
		})
	}
}

func TestRecordRiskAssessmentError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	recorder, err := NewTelemetryRecorder(ctx)
	require.NoError(t, err)

	tests := []struct {
		name      string
		errorType string
	}{
		{
			name:      "user baseline not found",
			errorType: "baseline_not_found",
		},
		{
			name:      "geoip lookup failed",
			errorType: "geoip_lookup_failed",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			require.NotPanics(t, func() {
				recorder.RecordRiskAssessmentError(ctx, tc.errorType)
			})
		})
	}
}
