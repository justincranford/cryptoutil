---
name: api-handler
description: "Map an OpenAPI operation to a strict server handler implementation for cryptoutil services. Use when adding a new API endpoint to generate the handler method, request/response mapping, and route registration boilerplate from the generated StrictServerInterface."
argument-hint: "[operation-name]"
---

Map an OpenAPI operation to a strict server handler implementation.

## Purpose

Use when adding a new API endpoint after defining it in the OpenAPI spec and
running oapi-codegen. Generates the handler method, domain-to-API mapping,
and route registration following the strict server pattern.

## Key Rules

- Handler DTOs MUST come from generated `api/*/server/` and `api/model/` packages
- NEVER hand-roll request/response structs that duplicate generated models
- ALWAYS use `strict-server: true` in oapi-codegen config
- Compile-time assertion: `var _ server.StrictServerInterface = (*StrictServer)(nil)`
- Handler methods receive `RequestObject`, return `ResponseObject`
- Business logic belongs in service layer, NOT in handlers
- Use a mapper struct or function for domain-to-API type conversion
- Pagination: `page` (default 1, min 1), `size` (default 50, min 1, max 1000)

## Handler Method Signature

```go
func (s *StrictServer) OperationName(
    ctx context.Context,
    request cryptoutilServer.OperationNameRequestObject,
) (cryptoutilServer.OperationNameResponseObject, error) {
    // 1. Extract parameters from request
    // 2. Call business logic service
    // 3. Map domain result to API response
    // 4. Return typed response object
}
```

## Template: StrictServer Struct

```go
package handler

import (
    "context"

    cryptoutilServer "cryptoutil/api/PS-ID/server"
    cryptoutilDomain "cryptoutil/internal/apps/PS-ID/domain"
    cryptoutilRepository "cryptoutil/internal/apps/PS-ID/repository"
    cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
)

// StrictServer implements the generated StrictServerInterface.
type StrictServer struct {
    repo    *cryptoutilRepository.ResourceRepository
    service *service.ResourceService
}

// NewStrictServer creates a new StrictServer.
func NewStrictServer(
    repo *cryptoutilRepository.ResourceRepository,
    svc *service.ResourceService,
) *StrictServer {
    return &StrictServer{repo: repo, service: svc}
}

// Compile-time assertion.
var _ cryptoutilServer.StrictServerInterface = (*StrictServer)(nil)
```

## Template: List Handler (with pagination)

```go
var (
    defaultPage = cryptoutilSharedMagic.DefaultPaginationPage
    defaultSize = cryptoutilSharedMagic.DefaultPaginationSize
)

// ListResources lists resources with pagination.
// (GET /resources).
func (s *StrictServer) ListResources(
    ctx context.Context,
    request cryptoutilServer.ListResourcesRequestObject,
) (cryptoutilServer.ListResourcesResponseObject, error) {
    page := defaultPage
    if request.Params.Page != nil {
        page = *request.Params.Page
    }

    size := defaultSize
    if request.Params.Size != nil {
        size = *request.Params.Size
    }

    items, total, err := s.repo.List(ctx, tenantID, page, size)
    if err != nil {
        return cryptoutilServer.ListResources500JSONResponse{
            Code:    "INTERNAL_ERROR",
            Message: "Failed to list resources",
        }, nil
    }

    apiItems := make([]cryptoutilServer.Resource, 0, len(items))
    for i := range items {
        apiItems = append(apiItems, domainToAPI(&items[i]))
    }

    return cryptoutilServer.ListResources200JSONResponse{
        Items:      apiItems,
        Pagination: cryptoutilServer.Pagination{Page: page, Size: size, Total: total},
    }, nil
}
```

## Template: Create Handler

