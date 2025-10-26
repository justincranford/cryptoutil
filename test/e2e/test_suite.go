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
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// TestStep represents a single test step with timing and status information.
type TestStep struct {
	Name        string
	StartTime   time.Time
	EndTime     time.Time
	Status      string // "PASS", "FAIL", "SKIP"
	Duration    time.Duration
	Description string
}

// TestSummary tracks overall test execution information.
type TestSummary struct {
	StartTime    time.Time
	EndTime      time.Time
	TotalSteps   int
	PassedSteps  int
	FailedSteps  int
	SkippedSteps int
	Steps        []TestStep
}

// E2ETestSuite provides a structured test suite for end-to-end testing.
type E2ETestSuite struct {
	suite.Suite
	fixture    *TestFixture
	assertions *ServiceAssertions
	summary    *TestSummary
}

// SetupSuite runs once before all tests in the suite.
func (suite *E2ETestSuite) SetupSuite() {
	suite.summary = &TestSummary{
		StartTime: time.Now(),
		Steps:     make([]TestStep, 0),
	}

	suite.logStep("E2E Test Suite Setup", "Starting E2E test suite initialization")

	// Create test fixture
	suite.fixture = NewTestFixture(suite.T())

	// Create assertions helper
	suite.assertions = NewServiceAssertions(suite.T(), suite.fixture.logger)

	// Setup infrastructure
	suite.fixture.Setup()

	suite.completeStep("PASS", "E2E test suite setup completed successfully")
}

// TearDownSuite runs once after all tests in the suite.
func (suite *E2ETestSuite) TearDownSuite() {
	suite.logStep("E2E Test Suite Cleanup", "Starting test suite cleanup")

	// Generate final summary report before tearing down infrastructure
	suite.generateSummaryReport()

	// Teardown infrastructure
	suite.fixture.Teardown()

	suite.completeStep("PASS", "Test suite cleanup completed")
}

// SetupTest runs before each test method.
func (suite *E2ETestSuite) SetupTest() {
	// Initialize API clients for each test
	suite.fixture.InitializeAPIClients()

	// Log test setup
	if suite.fixture != nil {
		LogTestSetup(suite.fixture.logger, suite.T().Name())
	}
}

// TearDownTest runs after each test method.
func (suite *E2ETestSuite) TearDownTest() {
	// Log test cleanup
	if suite.fixture != nil {
		LogTestCleanup(suite.fixture.logger, suite.T().Name())
	}

	// Clean up any test data created during the test
	suite.cleanupTestData()
}

// TestInfrastructureHealth verifies all services are healthy.
func (suite *E2ETestSuite) TestInfrastructureHealth() {
	suite.logStep("Infrastructure Health Check", "Verifying all Docker services are healthy")

	suite.withTestStepRecovery("Infrastructure health check failed: %v", func() string { return "All infrastructure services are healthy" }, func() {
		suite.assertions.AssertDockerServicesHealthy()
		suite.assertions.AssertHTTPReady(suite.fixture.ctx, suite.fixture.GetServiceURL("grafana")+"/api/health", cryptoutilMagic.TestTimeoutCryptoutilReady)
		suite.assertions.AssertHTTPReady(suite.fixture.ctx, suite.fixture.GetServiceURL("otel"), cryptoutilMagic.TestTimeoutCryptoutilReady)
	})
}

// TestCryptoutilSQLite tests SQLite-based cryptoutil instance.
func (suite *E2ETestSuite) TestCryptoutilSQLite() {
	suite.logStep("SQLite Cryptoutil Tests", "Testing SQLite-based cryptoutil instance")

	suite.withTestStepRecovery("SQLite cryptoutil tests failed: %v", func() string { return "SQLite cryptoutil instance tests completed successfully" }, func() {
		suite.testCryptoutilInstance("sqlite")
	})
}

// TestCryptoutilPostgres1 tests PostgreSQL-based cryptoutil instance #1.
func (suite *E2ETestSuite) TestCryptoutilPostgres1() {
	suite.logStep("PostgreSQL #1 Cryptoutil Tests", "Testing PostgreSQL instance #1 cryptoutil")

	suite.withTestStepRecovery("PostgreSQL #1 cryptoutil tests failed: %v", func() string { return "PostgreSQL #1 cryptoutil instance tests completed successfully" }, func() {
		suite.testCryptoutilInstance("postgres1")
	})
}

// TestCryptoutilPostgres2 tests PostgreSQL-based cryptoutil instance #2.
func (suite *E2ETestSuite) TestCryptoutilPostgres2() {
	suite.logStep("PostgreSQL #2 Cryptoutil Tests", "Testing PostgreSQL instance #2 cryptoutil")

	suite.withTestStepRecovery("PostgreSQL #2 cryptoutil tests failed: %v", func() string { return "PostgreSQL #2 cryptoutil instance tests completed successfully" }, func() {
		suite.testCryptoutilInstance("postgres2")
	})
}

