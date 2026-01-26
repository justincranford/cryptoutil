// Copyright (c) 2025 Justin Cranford

// Package observability provides metrics, tracing, and logging for the CA.
package observability

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// MetricType represents the type of metric.
type MetricType string

// Metric type constants.
const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
)

// Metric represents a single metric.
type Metric struct {
	Name        string            `json:"name"`
	Type        MetricType        `json:"type"`
	Description string            `json:"description"`
	Labels      map[string]string `json:"labels,omitempty"`
	Value       float64           `json:"value"`
	Timestamp   time.Time         `json:"timestamp"`
}

// MetricsRegistry holds all CA metrics.
type MetricsRegistry struct {
	counters   map[string]*atomic.Int64
	gauges     map[string]*atomic.Int64
	histograms map[string]*Histogram
	mu         sync.RWMutex
}

// Histogram tracks distribution of values.
type Histogram struct {
	buckets []histogramBucket
	count   atomic.Int64
	sum     atomic.Int64
}

type histogramBucket struct {
	upperBound float64
	count      atomic.Int64
}

// NewMetricsRegistry creates a new metrics registry.
func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{
		counters:   make(map[string]*atomic.Int64),
		gauges:     make(map[string]*atomic.Int64),
		histograms: make(map[string]*Histogram),
	}
}

// Counter operations.

// IncrementCounter increments a counter by 1.
func (r *MetricsRegistry) IncrementCounter(name string) {
	r.AddToCounter(name, 1)
}

// AddToCounter adds a value to a counter.
func (r *MetricsRegistry) AddToCounter(name string, delta int64) {
	r.mu.RLock()
	counter, exists := r.counters[name]
	r.mu.RUnlock()

	if !exists {
		r.mu.Lock()

		if counter, exists = r.counters[name]; !exists {
			counter = &atomic.Int64{}
			r.counters[name] = counter
		}

		r.mu.Unlock()
	}

	counter.Add(delta)
}

// GetCounter returns the current value of a counter.
func (r *MetricsRegistry) GetCounter(name string) int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if counter, exists := r.counters[name]; exists {
		return counter.Load()
	}

	return 0
}

// Gauge operations.

// SetGauge sets a gauge value.
func (r *MetricsRegistry) SetGauge(name string, value int64) {
	r.mu.RLock()
	gauge, exists := r.gauges[name]
	r.mu.RUnlock()

	if !exists {
		r.mu.Lock()

		if gauge, exists = r.gauges[name]; !exists {
			gauge = &atomic.Int64{}
			r.gauges[name] = gauge
		}

		r.mu.Unlock()
	}

	gauge.Store(value)
}

// GetGauge returns the current value of a gauge.
func (r *MetricsRegistry) GetGauge(name string) int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if gauge, exists := r.gauges[name]; exists {
		return gauge.Load()
	}

	return 0
}

// IncrementGauge increments a gauge by 1.
func (r *MetricsRegistry) IncrementGauge(name string) {
	r.mu.RLock()
	gauge, exists := r.gauges[name]
	r.mu.RUnlock()

	if !exists {
		r.mu.Lock()

		if gauge, exists = r.gauges[name]; !exists {
			gauge = &atomic.Int64{}
			r.gauges[name] = gauge
		}

		r.mu.Unlock()
	}

	gauge.Add(1)
}

// DecrementGauge decrements a gauge by 1.
func (r *MetricsRegistry) DecrementGauge(name string) {
	r.mu.RLock()
	gauge, exists := r.gauges[name]
	r.mu.RUnlock()

	if !exists {
		r.mu.Lock()

		if gauge, exists = r.gauges[name]; !exists {
			gauge = &atomic.Int64{}
			r.gauges[name] = gauge
		}

		r.mu.Unlock()
	}

	gauge.Add(-1)
}

// Histogram operations.

// Default histogram buckets (in milliseconds).
var defaultBuckets = []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000}

// NewHistogram creates a new histogram with default buckets.
func NewHistogram() *Histogram {
	return NewHistogramWithBuckets(defaultBuckets)
}

