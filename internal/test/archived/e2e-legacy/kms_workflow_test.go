// Copyright (c) 2025 Justin Cranford

//go:build e2e

package test

import (
	"context"
	"strings"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"

	"github.com/stretchr/testify/suite"
)

// TestKMSWorkflow runs KMS encrypt/decrypt E2E test.
func TestKMSWorkflow(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(KMSWorkflowSuite))
}

// KMSWorkflowSuite tests KMS key management and cryptographic operations.
type KMSWorkflowSuite struct {
	suite.Suite
	fixture    *TestFixture
	assertions *ServiceAssertions
}

// SetupSuite runs once before all tests.
func (suite *KMSWorkflowSuite) SetupSuite() {
	suite.fixture = NewTestFixture(suite.T())
	suite.assertions = NewServiceAssertions(suite.T(), suite.fixture.logger)
	// Deploy infrastructure once for all tests
	suite.fixture.Setup()
}

// TearDownSuite runs once after all tests.
func (suite *KMSWorkflowSuite) TearDownSuite() {
	// Teardown infrastructure after all tests complete
	suite.fixture.Teardown()
}

// TestEncryptDecryptWorkflow tests complete encrypt/decrypt cycle.
func (suite *KMSWorkflowSuite) TestEncryptDecryptWorkflow() {
	ctx := context.Background()

	// Wait for KMS health (infrastructure already set up in SetupSuite)
	suite.T().Log("Waiting for KMS service health checks...")
	err := suite.fixture.infraMgr.WaitForDockerServicesHealthy(ctx)
	suite.Require().NoError(err, "KMS service should be healthy")

	// Step 2: Create elastic key
	suite.T().Log("Creating elastic key...")

	elasticKeyName := cryptoutilOpenapiModel.ElasticKeyName("e2e-test-encrypt-key")
	elasticKeyAlg := cryptoutilOpenapiModel.ElasticKeyAlgorithm("A256GCM/A256KW")
	elasticKeyProvider := cryptoutilOpenapiModel.ElasticKeyProvider("Internal")
	elasticKeyDescription := cryptoutilOpenapiModel.ElasticKeyDescription("E2E test encryption key")

	createKeyReq := cryptoutilOpenapiModel.ElasticKeyCreate{
		Name:        elasticKeyName,
		Algorithm:   &elasticKeyAlg,
		Provider:    &elasticKeyProvider,
		Description: elasticKeyDescription,
	}

	createResp, err := suite.fixture.sqliteClient.PostElastickeyWithResponse(ctx, createKeyReq)
	suite.Require().NoError(err)
	suite.Require().Equal(200, createResp.StatusCode(), "Create elastic key should return 200")
	suite.Require().NotNil(createResp.JSON200)

	elasticKeyID := createResp.JSON200.ElasticKeyID
	suite.Require().NotNil(elasticKeyID)
	suite.T().Logf("Created elastic key: %s", *elasticKeyID)

	// Step 3: Generate material key (oct/256 via A256GCM/A256KW elastic key)
	suite.T().Log("Generating material key (oct/256 via A256GCM/A256KW elastic key)...")
	materialKeyReq := cryptoutilOpenapiModel.MaterialKeyGenerate{}
	genResp, err := suite.fixture.sqliteClient.PostElastickeyElasticKeyIDMaterialkeyWithResponse(ctx, *elasticKeyID, materialKeyReq)
	suite.Require().NoError(err)
	suite.Require().Equal(200, genResp.StatusCode(), "Generate material key should return 200")
	suite.Require().NotNil(genResp.JSON200)
	suite.T().Log("Generated material key (oct/256 via A256GCM/A256KW elastic key)")

	// Step 4: Encrypt plaintext
	suite.T().Log("Encrypting plaintext...")

	plaintext := "Hello World from E2E test!"
	plaintextBody := plaintext

	encryptResp, err := suite.fixture.sqliteClient.PostElastickeyElasticKeyIDEncryptWithTextBodyWithResponse(ctx, *elasticKeyID, nil, plaintextBody)
	suite.Require().NoError(err)
	suite.Require().Equal(200, encryptResp.StatusCode(), "Encrypt should return 200")
	suite.Require().NotNil(encryptResp.Body)

	ciphertext := string(encryptResp.Body)
	suite.Require().NotEmpty(ciphertext)
	suite.T().Logf("Encrypted data (JWE length: %d bytes)", len(ciphertext))

	// Verify JWE format (5 parts separated by dots)
	jweParts := strings.Split(ciphertext, ".")
	suite.Require().Equal(5, len(jweParts), "JWE should have 5 parts")

	// Step 5: Decrypt ciphertext
	suite.T().Log("Decrypting ciphertext...")
	decryptResp, err := suite.fixture.sqliteClient.PostElastickeyElasticKeyIDDecryptWithTextBodyWithResponse(ctx, *elasticKeyID, ciphertext)
	suite.Require().NoError(err)
	suite.Require().Equal(200, decryptResp.StatusCode(), "Decrypt should return 200")
	suite.Require().NotNil(decryptResp.Body)

	decrypted := string(decryptResp.Body)
	suite.Require().Equal(plaintext, decrypted, "Decrypted text should match original plaintext")
	suite.T().Logf("✅ Encryption/Decryption cycle successful: '%s' → (encrypted) → '%s'", plaintext, decrypted)
}

