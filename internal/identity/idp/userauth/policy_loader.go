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
type PolicyLoader interface {
	// LoadRiskScoringPolicy loads risk scoring configuration.
	LoadRiskScoringPolicy(ctx context.Context) (*RiskScoringPolicy, error)

	// LoadStepUpPolicies loads step-up authentication policies.
	LoadStepUpPolicies(ctx context.Context) (*StepUpPolicies, error)

	// LoadAdaptiveAuthPolicy loads adaptive authentication policy.
	LoadAdaptiveAuthPolicy(ctx context.Context) (*AdaptiveAuthPolicy, error)

	// EnableHotReload enables automatic policy reload on file changes.
	EnableHotReload(ctx context.Context, interval time.Duration) error

	// DisableHotReload stops automatic policy reload.
	DisableHotReload()
}

// RiskScoringPolicy represents risk scoring configuration loaded from risk_scoring.yml.
type RiskScoringPolicy struct {
	Version string `yaml:"version"`

	// Risk factor weights (must sum to 1.0).
	RiskFactors map[string]RiskFactorConfig `yaml:"risk_factors"`

	// Risk thresholds defining categorization.
	RiskThresholds map[string]RiskThreshold `yaml:"risk_thresholds"`

	// Confidence scoring weights.
	ConfidenceWeights ConfidenceWeights `yaml:"confidence_weights"`

	// Network risk scores.
	NetworkRisks map[string]NetworkRisk `yaml:"network_risks"`

	// Geographic risk scores.
	GeographicRisks GeographicRisks `yaml:"geographic_risks"`

	// Velocity limits.
	VelocityLimits map[string]VelocityLimit `yaml:"velocity_limits"`

	// Time-based risk scores.
	TimeRisks map[string]TimeRisk `yaml:"time_risks"`

	// Behavior-based risk scores.
	BehaviorRisks map[string]BehaviorRisk `yaml:"behavior_risks"`
}

// RiskFactorConfig represents weight and description for a risk factor from policy.
type RiskFactorConfig struct {
	Weight      float64 `yaml:"weight"`
	Description string  `yaml:"description"`
}

// RiskThreshold represents authentication requirements for a risk level.
type RiskThreshold struct {
	Min                 float64  `yaml:"min"`
	Max                 float64  `yaml:"max"`
	AuthRequirements    []string `yaml:"auth_requirements"`
	MaxSessionDuration  string   `yaml:"max_session_duration"`
	StepUpRequired      bool     `yaml:"step_up_required,omitempty"`
	AlertSecurityTeam   bool     `yaml:"alert_security_team,omitempty"`
	BlockAuthentication bool     `yaml:"block_authentication,omitempty"`
	Description         string   `yaml:"description"`
}

// ConfidenceWeights represents weights for confidence scoring components.
type ConfidenceWeights struct {
	FactorCount     float64 `yaml:"factor_count"`
	BaselineData    float64 `yaml:"baseline_data"`
	BehaviorProfile float64 `yaml:"behavior_profile"`
	Description     string  `yaml:"description"`
}

// NetworkRisk represents risk score for network type.
type NetworkRisk struct {
	Score       float64 `yaml:"score"`
	Description string  `yaml:"description"`
}

// GeographicRisks represents risk scores for geographic locations.
type GeographicRisks struct {
	HighRiskCountries  HighRiskCountries  `yaml:"high_risk_countries"`
	EmbargoedCountries EmbargoedCountries `yaml:"embargoed_countries"`
}

// HighRiskCountries represents countries with elevated risk.
type HighRiskCountries struct {
	Countries   []string `yaml:"countries"`
	Score       float64  `yaml:"score"`
	Description string   `yaml:"description"`
}

// EmbargoedCountries represents countries under embargo.
type EmbargoedCountries struct {
	Countries   []string `yaml:"countries"`
	Score       float64  `yaml:"score"`
	Description string   `yaml:"description"`
}

