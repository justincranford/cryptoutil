// Copyright (c) 2025 Justin Cranford
//

package server

import (
	"fmt"

	cryptoutilAppsJoseJaRepository "cryptoutil/internal/apps/jose/ja/repository"
	cryptoutilAppsJoseJaServerApis "cryptoutil/internal/apps/jose/ja/server/apis"
	cryptoutilAppsTemplateServiceServer "cryptoutil/internal/apps/template/service/server"
	cryptoutilAppsTemplateServiceServerBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilAppsTemplateServiceServerBusinesslogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilAppsTemplateServiceServerMiddleware "cryptoutil/internal/apps/template/service/server/middleware"
	cryptoutilAppsTemplateServiceServerService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilSharedCryptoJose "cryptoutil/internal/shared/crypto/jose"
)

// PublicServer implements the jose-ja public server by embedding PublicServerBase.
type PublicServer struct {
	base *cryptoutilAppsTemplateServiceServer.PublicServerBase // Reusable server infrastructure

	elasticJWKRepo        cryptoutilAppsJoseJaRepository.ElasticJWKRepository
	materialJWKRepo       cryptoutilAppsJoseJaRepository.MaterialJWKRepository
	auditConfigRepo       cryptoutilAppsJoseJaRepository.AuditConfigRepository
	auditLogRepo          cryptoutilAppsJoseJaRepository.AuditLogRepository
	jwkGenService         *cryptoutilSharedCryptoJose.JWKGenService
	barrierService        *cryptoutilAppsTemplateServiceServerBarrier.Service
	sessionManagerService *cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService
	realmService          cryptoutilAppsTemplateServiceServerService.RealmService

	// Handlers (composition pattern).
	jwkHandler *cryptoutilAppsJoseJaServerApis.JWKHandler
}

// NewPublicServer creates a new jose-ja public server using builder-provided infrastructure.
// Used by ServerBuilder during route registration.
func NewPublicServer(
	base *cryptoutilAppsTemplateServiceServer.PublicServerBase,
	sessionManagerService *cryptoutilAppsTemplateServiceServerBusinesslogic.SessionManagerService,
	realmService cryptoutilAppsTemplateServiceServerService.RealmService,
	elasticJWKRepo cryptoutilAppsJoseJaRepository.ElasticJWKRepository,
	materialJWKRepo cryptoutilAppsJoseJaRepository.MaterialJWKRepository,
	auditConfigRepo cryptoutilAppsJoseJaRepository.AuditConfigRepository,
	auditLogRepo cryptoutilAppsJoseJaRepository.AuditLogRepository,
	jwkGenService *cryptoutilSharedCryptoJose.JWKGenService,
	barrierService *cryptoutilAppsTemplateServiceServerBarrier.Service,
) (*PublicServer, error) {
	if base == nil {
		return nil, fmt.Errorf("public server base cannot be nil")
	} else if sessionManagerService == nil {
		return nil, fmt.Errorf("session manager service cannot be nil")
	} else if realmService == nil {
		return nil, fmt.Errorf("realm service cannot be nil")
	} else if elasticJWKRepo == nil {
		return nil, fmt.Errorf("elastic JWK repository cannot be nil")
	} else if materialJWKRepo == nil {
		return nil, fmt.Errorf("material JWK repository cannot be nil")
	} else if auditConfigRepo == nil {
		return nil, fmt.Errorf("audit config repository cannot be nil")
	} else if auditLogRepo == nil {
		return nil, fmt.Errorf("audit log repository cannot be nil")
	} else if jwkGenService == nil {
		return nil, fmt.Errorf("JWK generation service cannot be nil")
	} else if barrierService == nil {
		return nil, fmt.Errorf("barrier service cannot be nil")
	}

	s := &PublicServer{
		base:                  base,
		elasticJWKRepo:        elasticJWKRepo,
		materialJWKRepo:       materialJWKRepo,
		auditConfigRepo:       auditConfigRepo,
		auditLogRepo:          auditLogRepo,
		jwkGenService:         jwkGenService,
		barrierService:        barrierService,
		sessionManagerService: sessionManagerService,
		realmService:          realmService,
	}

	// Create JWK handler (business logic).
	s.jwkHandler = cryptoutilAppsJoseJaServerApis.NewJWKHandler(
		elasticJWKRepo,
		materialJWKRepo,
		auditConfigRepo,
		auditLogRepo,
		jwkGenService,
		barrierService,
	)

	return s, nil
}

