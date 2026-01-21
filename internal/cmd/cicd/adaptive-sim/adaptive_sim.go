// Copyright (c) 2025 Justin Cranford

// Package main provides the adaptive simulator for performance testing.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	cryptoutilIdentityUserauth "cryptoutil/internal/identity/idp/userauth"
)

// Version information for the adaptive simulation CLI.
const (
	Version = "1.0.0"
)

const (
	exitSuccess = 0
	exitError   = 1
)

const (
	dirPerms755  = 0o755
	filePerms600 = 0o600
)

const (
	authLevelNone   = 0
	authLevelBasic  = 1
	authLevelMFA    = 2
	authLevelStepUp = 3
	authLevelBlock  = 4
)

const (
	decisionAllow  = "allow"
	decisionStepUp = "step_up"
	decisionBlock  = "block"
)

// AdaptiveSimulator simulates adaptive authentication policy changes against historical data.
type AdaptiveSimulator struct {
	policyLoader cryptoutilIdentityUserauth.PolicyLoader
	outputDir    string
}

// SimulationResult represents the outcome of policy simulation.
type SimulationResult struct {
	PolicyVersion     string             `json:"policy_version"`
	SimulationTime    time.Time          `json:"simulation_time"`
	TotalAttempts     int                `json:"total_attempts"`
	StepUpRequired    int                `json:"step_up_required"`
	BlockedOperations int                `json:"blocked_operations"`
	AllowedOperations int                `json:"allowed_operations"`
	StepUpRate        float64            `json:"step_up_rate"`
	BlockedRate       float64            `json:"blocked_rate"`
	RiskDistribution  map[string]int     `json:"risk_distribution"`
	PolicyEvaluations []PolicyEvaluation `json:"policy_evaluations"`
	Recommendations   []string           `json:"recommendations"`
}

// PolicyEvaluation represents a single policy evaluation.
type PolicyEvaluation struct {
	Timestamp      time.Time `json:"timestamp"`
	UserID         string    `json:"user_id"`
	Operation      string    `json:"operation"`
	RiskScore      float64   `json:"risk_score"`
	RiskLevel      string    `json:"risk_level"`
	Decision       string    `json:"decision"` // "allow", "step_up", "block".
	RequiredLevel  string    `json:"required_level"`
	CurrentLevel   string    `json:"current_level"`
	StepUpRequired bool      `json:"step_up_required"`
}

// HistoricalAuthLog represents authentication attempt from logs.
type HistoricalAuthLog struct {
	Timestamp        time.Time      `json:"timestamp"`
	UserID           string         `json:"user_id"`
	Operation        string         `json:"operation"`
	IPAddress        string         `json:"ip_address"`
	DeviceID         string         `json:"device_id"`
	Country          string         `json:"country"`
	City             string         `json:"city"`
	IsVPN            bool           `json:"is_vpn"`
	IsProxy          bool           `json:"is_proxy"`
	CurrentAuthLevel string         `json:"current_auth_level"`
	Success          bool           `json:"success"`
	Metadata         map[string]any `json:"metadata"`
}

func main() {
	os.Exit(internalMain(os.Args, os.Stdin, os.Stdout, os.Stderr))
}