// NewHistogramWithBuckets creates a histogram with custom buckets.
func NewHistogramWithBuckets(buckets []float64) *Histogram {
	h := &Histogram{
		buckets: make([]histogramBucket, len(buckets)),
	}

	for i, b := range buckets {
		h.buckets[i] = histogramBucket{upperBound: b}
	}

	return h
}

// Observe records a value in the histogram.
func (h *Histogram) Observe(value float64) {
	h.count.Add(1)
	h.sum.Add(int64(value * histogramPrecision))

	for i := range h.buckets {
		if value <= h.buckets[i].upperBound {
			h.buckets[i].count.Add(1)
		}
	}
}

// Count returns the total number of observations.
func (h *Histogram) Count() int64 {
	return h.count.Load()
}

// Sum returns the sum of all observations.
func (h *Histogram) Sum() float64 {
	return float64(h.sum.Load()) / histogramPrecision
}

// GetBucketCounts returns the counts for each bucket.
func (h *Histogram) GetBucketCounts() map[float64]int64 {
	result := make(map[float64]int64)

	for i := range h.buckets {
		result[h.buckets[i].upperBound] = h.buckets[i].count.Load()
	}

	return result
}

// histogramPrecision for converting float64 to int64.
const histogramPrecision = 1000.0

// RecordHistogram records a value in a named histogram.
func (r *MetricsRegistry) RecordHistogram(name string, value float64) {
	r.mu.RLock()
	histogram, exists := r.histograms[name]
	r.mu.RUnlock()

	if !exists {
		r.mu.Lock()

		if histogram, exists = r.histograms[name]; !exists {
			histogram = NewHistogram()
			r.histograms[name] = histogram
		}

		r.mu.Unlock()
	}

	histogram.Observe(value)
}

// GetHistogram returns a histogram by name.
func (r *MetricsRegistry) GetHistogram(name string) *Histogram {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.histograms[name]
}

// GetAllMetrics returns all metrics as a slice.
func (r *MetricsRegistry) GetAllMetrics() []Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now().UTC()

	metrics := make([]Metric, 0, len(r.counters)+len(r.gauges)+len(r.histograms)*2)

	for name, counter := range r.counters {
		metrics = append(metrics, Metric{
			Name:      name,
			Type:      MetricTypeCounter,
			Value:     float64(counter.Load()),
			Timestamp: now,
		})
	}

	for name, gauge := range r.gauges {
		metrics = append(metrics, Metric{
			Name:      name,
			Type:      MetricTypeGauge,
			Value:     float64(gauge.Load()),
			Timestamp: now,
		})
	}

	for name, histogram := range r.histograms {
		metrics = append(metrics, Metric{
			Name:      name + "_count",
			Type:      MetricTypeHistogram,
			Value:     float64(histogram.Count()),
			Timestamp: now,
		})

		metrics = append(metrics, Metric{
			Name:      name + "_sum",
			Type:      MetricTypeHistogram,
			Value:     histogram.Sum(),
			Timestamp: now,
		})
	}

	return metrics
}

// Reset clears all metrics.
func (r *MetricsRegistry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.counters = make(map[string]*atomic.Int64)
	r.gauges = make(map[string]*atomic.Int64)
	r.histograms = make(map[string]*Histogram)
}

// Standard CA metric names.
const (
	MetricCertificatesIssued   = "ca_certificates_issued_total"
	MetricCertificatesRevoked  = "ca_certificates_revoked_total"
	MetricCertificatesExpired  = "ca_certificates_expired_total"
	MetricCertificatesActive   = "ca_certificates_active"
	MetricCRLGenerations       = "ca_crl_generations_total"
	MetricOCSPRequests         = "ca_ocsp_requests_total"
	MetricOCSPResponseTime     = "ca_ocsp_response_time_ms"
	MetricEnrollmentRequests   = "ca_enrollment_requests_total"
	MetricEnrollmentApprovals  = "ca_enrollment_approvals_total"
	MetricEnrollmentRejections = "ca_enrollment_rejections_total"
	MetricKeyGenerations       = "ca_key_generations_total"
	MetricSigningOperations    = "ca_signing_operations_total"
	MetricSigningOperationTime = "ca_signing_operation_time_ms"
	MetricValidationErrors     = "ca_validation_errors_total"
	MetricTimestampsIssued     = "ca_timestamps_issued_total"
)

