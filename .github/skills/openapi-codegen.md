# openapi-codegen

Generate oapi-codegen configuration files and OpenAPI spec skeletons for cryptoutil services.

## Purpose

Use when creating a new service or adding API endpoints. Generates the 3 standard
oapi-codegen config files and a baseline OpenAPI 3.0.3 spec.

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

## Generate Code

```bash
go generate ./api/...
# Or directly:
oapi-codegen -config openapi-gen_config_server.yaml openapi_spec_paths.yaml
oapi-codegen -config openapi-gen_config_model.yaml openapi_spec_paths.yaml
oapi-codegen -config openapi-gen_config_client.yaml openapi_spec_paths.yaml
```

## References

See [ARCHITECTURE.md Section 8.1 OpenAPI-First Design](../../docs/ARCHITECTURE.md#81-openapi-first-design) for strict-server requirements and code generation patterns.
See [ARCHITECTURE.md Section 8.4 Error Handling](../../docs/ARCHITECTURE.md#84-error-handling) for HTTP status codes and error schema.
