# Template Route Registration Guide

## Overview

The service template provides reusable infrastructure for tenant registration. Services using the template must register routes in the `WithPublicRouteRegistration` callback.

## Registration Handler Integration

### Step 1: Create Registration Service in Builder Callback

```go
builder.WithPublicRouteRegistration(func(
    base *cryptoutilTemplateServer.PublicServerBase,
    res *cryptoutilTemplateBuilder.ServiceResources,
) error {
    // Create tenant registration service (template business logic).
    tenantRegistrationService := cryptoutilTemplateBusinessLogic.NewTenantRegistrationService(
        res.DB,
        res.RealmRepository, // TenantRealmRepository provides tenant CRUD.
        // TODO: Add userRepo when tenant-user association implemented.
        // TODO: Add joinRequestRepo when TenantJoinRequestRepository implemented.
    )

    // Create registration handlers (template API layer).
    registrationHandlers := cryptoutilTemplateApis.NewRegistrationHandlers(tenantRegistrationService)

    // Register routes on Fiber app.
    app := base.App()

    // Public registration endpoint (no authentication).
    app.Post("/browser/api/v1/auth/register", registrationHandlers.HandleRegisterUser)
    app.Post("/service/api/v1/auth/register", registrationHandlers.HandleRegisterUser)

    // Admin join request management (requires authentication + admin role).
    // TODO: Add authentication middleware when implemented.
    app.Get("/browser/api/v1/admin/join-requests", registrationHandlers.HandleListJoinRequests)
    app.Post("/browser/api/v1/admin/join-requests/:id/approve", registrationHandlers.HandleApproveJoinRequest)
    app.Post("/browser/api/v1/admin/join-requests/:id/reject", registrationHandlers.HandleRejectJoinRequest)

    app.Get("/service/api/v1/admin/join-requests", registrationHandlers.HandleListJoinRequests)
    app.Post("/service/api/v1/admin/join-requests/:id/approve", registrationHandlers.HandleApproveJoinRequest)
    app.Post("/service/api/v1/admin/join-requests/:id/reject", registrationHandlers.HandleRejectJoinRequest)

    return nil
})
```

### Step 2: Testing Route Integration

Integration tests should verify end-to-end flow:

```go
func TestRegistrationFlow_E2E(t *testing.T) {
    // Setup: Start template service with route registration.
    server := setupTestServer(t)
    defer server.Shutdown()

    // Test 1: Register user with new tenant.
    reqBody := `{"username":"alice","email":"alice@example.com","password":"password123","tenant_name":"Acme Corp","create_tenant":true}`
    resp := makeRequest(t, "POST", server.PublicBaseURL()+"/browser/api/v1/auth/register", reqBody)
    require.Equal(t, 201, resp.StatusCode)

    // Test 2: Admin lists join requests.
    // TODO: Add authentication token to request.
    resp = makeRequest(t, "GET", server.PublicBaseURL()+"/browser/api/v1/admin/join-requests", "")
    require.Equal(t, 200, resp.StatusCode)

    // Test 3: Admin approves join request.
    // TODO: Add authentication token to request.
    resp = makeRequest(t, "POST", server.PublicBaseURL()+"/browser/api/v1/admin/join-requests/"+requestID+"/approve", `{"approved":true}`)
    require.Equal(t, 200, resp.StatusCode)
}
```

## Current Limitations

### Authentication Not Yet Implemented

- Registration endpoint (`/auth/register`) is public (no auth required).
- Admin endpoints (`/admin/join-requests/*`) currently have NO authentication middleware.
- **Security Risk**: Anyone can approve/reject join requests without authentication.
- **TODO**: Add authentication middleware in future tasks when auth framework implemented.

### User and Join Request Repositories Not Yet Integrated

- TenantRegistrationService currently uses only RealmRepository (for tenant CRUD).
- **TODO**: Add UserRepository when tenant-user association implemented.
- **TODO**: Add TenantJoinRequestRepository when join workflow fully implemented.

### Placeholder Business Logic

- `RegisterUserWithTenant` with `create_tenant=false` returns "not yet implemented" error.
- **TODO**: Implement join existing tenant flow (creates join request, admin approves, user assigned to tenant).

- `AuthorizeJoinRequest` does NOT verify admin permission.
- **TODO**: Add admin role check when authentication/authorization implemented.

- `AuthorizeJoinRequest` does NOT assign user/client to tenant on approval.
- **TODO**: Implement tenant assignment when user-tenant association implemented.

## Example: sm-im Service

See `internal/apps/sm/im/server/server.go` for complete route registration example:

```go
builder.WithPublicRouteRegistration(func(
    base *cryptoutilTemplateServer.PublicServerBase,
    res *cryptoutilTemplateBuilder.ServiceResources,
) error {
    // Create sm-im specific repositories.
    userRepo := repository.NewUserRepository(res.DB)
    messageRepo := repository.NewMessageRepository(res.DB)
    messageRecipientJWKRepo := repository.NewMessageRecipientJWKRepository(res.DB, res.BarrierService)

    // Create public server with handlers.
    publicServer, err := NewPublicServer(
        base,
        res.SessionManager,
        res.RealmService,
        userRepo,
        messageRepo,
        messageRecipientJWKRepo,
        res.JWKGenService,
        res.BarrierService,
    )
    if err != nil {
        return fmt.Errorf("failed to create public server: %w", err)
    }

    // Register all routes.
    if err := publicServer.registerRoutes(); err != nil {
        return fmt.Errorf("failed to register public routes: %w", err)
    }

    return nil
})
```

## Next Steps

1. **Task 0.10**: Implement TestMain pattern for database-backed integration tests.
2. **Task 0.11**: Create E2E tests with full HTTP request/response cycle.
3. **Task 0.12**: Phase 0 validation (coverage, mutation, quality gates).
4. **Future**: Add authentication middleware to protect admin endpoints.
5. **Future**: Implement join existing tenant flow (create join request → admin approves → user assigned).
6. **Future**: Add UserRepository and TenantJoinRequestRepository to TenantRegistrationService.
