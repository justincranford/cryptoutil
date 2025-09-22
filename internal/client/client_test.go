package client

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	cryptoutilOpenapiModel "cryptoutil/api/model"
	cryptoutilConfig "cryptoutil/internal/common/config"
	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilServerApplication "cryptoutil/internal/server/application"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"

	"github.com/stretchr/testify/require"
)

var (
	testSettings         = cryptoutilConfig.RequireNewForTest("client_test")
	testServerPublicURL  = testSettings.BindPublicProtocol + "://" + testSettings.BindPublicAddress + ":" + strconv.Itoa(int(testSettings.BindPublicPort))
	testServerPrivateURL = testSettings.BindPrivateProtocol + "://" + testSettings.BindPrivateAddress + ":" + strconv.Itoa(int(testSettings.BindPrivatePort))
)

func TestMain(m *testing.M) {
	var rc int
	func() {
		startServerListenerApplication, err := cryptoutilServerApplication.StartServerListenerApplication(testSettings)
		if err != nil {
			log.Fatalf("failed to start server application: %v", err)
		}
		go startServerListenerApplication.StartFunction()
		defer startServerListenerApplication.ShutdownFunction()
		WaitUntilReady(&testServerPrivateURL, 3*time.Second, 100*time.Millisecond, startServerListenerApplication.PrivateTLSServer.RootCAsPool)

		rc = m.Run()
	}()
	os.Exit(rc)
}

type elasticKeyTestCase struct {
	name              string
	description       string
	algorithm         string
	provider          string
	importAllowed     bool
	versioningAllowed bool
}

var uniqueElasticKeyTestNum = 0

func nextElasticKeyName() *string {
	uniqueElasticKeyTestNum++
	nextElasticKeyName := "Client Test Elastic Key " + strconv.Itoa(uniqueElasticKeyTestNum)
	return &nextElasticKeyName
}

func nextElasticKeyDesc() *string {
	nextElasticKeyDesc := "Client Test Elastic Key Description" + strconv.Itoa(uniqueElasticKeyTestNum)
	return &nextElasticKeyDesc
}

var happyPathGenerateAlgorithmTestCases = []cryptoutilOpenapiModel.GenerateAlgorithm{
	cryptoutilOpenapiModel.RSA4096,
	cryptoutilOpenapiModel.RSA3072,
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

var happyPathElasticKeyTestCasesEncrypt = []elasticKeyTestCase{
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},

	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/A256GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/A256GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/A256GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/A192GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/A192GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/A192GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/A128GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/A128GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/A128GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},

	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/dir", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/dir", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/dir", provider: "Internal", importAllowed: false, versioningAllowed: true},

	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/RSA-OAEP-512", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/RSA-OAEP-512", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/RSA-OAEP-512", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/RSA-OAEP-384", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/RSA-OAEP-384", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/RSA-OAEP-384", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/RSA-OAEP-256", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/RSA-OAEP-256", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/RSA-OAEP-256", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/RSA-OAEP", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/RSA-OAEP", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/RSA-OAEP", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/RSA1_5", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/RSA1_5", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/RSA1_5", provider: "Internal", importAllowed: false, versioningAllowed: true},

	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/ECDH-ES+A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/ECDH-ES+A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/ECDH-ES+A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/ECDH-ES+A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/ECDH-ES+A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/ECDH-ES+A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/ECDH-ES+A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/ECDH-ES+A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/ECDH-ES+A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256GCM/ECDH-ES", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192GCM/ECDH-ES", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128GCM/ECDH-ES", provider: "Internal", importAllowed: false, versioningAllowed: true},

	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},

	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A256GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A256GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A256GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A192GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A192GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A192GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A128GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A128GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A128GCMKW", provider: "Internal", importAllowed: false, versioningAllowed: true},

	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/dir", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/dir", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/dir", provider: "Internal", importAllowed: false, versioningAllowed: true},

	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/RSA-OAEP-512", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/RSA-OAEP-512", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/RSA-OAEP-512", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/RSA-OAEP-384", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/RSA-OAEP-384", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/RSA-OAEP-384", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/RSA-OAEP-256", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/RSA-OAEP-256", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/RSA-OAEP-256", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/RSA-OAEP", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/RSA-OAEP", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/RSA-OAEP", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/RSA1_5", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/RSA1_5", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/RSA1_5", provider: "Internal", importAllowed: false, versioningAllowed: true},

	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/ECDH-ES+A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/ECDH-ES+A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/ECDH-ES+A256KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/ECDH-ES+A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/ECDH-ES+A192KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/ECDH-ES+A128KW", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A256CBC-HS512/ECDH-ES", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A192CBC-HS384/ECDH-ES", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "A128CBC-HS256/ECDH-ES", provider: "Internal", importAllowed: false, versioningAllowed: true},
}