// internalMain is the testable main function with injected dependencies.
func internalMain(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		riskScoringPath = fs.String("risk-scoring", "configs/identity/policies/risk_scoring.yml", "Path to risk scoring policy")
		stepUpPath      = fs.String("step-up", "configs/identity/policies/step_up.yml", "Path to step-up policy")
		adaptivePath    = fs.String("adaptive", "configs/identity/policies/adaptive_auth.yml", "Path to adaptive auth policy")
		historicalLogs  = fs.String("logs", "", "Path to historical authentication logs (JSON)")
		outputDir       = fs.String("output", "test-output/adaptive-sim", "Output directory for simulation results")
		policyVersion   = fs.String("version", "v1.0", "Policy version identifier")
	)

	if err := fs.Parse(args[1:]); err != nil {
		return exitError
	}

	if *historicalLogs == "" {
		_, _ = fmt.Fprintln(stderr, "Error: --logs flag is required")
		_, _ = fmt.Fprintln(stderr, "\nUsage: adaptive-sim --logs=auth_logs.json [options]")
		_, _ = fmt.Fprintln(stderr, "\nExample historical log format:")
		_, _ = fmt.Fprintln(stderr, exampleLogFormat)

		return exitError
	}

	// Create output directory.
	if err := os.MkdirAll(*outputDir, dirPerms755); err != nil {
		_, _ = fmt.Fprintf(stderr, "Failed to create output directory: %v\n", err)

		return exitError
	}

	// Create policy loader.
	loader := cryptoutilIdentityUserauth.NewYAMLPolicyLoader(*riskScoringPath, *stepUpPath, *adaptivePath)

	simulator := &AdaptiveSimulator{
		policyLoader: loader,
		outputDir:    *outputDir,
	}

	// Run simulation.
	ctx := context.Background()

	result, err := simulator.Simulate(ctx, *historicalLogs, *policyVersion)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "Simulation failed: %v\n", err)

		return exitError
	}

	// Save results.
	if err := simulator.SaveResults(result, stdout); err != nil {
		_, _ = fmt.Fprintf(stderr, "Failed to save results: %v\n", err)

		return exitError
	}

	// Print summary.
	simulator.PrintSummary(result, stdout)

	return exitSuccess
}

// Simulate runs policy simulation against historical logs.
func (s *AdaptiveSimulator) Simulate(ctx context.Context, logsPath, policyVersion string) (*SimulationResult, error) {
	// Load policies.
	riskPolicy, err := s.policyLoader.LoadRiskScoringPolicy(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load risk scoring policy: %w", err)
	}

	stepUpPolicy, err := s.policyLoader.LoadStepUpPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load step-up policies: %w", err)
	}

	adaptivePolicy, err := s.policyLoader.LoadAdaptiveAuthPolicy(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load adaptive auth policy: %w", err)
	}

	// Load historical logs.
	logs, err := s.LoadHistoricalLogs(logsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load historical logs: %w", err)
	}

	// Initialize result.
	result := &SimulationResult{
		PolicyVersion:     policyVersion,
		SimulationTime:    time.Now(),
		TotalAttempts:     len(logs),
		RiskDistribution:  make(map[string]int),
		PolicyEvaluations: make([]PolicyEvaluation, 0, len(logs)),
	}

	// Simulate each auth attempt.
	for _, log := range logs {
		eval := s.EvaluateAuthAttempt(log, riskPolicy, stepUpPolicy, adaptivePolicy)
		result.PolicyEvaluations = append(result.PolicyEvaluations, eval)

		// Update counters.
		result.RiskDistribution[eval.RiskLevel]++

		switch eval.Decision {
		case decisionAllow:
			result.AllowedOperations++
		case decisionStepUp:
			result.StepUpRequired++
		case decisionBlock:
			result.BlockedOperations++
		}
	}

	// Calculate rates.
	if result.TotalAttempts > 0 {
		result.StepUpRate = float64(result.StepUpRequired) / float64(result.TotalAttempts)
		result.BlockedRate = float64(result.BlockedOperations) / float64(result.TotalAttempts)
	}

	// Generate recommendations.
	result.Recommendations = s.GenerateRecommendations(result)

	return result, nil
}

// EvaluateAuthAttempt simulates policy evaluation for single auth attempt.
func (s *AdaptiveSimulator) EvaluateAuthAttempt(
	log HistoricalAuthLog,
	riskPolicy *cryptoutilIdentityUserauth.RiskScoringPolicy,
	stepUpPolicy *cryptoutilIdentityUserauth.StepUpPolicies,
	adaptivePolicy *cryptoutilIdentityUserauth.AdaptiveAuthPolicy, //nolint:revive
) PolicyEvaluation {
	// Simplified risk scoring (real implementation would use full BehavioralRiskEngine).
	riskScore := s.CalculateRiskScore(log, riskPolicy)

	riskLevel := s.DetermineRiskLevel(riskScore, riskPolicy)

	// Determine required auth level for operation.
	requiredLevel, stepUpNeeded := s.DetermineRequiredLevel(log.Operation, log.CurrentAuthLevel, stepUpPolicy)

	// Make decision based on risk and policy.
	decision := s.MakeDecision(riskLevel, stepUpNeeded, adaptivePolicy)

	return PolicyEvaluation{
		Timestamp:      log.Timestamp,
		UserID:         log.UserID,
		Operation:      log.Operation,
		RiskScore:      riskScore,
		RiskLevel:      riskLevel,
		Decision:       decision,
		RequiredLevel:  requiredLevel,
		CurrentLevel:   log.CurrentAuthLevel,
		StepUpRequired: stepUpNeeded,
	}
}

