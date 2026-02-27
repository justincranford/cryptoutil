// Copyright (c) 2025 Justin Cranford

package observability

import (
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewMetricsRegistry(t *testing.T) {
	t.Parallel()

	registry := NewMetricsRegistry()
	require.NotNil(t, registry)
	require.NotNil(t, registry.counters)
	require.NotNil(t, registry.gauges)
	require.NotNil(t, registry.histograms)
}

func TestMetricsRegistry_Counter(t *testing.T) {
	t.Parallel()

	registry := NewMetricsRegistry()

	// Initial value should be 0.
	require.Equal(t, int64(0), registry.GetCounter("test_counter"))

	// Increment.
	registry.IncrementCounter("test_counter")
	require.Equal(t, int64(1), registry.GetCounter("test_counter"))

	// Add value.
	registry.AddToCounter("test_counter", cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	require.Equal(t, int64(cryptoutilSharedMagic.DefaultEmailOTPLength), registry.GetCounter("test_counter"))

	// Get non-existent counter.
	require.Equal(t, int64(0), registry.GetCounter("non_existent"))
}

func TestMetricsRegistry_Gauge(t *testing.T) {
	t.Parallel()

	registry := NewMetricsRegistry()

	// Initial value should be 0.
	require.Equal(t, int64(0), registry.GetGauge("test_gauge"))

	// Set value.
	registry.SetGauge("test_gauge", cryptoutilSharedMagic.AnswerToLifeUniverseEverything)
	require.Equal(t, int64(cryptoutilSharedMagic.AnswerToLifeUniverseEverything), registry.GetGauge("test_gauge"))

	// Increment.
	registry.IncrementGauge("test_gauge")
	require.Equal(t, int64(cryptoutilSharedMagic.DefaultCodeChallengeLength), registry.GetGauge("test_gauge"))

	// Decrement.
	registry.DecrementGauge("test_gauge")
	require.Equal(t, int64(cryptoutilSharedMagic.AnswerToLifeUniverseEverything), registry.GetGauge("test_gauge"))

	// Get non-existent gauge.
	require.Equal(t, int64(0), registry.GetGauge("non_existent"))
}

func TestMetricsRegistry_Histogram(t *testing.T) {
	t.Parallel()

	registry := NewMetricsRegistry()

	// Record values.
	registry.RecordHistogram("test_histogram", cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	registry.RecordHistogram("test_histogram", cryptoutilSharedMagic.MaxErrorDisplay)
	registry.RecordHistogram("test_histogram", cryptoutilSharedMagic.TLSTestEndEntityCertValidity30Days)

	histogram := registry.GetHistogram("test_histogram")
	require.NotNil(t, histogram)
	require.Equal(t, int64(3), histogram.Count())
	require.InDelta(t, cryptoutilSharedMagic.RateLimitSecondsPerMinute, histogram.Sum(), 0.001)

	// Get non-existent histogram.
	require.Nil(t, registry.GetHistogram("non_existent"))
}

func TestHistogram(t *testing.T) {
	t.Parallel()

	histogram := NewHistogram()

	// Observe values.
	histogram.Observe(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	histogram.Observe(15)
	histogram.Observe(cryptoutilSharedMagic.IMMaxUsernameLength)
	histogram.Observe(cryptoutilSharedMagic.JoseJAMaxMaterials)
	histogram.Observe(cryptoutilSharedMagic.TestDefaultRateLimitServiceIP)

	require.Equal(t, int64(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries), histogram.Count())
	require.InDelta(t, 670.0, histogram.Sum(), 0.001)

	buckets := histogram.GetBucketCounts()
	require.NotEmpty(t, buckets)
}

func TestHistogramWithCustomBuckets(t *testing.T) {
	t.Parallel()

	buckets := []float64{cryptoutilSharedMagic.JoseJADefaultMaxMaterials, cryptoutilSharedMagic.IMMaxUsernameLength, cryptoutilSharedMagic.JoseJAMaxMaterials}
	histogram := NewHistogramWithBuckets(buckets)

	histogram.Observe(cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)
	histogram.Observe(cryptoutilSharedMagic.TLSMaxValidityCACertYears)
	histogram.Observe(75)
	histogram.Observe(200)

	require.Equal(t, int64(4), histogram.Count())

	bucketCounts := histogram.GetBucketCounts()
	require.Equal(t, int64(1), bucketCounts[cryptoutilSharedMagic.JoseJADefaultMaxMaterials])
	require.Equal(t, int64(2), bucketCounts[cryptoutilSharedMagic.IMMaxUsernameLength])
	require.Equal(t, int64(3), bucketCounts[cryptoutilSharedMagic.JoseJAMaxMaterials])
}

func TestMetricsRegistry_GetAllMetrics(t *testing.T) {
	t.Parallel()

	registry := NewMetricsRegistry()

	// Add various metrics.
	registry.IncrementCounter("counter1")
	registry.SetGauge("gauge1", cryptoutilSharedMagic.JoseJADefaultMaxMaterials)
	registry.RecordHistogram("histogram1", cryptoutilSharedMagic.DefaultSidecarHealthCheckMaxRetries)

	metrics := registry.GetAllMetrics()
	require.NotEmpty(t, metrics)

	// Should have counter, gauge, and histogram count/sum.
	require.GreaterOrEqual(t, len(metrics), 4)
}

func TestMetricsRegistry_Reset(t *testing.T) {
	t.Parallel()

	registry := NewMetricsRegistry()

	// Add metrics.
	registry.IncrementCounter("counter1")
	registry.SetGauge("gauge1", cryptoutilSharedMagic.JoseJADefaultMaxMaterials)

	// Reset.
	registry.Reset()

	// Verify cleared.
	require.Equal(t, int64(0), registry.GetCounter("counter1"))
	require.Equal(t, int64(0), registry.GetGauge("gauge1"))
}

func TestNewCAMetrics(t *testing.T) {
	t.Parallel()

	metrics := NewCAMetrics()
	require.NotNil(t, metrics)
	require.NotNil(t, metrics.Registry())
}

func TestCAMetrics_CertificateOperations(t *testing.T) {
	t.Parallel()

	metrics := NewCAMetrics()

	// Record operations.
	metrics.RecordCertificateIssued()
	metrics.RecordCertificateIssued()
	metrics.RecordCertificateRevoked()
	metrics.RecordCertificateExpired()

	registry := metrics.Registry()
	require.Equal(t, int64(2), registry.GetCounter(MetricCertificatesIssued))
	require.Equal(t, int64(1), registry.GetCounter(MetricCertificatesRevoked))
	require.Equal(t, int64(1), registry.GetCounter(MetricCertificatesExpired))
	require.Equal(t, int64(0), registry.GetGauge(MetricCertificatesActive))
}

func TestCAMetrics_CRLAndOCSP(t *testing.T) {
	t.Parallel()

	metrics := NewCAMetrics()

	metrics.RecordCRLGeneration()
	metrics.RecordOCSPRequest(10.5)
	metrics.RecordOCSPRequest(20.3)

	registry := metrics.Registry()
	require.Equal(t, int64(1), registry.GetCounter(MetricCRLGenerations))
	require.Equal(t, int64(2), registry.GetCounter(MetricOCSPRequests))

	histogram := registry.GetHistogram(MetricOCSPResponseTime)
	require.NotNil(t, histogram)
	require.Equal(t, int64(2), histogram.Count())
}

func TestCAMetrics_Enrollment(t *testing.T) {
	t.Parallel()

	metrics := NewCAMetrics()

	metrics.RecordEnrollmentRequest()
	metrics.RecordEnrollmentApproval()
	metrics.RecordEnrollmentRejection()

	registry := metrics.Registry()
	require.Equal(t, int64(1), registry.GetCounter(MetricEnrollmentRequests))
	require.Equal(t, int64(1), registry.GetCounter(MetricEnrollmentApprovals))
	require.Equal(t, int64(1), registry.GetCounter(MetricEnrollmentRejections))
}

func TestCAMetrics_Operations(t *testing.T) {
	t.Parallel()

	metrics := NewCAMetrics()

	metrics.RecordKeyGeneration()
	metrics.RecordSigningOperation(15.5)
	metrics.RecordValidationError()
	metrics.RecordTimestampIssued()

	registry := metrics.Registry()
	require.Equal(t, int64(1), registry.GetCounter(MetricKeyGenerations))
	require.Equal(t, int64(1), registry.GetCounter(MetricSigningOperations))
	require.Equal(t, int64(1), registry.GetCounter(MetricValidationErrors))
	require.Equal(t, int64(1), registry.GetCounter(MetricTimestampsIssued))
}

func TestNewTracer(t *testing.T) {
	t.Parallel()

	tracer := NewTracer()
	require.NotNil(t, tracer)
	require.NotNil(t, tracer.spans)
}

func TestTracer_StartSpan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tracer := NewTracer()

	span := tracer.StartSpan(ctx, "test-operation")
	require.NotNil(t, span)
	require.NotEmpty(t, span.TraceID)
	require.NotEmpty(t, span.SpanID)
	require.Equal(t, "test-operation", span.Name)
	require.Equal(t, SpanStatusUnset, span.Status)
}

func TestTracer_EndSpan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tracer := NewTracer()

	span := tracer.StartSpan(ctx, "test-operation")

	time.Sleep(time.Millisecond) // Ensure non-zero duration.
	tracer.EndSpan(span, SpanStatusOK)

	require.NotZero(t, span.EndTime)
	require.Greater(t, span.Duration, time.Duration(0))
	require.Equal(t, SpanStatusOK, span.Status)
}

func TestTracer_EndSpan_Nil(t *testing.T) {
	t.Parallel()

	tracer := NewTracer()
	tracer.EndSpan(nil, SpanStatusOK) // Should not panic.
}

func TestTracer_AddAttribute(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tracer := NewTracer()

	span := tracer.StartSpan(ctx, "test-operation")
	tracer.AddAttribute(span, "key", "value")

	require.Equal(t, "value", span.Attributes["key"])
}

func TestTracer_AddAttribute_Nil(t *testing.T) {
	t.Parallel()

	tracer := NewTracer()
	tracer.AddAttribute(nil, "key", "value") // Should not panic.
}

func TestTracer_AddEvent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tracer := NewTracer()

	span := tracer.StartSpan(ctx, "test-operation")
	tracer.AddEvent(span, "test-event", map[string]string{"attr": "value"})

	require.Len(t, span.Events, 1)
	require.Equal(t, "test-event", span.Events[0].Name)
	require.Equal(t, "value", span.Events[0].Attributes["attr"])
}

func TestTracer_AddEvent_Nil(t *testing.T) {
	t.Parallel()

	tracer := NewTracer()
	tracer.AddEvent(nil, "test-event", nil) // Should not panic.
}

func TestTracer_GetSpan(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tracer := NewTracer()

	span := tracer.StartSpan(ctx, "test-operation")
	retrieved := tracer.GetSpan(span.SpanID)

	require.Equal(t, span, retrieved)

	// Get non-existent.
	require.Nil(t, tracer.GetSpan("non-existent"))
}

func TestNewAuditLogger(t *testing.T) {
	t.Parallel()

	logger := NewAuditLogger()
	require.NotNil(t, logger)
	require.Empty(t, logger.events)
}

func TestAuditLogger_Log(t *testing.T) {
	t.Parallel()

	logger := NewAuditLogger()

	event := AuditEvent{
		EventType: AuditTypeCertificate,
		Actor:     "admin",
		Action:    "issue",
		Resource:  "cert-123",
		Outcome:   AuditOutcomeSuccess,
		Details:   map[string]string{cryptoutilSharedMagic.ClaimProfile: "tls-server"},
	}

	logger.Log(event)

	events := logger.GetEvents()
	require.Len(t, events, 1)
	require.NotEmpty(t, events[0].ID)
	require.NotZero(t, events[0].Timestamp)
	require.Equal(t, AuditTypeCertificate, events[0].EventType)
}

func TestAuditLogger_GetEventsByType(t *testing.T) {
	t.Parallel()

	logger := NewAuditLogger()

	// Log different event types.
	logger.Log(AuditEvent{EventType: AuditTypeCertificate, Actor: "admin", Action: "issue"})
	logger.Log(AuditEvent{EventType: AuditTypeAuthentication, Actor: "user", Action: cryptoutilSharedMagic.PromptLogin})
	logger.Log(AuditEvent{EventType: AuditTypeCertificate, Actor: "admin", Action: "revoke"})

	// Get by type.
	certEvents := logger.GetEventsByType(AuditTypeCertificate)
	require.Len(t, certEvents, 2)

	authEvents := logger.GetEventsByType(AuditTypeAuthentication)
	require.Len(t, authEvents, 1)

	keyEvents := logger.GetEventsByType(AuditTypeKey)
	require.Empty(t, keyEvents)
}

func TestAuditLogger_Clear(t *testing.T) {
	t.Parallel()

	logger := NewAuditLogger()

	logger.Log(AuditEvent{EventType: AuditTypeCertificate, Actor: "admin"})
	require.Len(t, logger.GetEvents(), 1)

	logger.Clear()
	require.Empty(t, logger.GetEvents())
}

func TestSpanStatus_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status   SpanStatus
		expected string
	}{
		{SpanStatusUnset, "unset"},
		{SpanStatusOK, "ok"},
		{SpanStatusError, cryptoutilSharedMagic.StringError},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, string(tc.status))
		})
	}
}

func TestAuditEventType_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		eventType AuditEventType
		expected  string
	}{
		{AuditTypeAuthentication, "authentication"},
		{AuditTypeAuthorization, "authorization"},
		{AuditTypeCertificate, "certificate"},
		{AuditTypeKey, "key"},
		{AuditTypeConfiguration, "configuration"},
		{AuditTypeAdministration, "administration"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, string(tc.eventType))
		})
	}
}

func TestAuditOutcome_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		outcome  AuditOutcome
		expected string
	}{
		{AuditOutcomeSuccess, "success"},
		{AuditOutcomeFailure, "failure"},
		{AuditOutcomeDenied, "denied"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, string(tc.outcome))
		})
	}
}

func TestMetricType_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		metricType MetricType
		expected   string
	}{
		{MetricTypeCounter, "counter"},
		{MetricTypeGauge, "gauge"},
		{MetricTypeHistogram, "histogram"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, string(tc.metricType))
		})
	}
}
