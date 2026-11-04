// Copyright (c) 2025-2026 Justin Cranford.
// Package readiness provides bounded background observations for readiness probes.
package readiness

import (
	"context"
	crand "crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"sort"
	"sync"
	"time"
)

const (
	defaultPeriod       = 2 * time.Second
	defaultCheckTimeout = 500 * time.Millisecond
	defaultMaxStaleness = 6 * time.Second
	jitterWindowDivisor = 10
)

var (
	// ErrInitialChecksFailed indicates that one or more required observations failed at startup.
	ErrInitialChecksFailed = errors.New("initial readiness checks failed")
	safeNamePattern        = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
)

// CheckFunc performs one bounded, read-only dependency observation.
type CheckFunc func(context.Context) error

// Check describes a required health observation. ReasonCode must be a sanitized,
// stable identifier; the underlying error is never exposed by the monitor.
type Check struct {
	Name       string
	ReasonCode string
	Run        CheckFunc
}

// Config controls polling cadence and fail-closed freshness.
type Config struct {
	Checks       []Check
	Period       time.Duration
	CheckTimeout time.Duration
	MaxStaleness time.Duration
	Now          func() time.Time
	Jitter       func(time.Duration) time.Duration
}

// CheckObservation contains only sanitized health state and timing information.
type CheckObservation struct {
	Name       string
	ReasonCode string
	Healthy    bool
	Stale      bool
	Age        time.Duration
	ObservedAt time.Time
}

// Snapshot is the immutable aggregate state used by readiness handlers.
type Snapshot struct {
	Ready      bool
	ObservedAt time.Time
	Checks     []CheckObservation
}

// Monitor runs configured checks outside of probe handlers and stores bounded-staleness results.
type Monitor struct {
	config Config

	mu          sync.RWMutex
	snapshot    Snapshot
	started     bool
	initialized bool
	initialDone chan struct{}
	done        chan struct{}
	cancel      context.CancelFunc
}

// New validates check metadata and applies the contract's default observation bounds.
func New(config Config) (*Monitor, error) {
	if len(config.Checks) == 0 {
		return nil, errors.New("at least one readiness check is required")
	}

	if config.Period == 0 {
		config.Period = defaultPeriod
	}

	if config.CheckTimeout == 0 {
		config.CheckTimeout = defaultCheckTimeout
	}

	if config.MaxStaleness == 0 {
		config.MaxStaleness = defaultMaxStaleness
	}

	if config.Period <= 0 || config.CheckTimeout <= 0 || config.MaxStaleness <= 0 {
		return nil, errors.New("readiness durations must be positive")
	}

	if config.Now == nil {
		config.Now = time.Now
	}

	if config.Jitter == nil {
		config.Jitter = jitteredInterval
	}

	names := make(map[string]struct{}, len(config.Checks))
	for _, check := range config.Checks {
		if !safeNamePattern.MatchString(check.Name) || !safeNamePattern.MatchString(check.ReasonCode) {
			return nil, errors.New("readiness check names and reason codes must be sanitized identifiers")
		}

		if check.Run == nil {
			return nil, fmt.Errorf("readiness check %q has no function", check.Name)
		}

		if _, exists := names[check.Name]; exists {
			return nil, fmt.Errorf("duplicate readiness check %q", check.Name)
		}

		names[check.Name] = struct{}{}
	}

	checks := append([]Check(nil), config.Checks...)
	sort.Slice(checks, func(i, j int) bool { return checks[i].Name < checks[j].Name })
	config.Checks = checks

	return &Monitor{
		config:      config,
		initialDone: make(chan struct{}),
		done:        make(chan struct{}),
	}, nil
}

// Start begins immediate observations followed by jittered periodic observations.
func (m *Monitor) Start(parent context.Context) error {
	if parent == nil {
		return errors.New("context cannot be nil")
	}

	m.mu.Lock()
	if m.started {
		m.mu.Unlock()

		return errors.New("readiness monitor already started")
	}

	ctx, cancel := context.WithCancel(parent)
	m.cancel = cancel
	m.started = true
	m.mu.Unlock()

	go m.run(ctx)

	return nil
}

// WaitInitial waits until every required check has been observed once.
// Failed checks return sanitized reason codes rather than raw dependency errors.
func (m *Monitor) WaitInitial(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}

	select {
	case <-m.initialDone:
		snapshot := m.Snapshot()
		if snapshot.Ready {
			return nil
		}

		return fmt.Errorf("%w: %v", ErrInitialChecksFailed, failedReadinessReasons(snapshot))
	case <-ctx.Done():
		return fmt.Errorf("waiting for initial readiness checks: %w", ctx.Err())
	}
}

