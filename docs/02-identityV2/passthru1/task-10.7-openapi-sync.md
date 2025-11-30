# Task 10.7: OpenAPI Specification Synchronization and Client Generation

## Task Reflection

### What Went Well

- ✅ **Task 10.5 Endpoints**: Core OAuth/OIDC endpoints implemented (/authorize, /token, /login, /health)
- ✅ **Task 10.6 CLI**: Unified CLI enables easy service startup for testing OpenAPI integration
- ✅ **Existing OpenAPI Infrastructure**: `api/` directory has generation configs and partial specs

### At Risk Items

- ❌ **Spec Drift**: OpenAPI specs don't reflect implemented endpoints (still reference placeholder paths)
- ❌ **Missing Endpoints in Spec**: New endpoints from Task 10.5 not documented in OpenAPI files
- ❌ **Client Generation Stale**: Generated client code may not match actual server behavior
- ❌ **Swagger UI Incomplete**: UI doesn't show full API surface area for manual testing

### Could Be Improved

- **API Documentation**: Current specs lack detailed descriptions, examples, security schemes
- **Response Schemas**: Many endpoints have generic responses instead of typed schemas
- **Error Responses**: OAuth 2.1 error format not fully documented
- **Code Generation**: No automated validation that generated code matches specs

### Dependencies and Blockers

- **Dependency on Task 10.5**: Requires implemented endpoints to document
- **Dependency on Task 10.6**: Unified CLI simplifies running services for spec validation
- **Enables Tasks 11-15**: Feature additions require up-to-date OpenAPI specs for client generation
- **Enables External Integrations**: Third parties need accurate specs for client implementation

---

## Objective

Synchronize OpenAPI 3.0 specifications with the implemented identity service endpoints, regenerate client libraries with oapi-codegen, update Swagger UI configuration, and establish automated validation that specs match running services.

**Acceptance Criteria**:

- OpenAPI specs include all endpoints from Task 10.5 (/authorize, /token, /login, /health)
- Specs validate against OpenAPI 3.0.3 schema
- oapi-codegen regenerates client libraries without errors
- Swagger UI displays all endpoints with accurate request/response schemas
- Automated test validates spec matches server behavior (contract testing)

---

## Historical Context

- **Original OpenAPI Setup**: `api/openapi_spec_components.yaml` and `api/openapi_spec_paths.yaml` created early in project
- **Identity API Gap**: Identity services added later without updating OpenAPI specs
- **Task 16 Deferral**: Original plan had "OpenAPI 3.0 Spec Modernization" as Task 16 (late in sequence)
- **Moved Earlier**: Refactored plan moves OpenAPI work to Task 10.7 to document working APIs before feature additions

---

## Scope

### In-Scope

1. **OpenAPI Specification Files**:
   - `api/identity/openapi_spec_authz.yaml`: OAuth 2.1 Authorization Server endpoints
   - `api/identity/openapi_spec_idp.yaml`: OIDC Identity Provider endpoints
   - `api/identity/openapi_spec_rs.yaml`: Resource Server endpoints (validate existing)
   - `api/identity/openapi_spec_components.yaml`: Shared components (schemas, security schemes, responses)

2. **Endpoint Documentation** (AuthZ):
   - `POST /oauth2/v1/authorize`: Authorization request with PKCE
   - `POST /oauth2/v1/token`: Token endpoint (authorization_code, refresh_token, client_credentials grants)
   - `GET /health`: Health check endpoint

3. **Endpoint Documentation** (IdP):
   - `GET/POST /oidc/v1/login`: User authentication endpoint
   - `GET /health`: Health check endpoint

4. **Endpoint Documentation** (RS - validate existing):
   - OAuth 2.0 Bearer token protected endpoints
   - Scope enforcement patterns

5. **Client Code Generation**:
   - Regenerate with oapi-codegen using configs in `api/openapi-gen_config_*.yaml`
   - Validate generated code compiles
   - Update imports in server code if needed

