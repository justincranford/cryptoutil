package client

import (
	"context"
	cryptoutilOpenapiClient "cryptoutil/internal/openapi/client"
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"
	cryptoutilServer "cryptoutil/internal/server/listener"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	start, stop, err := cryptoutilServer.NewHttpListener("localhost", 8080, true)
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	go start()
	defer stop()

	openapiClient := RequireClientWithResponses(t, "http://localhost:8080/")

	createdKeyPools := make([]*cryptoutilOpenapiModel.KeyPool, 0, 10)
	for i := range cap(createdKeyPools) {
		suffix := strconv.Itoa(i + 1)
		openapiCreateKeyPoolRequest := RequireCreateKeyPoolRequest(t, "Name "+suffix, "Description "+suffix, "A256GCM/A256KW", "Internal", false, false, true)
		openapiCreateKeyPoolResponse := RequireCreateKeyPoolResponse(t, context.Background(), openapiClient, openapiCreateKeyPoolRequest)
		createdKeyPools = append(createdKeyPools, openapiCreateKeyPoolResponse.JSON200)
	}

	t.Logf("Created %d key pools", len(createdKeyPools))
	for _, createdKeyPool := range createdKeyPools {
		responseJson, err := json.MarshalIndent(*createdKeyPool, "", "  ")
		require.NoError(t, err)
		t.Log(responseJson)
	}
}

func RequireClientWithResponses(t *testing.T, baseUrl string) *cryptoutilOpenapiClient.ClientWithResponses {
	openapiClient, err := cryptoutilOpenapiClient.NewClientWithResponses(baseUrl)
	require.NoError(t, err)
	require.NotNil(t, openapiClient)
	return openapiClient
}

func RequireCreateKeyPoolRequest(t *testing.T, name string, description string, algorithm string, provider string, exportAllowed bool, importAllowed bool, versioningAllowed bool) *cryptoutilOpenapiClient.PostKeypoolJSONRequestBody {
	openapiCreateKeyPoolRequest, err := mapKeyPoolCreate(t, name, description, algorithm, provider, exportAllowed, importAllowed, versioningAllowed)
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

func mapKeyPoolCreate(t *testing.T, name string, description string, algorithm string, provider string, exportAllowed bool, importAllowed bool, versioningAllowed bool) (*cryptoutilOpenapiClient.PostKeypoolJSONRequestBody, error) {
	keyPoolName, err1 := mapKeyPoolName(name)
	keyPoolDescription, err2 := mapKeyPoolDescription(description)
	keyPoolAlgorithm, err3 := mapKeyPoolAlgorithm(algorithm)
	keyPoolProvider, err4 := mapKeyPoolProvider(provider)
	keyPoolKeyPoolExportAllowed := mapKeyPoolExportAllowed(exportAllowed)
	keyPoolKeyPoolImportAllowed := mapKeyPoolImportAllowed(importAllowed)
	keyPoolKeyPoolVersioningAllowed := mapKeyPoolVersioningAllowed(versioningAllowed)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		t.Fatalf("failed to map key pool: %v", errors.Join(err1, err2, err3, err4))
	}
	return &cryptoutilOpenapiModel.KeyPoolCreate{
		Name:              *keyPoolName,
		Description:       *keyPoolDescription,
		Provider:          keyPoolProvider,
		Algorithm:         keyPoolAlgorithm,
		ExportAllowed:     keyPoolKeyPoolExportAllowed,
		ImportAllowed:     keyPoolKeyPoolImportAllowed,
		VersioningAllowed: keyPoolKeyPoolVersioningAllowed,
	}, nil
}

func mapKeyPoolName(name string) (*cryptoutilOpenapiModel.KeyPoolName, error) {
	if err := ValidateString(name); err != nil {
		return nil, fmt.Errorf("invalid key pool name: %w", err)
	}
	keyPoolName := cryptoutilOpenapiModel.KeyPoolName(name)
	return &keyPoolName, nil
}

func ValidateString(value string) error {
	length := len(value)
	trimmedLength := len(strings.TrimSpace(value))
	if length == 0 {
		return fmt.Errorf("string can't be empty")
	} else if trimmedLength == 0 {
		return fmt.Errorf("string can't contain all whitespace")
	} else if trimmedLength != length {
		return fmt.Errorf("string can't contain leading or trailing whitespace")
	}
	return nil
}

func mapKeyPoolDescription(description string) (*cryptoutilOpenapiModel.KeyPoolDescription, error) {
	if err := ValidateString(description); err != nil {
		return nil, fmt.Errorf("invalid key pool description: %w", err)
	}
	keyPoolDescription := cryptoutilOpenapiModel.KeyPoolDescription(description)
	return &keyPoolDescription, nil
}

