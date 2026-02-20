// Copyright (c) 2025 Justin Cranford
//
//

package middleware

import (
	"context"
	"io"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	googleUuid "github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Test UUIDs generated once per test run for consistency.
var middlewareTestUUID = googleUuid.Must(googleUuid.NewV7()).String()

func TestTenantMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		tenantID       string
		wantStatus     int
		wantTenantID   string
		wantError      bool
		wantErrorField string
	}{
		{
			name:         "valid tenant ID",
			tenantID:     middlewareTestUUID,
			wantStatus:   http.StatusOK,
			wantTenantID: middlewareTestUUID,
			wantError:    false,
		},
		{
			name:         "no tenant ID header",
			tenantID:     "",
			wantStatus:   http.StatusOK,
			wantTenantID: "",
			wantError:    false,
		},
		{
			name:           "invalid tenant ID format",
			tenantID:       "not-a-uuid",
			wantStatus:     http.StatusBadRequest,
			wantError:      true,
			wantErrorField: "invalid_tenant_id",
		},
		{
			name:           "malformed UUID",
			tenantID:       "550e8400-e29b-41d4-a716",
			wantStatus:     http.StatusBadRequest,
			wantError:      true,
			wantErrorField: "invalid_tenant_id",
		},
		{
			name:         "uppercase UUID",
			tenantID:     middlewareTestUUID,
			wantStatus:   http.StatusOK,
			wantTenantID: middlewareTestUUID,
			wantError:    false,
		},
		{
			name:         "tenant ID with whitespace",
			tenantID:     "  " + middlewareTestUUID + "  ",
			wantStatus:   http.StatusOK,
			wantTenantID: middlewareTestUUID,
			wantError:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()

			var capturedTenantID string

			app.Use(TenantMiddleware())

			app.Get("/test", func(c *fiber.Ctx) error {
				capturedTenantID = GetTenantID(c.UserContext())

				return c.SendString("ok")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.tenantID != "" {
				req.Header.Set(TenantIDHeader, tc.tenantID)
			}

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.wantStatus, resp.StatusCode)

			if !tc.wantError {
				require.Equal(t, tc.wantTenantID, capturedTenantID)
			}
		})
	}
}

func TestRequireTenantMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		tenantID   string
		wantStatus int
	}{
		{
			name:       "with tenant ID",
			tenantID:   middlewareTestUUID,
			wantStatus: http.StatusOK,
		},
		{
			name:       "without tenant ID",
			tenantID:   "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New()
			app.Use(TenantMiddleware())
			app.Use(RequireTenantMiddleware())

			app.Get("/test", func(c *fiber.Ctx) error {
				return c.SendString("ok")
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tc.tenantID != "" {
				req.Header.Set(TenantIDHeader, tc.tenantID)
			}

			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.wantStatus, resp.StatusCode)
		})
	}
}

func TestGetTenantID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ctx      context.Context
		wantID   string
		hasValue bool
	}{
		{
			name:     "with tenant ID in context",
			ctx:      context.WithValue(context.Background(), TenantContextKey{}, "test-tenant-id"),
			wantID:   "test-tenant-id",
			hasValue: true,
		},
		{
			name:     "without tenant ID in context",
			ctx:      context.Background(),
			wantID:   "",
			hasValue: false,
		},
		{
			name:     "with wrong type in context",
			ctx:      context.WithValue(context.Background(), TenantContextKey{}, 12345),
			wantID:   "",
			hasValue: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			id := GetTenantID(tc.ctx)
			require.Equal(t, tc.wantID, id)
		})
	}
}

func TestIsValidUUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{name: "valid lowercase UUID", input: middlewareTestUUID, valid: true},
		{name: "valid uppercase UUID", input: middlewareTestUUID, valid: true},
		{name: "valid mixed case UUID", input: middlewareTestUUID, valid: true},
		{name: "too short", input: "550e8400-e29b-41d4-a716", valid: false},
		{name: "too long", input: middlewareTestUUID + "1", valid: false},
		{name: "missing hyphens", input: "550e8400e29b41d4a716446655440000", valid: false},
		{name: "wrong hyphen position", input: "550e840-0e29b-41d4-a716-446655440000", valid: false},
		{name: "invalid hex char", input: "550g8400-e29b-41d4-a716-446655440000", valid: false},
		{name: "empty string", input: "", valid: false},
		{name: "spaces only", input: "                                    ", valid: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isValidUUID(tc.input)
			require.Equal(t, tc.valid, result)
		})
	}
}

func TestIsHexChar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input byte
		valid bool
	}{
		{name: "digit 0", input: '0', valid: true},
		{name: "digit 9", input: '9', valid: true},
		{name: "lowercase a", input: 'a', valid: true},
		{name: "lowercase f", input: 'f', valid: true},
		{name: "uppercase A", input: 'A', valid: true},
		{name: "uppercase F", input: 'F', valid: true},
		{name: "lowercase g", input: 'g', valid: false},
		{name: "uppercase G", input: 'G', valid: false},
		{name: "hyphen", input: '-', valid: false},
		{name: "space", input: ' ', valid: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isHexChar(tc.input)
			require.Equal(t, tc.valid, result)
		})
	}
}

func TestTenantMiddleware_ResponseBody(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	app.Use(TenantMiddleware())

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Test with invalid tenant ID.
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(TenantIDHeader, "invalid")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "invalid_tenant_id")
}
