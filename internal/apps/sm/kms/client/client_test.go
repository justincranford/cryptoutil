//go:build integration || e2e

// Copyright (c) 2025 Justin Cranford
//
// NOTE: These tests require a PostgreSQL database and are skipped in CI without the integration tag.
//

package client

import (
	"context"
	"crypto/x509"
	json "encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilServerApplication "cryptoutil/internal/apps/sm/kms/server/application"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilRandom "cryptoutil/internal/shared/util/random"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"

	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var (
	testSettings         = cryptoutilAppsTemplateServiceConfig.RequireNewForTest("application_test")
	testServerPublicURL  string
	testServerPrivateURL string
	testRootCAsPool      *x509.CertPool
)

func TestMain(m *testing.M) {
	var err error

	var startServerListenerApplication *cryptoutilServerApplication.ServerApplicationListener

	startServerListenerApplication, err = cryptoutilServerApplication.StartServerListenerApplication(testSettings)
	if err != nil {
		log.Fatalf("failed to start server application: %v", err)
	}

	go startServerListenerApplication.StartFunction()

	defer startServerListenerApplication.ShutdownFunction()

	// Build URLs using actual port bindings
	testServerPublicURL = testSettings.BindPublicProtocol + "://" + testSettings.BindPublicAddress + ":" + strconv.Itoa(int(startServerListenerApplication.ActualPublicPort))
	testServerPrivateURL = testSettings.BindPrivateProtocol + "://" + testSettings.BindPrivateAddress + ":" + strconv.Itoa(int(startServerListenerApplication.ActualPrivatePort))

	// Store the root CA pool for use in tests - use public server's pool since tests connect to public API
	testRootCAsPool = startServerListenerApplication.PublicTLSServer.RootCAsPool
	WaitUntilReady(&testServerPrivateURL, cryptoutilSharedMagic.TimeoutTestServerReady, cryptoutilSharedMagic.TimeoutTestServerReadyRetryDelay, startServerListenerApplication.PrivateTLSServer.RootCAsPool)

	os.Exit(m.Run())
}

type elasticKeyTestCase struct {
	name              string
	description       string
	algorithm         string
	provider          string
	importAllowed     bool
	versioningAllowed bool
}

func nextElasticKeyName() *string {
	// Use UUIDv7 for time-ordered uniqueness across concurrent test runs
	uniqueID := googleUuid.Must(googleUuid.NewV7()).String()
	nextElasticKeyName := fmt.Sprintf("Client Test Elastic Key %s", uniqueID)

	return &nextElasticKeyName
}

func nextElasticKeyDesc() *string {
	// Use UUIDv7 for time-ordered uniqueness across concurrent test runs
	uniqueID := googleUuid.Must(googleUuid.NewV7()).String()
	nextElasticKeyDesc := fmt.Sprintf("Client Test Elastic Key Description %s", uniqueID)

	return &nextElasticKeyDesc
}

var happyPathGenerateAlgorithmTestCases = []cryptoutilOpenapiModel.GenerateAlgorithm{
	// P0.4 optimization: Test only RSA2048 - RSA logic identical for all sizes
	cryptoutilOpenapiModel.RSA2048,
	cryptoutilOpenapiModel.ECP521,
	cryptoutilOpenapiModel.ECP384,
	cryptoutilOpenapiModel.ECP256,
	cryptoutilOpenapiModel.OKPEd25519,
	cryptoutilOpenapiModel.Oct512,
	cryptoutilOpenapiModel.Oct384,
	cryptoutilOpenapiModel.Oct256,
	cryptoutilOpenapiModel.Oct192,
	cryptoutilOpenapiModel.Oct128,
}

func generateHappyPathElasticKeyTestCasesEncrypt() []elasticKeyTestCase {
	return []elasticKeyTestCase{
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},

		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/A256GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/A256GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/A256GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/A192GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/A192GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/A192GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/A128GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/A128GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/A128GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},

		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/dir", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/dir", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/dir", provider: "Internal", importAllowed: false, versioningAllowed: true},

		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/RSA-OAEP-512", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/RSA-OAEP-512", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/RSA-OAEP-512", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/RSA-OAEP-384", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/RSA-OAEP-384", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/RSA-OAEP-384", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/RSA-OAEP-256", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/RSA-OAEP-256", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/RSA-OAEP-256", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/RSA-OAEP", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/RSA-OAEP", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/RSA-OAEP", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/RSA1_5", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/RSA1_5", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/RSA1_5", provider: "Internal", importAllowed: false, versioningAllowed: true},

		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/ECDH-ES+A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/ECDH-ES+A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/ECDH-ES+A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/ECDH-ES+A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/ECDH-ES+A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/ECDH-ES+A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/ECDH-ES+A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/ECDH-ES+A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/ECDH-ES+A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/ECDH-ES", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/ECDH-ES", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/ECDH-ES", provider: "Internal", importAllowed: false, versioningAllowed: true},

		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},

		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A256GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A256GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A256GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A192GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A192GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A192GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A128GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A128GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A128GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},

		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/dir", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/dir", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/dir", provider: "Internal", importAllowed: false, versioningAllowed: true},

		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/RSA-OAEP-512", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/RSA-OAEP-512", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/RSA-OAEP-512", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/RSA-OAEP-384", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/RSA-OAEP-384", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/RSA-OAEP-384", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/RSA-OAEP-256", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/RSA-OAEP-256", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/RSA-OAEP-256", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/RSA-OAEP", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/RSA-OAEP", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/RSA-OAEP", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/RSA1_5", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/RSA1_5", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/RSA1_5", provider: "Internal", importAllowed: false, versioningAllowed: true},

		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/ECDH-ES+A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/ECDH-ES+A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/ECDH-ES+A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/ECDH-ES+A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/ECDH-ES+A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/ECDH-ES+A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/ECDH-ES", provider: "Internal", importAllowed: false, versioningAllowed: true},
		// {name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/ECDH-ES", provider: "Internal", importAllowed: false, versioningAllowed: true},
		{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/ECDH-ES", provider: "Internal", importAllowed: false, versioningAllowed: true},
	}
}

