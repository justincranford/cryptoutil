// Copyright (c) 2025 Justin Cranford

//go:build e2e && blocked

// Package test provides E2E tests for JOSE Authority.
// BLOCKED: Requires JOSE OpenAPI client generation (api/jose/).
// Remove 'blocked' build constraint after generating JOSE OpenAPI spec and client.
package test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	cryptoutilOpenapiModel "cryptoutil/api/model"
)

// TestJOSEWorkflow runs JOSE E2E test.
func TestJOSEWorkflow(t *testing.T) {
	suite.Run(t, new(JOSEWorkflowSuite))
}

// JOSEWorkflowSuite tests JOSE Authority JWT sign/verify operations.
type JOSEWorkflowSuite struct {
	suite.Suite
	fixture    *TestFixture
	assertions *ServiceAssertions
}

// SetupSuite runs once before all tests.
func (suite *JOSEWorkflowSuite) SetupSuite() {
	suite.fixture = NewTestFixture(suite.T())
	suite.assertions = NewServiceAssertions(suite.T(), suite.fixture.logger)
}

// TearDownSuite runs once after all tests.
func (suite *JOSEWorkflowSuite) TearDownSuite() {
	// Cleanup if needed.
}

// TestSignVerifyWorkflow tests complete JWT sign/verify cycle.
func (suite *JOSEWorkflowSuite) TestSignVerifyWorkflow() {
	ctx := context.Background()

	// Step 1: Deploy JOSE services
	suite.fixture.Setup()
	defer suite.fixture.Teardown()

	// Wait for JOSE health
	suite.T().Log("Waiting for JOSE service health checks...")
	err := suite.fixture.infraMgr.WaitForDockerServicesHealthy(ctx)
	suite.NoError(err, "JOSE services should become healthy")

	suite.T().Log("=== JOSE Sign/Verify Workflow E2E Test ===")

	// Step 2: Generate JWK (ES384 for signing)
	suite.T().Log("Step 1: Generating ES384 JWK for signing...")
	jwkID := suite.generateJWK(ctx, "ES384")

	// Step 3: Create JWT claims
	suite.T().Log("Step 2: Creating JWT with claims...")
	claims := map[string]any{
		"sub":   "user-" + googleUuid.NewString(),
		"name":  "Test User",
		"email": "test@example.com",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(1 * time.Hour).Unix(),
		"aud":   "jose-e2e-test",
		"iss":   "jose-authority",
	}

	// Step 4: Sign JWT
	suite.T().Log("Step 3: Signing JWT...")
	jws := suite.signJWT(ctx, jwkID, claims)
	suite.NotEmpty(jws, "JWS should not be empty")
	suite.True(strings.HasPrefix(jws, "eyJ"), "JWS should start with Base64URL header")

	// Step 5: Verify JWT signature
	suite.T().Log("Step 4: Verifying JWT signature...")
	verifiedClaims := suite.verifyJWT(ctx, jws)
	suite.NotNil(verifiedClaims, "Verified claims should not be nil")
	suite.Equal(claims["sub"], verifiedClaims["sub"], "Subject claim should match")
	suite.Equal(claims["email"], verifiedClaims["email"], "Email claim should match")

	// Step 6: Test invalid signature rejection
	suite.T().Log("Step 5: Testing invalid signature rejection...")
	tamperedJWS := jws[:len(jws)-10] + "TAMPERED00"
	suite.expectVerificationFailure(ctx, tamperedJWS)

	// Step 7: Test expired token rejection
	suite.T().Log("Step 6: Testing expired token rejection...")
	expiredClaims := map[string]any{
		"sub": "expired-user",
		"exp": time.Now().Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	}
	expiredJWS := suite.signJWT(ctx, jwkID, expiredClaims)
	suite.expectVerificationFailure(ctx, expiredJWS)

	// Step 8: Cleanup - Delete JWK
	suite.T().Log("Step 7: Cleaning up JWK...")
	suite.deleteJWK(ctx, jwkID)

	suite.T().Log("=== JOSE Sign/Verify Workflow E2E Test Complete ===")
}

// generateJWK creates an ES384 JWK for signing.
func (suite *JOSEWorkflowSuite) generateJWK(ctx context.Context, algorithm string) string {
	createReq := cryptoutilOpenapiModel.JoseJwkCreateRequest{
		Algorithm: &algorithm,
	}

	resp, err := suite.fixture.joseClient.PostJoseJwkWithResponse(ctx, createReq)
	suite.NoError(err, "JWK creation should succeed")
	suite.Equal(200, resp.StatusCode(), "JWK creation should return 200")
	suite.NotNil(resp.JSON200, "Response should contain JWK")

	kid := *resp.JSON200.Kid
	suite.T().Logf("Generated JWK: kid=%s, algorithm=%s", kid, algorithm)

	return kid
}

