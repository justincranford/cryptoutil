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

	cryptoutilJose "cryptoutil/internal/common/crypto/jose"
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"
	cryptoutilServerApplication "cryptoutil/internal/server/application"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
	joseJwk "github.com/lestrrat-go/jwx/v3/jwk"
	joseJws "github.com/lestrrat-go/jwx/v3/jws"

	"github.com/stretchr/testify/require"
)

type TestCase struct {
	name              string
	description       string
	algorithm         string
	provider          string
	exportAllowed     bool
	importAllowed     bool
	versioningAllowed bool
}

var uniqueTestNum = 0

func nextElasticKeyName() string {
	uniqueTestNum++
	return "Client Test Elastic Key " + strconv.Itoa(uniqueTestNum)
}
func nextElasticKeyDesc() string {
	return "Client Test Elastic Key Description" + strconv.Itoa(uniqueTestNum)
}

var happyPathTestCasesEncrypt = []TestCase{
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/A128KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/A128KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/A128KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/A192GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/A192GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/A192GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/A128GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/A128GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/A128GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/RSA-OAEP-512", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/RSA-OAEP-512", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/RSA-OAEP-512", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/RSA-OAEP-384", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/RSA-OAEP-384", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/RSA-OAEP-384", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/RSA-OAEP-256", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/RSA-OAEP-256", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/RSA-OAEP-256", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/RSA-OAEP", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/RSA-OAEP", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/RSA-OAEP", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/RSA1_5", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/RSA1_5", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/RSA1_5", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/ECDH-ES+A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/ECDH-ES+A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/ECDH-ES+A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/ECDH-ES+A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/ECDH-ES+A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/ECDH-ES+A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/ECDH-ES+A128KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/ECDH-ES+A128KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/ECDH-ES+A128KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256GCM/ECDH-ES", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192GCM/ECDH-ES", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128GCM/ECDH-ES", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A128KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A128KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A128KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A192GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A192GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A192GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256CBC-HS512/A128GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192CBC-HS384/A128GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/A128GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256CBC-HS512/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192CBC-HS384/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256CBC-HS512/RSA-OAEP-512", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192CBC-HS384/RSA-OAEP-512", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/RSA-OAEP-512", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256CBC-HS512/RSA-OAEP-384", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192CBC-HS384/RSA-OAEP-384", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/RSA-OAEP-384", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256CBC-HS512/RSA-OAEP-256", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192CBC-HS384/RSA-OAEP-256", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/RSA-OAEP-256", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256CBC-HS512/RSA-OAEP", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192CBC-HS384/RSA-OAEP", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/RSA-OAEP", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256CBC-HS512/RSA1_5", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192CBC-HS384/RSA1_5", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/RSA1_5", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256CBC-HS512/ECDH-ES+A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192CBC-HS384/ECDH-ES+A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/ECDH-ES+A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192CBC-HS384/ECDH-ES+A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/ECDH-ES+A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/ECDH-ES+A128KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A256CBC-HS512/ECDH-ES", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A192CBC-HS384/ECDH-ES", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "A128CBC-HS256/ECDH-ES", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
}

var happyPathTestCasesSign = []TestCase{
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "RS256", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "RS384", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "RS512", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "PS256", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "PS384", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "PS512", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "ES256", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "ES384", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "ES512", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "HS256", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "HS384", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "HS512", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: nextElasticKeyName(), description: nextElasticKeyDesc(), algorithm: "EdDSA", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
}

var (
	testServerHostname = "localhost"
	testServerPort     = 8080
	testServerBaseUrl  = "http://" + testServerHostname + ":" + strconv.Itoa(testServerPort) + "/"
)