var happyPathElasticKeyTestCasesSign = []elasticKeyTestCase{
	{name: "placeholder", description: "placeholder", algorithm: "RS256", provider: "Internal", importAllowed: false, versioningAllowed: true},
	// {name: "placeholder", description: "placeholder", algorithm: "RS384", provider: "Internal", importAllowed: false, versioningAllowed: true},
	// {name: "placeholder", description: "placeholder", algorithm: "RS512", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: "placeholder", description: "placeholder", algorithm: "PS256", provider: "Internal", importAllowed: false, versioningAllowed: true},
	// {name: "placeholder", description: "placeholder", algorithm: "PS384", provider: "Internal", importAllowed: false, versioningAllowed: true},
	// {name: "placeholder", description: "placeholder", algorithm: "PS512", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: "placeholder", description: "placeholder", algorithm: "ES256", provider: "Internal", importAllowed: false, versioningAllowed: true},
	// {name: "placeholder", description: "placeholder", algorithm: "ES384", provider: "Internal", importAllowed: false, versioningAllowed: true},
	// {name: "placeholder", description: "placeholder", algorithm: "ES512", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: "placeholder", description: "placeholder", algorithm: "HS256", provider: "Internal", importAllowed: false, versioningAllowed: true},
	// {name: "placeholder", description: "placeholder", algorithm: "HS384", provider: "Internal", importAllowed: false, versioningAllowed: true},
	// {name: "placeholder", description: "placeholder", algorithm: "HS512", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: "placeholder", description: "placeholder", algorithm: "EdDSA", provider: "Internal", importAllowed: false, versioningAllowed: true},
}

