// Copyright (c) 2025 Justin Cranford
//
//

package demo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultDemoEndpoints(t *testing.T) {
	t.Parallel()

	endpoints := DefaultDemoEndpoints()
	require.NotNil(t, endpoints)
	require.NotEmpty(t, endpoints.TokenEndpoint)
	require.NotEmpty(t, endpoints.JWKSEndpoint)
	require.NotEmpty(t, endpoints.KMSAPIEndpoint)
	require.NotEmpty(t, endpoints.KMSHealthEndpoint)
}

func TestDefaultDemoCredentials(t *testing.T) {
	t.Parallel()

	creds := DefaultDemoCredentials()
	require.NotNil(t, creds)
	require.NotEmpty(t, creds.ClientID)
	require.NotEmpty(t, creds.ClientSecret)
	require.NotEmpty(t, creds.Scopes)
}

func TestNewDemoScript(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()
	script := NewDemoScript(config)
	require.NotNil(t, script)
	require.NotNil(t, script.endpoints)
	require.NotNil(t, script.credentials)
	require.NotNil(t, script.httpClient)
	require.NotNil(t, script.progress)
	require.NotNil(t, script.errors)
}
