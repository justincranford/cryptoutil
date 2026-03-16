// Copyright (c) 2025 Justin Cranford

package userauth

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	meterName = "cryptoutil.identity.userauth"
)

// TelemetryRecorder records telemetry for adaptive authentication operations.
type TelemetryRecorder struct {
	meter metric.Meter

	// Risk scoring metrics.
	riskScoreHistogram       metric.Float64Histogram
	riskLevelCounter         metric.Int64Counter
	confidenceScoreHistogram metric.Float64Histogram

	// Step-up metrics.
	stepUpTriggeredCounter metric.Int64Counter
	stepUpMethodCounter    metric.Int64Counter
	stepUpSuccessCounter   metric.Int64Counter
	stepUpFailureCounter   metric.Int64Counter

	// Policy evaluation metrics.
	policyEvaluationDuration metric.Float64Histogram
	policyLoadDuration       metric.Float64Histogram
	policyReloadCounter      metric.Int64Counter

	// Blocking metrics.
	blockedOperationsCounter metric.Int64Counter
	allowedOperationsCounter metric.Int64Counter

	// Error metrics.
	riskAssessmentErrorCounter metric.Int64Counter
	policyLoadErrorCounter     metric.Int64Counter
}

// NewTelemetryRecorder creates new telemetry recorder for adaptive auth.
func NewTelemetryRecorder(_ context.Context) (*TelemetryRecorder, error) {
	meter := otel.Meter(meterName)

	// Risk scoring metrics.
	riskScoreHistogram, err := meter.Float64Histogram(
		"identity.risk.score",
		metric.WithDescription("Distribution of risk scores calculated for authentication attempts"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create risk score histogram: %w", err)
	}

	riskLevelCounter, err := meter.Int64Counter(
		"identity.risk.level.total",
		metric.WithDescription("Total count of risk levels assigned to authentication attempts"),
		metric.WithUnit("{attempts}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create risk level counter: %w", err)
	}

	confidenceScoreHistogram, err := meter.Float64Histogram(
		"identity.risk.confidence",
		metric.WithDescription("Distribution of confidence scores for risk assessments"),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create confidence score histogram: %w", err)
	}

	// Step-up metrics.
	stepUpTriggeredCounter, err := meter.Int64Counter(
		"identity.stepup.triggered.total",
		metric.WithDescription("Total count of step-up authentication triggers"),
		metric.WithUnit("{triggers}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create step-up triggered counter: %w", err)
	}

	stepUpMethodCounter, err := meter.Int64Counter(
		"identity.stepup.method.total",
		metric.WithDescription("Total count of step-up methods used"),
		metric.WithUnit("{attempts}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create step-up method counter: %w", err)
	}

	stepUpSuccessCounter, err := meter.Int64Counter(
		"identity.stepup.success.total",
		metric.WithDescription("Total count of successful step-up authentications"),
		metric.WithUnit("{successes}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create step-up success counter: %w", err)
	}

	stepUpFailureCounter, err := meter.Int64Counter(
		"identity.stepup.failure.total",
		metric.WithDescription("Total count of failed step-up authentications"),
		metric.WithUnit("{failures}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create step-up failure counter: %w", err)
	}

	// Policy evaluation metrics.
	policyEvaluationDuration, err := meter.Float64Histogram(
		"identity.policy.evaluation.duration",
		metric.WithDescription("Duration of policy evaluation operations"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy evaluation duration histogram: %w", err)
	}

	policyLoadDuration, err := meter.Float64Histogram(
		"identity.policy.load.duration",
		metric.WithDescription("Duration of policy loading operations"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy load duration histogram: %w", err)
	}

	policyReloadCounter, err := meter.Int64Counter(
		"identity.policy.reload.total",
		metric.WithDescription("Total count of policy reload operations"),
		metric.WithUnit("{reloads}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy reload counter: %w", err)
	}

	// Blocking metrics.
	blockedOperationsCounter, err := meter.Int64Counter(
		"identity.operations.blocked.total",
		metric.WithDescription("Total count of blocked operations due to risk level"),
		metric.WithUnit("{operations}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create blocked operations counter: %w", err)
	}

	allowedOperationsCounter, err := meter.Int64Counter(
		"identity.operations.allowed.total",
		metric.WithDescription("Total count of allowed operations"),
		metric.WithUnit("{operations}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create allowed operations counter: %w", err)
	}

	// Error metrics.
	riskAssessmentErrorCounter, err := meter.Int64Counter(
		"identity.risk.assessment.errors.total",
		metric.WithDescription("Total count of risk assessment errors"),
		metric.WithUnit("{errors}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create risk assessment error counter: %w", err)
	}

	policyLoadErrorCounter, err := meter.Int64Counter(
		"identity.policy.load.errors.total",
		metric.WithDescription("Total count of policy loading errors"),
		metric.WithUnit("{errors}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy load error counter: %w", err)
	}

	return &TelemetryRecorder{
		meter:                      meter,
		riskScoreHistogram:         riskScoreHistogram,
		riskLevelCounter:           riskLevelCounter,
		confidenceScoreHistogram:   confidenceScoreHistogram,
		stepUpTriggeredCounter:     stepUpTriggeredCounter,
		stepUpMethodCounter:        stepUpMethodCounter,
		stepUpSuccessCounter:       stepUpSuccessCounter,
		stepUpFailureCounter:       stepUpFailureCounter,
		policyEvaluationDuration:   policyEvaluationDuration,
		policyLoadDuration:         policyLoadDuration,
		policyReloadCounter:        policyReloadCounter,
		blockedOperationsCounter:   blockedOperationsCounter,
		allowedOperationsCounter:   allowedOperationsCounter,
		riskAssessmentErrorCounter: riskAssessmentErrorCounter,
		policyLoadErrorCounter:     policyLoadErrorCounter,
	}, nil
}

// RecordRiskScore records risk score for authentication attempt.
func (t *TelemetryRecorder) RecordRiskScore(ctx context.Context, score float64, level RiskLevel, confidence float64) {
	t.riskScoreHistogram.Record(ctx, score)

	t.riskLevelCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("risk_level", level.String()),
	))

	t.confidenceScoreHistogram.Record(ctx, confidence)
}

// RecordStepUpTriggered records step-up authentication trigger.
func (t *TelemetryRecorder) RecordStepUpTriggered(ctx context.Context, operation string, currentLevel, requiredLevel string) {
	t.stepUpTriggeredCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("current_level", currentLevel),
		attribute.String("required_level", requiredLevel),
	))
}