// TestTelemetryFlow verifies telemetry is flowing correctly.
func (suite *E2ETestSuite) TestTelemetryFlow() {
	suite.logStep("Telemetry Flow Tests", "Verifying telemetry data flow between services")

	suite.withTestStepRecovery("Telemetry flow tests failed: %v", func() string { return "Telemetry flow verification completed successfully" }, func() {
		suite.assertions.AssertTelemetryFlow(
			suite.fixture.ctx,
			suite.fixture.GetServiceURL("grafana"),
			suite.fixture.GetServiceURL("otel"),
		)
	})
}

// testCryptoutilInstance tests a single cryptoutil instance.
func (suite *E2ETestSuite) testCryptoutilInstance(instanceName string) {
	caser := cases.Title(language.English)
	stepName := fmt.Sprintf("%s Instance Tests", caser.String(instanceName))
	suite.logStep(stepName, fmt.Sprintf("Testing %s cryptoutil instance functionality", instanceName))

	suite.withTestStepRecovery("%s instance tests failed: %v", func() string {
		return fmt.Sprintf("%s instance tests completed successfully", caser.String(instanceName))
	}, func() {
		client := suite.fixture.GetClient(instanceName)
		baseURL := suite.fixture.GetServiceURL(instanceName)

		// Test health check
		suite.assertions.AssertCryptoutilReady(suite.fixture.ctx, baseURL, suite.fixture.rootCAsPool)

		// Test core functionality
		encryptionKey := suite.testCreateEncryptionKey(client, instanceName)
		suite.testGenerateMaterialKey(client, encryptionKey)
		suite.testEncryptDecryptCycle(client, encryptionKey)

		signingKey := suite.testCreateSigningKey(client, instanceName)
		suite.testSignVerifyCycle(client, signingKey)
	})
}

// testCreateEncryptionKey creates a test elastic key for encryption operations.
func (suite *E2ETestSuite) testCreateEncryptionKey(client *cryptoutilOpenapiClient.ClientWithResponses, instanceName string) *cryptoutilOpenapiModel.ElasticKey {
	suite.logStep("Create Encryption Key", "Creating test elastic key for encryption operations")

	var elasticKey *cryptoutilOpenapiModel.ElasticKey

	suite.withTestStepRecovery("Encryption key creation failed: %v", func() string { return fmt.Sprintf("Encryption key created with ID: %s", *elasticKey.ElasticKeyID) }, func() {
		// Create instance-specific key name to avoid conflicts
		instanceKeyName := fmt.Sprintf("e2e-test-encrypt-key-%s", instanceName)
		instanceKeyDescription := fmt.Sprintf("E2E integration test encryption key for %s", instanceName)

		encryptionAlgorithm := cryptoutilMagic.TestJwkJweAlgorithm // JWE algorithm for encryption

		elasticKeyCreate := cryptoutilClient.RequireCreateElasticKeyRequest(
			suite.T(), &instanceKeyName, &instanceKeyDescription,
			&encryptionAlgorithm, &cryptoutilMagic.StringProviderInternal, &cryptoutilMagic.TestElasticKeyImportAllowed, &cryptoutilMagic.TestElasticKeyVersioningAllowed,
		)

		elasticKey = cryptoutilClient.RequireCreateElasticKeyResponse(suite.T(), suite.fixture.ctx, client, elasticKeyCreate)
		require.NotNil(suite.T(), elasticKey.ElasticKeyID)
	})

	return elasticKey
}

// testCreateSigningKey creates a test elastic key for signing operations.
func (suite *E2ETestSuite) testCreateSigningKey(client *cryptoutilOpenapiClient.ClientWithResponses, instanceName string) *cryptoutilOpenapiModel.ElasticKey {
	suite.logStep("Create Signing Key", "Creating test elastic key for signing operations")

	var elasticKey *cryptoutilOpenapiModel.ElasticKey

	suite.withTestStepRecovery("Signing key creation failed: %v", func() string { return fmt.Sprintf("Signing key created with ID: %s", *elasticKey.ElasticKeyID) }, func() {
		// Create instance-specific key name to avoid conflicts
		instanceKeyName := fmt.Sprintf("e2e-test-sign-key-%s", instanceName)
		instanceKeyDescription := fmt.Sprintf("E2E integration test signing key for %s", instanceName)

		signingAlgorithm := cryptoutilMagic.TestJwkJwsAlgorithm // JWS algorithm for signing

		elasticKeyCreate := cryptoutilClient.RequireCreateElasticKeyRequest(
			suite.T(), &instanceKeyName, &instanceKeyDescription,
			&signingAlgorithm, &cryptoutilMagic.StringProviderInternal, &cryptoutilMagic.TestElasticKeyImportAllowed, &cryptoutilMagic.TestElasticKeyVersioningAllowed,
		)

		elasticKey = cryptoutilClient.RequireCreateElasticKeyResponse(suite.T(), suite.fixture.ctx, client, elasticKeyCreate)
		require.NotNil(suite.T(), elasticKey.ElasticKeyID)
	})

	return elasticKey
}

