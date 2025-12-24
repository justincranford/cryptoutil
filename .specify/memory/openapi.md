# OpenAPI Specifications and Code Generation

**Referenced by**: `.github/instructions/02-06.openapi.instructions.md`

## OpenAPI Version - MANDATORY

**ALWAYS use OpenAPI 3.0.3**:

- Specification: <https://spec.openapis.org/oas/v3.0.3>
- NOT OpenAPI 2.0 (Swagger)
- NOT OpenAPI 3.1.x (adds JSON Schema compatibility but less tooling support)

**Why 3.0.3**:

- Mature ecosystem with wide tool support
- oapi-codegen excellent Go support
- Stable spec (released 2020)
- Industry standard for REST APIs

---

## API Specification Structure

### File Organization Pattern

**Split specifications into separate files**:

- `openapi_spec_components.yaml` - Reusable components (schemas, responses, parameters, examples, requestBodies, headers, securitySchemes, links, callbacks)
- `openapi_spec_paths.yaml` - API endpoints and operations
- Main spec file references these via `$ref`

**Benefits**:

- Easier to review and merge changes (smaller diffs)
- Reduces git conflicts (separate files for components vs paths)
- Enables component reuse across multiple APIs
- Better IDE/editor performance (smaller files)
- Clear separation of concerns

**Example Main Spec**:

```yaml
openapi: 3.0.3
info:
  title: Cryptoutil API
  version: 1.0.0
servers:
  - url: https://localhost:8080/service/api/v1
components:
  $ref: './openapi_spec_components.yaml#/components'
paths:
  $ref: './openapi_spec_paths.yaml#/paths'
```

---

## Code Generation with oapi-codegen

### Generator Tool

**Use oapi-codegen for Go code generation**:

- Repository: <https://github.com/deepmap/oapi-codegen>
- Generates Go code from OpenAPI 3.0 specs
- Supports strict server/client patterns
- Built-in validation and marshaling

### Configuration Files Pattern

**Three separate config files for different outputs**:

**1. Server Configuration** (`openapi-gen_config_server.yaml`):

```yaml
package: server
generate:
  strict-server: true
  models: false  # Use shared models
output: api/server/server.gen.go
```

**2. Model Configuration** (`openapi-gen_config_model.yaml`):

```yaml
package: model
generate:
  models: true
  strict-server: false
  client: false
output: api/model/models.gen.go
```

**3. Client Configuration** (`openapi-gen_config_client.yaml`):

```yaml
package: client
generate:
  client: true
  models: false  # Use shared models
output: api/client/client.gen.go
```

**Benefits**:

- Single source of truth for models (prevents drift)
- Server and client use same model types
- Clear separation of concerns
- Easier to version and maintain

---

## Strict Server Pattern - MANDATORY

**ALWAYS use strict-server mode**:

```yaml
generate:
  strict-server: true
```

**Strict Server Benefits**:

- **Type Safety**: All request/response types are strongly typed
- **Validation**: Request validation happens before handler execution
- **Separation**: Business logic in handlers, validation in generated code
- **Error Handling**: Consistent error responses for validation failures

**Pattern**:

```go
// Generated strict server interface
type StrictServerInterface interface {
    CreateKey(ctx context.Context, request CreateKeyRequest) (CreateKeyResponse, error)
}

// Implementation
type Handler struct {
    // dependencies
}

func (h *Handler) CreateKey(ctx context.Context, request CreateKeyRequest) (CreateKeyResponse, error) {
    // Request already validated by generated code
    // Business logic only
    key, err := h.keyService.Create(ctx, request)
    if err != nil {
        return CreateKey500Response{}, err
    }
    return CreateKey200Response{Key: key}, nil
}
```

---

## Request/Response Validation - MANDATORY

**ALWAYS include validation rules in OpenAPI spec**:

**String Validation**:

```yaml
properties:
  keyId:
    type: string
    format: uuid
    description: Unique key identifier
  algorithm:
    type: string
    enum: [RSA-2048, RSA-3072, ECDSA-P256, ECDSA-P384]
  name:
    type: string
    minLength: 1
    maxLength: 255
    pattern: '^[a-zA-Z0-9_-]+$'
```

**Number Validation**:

```yaml
properties:
  keySize:
    type: integer
    minimum: 2048
    maximum: 4096
    multipleOf: 1024
  rotationDays:
    type: integer
    minimum: 30
    maximum: 365
```

**Array Validation**:

```yaml
properties:
  scopes:
    type: array
    minItems: 1
    maxItems: 50
    uniqueItems: true
    items:
      type: string
      enum: [read, write, admin]
```

**Object Validation**:

