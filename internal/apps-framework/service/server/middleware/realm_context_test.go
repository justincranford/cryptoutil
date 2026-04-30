// Copyright (c) 2025-2026 Justin Cranford.
//
//

package middleware

import (
	"context"
	"testing"

	googleUuid "github.com/google/uuid"

	"github.com/stretchr/testify/require"
)

func TestGetRealmContext_NotSet(t *testing.T) {
	t.Parallel()

	rc := GetRealmContext(context.Background())

	require.Nil(t, rc)
}

func TestGetRealmContext_Set(t *testing.T) {
	t.Parallel()

	tenantID := googleUuid.New()

	expected := &RealmContext{TenantID: tenantID, Source: "session"}

	ctx := context.WithValue(context.Background(), RealmContextKey{}, expected)

	got := GetRealmContext(ctx)

	require.NotNil(t, got)

	require.Equal(t, tenantID, got.TenantID)

	require.Equal(t, "session", got.Source)
}