// VelocityLimit represents threshold for velocity-based risk.
type VelocityLimit struct {
	Window       string  `yaml:"window"`
	MaxAttempts  int     `yaml:"max_attempts,omitempty"`
	MaxLocations int     `yaml:"max_locations,omitempty"`
	MaxDevices   int     `yaml:"max_devices,omitempty"`
	RiskScore    float64 `yaml:"risk_score"`
	Description  string  `yaml:"description"`
}

// TimeRisk represents risk score for time-based anomalies.
type TimeRisk struct {
	Score       float64 `yaml:"score"`
	Description string  `yaml:"description"`
}

// BehaviorRisk represents risk score for behavior patterns.
type BehaviorRisk struct {
	Score       float64 `yaml:"score"`
	Description string  `yaml:"description"`
}

// StepUpPolicies represents step-up authentication policies loaded from step_up.yml.
type StepUpPolicies struct {
	Version string `yaml:"version"`

	// Operation-specific policies.
	Policies map[string]OperationPolicy `yaml:"policies"`

	// Default policy for unlisted operations.
	DefaultPolicy OperationPolicy `yaml:"default_policy"`

	// Auth level definitions (documentation reference).
	AuthLevels map[string]int `yaml:"auth_levels"`

	// Step-up method configurations.
	StepUpMethods map[string]StepUpMethod `yaml:"step_up_methods"`

	// Session durations by auth level.
	SessionDurations map[string]string `yaml:"session_durations"`

	// Monitoring thresholds.
	Monitoring MonitoringThresholds `yaml:"monitoring"`
}

// OperationPolicy represents policy for specific operation.
type OperationPolicy struct {
	OperationPattern   string            `yaml:"operation_pattern,omitempty"`
	RequiredLevel      string            `yaml:"required_level"`
	AllowedMethods     []string          `yaml:"allowed_methods"`
	MaxAge             string            `yaml:"max_age"`
	RiskLevelOverrides map[string]string `yaml:"risk_level_overrides,omitempty"`
	Description        string            `yaml:"description,omitempty"`
}

// StepUpMethod represents configuration for step-up method.
type StepUpMethod struct {
	Strength         string `yaml:"strength"`
	FallbackPriority int    `yaml:"fallback_priority"`
	Description      string `yaml:"description"`
}

// MonitoringThresholds represents thresholds for monitoring step-up behavior.
type MonitoringThresholds struct {
	StepUpRate        string `yaml:"step_up_rate"`
	BlockedOperations string `yaml:"blocked_operations"`
	FallbackMethods   string `yaml:"fallback_methods"`
	Description       string `yaml:"description"`
}

// AdaptiveAuthPolicy represents adaptive authentication policy loaded from adaptive_auth.yml.
type AdaptiveAuthPolicy struct {
	Version     string `yaml:"version"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`

	// Risk-based authentication requirements.
	RiskBasedAuth map[string]RiskBasedAuthRequirement `yaml:"risk_based_auth"`

	// Fallback policy.
	FallbackPolicy FallbackPolicy `yaml:"fallback_policy"`

	// Grace periods.
	GracePeriods map[string]GracePeriod `yaml:"grace_periods"`

	// Device trust settings.
	DeviceTrust DeviceTrust `yaml:"device_trust"`

	// Location trust settings.
	LocationTrust LocationTrust `yaml:"location_trust"`

	// Behavior trust settings.
	BehaviorTrust BehaviorTrust `yaml:"behavior_trust"`

	// Tuning parameters.
	Tuning TuningParameters `yaml:"tuning"`
}

// RiskBasedAuthRequirement represents authentication requirements for risk level.
type RiskBasedAuthRequirement struct {
	RiskScoreRange             RiskScoreRange   `yaml:"risk_score_range"`
	RequiredMethods            []string         `yaml:"required_methods"`
	OptionalMethods            []string         `yaml:"optional_methods,omitempty"`
	MFAMethodsAllowed          []string         `yaml:"mfa_methods_allowed,omitempty"`
	StrongMFAMethodsAllowed    []string         `yaml:"strong_mfa_methods_allowed,omitempty"`
	FallbackMethodsAllowed     []string         `yaml:"fallback_methods_allowed,omitempty"`
	SessionDuration            string           `yaml:"session_duration"`
	IdleTimeout                string           `yaml:"idle_timeout"`
	StepUpRequired             bool             `yaml:"step_up_required"`
	AllowNewDeviceRegistration bool             `yaml:"allow_new_device_registration"`
	AllowPasswordReset         bool             `yaml:"allow_password_reset"`
	AdditionalChecks           []string         `yaml:"additional_checks,omitempty"`
	Monitoring                 MonitoringConfig `yaml:"monitoring"`
	Description                string           `yaml:"description"`
}