// CAMetrics provides pre-configured CA metrics.
type CAMetrics struct {
	registry *MetricsRegistry
}

// NewCAMetrics creates a new CA metrics instance.
func NewCAMetrics() *CAMetrics {
	return &CAMetrics{
		registry: NewMetricsRegistry(),
	}
}

// Registry returns the underlying metrics registry.
func (m *CAMetrics) Registry() *MetricsRegistry {
	return m.registry
}

// RecordCertificateIssued records a certificate issuance.
func (m *CAMetrics) RecordCertificateIssued() {
	m.registry.IncrementCounter(MetricCertificatesIssued)
	m.registry.IncrementGauge(MetricCertificatesActive)
}

// RecordCertificateRevoked records a certificate revocation.
func (m *CAMetrics) RecordCertificateRevoked() {
	m.registry.IncrementCounter(MetricCertificatesRevoked)
	m.registry.DecrementGauge(MetricCertificatesActive)
}

// RecordCertificateExpired records a certificate expiration.
func (m *CAMetrics) RecordCertificateExpired() {
	m.registry.IncrementCounter(MetricCertificatesExpired)
	m.registry.DecrementGauge(MetricCertificatesActive)
}

// RecordCRLGeneration records a CRL generation.
func (m *CAMetrics) RecordCRLGeneration() {
	m.registry.IncrementCounter(MetricCRLGenerations)
}

// RecordOCSPRequest records an OCSP request with response time.
func (m *CAMetrics) RecordOCSPRequest(responseTimeMs float64) {
	m.registry.IncrementCounter(MetricOCSPRequests)
	m.registry.RecordHistogram(MetricOCSPResponseTime, responseTimeMs)
}

// RecordEnrollmentRequest records an enrollment request.
func (m *CAMetrics) RecordEnrollmentRequest() {
	m.registry.IncrementCounter(MetricEnrollmentRequests)
}

// RecordEnrollmentApproval records an enrollment approval.
func (m *CAMetrics) RecordEnrollmentApproval() {
	m.registry.IncrementCounter(MetricEnrollmentApprovals)
}

// RecordEnrollmentRejection records an enrollment rejection.
func (m *CAMetrics) RecordEnrollmentRejection() {
	m.registry.IncrementCounter(MetricEnrollmentRejections)
}

// RecordKeyGeneration records a key generation.
func (m *CAMetrics) RecordKeyGeneration() {
	m.registry.IncrementCounter(MetricKeyGenerations)
}

// RecordSigningOperation records a signing operation with duration.
func (m *CAMetrics) RecordSigningOperation(durationMs float64) {
	m.registry.IncrementCounter(MetricSigningOperations)
	m.registry.RecordHistogram(MetricSigningOperationTime, durationMs)
}

// RecordValidationError records a validation error.
func (m *CAMetrics) RecordValidationError() {
	m.registry.IncrementCounter(MetricValidationErrors)
}

// RecordTimestampIssued records a timestamp issuance.
func (m *CAMetrics) RecordTimestampIssued() {
	m.registry.IncrementCounter(MetricTimestampsIssued)
}

// Span represents a tracing span.
type Span struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	ParentID   string            `json:"parent_id,omitempty"`
	Name       string            `json:"name"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time,omitempty"`
	Duration   time.Duration     `json:"duration,omitempty"`
	Status     SpanStatus        `json:"status"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Events     []SpanEvent       `json:"events,omitempty"`
}

// SpanStatus represents the status of a span.
type SpanStatus string

// Span status constants.
const (
	SpanStatusUnset SpanStatus = "unset"
	SpanStatusOK    SpanStatus = "ok"
	SpanStatusError SpanStatus = "error"
)

// SpanEvent represents an event within a span.
type SpanEvent struct {
	Name       string            `json:"name"`
	Timestamp  time.Time         `json:"timestamp"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// Tracer provides tracing functionality.
