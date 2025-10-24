//go:build e2e

package test

import (
	"fmt"
	"strings"
	"time"

	cryptoutilOpenapiClient "cryptoutil/api/client"
	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilClient "cryptoutil/internal/client"
	cryptoutilMagic "cryptoutil/internal/common/magic"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestStep represents a single test step with timing and status information
type TestStep struct {
	Name        string
	StartTime   time.Time
	EndTime     time.Time
	Status      string // "PASS", "FAIL", "SKIP"
	Duration    time.Duration
	Description string
}

// TestSummary tracks overall test execution information
type TestSummary struct {
	StartTime    time.Time
	EndTime      time.Time
	TotalSteps   int
	PassedSteps  int
	FailedSteps  int
	SkippedSteps int
	Steps        []TestStep
}

// E2ETestSuite provides a structured test suite for end-to-end testing
type E2ETestSuite struct {
	suite.Suite
	fixture    *TestFixture
	assertions *ServiceAssertions
	summary    *TestSummary
}

// SetupSuite runs once before all tests in the suite
func (suite *E2ETestSuite) SetupSuite() {
	suite.summary = &TestSummary{
		StartTime: time.Now(),
		Steps:     make([]TestStep, 0),
	}

	suite.logStep("E2E Test Suite Setup", "Starting E2E test suite initialization")

	// Create test fixture
	suite.fixture = NewTestFixture(suite.T())

	// Create assertions helper
	suite.assertions = NewServiceAssertions(suite.T(), suite.fixture.startTime, suite.fixture.logFile)

	// Setup infrastructure
	suite.fixture.Setup()

	suite.completeStep("PASS", "E2E test suite setup completed successfully")
}

// TearDownSuite runs once after all tests in the suite
func (suite *E2ETestSuite) TearDownSuite() {
	suite.logStep("E2E Test Suite Cleanup", "Starting test suite cleanup")

	// Teardown infrastructure
	suite.fixture.Teardown()

	suite.completeStep("PASS", "Test suite cleanup completed")

	// Generate final summary report
	suite.generateSummaryReport()
}

// SetupTest runs before each test method
func (suite *E2ETestSuite) SetupTest() {
	fmt.Printf("[%s] [%v] 📋 Setting up test: %s\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second),
		suite.T().Name())

	// Initialize API clients for each test
	suite.fixture.InitializeAPIClients()
}

// TearDownTest runs after each test method
func (suite *E2ETestSuite) TearDownTest() {
	fmt.Printf("[%s] [%v] 🧹 Cleaning up test: %s\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second),
		suite.T().Name())

	// Clean up any test data created during the test
	suite.cleanupTestData()
}

// TestInfrastructureHealth verifies all services are healthy
func (suite *E2ETestSuite) TestInfrastructureHealth() {
	suite.logStep("Infrastructure Health Check", "Verifying all Docker services are healthy")

	defer func() {
		if r := recover(); r != nil {
			suite.completeStep("FAIL", fmt.Sprintf("Infrastructure health check failed: %v", r))
			panic(r)
		}
		suite.completeStep("PASS", "All infrastructure services are healthy")
	}()

	suite.assertions.AssertDockerServicesHealthy()
	suite.assertions.AssertHTTPReady(suite.fixture.ctx, suite.fixture.GetServiceURL("grafana")+"/api/health", cryptoutilMagic.TestTimeoutCryptoutilReady)
	suite.assertions.AssertHTTPReady(suite.fixture.ctx, suite.fixture.GetServiceURL("otel")+"/metrics", cryptoutilMagic.TestTimeoutCryptoutilReady)
}

// TestCryptoutilSQLite tests SQLite-based cryptoutil instance
func (suite *E2ETestSuite) TestCryptoutilSQLite() {
	suite.logStep("SQLite Cryptoutil Tests", "Testing SQLite-based cryptoutil instance")

	defer func() {
		if r := recover(); r != nil {
			suite.completeStep("FAIL", fmt.Sprintf("SQLite cryptoutil tests failed: %v", r))
			panic(r)
		}
		suite.completeStep("PASS", "SQLite cryptoutil instance tests completed successfully")
	}()

	suite.testCryptoutilInstance("sqlite")
}

// TestCryptoutilPostgres1 tests PostgreSQL-based cryptoutil instance #1
func (suite *E2ETestSuite) TestCryptoutilPostgres1() {
	suite.logStep("PostgreSQL #1 Cryptoutil Tests", "Testing PostgreSQL instance #1 cryptoutil")

	defer func() {
		if r := recover(); r != nil {
			suite.completeStep("FAIL", fmt.Sprintf("PostgreSQL #1 cryptoutil tests failed: %v", r))
			panic(r)
		}
		suite.completeStep("PASS", "PostgreSQL #1 cryptoutil instance tests completed successfully")
	}()

	suite.testCryptoutilInstance("postgres1")
}

// TestCryptoutilPostgres2 tests PostgreSQL-based cryptoutil instance #2
func (suite *E2ETestSuite) TestCryptoutilPostgres2() {
	suite.logStep("PostgreSQL #2 Cryptoutil Tests", "Testing PostgreSQL instance #2 cryptoutil")

	defer func() {
		if r := recover(); r != nil {
			suite.completeStep("FAIL", fmt.Sprintf("PostgreSQL #2 cryptoutil tests failed: %v", r))
			panic(r)
		}
		suite.completeStep("PASS", "PostgreSQL #2 cryptoutil instance tests completed successfully")
	}()

	suite.testCryptoutilInstance("postgres2")
}

// TestTelemetryFlow verifies telemetry is flowing correctly
func (suite *E2ETestSuite) TestTelemetryFlow() {
	suite.logStep("Telemetry Flow Tests", "Verifying telemetry data flow between services")

	defer func() {
		if r := recover(); r != nil {
			suite.completeStep("FAIL", fmt.Sprintf("Telemetry flow tests failed: %v", r))
			panic(r)
		}
		suite.completeStep("PASS", "Telemetry flow verification completed successfully")
	}()

	suite.assertions.AssertTelemetryFlow(
		suite.fixture.ctx,
		suite.fixture.GetServiceURL("grafana"),
		suite.fixture.GetServiceURL("otel"),
	)
}

// testCryptoutilInstance tests a single cryptoutil instance
func (suite *E2ETestSuite) testCryptoutilInstance(instanceName string) {
	stepName := fmt.Sprintf("%s Instance Tests", strings.Title(instanceName))
	suite.logStep(stepName, fmt.Sprintf("Testing %s cryptoutil instance functionality", instanceName))

	defer func() {
		if r := recover(); r != nil {
			suite.completeStep("FAIL", fmt.Sprintf("%s instance tests failed: %v", instanceName, r))
			panic(r)
		}
	}()

	client := suite.fixture.GetClient(instanceName)
	baseURL := suite.fixture.GetServiceURL(instanceName)

	// Test health check
	suite.assertions.AssertCryptoutilHealth(baseURL, suite.fixture.rootCAsPool)

	// Test core functionality
	elasticKey := suite.testCreateElasticKey(client)
	suite.testGenerateMaterialKey(client, elasticKey)
	suite.testEncryptDecryptCycle(client, elasticKey)
	suite.testSignVerifyCycle(client, elasticKey)

	suite.completeStep("PASS", fmt.Sprintf("%s instance tests completed successfully", instanceName))
}

// testCreateElasticKey creates a test elastic key
func (suite *E2ETestSuite) testCreateElasticKey(client *cryptoutilOpenapiClient.ClientWithResponses) *cryptoutilOpenapiModel.ElasticKey {
	suite.logStep("Create Elastic Key", "Creating test elastic key for cryptographic operations")

	defer func() {
		if r := recover(); r != nil {
			suite.completeStep("FAIL", fmt.Sprintf("Elastic key creation failed: %v", r))
			panic(r)
		}
	}()

	elasticKeyCreate := cryptoutilClient.RequireCreateElasticKeyRequest(
		suite.T(), &testElasticKeyName, &testElasticKeyDescription,
		&testAlgorithm, &testProvider, &importAllowed, &versioningAllowed,
	)

	elasticKey := cryptoutilClient.RequireCreateElasticKeyResponse(suite.T(), suite.fixture.ctx, client, elasticKeyCreate)
	require.NotNil(suite.T(), elasticKey.ElasticKeyID)

	suite.completeStep("PASS", fmt.Sprintf("Elastic key created with ID: %s", *elasticKey.ElasticKeyID))
	return elasticKey
}

// testGenerateMaterialKey generates a material key
func (suite *E2ETestSuite) testGenerateMaterialKey(client *cryptoutilOpenapiClient.ClientWithResponses, elasticKey *cryptoutilOpenapiModel.ElasticKey) {
	suite.logStep("Generate Material Key", "Generating material key from elastic key")

	defer func() {
		if r := recover(); r != nil {
			suite.completeStep("FAIL", fmt.Sprintf("Material key generation failed: %v", r))
			panic(r)
		}
	}()

	keyGenerate := cryptoutilClient.RequireMaterialKeyGenerateRequest(suite.T())
	materialKey := cryptoutilClient.RequireMaterialKeyGenerateResponse(suite.T(), suite.fixture.ctx, client, elasticKey.ElasticKeyID, keyGenerate)
	require.NotNil(suite.T(), materialKey.MaterialKeyID)

	suite.completeStep("PASS", fmt.Sprintf("Material key generated with ID: %s", materialKey.MaterialKeyID))
}

// testEncryptDecryptCycle tests full encrypt/decrypt cycle
func (suite *E2ETestSuite) testEncryptDecryptCycle(client *cryptoutilOpenapiClient.ClientWithResponses, elasticKey *cryptoutilOpenapiModel.ElasticKey) {
	suite.logStep("Encrypt/Decrypt Cycle", "Testing full encryption and decryption cycle")

	defer func() {
		if r := recover(); r != nil {
			suite.completeStep("FAIL", fmt.Sprintf("Encrypt/decrypt cycle failed: %v", r))
			panic(r)
		}
	}()

	// Encrypt
	encryptRequest := cryptoutilClient.RequireEncryptRequest(suite.T(), &cryptoutilMagic.TestCleartext)
	encryptedText := cryptoutilClient.RequireEncryptResponse(suite.T(), suite.fixture.ctx, client, elasticKey.ElasticKeyID, nil, encryptRequest)
	require.NotEmpty(suite.T(), *encryptedText)

	// Decrypt
	decryptRequest := cryptoutilClient.RequireDecryptRequest(suite.T(), encryptedText)
	decryptedText := cryptoutilClient.RequireDecryptResponse(suite.T(), suite.fixture.ctx, client, elasticKey.ElasticKeyID, decryptRequest)
	require.Equal(suite.T(), cryptoutilMagic.TestCleartext, *decryptedText)

	suite.completeStep("PASS", "Encrypt/decrypt cycle completed successfully")
}

// testSignVerifyCycle tests full sign/verify cycle
func (suite *E2ETestSuite) testSignVerifyCycle(client *cryptoutilOpenapiClient.ClientWithResponses, elasticKey *cryptoutilOpenapiModel.ElasticKey) {
	suite.logStep("Sign/Verify Cycle", "Testing full digital signature and verification cycle")

	defer func() {
		if r := recover(); r != nil {
			suite.completeStep("FAIL", fmt.Sprintf("Sign/verify cycle failed: %v", r))
			panic(r)
		}
	}()

	// Sign
	signRequest := cryptoutilClient.RequireSignRequest(suite.T(), &cryptoutilMagic.TestCleartext)
	signedText := cryptoutilClient.RequireSignResponse(suite.T(), suite.fixture.ctx, client, elasticKey.ElasticKeyID, nil, signRequest)
	require.NotEmpty(suite.T(), *signedText)

	// Verify
	verifyRequest := cryptoutilClient.RequireVerifyRequest(suite.T(), signedText)
	verifyResponse := cryptoutilClient.RequireVerifyResponse(suite.T(), suite.fixture.ctx, client, elasticKey.ElasticKeyID, verifyRequest)
	require.Equal(suite.T(), "true", *verifyResponse)

	suite.completeStep("PASS", "Sign/verify cycle completed successfully")
}

// cleanupTestData cleans up any test data created during tests
func (suite *E2ETestSuite) cleanupTestData() {
	// This could include deleting test keys, clearing databases, etc.
	// Implementation depends on what test data is created
}

// logStep starts tracking a new test step
func (suite *E2ETestSuite) logStep(name, description string) {
	step := TestStep{
		Name:        name,
		StartTime:   time.Now(),
		Description: description,
	}
	suite.summary.Steps = append(suite.summary.Steps, step)

	// Only log to fixture if it exists (it won't exist during very early setup)
	if suite.fixture != nil {
		suite.fixture.log("[%s] [%v] 📋 %s: %s",
			step.StartTime.Format("15:04:05"),
			time.Since(suite.fixture.startTime).Round(time.Second),
			name, description)
	}
}

// completeStep marks the current step as completed with a status
func (suite *E2ETestSuite) completeStep(status, result string) {
	if len(suite.summary.Steps) == 0 {
		return
	}

	step := &suite.summary.Steps[len(suite.summary.Steps)-1]
	step.EndTime = time.Now()
	step.Duration = step.EndTime.Sub(step.StartTime)
	step.Status = status

	suite.summary.TotalSteps++

	switch status {
	case "PASS":
		suite.summary.PassedSteps++
	case "FAIL":
		suite.summary.FailedSteps++
	case "SKIP":
		suite.summary.SkippedSteps++
	}

	statusEmoji := "✅"
	if status == "FAIL" {
		statusEmoji = "❌"
	} else if status == "SKIP" {
		statusEmoji = "⏭️"
	}

	// Only log to fixture if it exists
	if suite.fixture != nil {
		suite.fixture.log("[%s] [%v] %s %s: %s (took %v)",
			step.EndTime.Format("15:04:05"),
			time.Since(suite.fixture.startTime).Round(time.Second),
			statusEmoji, step.Name, result, step.Duration.Round(time.Millisecond))
	}
}

// generateSummaryReport creates and displays a detailed summary report
func (suite *E2ETestSuite) generateSummaryReport() {
	suite.summary.EndTime = time.Now()
	totalDuration := suite.summary.EndTime.Sub(suite.summary.StartTime)

	// Generate summary report
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
		statusEmoji := "✅"
		if step.Status == "FAIL" {
			statusEmoji = "❌"
		} else if step.Status == "SKIP" {
			statusEmoji = "⏭️"
		}

		report.WriteString(fmt.Sprintf("%2d. %s %-20s %8v  %s\n",
			i+1, statusEmoji, step.Name, step.Duration.Round(time.Millisecond), step.Description))
	}

	report.WriteString("\n" + strings.Repeat("=", 80) + "\n")

	if suite.summary.FailedSteps > 0 {
		report.WriteString("⚠️  EXECUTION STATUS: PARTIAL SUCCESS\n")
	} else {
		report.WriteString("🎉 EXECUTION STATUS: FULL SUCCESS\n")
	}
	report.WriteString(strings.Repeat("=", 80) + "\n")

	// Log the report to both console and file
	suite.fixture.log("%s", report.String())
}

// Test constants (moved from original file)
var (
	testElasticKeyName        = "e2e-test-key"
	testElasticKeyDescription = "E2E integration test key"
	testAlgorithm             = "RSA"
	testProvider              = "GO"
	importAllowed             = false
	versioningAllowed         = true
)
