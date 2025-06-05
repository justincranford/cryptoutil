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

	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"
	cryptoutilServer "cryptoutil/internal/server/listener"

	joseJwe "github.com/lestrrat-go/jwx/v3/jwe"
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

// TODO Add asymmetric key pool tests
var happyPathTestCasesEncrypt = []TestCase{
	{name: "Key Pool E01", description: "Description 01", algorithm: "A256GCM/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E02", description: "Description 02", algorithm: "A192GCM/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E03", description: "Description 03", algorithm: "A128GCM/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E04", description: "Description 04", algorithm: "A192GCM/A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E05", description: "Description 05", algorithm: "A128GCM/A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E06", description: "Description 06", algorithm: "A128GCM/A128KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: "Key Pool E07", description: "Description 07", algorithm: "A256GCM/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E08", description: "Description 08", algorithm: "A192GCM/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E09", description: "Description 09", algorithm: "A128GCM/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E10", description: "Description 10", algorithm: "A192GCM/A192GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E11", description: "Description 11", algorithm: "A128GCM/A192GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E12", description: "Description 12", algorithm: "A128GCM/A128GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: "Key Pool E13", description: "Description 13", algorithm: "A256GCM/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E14", description: "Description 14", algorithm: "A192GCM/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E15", description: "Description 15", algorithm: "A128GCM/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: "Key Pool E16", description: "Description 16", algorithm: "A256CBC-HS512/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E17", description: "Description 17", algorithm: "A192CBC-HS384/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E18", description: "Description 18", algorithm: "A128CBC-HS256/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E19", description: "Description 19", algorithm: "A192CBC-HS384/A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E20", description: "Description 20", algorithm: "A128CBC-HS256/A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E21", description: "Description 21", algorithm: "A128CBC-HS256/A128KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: "Key Pool E22", description: "Description 22", algorithm: "A256CBC-HS512/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E23", description: "Description 23", algorithm: "A192CBC-HS384/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E24", description: "Description 24", algorithm: "A128CBC-HS256/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E25", description: "Description 25", algorithm: "A192CBC-HS384/A192GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E26", description: "Description 26", algorithm: "A128CBC-HS256/A192GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E27", description: "Description 27", algorithm: "A128CBC-HS256/A128GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: "Key Pool E28", description: "Description 28", algorithm: "A256CBC-HS512/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E29", description: "Description 29", algorithm: "A192CBC-HS384/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool E30", description: "Description 30", algorithm: "A128CBC-HS256/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
}

var happyPathTestCasesSign = []TestCase{
	{name: "Key Pool S01", description: "Description 01", algorithm: "RS256", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	// {name: "Key Pool S02", description: "Description 02", algorithm: "RS384", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	// {name: "Key Pool S03", description: "Description 03", algorithm: "RS512", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	// {name: "Key Pool S04", description: "Description 04", algorithm: "PS256", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	// {name: "Key Pool S05", description: "Description 05", algorithm: "PS384", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	// {name: "Key Pool S06", description: "Description 06", algorithm: "PS512", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	// {name: "Key Pool S07", description: "Description 07", algorithm: "ES256", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	// {name: "Key Pool S08", description: "Description 08", algorithm: "ES384", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	// {name: "Key Pool S09", description: "Description 09", algorithm: "ES512", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	// {name: "Key Pool S10", description: "Description 10", algorithm: "HS256", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	// {name: "Key Pool S11", description: "Description 11", algorithm: "HS384", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	// {name: "Key Pool S12", description: "Description 12", algorithm: "HS512", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	// {name: "Key Pool S13", description: "Description 13", algorithm: "EdDSA", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
}

var (
	testServerHostname = "localhost"
	testServerPort     = 8080
	testServerBaseUrl  = "http://" + testServerHostname + ":" + strconv.Itoa(testServerPort) + "/"
)

func TestMain(m *testing.M) {
	var rc int
	func() {
		start, stop, err := cryptoutilServer.NewHttpListener(testServerHostname, testServerPort, true)
		if err != nil {
			log.Fatalf("failed to create listener: %v", err)
		}
		go start()
		defer stop()
		WaitUntilReady(testServerBaseUrl, 5*time.Second, 100*time.Millisecond)

		rc = m.Run()
	}()
	os.Exit(rc)
}

func TestAllKeyPoolCipherAlgorithms(t *testing.T) {
	context := context.Background()
	openapiClient := RequireClientWithResponses(t, testServerBaseUrl)

	for i, testCase := range happyPathTestCasesEncrypt {
		testCaseNamePrefix := strings.ReplaceAll(testCase.algorithm, "/", "_")
		t.Run(testCaseNamePrefix, func(t *testing.T) {
			t.Parallel() // PostgreSQL supports N concurrent writers, SQLite supports 1 concurrent writer; concurrent perf is better with PostgreSQL
			var keyPool *cryptoutilOpenapiModel.KeyPool
			t.Run(testCaseNamePrefix+"  Create Key Pool", func(t *testing.T) {
				keyPoolCreate := RequireCreateKeyPoolRequest(t, testCase.name, testCase.description, testCase.algorithm, testCase.provider, testCase.exportAllowed, testCase.importAllowed, testCase.versioningAllowed)
				keyPool = RequireCreateKeyPoolResponse(t, context, openapiClient, keyPoolCreate)
				logObjectAsJson(t, keyPool)
			})
			if keyPool == nil {
				return
			}

			t.Run(testCaseNamePrefix+"  Generate Key", func(t *testing.T) {
				keyGenerate := RequireKeyGenerateRequest(t)
				key := RequireKeyGenerateResponse(t, context, openapiClient, keyPool.Id, keyGenerate)
				logObjectAsJson(t, key)
			})

			var cleartext *string
			var ciphertext *string
			t.Run(testCaseNamePrefix+"  Encrypt", func(t *testing.T) {
				str := "Hello World " + strconv.Itoa(i)
				cleartext = &str
				encryptRequest := RequireEncryptRequest(t, cleartext)
				ciphertext = RequireEncryptResponse(t, context, openapiClient, keyPool.Id, nil, encryptRequest)
				logJwe(t, ciphertext)
			})

			t.Run(testCaseNamePrefix+"  Generate Key", func(t *testing.T) {
				keyGenerate := RequireKeyGenerateRequest(t)
				key := RequireKeyGenerateResponse(t, context, openapiClient, keyPool.Id, keyGenerate)
				logObjectAsJson(t, key)
			})

			var decryptedtext *string
			t.Run(testCaseNamePrefix+"  Decrypt", func(t *testing.T) {
				decryptRequest := RequireDecryptRequest(t, ciphertext)
				decryptedtext = RequireDecryptResponse(t, context, openapiClient, keyPool.Id, decryptRequest)
			})

			require.NotNil(t, decryptedtext)
			require.Equal(t, *cleartext, *decryptedtext)
		})
	}
}

// func TestAllKeyPoolSignatureAlgorithms(t *testing.T) {
// 	context := context.Background()
// 	openapiClient := RequireClientWithResponses(t, testServerBaseUrl)

// 	for _, testCase := range happyPathTestCasesSign {
// 		testCaseNamePrefix := strings.ReplaceAll(testCase.algorithm, "/", "_")
// 		t.Run(testCaseNamePrefix, func(t *testing.T) {
// 			// t.Parallel() // PostgreSQL supports N concurrent writers, SQLite supports 1 concurrent writer; concurrent perf is better with PostgreSQL
// 			var keyPool *cryptoutilOpenapiModel.KeyPool
// 			t.Run(testCaseNamePrefix+"  Create Key Pool", func(t *testing.T) {
// 				keyPoolCreate := RequireCreateKeyPoolRequest(t, testCase.name, testCase.description, testCase.algorithm, testCase.provider, testCase.exportAllowed, testCase.importAllowed, testCase.versioningAllowed)
// 				keyPool = RequireCreateKeyPoolResponse(t, context, openapiClient, keyPoolCreate)
// 				logObjectAsJson(t, keyPool)
// 			})
// 			if keyPool == nil {
// 				return
// 			}

// 			t.Run(testCaseNamePrefix+"  Generate Key", func(t *testing.T) {
// 				keyGenerate := RequireKeyGenerateRequest(t)
// 				key := RequireKeyGenerateResponse(t, context, openapiClient, keyPool.Id, keyGenerate)
// 				logObjectAsJson(t, key)
// 			})

// 			// var cleartext *string
// 			// var ciphertext *string
// 			// t.Run(testCaseNamePrefix+"  Sign", func(t *testing.T) {
// 			// 	str := "Hello World " + strconv.Itoa(i)
// 			// 	cleartext = &str
// 			// 	encryptRequest := RequireEncryptRequest(t, cleartext)
// 			// 	ciphertext = RequireEncryptResponse(t, context, openapiClient, keyPool.Id, nil, encryptRequest)
// 			// 	logJwe(t, ciphertext)
// 			// })

// 			t.Run(testCaseNamePrefix+"  Generate Key", func(t *testing.T) {
// 				keyGenerate := RequireKeyGenerateRequest(t)
// 				key := RequireKeyGenerateResponse(t, context, openapiClient, keyPool.Id, keyGenerate)
// 				logObjectAsJson(t, key)
// 			})

// 			// var decryptedtext *string
// 			// t.Run(testCaseNamePrefix+"  Decrypt", func(t *testing.T) {
// 			// 	decryptRequest := RequireDecryptRequest(t, ciphertext)
// 			// 	decryptedtext = RequireDecryptResponse(t, context, openapiClient, keyPool.Id, decryptRequest)
// 			// })

// 			// require.NotNil(t, decryptedtext)
// 			// require.Equal(t, *cleartext, *decryptedtext)
// 		})
// 	}
// }

func logObjectAsJson(t *testing.T, object any) {
	jsonString, err := json.MarshalIndent(object, "", " ")
	require.NoError(t, err)
	t.Log(string(jsonString))
}

func logJwe(t *testing.T, encodedJweMessage *string) {
	jweMessage, err := joseJwe.Parse([]byte(*encodedJweMessage))
	require.NoError(t, err)
	logObjectAsJson(t, jweMessage)
}

// func logJws(t *testing.T, encodedJwsMessage *string) {
// 	jwsMessage, err := joseJws.Parse([]byte(*encodedJwsMessage))
// 	require.NoError(t, err)
// 	logObjectAsJson(t, jwsMessage)
// }