6. **Swagger UI Integration**:
   - Update Swagger UI to serve identity API specs
   - Configure proper base URLs and security schemes
   - Test interactive API exploration

7. **Contract Testing**:
   - Automated test that validates OpenAPI spec against running server
   - Use schemathesis or similar tool to generate requests from spec
   - Assert server responses match spec schemas

### Out-of-Scope

- **OIDC Discovery**: `/.well-known/openid-configuration` endpoint (defer to future task)
- **OAuth 2.1 Metadata**: `/.well-known/oauth-authorization-server` (defer to future task)
- **Advanced Security**: OAuth 2.0 Token Introspection/Revocation specs (defer to future task)
- **Client SDKs**: Multi-language client generation (focus on Go only)
- **API Versioning Strategy**: Comprehensive versioning plan (focus on /v1 paths)

---

## Deliverables

### 1. AuthZ OpenAPI Specification

**File**: `api/identity/openapi_spec_authz.yaml`

**Content**:

```yaml
openapi: 3.0.3
info:
  title: OAuth 2.1 Authorization Server API
  version: 1.0.0
  description: OAuth 2.1 and OpenID Connect Authorization Server
  contact:
    name: Identity Team
    email: identity@cryptoutil.local

servers:
  - url: https://localhost:8080
    description: Local Development
  - url: https://authz.cryptoutil.local
    description: Production

paths:
  /oauth2/v1/authorize:
    post:
      summary: OAuth 2.1 Authorization Request
      description: Initiate authorization code flow with PKCE
      operationId: authorize
      tags:
        - OAuth 2.1
      requestBody:
        required: true
        content:
          application/x-www-form-urlencoded:
            schema:
              type: object
              required:
                - response_type
                - client_id
                - redirect_uri
                - code_challenge
                - code_challenge_method
              properties:
                response_type:
                  type: string
                  enum: [code]
                client_id:
                  type: string
                  format: uuid
                redirect_uri:
                  type: string
                  format: uri
                scope:
                  type: string
                  example: "openid profile email"
                state:
                  type: string
                code_challenge:
                  type: string
                  minLength: 43
                  maxLength: 128
                code_challenge_method:
                  type: string
                  enum: [S256]
      responses:
        '302':
          description: Redirect to IdP login or redirect_uri with authorization code
          headers:
            Location:
              schema:
                type: string
                format: uri
        '400':
          $ref: '#/components/responses/OAuth2Error'
        '401':
          $ref: '#/components/responses/OAuth2Error'

  /oauth2/v1/token:
    post:
      summary: OAuth 2.1 Token Endpoint
      description: Exchange authorization code or refresh token for access token
      operationId: token
      tags:
        - OAuth 2.1
      security:
        - clientBasicAuth: []
        - clientSecretPost: []
      requestBody:
        required: true
        content:
          application/x-www-form-urlencoded:
            schema:
              type: object
              required:
                - grant_type
              properties:
                grant_type:
                  type: string
                  enum: [authorization_code, refresh_token, client_credentials]
                code:
                  type: string
                  description: Authorization code (for authorization_code grant)
                redirect_uri:
                  type: string
                  format: uri
                code_verifier:
                  type: string
                  description: PKCE code verifier
                refresh_token:
                  type: string
                  description: Refresh token (for refresh_token grant)
      responses:
        '200':
          description: Successful token response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TokenResponse'
        '400':
          $ref: '#/components/responses/OAuth2Error'
        '401':
          $ref: '#/components/responses/OAuth2Error'

  /health:
    get:
      summary: Health Check
      description: Service health status
      operationId: healthCheck
      tags:
        - Health
      responses:
        '200':
          description: Service is healthy
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HealthResponse'
        '503':
          description: Service is unhealthy
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HealthResponse'

components:
  schemas:
    TokenResponse:
      type: object
      required:
        - access_token
        - token_type
        - expires_in
      properties:
        access_token:
          type: string
          description: OAuth 2.0 access token
        token_type:
          type: string
          enum: [Bearer]
        expires_in:
          type: integer
          description: Token lifetime in seconds
        refresh_token:
          type: string
          description: Refresh token (optional)
        id_token:
          type: string
          description: OpenID Connect ID Token (if openid scope requested)
        scope:
          type: string
          description: Granted scopes

    OAuth2Error:
      type: object
      required:
        - error
      properties:
        error:
          type: string
          enum:
            - invalid_request
            - invalid_client
            - invalid_grant
            - unauthorized_client
            - unsupported_grant_type
            - invalid_scope
        error_description:
          type: string
        error_uri:
          type: string
          format: uri

    HealthResponse:
      type: object
      required:
        - status
      properties:
        status:
          type: string
          enum: [healthy, unhealthy]
        database:
          type: string
          enum: [ok, error]
        uptime:
          type: integer
          description: Uptime in seconds

  securitySchemes:
    clientBasicAuth:
      type: http
      scheme: basic
      description: Client ID and secret via HTTP Basic Authentication
    clientSecretPost:
      type: apiKey
      in: formData
      name: client_secret
      description: Client secret in request body
```