func TestAllElasticKeyCipherAlgorithms(t *testing.T) {
	t.Parallel() // PostgreSQL supports N concurrent writers, SQLite supports 1 concurrent writer; concurrent perf is better with PostgreSQL

	context := context.Background()
	testPublicServiceAPIUrl := testServerPublicURL + testSettings.PublicServiceAPIContextPath
	openapiClient := RequireClientWithResponses(t, &testPublicServiceAPIUrl, testRootCAsPool)

	testCases := generateHappyPathElasticKeyTestCasesEncrypt()

	for i, testCase := range testCases {
		testCaseNamePrefix := strings.ReplaceAll(testCase.algorithm, "/", "_")
		t.Run(testCaseNamePrefix, func(t *testing.T) {
			t.Parallel() // PostgreSQL supports N concurrent writers, SQLite supports 1 concurrent writer; concurrent perf is better with PostgreSQL

			// P1.14: Probabilistic execution based on P1.13 timing baseline analysis
			// High priority (>3.0s): TestProbTenth (10% execution)
			// Medium priority (1.5-3.0s): TestProbQuarter (25% execution)
			// Base algorithms (<1.5s): TestProbAlways (100% execution)
			switch testCase.algorithm {
			case "A128CBC-HS256/RSA1_5", "A128CBC-HS256/RSA-OAEP", "A128CBC-HS256/dir", "A128GCM/A128KW", "A128CBC-HS256/ECDH-ES", "A128CBC-HS256/ECDH-ES+A128KW", "A128CBC-HS256/RSA-OAEP-256":
				cryptoutilSharedUtilRandom.SkipByProbability(t, cryptoutilSharedMagic.TestProbTenth) // 10% execution for >3.0s tests
			case "A128CBC-HS256/A128GCMKW", "A128GCM/ECDH-ES", "A128CBC-HS256/A128KW", "A128GCM/ECDH-ES+A128KW", "A128GCM/RSA-OAEP", "A128GCM/RSA1_5", "A128GCM/A128GCMKW", "A128GCM/RSA-OAEP-256":
				cryptoutilSharedUtilRandom.SkipByProbability(t, cryptoutilSharedMagic.TestProbQuarter) // 25% execution for 1.5-3.0s tests
			default:
				cryptoutilSharedUtilRandom.SkipByProbability(t, cryptoutilSharedMagic.TestProbAlways) // 100% execution for <1.5s tests (base algorithms)
			}

			// Generate unique names per subtest to avoid UNIQUE constraint violations in concurrent tests
			uniqueName := nextElasticKeyName()
			uniqueDesc := nextElasticKeyDesc()

			var elasticKey *cryptoutilOpenapiModel.ElasticKey

			t.Run(testCaseNamePrefix+"  Create Elastic Key", func(t *testing.T) {
				elasticKeyCreate := RequireCreateElasticKeyRequest(t, uniqueName, uniqueDesc, &testCase.algorithm, &testCase.provider, &testCase.importAllowed, &testCase.versioningAllowed)
				elasticKey = RequireCreateElasticKeyResponse(context, t, openapiClient, elasticKeyCreate)
				logObjectAsJSON(t, elasticKey)
			})

			if elasticKey == nil {
				return
			}

			elasticKeyAlgorithm := cryptoutilOpenapiModel.ElasticKeyAlgorithm(testCase.algorithm)
			isAsymmetric, err := cryptoutilSharedCryptoJose.IsAsymmetric(&elasticKeyAlgorithm)
			require.NoError(t, err)

			t.Run(testCaseNamePrefix+"  Generate Key", func(t *testing.T) {
				keyGenerate := RequireMaterialKeyGenerateRequest(t)
				key := RequireMaterialKeyGenerateResponse(context, t, openapiClient, elasticKey.ElasticKeyID, keyGenerate)
				validateKeyResponsePublicKey(t, key, isAsymmetric)
				logObjectAsJSON(t, key)
			})

			var cleartext *string

			var ciphertext *string

			t.Run(testCaseNamePrefix+"  Encrypt", func(t *testing.T) {
				str := "Hello World " + strconv.Itoa(i)
				cleartext = &str
				encryptRequest := RequireEncryptRequest(t, cleartext)
				ciphertext = RequireEncryptResponse(context, t, openapiClient, elasticKey.ElasticKeyID, nil, encryptRequest)
				logJWE(t, ciphertext)
			})

			t.Run(testCaseNamePrefix+"  Generate Key", func(t *testing.T) {
				keyGenerate := RequireMaterialKeyGenerateRequest(t)
				key := RequireMaterialKeyGenerateResponse(context, t, openapiClient, elasticKey.ElasticKeyID, keyGenerate)
				validateKeyResponsePublicKey(t, key, isAsymmetric)
				logObjectAsJSON(t, key)
			})

			var decryptedtext *string

			t.Run(testCaseNamePrefix+"  Decrypt", func(t *testing.T) {
				decryptRequest := RequireDecryptRequest(t, ciphertext)
				decryptedtext = RequireDecryptResponse(context, t, openapiClient, elasticKey.ElasticKeyID, decryptRequest)
			})

			for _, generateAlgorithm := range happyPathGenerateAlgorithmTestCases {
				var generateDataKeyResponse *string

				algorithmSuffix := strings.ReplaceAll(((string)(generateAlgorithm)), "/", "_")
				t.Run(testCaseNamePrefix+"  Generate Data Key  "+algorithmSuffix, func(t *testing.T) {
					generateDataKeyParams := &cryptoutilOpenapiModel.GenerateParams{Context: nil, Alg: &generateAlgorithm}
					generateDataKeyResponse = RequireGenerateResponse(context, t, openapiClient, elasticKey.ElasticKeyID, generateDataKeyParams)
					logObjectAsJSON(t, generateDataKeyResponse)
				})

				var decryptedDataKey *string

				t.Run(testCaseNamePrefix+"  Decrypt Data Key  "+algorithmSuffix, func(t *testing.T) {
					decryptDataKeyRequest := RequireDecryptRequest(t, generateDataKeyResponse)
					decryptedDataKey = RequireDecryptResponse(context, t, openapiClient, elasticKey.ElasticKeyID, decryptDataKeyRequest)
					t.Log("decrypted data key", *decryptedDataKey)
				})

				t.Run(testCaseNamePrefix+"  Validate Data Key  "+algorithmSuffix, func(t *testing.T) {
					dataKeyJWK, err := joseJwk.ParseKey([]byte(*decryptedDataKey))
					require.NoError(t, err)
					require.NotNil(t, dataKeyJWK)
					logObjectAsJSON(t, dataKeyJWK)

					kidUUID, err := cryptoutilSharedCryptoJose.ExtractKidUUID(dataKeyJWK)
					require.NoError(t, err)
					require.NotNil(t, kidUUID)

					kty, err := cryptoutilSharedCryptoJose.ExtractKty(dataKeyJWK)
					require.NoError(t, err)
					require.NotNil(t, kty)
				})
			}

			require.NotNil(t, decryptedtext)
			require.Equal(t, *cleartext, *decryptedtext)
		})
	}
}