// TestSignVerifyWorkflow tests complete sign/verify cycle.
func (suite *KMSWorkflowSuite) TestSignVerifyWorkflow() {
	ctx := context.Background()

	// Step 1: Deploy KMS services
	suite.fixture.Setup()
	defer suite.fixture.Teardown()

	// Wait for KMS health
	suite.T().Log("Waiting for KMS service health checks...")
	err := suite.fixture.infraMgr.WaitForDockerServicesHealthy(ctx)
	suite.Require().NoError(err, "KMS service should be healthy")

	// Step 2: Create elastic key for signing
	suite.T().Log("Creating elastic key for signing...")

	elasticKeyName := cryptoutilOpenapiModel.ElasticKeyName("e2e-test-sign-key")
	elasticKeyAlg := cryptoutilOpenapiModel.ElasticKeyAlgorithm("ES384")
	elasticKeyProvider := cryptoutilOpenapiModel.ElasticKeyProvider("Internal")
	elasticKeyDescription := cryptoutilOpenapiModel.ElasticKeyDescription("E2E test signing key")

	createKeyReq := cryptoutilOpenapiModel.ElasticKeyCreate{
		Name:        elasticKeyName,
		Algorithm:   &elasticKeyAlg,
		Provider:    &elasticKeyProvider,
		Description: elasticKeyDescription,
	}

	createResp, err := suite.fixture.sqliteClient.PostElastickeyWithResponse(ctx, createKeyReq)
	suite.Require().NoError(err)
	suite.Require().Equal(200, createResp.StatusCode(), "Create elastic key should return 200")
	suite.Require().NotNil(createResp.JSON200)

	elasticKeyID := createResp.JSON200.ElasticKeyID
	suite.Require().NotNil(elasticKeyID)
	suite.T().Logf("Created elastic key: %s", *elasticKeyID)

	// Step 3: Generate material key (ECDSA P-384 via ES384 elastic key)
	suite.T().Log("Generating material key (ECDSA P-384 via ES384 elastic key)...")

	materialKeyReq := cryptoutilOpenapiModel.MaterialKeyGenerate{}
	genResp, err := suite.fixture.sqliteClient.PostElastickeyElasticKeyIDMaterialkeyWithResponse(ctx, *elasticKeyID, materialKeyReq)
	suite.Require().NoError(err)
	suite.Require().Equal(200, genResp.StatusCode(), "Generate material key should return 200")
	suite.Require().NotNil(genResp.JSON200)
	suite.T().Log("Generated material key (ECDSA P-384 via ES384 elastic key)")

	// Step 4: Sign payload
	suite.T().Log("Signing payload...")

	payload := "E2E test message to sign"
	payloadBody := payload

	signResp, err := suite.fixture.sqliteClient.PostElastickeyElasticKeyIDSignWithTextBodyWithResponse(ctx, *elasticKeyID, nil, payloadBody)
	suite.Require().NoError(err)
	suite.Require().Equal(200, signResp.StatusCode(), "Sign should return 200")
	suite.Require().NotNil(signResp.Body)

	signature := string(signResp.Body)
	suite.Require().NotEmpty(signature)
	suite.T().Logf("Signed data (JWS length: %d bytes)", len(signature))

	// Verify JWS format (3 parts separated by dots)
	jwsParts := strings.Split(signature, ".")
	suite.Require().Equal(3, len(jwsParts), "JWS should have 3 parts")

	// Step 5: Verify signature
	suite.T().Log("Verifying signature...")
	verifyResp, err := suite.fixture.sqliteClient.PostElastickeyElasticKeyIDVerifyWithTextBodyWithResponse(ctx, *elasticKeyID, signature)
	suite.Require().NoError(err)
	suite.Require().Equal(204, verifyResp.StatusCode(), "Verify should return 204 No Content on success")
	// Note: 204 No Content means signature verification succeeded, no body returned
	suite.T().Log("✅ Sign/Verify cycle successful: signature verified")

	// Step 6: Test invalid signature detection
	suite.T().Log("Testing invalid signature detection...")

	invalidSignature := strings.Replace(signature, jwsParts[2], "invalid_signature_part", 1)

	invalidVerifyResp, err := suite.fixture.sqliteClient.PostElastickeyElasticKeyIDVerifyWithTextBodyWithResponse(ctx, *elasticKeyID, invalidSignature)
	suite.Require().NoError(err)
	suite.Require().NotEqual(200, invalidVerifyResp.StatusCode(), "Invalid signature should not return 200")
	suite.T().Log("✅ Invalid signature correctly rejected")
}