// CalculateRiskScore computes risk score based on log data and policy.
func (s *AdaptiveSimulator) CalculateRiskScore(
	log HistoricalAuthLog,
	policy *cryptoutilIdentityUserauth.RiskScoringPolicy,
) float64 {
	score := 0.0

	// Network risk.
	if log.IsVPN {
		if vpnRisk, ok := policy.NetworkRisks["vpn"]; ok {
			score += vpnRisk.Score * policy.RiskFactors["network"].Weight
		}
	}

	if log.IsProxy {
		if proxyRisk, ok := policy.NetworkRisks["proxy"]; ok {
			score += proxyRisk.Score * policy.RiskFactors["network"].Weight
		}
	}

	// Geographic risk (simplified).
	for _, highRiskCountry := range policy.GeographicRisks.HighRiskCountries.Countries {
		if log.Country == highRiskCountry {
			score += policy.GeographicRisks.HighRiskCountries.Score * policy.RiskFactors["location"].Weight

			break
		}
	}

	// Device risk (simplified - assume new device if not in baseline).
	// Real implementation would check user baseline data.
	const newDeviceRisk = 0.6

	score += newDeviceRisk * policy.RiskFactors["device"].Weight

	return score
}

// DetermineRiskLevel categorizes risk score into risk level.
func (s *AdaptiveSimulator) DetermineRiskLevel(
	score float64,
	policy *cryptoutilIdentityUserauth.RiskScoringPolicy,
) string {
	for level, threshold := range policy.RiskThresholds {
		if score >= threshold.Min && score <= threshold.Max {
			return level
		}
	}

	return "medium" // Default.
}

// DetermineRequiredLevel checks if operation requires step-up.
func (s *AdaptiveSimulator) DetermineRequiredLevel(
	operation string,
	currentLevel string,
	policy *cryptoutilIdentityUserauth.StepUpPolicies,
) (requiredLevel string, stepUpNeeded bool) {
	// Find operation policy.
	opPolicy, found := policy.Policies[operation]
	if !found {
		opPolicy = policy.DefaultPolicy
	}

	requiredLevel = opPolicy.RequiredLevel

	// Compare with current level.
	currentLevelInt := s.AuthLevelToInt(currentLevel)
	requiredLevelInt := s.AuthLevelToInt(requiredLevel)

	stepUpNeeded = requiredLevelInt > currentLevelInt

	return requiredLevel, stepUpNeeded
}

// AuthLevelToInt converts auth level string to integer for comparison.
func (s *AdaptiveSimulator) AuthLevelToInt(level string) int {
	levels := map[string]int{
		"none":    authLevelNone,
		"basic":   authLevelBasic,
		"mfa":     authLevelMFA,
		"step_up": authLevelStepUp,
		"block":   authLevelBlock,
	}

	if val, ok := levels[level]; ok {
		return val
	}

	return 1 // Default to basic.
}

// MakeDecision determines final authentication decision.
func (s *AdaptiveSimulator) MakeDecision(
	riskLevel string,
	stepUpNeeded bool,
	policy *cryptoutilIdentityUserauth.AdaptiveAuthPolicy,
) string {
	// Check if risk level requires blocking.
	if riskLevel == "critical" {
		return "block"
	}

	// Check if step-up required by operation policy.
	if stepUpNeeded {
		return "step_up"
	}

	return "allow"
}

