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

// Package middleware provides context types for KMS server route handlers.
package middleware

import (
	"context"

	googleUuid "github.com/google/uuid"
)

// RealmContextKey is the context key for realm context.
type RealmContextKey struct{}

// RealmContext holds tenant and realm information extracted from authentication.
type RealmContext struct {
	TenantID googleUuid.UUID

	RealmID googleUuid.UUID

	UserID googleUuid.UUID

	ClientID googleUuid.UUID

	Scopes []string

	Source string // "session", "jwt", "oidc", "header", "mtls"
}

// GetRealmContext extracts the RealmContext from a Go context.
// Returns nil if no realm context is set.
func GetRealmContext(ctx context.Context) *RealmContext {
	if ctx == nil {
		return nil
	}

	rc, _ := ctx.Value(RealmContextKey{}).(*RealmContext)

	return rc
}