// RiskScoreRange represents min/max risk score range.
type RiskScoreRange struct {
	Min float64 `yaml:"min"`
	Max float64 `yaml:"max"`
}

// MonitoringConfig represents monitoring configuration for risk level.
type MonitoringConfig struct {
	LogLevel              string `yaml:"log_level"`
	AlertOnFailure        bool   `yaml:"alert_on_failure"`
	AlertSecurityTeam     bool   `yaml:"alert_security_team,omitempty"`
	AlertFraudTeam        bool   `yaml:"alert_fraud_team,omitempty"`
	TrackLocation         bool   `yaml:"track_location,omitempty"`
	TrackDevice           bool   `yaml:"track_device,omitempty"`
	CaptureRequestDetails bool   `yaml:"capture_request_details,omitempty"`
	CaptureNetworkDetails bool   `yaml:"capture_network_details,omitempty"`
}

// FallbackPolicy represents fallback behavior when risk assessment fails.
type FallbackPolicy struct {
	OnError         string `yaml:"on_error"`
	OnLowConfidence string `yaml:"on_low_confidence"`
	Description     string `yaml:"description"`
}

// GracePeriod represents grace period for authentication transitions.
type GracePeriod struct {
	Duration          string `yaml:"duration"`
	RiskLevelOverride string `yaml:"risk_level_override"`
	Description       string `yaml:"description"`
}

// DeviceTrust represents device trust settings.
type DeviceTrust struct {
	RememberDeviceDuration   string   `yaml:"remember_device_duration"`
	MaxTrustedDevices        int      `yaml:"max_trusted_devices"`
	RequireReauthOnNewDevice bool     `yaml:"require_reauth_on_new_device"`
	DeviceFingerprintFactors []string `yaml:"device_fingerprint_factors"`
}

// LocationTrust represents location trust settings.
type LocationTrust struct {
	RememberLocationDuration  string   `yaml:"remember_location_duration"`
	ImpossibleTravelThreshold string   `yaml:"impossible_travel_threshold"`
	HighRiskCountriesBlock    bool     `yaml:"high_risk_countries_block"`
	LocationFactors           []string `yaml:"location_factors"`
}

// BehaviorTrust represents behavior trust settings.
type BehaviorTrust struct {
	BaselineEstablishmentPeriod string   `yaml:"baseline_establishment_period"`
	MinEventsForBaseline        int      `yaml:"min_events_for_baseline"`
	TrackedPatterns             []string `yaml:"tracked_patterns"`
}

// TuningParameters represents adaptive authentication tuning parameters.
type TuningParameters struct {
	RiskScoreDecayRate         float64 `yaml:"risk_score_decay_rate"`
	RiskScoreSpikeFactor       float64 `yaml:"risk_score_spike_factor"`
	ConfidenceThresholdLow     float64 `yaml:"confidence_threshold_low"`
	ConfidenceThresholdMedium  float64 `yaml:"confidence_threshold_medium"`
	ConfidenceThresholdHigh    float64 `yaml:"confidence_threshold_high"`
	BaselineStalenessThreshold string  `yaml:"baseline_staleness_threshold"`
}

// YAMLPolicyLoader implements PolicyLoader using YAML files.
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
func (l *YAMLPolicyLoader) LoadStepUpPolicies(ctx context.Context) (*StepUpPolicies, error) {
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
func (l *YAMLPolicyLoader) LoadAdaptiveAuthPolicy(ctx context.Context) (*AdaptiveAuthPolicy, error) {
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
