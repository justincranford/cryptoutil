package client

import (
	"context"
	cryptoutilOpenapiClient "cryptoutil/internal/openapi/client"
	cryptoutilOpenapiModel "cryptoutil/internal/openapi/model"
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