// WaitReady waits for a healthy, fresh aggregate rather than treating a transient
// initial observation failure as fatal. The caller must bound startup with a
// context deadline; each observation retains the configured probe deadline.
func (m *Monitor) WaitReady(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context cannot be nil")
	}

	ticker := time.NewTicker(m.config.Period)
	defer ticker.Stop()
	initialDone := m.initialDone

	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("waiting for successful readiness observations: %w: %w: %v",
				ErrInitialChecksFailed, err, failedReadinessReasons(m.Snapshot()))
		}

		if m.Snapshot().Ready {
			return nil
		}

		select {
		case <-ctx.Done():
		case <-initialDone:
			initialDone = nil
		case <-ticker.C:
		}
	}
}

func failedReadinessReasons(snapshot Snapshot) []string {
	reasons := make([]string, 0, len(snapshot.Checks))
	for _, check := range snapshot.Checks {
		if !check.Healthy || check.Stale {
			reasons = append(reasons, check.ReasonCode)
		}
	}

	return reasons
}

// Snapshot returns a copy of the latest observations, failing closed when any result is stale.
func (m *Monitor) Snapshot() Snapshot {
	m.mu.RLock()
	snapshot := Snapshot{
		ObservedAt: m.snapshot.ObservedAt,
		Checks:     append([]CheckObservation(nil), m.snapshot.Checks...),
	}
	initialized := m.initialized
	m.mu.RUnlock()

	now := m.config.Now()

	snapshot.Ready = initialized && len(snapshot.Checks) == len(m.config.Checks)
	for i := range snapshot.Checks {
		age := now.Sub(snapshot.Checks[i].ObservedAt)
		snapshot.Checks[i].Age = age

		snapshot.Checks[i].Stale = age < 0 || age > m.config.MaxStaleness
		if !snapshot.Checks[i].Healthy || snapshot.Checks[i].Stale {
			snapshot.Ready = false
		}
	}

	return snapshot
}

// Stop cancels polling and waits for the monitor goroutine to exit.
func (m *Monitor) Stop(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	m.mu.RLock()
	started := m.started
	cancel := m.cancel
	m.mu.RUnlock()

	if !started {
		return nil
	}

	cancel()

	select {
	case <-m.done:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("stopping readiness monitor: %w", ctx.Err())
	}
}

func (m *Monitor) run(ctx context.Context) {
	defer close(m.done)

	for {
		m.observe(ctx)

		m.mu.Lock()
		if !m.initialized {
			m.initialized = true
			close(m.initialDone)
		}
		m.mu.Unlock()

		interval := m.config.Jitter(m.config.Period)
		if interval <= 0 {
			interval = time.Millisecond
		}

		timer := time.NewTimer(interval)
		select {
		case <-timer.C:
		case <-ctx.Done():
			stopAndDrainTimer(timer)

			return
		}
	}
}

func (m *Monitor) observe(parent context.Context) {
	type result struct {
		observation CheckObservation
	}

	results := make(chan result, len(m.config.Checks))

	var checks sync.WaitGroup

	for _, check := range m.config.Checks {
		checks.Add(1)

		go func() {
			defer checks.Done()

			ctx, cancel := context.WithTimeout(parent, m.config.CheckTimeout)
			err := check.Run(ctx)

			cancel()

			results <- result{observation: CheckObservation{
				Name:       check.Name,
				ReasonCode: check.ReasonCode,
				Healthy:    err == nil,
				ObservedAt: m.config.Now(),
			}}
		}()
	}

	checks.Wait()
	close(results)

	snapshot := Snapshot{
		ObservedAt: m.config.Now(),
		Checks:     make([]CheckObservation, 0, len(m.config.Checks)),
	}

	snapshot.Ready = true
	for result := range results {
		snapshot.Checks = append(snapshot.Checks, result.observation)
		if !result.observation.Healthy {
			snapshot.Ready = false
		}
	}

	sort.Slice(snapshot.Checks, func(i, j int) bool {
		return snapshot.Checks[i].Name < snapshot.Checks[j].Name
	})

	m.mu.Lock()
	m.snapshot = snapshot
	m.mu.Unlock()
}

func stopAndDrainTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}

func jitteredInterval(period time.Duration) time.Duration {
	return jitteredIntervalWithReader(period, crand.Reader)
}

func jitteredIntervalWithReader(period time.Duration, randomReader io.Reader) time.Duration {
	window := period / jitterWindowDivisor
	if window <= 0 {
		return period
	}

	randomRange := big.NewInt(int64(2*window + 1))

	value, err := crand.Int(randomReader, randomRange)
	if err != nil {
		return period
	}

	return period - window + time.Duration(value.Int64())
}
