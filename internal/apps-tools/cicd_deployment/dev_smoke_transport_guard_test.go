// Copyright (c) 2025-2026 Justin Cranford.

package cicd_deployment

import (
	"io"
	"testing"

	gen "cryptoutil/internal/apps-tools/deploymentgen"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentSmokeRequiresAllExplicitTransportsAndCapturedCompute(t *testing.T) {
	t.Parallel()
	err := executeDevelopmentSmokeWithOperations(t.Context(), developmentSmokeInput{}, nil, io.Discard, developmentSmokeOperations{})
	require.ErrorContains(t, err, "transports required")
	err = importDevelopmentImage(t.Context(), developmentSmokeInput{}, gen.LocalK3sPlan{}, "unvalidated-reference", nil)
	require.ErrorContains(t, err, "captured compute")
}