type Tracer struct {
	spans map[string]*Span
	mu    sync.RWMutex
}

// NewTracer creates a new tracer.
func NewTracer() *Tracer {
	return &Tracer{
		spans: make(map[string]*Span),
	}
}

// StartSpan starts a new span.
func (t *Tracer) StartSpan(_ context.Context, name string) *Span {
	span := &Span{
		TraceID:    generateID(),
		SpanID:     generateID(),
		Name:       name,
		StartTime:  time.Now().UTC(),
		Status:     SpanStatusUnset,
		Attributes: make(map[string]string),
		Events:     []SpanEvent{},
	}

	t.mu.Lock()
	t.spans[span.SpanID] = span
	t.mu.Unlock()

	return span
}

// EndSpan ends a span.
func (t *Tracer) EndSpan(span *Span, status SpanStatus) {
	if span == nil {
		return
	}

	span.EndTime = time.Now().UTC()
	span.Duration = span.EndTime.Sub(span.StartTime)
	span.Status = status
}

// AddAttribute adds an attribute to a span.
func (t *Tracer) AddAttribute(span *Span, key, value string) {
	if span == nil {
		return
	}

	span.Attributes[key] = value
}

// AddEvent adds an event to a span.
func (t *Tracer) AddEvent(span *Span, name string, attributes map[string]string) {
	if span == nil {
		return
	}

	event := SpanEvent{
		Name:       name,
		Timestamp:  time.Now().UTC(),
		Attributes: attributes,
	}

	span.Events = append(span.Events, event)
}

// GetSpan retrieves a span by ID.
func (t *Tracer) GetSpan(spanID string) *Span {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.spans[spanID]
}

// generateID generates a random hex ID.
func generateID() string {
	// Simple implementation - in production use proper trace ID generation.
	return time.Now().UTC().Format("20060102150405.000000000")
}

// AuditEvent represents a CA audit event.
type AuditEvent struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	EventType AuditEventType    `json:"event_type"`
	Actor     string            `json:"actor"`
	Action    string            `json:"action"`
	Resource  string            `json:"resource"`
	Outcome   AuditOutcome      `json:"outcome"`
	Details   map[string]string `json:"details,omitempty"`
	IPAddress string            `json:"ip_address,omitempty"`
	UserAgent string            `json:"user_agent,omitempty"`
}

// AuditEventType represents the type of audit event.
type AuditEventType string

// Audit event type constants.
const (
	AuditTypeAuthentication AuditEventType = "authentication"
	AuditTypeAuthorization  AuditEventType = "authorization"
	AuditTypeCertificate    AuditEventType = "certificate"
	AuditTypeKey            AuditEventType = "key"
	AuditTypeConfiguration  AuditEventType = "configuration"
	AuditTypeAdministration AuditEventType = "administration"
)

// AuditOutcome represents the outcome of an audit event.
type AuditOutcome string

// Audit outcome constants.
const (
	AuditOutcomeSuccess AuditOutcome = "success"
	AuditOutcomeFailure AuditOutcome = "failure"
	AuditOutcomeDenied  AuditOutcome = "denied"
)

// AuditLogger provides audit logging functionality.
type AuditLogger struct {
	events []AuditEvent
	mu     sync.RWMutex
}

// NewAuditLogger creates a new audit logger.
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		events: []AuditEvent{},
	}
}

// Log records an audit event.
func (l *AuditLogger) Log(event AuditEvent) {
	event.ID = generateID()
	event.Timestamp = time.Now().UTC()

	l.mu.Lock()
	l.events = append(l.events, event)
	l.mu.Unlock()
}

// GetEvents returns all audit events.
func (l *AuditLogger) GetEvents() []AuditEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]AuditEvent, len(l.events))
	copy(result, l.events)

	return result
}

// GetEventsByType returns events filtered by type.
func (l *AuditLogger) GetEventsByType(eventType AuditEventType) []AuditEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var result []AuditEvent

	for _, event := range l.events {
		if event.EventType == eventType {
			result = append(result, event)
		}
	}

	return result
}

// Clear clears all audit events.
func (l *AuditLogger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.events = []AuditEvent{}
}