// RecordStepUpMethod records step-up method usage.
func (t *TelemetryRecorder) RecordStepUpMethod(ctx context.Context, method string, success bool) {
	t.stepUpMethodCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("method", method),
	))

	if success {
		t.stepUpSuccessCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("method", method),
		))
	} else {
		t.stepUpFailureCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("method", method),
		))
	}
}

// RecordPolicyEvaluation records policy evaluation duration.
func (t *TelemetryRecorder) RecordPolicyEvaluation(ctx context.Context, operation string, duration time.Duration) {
	t.policyEvaluationDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("operation", operation),
	))
}

// RecordPolicyLoad records policy loading operation.
func (t *TelemetryRecorder) RecordPolicyLoad(ctx context.Context, policyType string, duration time.Duration, success bool) {
	t.policyLoadDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(
		attribute.String("policy_type", policyType),
		attribute.Bool("success", success),
	))

	if !success {
		t.policyLoadErrorCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("policy_type", policyType),
		))
	}
}

// RecordPolicyReload records policy reload event.
func (t *TelemetryRecorder) RecordPolicyReload(ctx context.Context, policyType string) {
	t.policyReloadCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("policy_type", policyType),
	))
}

// RecordOperationDecision records authentication decision.
func (t *TelemetryRecorder) RecordOperationDecision(ctx context.Context, operation string, riskLevel RiskLevel, blocked bool) {
	if blocked {
		t.blockedOperationsCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("operation", operation),
			attribute.String("risk_level", riskLevel.String()),
		))
	} else {
		t.allowedOperationsCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("operation", operation),
			attribute.String("risk_level", riskLevel.String()),
		))
	}
}

// RecordRiskAssessmentError records risk assessment error.
func (t *TelemetryRecorder) RecordRiskAssessmentError(ctx context.Context, errorType string) {
	t.riskAssessmentErrorCounter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("error_type", errorType),
	))
}
