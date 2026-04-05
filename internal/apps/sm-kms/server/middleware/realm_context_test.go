// Copyright (c) 2025 Justin Cranford
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