**Tests**: Validate with `swagger-cli validate api/identity/openapi_spec_authz.yaml`

### 2. IdP OpenAPI Specification

**File**: `api/identity/openapi_spec_idp.yaml`

**Content**:

```yaml
openapi: 3.0.3
info:
  title: OpenID Connect Identity Provider API
  version: 1.0.0
  description: OIDC Identity Provider for user authentication

servers:
  - url: https://localhost:8081
    description: Local Development

paths:
  /oidc/v1/login:
    get:
      summary: Login Form
      description: Render user authentication form
      operationId: getLoginForm
      tags:
        - Authentication
      parameters:
        - name: return_url
          in: query
          required: true
          schema:
            type: string
            format: uri
          description: Redirect URL after successful authentication
      responses:
        '200':
          description: Login form HTML
          content:
            text/html:
              schema:
                type: string

    post:
      summary: Authenticate User
      description: Process login credentials
      operationId: authenticateUser
      tags:
        - Authentication
      requestBody:
        required: true
        content:
          application/x-www-form-urlencoded:
            schema:
              type: object
              required:
                - username
                - password
                - csrf_token
              properties:
                username:
                  type: string
                password:
                  type: string
                  format: password
                csrf_token:
                  type: string
                return_url:
                  type: string
                  format: uri
      responses:
        '302':
          description: Redirect to return_url with session cookie
          headers:
            Location:
              schema:
                type: string
                format: uri
            Set-Cookie:
              schema:
                type: string
        '401':
          description: Invalid credentials
          content:
            text/html:
              schema:
                type: string

  /health:
    get:
      summary: Health Check
      operationId: healthCheckIdP
      tags:
        - Health
      responses:
        '200':
          description: Service is healthy
          content:
            application/json:
              schema:
                $ref: './openapi_spec_authz.yaml#/components/schemas/HealthResponse'
```

### 3. RS OpenAPI Specification Validation

**File**: `api/identity/openapi_spec_rs.yaml` (validate existing)

**Actions**:

- Review current Resource Server spec
- Validate all endpoints documented
- Add examples for OAuth 2.0 Bearer token usage
- Document scope requirements per endpoint

### 4. Shared Components

**File**: `api/identity/openapi_spec_components.yaml`

**Content**:

- Common schemas (HealthResponse, ErrorResponse, PaginationParams)
- Security schemes (OAuth2, Bearer token)
- Common responses (401 Unauthorized, 403 Forbidden, 500 Internal Server Error)

**Refactoring**:

- Extract duplicated schemas from service-specific specs
- Use `$ref` to reference shared components
- Validate all references resolve correctly

### 5. Client Code Generation

**Files**: `api/openapi-gen_config_*.yaml`

**Validation**:

- Ensure configs reference new spec files
- Regenerate client code: `go generate ./api/...`
- Fix any compilation errors in generated code
- Update server handler registrations if needed

**Generated Files**:

