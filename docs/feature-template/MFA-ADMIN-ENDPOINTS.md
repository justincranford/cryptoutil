# MFA Admin Endpoints Implementation Plan

## Overview

**Feature**: Administrative MFA Factor Management
**Status**: ❌ Not Implemented → ✅ 100% Complete
**Priority**: HIGH (MANDATORY)
**Estimated Time**: 4 hours

## Endpoints to Implement

1. **POST /oidc/v1/mfa/enroll** - Enroll MFA factor for user
2. **GET /oidc/v1/mfa/factors** - List user MFA factors
3. **DELETE /oidc/v1/mfa/factors/{id}** - Remove MFA factor

## Implementation Tasks

### Task 1: API Handlers (2 hours)

**File**: `internal/identity/authz/handlers_mfa_admin.go`

#### POST /oidc/v1/mfa/enroll

- Request: `{ "user_id": "uuid", "factor_type": "totp|email_otp|passkey|...", "name": "My Device", "required": false }`
- Response (201): `{ "id": "uuid", "factor_type": "totp", "name": "My Device", "created_at": "..." }`
- Validation: user exists, factor_type valid, name unique per user
- Business logic: Create MFAFactor, link to user's auth profile

#### GET /oidc/v1/mfa/factors

- Query params: `user_id=uuid` (required)
- Response (200): `{ "factors": [{ "id": "uuid", "factor_type": "totp", "name": "My Device", "enabled": true, "created_at": "..." }] }`
- Business logic: Fetch user's auth profile, list all MFA factors

#### DELETE /oidc/v1/mfa/factors/{id}

- Path param: `id=uuid`
- Query param: `user_id=uuid` (for verification)
- Response (204): No content
- Validation: factor exists, belongs to user
- Business logic: Soft delete MFA factor

### Task 2: Route Registration (15 minutes)

**File**: `internal/identity/authz/routes.go`

Add to `/oidc/v1` group:

```go
oidc.Post("/mfa/enroll", s.handleEnrollMFA)
oidc.Get("/mfa/factors", s.handleListMFAFactors)
oidc.Delete("/mfa/factors/:id", s.handleDeleteMFAFactor)
```

### Task 3: Handler Tests (1.5 hours)

**File**: `internal/identity/authz/handlers_mfa_admin_test.go`

Test cases:

- **Enroll**: happy path, invalid user, invalid factor type, duplicate name
- **List**: happy path, no factors, user not found
- **Delete**: happy path, factor not found, unauthorized (wrong user)

### Task 4: Integration Testing (30 minutes)

- Test full MFA enrollment workflow
- Verify soft delete (factor still in DB with deleted_at set)
- Test listing filters out deleted factors

## Dependencies

- ✅ MFAFactor domain model exists
- ✅ MFAFactorRepository interface exists
- ✅ GORM implementation exists (NewMFAFactorRepository)
- ✅ User repository exists
- ✅ AuthProfile repository exists

## Security Considerations

- **Authorization**: Admin endpoints require elevated privileges (future: check admin role)
- **Validation**: Enforce factor_type enum, unique names per user
- **Audit**: Log all MFA enrollment/deletion events
- **Rate Limiting**: Prevent abuse of enrollment endpoint

## Success Criteria

- [ ] All 3 endpoints implemented and working
- [ ] 10+ handler tests passing
- [ ] Routes registered in /oidc/v1 group
- [ ] Spec.md updated (❌ → ✅)
- [ ] Integration tests pass

## Out of Scope

- MFA enforcement in login flow (separate feature)
- Factor-specific configuration (TOTP secret, email verification)
- User-facing MFA enrollment (admin-only for now)