func TestAllElasticKeySignatureAlgorithms(t *testing.T) {
	t.Parallel() // PostgreSQL supports N concurrent writers, SQLite supports 1 concurrent writer; concurrent perf is better with PostgreSQL

	context := context.Background()
	testPublicServiceAPIUrl := testServerPublicURL + testSettings.PublicServiceAPIContextPath
	openapiClient := RequireClientWithResponses(t, &testPublicServiceAPIUrl, testRootCAsPool)

	for i, testCase := range happyPathElasticKeyTestCasesSign {
		testCaseNamePrefix := strings.ReplaceAll(testCase.algorithm, "/", "_")
		t.Run(testCaseNamePrefix, func(t *testing.T) {
			t.Parallel() // PostgreSQL supports N concurrent writers, SQLite supports 1 concurrent writer; concurrent perf is better with PostgreSQL

			var elasticKey *cryptoutilOpenapiModel.ElasticKey

			t.Run(testCaseNamePrefix+"  Create Elastic Key", func(t *testing.T) {
				uniqueName := nextElasticKeyName()
				uniqueDesc := nextElasticKeyDesc()
				elasticKeyCreate := RequireCreateElasticKeyRequest(t, uniqueName, uniqueDesc, &testCase.algorithm, &testCase.provider, &testCase.importAllowed, &testCase.versioningAllowed)
				elasticKey = RequireCreateElasticKeyResponse(context, t, openapiClient, elasticKeyCreate)
				logObjectAsJSON(t, elasticKey)
			})

			if elasticKey == nil {
				return
			}

			oamElasticKeyAlgorithm, err := cryptoutilSharedCryptoJose.ToElasticKeyAlgorithm(&testCase.algorithm)
			require.NoError(t, err)
			require.NotNil(t, oamElasticKeyAlgorithm)

			elasticKeyAlgorithm := cryptoutilOpenapiModel.ElasticKeyAlgorithm(testCase.algorithm)

			isAsymmetric, err := cryptoutilSharedCryptoJose.IsAsymmetric(&elasticKeyAlgorithm)
			require.NoError(t, err)

			t.Run(testCaseNamePrefix+"  Generate Key", func(t *testing.T) {
				keyGenerate := RequireMaterialKeyGenerateRequest(t)

				key := RequireMaterialKeyGenerateResponse(context, t, openapiClient, elasticKey.ElasticKeyID, keyGenerate)
				validateKeyResponsePublicKey(t, key, isAsymmetric)
				logObjectAsJSON(t, key)
			})

			var cleartext *string

			var signedtext *string

			t.Run(testCaseNamePrefix+"  Sign", func(t *testing.T) {
				str := "Hello World " + strconv.Itoa(i)
				cleartext = &str
				signRequest := RequireSignRequest(t, cleartext)
				signedtext = RequireSignResponse(context, t, openapiClient, elasticKey.ElasticKeyID, nil, signRequest)
				logJWS(t, signedtext)
			})

			t.Run(testCaseNamePrefix+"  Generate Key", func(t *testing.T) {
				keyGenerate := RequireMaterialKeyGenerateRequest(t)
				key := RequireMaterialKeyGenerateResponse(context, t, openapiClient, elasticKey.ElasticKeyID, keyGenerate)
				validateKeyResponsePublicKey(t, key, isAsymmetric)
				logObjectAsJSON(t, key)
			})

			var verifiedtest *string

			t.Run(testCaseNamePrefix+"  Verify", func(t *testing.T) {
				verifyRequest := RequireVerifyRequest(t, signedtext)
				verifiedtest = RequireVerifyResponse(context, t, openapiClient, elasticKey.ElasticKeyID, verifyRequest)
			})

			// Verify endpoint returns 204 No Content on success, so verifiedtest will be empty
			// The success of RequireVerifyResponse call indicates signature verification passed
			require.NotNil(t, verifiedtest)
			require.Empty(t, *verifiedtest)
		})
	}
}

