// Copyright (c) 2025 Justin Cranford

//go:build e2e

// Package test provides E2E tests for JOSE Authority.
package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	cryptoutilJOSEClient "cryptoutil/api/jose/client"
)

// TestJOSEWorkflow runs JOSE E2E test.
func TestJOSEWorkflow(t *testing.T) {
	t.Parallel()
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
	t := suite.T()
	t.Skip("TODO P4: Implement sign/verify handlers (currently return 501 Not Implemented) and ensure Docker Desktop running")

	ctx := context.Background()

	// 1. Deploy JOSE E2E service from compose
	suite.fixture.Setup()
	defer suite.fixture.Teardown()

	// 2. Generate JWK (ES384) via GenerateJWKWithResponse
	algorithm := cryptoutilJOSEClient.JWKGenerateRequestAlgorithm("ES384")
	use := cryptoutilJOSEClient.JWKGenerateRequestUse("sig")
	genReq := cryptoutilJOSEClient.JWKGenerateRequest{
		Algorithm: algorithm,
		Use:       &use,
	}
	genResp, err := suite.fixture.GetJOSEClient().GenerateJWKWithResponse(ctx, genReq)
	require.NoError(t, err, "Failed to generate JWK")
	require.NotNil(t, genResp.JSON201, "Generate JWK response should be 201 Created")
	kid := genResp.JSON201.Kid
	require.NotEmpty(t, kid, "Generated kid should not be empty")

	// 3. Create JWT claims (sub, name, email, iat, exp, aud, iss)
	// exp set to 1 year from now for valid token test
	iat := time.Now().UTC().Unix()
	exp := time.Now().UTC().Add(365 * 24 * time.Hour).Unix()
	claims := fmt.Sprintf(
		`{"sub":"user-123","name":"Test User","email":"test@example.com","iat":%d,"exp":%d,"aud":"test-audience","iss":"jose-e2e-test"}`,
		iat, exp,
	)

	// 4. Sign JWT via SignJWSWithResponse
	signReq := cryptoutilJOSEClient.JWSSignRequest{
		Kid:     kid,
		Payload: claims,
	}
	signResp, err := suite.fixture.GetJOSEClient().SignJWSWithResponse(ctx, signReq)
	require.NoError(t, err, "Failed to sign JWS")
	require.NotNil(t, signResp.JSON200, "Sign JWS response should be 200 OK")
	jws := signResp.JSON200.JWS
	require.NotEmpty(t, jws, "Generated JWS should not be empty")

	// 5. Verify JWT signature via VerifyJWSWithResponse
	verifyReq := cryptoutilJOSEClient.JWSVerifyRequest{JWS: jws}
	verifyResp, err := suite.fixture.GetJOSEClient().VerifyJWSWithResponse(ctx, verifyReq)
	require.NoError(t, err, "Failed to verify JWS")
	require.NotNil(t, verifyResp.JSON200, "Verify JWS response should be 200 OK")
	require.True(t, verifyResp.JSON200.Valid, "JWS should be valid")

	// 6. Test invalid signature rejection (tamper JWS, expect verification failure)
	// Tamper with JWS by changing last character
	tamperedJWS := jws[:len(jws)-1] + "X"
	tamperedVerifyReq := cryptoutilJOSEClient.JWSVerifyRequest{JWS: tamperedJWS}
	tamperedVerifyResp, err := suite.fixture.GetJOSEClient().VerifyJWSWithResponse(ctx, tamperedVerifyReq)
	require.NoError(t, err, "Verify request should not error (HTTP layer)")
	// Expect either 400 Bad Request or 200 with Valid=false
	if tamperedVerifyResp.JSON200 != nil {
		require.False(t, tamperedVerifyResp.JSON200.Valid, "Tampered JWS should be invalid")
	} else {
		require.NotNil(t, tamperedVerifyResp.JSON400, "Tampered JWS should return 400 Bad Request")
	}

	// 7. Test expired token rejection (set exp in past, expect validation error)
	expiredExp := time.Now().UTC().Add(-1 * time.Hour).Unix() // 1 hour ago
	expiredClaims := fmt.Sprintf(
		`{"sub":"user-456","name":"Expired User","email":"expired@example.com","iat":%d,"exp":%d,"aud":"test-audience","iss":"jose-e2e-test"}`,
		iat, expiredExp,
	)
	expiredSignReq := cryptoutilJOSEClient.JWSSignRequest{
		Kid:     kid,
		Payload: expiredClaims,
	}
	expiredSignResp, err := suite.fixture.GetJOSEClient().SignJWSWithResponse(ctx, expiredSignReq)
	require.NoError(t, err, "Failed to sign expired JWS")
	require.NotNil(t, expiredSignResp.JSON200, "Sign expired JWS response should be 200 OK")
	expiredJWS := expiredSignResp.JSON200.JWS

	expiredVerifyReq := cryptoutilJOSEClient.JWSVerifyRequest{JWS: expiredJWS}
	expiredVerifyResp, err := suite.fixture.GetJOSEClient().VerifyJWSWithResponse(ctx, expiredVerifyReq)
	require.NoError(t, err, "Verify request should not error (HTTP layer)")
	// Expect either 400 Bad Request or 200 with Valid=false for expired token
	if expiredVerifyResp.JSON200 != nil {
		require.False(t, expiredVerifyResp.JSON200.Valid, "Expired JWS should be invalid")
	} else {
		require.NotNil(t, expiredVerifyResp.JSON400, "Expired JWS should return 400 Bad Request")
	}

	// 8. Cleanup: Delete JWK via DeleteJWKWithResponse
	deleteResp, err := suite.fixture.GetJOSEClient().DeleteJWKWithResponse(ctx, kid)
	require.NoError(t, err, "Failed to delete JWK")
	require.Equal(t, 204, deleteResp.StatusCode(), "Delete JWK response should be 204 No Content")
	// Example usage:
	//   ctx := context.Background()
	//   suite.fixture.Setup()
	//   defer suite.fixture.Teardown()
	//
	//   genReq := cryptoutilJOSEClient.JWKGenerateRequest{
	//     Alg: ptr("ES384"),
	//     Use: ptr("sig"),
	//   }
	//   genResp, _ := suite.fixture.GetJOSEClient().GenerateJWKWithResponse(ctx, genReq)
	//   kid := genResp.JSON201.Kid
	//
	//   signReq := cryptoutilJOSEClient.JWSSignRequest{
	//     Kid:     kid,
	//     Payload: `{"sub":"user-123","exp":1735689600}`,
	//   }
	//   signResp, _ := suite.fixture.GetJOSEClient().SignJWSWithResponse(ctx, signReq)
	//   jws := signResp.JSON201.Jws
	//
	//   verifyReq := cryptoutilJOSEClient.JWSVerifyRequest{Jws: jws}
	//   verifyResp, _ := suite.fixture.GetJOSEClient().VerifyJWSWithResponse(ctx, verifyReq)
	//   assert(verifyResp.JSON200.Valid == true)
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