func TestMain(m *testing.M) {
	var rc int
	func() {
		start, stop, err := cryptoutilServerApplication.StartServerApplication(testServerHostname, testServerPort, true)
		if err != nil {
			log.Fatalf("failed to start server application: %v", err)
		}
		go start()
		defer stop()
		WaitUntilReady(testServerBaseUrl, 5*time.Second, 100*time.Millisecond)

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestAllElasticKeyCipherAlgorithms(t *testing.T) {
	context := context.Background()
	openapiClient := RequireClientWithResponses(t, testServerBaseUrl)

	for i, testCase := range happyPathTestCasesEncrypt {
		testCaseNamePrefix := strings.ReplaceAll(testCase.algorithm, "/", "_")
		t.Run(testCaseNamePrefix, func(t *testing.T) {
			t.Parallel() // PostgreSQL supports N concurrent writers, SQLite supports 1 concurrent writer; concurrent perf is better with PostgreSQL
			var elasticKey *cryptoutilOpenapiModel.ElasticKey
			t.Run(testCaseNamePrefix+"  Create Elastic Key", func(t *testing.T) {
				elasticKeyCreate := RequireCreateElasticKeyRequest(t, testCase.name, testCase.description, testCase.algorithm, testCase.provider, testCase.exportAllowed, testCase.importAllowed, testCase.versioningAllowed)
				elasticKey = RequireCreateElasticKeyResponse(t, context, openapiClient, elasticKeyCreate)
				logObjectAsJson(t, elasticKey)
			})
			if elasticKey == nil {
				return
			}

			t.Run(testCaseNamePrefix+"  Generate Key", func(t *testing.T) {
				keyGenerate := RequireMaterialKeyGenerateRequest(t)
				key := RequireMaterialKeyGenerateResponse(t, context, openapiClient, elasticKey.ElasticKeyID, keyGenerate)
				logObjectAsJson(t, key)
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
				logObjectAsJson(t, key)
			})

			var decryptedtext *string
			t.Run(testCaseNamePrefix+"  Decrypt", func(t *testing.T) {
				decryptRequest := RequireDecryptRequest(t, ciphertext)
				decryptedtext = RequireDecryptResponse(t, context, openapiClient, elasticKey.ElasticKeyID, decryptRequest)
			})

			require.NotNil(t, decryptedtext)
			require.Equal(t, *cleartext, *decryptedtext)
		})
	}
}

func TestAllElasticKeySignatureAlgorithms(t *testing.T) {
	context := context.Background()
	openapiClient := RequireClientWithResponses(t, testServerBaseUrl)

	for i, testCase := range happyPathTestCasesSign {
		testCaseNamePrefix := strings.ReplaceAll(testCase.algorithm, "/", "_")
		t.Run(testCaseNamePrefix, func(t *testing.T) {
			// t.Parallel() // PostgreSQL supports N concurrent writers, SQLite supports 1 concurrent writer; concurrent perf is better with PostgreSQL
			var elasticKey *cryptoutilOpenapiModel.ElasticKey
			t.Run(testCaseNamePrefix+"  Create Elastic Key", func(t *testing.T) {
				elasticKeyCreate := RequireCreateElasticKeyRequest(t, testCase.name, testCase.description, testCase.algorithm, testCase.provider, testCase.exportAllowed, testCase.importAllowed, testCase.versioningAllowed)
				elasticKey = RequireCreateElasticKeyResponse(t, context, openapiClient, elasticKeyCreate)
				logObjectAsJson(t, elasticKey)
			})
			if elasticKey == nil {
				return
			}
			oamElasticKeyAlgorithm, err := cryptoutilJose.ToElasticKeyAlgorithm(testCase.algorithm)
			require.NoError(t, err)
			require.NotNil(t, oamElasticKeyAlgorithm)
			elasticKeyAlgorithm := cryptoutilOpenapiModel.ElasticKeyAlgorithm(testCase.algorithm)

			t.Run(testCaseNamePrefix+"  Generate Key", func(t *testing.T) {
				keyGenerate := RequireMaterialKeyGenerateRequest(t)
				key := RequireMaterialKeyGenerateResponse(t, context, openapiClient, elasticKey.ElasticKeyID, keyGenerate)
				isAsymmetric, err := cryptoutilJose.IsAsymmetric(&elasticKeyAlgorithm)
				require.NoError(t, err)
				if isAsymmetric {
					require.NotNil(t, key.ClearPublic)
					jwk, err := joseJwk.ParseKey([]byte(string(*key.ClearPublic)))
					require.NoError(t, err)
					require.NotNil(t, jwk)
					// TODO validate public key does not contain any private key or secret key material
				} else {
					require.Nil(t, key.ClearPublic)
				}
				logObjectAsJson(t, key)
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
				logObjectAsJson(t, key)
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

func logObjectAsJson(t *testing.T, object any) {
	jsonString, err := json.MarshalIndent(object, "", " ")
	require.NoError(t, err)
	t.Log(string(jsonString))
}

func logJwe(t *testing.T, encodedJweMessage *string) {
	t.Log("JWE Message: {}", *encodedJweMessage)

	jweMessage, err := joseJwe.Parse([]byte(*encodedJweMessage))
	require.NoError(t, err)
	logObjectAsJson(t, jweMessage)
}

func logJws(t *testing.T, encodedJwsMessage *string) {
	t.Log("JWS Message: {}", *encodedJwsMessage)

	jwsMessage, err := joseJws.Parse([]byte(*encodedJwsMessage))
	require.NoError(t, err)
	logObjectAsJson(t, jwsMessage)
}
