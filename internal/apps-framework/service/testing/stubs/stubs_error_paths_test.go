// Copyright (c) 2025-2026 Justin Cranford.

package stubs

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	cryptoutilAppsFrameworkServiceServer "cryptoutil/internal/apps-framework/service/server"
)

func TestApplicationOrFatal_ConstructionError(t *testing.T) {
	t.Parallel()

	injectedErr := errors.New("injected application construction error")

	var (
		fatalFormat string
		fatalArgs   []any
	)

	app := newTestApplication(
		t,
		func(
			context.Context,
			cryptoutilAppsFrameworkServiceServer.IPublicServer,
			cryptoutilAppsFrameworkServiceServer.IAdminServer,
		) (*cryptoutilAppsFrameworkServiceServer.Application, error) {
			return nil, injectedErr
		},
		func(format string, args ...any) {
			fatalFormat = format
			fatalArgs = args
		},
	)

	require.Nil(t, app)
	require.Contains(t, fatalFormat, "failed to create test application")
	require.Len(t, fatalArgs, 1)
	require.Equal(t, injectedErr, fatalArgs[0])
}
