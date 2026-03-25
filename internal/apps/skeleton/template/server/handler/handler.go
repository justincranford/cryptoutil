// Copyright (c) 2025 Justin Cranford
//
// TEMPLATE: Copy and rename 'skeleton' -> your-service-name before use.

// Package handler implements the OpenAPI strict server for the skeleton-template service.
package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	cryptoutilSkeletonTemplateServer "cryptoutil/api/skeleton-template/server"
	cryptoutilAppsSkeletonTemplateDomain "cryptoutil/internal/apps/skeleton/template/domain"
	cryptoutilAppsSkeletonTemplateRepository "cryptoutil/internal/apps/skeleton/template/repository"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"

	googleUuid "github.com/google/uuid"
	"gorm.io/gorm"
)

// defaultTenantID is a fixed tenant ID for the skeleton-template.
// Real services derive tenant from authentication context.
var defaultTenantID = googleUuid.MustParse("00000000-0000-0000-0000-000000000001")

// defaultPage is the default page number for list operations.
var defaultPage = 1

// defaultSize is the default page size for list operations.
var defaultSize = cryptoutilSharedMagic.SeparatorLength

// StrictServer implements the generated StrictServerInterface for Item CRUD.
type StrictServer struct {
	itemRepo *cryptoutilAppsSkeletonTemplateRepository.ItemRepository
}

// NewStrictServer creates a new StrictServer with the given repository.
func NewStrictServer(itemRepo *cryptoutilAppsSkeletonTemplateRepository.ItemRepository) *StrictServer {
	return &StrictServer{itemRepo: itemRepo}
}

// Compile-time assertion: StrictServer implements StrictServerInterface.
var _ cryptoutilSkeletonTemplateServer.StrictServerInterface = (*StrictServer)(nil)

// ListItems lists items for the default tenant with pagination.
// (GET /items).
func (s *StrictServer) ListItems(_ context.Context, request cryptoutilSkeletonTemplateServer.ListItemsRequestObject) (cryptoutilSkeletonTemplateServer.ListItemsResponseObject, error) {
	page := defaultPage
	if request.Params.Page != nil {
		page = *request.Params.Page
	}

	size := defaultSize
	if request.Params.Size != nil {
		size = *request.Params.Size
	}

	items, total, err := s.itemRepo.List(context.Background(), defaultTenantID, page, size)
	if err != nil {
		return listItems500("Failed to list items")
	}

	apiItems := make([]cryptoutilSkeletonTemplateServer.Item, 0, len(items))

	for i := range items {
		apiItems = append(apiItems, domainToAPI(&items[i]))
	}

	return cryptoutilSkeletonTemplateServer.ListItems200JSONResponse{
		Items: apiItems,
		Pagination: cryptoutilSkeletonTemplateServer.Pagination{
			Page:  page,
			Size:  size,
			Total: total,
		},
	}, nil
}