// TestKeyRotationWorkflow tests key rotation and version management.
func (suite *KMSWorkflowSuite) TestKeyRotationWorkflow() {
	ctx := context.Background()

	// Wait for KMS health (infrastructure already set up in SetupSuite)
	suite.T().Log("Waiting for KMS service health checks...")
	err := suite.fixture.infraMgr.WaitForDockerServicesHealthy(ctx)
	suite.Require().NoError(err, "KMS service should be healthy")

	// Step 2: Create elastic key
	suite.T().Log("Creating elastic key for rotation test...")

	elasticKeyName := cryptoutilOpenapiModel.ElasticKeyName("e2e-test-rotation-key")
	elasticKeyAlg := cryptoutilOpenapiModel.ElasticKeyAlgorithm("A256GCM/A256KW")
	elasticKeyProvider := cryptoutilOpenapiModel.ElasticKeyProvider("Internal")
	elasticKeyDescription := cryptoutilOpenapiModel.ElasticKeyDescription("E2E test key rotation")

	createKeyReq := cryptoutilOpenapiModel.ElasticKeyCreate{
		Name:        elasticKeyName,
		Algorithm:   &elasticKeyAlg,
		Provider:    &elasticKeyProvider,
		Description: elasticKeyDescription,
	}

	createResp, err := suite.fixture.sqliteClient.PostElastickeyWithResponse(ctx, createKeyReq)
	suite.Require().NoError(err)
	suite.Require().Equal(200, createResp.StatusCode())

	elasticKeyID := createResp.JSON200.ElasticKeyID
	suite.T().Logf("Created elastic key: %s", *elasticKeyID)

	// Step 3: Generate material key version 1
	suite.T().Log("Generating material key version 1...")

	materialKeyReq1 := cryptoutilOpenapiModel.MaterialKeyGenerate{}
	genResp1, err := suite.fixture.sqliteClient.PostElastickeyElasticKeyIDMaterialkeyWithResponse(ctx, *elasticKeyID, materialKeyReq1)
	suite.Require().NoError(err)
	suite.Require().Equal(200, genResp1.StatusCode())

	// Step 4: Encrypt data with version 1
	suite.T().Log("Encrypting data with version 1...")

	plaintext1 := "Data encrypted with version 1"
	encryptResp1, err := suite.fixture.sqliteClient.PostElastickeyElasticKeyIDEncryptWithTextBodyWithResponse(ctx, *elasticKeyID, nil, plaintext1)
	suite.Require().NoError(err)
	suite.Require().Equal(200, encryptResp1.StatusCode())
	ciphertext1 := string(encryptResp1.Body)
	suite.T().Logf("Encrypted with v1 (length: %d bytes)", len(ciphertext1))

	// Step 5: Rotate key (create version 2)
	suite.T().Log("Rotating key (generating version 2)...")

	materialKeyReq2 := cryptoutilOpenapiModel.MaterialKeyGenerate{}
	genResp2, err := suite.fixture.sqliteClient.PostElastickeyElasticKeyIDMaterialkeyWithResponse(ctx, *elasticKeyID, materialKeyReq2)
	suite.Require().NoError(err)
	suite.Require().Equal(200, genResp2.StatusCode())
	suite.T().Log("✅ Key rotated - version 2 created")

	// Step 6: Encrypt new data with version 2
	suite.T().Log("Encrypting data with version 2...")

	plaintext2 := "Data encrypted with version 2"
	encryptResp2, err := suite.fixture.sqliteClient.PostElastickeyElasticKeyIDEncryptWithTextBodyWithResponse(ctx, *elasticKeyID, nil, plaintext2)
	suite.Require().NoError(err)
	suite.Require().Equal(200, encryptResp2.StatusCode())
	ciphertext2 := string(encryptResp2.Body)
	suite.T().Logf("Encrypted with v2 (length: %d bytes)", len(ciphertext2))

	// Step 7: Decrypt old data with version 1 (historical lookup via kid in JWE)
	suite.T().Log("Decrypting version 1 data (historical key lookup)...")
	decryptResp1, err := suite.fixture.sqliteClient.PostElastickeyElasticKeyIDDecryptWithTextBodyWithResponse(ctx, *elasticKeyID, ciphertext1)
	suite.Require().NoError(err)
	suite.Require().Equal(200, decryptResp1.StatusCode())
	decrypted1 := string(decryptResp1.Body)
	suite.Require().Equal(plaintext1, decrypted1)
	suite.T().Log("✅ Successfully decrypted v1 data with historical key")

	// Step 8: Decrypt new data with version 2 (latest key)
	suite.T().Log("Decrypting version 2 data (latest key)...")
	decryptResp2, err := suite.fixture.sqliteClient.PostElastickeyElasticKeyIDDecryptWithTextBodyWithResponse(ctx, *elasticKeyID, ciphertext2)
	suite.Require().NoError(err)
	suite.Require().Equal(200, decryptResp2.StatusCode())
	decrypted2 := string(decryptResp2.Body)
	suite.Require().Equal(plaintext2, decrypted2)
	suite.T().Log("✅ Successfully decrypted v2 data with latest key")

	suite.T().Log("✅ Key rotation workflow complete - both versions work correctly")
}
