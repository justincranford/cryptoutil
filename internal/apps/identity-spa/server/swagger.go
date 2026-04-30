// Copyright (c) 2025-2026 Justin Cranford.
//

//nolint:wrapcheck // Fiber HTTP handlers return framework errors directly
package server

import (
	json "encoding/json"

	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	fiber "github.com/gofiber/fiber/v2"
)

// ServeOpenAPISpec serves a placeholder OpenAPI specification for the Identity Single Page App service.
// This is a BFF (Backend-for-Frontend) service; full spec generation is pending.
func ServeOpenAPISpec() (func(c *fiber.Ctx) error, error) {
	spec := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":                                 "Single Page App Backend-for-Frontend API",
			cryptoutilSharedMagic.CLIVersionCommand: cryptoutilSharedMagic.ServiceVersion,
		},
		"paths": map[string]any{},
	}

	specBytes, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}

	handler := func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")

		return c.Send(specBytes)
	}

	return handler, nil
}