```go
// CreateResource creates a new resource.
// (POST /resources).
func (s *StrictServer) CreateResource(
    ctx context.Context,
    request cryptoutilServer.CreateResourceRequestObject,
) (cryptoutilServer.CreateResourceResponseObject, error) {
    domainObj := apiToDomain(request.Body)

    created, err := s.service.Create(ctx, domainObj)
    if err != nil {
        return cryptoutilServer.CreateResource500JSONResponse{
            Code:    "INTERNAL_ERROR",
            Message: "Failed to create resource",
        }, nil
    }

    return cryptoutilServer.CreateResource201JSONResponse(domainToAPI(created)), nil
}
```

## Template: Get-by-ID Handler

```go
// GetResource gets a resource by ID.
// (GET /resources/{resourceId}).
func (s *StrictServer) GetResource(
    ctx context.Context,
    request cryptoutilServer.GetResourceRequestObject,
) (cryptoutilServer.GetResourceResponseObject, error) {
    result, err := s.service.GetByID(ctx, request.ResourceId)
    if err != nil {
        return cryptoutilServer.GetResource404JSONResponse{
            Code:    "NOT_FOUND",
            Message: "Resource not found",
        }, nil
    }

    return cryptoutilServer.GetResource200JSONResponse(domainToAPI(result)), nil
}
```

## Template: Domain-to-API Mapper

```go
func domainToAPI(d *cryptoutilDomain.Resource) cryptoutilServer.Resource {
    return cryptoutilServer.Resource{
        Id:        d.ID.String(),
        Name:      d.Name,
        CreatedAt: d.CreatedAt,
        UpdatedAt: d.UpdatedAt,
    }
}

func apiToDomain(a *cryptoutilServer.CreateResourceRequest) *cryptoutilDomain.Resource {
    return &cryptoutilDomain.Resource{
        Name: a.Name,
    }
}
```

## Template: Error Response Helpers

```go
func listResources500(message string) (cryptoutilServer.ListResourcesResponseObject, error) {
    return cryptoutilServer.ListResources500JSONResponse{
        Code:    "INTERNAL_ERROR",
        Message: message,
    }, nil
}
```

## Route Registration (in builder)

```go
builder.WithPublicRouteRegistration(func(
    base *cryptoutilFrameworkServer.PublicServerBase,
    res *cryptoutilFrameworkBuilder.ServiceResources,
) error {
    repo := cryptoutilRepository.NewResourceRepository(res.DB())
    svc := service.NewResourceService(repo)
    handler := handler.NewStrictServer(repo, svc)

    strictHandler := cryptoutilServer.NewStrictHandler(handler, nil)
    cryptoutilServer.RegisterHandlers(base.App(), strictHandler)

    return nil
})
```

## Validation Checklist

- [ ] Handler implements `StrictServerInterface` (compile-time assertion present)
- [ ] All DTOs from generated `api/*/server/` packages (no hand-rolled structs)
- [ ] Pagination defaults: page=1, size=50 (from magic constants)
- [ ] Error responses use generated typed response objects (e.g., `500JSONResponse`)
- [ ] Domain-to-API and API-to-domain mappers exist (not inline conversion)
- [ ] Business logic in service layer (handler only does mapping)
- [ ] Route registration via builder `WithPublicRouteRegistration`
- [ ] `go build ./...` passes after implementation

## References

Read [ARCHITECTURE.md Section 8.1 OpenAPI-First Design](../../../docs/ARCHITECTURE.md#81-openapi-first-design) for strict server pattern — follow type safety requirements and code generation configuration.

Read [ARCHITECTURE.md Section 5.2 Service Builder Pattern](../../../docs/ARCHITECTURE.md#52-service-builder-pattern) for route registration — use `WithPublicRouteRegistration` to register handlers with the service framework.

Read [ARCHITECTURE.md Section 8.2 REST Conventions](../../../docs/ARCHITECTURE.md#82-rest-conventions) for pagination and naming — apply plural nouns, kebab-case paths, and mandatory pagination on list endpoints.

Read [ARCHITECTURE.md Section 8.4 Error Handling](../../../docs/ARCHITECTURE.md#84-error-handling) for error response patterns — use the standard Error schema with code, message, details, and requestId.
