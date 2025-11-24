# R08: OpenAPI Specification Synchronization - Analysis

**Analysis Date**: November 23, 2025
**Status**: 🔍 In Progress - Initial Review Complete

---

## OpenAPI Specs vs Actual Implementation

### Authz Service (OAuth 2.1 Authorization Server)

**OpenAPI Spec**: `api/identity/openapi_spec_authz.yaml`

| Endpoint | Method | OpenAPI Defined | Implementation Status | File | Notes |
|----------|--------|----------------|----------------------|------|-------|
| `/health` | GET | ✅ Yes | ✅ Implemented | handlers_health.go:13 | ✅ MATCH |
| `/oauth2/v1/authorize` | GET | ❌ No (only POST) | ✅ Implemented | handlers_authorize.go:22 | ⚠️ MISSING GET in spec |
| `/oauth2/v1/authorize` | POST | ✅ Yes | ✅ Implemented | handlers_authorize.go:160 | ✅ MATCH |
| `/oauth2/v1/token` | POST | ✅ Yes | ✅ Implemented | handlers_token.go:22 | ✅ MATCH (with grant handlers) |
| `/oauth2/v1/introspect` | POST | ✅ Yes | ✅ Implemented | handlers_introspect_revoke.go:19 | ✅ MATCH |
| `/oauth2/v1/revoke` | POST | ✅ Yes | ✅ Implemented | handlers_introspect_revoke.go:86 | ✅ MATCH |
| `/ui/swagger/doc.json` | GET | ❌ No | ✅ Implemented | routes.go:20 | ⚠️ MISSING in spec (OK - internal) |

**Grant Type Handlers** (all under `/token` endpoint):

- ✅ `authorization_code` - handlers_token.go:42 `handleAuthorizationCodeGrant`
- ✅ `client_credentials` - handlers_token.go:230 `handleClientCredentialsGrant`
- ✅ `refresh_token` - handlers_token.go:272 `handleRefreshTokenGrant`

**Route Registration**: `internal/identity/authz/routes.go`

---

### IdP Service (OIDC Identity Provider)

**OpenAPI Spec**: `api/identity/openapi_spec_idp.yaml`

| Endpoint | Method | OpenAPI Defined | Implementation Status | File | Notes |
|----------|--------|----------------|----------------------|------|-------|
| `/health` | GET | ✅ Yes | ✅ Implemented | handlers_health.go:13 | ✅ MATCH |
| `/oidc/v1/login` | GET | ✅ Yes | ✅ Implemented | handlers_login.go:20 | ✅ MATCH |
| `/oidc/v1/login` | POST | ✅ Yes | ✅ Implemented | handlers_login.go:41 | ✅ MATCH |
| `/oidc/v1/consent` | GET | ✅ Yes | ✅ Implemented | handlers_consent.go:64 | ✅ MATCH |
| `/oidc/v1/consent` | POST | ✅ Yes | ✅ Implemented | handlers_consent.go:174 | ✅ MATCH |
| `/oidc/v1/userinfo` | GET | ✅ Yes | ✅ Implemented | handlers_userinfo.go:17 | ✅ MATCH |
| `/oidc/v1/logout` | POST | ✅ Yes | ✅ Implemented | handlers_logout.go:15 | ✅ MATCH |
| `/ui/swagger/doc.json` | GET | ❌ No | ✅ Implemented | routes.go:20 | ⚠️ MISSING in spec (OK - internal) |

**Route Registration**: `internal/identity/idp/routes.go`

---

## Identified Gaps

### Critical Issues

1. **GET /oauth2/v1/authorize Missing from Spec**
   - **Impact**: OpenAPI spec only documents POST, but GET is implemented for initial authorize requests
   - **Used By**: OAuth 2.1 authorization code flow (step 1: client redirects user to GET /authorize)
   - **Action Required**: Add GET method to openapi_spec_authz.yaml
   - **Implementation Details**: Query parameters (response_type, client_id, redirect_uri, scope, state, code_challenge, code_challenge_method)
   - **Response**: 302 redirect to IdP login or consent form

### Minor Omissions (Acceptable)

1. **Swagger UI endpoint not in spec**
   - `/ui/swagger/doc.json` is an internal documentation endpoint
   - Not part of public API contract
   - No action required

### Schema Validation Needed

1. **Token response schemas**: Verify `AuthZTokenResponse` matches actual JWT claims structure
2. **Error response schemas**: Verify `OAuth2Error` matches identityDomainApperr.AppError format
3. **Introspection response**: Verify claim names match actual implementation
4. **UserInfo response**: Verify claim names match domain.User fields

---

## Next Steps

### Phase 1: Specification Updates (HIGH PRIORITY)

1. **Add GET /oauth2/v1/authorize to authz spec**
   - Add query parameter schema (same fields as POST body)
   - Document 302 redirect responses (to IdP login, to consent, with error)
   - Clarify relationship between GET (initial request) and POST (after authentication)

2. **Review request/response schemas for accuracy**
   - Compare AuthZTokenResponse to actual JWT structure in handlers_token.go
   - Compare OAuth2Error to identityDomainApperr.AppError
   - Compare IntrospectionResponse to token claims returned
   - Compare UserInfoResponse to domain.User fields

### Phase 2: Client Code Regeneration (MEDIUM PRIORITY)

1. **Regenerate authz client** after GET /authorize added
   - Config: `api/identity/openapi-gen_config_authz.yaml`
   - Command: `oapi-codegen -config api/identity/openapi-gen_config_authz.yaml api/identity/openapi_spec_authz.yaml`
   - Output: Update `api/client/authz_generated.go` (or similar)

2. **Regenerate idp client** (no changes expected, but verify)
   - Config: `api/identity/openapi-gen_config_idp.yaml`
   - Command: `oapi-codegen -config api/identity/openapi-gen_config_idp.yaml api/identity/openapi_spec_idp.yaml`
   - Output: Update `api/client/idp_generated.go` (or similar)

3. **Regenerate models** after schema reviews/updates
   - Config: `api/identity/openapi-gen_config_models.yaml`
   - Command: `oapi-codegen -config api/identity/openapi-gen_config_models.yaml api/identity/openapi_spec_components.yaml`
   - Output: Update `api/model/*_generated.go`

### Phase 3: Swagger UI Validation (LOW PRIORITY)

1. **Manual testing via Swagger UI** at `https://localhost:8080/ui/swagger`
   - Test all authz endpoints via UI
   - Verify request/response examples accurate
   - Check parameter validation matches implementation

2. **Manual testing via Swagger UI** at `https://localhost:8080/ui/swagger` (idp)
   - Test all idp endpoints via UI
   - Verify session cookie handling documented correctly
   - Check error responses match documented schemas

---

## Acceptance Criteria Checklist

- [ ] GET /oauth2/v1/authorize added to openapi_spec_authz.yaml
- [ ] All response schemas verified against actual implementation
- [ ] Client libraries regenerated and compilable
- [ ] No placeholder/TODO endpoints remain in specs
- [ ] Swagger UI reflects all actual endpoints
- [ ] All endpoints documented match routes.go registrations

---

**Analysis Completed**: November 23, 2025
**Ready for Implementation**: Phase 1 (Specification Updates)
