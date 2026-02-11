// Copyright (c) 2025 Justin Cranford
//
//

package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	googleUuid "github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// MFATelemetry provides observability instrumentation for MFA operations.
type MFATelemetry struct {
	logger                  *slog.Logger
	tracer                  trace.Tracer
	validationCounter       metric.Int64Counter
	validationDuration      metric.Float64Histogram
	replayAttemptsCounter   metric.Int64Counter
	requiresMFACounter      metric.Int64Counter
	getRequiredFactorsGauge metric.Int64UpDownCounter
}

// NewMFATelemetry creates MFA telemetry instrumentation.
func NewMFATelemetry(logger *slog.Logger, metricsProvider metric.MeterProvider, tracesProvider trace.TracerProvider) (*MFATelemetry, error) {
	meter := metricsProvider.Meter("identity.idp.auth.mfa")
	tracer := tracesProvider.Tracer("identity.idp.auth.mfa")

	validationCounter, err := meter.Int64Counter(
		"mfa.validation.total",
		metric.WithDescription("Total MFA validation attempts"),
		metric.WithUnit("{validations}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create validation counter: %w", err)
	}

	validationDuration, err := meter.Float64Histogram(
		"mfa.validation.duration",
		metric.WithDescription("MFA validation duration"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create validation duration histogram: %w", err)
	}

	replayAttemptsCounter, err := meter.Int64Counter(
		"mfa.replay_attempts.total",
		metric.WithDescription("Total MFA replay attack attempts detected"),
		metric.WithUnit("{attempts}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create replay attempts counter: %w", err)
	}

	requiresMFACounter, err := meter.Int64Counter(
		"mfa.requires_mfa.total",
		metric.WithDescription("Total RequiresMFA checks"),
		metric.WithUnit("{checks}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create requires mfa counter: %w", err)
	}

	getRequiredFactorsGauge, err := meter.Int64UpDownCounter(
		"mfa.required_factors.count",
		metric.WithDescription("Number of required MFA factors per profile"),
		metric.WithUnit("{factors}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create required factors gauge: %w", err)
	}

	return &MFATelemetry{
		logger:                  logger,
		tracer:                  tracer,
		validationCounter:       validationCounter,
		validationDuration:      validationDuration,
		replayAttemptsCounter:   replayAttemptsCounter,
		requiresMFACounter:      requiresMFACounter,
		getRequiredFactorsGauge: getRequiredFactorsGauge,
	}, nil
}

// RecordValidation records MFA validation telemetry.
func (t *MFATelemetry) RecordValidation(ctx context.Context, factorType string, success bool, duration time.Duration, isReplay bool) {
	attrs := []attribute.KeyValue{
		attribute.String("factor_type", factorType),
		attribute.Bool("success", success),
		attribute.Bool("is_replay", isReplay),
	}

	t.validationCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	t.validationDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))

	if isReplay {
		t.replayAttemptsCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("factor_type", factorType),
		))

		t.logger.WarnContext(ctx, "MFA replay attack detected",
			"factor_type", factorType,
			"duration_ms", duration.Milliseconds(),
		)
	} else {
		logLevel := slog.LevelInfo
		if !success {
			logLevel = slog.LevelWarn
		}

		t.logger.Log(ctx, logLevel, "MFA validation completed",
			"factor_type", factorType,
			"success", success,
			"duration_ms", duration.Milliseconds(),
		)
	}
}

// RecordRequiresMFA records RequiresMFA check telemetry.
func (t *MFATelemetry) RecordRequiresMFA(ctx context.Context, authProfileID googleUuid.UUID, requiresMFA bool) {
	attrs := []attribute.KeyValue{
		attribute.String("auth_profile_id", authProfileID.String()),
		attribute.Bool("requires_mfa", requiresMFA),
	}

	t.requiresMFACounter.Add(ctx, 1, metric.WithAttributes(attrs...))

	t.logger.DebugContext(ctx, "MFA requirement check",
		"auth_profile_id", authProfileID.String(),
		"requires_mfa", requiresMFA,
	)
}

// RecordRequiredFactors records required factors count telemetry.
func (t *MFATelemetry) RecordRequiredFactors(ctx context.Context, authProfileID googleUuid.UUID, factorCount int) {
	attrs := []attribute.KeyValue{
		attribute.String("auth_profile_id", authProfileID.String()),
	}

	t.getRequiredFactorsGauge.Add(ctx, int64(factorCount), metric.WithAttributes(attrs...))

	t.logger.DebugContext(ctx, "MFA required factors",
		"auth_profile_id", authProfileID.String(),
		"factor_count", factorCount,
	)
}

// StartValidationSpan starts distributed tracing span for MFA validation.
func (t *MFATelemetry) StartValidationSpan(ctx context.Context, factorType string, authProfileID googleUuid.UUID) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, "mfa.validate_factor",
		trace.WithAttributes(
			attribute.String("factor_type", factorType),
			attribute.String("auth_profile_id", authProfileID.String()),
		),
	)
}

// StartRequiresMFASpan starts distributed tracing span for RequiresMFA check.
func (t *MFATelemetry) StartRequiresMFASpan(ctx context.Context, authProfileID googleUuid.UUID) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, "mfa.requires_mfa",
		trace.WithAttributes(
			attribute.String("auth_profile_id", authProfileID.String()),
		),
	)
}

// StartGetRequiredFactorsSpan starts distributed tracing span for GetRequiredFactors.
func (t *MFATelemetry) StartGetRequiredFactorsSpan(ctx context.Context, authProfileID googleUuid.UUID) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, "mfa.get_required_factors",
		trace.WithAttributes(
			attribute.String("auth_profile_id", authProfileID.String()),
		),
	)
}
