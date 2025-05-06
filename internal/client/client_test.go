package client

import (
	"context"
	cryptoutilOpenapiClient "cryptoutil/internal/openapi/client"
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"
	cryptoutilServer "cryptoutil/internal/server/listener"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestClient(t *testing.T) {
	start, stop, err := cryptoutilServer.NewHttpListener("localhost", 8080, true)
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	go start()
	defer stop()

	openapiClient := RequireClientWithResponses(t, "http://localhost:8080/")

	createdKeyPools := make([]*cryptoutilOpenapiModel.KeyPool, 0)
	for _, testCase := range happyPathTestCases {
		t.Run("Create Key Pool  "+strings.ReplaceAll(testCase.algorithm, "/", "_"), func(t *testing.T) {
			openapiCreateKeyPoolRequest := RequireCreateKeyPoolRequest(t, testCase.name, testCase.description, testCase.algorithm, testCase.provider, testCase.exportAllowed, testCase.importAllowed, testCase.versioningAllowed)
			openapiCreateKeyPoolResponse := RequireCreateKeyPoolResponse(t, context.Background(), openapiClient, openapiCreateKeyPoolRequest)
			createdKeyPool := openapiCreateKeyPoolResponse.JSON200
			createdKeyPools = append(createdKeyPools, createdKeyPool)
			responseJson, err := json.MarshalIndent(*createdKeyPool, "", " ")
			require.NoError(t, err)
			t.Log(string(responseJson))
		})
	}
	t.Logf("Created %d key pools", len(createdKeyPools))

}

func RequireClientWithResponses(t *testing.T, baseUrl string) *cryptoutilOpenapiClient.ClientWithResponses {
	openapiClient, err := cryptoutilOpenapiClient.NewClientWithResponses(baseUrl)
	require.NoError(t, err)
	require.NotNil(t, openapiClient)
	return openapiClient
}

func RequireCreateKeyPoolRequest(t *testing.T, name string, description string, algorithm string, provider string, exportAllowed bool, importAllowed bool, versioningAllowed bool) *cryptoutilOpenapiClient.PostKeypoolJSONRequestBody {
	openapiCreateKeyPoolRequest, err := MapKeyPoolCreate(t, name, description, algorithm, provider, exportAllowed, importAllowed, versioningAllowed)
	require.NotNil(t, openapiCreateKeyPoolRequest)
	require.NoError(t, err)
	return openapiCreateKeyPoolRequest
}

func RequireCreateKeyPoolResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, openapiCreateKeyPoolRequest *cryptoutilOpenapiClient.PostKeypoolJSONRequestBody) *cryptoutilOpenapiClient.PostKeypoolResponse {
	openapiCreateKeyPoolResponse, err := openapiClient.PostKeypoolWithResponse(context, cryptoutilOpenapiClient.PostKeypoolJSONRequestBody(*openapiCreateKeyPoolRequest))
	require.NoError(t, err)
	require.NotNil(t, openapiCreateKeyPoolResponse)
	require.NotNil(t, openapiCreateKeyPoolResponse.HTTPResponse)
	t.Logf("HTTP Response, Status: %v, Message: %s", openapiCreateKeyPoolResponse.HTTPResponse.StatusCode, openapiCreateKeyPoolResponse.HTTPResponse.Status)
	switch openapiCreateKeyPoolResponse.HTTPResponse.StatusCode {
	case 200:
		require.NotNil(t, openapiCreateKeyPoolResponse.Body)
		require.NotNil(t, openapiCreateKeyPoolResponse.JSON200)
		require.NotNil(t, openapiCreateKeyPoolResponse.JSON200.Id)
		require.NotNil(t, openapiCreateKeyPoolResponse.JSON200.Name)
		require.NotNil(t, openapiCreateKeyPoolResponse.JSON200.Description)
		require.NotNil(t, openapiCreateKeyPoolResponse.JSON200.Algorithm)
		require.NotNil(t, openapiCreateKeyPoolResponse.JSON200.Provider)
		require.NotNil(t, openapiCreateKeyPoolResponse.JSON200.ExportAllowed)
		require.NotNil(t, openapiCreateKeyPoolResponse.JSON200.ImportAllowed)
		require.NotNil(t, openapiCreateKeyPoolResponse.JSON200.VersioningAllowed)
		require.NotNil(t, openapiCreateKeyPoolResponse.JSON200.Status)
		require.Equal(t, openapiCreateKeyPoolRequest.Name, *openapiCreateKeyPoolResponse.JSON200.Name)
		require.Equal(t, openapiCreateKeyPoolRequest.Description, *openapiCreateKeyPoolResponse.JSON200.Description)
		require.Equal(t, *openapiCreateKeyPoolRequest.Algorithm, *openapiCreateKeyPoolResponse.JSON200.Algorithm)
		require.Equal(t, *openapiCreateKeyPoolRequest.Provider, *openapiCreateKeyPoolResponse.JSON200.Provider)
		require.Equal(t, *openapiCreateKeyPoolRequest.ExportAllowed, *openapiCreateKeyPoolResponse.JSON200.ExportAllowed)
		require.Equal(t, *openapiCreateKeyPoolRequest.ImportAllowed, *openapiCreateKeyPoolResponse.JSON200.ImportAllowed)
		require.Equal(t, *openapiCreateKeyPoolRequest.VersioningAllowed, *openapiCreateKeyPoolResponse.JSON200.VersioningAllowed)
		require.Equal(t, cryptoutilOpenapiModel.Active, *openapiCreateKeyPoolResponse.JSON200.Status)
	default:
		assert.FailNowf(t, "", "failed to create key pool, Status: %v, Message: %s", openapiCreateKeyPoolResponse.HTTPResponse.StatusCode, openapiCreateKeyPoolResponse.HTTPResponse.Status)
	}
	return openapiCreateKeyPoolResponse
}
