//go:build e2e

package test

import (
	"fmt"
	"time"

	cryptoutilOpenapiClient "cryptoutil/api/client"
	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilClient "cryptoutil/internal/client"
	cryptoutilMagic "cryptoutil/internal/common/magic"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// E2ETestSuite provides a structured test suite for end-to-end testing
type E2ETestSuite struct {
	suite.Suite
	fixture    *TestFixture
	assertions *ServiceAssertions
}

// SetupSuite runs once before all tests in the suite
func (suite *E2ETestSuite) SetupSuite() {
	fmt.Printf("[%s] [0s] 🚀 E2E TEST SUITE STARTING\n", time.Now().Format("15:04:05"))

	// Create test fixture
	suite.fixture = NewTestFixture(suite.T())

	// Create assertions helper
	suite.assertions = NewServiceAssertions(suite.T(), suite.fixture.startTime)

	// Setup infrastructure
	suite.fixture.Setup()

	fmt.Printf("[%s] [0s] ✅ E2E TEST SUITE SETUP COMPLETE\n", time.Now().Format("15:04:05"))
}

// TearDownSuite runs once after all tests in the suite
func (suite *E2ETestSuite) TearDownSuite() {
	fmt.Printf("[%s] [%v] 🧹 E2E TEST SUITE CLEANUP\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second))

	// Teardown infrastructure
	suite.fixture.Teardown()

	fmt.Printf("[%s] [%v] ✅ E2E TEST SUITE COMPLETED\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second))
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
	suite.assertions.AssertDockerServicesHealthy()
	suite.assertions.AssertHTTPReady(suite.fixture.ctx, suite.fixture.GetServiceURL("grafana")+"/api/health", cryptoutilMagic.TestTimeoutCryptoutilReady)
	suite.assertions.AssertHTTPReady(suite.fixture.ctx, suite.fixture.GetServiceURL("otel")+"/metrics", cryptoutilMagic.TestTimeoutCryptoutilReady)
}

// TestCryptoutilSQLite tests SQLite-based cryptoutil instance
func (suite *E2ETestSuite) TestCryptoutilSQLite() {
	suite.testCryptoutilInstance("sqlite")
}

// TestCryptoutilPostgres1 tests PostgreSQL-based cryptoutil instance #1
func (suite *E2ETestSuite) TestCryptoutilPostgres1() {
	suite.testCryptoutilInstance("postgres1")
}

// TestCryptoutilPostgres2 tests PostgreSQL-based cryptoutil instance #2
func (suite *E2ETestSuite) TestCryptoutilPostgres2() {
	suite.testCryptoutilInstance("postgres2")
}

// TestTelemetryFlow verifies telemetry is flowing correctly
func (suite *E2ETestSuite) TestTelemetryFlow() {
	suite.assertions.AssertTelemetryFlow(
		suite.fixture.ctx,
		suite.fixture.GetServiceURL("grafana"),
		suite.fixture.GetServiceURL("otel"),
	)
}

// testCryptoutilInstance tests a single cryptoutil instance
func (suite *E2ETestSuite) testCryptoutilInstance(instanceName string) {
	fmt.Printf("[%s] [%v] 🧪 Testing %s instance\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second),
		instanceName)

	client := suite.fixture.GetClient(instanceName)
	baseURL := suite.fixture.GetServiceURL(instanceName)

	// Test health check
	suite.assertions.AssertCryptoutilHealth(baseURL, suite.fixture.rootCAsPool)

	// Test core functionality
	elasticKey := suite.testCreateElasticKey(client)
	suite.testGenerateMaterialKey(client, elasticKey)
	suite.testEncryptDecryptCycle(client, elasticKey)
	suite.testSignVerifyCycle(client, elasticKey)

	fmt.Printf("[%s] [%v] ✅ %s instance tests passed\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second),
		instanceName)
}

// testCreateElasticKey creates a test elastic key
func (suite *E2ETestSuite) testCreateElasticKey(client *cryptoutilOpenapiClient.ClientWithResponses) *cryptoutilOpenapiModel.ElasticKey {
	fmt.Printf("[%s] [%v] 🔑 Creating elastic key\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second))

	elasticKeyCreate := cryptoutilClient.RequireCreateElasticKeyRequest(
		suite.T(), &testElasticKeyName, &testElasticKeyDescription,
		&testAlgorithm, &testProvider, &importAllowed, &versioningAllowed,
	)

	elasticKey := cryptoutilClient.RequireCreateElasticKeyResponse(suite.T(), suite.fixture.ctx, client, elasticKeyCreate)
	require.NotNil(suite.T(), elasticKey.ElasticKeyID)

	fmt.Printf("[%s] [%v] ✅ Elastic key created with ID: %s\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second),
		*elasticKey.ElasticKeyID)

	return elasticKey
}

// testGenerateMaterialKey generates a material key
func (suite *E2ETestSuite) testGenerateMaterialKey(client *cryptoutilOpenapiClient.ClientWithResponses, elasticKey *cryptoutilOpenapiModel.ElasticKey) {
	fmt.Printf("[%s] [%v] 🗝️  Generating material key\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second))

	keyGenerate := cryptoutilClient.RequireMaterialKeyGenerateRequest(suite.T())
	materialKey := cryptoutilClient.RequireMaterialKeyGenerateResponse(suite.T(), suite.fixture.ctx, client, elasticKey.ElasticKeyID, keyGenerate)
	require.NotNil(suite.T(), materialKey.MaterialKeyID)

	fmt.Printf("[%s] [%v] ✅ Material key generated with ID: %s\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second),
		materialKey.MaterialKeyID)
}

// testEncryptDecryptCycle tests full encrypt/decrypt cycle
func (suite *E2ETestSuite) testEncryptDecryptCycle(client *cryptoutilOpenapiClient.ClientWithResponses, elasticKey *cryptoutilOpenapiModel.ElasticKey) {
	fmt.Printf("[%s] [%v] 🔐 Testing encrypt/decrypt cycle\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second))

	// Encrypt
	encryptRequest := cryptoutilClient.RequireEncryptRequest(suite.T(), &cryptoutilMagic.TestCleartext)
	encryptedText := cryptoutilClient.RequireEncryptResponse(suite.T(), suite.fixture.ctx, client, elasticKey.ElasticKeyID, nil, encryptRequest)
	require.NotEmpty(suite.T(), *encryptedText)

	fmt.Printf("[%s] [%v] ✅ Text encrypted successfully\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second))

	// Decrypt
	decryptRequest := cryptoutilClient.RequireDecryptRequest(suite.T(), encryptedText)
	decryptedText := cryptoutilClient.RequireDecryptResponse(suite.T(), suite.fixture.ctx, client, elasticKey.ElasticKeyID, decryptRequest)
	require.Equal(suite.T(), cryptoutilMagic.TestCleartext, *decryptedText)

	fmt.Printf("[%s] [%v] ✅ Text decrypted successfully\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second))
}

// testSignVerifyCycle tests full sign/verify cycle
func (suite *E2ETestSuite) testSignVerifyCycle(client *cryptoutilOpenapiClient.ClientWithResponses, elasticKey *cryptoutilOpenapiModel.ElasticKey) {
	fmt.Printf("[%s] [%v] ✍️  Testing sign/verify cycle\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second))

	// Sign
	signRequest := cryptoutilClient.RequireSignRequest(suite.T(), &cryptoutilMagic.TestCleartext)
	signedText := cryptoutilClient.RequireSignResponse(suite.T(), suite.fixture.ctx, client, elasticKey.ElasticKeyID, nil, signRequest)
	require.NotEmpty(suite.T(), *signedText)

	fmt.Printf("[%s] [%v] ✅ Text signed successfully\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second))

	// Verify
	verifyRequest := cryptoutilClient.RequireVerifyRequest(suite.T(), signedText)
	verifyResponse := cryptoutilClient.RequireVerifyResponse(suite.T(), suite.fixture.ctx, client, elasticKey.ElasticKeyID, verifyRequest)
	require.Equal(suite.T(), "true", *verifyResponse)

	fmt.Printf("[%s] [%v] ✅ Signature verified successfully\n",
		time.Now().Format("15:04:05"),
		time.Since(suite.fixture.startTime).Round(time.Second))
}

// cleanupTestData cleans up any test data created during tests
func (suite *E2ETestSuite) cleanupTestData() {
	// This could include deleting test keys, clearing databases, etc.
	// Implementation depends on what test data is created
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
