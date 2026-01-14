// Copyright (c) 2025 Justin Cranford
//

package middleware

import (
	"github.com/gofiber/fiber/v2"

	"cryptoutil/internal/apps/cipher/im/server/businesslogic"
	cryptoutilTemplateMiddleware "cryptoutil/internal/apps/template/service/server/middleware"
)

// BrowserSessionMiddleware creates middleware for browser session validation.
// Delegates to the template middleware.
func BrowserSessionMiddleware(sessionManager *businesslogic.SessionManagerService) fiber.Handler {
	return cryptoutilTemplateMiddleware.BrowserSessionMiddleware(sessionManager)
}

// ServiceSessionMiddleware creates middleware for service session validation.
// Delegates to the template middleware.
func ServiceSessionMiddleware(sessionManager *businesslogic.SessionManagerService) fiber.Handler {
	return cryptoutilTemplateMiddleware.ServiceSessionMiddleware(sessionManager)
}
