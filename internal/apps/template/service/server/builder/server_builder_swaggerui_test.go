// Copyright (c) 2025 Justin Cranford
//

package builder

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsTemplateServiceConfig "cryptoutil/internal/apps/template/service/config"
)

func TestWithSwaggerUI_EmptySpec(t *testing.T) {
	t.Parallel()

	b := &ServerBuilder{}

	result := b.WithSwaggerUI("user", "pass", nil)

	require.Same(t, b, result)
	require.Error(t, b.err)
	require.Contains(t, b.err.Error(), "cannot be empty")
}

func TestWithSwaggerUI_ValidSpec(t *testing.T) {
	t.Parallel()

	minCfg := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{}
	b := &ServerBuilder{config: minCfg}
	spec := []byte(`{"openapi":"3.0.3"}`)

	result := b.WithSwaggerUI("user", "pass", spec)

	require.Same(t, b, result)
	require.NoError(t, b.err)
	require.NotNil(t, b.swaggerUIConfig)
	require.Equal(t, "user", b.swaggerUIConfig.Username)
	require.Equal(t, "pass", b.swaggerUIConfig.Password)
	require.Equal(t, spec, b.swaggerUIConfig.OpenAPISpecJSON)
}

func TestWithSwaggerUI_PriorError(t *testing.T) {
	t.Parallel()

	priorErr := ErrBarrierConfigRequired
	b := &ServerBuilder{err: priorErr}
	spec := []byte(`{"openapi":"3.0.3"}`)

	result := b.WithSwaggerUI("user", "pass", spec)

	require.Same(t, b, result)
	require.ErrorIs(t, b.err, priorErr)
	require.Nil(t, b.swaggerUIConfig)
}

func TestWithSwaggerUI_NoAuth(t *testing.T) {
	t.Parallel()

	minCfg := &cryptoutilAppsTemplateServiceConfig.ServiceTemplateServerSettings{}
	b := &ServerBuilder{config: minCfg}
	spec := []byte(`{"openapi":"3.0.3"}`)

	result := b.WithSwaggerUI("", "", spec)

	require.Same(t, b, result)
	require.NoError(t, b.err)
	require.NotNil(t, b.swaggerUIConfig)
	require.Empty(t, b.swaggerUIConfig.Username)
	require.Empty(t, b.swaggerUIConfig.Password)
}
