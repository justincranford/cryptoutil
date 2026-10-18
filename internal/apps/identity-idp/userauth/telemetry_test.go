// Copyright (c) 2025-2026 Justin Cranford.
package userauth

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/require"
)

const telemetryTestInjectedInstrumentError = "injected metric instrument creation failure"

type telemetryFailingMeter struct {
	metric.Meter
	instrumentName string
}

type telemetryTestMeterProvider struct {
	metric.MeterProvider
	meter metric.Meter
}

func (p telemetryTestMeterProvider) Meter(_ string, _ ...metric.MeterOption) metric.Meter {
	return p.meter
}

func (m telemetryFailingMeter) Float64Histogram(
	name string,
	options ...metric.Float64HistogramOption,
) (metric.Float64Histogram, error) {
	if name == m.instrumentName {
		return nil, errors.New(telemetryTestInjectedInstrumentError)
	}

	return m.Meter.Float64Histogram(name, options...)
}

func (m telemetryFailingMeter) Int64Counter(
	name string,
	options ...metric.Int64CounterOption,
) (metric.Int64Counter, error) {
	if name == m.instrumentName {
		return nil, errors.New(telemetryTestInjectedInstrumentError)
	}

	return m.Meter.Int64Counter(name, options...)
}

func TestNewTelemetryAuditLogger_InstrumentCreationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		instrumentName string
		wantError      string
	}{
		{
			name:           "generation counter",
			instrumentName: "identity.audit.token.generated",
			wantError:      "failed to create generation counter",
		},
		{
			name:           "validation counter",
			instrumentName: "identity.audit.token.validated",
			wantError:      "failed to create validation counter",
		},
		{
			name:           "invalidation counter",
			instrumentName: "identity.audit.token.invalidated",
			wantError:      "failed to create invalidation counter",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			baseMeter := noop.NewMeterProvider().Meter(meterName)
			meterProvider := telemetryTestMeterProvider{
				MeterProvider: noop.NewMeterProvider(),
				meter: telemetryFailingMeter{
					Meter:          baseMeter,
					instrumentName: tc.instrumentName,
				},
			}

			logger, err := NewTelemetryAuditLogger(meterProvider)

			require.Nil(t, logger)
			require.ErrorContains(t, err, tc.wantError)
			require.ErrorContains(t, err, telemetryTestInjectedInstrumentError)
		})
	}
}

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

func TestNewTelemetryRecorder_InstrumentCreationErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		instrumentName string
		wantError      string
	}{
		{name: "risk score histogram", instrumentName: "identity.risk.score", wantError: "failed to create risk score histogram"},
		{name: "risk level counter", instrumentName: "identity.risk.level.total", wantError: "failed to create risk level counter"},
		{name: "confidence score histogram", instrumentName: "identity.risk.confidence", wantError: "failed to create confidence score histogram"},
		{name: "step-up triggered counter", instrumentName: "identity.stepup.triggered.total", wantError: "failed to create step-up triggered counter"},
		{name: "step-up method counter", instrumentName: "identity.stepup.method.total", wantError: "failed to create step-up method counter"},
		{name: "step-up success counter", instrumentName: "identity.stepup.success.total", wantError: "failed to create step-up success counter"},
		{name: "step-up failure counter", instrumentName: "identity.stepup.failure.total", wantError: "failed to create step-up failure counter"},
		{name: "policy evaluation histogram", instrumentName: "identity.policy.evaluation.duration", wantError: "failed to create policy evaluation duration histogram"},
		{name: "policy load histogram", instrumentName: "identity.policy.load.duration", wantError: "failed to create policy load duration histogram"},
		{name: "policy reload counter", instrumentName: "identity.policy.reload.total", wantError: "failed to create policy reload counter"},
		{name: "blocked operations counter", instrumentName: "identity.operations.blocked.total", wantError: "failed to create blocked operations counter"},
		{name: "allowed operations counter", instrumentName: "identity.operations.allowed.total", wantError: "failed to create allowed operations counter"},
		{name: "risk assessment error counter", instrumentName: "identity.risk.assessment.errors.total", wantError: "failed to create risk assessment error counter"},
		{name: "policy load error counter", instrumentName: "identity.policy.load.errors.total", wantError: "failed to create policy load error counter"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			baseMeter := noop.NewMeterProvider().Meter(meterName)
			recorder, err := newTelemetryRecorder(telemetryFailingMeter{
				Meter:          baseMeter,
				instrumentName: tc.instrumentName,
			})

			require.Nil(t, recorder)
			require.ErrorContains(t, err, tc.wantError)
			require.ErrorContains(t, err, telemetryTestInjectedInstrumentError)
		})
	}
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
			operation:     valueTransferFunds,
			currentLevel:  authLevelStringBasic,
			requiredLevel: cryptoutilSharedMagic.AMRMultiFactor,
		},
		{
			name:          "sensitive op requires strong mfa",
			operation:     valueChangePassword,
			currentLevel:  cryptoutilSharedMagic.AMRMultiFactor,
			requiredLevel: authLevelStringStrongMFA,
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
			operation: testViewBalance,
			duration:  cryptoutilSharedMagic.JoseJADefaultMaxMaterials * time.Millisecond,
		},
		{
			name:      "slow evaluation",
			operation: valueTransferFunds,
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
			policyType: testRiskScoring,
			duration:   cryptoutilSharedMagic.IMMaxUsernameLength * time.Millisecond,
			success:    true,
		},
		{
			name:       "failed step-up policy load",
			policyType: authLevelStringStepUp,
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
			policyType: testRiskScoring,
		},
		{
			name:       "reload step-up policy",
			policyType: authLevelStringStepUp,
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
			operation: testViewBalance,
			riskLevel: RiskLevelLow,
			blocked:   false,
		},
		{
			name:      "block high risk operation",
			operation: valueTransferFunds,
			riskLevel: RiskLevelHigh,
			blocked:   true,
		},
		{
			name:      "block critical risk operation",
			operation: valueChangePassword,
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