- `api/identity/client/openapi_gen_client.go`
- `api/identity/model/openapi_gen_model.go`
- `api/identity/server/openapi_gen_server.go`

**Tests**: `go test ./api/identity/...`

### 6. Swagger UI Integration

**Files**: Swagger UI configuration (if separate deployment)

**Actions**:

- Update Swagger UI `urls` config to include identity specs
- Configure proper `basePath` and `servers`
- Set up OAuth 2.0 authorization in Swagger UI (client credentials for testing)
- Test interactive API exploration

**Verification**:

- Navigate to Swagger UI: `https://localhost:8080/ui/swagger`
- Verify all identity endpoints visible
- Test "Try it out" functionality for /health endpoints
- Document OAuth flow testing in Swagger UI

### 7. Contract Testing Suite

**File**: `internal/identity/contract/contract_test.go` (new package)

**Functionality**:

- Load OpenAPI spec from YAML files
- Start identity services via unified CLI
- Use schemathesis/openapi-fuzzer to generate test requests
- Execute requests against running servers
- Validate responses match spec schemas
- Report violations

**Implementation**:

```go
package contract_test

import (
    "testing"
    "github.com/getkin/kin-openapi/openapi3"
    // Contract testing library
)

func TestAuthZSpecMatchesServer(t *testing.T) {
    // Load spec
    loader := openapi3.NewLoader()
    spec, err := loader.LoadFromFile("api/identity/openapi_spec_authz.yaml")
    require.NoError(t, err)

    // Start services
    // TODO: Use unified CLI to start services

    // For each path in spec
    for path, pathItem := range spec.Paths {
        // For each operation (GET, POST, etc.)
        // Generate request based on spec
        // Execute request
        // Validate response matches spec schema
    }
}
```

**Tests**: `go test ./internal/identity/contract/...`

### 8. Documentation

**File**: `docs/identityV2/openapi-guide.md`

**Content**:

- OpenAPI file structure and organization
- Client code generation workflow
- Swagger UI usage guide
- Contract testing approach
- Spec maintenance guidelines
- How to add new endpoints to specs

**File**: `README.md` (update API documentation section)

**Changes**:

```markdown
## API Documentation

Identity services provide OpenAPI 3.0 specifications:

- AuthZ: `api/identity/openapi_spec_authz.yaml`
- IdP: `api/identity/openapi_spec_idp.yaml`
- RS: `api/identity/openapi_spec_rs.yaml`

View interactive docs:
```bash
./identity start --profile demo
# Navigate to https://localhost:8080/ui/swagger
```

Regenerate client code:

```bash
go generate ./api/identity/...
```

```

---

## Validation Criteria

### Automated Tests

- ✅ OpenAPI specs validate: `swagger-cli validate api/identity/*.yaml`
- ✅ Client code regenerates: `go generate ./api/identity/...`
- ✅ Generated code compiles: `go build ./api/identity/...`
- ✅ Contract tests pass: `go test ./internal/identity/contract/...`
- ✅ Linting passes: `golangci-lint run`

### Manual Testing

1. **Spec Validation**:

   ```bash
   npm install -g @apidevtools/swagger-cli
   swagger-cli validate api/identity/openapi_spec_authz.yaml
   # Expect: No errors
   ```

2. **Swagger UI Exploration**:

   ```bash
   ./identity start --profile demo
   # Navigate to https://localhost:8080/ui/swagger
   # Verify all endpoints visible
   # Test /health endpoint via "Try it out"
   ```

3. **Client Code Generation**:

   ```bash
   cd api/identity
   oapi-codegen -config openapi-gen_config_client.yaml openapi_spec_authz.yaml > client/openapi_gen_client.go
   # Expect: File generated with no errors

   go build ./client/...
   # Expect: Compilation succeeds
   ```

4. **Contract Testing**:

   ```bash
   go test ./internal/identity/contract/... -v
   # Expect: All contract tests pass
   ```

### Success Metrics

