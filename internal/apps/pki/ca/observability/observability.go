// Copyright (c) 2025 Justin Cranford

// Package observability provides metrics, tracing, and logging for the CA.
package observability

import (
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
