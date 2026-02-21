// Copyright (c) 2025 Justin Cranford.
// SPDX-License-Identifier: Apache-2.0.

//go:build !integration

package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

// TestSessionMiddleware_EmptyBearerToken covers the defensive dead code path at
// session.go:75-79 where the token after "Bearer " is empty. This path cannot
// be triggered via real HTTP requests because Fiber trims trailing whitespace
// from headers, but is exercised via the injectable sessionMiddlewareStringsSplitNFn.
//
// NOTE: NOT parallel — modifies package-level injectable var.
func TestSessionMiddleware_EmptyBearerToken(t *testing.T) {
	// Override SplitN to simulate "Bearer " with empty token part.
	orig := sessionMiddlewareStringsSplitNFn
	sessionMiddlewareStringsSplitNFn = func(s, sep string, n int) []string {
		if strings.HasPrefix(strings.ToLower(s), "bearer") {
			return []string{"Bearer", ""}
		}

		return strings.SplitN(s, sep, n)
	}

	defer func() { sessionMiddlewareStringsSplitNFn = orig }()

	validator := &mockSessionValidator{}
	app := createTestApp()
	app.Get("/test", SessionMiddleware(validator, true), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer sometoken")
	resp, err := app.Test(req, -1)

	require.NoError(t, err)
	require.Equal(t, 401, resp.StatusCode)
}