```yaml
properties:
  metadata:
    type: object
    required: [createdAt, version]
    properties:
      createdAt:
        type: string
        format: date-time
      version:
        type: integer
        minimum: 1
```

---

## HTTP Status Codes - MANDATORY Standards

**Use appropriate status codes for all responses**:

**Success Codes**:

- `200 OK` - GET, PUT, PATCH successful
- `201 Created` - POST successful (resource created)
- `204 No Content` - DELETE successful, PATCH successful (no content)

**Client Error Codes**:

- `400 Bad Request` - Validation error, malformed request
- `401 Unauthorized` - Missing/invalid authentication
- `403 Forbidden` - Valid auth but insufficient permissions
- `404 Not Found` - Resource does not exist
- `409 Conflict` - Duplicate resource, optimistic lock failure
- `422 Unprocessable Entity` - Semantic validation error

**Server Error Codes**:

- `500 Internal Server Error` - Unhandled server error
- `503 Service Unavailable` - Temporary unavailability (maintenance, overload)

**Example**:

```yaml
paths:
  /keys/{keyId}:
    get:
      responses:
        '200':
          description: Key retrieved successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Key'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '403':
          $ref: '#/components/responses/Forbidden'
        '404':
          $ref: '#/components/responses/NotFound'
        '500':
          $ref: '#/components/responses/InternalServerError'
```

---

## REST Conventions - MANDATORY

**Resource Naming**:

- Use plural nouns for collections: `/keys`, `/certificates`, `/users`
- Use singular for singletons: `/config`, `/health`
- Use kebab-case for multi-word resources: `/api-keys`, `/access-tokens`

**HTTP Method Semantics**:

- `GET /keys` - List all keys (with pagination)
- `POST /keys` - Create new key
- `GET /keys/{keyId}` - Get specific key
- `PUT /keys/{keyId}` - Replace key (full update)
- `PATCH /keys/{keyId}` - Update key (partial update)
- `DELETE /keys/{keyId}` - Delete key

**Idempotency**:

- GET, PUT, DELETE are idempotent (repeated calls have same effect)
- POST is NOT idempotent (creates new resource each time)
- PATCH may or may not be idempotent (depends on operation)

---

## JSON Content Types - MANDATORY

**ALWAYS use application/json**:

```yaml
requestBody:
  required: true
  content:
    application/json:
      schema:
        $ref: '#/components/schemas/CreateKeyRequest'

responses:
  '200':
    description: Success
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/Key'
```

**NEVER use**:

- `text/plain` for structured data
- `application/x-www-form-urlencoded` for complex objects
- `application/xml` (unless legacy requirement)

---

## Error Schemas - MANDATORY

**Consistent error response format**:

```yaml
components:
  schemas:
    Error:
      type: object
      required:
        - code
        - message
      properties:
        code:
          type: string
          description: Machine-readable error code
          example: INVALID_KEY_SIZE
        message:
          type: string
          description: Human-readable error message
          example: Key size must be 2048, 3072, or 4096 bits
        details:
          type: object
          description: Additional error context
          additionalProperties: true
          example:
            field: keySize
            value: 1024
            allowed: [2048, 3072, 4096]
        requestId:
          type: string
          format: uuid
          description: Unique request identifier for troubleshooting
          example: 550e8400-e29b-41d4-a716-446655440000
```

**Usage**:

```yaml
responses:
  '400':
    description: Bad request
    content:
      application/json:
        schema:
          $ref: '#/components/schemas/Error'
```

---

## Pagination Support - MANDATORY

**All list endpoints MUST support pagination**:

**Query Parameters**:

```yaml
parameters:
  - name: page
    in: query
    description: Page number (1-indexed)
    schema:
      type: integer
      minimum: 1
      default: 1
  - name: size
    in: query
    description: Items per page
    schema:
      type: integer
      minimum: 1
      maximum: 1000
      default: 50
```

**Response Schema**:

```yaml
components:
  schemas:
    KeyListResponse:
      type: object
      required:
        - items
        - pagination
      properties:
        items:
          type: array
          items:
            $ref: '#/components/schemas/Key'
        pagination:
          type: object
          required:
            - page
            - size
            - total
          properties:
            page:
              type: integer
              minimum: 1
            size:
              type: integer
              minimum: 1
            total:
              type: integer
              minimum: 0
              description: Total number of items across all pages
```

---

## Cross-References

**Related Documentation**:

- Service template: `.specify/memory/service-template.md`
- HTTPS ports: `.specify/memory/https-ports.md`
- Testing: `.specify/memory/testing.md`

**Tools**:

- oapi-codegen: <https://github.com/deepmap/oapi-codegen>
- OpenAPI 3.0.3: <https://spec.openapis.org/oas/v3.0.3>
- Swagger Editor: <https://editor.swagger.io/>
