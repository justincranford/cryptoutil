---
name: openapi-codegen
description: "Generate oapi-codegen configuration files and OpenAPI 3.0.3 spec skeletons for cryptoutil services. Use when creating or extending service APIs to produce the three standard configs (server/model/client) and a baseline spec with dual /service/ and /browser/ paths."
argument-hint: "[service-name]"
---

Generate oapi-codegen configuration files and OpenAPI spec skeletons for cryptoutil services.

## Purpose

Use when creating a new service or adding API endpoints. Generates the 3 standard
oapi-codegen config files and a baseline OpenAPI 3.0.3 spec.

## Key Rules

- OpenAPI version MUST be 3.0.3 (NOT 2.0/Swagger, NOT 3.1.x)
- Generate THREE config files: server (`strict-server: true`), model, client
- API MUST duplicate under BOTH `/service/` and `/browser/` paths
- Content type: `application/json` ONLY (no form, multipart, or other types)
- `strict-server: true` is MANDATORY in server config
- All `openapi-gen_config*.yaml` MUST include the full base initialisms list from ARCHITECTURE.md §8

## Three Config Files Per Service

### 1. Server Config: `openapi-gen_config_server.yaml`

```yaml
package: server
generate:
  strict-server: true
  embedded-spec: true
output: api/server/server.gen.go
```

### 2. Model Config: `openapi-gen_config_model.yaml`

```yaml
package: model
generate:
  models: true
output: api/model/models.gen.go
```

### 3. Client Config: `openapi-gen_config_client.yaml`

```yaml
package: client
generate:
  client: true
  models: true
output: api/client/client.gen.go
```

## OpenAPI Spec Skeleton

`openapi_spec_paths.yaml`:

```yaml
openapi: "3.0.3"
info:
  title: SERVICE-NAME API
  version: "1.0"
paths:
  /service/api/v1/resources:
    get:
      operationId: listResources
      summary: List resources
      parameters:
        - name: page
          in: query
          schema: {type: integer, default: 1, minimum: 1}
        - name: size
          in: query
          schema: {type: integer, default: 50, minimum: 1, maximum: 1000}
      responses:
        "200":
          description: Success
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ResourceListResponse"
        "400":
          $ref: "openapi_spec_components.yaml#/components/responses/BadRequest"
        "500":
          $ref: "openapi_spec_components.yaml#/components/responses/InternalServerError"
```

`openapi_spec_components.yaml`:

```yaml
components:
  schemas:
    Error:
      type: object
      required: [code, message]
      properties:
        code: {type: string}
        message: {type: string}
        details: {type: object, additionalProperties: true}
        requestId: {type: string, format: uuid}
    Pagination:
      type: object
      required: [page, size, total]
      properties:
        page: {type: integer}
        size: {type: integer}
        total: {type: integer}
  responses:
    BadRequest:
      description: Validation error
      content:
        application/json:
          schema: {$ref: "#/components/schemas/Error"}
    InternalServerError:
      description: Internal server error
      content:
        application/json:
          schema: {$ref: "#/components/schemas/Error"}
```

## Mandatory Checklist

- [ ] `openapi-gen_config_server.yaml` created with `strict-server: true`, output `api/server/server.gen.go`
- [ ] `openapi-gen_config_model.yaml` created with `models: true`, output `api/model/models.gen.go`
- [ ] `openapi-gen_config_client.yaml` created with `client: true`, output `api/client/client.gen.go`
- [ ] `openapi_spec_paths.yaml` — both `/service/api/v1/` and `/browser/api/v1/` path prefixes present
- [ ] `openapi_spec_components.yaml` — `Error` schema with `code`, `message`, `details`, `requestId` present
- [ ] All list endpoints include `page` (default 1) and `size` (default 50, max 1000) query params
- [ ] `go generate ./api/...` exits 0 cleanly after files are created

## Generate Code

```bash
go generate ./api/...
# Or directly:
oapi-codegen -config openapi-gen_config_server.yaml openapi_spec_paths.yaml
oapi-codegen -config openapi-gen_config_model.yaml openapi_spec_paths.yaml
oapi-codegen -config openapi-gen_config_client.yaml openapi_spec_paths.yaml
```

## References

Read [ARCHITECTURE.md Section 8.1 OpenAPI-First Design](../../../docs/ARCHITECTURE.md#81-openapi-first-design) for strict-server requirements and code generation patterns — ensure all three config files (server/model/client) are generated with `strict-server: true` and correct output paths.
Read [ARCHITECTURE.md Section 8.4 Error Handling](../../../docs/ARCHITECTURE.md#84-error-handling) for HTTP status codes and error schema — apply the standard error schema and status code table when generating response definitions.