// signJWT signs claims into a JWS token.
func (suite *JOSEWorkflowSuite) signJWT(ctx context.Context, kid string, claims map[string]any) string {
	claimsJSON, err := json.Marshal(claims)
	suite.NoError(err, "Claims serialization should succeed")

	signReq := cryptoutilOpenapiModel.JoseJwsSignRequest{
		Kid:     &kid,
		Payload: string(claimsJSON),
	}

	resp, err := suite.fixture.joseClient.PostJoseJwsSignWithResponse(ctx, signReq)
	suite.NoError(err, "JWS sign should succeed")
	suite.Equal(200, resp.StatusCode(), "JWS sign should return 200")
	suite.NotNil(resp.JSON200, "Response should contain JWS")

	jws := *resp.JSON200.Jws
	suite.T().Logf("Signed JWT: %d bytes", len(jws))

	return jws
}

// verifyJWT verifies a JWS token and returns claims.
func (suite *JOSEWorkflowSuite) verifyJWT(ctx context.Context, jws string) map[string]any {
	verifyReq := cryptoutilOpenapiModel.JoseJwsVerifyRequest{
		Jws: &jws,
	}

	resp, err := suite.fixture.joseClient.PostJoseJwsVerifyWithResponse(ctx, verifyReq)
	suite.NoError(err, "JWS verify should succeed")
	suite.Equal(200, resp.StatusCode(), "JWS verify should return 200")
	suite.NotNil(resp.JSON200, "Response should contain verification result")
	suite.True(*resp.JSON200.Valid, "JWS signature should be valid")

	var claims map[string]any
	err = json.Unmarshal([]byte(*resp.JSON200.Payload), &claims)
	suite.NoError(err, "Claims deserialization should succeed")

	suite.T().Log("JWT verified successfully")

	return claims
}

// expectVerificationFailure verifies that invalid/expired JWS is rejected.
func (suite *JOSEWorkflowSuite) expectVerificationFailure(ctx context.Context, jws string) {
	verifyReq := cryptoutilOpenapiModel.JoseJwsVerifyRequest{
		Jws: &jws,
	}

	resp, err := suite.fixture.joseClient.PostJoseJwsVerifyWithResponse(ctx, verifyReq)
	suite.NoError(err, "Verify request should complete (even if signature invalid)")

	// Expect either 400 (bad request) or 200 with valid=false
	if resp.StatusCode() == 200 {
		suite.NotNil(resp.JSON200, "Response should contain verification result")
		suite.False(*resp.JSON200.Valid, "Invalid JWS should fail verification")
		suite.T().Log("Invalid JWS correctly rejected (valid=false)")
	} else if resp.StatusCode() == 400 {
		suite.T().Log("Invalid JWS correctly rejected (400 Bad Request)")
	} else {
		suite.Fail(fmt.Sprintf("Unexpected status code for invalid JWS: %d", resp.StatusCode()))
	}
}

// deleteJWK removes a JWK.
func (suite *JOSEWorkflowSuite) deleteJWK(ctx context.Context, kid string) {
	resp, err := suite.fixture.joseClient.DeleteJoseJwkKidWithResponse(ctx, kid)
	suite.NoError(err, "JWK deletion should succeed")
	suite.Equal(204, resp.StatusCode(), "JWK deletion should return 204")

	suite.T().Logf("Deleted JWK: kid=%s", kid)
}

// TestJWKSEndpointWorkflow tests JWKS discovery endpoint.
func (suite *JOSEWorkflowSuite) TestJWKSEndpointWorkflow() {
	suite.T().Skip("TODO P4.4: Implement JOSE JWKS endpoint E2E test")

	// TODO: Implement E2E test covering:
	// 1. Generate multiple JWKs (ES384, RS256)
	// 2. Fetch JWKS from /.well-known/jwks.json
	// 3. Verify public keys published correctly
	// 4. Verify key IDs (kid) match
	// 5. Verify private keys NOT exposed in JWKS
	// 6. Use JWKS public keys to verify JWTs
}

// TestJWKRotationWorkflow tests JWK rotation and backward compatibility.
func (suite *JOSEWorkflowSuite) TestJWKRotationWorkflow() {
	suite.T().Skip("TODO P4.4: Implement JOSE JWK rotation E2E test")

	// TODO: Implement E2E test covering:
	// 1. Generate JWK version 1
	// 2. Sign JWT with version 1
	// 3. Rotate to JWK version 2
	// 4. Sign new JWT with version 2
	// 5. Verify both JWTs with JWKS endpoint
	// 6. Verify old JWT still validates (backward compatibility)
	// 7. Verify new JWTs use version 2 kid
	// 8. Test JWKS contains both versions during rotation
}

// TestJWEEncryptionWorkflow tests JWE encryption and decryption.
func (suite *JOSEWorkflowSuite) TestJWEEncryptionWorkflow() {
	suite.T().Skip("TODO P4.4: Implement JOSE JWE encryption E2E test")

	// TODO: Implement E2E test covering:
	// 1. Generate encryption key (A256GCM)
	// 2. Create plaintext payload
	// 3. Encrypt payload as JWE
	// 4. Verify JWE structure (header, encrypted_key, iv, ciphertext, tag)
	// 5. Decrypt JWE
	// 6. Verify decrypted plaintext matches original
	// 7. Test with different encryption algorithms (A128GCM, A256GCM)
}
