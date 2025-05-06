package client

import (
	"context"
	cryptoutilOpenapiClient "cryptoutil/internal/openapi/client"
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func RequireClientWithResponses(t *testing.T, baseUrl string) *cryptoutilOpenapiClient.ClientWithResponses {
	openapiClient, err := cryptoutilOpenapiClient.NewClientWithResponses(baseUrl)
	require.NoError(t, err)
	require.NotNil(t, openapiClient)
	return openapiClient
}

func RequireCreateKeyPoolRequest(t *testing.T, name string, description string, algorithm string, provider string, exportAllowed bool, importAllowed bool, versioningAllowed bool) *cryptoutilOpenapiModel.KeyPoolCreate {
	openapiCreateKeyPoolRequest, err := MapKeyPoolCreate(name, description, algorithm, provider, exportAllowed, importAllowed, versioningAllowed)
	require.NotNil(t, openapiCreateKeyPoolRequest)
	require.NoError(t, err)
	return openapiCreateKeyPoolRequest
}

func RequireCreateKeyPoolResponse(t *testing.T, context context.Context, openapiClient *cryptoutilOpenapiClient.ClientWithResponses, keyPoolCreate *cryptoutilOpenapiModel.KeyPoolCreate) *cryptoutilOpenapiModel.KeyPool {
	openapiCreateKeyPoolResponse, err := openapiClient.PostKeypoolWithResponse(context, cryptoutilOpenapiClient.PostKeypoolJSONRequestBody(*keyPoolCreate))
	require.NoError(t, err)

	keyPool, err := MapKeyPool(openapiCreateKeyPoolResponse)
	require.NoError(t, err)
	require.NotNil(t, keyPool)

	err = ValidateCreateKeyPoolVsKeyPool(keyPoolCreate, keyPool)
	require.NoError(t, err)

	return keyPool
}

func ValidateCreateKeyPoolVsKeyPool(keyPoolCreate *cryptoutilOpenapiModel.KeyPoolCreate, keyPool *cryptoutilOpenapiModel.KeyPool) error {
	if keyPoolCreate == nil {
		return fmt.Errorf("key pool create is nil")
	} else if keyPool == nil {
		return fmt.Errorf("key pool is nil")
	} else if keyPool.Id == nil {
		return fmt.Errorf("key pool ID is nil")
	} else if keyPoolCreate.Name != *keyPool.Name {
		return fmt.Errorf("name mismatch: expected %s, got %s", keyPoolCreate.Name, *keyPool.Name)
	} else if keyPoolCreate.Description != *keyPool.Description {
		return fmt.Errorf("description mismatch: expected %s, got %s", keyPoolCreate.Description, *keyPool.Description)
	} else if *keyPoolCreate.Algorithm != *keyPool.Algorithm {
		return fmt.Errorf("algorithm mismatch: expected %s, got %s", *keyPoolCreate.Algorithm, *keyPool.Algorithm)
	} else if *keyPoolCreate.Provider != *keyPool.Provider {
		return fmt.Errorf("provider mismatch: expected %s, got %s", *keyPoolCreate.Provider, *keyPool.Provider)
	} else if *keyPoolCreate.ExportAllowed != *keyPool.ExportAllowed {
		return fmt.Errorf("exportAllowed mismatch: expected %t, got %t", *keyPoolCreate.ExportAllowed, *keyPool.ExportAllowed)
	} else if *keyPoolCreate.ImportAllowed != *keyPool.ImportAllowed {
		return fmt.Errorf("importAllowed mismatch: expected %t, got %t", *keyPoolCreate.ImportAllowed, *keyPool.ImportAllowed)
	} else if *keyPoolCreate.VersioningAllowed != *keyPool.VersioningAllowed {
		return fmt.Errorf("versioningAllowed mismatch: expected %t, got %t", *keyPoolCreate.VersioningAllowed, *keyPool.VersioningAllowed)
	} else if cryptoutilOpenapiModel.Active != *keyPool.Status {
		return fmt.Errorf("status mismatch: expected %s, got %s", cryptoutilOpenapiModel.Active, *keyPool.Status)
	}
	return nil
}
