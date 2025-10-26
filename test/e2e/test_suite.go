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
	fmt.Printf("[%s] [%v] 📋 Setting up test: %s\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second),
		suite.T().Name())

	// Initialize API clients for each test
	suite.fixture.InitializeAPIClients()
}

// TearDownTest runs after each test method.
func (suite *E2ETestSuite) TearDownTest() {
	fmt.Printf("[%s] [%v] 🧹 Cleaning up test: %s\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second),
		suite.T().Name())

	// Clean up any test data created during the test
	suite.cleanupTestData()
}

// TestInfrastructureHealth verifies all services are healthy.
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
	suite.assertions.AssertHTTPReady(suite.fixture.ctx, suite.fixture.GetServiceURL("otel"), cryptoutilMagic.TestTimeoutCryptoutilReady)
}

// TestCryptoutilSQLite tests SQLite-based cryptoutil instance.
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

// TestCryptoutilPostgres1 tests PostgreSQL-based cryptoutil instance #1.
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

// TestCryptoutilPostgres2 tests PostgreSQL-based cryptoutil instance #2.
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

// TestTelemetryFlow verifies telemetry is flowing correctly.
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

// testCryptoutilInstance tests a single cryptoutil instance.
func (suite *E2ETestSuite) testCryptoutilInstance(instanceName string) {
	caser := cases.Title(language.English)
	stepName := fmt.Sprintf("%s Instance Tests", caser.String(instanceName))
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
	suite.assertions.AssertCryptoutilReady(suite.fixture.ctx, baseURL, suite.fixture.rootCAsPool)

	// Test core functionality
	encryptionKey := suite.testCreateEncryptionKey(client, instanceName)
	suite.testGenerateMaterialKey(client, encryptionKey)
	suite.testEncryptDecryptCycle(client, encryptionKey)

	signingKey := suite.testCreateSigningKey(client, instanceName)
	suite.testSignVerifyCycle(client, signingKey)

	suite.completeStep("PASS", fmt.Sprintf("%s instance tests completed successfully", instanceName))
}

// testCreateEncryptionKey creates a test elastic key for encryption operations.
func (suite *E2ETestSuite) testCreateEncryptionKey(client *cryptoutilOpenapiClient.ClientWithResponses, instanceName string) *cryptoutilOpenapiModel.ElasticKey {
	suite.logStep("Create Encryption Key", "Creating test elastic key for encryption operations")

	defer func() {
		if r := recover(); r != nil {
			suite.completeStep("FAIL", fmt.Sprintf("Encryption key creation failed: %v", r))
			panic(r)
		}
	}()

	// Create instance-specific key name to avoid conflicts
	instanceKeyName := fmt.Sprintf("e2e-test-encrypt-key-%s", instanceName)
	instanceKeyDescription := fmt.Sprintf("E2E integration test encryption key for %s", instanceName)

	encryptionAlgorithm := cryptoutilMagic.TestJwkJweAlgorithm // JWE algorithm for encryption

	elasticKeyCreate := cryptoutilClient.RequireCreateElasticKeyRequest(
		suite.T(), &instanceKeyName, &instanceKeyDescription,
		&encryptionAlgorithm, &cryptoutilMagic.StringProviderInternal, &cryptoutilMagic.TestElasticKeyImportAllowed, &cryptoutilMagic.TestElasticKeyVersioningAllowed,
	)

	elasticKey := cryptoutilClient.RequireCreateElasticKeyResponse(suite.T(), suite.fixture.ctx, client, elasticKeyCreate)
	require.NotNil(suite.T(), elasticKey.ElasticKeyID)

	suite.completeStep("PASS", fmt.Sprintf("Encryption key created with ID: %s", *elasticKey.ElasticKeyID))

	return elasticKey
}

// testCreateSigningKey creates a test elastic key for signing operations.
func (suite *E2ETestSuite) testCreateSigningKey(client *cryptoutilOpenapiClient.ClientWithResponses, instanceName string) *cryptoutilOpenapiModel.ElasticKey {
	suite.logStep("Create Signing Key", "Creating test elastic key for signing operations")

	defer func() {
		if r := recover(); r != nil {
			suite.completeStep("FAIL", fmt.Sprintf("Signing key creation failed: %v", r))
			panic(r)
		}
	}()

	// Create instance-specific key name to avoid conflicts
	instanceKeyName := fmt.Sprintf("e2e-test-sign-key-%s", instanceName)
	instanceKeyDescription := fmt.Sprintf("E2E integration test signing key for %s", instanceName)

	signingAlgorithm := cryptoutilMagic.TestJwkJwsAlgorithm // JWS algorithm for signing

	elasticKeyCreate := cryptoutilClient.RequireCreateElasticKeyRequest(
		suite.T(), &instanceKeyName, &instanceKeyDescription,
		&signingAlgorithm, &cryptoutilMagic.StringProviderInternal, &cryptoutilMagic.TestElasticKeyImportAllowed, &cryptoutilMagic.TestElasticKeyVersioningAllowed,
	)

	elasticKey := cryptoutilClient.RequireCreateElasticKeyResponse(suite.T(), suite.fixture.ctx, client, elasticKeyCreate)
	require.NotNil(suite.T(), elasticKey.ElasticKeyID)

	suite.completeStep("PASS", fmt.Sprintf("Signing key created with ID: %s", *elasticKey.ElasticKeyID))

	return elasticKey
}

// testGenerateMaterialKey generates a material key.
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

// testEncryptDecryptCycle tests full encrypt/decrypt cycle.
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

// testSignVerifyCycle tests full sign/verify cycle.
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
	// For successful verification, API returns 204 No Content with empty body
	require.Equal(suite.T(), "", *verifyResponse)

	suite.completeStep("PASS", "Sign/verify cycle completed successfully")
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
		suite.fixture.logger.LogTestStep(name, description)
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
	if status == cryptoutilMagic.TestStatusFail {
		statusEmoji = cryptoutilMagic.TestStatusEmojiFail
	} else if status == cryptoutilMagic.TestStatusSkip {
		statusEmoji = cryptoutilMagic.TestStatusEmojiSkip
	}

	// Only log to fixture if it exists
	if suite.fixture != nil {
		suite.fixture.logger.LogTestStepCompletion(statusEmoji, step.Name, result, step.Duration)
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
		if step.Status == cryptoutilMagic.TestStatusFail {
			statusEmoji = cryptoutilMagic.TestStatusEmojiFail
		} else if step.Status == cryptoutilMagic.TestStatusSkip {
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
	suite.fixture.logger.Log("%s", report.String())
}
