// Copyright (c) 2025 Justin Cranford
//

// Package businesslogic provides business logic services for cipher-im.
// Session management is delegated to the template package.
package businesslogic

import (
	"context"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"

	cryptoutilConfig "cryptoutil/internal/apps/template/service/config"
	cryptoutilTemplateBarrier "cryptoutil/internal/apps/template/service/server/barrier"
	cryptoutilTemplateBusinessLogic "cryptoutil/internal/apps/template/service/server/businesslogic"
	cryptoutilJose "cryptoutil/internal/shared/crypto/jose"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// SessionManagerService is an alias to the template's SessionManagerService.
type SessionManagerService = cryptoutilTemplateBusinessLogic.SessionManagerService

// NewSessionManagerService creates a new SessionManagerService instance.
// The defaultTenantID and defaultRealmID are used by the single-tenant convenience methods.
func NewSessionManagerService(
	ctx context.Context,
	db *gorm.DB,
	telemetryService *cryptoutilTelemetry.TelemetryService,
	jwkGenService *cryptoutilJose.JWKGenService,
	barrierService *cryptoutilTemplateBarrier.BarrierService,
	config *cryptoutilConfig.ServiceTemplateServerSettings,
	defaultTenantID googleUuid.UUID,
	defaultRealmID googleUuid.UUID,
) (*SessionManagerService, error) {
	return cryptoutilTemplateBusinessLogic.NewSessionManagerService(
		ctx,
		db,
		telemetryService,
		jwkGenService,
		barrierService,
		config,
		defaultTenantID,
		defaultRealmID,
	)
}