func mapKeyPoolAlgorithm(algorithm string) (*cryptoutilOpenapiModel.KeyPoolAlgorithm, error) {
	if err := ValidateString(algorithm); err != nil {
		return nil, fmt.Errorf("invalid key pool algorithm: %w", err)
	}
	var keyPoolAlgorithm cryptoutilOpenapiModel.KeyPoolAlgorithm
	switch algorithm {
	case string(cryptoutilOpenapiModel.A128CBCHS256A128GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256A128GCMKW
	case string(cryptoutilOpenapiModel.A128CBCHS256A128KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256A128KW
	case string(cryptoutilOpenapiModel.A128CBCHS256A192GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256A192GCMKW
	case string(cryptoutilOpenapiModel.A128CBCHS256A192KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256A192KW
	case string(cryptoutilOpenapiModel.A128CBCHS256A256GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256A256GCMKW
	case string(cryptoutilOpenapiModel.A128CBCHS256A256KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256A256KW
	case string(cryptoutilOpenapiModel.A128CBCHS256Dir):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128CBCHS256Dir
	case string(cryptoutilOpenapiModel.A128GCMA128GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMA128GCMKW
	case string(cryptoutilOpenapiModel.A128GCMA128KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMA128KW
	case string(cryptoutilOpenapiModel.A128GCMA192GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMA192GCMKW
	case string(cryptoutilOpenapiModel.A128GCMA192KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMA192KW
	case string(cryptoutilOpenapiModel.A128GCMA256GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMA256GCMKW
	case string(cryptoutilOpenapiModel.A128GCMA256KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMA256KW
	case string(cryptoutilOpenapiModel.A128GCMDir):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A128GCMDir
	case string(cryptoutilOpenapiModel.A192CBCHS384A192GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192CBCHS384A192GCMKW
	case string(cryptoutilOpenapiModel.A192CBCHS384A192KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192CBCHS384A192KW
	case string(cryptoutilOpenapiModel.A192CBCHS384A256GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192CBCHS384A256GCMKW
	case string(cryptoutilOpenapiModel.A192CBCHS384A256KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192CBCHS384A256KW
	case string(cryptoutilOpenapiModel.A192CBCHS384Dir):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192CBCHS384Dir
	case string(cryptoutilOpenapiModel.A192GCMA192GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192GCMA192GCMKW
	case string(cryptoutilOpenapiModel.A192GCMA192KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192GCMA192KW
	case string(cryptoutilOpenapiModel.A192GCMA256GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192GCMA256GCMKW
	case string(cryptoutilOpenapiModel.A192GCMA256KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192GCMA256KW
	case string(cryptoutilOpenapiModel.A192GCMDir):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A192GCMDir
	case string(cryptoutilOpenapiModel.A256CBCHS512A256GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256CBCHS512A256GCMKW
	case string(cryptoutilOpenapiModel.A256CBCHS512A256KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256CBCHS512A256KW
	case string(cryptoutilOpenapiModel.A256CBCHS512Dir):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256CBCHS512Dir
	case string(cryptoutilOpenapiModel.A256GCMA256GCMKW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256GCMA256GCMKW
	case string(cryptoutilOpenapiModel.A256GCMA256KW):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256GCMA256KW
	case string(cryptoutilOpenapiModel.A256GCMDir):
		keyPoolAlgorithm = cryptoutilOpenapiModel.A256GCMDir
	default:
		return nil, fmt.Errorf("invalid key pool algorithm: %s", algorithm)
	}
	return &keyPoolAlgorithm, nil
}

func mapKeyPoolProvider(provider string) (*cryptoutilOpenapiModel.KeyPoolProvider, error) {
	if err := ValidateString(provider); err != nil {
		return nil, fmt.Errorf("invalid key pool provider: %w", err)
	}
	var keyPoolProvider cryptoutilOpenapiModel.KeyPoolProvider
	switch provider {
	case string(cryptoutilOpenapiModel.Internal):
		keyPoolProvider = cryptoutilOpenapiModel.Internal
	default:
		return nil, fmt.Errorf("invalid key pool provider: %s", provider)
	}
	return &keyPoolProvider, nil
}

func mapKeyPoolImportAllowed(importAllowed bool) *cryptoutilOpenapiModel.KeyPoolImportAllowed {
	keyPoolKeyPoolImportAllowed := cryptoutilOpenapiModel.KeyPoolImportAllowed(importAllowed)
	return &keyPoolKeyPoolImportAllowed
}

func mapKeyPoolExportAllowed(exportAllowed bool) *cryptoutilOpenapiModel.KeyPoolExportAllowed {
	keyPoolKeyPoolExportAllowed := cryptoutilOpenapiModel.KeyPoolExportAllowed(exportAllowed)
	return &keyPoolKeyPoolExportAllowed
}

func mapKeyPoolVersioningAllowed(versioningAllowed bool) *cryptoutilOpenapiModel.KeyPoolVersioningAllowed {
	keyPoolKeyPoolVersioningAllowed := cryptoutilOpenapiModel.KeyPoolVersioningAllowed(versioningAllowed)
	return &keyPoolKeyPoolVersioningAllowed
}
