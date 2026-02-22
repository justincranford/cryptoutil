package middleware

import (
	"context"
	json "encoding/json"
	http "net/http"
	"net/http/httptest"
	"testing"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
)

func TestRequireScopeMiddleware(t *testing.T) {
	t.Parallel()

	validator := NewScopeValidator(DefaultScopeConfig())

	tests := []struct {
		name           string
		scope          string
		contextScopes  []string
		setScopes      bool
		expectedStatus int
	}{
		{
			name:           "has required scope",
			scope:          "kms:read",
			contextScopes:  []string{"kms:read", "kms:write"},
			setScopes:      true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing required scope",
			scope:          "kms:admin",
			contextScopes:  []string{"kms:read"},
			setScopes:      true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "no scopes in context",
			scope:          "kms:read",
			contextScopes:  nil,
			setScopes:      false,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})

			app.Use(func(c *fiber.Ctx) error {
				if tc.setScopes {
					ctx := context.WithValue(c.UserContext(), ScopeContextKey{}, tc.contextScopes)
					c.SetUserContext(ctx)
				}

				return c.Next()
			})
			app.Get("/test", RequireScope(validator, tc.scope), func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestRequireAnyScopeMiddleware(t *testing.T) {
	t.Parallel()

	validator := NewScopeValidator(DefaultScopeConfig())

	tests := []struct {
		name           string
		scopes         []string
		contextScopes  []string
		setScopes      bool
		expectedStatus int
	}{
		{
			name:           "has one of required scopes",
			scopes:         []string{"kms:admin", "kms:read"},
			contextScopes:  []string{"kms:read"},
			setScopes:      true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing all required scopes",
			scopes:         []string{"kms:admin", "kms:write"},
			contextScopes:  []string{"kms:read"},
			setScopes:      true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "no scopes in context",
			scopes:         []string{"kms:read"},
			contextScopes:  nil,
			setScopes:      false,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})

			app.Use(func(c *fiber.Ctx) error {
				if tc.setScopes {
					ctx := context.WithValue(c.UserContext(), ScopeContextKey{}, tc.contextScopes)
					c.SetUserContext(ctx)
				}

				return c.Next()
			})
			app.Get("/test", RequireAnyScope(validator, tc.scopes...), func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestRequireAllScopesMiddleware(t *testing.T) {
	t.Parallel()

	validator := NewScopeValidator(DefaultScopeConfig())

	tests := []struct {
		name           string
		scopes         []string
		contextScopes  []string
		setScopes      bool
		expectedStatus int
	}{
		{
			name:           "has all required scopes",
			scopes:         []string{"kms:read", "kms:write"},
			contextScopes:  []string{"kms:read", "kms:write", "kms:admin"},
			setScopes:      true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing one required scope",
			scopes:         []string{"kms:read", "kms:admin"},
			contextScopes:  []string{"kms:read"},
			setScopes:      true,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "no scopes in context",
			scopes:         []string{"kms:read"},
			contextScopes:  nil,
			setScopes:      false,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app := fiber.New(fiber.Config{DisableStartupMessage: true})

			app.Use(func(c *fiber.Ctx) error {
				if tc.setScopes {
					ctx := context.WithValue(c.UserContext(), ScopeContextKey{}, tc.contextScopes)
					c.SetUserContext(ctx)
				}

				return c.Next()
			})
			app.Get("/test", RequireAllScopes(validator, tc.scopes...), func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
		})
	}
}

func TestInsufficientScopeError_DetailLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		detailLevel    string
		provided       []string
		expectRequired bool
		expectProvided bool
	}{
		{
			name:           "minimal detail",
			detailLevel:    errorDetailLevelMin,
			provided:       []string{"kms:read"},
			expectRequired: false,
			expectProvided: false,
		},
		{
			name:           "standard detail",
			detailLevel:    errorDetailLevelStd,
			provided:       []string{"kms:read"},
			expectRequired: true,
			expectProvided: false,
		},
		{
			name:           "debug detail with provided",
			detailLevel:    errorDetailLevelDebug,
			provided:       []string{"kms:read"},
			expectRequired: true,
			expectProvided: true,
		},
		{
			name:           "debug detail without provided",
			detailLevel:    errorDetailLevelDebug,
			provided:       nil,
			expectRequired: true,
			expectProvided: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := DefaultScopeConfig()
			config.ErrorDetailLevel = tc.detailLevel
			validator := NewScopeValidator(config)

			app := fiber.New(fiber.Config{DisableStartupMessage: true})

			app.Use(func(c *fiber.Ctx) error {
				if tc.provided != nil {
					ctx := context.WithValue(c.UserContext(), ScopeContextKey{}, tc.provided)
					c.SetUserContext(ctx)
				}

				return c.Next()
			})
			app.Get("/test", RequireScope(validator, "kms:admin"), func(c *fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)

			defer func() { _ = resp.Body.Close() }()

			require.Equal(t, http.StatusForbidden, resp.StatusCode)

			var body map[string]any

			err = json.NewDecoder(resp.Body).Decode(&body)
			require.NoError(t, err)
			require.Equal(t, "insufficient_scope", body["error"])

			if tc.expectRequired {
				require.Contains(t, body, "required_scope")
				require.Contains(t, body, "error_description")
			} else {
				require.NotContains(t, body, "required_scope")
			}

			if tc.expectProvided {
				require.Contains(t, body, "provided_scopes")
			} else {
				require.NotContains(t, body, "provided_scopes")
			}
		})
	}
}
