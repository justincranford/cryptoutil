// Copyright (c) 2025 Justin Cranford

package userauth

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// PolicyLoader defines interface for loading policy configurations from YAML.
type YAMLPolicyLoader struct {
	riskScoringPath    string
	stepUpPoliciesPath string
	adaptiveAuthPath   string

	// Cached policies with RWMutex for hot-reload support.
	mu                 sync.RWMutex
	riskScoringPolicy  *RiskScoringPolicy
	stepUpPolicies     *StepUpPolicies
	adaptiveAuthPolicy *AdaptiveAuthPolicy

	// Hot-reload management.
	hotReloadEnabled bool
	hotReloadCancel  context.CancelFunc
	hotReloadWg      sync.WaitGroup
}

// NewYAMLPolicyLoader creates new YAML policy loader.
func NewYAMLPolicyLoader(riskScoringPath, stepUpPoliciesPath, adaptiveAuthPath string) *YAMLPolicyLoader {
	return &YAMLPolicyLoader{
		riskScoringPath:    riskScoringPath,
		stepUpPoliciesPath: stepUpPoliciesPath,
		adaptiveAuthPath:   adaptiveAuthPath,
	}
}

// LoadRiskScoringPolicy loads risk scoring configuration from risk_scoring.yml.
func (l *YAMLPolicyLoader) LoadRiskScoringPolicy(_ context.Context) (*RiskScoringPolicy, error) {
	// Check cache first.
	l.mu.RLock()

	if l.riskScoringPolicy != nil {
		cached := l.riskScoringPolicy
		l.mu.RUnlock()

		return cached, nil
	}

	l.mu.RUnlock()

	// Load from file.
	data, err := os.ReadFile(l.riskScoringPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read risk scoring policy: %w", err)
	}

	var policy RiskScoringPolicy
	if err := yaml.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse risk scoring policy: %w", err)
	}

	// Validate policy.
	if err := validateRiskScoringPolicy(&policy); err != nil {
		return nil, fmt.Errorf("invalid risk scoring policy: %w", err)
	}

	// Cache policy.
	l.mu.Lock()
	l.riskScoringPolicy = &policy
	l.mu.Unlock()

	return &policy, nil
}

// LoadStepUpPolicies loads step-up authentication policies from step_up.yml.
func (l *YAMLPolicyLoader) LoadStepUpPolicies(_ context.Context) (*StepUpPolicies, error) {
	// Check cache first.
	l.mu.RLock()

	if l.stepUpPolicies != nil {
		cached := l.stepUpPolicies
		l.mu.RUnlock()

		return cached, nil
	}

	l.mu.RUnlock()

	// Load from file.
	data, err := os.ReadFile(l.stepUpPoliciesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read step-up policies: %w", err)
	}

	var policies StepUpPolicies
	if err := yaml.Unmarshal(data, &policies); err != nil {
		return nil, fmt.Errorf("failed to parse step-up policies: %w", err)
	}

	// Validate policies.
	if err := validateStepUpPolicies(&policies); err != nil {
		return nil, fmt.Errorf("invalid step-up policies: %w", err)
	}

	// Cache policies.
	l.mu.Lock()
	l.stepUpPolicies = &policies
	l.mu.Unlock()

	return &policies, nil
}

// LoadAdaptiveAuthPolicy loads adaptive authentication policy from adaptive_auth.yml.
func (l *YAMLPolicyLoader) LoadAdaptiveAuthPolicy(_ context.Context) (*AdaptiveAuthPolicy, error) {
	// Check cache first.
	l.mu.RLock()

	if l.adaptiveAuthPolicy != nil {
		cached := l.adaptiveAuthPolicy
		l.mu.RUnlock()

		return cached, nil
	}

	l.mu.RUnlock()

	// Load from file.
	data, err := os.ReadFile(l.adaptiveAuthPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read adaptive auth policy: %w", err)
	}

	var policy AdaptiveAuthPolicy
	if err := yaml.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse adaptive auth policy: %w", err)
	}

	// Validate policy.
	if err := validateAdaptiveAuthPolicy(&policy); err != nil {
		return nil, fmt.Errorf("invalid adaptive auth policy: %w", err)
	}

	// Cache policy.
	l.mu.Lock()
	l.adaptiveAuthPolicy = &policy
	l.mu.Unlock()

	return &policy, nil
}

// EnableHotReload enables automatic policy reload on file changes.
func (l *YAMLPolicyLoader) EnableHotReload(ctx context.Context, interval time.Duration) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.hotReloadEnabled {
		return fmt.Errorf("hot-reload already enabled")
	}

	hotReloadCtx, cancel := context.WithCancel(ctx)
	l.hotReloadCancel = cancel
	l.hotReloadEnabled = true

	l.hotReloadWg.Add(1)

	go l.hotReloadWorker(hotReloadCtx, interval)

	return nil
}

// DisableHotReload stops automatic policy reload.
func (l *YAMLPolicyLoader) DisableHotReload() {
	l.mu.Lock()

	if !l.hotReloadEnabled {
		l.mu.Unlock()

		return
	}

	l.hotReloadCancel()
	l.hotReloadEnabled = false
	l.mu.Unlock()

	l.hotReloadWg.Wait()
}

// hotReloadWorker periodically checks for policy file changes and reloads.
func (l *YAMLPolicyLoader) hotReloadWorker(ctx context.Context, interval time.Duration) {
	defer l.hotReloadWg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Invalidate caches to force reload on next access.
			l.mu.Lock()
			l.riskScoringPolicy = nil
			l.stepUpPolicies = nil
			l.adaptiveAuthPolicy = nil
			l.mu.Unlock()
		}
	}
}

// validateRiskScoringPolicy validates risk scoring policy.
func validateRiskScoringPolicy(policy *RiskScoringPolicy) error {
	// Validate risk factor weights sum to 1.0.
	var weightSum float64
	for _, factor := range policy.RiskFactors {
		weightSum += factor.Weight
	}

	const tolerance = 0.001
	if weightSum < (1.0-tolerance) || weightSum > (1.0+tolerance) {
		return fmt.Errorf("risk factor weights must sum to 1.0, got %.3f", weightSum)
	}

	// Validate risk thresholds are non-overlapping and cover [0.0, 1.0].
	// (Simplified validation - full implementation would check ranges).
	if len(policy.RiskThresholds) == 0 {
		return fmt.Errorf("risk thresholds cannot be empty")
	}

	return nil
}

// validateStepUpPolicies validates step-up policies.
func validateStepUpPolicies(policies *StepUpPolicies) error {
	// Validate default policy exists.
	if policies.DefaultPolicy.RequiredLevel == "" {
		return fmt.Errorf("default policy required_level cannot be empty")
	}

	// Validate operation policies.
	if len(policies.Policies) == 0 {
		return fmt.Errorf("policies cannot be empty")
	}

	return nil
}

// validateAdaptiveAuthPolicy validates adaptive authentication policy.
func validateAdaptiveAuthPolicy(policy *AdaptiveAuthPolicy) error {
	// Validate risk-based auth requirements exist.
	if len(policy.RiskBasedAuth) == 0 {
		return fmt.Errorf("risk_based_auth cannot be empty")
	}

	// Validate fallback policy.
	if policy.FallbackPolicy.OnError == "" {
		return fmt.Errorf("fallback_policy.on_error cannot be empty")
	}

	return nil
}
