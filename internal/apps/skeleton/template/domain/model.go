// Copyright (c) 2025 Justin Cranford
//
// TEMPLATE: Copy and rename 'skeleton' → your-service-name before use.

// Package domain provides domain models for the skeleton-template service.
// These models represent a minimal template item for demonstration purposes.
package domain

import (
	"time"

	googleUuid "github.com/google/uuid"
)

// TemplateItem represents a minimal domain entity for the skeleton-template service.
// This model demonstrates best-practice GORM tagging with cross-DB compatibility.
// CRITICAL: TenantID for data scoping only - realms are authentication-only, NOT data scope.
type TemplateItem struct {
	ID        googleUuid.UUID `gorm:"type:text;primaryKey"`
	TenantID  googleUuid.UUID `gorm:"type:text;not null;index:idx_template_items_tenant"`
	CreatedAt time.Time       `gorm:"not null;autoCreateTime"`
}

// TableName specifies the database table name for TemplateItem.
func (TemplateItem) TableName() string {
	return "template_items"
}

// =============================================================================
// EXAMPLE DOMAIN PATTERN — commented reference for new service implementation.
// When creating a new service from this template, replace the examples below
// with real domain types following the patterns shown.
// =============================================================================

// --- Example: Rich Entity Model (GORM) ---
// type MyServiceItem struct {
// 	ID          googleUuid.UUID `gorm:"type:text;primaryKey"`
// 	TenantID    googleUuid.UUID `gorm:"type:text;not null;index:idx_my_service_items_tenant"`
// 	Name        string          `gorm:"type:text;not null"`
// 	Description string          `gorm:"type:text"`
// 	Tags        []string        `gorm:"serializer:json"`    // Cross-DB JSON array
// 	Status      string          `gorm:"type:text;not null;default:active"`
// 	CreatedAt   time.Time       `gorm:"not null;autoCreateTime"`
// 	UpdatedAt   time.Time       `gorm:"not null;autoUpdateTime"`
// }
//
// func (MyServiceItem) TableName() string { return "my_service_items" }

// --- Example: Repository Interface ---
// type Repository interface {
// 	Create(ctx context.Context, item *MyServiceItem) error
// 	GetByID(ctx context.Context, tenantID, id googleUuid.UUID) (*MyServiceItem, error)
// 	List(ctx context.Context, tenantID googleUuid.UUID, page, size int) ([]MyServiceItem, int64, error)
// 	Update(ctx context.Context, item *MyServiceItem) error
// 	Delete(ctx context.Context, tenantID, id googleUuid.UUID) error
// }

// --- Example: Service Layer Stub ---
// type Service interface {
// 	Create(ctx context.Context, tenantID googleUuid.UUID, name, description string) (*MyServiceItem, error)
// 	Get(ctx context.Context, tenantID, id googleUuid.UUID) (*MyServiceItem, error)
// 	List(ctx context.Context, tenantID googleUuid.UUID, page, size int) ([]MyServiceItem, int64, error)
// 	Delete(ctx context.Context, tenantID, id googleUuid.UUID) error
// }

// --- Example: OpenAPI Handler Stub (implements StrictServerInterface) ---
// type Handler struct{ service Service }
//
// func NewHandler(service Service) *Handler { return &Handler{service: service} }
//
// func (h *Handler) CreateItem(ctx context.Context, req CreateItemRequest) (CreateItemResponse, error) {
// 	item, err := h.service.Create(ctx, req.TenantID, req.Body.Name, req.Body.Description)
// 	if err != nil { return nil, err }
// 	return CreateItem201JSONResponse{ID: item.ID, Name: item.Name}, nil
// }