func validateKeyResponsePublicKey(t *testing.T, key *cryptoutilOpenapiModel.MaterialKey, isAsymmetric bool) {
	t.Helper()

	if isAsymmetric {
		require.NotNil(t, key.ClearPublic)
		jwk, err := joseJwk.ParseKey([]byte(*key.ClearPublic))
		require.NoError(t, err)

		isPublic, err := cryptoutilSharedCryptoJose.IsPublicJWK(jwk)
		require.NoError(t, err)
		require.True(t, isPublic, "parsed JWK must be a public key")

		require.NotNil(t, jwk)
	} else {
		require.Nil(t, key.ClearPublic)
	}
}

func logObjectAsJSON(t *testing.T, object any) {
	t.Helper()

	jsonString, err := json.MarshalIndent(object, "", " ")
	require.NoError(t, err)
	t.Log(string(jsonString))
}

func logJWE(t *testing.T, encodedJWEMessage *string) {
	t.Helper()
	t.Log("JWE Message: {}", *encodedJWEMessage)

	jweMessage, err := joseJwe.Parse([]byte(*encodedJWEMessage))
	require.NoError(t, err)
	logObjectAsJSON(t, jweMessage)
}

func logJWS(t *testing.T, encodedJWSMessage *string) {
	t.Helper()
	t.Log("JWS Message: {}", *encodedJWSMessage)

	jwsMessage, err := joseJws.Parse([]byte(*encodedJWSMessage))
	require.NoError(t, err)
	logObjectAsJSON(t, jwsMessage)
}

// List and delete PEM files created during testing.
func TestCleanupTestCertificates(t *testing.T) {
	// List PEM files in the current package directory
	files, err := os.ReadDir(".")
	if err != nil {
		t.Logf("Warning: Could not read directory for PEM file cleanup: %v", err)

		return
	}

	var pemFiles []string

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".pem") {
			pemFiles = append(pemFiles, file.Name())
		}
	}

	// List the PEM files found
	if len(pemFiles) > 0 {
		t.Logf("Found PEM files in %s directory:", "internal/client")

		for _, pemFile := range pemFiles {
			t.Logf("  - %s", pemFile)
		}

		// Delete the PEM files
		for _, pemFile := range pemFiles {
			err := os.Remove(pemFile)
			require.NoError(t, err, "Failed to delete PEM file %s", pemFile)
			t.Logf("Successfully deleted PEM file: %s", pemFile)
		}
	} else {
		t.Logf("No PEM files found in %s directory", "internal/client")
	}
}