var happyPathElasticKeyTestCasesSign = []elasticKeyTestCase{
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "RS256", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "RS384", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "RS512", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "PS256", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "PS384", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "PS512", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "ES256", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "ES384", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "ES512", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "HS256", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "HS384", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "HS512", provider: "Internal", importAllowed: false, versioningAllowed: true},
	{name: *nextElasticKeyName(), description: *nextElasticKeyDesc(), algorithm: "EdDSA", provider: "Internal", importAllowed: false, versioningAllowed: true},
}

func TestAllElasticKeyCipherAlgorithms(t *testing.T) {
	context := context.Background()
	testPublicServiceAPIUrl := testServerPublicURL + testSettings.PublicServiceAPIContextPath
	openapiClient := RequireClientWithResponses(t, &testPublicServiceAPIUrl)

	for i, testCase := range happyPathElasticKeyTestCasesEncrypt {
		testCaseNamePrefix := strings.ReplaceAll(testCase.algorithm, "/", "_")
		t.Run(testCaseNamePrefix, func(t *testing.T) {
			t.Parallel() // PostgreSQL supports N concurrent writers, SQLite supports 1 concurrent writer; concurrent perf is better with PostgreSQL
			var elasticKey *cryptoutilOpenapiModel.ElasticKey
			t.Run(testCaseNamePrefix+"  Create Elastic Key", func(t *testing.T) {
				elasticKeyCreate := RequireCreateElasticKeyRequest(t, &testCase.name, &testCase.description, &testCase.algorithm, &testCase.provider, &testCase.importAllowed, &testCase.versioningAllowed)
				elasticKey = RequireCreateElasticKeyResponse(t, context, openapiClient, elasticKeyCreate)
				logObjectAsJSON(t, elasticKey)
			})
			if elasticKey == nil {
				return
			}

			t.Run(testCaseNamePrefix+"  Generate Key", func(t *testing.T) {
				keyGenerate := RequireMaterialKeyGenerateRequest(t)
				key := RequireMaterialKeyGenerateResponse(t, context, openapiClient, elasticKey.ElasticKeyID, keyGenerate)
				logObjectAsJSON(t, key)
			})

			var cleartext *string
			var ciphertext *string
			t.Run(testCaseNamePrefix+"  Encrypt", func(t *testing.T) {
				str := "Hello World " + strconv.Itoa(i)
				cleartext = &str
				encryptRequest := RequireEncryptRequest(t, cleartext)
				ciphertext = RequireEncryptResponse(t, context, openapiClient, elasticKey.ElasticKeyID, nil, encryptRequest)
				logJwe(t, ciphertext)
			})

			t.Run(testCaseNamePrefix+"  Generate Key", func(t *testing.T) {
				keyGenerate := RequireMaterialKeyGenerateRequest(t)
				key := RequireMaterialKeyGenerateResponse(t, context, openapiClient, elasticKey.ElasticKeyID, keyGenerate)
				logObjectAsJSON(t, key)
			})

			var decryptedtext *string
			t.Run(testCaseNamePrefix+"  Decrypt", func(t *testing.T) {
				decryptRequest := RequireDecryptRequest(t, ciphertext)
				decryptedtext = RequireDecryptResponse(t, context, openapiClient, elasticKey.ElasticKeyID, decryptRequest)
			})

			for _, generateAlgorithm := range happyPathGenerateAlgorithmTestCases {
				var generateDataKeyResponse *string
				algorithmSuffix := strings.ReplaceAll(((string)(generateAlgorithm)), "/", "_")
				t.Run(testCaseNamePrefix+"  Generate Data Key  "+algorithmSuffix, func(t *testing.T) {
					generateDataKeyParams := &cryptoutilOpenapiModel.GenerateParams{Context: nil, Alg: &generateAlgorithm}
					generateDataKeyResponse = RequireGenerateResponse(t, context, openapiClient, elasticKey.ElasticKeyID, generateDataKeyParams)
					logObjectAsJSON(t, generateDataKeyResponse)
				})

				var decryptedDataKey *string
				t.Run(testCaseNamePrefix+"  Decrypt Data Key  "+algorithmSuffix, func(t *testing.T) {
					decryptDataKeyRequest := RequireDecryptRequest(t, generateDataKeyResponse)
					decryptedDataKey = RequireDecryptResponse(t, context, openapiClient, elasticKey.ElasticKeyID, decryptDataKeyRequest)
					t.Log("decrypted data key", *decryptedDataKey)
				})

				t.Run(testCaseNamePrefix+"  Validate Data Key  "+algorithmSuffix, func(t *testing.T) {
					dataKeyJwk, err := joseJwk.ParseKey([]byte(*decryptedDataKey))
					require.NoError(t, err)
					require.NotNil(t, dataKeyJwk)
					logObjectAsJSON(t, dataKeyJwk)

					kidUUID, err := cryptoutilJose.ExtractKidUUID(dataKeyJwk)
					require.NoError(t, err)
					require.NotNil(t, kidUUID)

					kty, err := cryptoutilJose.ExtractKty(dataKeyJwk)
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
	context := context.Background()
	testPublicServiceAPIUrl := testServerPublicURL + testSettings.PublicServiceAPIContextPath
	openapiClient := RequireClientWithResponses(t, &testPublicServiceAPIUrl)

	for i, testCase := range happyPathElasticKeyTestCasesSign {
		testCaseNamePrefix := strings.ReplaceAll(testCase.algorithm, "/", "_")
		t.Run(testCaseNamePrefix, func(t *testing.T) {
			// t.Parallel() // PostgreSQL supports N concurrent writers, SQLite supports 1 concurrent writer; concurrent perf is better with PostgreSQL
			var elasticKey *cryptoutilOpenapiModel.ElasticKey
			t.Run(testCaseNamePrefix+"  Create Elastic Key", func(t *testing.T) {
				elasticKeyCreate := RequireCreateElasticKeyRequest(t, &testCase.name, &testCase.description, &testCase.algorithm, &testCase.provider, &testCase.importAllowed, &testCase.versioningAllowed)
				elasticKey = RequireCreateElasticKeyResponse(t, context, openapiClient, elasticKeyCreate)
				logObjectAsJSON(t, elasticKey)
			})
			if elasticKey == nil {
				return
			}
			oamElasticKeyAlgorithm, err := cryptoutilJose.ToElasticKeyAlgorithm(&testCase.algorithm)
			require.NoError(t, err)
			require.NotNil(t, oamElasticKeyAlgorithm)
			elasticKeyAlgorithm := cryptoutilOpenapiModel.ElasticKeyAlgorithm(testCase.algorithm)

			isAsymmetric, err := cryptoutilJose.IsAsymmetric(&elasticKeyAlgorithm)
			require.NoError(t, err)

			t.Run(testCaseNamePrefix+"  Generate Key", func(t *testing.T) {
				keyGenerate := RequireMaterialKeyGenerateRequest(t)
				key := RequireMaterialKeyGenerateResponse(t, context, openapiClient, elasticKey.ElasticKeyID, keyGenerate)
				if isAsymmetric {
					require.NotNil(t, key.ClearPublic)
					jwk, err := joseJwk.ParseKey([]byte(*key.ClearPublic))
					require.NoError(t, err)
					require.NotNil(t, jwk)
					// TODO validate public key does not contain any private key or secret key material
				} else {
					require.Nil(t, key.ClearPublic)
				}
				logObjectAsJSON(t, key)
			})
			var cleartext *string
			var signedtext *string
			t.Run(testCaseNamePrefix+"  Sign", func(t *testing.T) {
				str := "Hello World " + strconv.Itoa(i)
				cleartext = &str
				signRequest := RequireSignRequest(t, cleartext)
				signedtext = RequireSignResponse(t, context, openapiClient, elasticKey.ElasticKeyID, nil, signRequest)
				logJws(t, signedtext)
			})

			t.Run(testCaseNamePrefix+"  Generate Key", func(t *testing.T) {
				keyGenerate := RequireMaterialKeyGenerateRequest(t)
				key := RequireMaterialKeyGenerateResponse(t, context, openapiClient, elasticKey.ElasticKeyID, keyGenerate)
				logObjectAsJSON(t, key)
			})

			var verifiedtest *string
			t.Run(testCaseNamePrefix+"  Verify", func(t *testing.T) {
				verifyRequest := RequireVerifyRequest(t, signedtext)
				verifiedtest = RequireVerifyResponse(t, context, openapiClient, elasticKey.ElasticKeyID, verifyRequest)
			})

			require.NotNil(t, *verifiedtest)
		})
	}
}

func logObjectAsJSON(t *testing.T, object any) {
	jsonString, err := json.MarshalIndent(object, "", " ")
	require.NoError(t, err)
	t.Log(string(jsonString))
}

func logJwe(t *testing.T, encodedJweMessage *string) {
	t.Log("JWE Message: {}", *encodedJweMessage)

	jweMessage, err := joseJwe.Parse([]byte(*encodedJweMessage))
	require.NoError(t, err)
	logObjectAsJSON(t, jweMessage)
}

func logJws(t *testing.T, encodedJwsMessage *string) {
	t.Log("JWS Message: {}", *encodedJwsMessage)

	jwsMessage, err := joseJws.Parse([]byte(*encodedJwsMessage))
	require.NoError(t, err)
	logObjectAsJSON(t, jwsMessage)
}