// CreateItem creates a new item for the default tenant.
// (POST /items).
func (s *StrictServer) CreateItem(_ context.Context, request cryptoutilSkeletonTemplateServer.CreateItemRequestObject) (cryptoutilSkeletonTemplateServer.CreateItemResponseObject, error) {
	if request.Body == nil {
		return cryptoutilSkeletonTemplateServer.CreateItem400JSONResponse{BadRequestJSONResponse: cryptoutilSkeletonTemplateServer.BadRequestJSONResponse{
			Code:    "INVALID_REQUEST",
			Message: "Request body is required",
		}}, nil
	}

	now := time.Now().UTC()
	item := &cryptoutilAppsSkeletonTemplateDomain.TemplateItem{
		ID:          googleUuid.Must(googleUuid.NewV7()),
		TenantID:    defaultTenantID,
		Name:        request.Body.Name,
		Description: derefString(request.Body.Description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.itemRepo.Create(context.Background(), item); err != nil {
		return createItem500("Failed to create item")
	}

	return cryptoutilSkeletonTemplateServer.CreateItem201JSONResponse(domainToAPI(item)), nil
}

// GetItem retrieves an item by ID.
// (GET /items/{itemID}).
func (s *StrictServer) GetItem(_ context.Context, request cryptoutilSkeletonTemplateServer.GetItemRequestObject) (cryptoutilSkeletonTemplateServer.GetItemResponseObject, error) {
	item, err := s.itemRepo.GetByID(context.Background(), defaultTenantID, googleUuid.UUID(request.ItemID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return cryptoutilSkeletonTemplateServer.GetItem404JSONResponse{NotFoundJSONResponse: cryptoutilSkeletonTemplateServer.NotFoundJSONResponse{
				Code:    "NOT_FOUND",
				Message: fmt.Sprintf("Item %s not found", request.ItemID),
			}}, nil
		}

		return getItem500("Failed to get item")
	}

	return cryptoutilSkeletonTemplateServer.GetItem200JSONResponse(domainToAPI(item)), nil
}

// UpdateItem updates an existing item.
// (PUT /items/{itemID}).
func (s *StrictServer) UpdateItem(_ context.Context, request cryptoutilSkeletonTemplateServer.UpdateItemRequestObject) (cryptoutilSkeletonTemplateServer.UpdateItemResponseObject, error) {
	if request.Body == nil {
		return cryptoutilSkeletonTemplateServer.UpdateItem400JSONResponse{BadRequestJSONResponse: cryptoutilSkeletonTemplateServer.BadRequestJSONResponse{
			Code:    "INVALID_REQUEST",
			Message: "Request body is required",
		}}, nil
	}

	existing, err := s.itemRepo.GetByID(context.Background(), defaultTenantID, googleUuid.UUID(request.ItemID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return cryptoutilSkeletonTemplateServer.UpdateItem404JSONResponse{NotFoundJSONResponse: cryptoutilSkeletonTemplateServer.NotFoundJSONResponse{
				Code:    "NOT_FOUND",
				Message: fmt.Sprintf("Item %s not found", request.ItemID),
			}}, nil
		}

		return updateItem500("Failed to get item for update")
	}

	existing.Name = request.Body.Name
	existing.Description = derefString(request.Body.Description)

	if err := s.itemRepo.Update(context.Background(), existing); err != nil {
		return updateItem500("Failed to update item")
	}

	return cryptoutilSkeletonTemplateServer.UpdateItem200JSONResponse(domainToAPI(existing)), nil
}

// DeleteItem deletes an item by ID.
// (DELETE /items/{itemID}).
func (s *StrictServer) DeleteItem(_ context.Context, request cryptoutilSkeletonTemplateServer.DeleteItemRequestObject) (cryptoutilSkeletonTemplateServer.DeleteItemResponseObject, error) {
	if err := s.itemRepo.Delete(context.Background(), defaultTenantID, googleUuid.UUID(request.ItemID)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return cryptoutilSkeletonTemplateServer.DeleteItem404JSONResponse{NotFoundJSONResponse: cryptoutilSkeletonTemplateServer.NotFoundJSONResponse{
				Code:    "NOT_FOUND",
				Message: fmt.Sprintf("Item %s not found", request.ItemID),
			}}, nil
		}

		return deleteItem500("Failed to delete item")
	}

	return cryptoutilSkeletonTemplateServer.DeleteItem204Response{}, nil
}

// domainToAPI converts a domain TemplateItem to an API Item.
func domainToAPI(item *cryptoutilAppsSkeletonTemplateDomain.TemplateItem) cryptoutilSkeletonTemplateServer.Item {
	desc := item.Description

	return cryptoutilSkeletonTemplateServer.Item{
		ID:          item.ID,
		TenantID:    item.TenantID,
		Name:        item.Name,
		Description: &desc,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}

// derefString returns the value of a string pointer, or empty string if nil.
func derefString(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

// listItems500 builds a ListItems 500 response.
func listItems500(message string) (cryptoutilSkeletonTemplateServer.ListItemsResponseObject, error) {
	return cryptoutilSkeletonTemplateServer.ListItems500JSONResponse{InternalServerErrorJSONResponse: cryptoutilSkeletonTemplateServer.InternalServerErrorJSONResponse{
		Code:    "INTERNAL_ERROR",
		Message: message,
	}}, nil
}

// createItem500 builds a CreateItem 500 response.
func createItem500(message string) (cryptoutilSkeletonTemplateServer.CreateItemResponseObject, error) {
	return cryptoutilSkeletonTemplateServer.CreateItem500JSONResponse{InternalServerErrorJSONResponse: cryptoutilSkeletonTemplateServer.InternalServerErrorJSONResponse{
		Code:    "INTERNAL_ERROR",
		Message: message,
	}}, nil
}

// getItem500 builds a GetItem 500 response.
func getItem500(message string) (cryptoutilSkeletonTemplateServer.GetItemResponseObject, error) {
	return cryptoutilSkeletonTemplateServer.GetItem500JSONResponse{InternalServerErrorJSONResponse: cryptoutilSkeletonTemplateServer.InternalServerErrorJSONResponse{
		Code:    "INTERNAL_ERROR",
		Message: message,
	}}, nil
}

// updateItem500 builds an UpdateItem 500 response.
func updateItem500(message string) (cryptoutilSkeletonTemplateServer.UpdateItemResponseObject, error) {
	return cryptoutilSkeletonTemplateServer.UpdateItem500JSONResponse{InternalServerErrorJSONResponse: cryptoutilSkeletonTemplateServer.InternalServerErrorJSONResponse{
		Code:    "INTERNAL_ERROR",
		Message: message,
	}}, nil
}

// deleteItem500 builds a DeleteItem 500 response.
func deleteItem500(message string) (cryptoutilSkeletonTemplateServer.DeleteItemResponseObject, error) {
	return cryptoutilSkeletonTemplateServer.DeleteItem500JSONResponse{InternalServerErrorJSONResponse: cryptoutilSkeletonTemplateServer.InternalServerErrorJSONResponse{
		Code:    "INTERNAL_ERROR",
		Message: message,
	}}, nil
}