// GenerateRecommendations analyzes simulation results and generates recommendations.
func (s *AdaptiveSimulator) GenerateRecommendations(result *SimulationResult) []string {
	recommendations := make([]string, 0)

	// Check step-up rate.
	const stepUpRateThreshold = 0.15
	if result.StepUpRate > stepUpRateThreshold {
		recommendations = append(recommendations, fmt.Sprintf(
			"High step-up rate detected (%.1f%%). Consider relaxing policies for low-risk operations.",
			result.StepUpRate*100, //nolint:mnd
		))
	}

	// Check blocked rate.
	const blockedRateThreshold = 0.05
	if result.BlockedRate > blockedRateThreshold {
		recommendations = append(recommendations, fmt.Sprintf(
			"High blocked rate detected (%.1f%%). Review risk thresholds and false positive cases.",
			result.BlockedRate*100, //nolint:mnd
		))
	}

	// Check risk distribution.
	criticalCount := result.RiskDistribution["critical"]

	const criticalPercentageThreshold = 0.10

	if float64(criticalCount)/float64(result.TotalAttempts) > criticalPercentageThreshold {
		recommendations = append(recommendations, fmt.Sprintf(
			"High critical-risk attempts (%d/%d = %.1f%%). Investigate potential attack patterns.",
			criticalCount, result.TotalAttempts, float64(criticalCount)/float64(result.TotalAttempts)*100, //nolint:mnd
		))
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "No policy adjustments recommended. Current policy performs well.")
	}

	return recommendations
}

// LoadHistoricalLogs loads authentication logs from JSON file.
func (s *AdaptiveSimulator) LoadHistoricalLogs(path string) ([]HistoricalAuthLog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs file: %w", err)
	}

	var logs []HistoricalAuthLog
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil, fmt.Errorf("failed to parse logs: %w", err)
	}

	return logs, nil
}

// SaveResults saves simulation results to JSON file.
func (s *AdaptiveSimulator) SaveResults(result *SimulationResult, stdout io.Writer) error {
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("simulation-%s.json", timestamp)
	outputPath := filepath.Join(s.outputDir, filename)

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	if err := os.WriteFile(outputPath, data, filePerms600); err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}

	_, _ = fmt.Fprintf(stdout, "\nSimulation results saved to: %s\n", outputPath)

	return nil
}

// PrintSummary prints simulation summary to stdout.
func (s *AdaptiveSimulator) PrintSummary(result *SimulationResult, stdout io.Writer) {
	_, _ = fmt.Fprintln(stdout, "\n=== Adaptive Authentication Policy Simulation ===")
	_, _ = fmt.Fprintf(stdout, "Policy Version: %s\n", result.PolicyVersion)
	_, _ = fmt.Fprintf(stdout, "Simulation Time: %s\n", result.SimulationTime.Format(time.RFC3339))
	_, _ = fmt.Fprintf(stdout, "Total Attempts: %d\n", result.TotalAttempts)
	_, _ = fmt.Fprintln(stdout)

	_, _ = fmt.Fprintln(stdout, "=== Decisions ===")
	_, _ = fmt.Fprintf(stdout, "Allowed: %d (%.1f%%)\n", result.AllowedOperations, float64(result.AllowedOperations)/float64(result.TotalAttempts)*100) //nolint:mnd
	_, _ = fmt.Fprintf(stdout, "Step-Up Required: %d (%.1f%%)\n", result.StepUpRequired, result.StepUpRate*100)                                         //nolint:mnd
	_, _ = fmt.Fprintf(stdout, "Blocked: %d (%.1f%%)\n", result.BlockedOperations, result.BlockedRate*100)                                              //nolint:mnd
	_, _ = fmt.Fprintln(stdout)

	_, _ = fmt.Fprintln(stdout, "=== Risk Distribution ===")

	for level, count := range result.RiskDistribution {
		_, _ = fmt.Fprintf(stdout, "%s: %d (%.1f%%)\n", level, count, float64(count)/float64(result.TotalAttempts)*100) //nolint:mnd
	}

	_, _ = fmt.Fprintln(stdout)

	_, _ = fmt.Fprintln(stdout, "=== Recommendations ===")

	for i, rec := range result.Recommendations {
		_, _ = fmt.Fprintf(stdout, "%d. %s\n", i+1, rec)
	}

	_, _ = fmt.Fprintln(stdout)
}

const exampleLogFormat = `[
  {
    "timestamp": "2025-01-15T10:30:00Z",
    "user_id": "user-123",
    "operation": "transfer_funds",
    "ip_address": "192.168.1.1",
    "device_id": "device-abc",
    "country": "US",
    "city": "New York",
    "is_vpn": false,
    "is_proxy": false,
    "current_auth_level": "basic",
    "success": true,
    "metadata": {}
  }
]`