// registerRoutes sets up the API endpoints.
// Called by ServerBuilder after NewPublicServer returns.
func (s *PublicServer) registerRoutes() error {
	// Create session handler.
	sessionHandler := cryptoutilAppsJoseJaServerApis.NewSessionHandler(s.sessionManagerService)

	// Create session middleware for browser and service paths using template middleware directly.
	browserSessionMiddleware := cryptoutilAppsTemplateServiceServerMiddleware.BrowserSessionMiddleware(s.sessionManagerService)
	serviceSessionMiddleware := cryptoutilAppsTemplateServiceServerMiddleware.ServiceSessionMiddleware(s.sessionManagerService)

	// Get underlying Fiber app from base for route registration.
	app := s.base.App()

	// Session management endpoints (no middleware - these endpoints create/validate sessions).
	app.Post("/service/api/v1/sessions/issue", sessionHandler.IssueSession)
	app.Post("/service/api/v1/sessions/validate", sessionHandler.ValidateSession)
	app.Post("/browser/api/v1/sessions/issue", sessionHandler.IssueSession)
	app.Post("/browser/api/v1/sessions/validate", sessionHandler.ValidateSession)

	// Elastic JWK management endpoints (CRUD operations on elastic keys).
	app.Post("/service/api/v1/elastic-jwks", serviceSessionMiddleware, s.jwkHandler.HandleCreateElasticJWK())
	app.Get("/service/api/v1/elastic-jwks", serviceSessionMiddleware, s.jwkHandler.HandleListElasticJWKs())
	app.Get("/service/api/v1/elastic-jwks/:kid", serviceSessionMiddleware, s.jwkHandler.HandleGetElasticJWK())
	app.Delete("/service/api/v1/elastic-jwks/:kid", serviceSessionMiddleware, s.jwkHandler.HandleDeleteElasticJWK())

	app.Post("/browser/api/v1/elastic-jwks", browserSessionMiddleware, s.jwkHandler.HandleCreateElasticJWK())
	app.Get("/browser/api/v1/elastic-jwks", browserSessionMiddleware, s.jwkHandler.HandleListElasticJWKs())
	app.Get("/browser/api/v1/elastic-jwks/:kid", browserSessionMiddleware, s.jwkHandler.HandleGetElasticJWK())
	app.Delete("/browser/api/v1/elastic-jwks/:kid", browserSessionMiddleware, s.jwkHandler.HandleDeleteElasticJWK())

	// Material JWK management endpoints (key material rotation and retrieval).
	app.Post("/service/api/v1/elastic-jwks/:kid/materials", serviceSessionMiddleware, s.jwkHandler.HandleCreateMaterialJWK())
	app.Get("/service/api/v1/elastic-jwks/:kid/materials", serviceSessionMiddleware, s.jwkHandler.HandleListMaterialJWKs())
	app.Get("/service/api/v1/elastic-jwks/:kid/materials/active", serviceSessionMiddleware, s.jwkHandler.HandleGetActiveMaterialJWK())
	app.Post("/service/api/v1/elastic-jwks/:kid/rotate", serviceSessionMiddleware, s.jwkHandler.HandleRotateMaterialJWK())

	app.Post("/browser/api/v1/elastic-jwks/:kid/materials", browserSessionMiddleware, s.jwkHandler.HandleCreateMaterialJWK())
	app.Get("/browser/api/v1/elastic-jwks/:kid/materials", browserSessionMiddleware, s.jwkHandler.HandleListMaterialJWKs())
	app.Get("/browser/api/v1/elastic-jwks/:kid/materials/active", browserSessionMiddleware, s.jwkHandler.HandleGetActiveMaterialJWK())
	app.Post("/browser/api/v1/elastic-jwks/:kid/rotate", browserSessionMiddleware, s.jwkHandler.HandleRotateMaterialJWK())

	// JWKS endpoint (public key set for verification - typically public, no auth required).
	app.Get("/service/api/v1/jwks.json", s.jwkHandler.HandleGetJWKS())
	app.Get("/browser/api/v1/jwks.json", s.jwkHandler.HandleGetJWKS())
	app.Get("/.well-known/jwks.json", s.jwkHandler.HandleGetJWKS()) // Standard well-known endpoint.

	// Cryptographic operations (sign, verify, encrypt, decrypt).
	app.Post("/service/api/v1/sign", serviceSessionMiddleware, s.jwkHandler.HandleSign())
	app.Post("/service/api/v1/verify", serviceSessionMiddleware, s.jwkHandler.HandleVerify())
	app.Post("/service/api/v1/encrypt", serviceSessionMiddleware, s.jwkHandler.HandleEncrypt())
	app.Post("/service/api/v1/decrypt", serviceSessionMiddleware, s.jwkHandler.HandleDecrypt())

	app.Post("/browser/api/v1/sign", browserSessionMiddleware, s.jwkHandler.HandleSign())
	app.Post("/browser/api/v1/verify", browserSessionMiddleware, s.jwkHandler.HandleVerify())
	app.Post("/browser/api/v1/encrypt", browserSessionMiddleware, s.jwkHandler.HandleEncrypt())
	app.Post("/browser/api/v1/decrypt", browserSessionMiddleware, s.jwkHandler.HandleDecrypt())

	return nil
}

// PublicBaseURL returns the base URL for public API access by delegating to PublicServerBase.
func (s *PublicServer) PublicBaseURL() string {
	return s.base.PublicBaseURL()
}
