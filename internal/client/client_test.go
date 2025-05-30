package client

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

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

var happyPathTestCases = []TestCase{
	{name: "Key Pool 01", description: "Description 01", algorithm: "A256GCM/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 02", description: "Description 02", algorithm: "A192GCM/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 03", description: "Description 03", algorithm: "A128GCM/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 04", description: "Description 04", algorithm: "A192GCM/A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 05", description: "Description 05", algorithm: "A128GCM/A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 06", description: "Description 06", algorithm: "A128GCM/A128KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: "Key Pool 07", description: "Description 07", algorithm: "A256GCM/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 08", description: "Description 08", algorithm: "A192GCM/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 09", description: "Description 09", algorithm: "A128GCM/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 10", description: "Description 10", algorithm: "A192GCM/A192GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 11", description: "Description 11", algorithm: "A128GCM/A192GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 12", description: "Description 12", algorithm: "A128GCM/A128GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: "Key Pool 13", description: "Description 13", algorithm: "A256GCM/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 14", description: "Description 14", algorithm: "A192GCM/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 15", description: "Description 15", algorithm: "A128GCM/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: "Key Pool 16", description: "Description 16", algorithm: "A256CBC-HS512/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 17", description: "Description 17", algorithm: "A192CBC-HS384/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 18", description: "Description 18", algorithm: "A128CBC-HS256/A256KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 19", description: "Description 19", algorithm: "A192CBC-HS384/A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 20", description: "Description 20", algorithm: "A128CBC-HS256/A192KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 21", description: "Description 21", algorithm: "A128CBC-HS256/A128KW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: "Key Pool 22", description: "Description 22", algorithm: "A256CBC-HS512/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 23", description: "Description 23", algorithm: "A192CBC-HS384/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 24", description: "Description 24", algorithm: "A128CBC-HS256/A256GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 25", description: "Description 25", algorithm: "A192CBC-HS384/A192GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 26", description: "Description 26", algorithm: "A128CBC-HS256/A192GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 27", description: "Description 27", algorithm: "A128CBC-HS256/A128GCMKW", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},

	{name: "Key Pool 28", description: "Description 28", algorithm: "A256CBC-HS512/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 29", description: "Description 29", algorithm: "A192CBC-HS384/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
	{name: "Key Pool 30", description: "Description 30", algorithm: "A128CBC-HS256/dir", provider: "Internal", exportAllowed: false, importAllowed: false, versioningAllowed: true},
}

var (
	testServerHostname = "localhost"
	testServerPort     = 8080
	testServerBaseUrl  = "http://" + testServerHostname + ":" + strconv.Itoa(testServerPort) + "/"
)

func TestMain(m *testing.M) {
	start, stop, err := cryptoutilServer.NewHttpListener(testServerHostname, testServerPort, true)
	if err != nil {
		log.Fatalf("failed to create listener: %v", err)
	}
	go start()
	defer stop()

	os.Exit(m.Run())
}

func TestAllKeyPoolAlgorithms(t *testing.T) {
	context := context.Background()
	openapiClient := RequireClientWithResponses(t, testServerBaseUrl)

	createdKeyPools := make([]*cryptoutilOpenapiModel.KeyPool, 0)
	createdKeys := make([]*cryptoutilOpenapiModel.Key, 0)
	cleartexts := make([]*string, 0)
	ciphertexts := make([]*string, 0)
	decryptedtexts := make([]*string, 0)
	for i, testCase := range happyPathTestCases {
		t.Run(strings.ReplaceAll(testCase.algorithm, "/", "_")+"  Create Key Pool", func(t *testing.T) {
			keyPoolCreate := RequireCreateKeyPoolRequest(t, testCase.name, testCase.description, testCase.algorithm, testCase.provider, testCase.exportAllowed, testCase.importAllowed, testCase.versioningAllowed)
			keyPool := RequireCreateKeyPoolResponse(t, context, openapiClient, keyPoolCreate)
			createdKeyPools = append(createdKeyPools, keyPool)
			logObjectAsJson(t, keyPool)
		})
		keyPool := createdKeyPools[i]

		t.Run(strings.ReplaceAll(testCase.algorithm, "/", "_")+"  Generate Key", func(t *testing.T) {
			keyGenerate := RequireKeyGenerateRequest(t)
			key := RequireKeyGenerateResponse(t, context, openapiClient, keyPool.Id, keyGenerate)
			createdKeys = append(createdKeys, key)
			logObjectAsJson(t, key)
		})
		// key := createdKeys[i]

		t.Run(strings.ReplaceAll(testCase.algorithm, "/", "_")+"  Encrypt", func(t *testing.T) {
			cleartext := "Hello World " + strconv.Itoa(i)
			cleartexts = append(cleartexts, &cleartext)
			encryptRequest := RequireEncryptRequest(t, &cleartext)
			ciphertext := RequireEncryptResponse(t, context, openapiClient, keyPool.Id, nil, encryptRequest)
			ciphertexts = append(ciphertexts, ciphertext)
			logJwe(t, ciphertext)
		})
		cleartext := cleartexts[i]
		ciphertext := ciphertexts[i]

		t.Run(strings.ReplaceAll(testCase.algorithm, "/", "_")+"  Decrypt", func(t *testing.T) {
			decryptRequest := RequireDecryptRequest(t, ciphertext)
			decryptedtext := RequireDecryptResponse(t, context, openapiClient, keyPool.Id, decryptRequest)
			decryptedtexts = append(decryptedtexts, decryptedtext)
		})
		decryptedtext := decryptedtexts[i]

		require.NotNil(t, decryptedtext)
		require.Equal(t, *decryptedtext, *cleartext)
	}
	t.Logf("Created %d key pools, %d keys, %d cleartexts, %d ciphertexts, %d decryptedtexts", len(createdKeyPools), len(createdKeys), len(cleartexts), len(createdKeys), len(decryptedtexts))
}

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
