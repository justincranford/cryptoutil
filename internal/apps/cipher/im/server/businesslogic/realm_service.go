// Copyright (c) 2025 Justin Cranford
//
// Package businesslogic provides business logic services for cipher-im.
// Realm management is delegated to the template package (follows session_manager_service pattern).
package businesslogic

import (
	"context"

	"gorm.io/gorm"

	cryptoutilTemplateService "cryptoutil/internal/apps/template/service/server/service"
	cryptoutilTelemetry "cryptoutil/internal/shared/telemetry"
)

// RealmService is an alias to the template's RealmService.
type RealmService = cryptoutilTemplateService.RealmService

// NewRealmService creates a new RealmService instance.
// Delegates to service-template for reusability across all 9 product-services.
func NewRealmService(
	ctx context.Context,
	db *gorm.DB,
	telemetryService *cryptoutilTelemetry.TelemetryService,
) (*RealmService, error) {
	return cryptoutilTemplateService.NewRealmService(ctx, db, telemetryService)
}
