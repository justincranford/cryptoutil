// Copyright (c) 2025 Justin Cranford

//go:build e2e

package test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	"github.com/stretchr/testify/suite"
)

// TestE2E runs the complete end-to-end test suite.
func TestE2E(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(E2ETestSuite))
}

// TestSummaryReportOnly runs a quick test to demonstrate summary report functionality.
func TestSummaryReportOnly(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(SummaryTestSuite))
}

// SummaryTestSuite provides a quick test suite to demonstrate summary reporting.
type SummaryTestSuite struct {
	suite.Suite
	fixture    *TestFixture
	assertions *ServiceAssertions
	summary    *TestSummary
}

// SetupSuite runs once before all tests in the suite.
func (suite *SummaryTestSuite) SetupSuite() {
	suite.summary = &TestSummary{
		StartTime: time.Now().UTC(),
		Steps:     make([]TestStep, 0),
	}

	suite.logStep("Summary Test Setup", "Setting up summary test suite")

	// Create test fixture
	suite.fixture = NewTestFixture(suite.T())

	// Create assertions helper
	suite.assertions = NewServiceAssertions(suite.T(), suite.fixture.logger)

	suite.completeStep("PASS", "Summary test setup completed successfully")
}

// TearDownSuite runs once after all tests in the suite.
func (suite *SummaryTestSuite) TearDownSuite() {
	suite.logStep("Summary Test Cleanup", "Cleaning up summary test suite")

	suite.completeStep("PASS", "Summary test cleanup completed")

	// Generate final summary report
	suite.generateSummaryReport()
}

// TestQuickDemo demonstrates summary tracking.
func (suite *SummaryTestSuite) TestQuickDemo() {
	suite.logStep("Quick Demo Test", "Demonstrating summary tracking functionality")

	defer func() {
		if r := recover(); r != nil {
			suite.completeStep("FAIL", fmt.Sprintf("Quick demo test failed: %v", r))
			panic(r)
		}

		suite.completeStep("PASS", "Quick demo test completed successfully")
	}()

	// Simulate some quick operations
	suite.logStep("Sub-operation 1", "Performing first sub-operation")
	time.Sleep(100 * time.Millisecond)
	suite.completeStep("PASS", "Sub-operation 1 completed")

	suite.logStep("Sub-operation 2", "Performing second sub-operation")
	time.Sleep(50 * time.Millisecond)
	suite.completeStep("PASS", "Sub-operation 2 completed")
}

// Helper methods (same as E2ETestSuite).
func (suite *SummaryTestSuite) logStep(name, description string) {
	step := TestStep{
		Name:        name,
		StartTime:   time.Now().UTC(),
		Description: description,
	}
	suite.summary.Steps = append(suite.summary.Steps, step)

	if suite.fixture != nil {
		LogTestStep(suite.fixture.logger, name, description)
	}
}

func (suite *SummaryTestSuite) completeStep(status, result string) {
	if len(suite.summary.Steps) == 0 {
		return
	}

	step := &suite.summary.Steps[len(suite.summary.Steps)-1]
	step.EndTime = time.Now().UTC()
	step.Duration = step.EndTime.Sub(step.StartTime)
	step.Status = status

	suite.summary.TotalSteps++

	switch status {
	case cryptoutilSharedMagic.TestStatusPass:
		suite.summary.PassedSteps++
	case cryptoutilSharedMagic.TestStatusFail:
		suite.summary.FailedSteps++
	case cryptoutilSharedMagic.TestStatusSkip:
		suite.summary.SkippedSteps++
	}

	statusEmoji := GetStatusEmoji(status)

	if suite.fixture != nil {
		LogTestStepCompletion(suite.fixture.logger, statusEmoji, step.Name, result, step.Duration)
	}
}

func (suite *SummaryTestSuite) generateSummaryReport() {
	suite.summary.EndTime = time.Now().UTC()
	totalDuration := suite.summary.EndTime.Sub(suite.summary.StartTime)

	report := strings.Builder{}
	report.WriteString("\n" + strings.Repeat("=", 80) + "\n")
	report.WriteString("🎯 E2E TEST EXECUTION SUMMARY REPORT\n")
	report.WriteString(strings.Repeat("=", 80) + "\n\n")

	report.WriteString(fmt.Sprintf("📅 Execution Date: %s\n", suite.summary.StartTime.Format("2006-01-02 15:04:05")))
	report.WriteString(fmt.Sprintf("⏱️  Total Duration: %v\n", totalDuration.Round(time.Millisecond)))
	report.WriteString(fmt.Sprintf("📊 Total Steps: %d\n", suite.summary.TotalSteps))
	report.WriteString(fmt.Sprintf("✅ Passed: %d\n", suite.summary.PassedSteps))
	report.WriteString(fmt.Sprintf("❌ Failed: %d\n", suite.summary.FailedSteps))
	report.WriteString(fmt.Sprintf("⏭️  Skipped: %d\n", suite.summary.SkippedSteps))

	if suite.summary.FailedSteps > 0 {
		report.WriteString(fmt.Sprintf("📈 Success Rate: %.1f%%\n", float64(suite.summary.PassedSteps)/float64(suite.summary.TotalSteps)*100))
	} else {
		report.WriteString("📈 Success Rate: 100.0%\n")
	}

	report.WriteString("\n" + strings.Repeat("-", 80) + "\n")
	report.WriteString("📋 DETAILED STEP BREAKDOWN\n")
	report.WriteString(strings.Repeat("-", 80) + "\n")

	for i, step := range suite.summary.Steps {
		statusEmoji := GetStatusEmoji(step.Status)

		report.WriteString(fmt.Sprintf("%2d. %s %-25s %8v  %s\n",
			i+1, statusEmoji, step.Name, step.Duration.Round(time.Millisecond), step.Description))
	}

	report.WriteString("\n" + strings.Repeat("=", 80) + "\n")

	if suite.summary.FailedSteps > 0 {
		report.WriteString("⚠️  EXECUTION STATUS: PARTIAL SUCCESS\n")
	} else {
		report.WriteString("🎉 EXECUTION STATUS: FULL SUCCESS\n")
	}

	report.WriteString(strings.Repeat("=", 80) + "\n")

	if suite.fixture != nil {
		Log(suite.fixture.logger, "%s", report.String())
	}
}
