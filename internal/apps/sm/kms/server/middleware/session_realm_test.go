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
	"io"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSessionMiddleware_EmptyTokenAfterParse(t *testing.T) {
	t.Parallel()

	validator := &mockSessionValidator{
		serviceSession: &SessionInfo{
			SessionID: googleUuid.New(),
			TenantID:  googleUuid.New(),
		},
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(ServiceSessionMiddleware(validator))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Test with "Bearer " followed by empty string
	// Note: HTTP headers are trimmed, so this becomes just "Bearer"
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	// Header is trimmed to "Bearer" which doesn't have 2 parts after split
	require.Contains(t, string(body), "Invalid Authorization header format")
}