// testGenerateMaterialKey generates a material key.
func (suite *E2ETestSuite) testGenerateMaterialKey(client *cryptoutilOpenapiClient.ClientWithResponses, elasticKey *cryptoutilOpenapiModel.ElasticKey) {
	suite.logStep("Generate Material Key", "Generating material key from elastic key")

	suite.withTestStepRecovery("Material key generation failed: %v", func() string {
		return fmt.Sprintf("Material key generated with ID: %s", "placeholder") // Will be updated when we have the actual key
	}, func() {
		keyGenerate := cryptoutilClient.RequireMaterialKeyGenerateRequest(suite.T())
		materialKey := cryptoutilClient.RequireMaterialKeyGenerateResponse(suite.T(), suite.fixture.ctx, client, elasticKey.ElasticKeyID, keyGenerate)
		require.NotNil(suite.T(), materialKey.MaterialKeyID)
		// Update the success message with the actual key ID
		// Note: This is a limitation of the current design - we can't dynamically update the success message
		// For now, we'll use a generic message
	})
}

// testEncryptDecryptCycle tests full encrypt/decrypt cycle.
func (suite *E2ETestSuite) testEncryptDecryptCycle(client *cryptoutilOpenapiClient.ClientWithResponses, elasticKey *cryptoutilOpenapiModel.ElasticKey) {
	suite.logStep("Encrypt/Decrypt Cycle", "Testing full encryption and decryption cycle")

	suite.withTestStepRecovery("Encrypt/decrypt cycle failed: %v", func() string { return "Encrypt/decrypt cycle completed successfully" }, func() {
		// Encrypt
		encryptRequest := cryptoutilClient.RequireEncryptRequest(suite.T(), &cryptoutilMagic.TestCleartext)
		encryptedText := cryptoutilClient.RequireEncryptResponse(suite.T(), suite.fixture.ctx, client, elasticKey.ElasticKeyID, nil, encryptRequest)
		require.NotEmpty(suite.T(), *encryptedText)

		// Decrypt
		decryptRequest := cryptoutilClient.RequireDecryptRequest(suite.T(), encryptedText)
		decryptedText := cryptoutilClient.RequireDecryptResponse(suite.T(), suite.fixture.ctx, client, elasticKey.ElasticKeyID, decryptRequest)
		require.Equal(suite.T(), cryptoutilMagic.TestCleartext, *decryptedText)
	})
}

// testSignVerifyCycle tests full sign/verify cycle.
func (suite *E2ETestSuite) testSignVerifyCycle(client *cryptoutilOpenapiClient.ClientWithResponses, elasticKey *cryptoutilOpenapiModel.ElasticKey) {
	suite.logStep("Sign/Verify Cycle", "Testing full digital signature and verification cycle")

	suite.withTestStepRecovery("Sign/verify cycle failed: %v", func() string { return "Sign/verify cycle completed successfully" }, func() {
		// Sign
		signRequest := cryptoutilClient.RequireSignRequest(suite.T(), &cryptoutilMagic.TestCleartext)
		signedText := cryptoutilClient.RequireSignResponse(suite.T(), suite.fixture.ctx, client, elasticKey.ElasticKeyID, nil, signRequest)
		require.NotEmpty(suite.T(), *signedText)

		// Verify
		verifyRequest := cryptoutilClient.RequireVerifyRequest(suite.T(), signedText)
		verifyResponse := cryptoutilClient.RequireVerifyResponse(suite.T(), suite.fixture.ctx, client, elasticKey.ElasticKeyID, verifyRequest)
		// For successful verification, API returns 204 No Content with empty body
		require.Equal(suite.T(), "", *verifyResponse)
	})
}

// withTestStepRecovery executes a test function with consistent panic recovery and step completion.
func (suite *E2ETestSuite) withTestStepRecovery(failMessageFormat string, successMessageFunc func() string, testFunc func()) {
	defer func() {
		if r := recover(); r != nil {
			suite.completeStep("FAIL", fmt.Sprintf(failMessageFormat, r))
			panic(r)
		}

		suite.completeStep("PASS", successMessageFunc())
	}()
	testFunc()
}