- All OpenAPI specs validate against OpenAPI 3.0.3 schema
- Generated client code compiles with zero errors
- Contract tests pass (server responses match spec schemas)
- Swagger UI displays all identity endpoints correctly
- No spec drift between documentation and implementation

---

## Dependencies

### Depends On (Must Be Complete)

- ✅ **Task 10.5**: Endpoints must be implemented to document
- ✅ **Task 10.6**: Unified CLI simplifies service startup for testing

### Enables (Blocked Until Complete)

- **Tasks 11-15**: Feature additions require up-to-date specs for client generation
- **External Integrations**: Third parties need accurate specs for client SDKs
- **Task 18**: E2E testing benefits from contract tests validating spec compliance

---

## Known Risks

1. **Spec Drift Over Time**
   - **Risk**: Specs become outdated as endpoints change
   - **Mitigation**: Contract tests catch drift; add pre-commit hook to validate specs

2. **Complex Schema Modeling**
   - **Risk**: OAuth 2.1 request/response patterns complex (form-encoded requests, error schemas)
   - **Mitigation**: Reference RFC 6749 examples; use existing cryptoutil OpenAPI patterns

3. **Code Generation Brittleness**
   - **Risk**: oapi-codegen may generate incompatible code with spec changes
   - **Mitigation**: Pin oapi-codegen version; test generation in CI/CD pipeline

4. **Circular References in Schemas**
   - **Risk**: Shared components may create circular $ref dependencies
   - **Mitigation**: Flatten nested schemas; validate with swagger-cli before committing

---

## Implementation Notes

### Phased Approach

1. **Phase 1**: Create AuthZ OpenAPI spec with /authorize, /token, /health endpoints
2. **Phase 2**: Create IdP OpenAPI spec with /login, /health endpoints
3. **Phase 3**: Validate RS spec (already exists)
4. **Phase 4**: Extract shared components to openapi_spec_components.yaml
5. **Phase 5**: Regenerate client code and fix compilation errors
6. **Phase 6**: Set up contract testing infrastructure
7. **Phase 7**: Update Swagger UI configuration

### Code Organization

- **Spec Files**: `api/identity/*.yaml` (one per service + shared components)
- **Generation Configs**: `api/identity/openapi-gen_config_*.yaml`
- **Generated Code**: `api/identity/{client,model,server}/openapi_gen_*.go`
- **Contract Tests**: `internal/identity/contract/contract_test.go`

### Testing Strategy

- **Spec Validation**: swagger-cli in CI/CD pipeline
- **Generation Testing**: `go generate` in pre-commit hooks
- **Contract Testing**: Automated validation of spec vs server
- **Manual Testing**: Swagger UI for interactive exploration

---

## Exit Criteria

- [ ] AuthZ OpenAPI spec complete with all endpoints
- [ ] IdP OpenAPI spec complete with all endpoints
- [ ] RS OpenAPI spec validated and updated if needed
- [ ] Shared components extracted to openapi_spec_components.yaml
- [ ] Client code regenerated successfully
- [ ] Contract tests passing
- [ ] Swagger UI displays all identity endpoints
- [ ] Documentation complete (openapi-guide.md, README updates)
- [ ] Linting passes with zero violations
- [ ] Code review complete
- [ ] Commit with message: `feat(identity): complete task 10.7 - openapi synchronization`

---

## References

- [OpenAPI 3.0.3 Specification](https://spec.openapis.org/oas/v3.0.3)
- [oapi-codegen Documentation](https://github.com/deepmap/oapi-codegen)
- [RFC 6749 - OAuth 2.0 Authorization Framework](https://www.rfc-editor.org/rfc/rfc6749)
- [OpenID Connect Core 1.0](https://openid.net/specs/openid-connect-core-1_0.html)
- [kin-openapi Validator](https://github.com/getkin/kin-openapi)
- [Schemathesis Contract Testing](https://schemathesis.readthedocs.io/)
- `api/openapi_spec_*.yaml` - Existing cryptoutil OpenAPI specs
- `docs/identityV2/task-10.5-authz-idp-endpoints.md` - Endpoint implementations
