// Copyright (c) 2025-2026 Justin Cranford.
package server

import (
	"fmt"

	cryptoutilApiCaServer "cryptoutil/api/sm-kms/pki/server"
	cryptoutilCAHandler "cryptoutil/internal/apps/sm-kms/server/pki/api/handler"

	fiber "github.com/gofiber/fiber/v2"
)

// RegisterPKIRoutes mounts the PKI API on the KMS service and browser route groups.
func RegisterPKIRoutes(serviceGroup, browserGroup fiber.Router, components *PKIComponents) error {
	switch {
	case serviceGroup == nil:
		return fmt.Errorf("PKI service route group is required")
	case browserGroup == nil:
		return fmt.Errorf("PKI browser route group is required")
	case components == nil || components.handler == nil:
		return fmt.Errorf("PKI components and handler are required")
	default:
		options := cryptoutilApiCaServer.FiberServerOptions{
			BaseURL:     "/pki",
			Middlewares: []cryptoutilApiCaServer.MiddlewareFunc{cryptoutilCAHandler.StrictRequestBodyErrorMiddleware},
		}
		strictHandler := cryptoutilApiCaServer.NewStrictHandler(
			cryptoutilCAHandler.NewStrictServer(components.handler),
			nil,
		)
		cryptoutilApiCaServer.RegisterHandlersWithOptions(serviceGroup, strictHandler, options)
		cryptoutilApiCaServer.RegisterHandlersWithOptions(browserGroup, strictHandler, options)
	}

	return nil
}