// cleanupTestData cleans up any test data created during tests.
func (suite *E2ETestSuite) cleanupTestData() {
	// This could include deleting test keys, clearing databases, etc.
	// Implementation depends on what test data is created
}

// logStep starts tracking a new test step.
func (suite *E2ETestSuite) logStep(name, description string) {
	step := TestStep{
		Name:        name,
		StartTime:   time.Now(),
		Description: description,
	}
	suite.summary.Steps = append(suite.summary.Steps, step)

	// Only log to fixture if it exists (it won't exist during very early setup)
	if suite.fixture != nil {
		LogTestStep(suite.fixture.logger, name, description)
	}
}

// completeStep marks the current step as completed with a status.
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
	case cryptoutilMagic.TestStatusPass:
		suite.summary.PassedSteps++
	case cryptoutilMagic.TestStatusFail:
		suite.summary.FailedSteps++
	case cryptoutilMagic.TestStatusSkip:
		suite.summary.SkippedSteps++
	}

	statusEmoji := cryptoutilMagic.TestStatusEmojiPass

	switch status {
	case cryptoutilMagic.TestStatusFail:
		statusEmoji = cryptoutilMagic.TestStatusEmojiFail
	case cryptoutilMagic.TestStatusSkip:
		statusEmoji = cryptoutilMagic.TestStatusEmojiSkip
	}

	// Only log to fixture if it exists
	if suite.fixture != nil {
		LogTestStepCompletion(suite.fixture.logger, statusEmoji, step.Name, result, step.Duration)
	}
}

// generateSummaryReport creates and displays a detailed summary report.
func (suite *E2ETestSuite) generateSummaryReport() {
	suite.summary.EndTime = time.Now()
	totalDuration := suite.summary.EndTime.Sub(suite.summary.StartTime)

	// Generate summary report
	report := strings.Builder{}
	report.WriteString("\n" + strings.Repeat("=", cryptoutilMagic.TestReportWidth) + "\n")
	report.WriteString("🎯 E2E TEST EXECUTION SUMMARY REPORT\n")
	report.WriteString(strings.Repeat("=", cryptoutilMagic.TestReportWidth) + "\n\n")

	report.WriteString(fmt.Sprintf("📅 Execution Date: %s\n", suite.summary.StartTime.Format("2006-01-02 15:04:05")))
	report.WriteString(fmt.Sprintf("⏱️  Total Duration: %v\n", totalDuration.Round(time.Millisecond)))
	report.WriteString(fmt.Sprintf("📊 Total Steps: %d\n", suite.summary.TotalSteps))
	report.WriteString(fmt.Sprintf("✅ Passed: %d\n", suite.summary.PassedSteps))
	report.WriteString(fmt.Sprintf("❌ Failed: %d\n", suite.summary.FailedSteps))
	report.WriteString(fmt.Sprintf("⏭️  Skipped: %d\n", suite.summary.SkippedSteps))

	if suite.summary.FailedSteps > 0 {
		report.WriteString(fmt.Sprintf("📈 Success Rate: %.1f%%\n", float64(suite.summary.PassedSteps)/float64(suite.summary.TotalSteps)*cryptoutilMagic.PercentageBasis100))
	} else {
		report.WriteString("📈 Success Rate: 100.0%\n")
	}

	report.WriteString("\n" + strings.Repeat("-", cryptoutilMagic.TestReportWidth) + "\n")
	report.WriteString("📋 DETAILED STEP BREAKDOWN\n")
	report.WriteString(strings.Repeat("-", cryptoutilMagic.TestReportWidth) + "\n")

	for i, step := range suite.summary.Steps {
		statusEmoji := cryptoutilMagic.TestStatusEmojiPass

		switch step.Status {
		case cryptoutilMagic.TestStatusFail:
			statusEmoji = cryptoutilMagic.TestStatusEmojiFail
		case cryptoutilMagic.TestStatusSkip:
			statusEmoji = cryptoutilMagic.TestStatusEmojiSkip
		}

		report.WriteString(fmt.Sprintf("%2d. %s %-20s %8v  %s\n",
			i+1, statusEmoji, step.Name, step.Duration.Round(time.Millisecond), step.Description))
	}

	report.WriteString("\n" + strings.Repeat("=", cryptoutilMagic.TestReportWidth) + "\n")

	if suite.summary.FailedSteps > 0 {
		report.WriteString("⚠️  EXECUTION STATUS: PARTIAL SUCCESS\n")
	} else {
		report.WriteString("🎉 EXECUTION STATUS: FULL SUCCESS\n")
	}

	report.WriteString(strings.Repeat("=", cryptoutilMagic.TestReportWidth) + "\n")

	// Log the report to both console and file
	Log(suite.fixture.logger, "%s", report.String())
}
