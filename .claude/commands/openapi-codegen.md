---
name: openapi-codegen
description: "Generate oapi-codegen configuration files and OpenAPI 3.0.3 spec skeletons for cryptoutil services. Use when creating or extending service APIs to produce the three standard configs (server/model/client) and a baseline spec with dual /service/ and /browser/ paths."
argument-hint: "[service-name]"
---

Generate three oapi-codegen configuration files and an OpenAPI 3.0.3 spec skeleton for a PS-ID service.

**Full Copilot original**: [.github/skills/openapi-codegen/SKILL.md](.github/skills/openapi-codegen/SKILL.md)

Provide the PS-ID (e.g., `sm-kms`) and list of resources.

## Key Rules

- OpenAPI version MUST be 3.0.3 (NOT 2.0/Swagger, NOT 3.1.x)
- Generate THREE config files: server (`strict-server: true`), model, client
- API MUST duplicate under BOTH `/service/` and `/browser/` paths
- Content type: `application/json` ONLY (no form, multipart, or other types)
- `strict-server: true` is MANDATORY in server config
- All `openapi-gen_config*.yaml` MUST include the full base initialisms list from ARCHITECTURE.md §8

## Three Config Files to Generate

### 1. Server Config (`api/{ps-id}/server-gen-config.yaml`)

```yaml
package: server
generate:
  strict-server: true
  embedded-spec: true
output: api/{ps-id}/server/server.gen.go
output-options:
  skip-prune: false
```

### 2. Model Config (`api/{ps-id}/model-gen-config.yaml`)

```yaml
package: model
generate:
  models: true
output: api/{ps-id}/model/models.gen.go
```

### 3. Client Config (`api/{ps-id}/client-gen-config.yaml`)

```yaml
package: client
generate:
  client: true
  models: true
output: api/{ps-id}/client/client.gen.go
```

## OpenAPI 3.0.3 Spec Skeleton

```yaml
openapi: "3.0.3"
info:
  title: "{PS-ID} API"
  version: "1.0.0"
servers:
  - url: "https://{host}/service/api/v1"
    description: Service API (service-to-service, mTLS)
  - url: "https://{host}/browser/api/v1"
    description: Browser API (session-based, CORS+CSRF)
paths:
  /resources:
    get:
      operationId: ListResources
      summary: List resources
      security:
        - bearerAuth: []
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ResourceList"
        "401":
          $ref: "#/components/responses/Unauthorized"
        "429":
          $ref: "#/components/responses/TooManyRequests"
        "500":
          $ref: "#/components/responses/InternalServerError"
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
  responses:
    Unauthorized:
      description: Unauthorized
    TooManyRequests:
      description: Too Many Requests
    InternalServerError:
      description: Internal Server Error
  schemas:
    ResourceList:
      type: object
      properties:
        items:
          type: array
          items:
            $ref: "#/components/schemas/Resource"
    Resource:
      type: object
      required: [id]
      properties:
        id:
          type: string
          format: uuid
```

## Run Codegen

```bash
oapi-codegen -config api/{ps-id}/server-gen-config.yaml api/{ps-id}/openapi.yaml
oapi-codegen -config api/{ps-id}/model-gen-config.yaml api/{ps-id}/openapi.yaml
oapi-codegen -config api/{ps-id}/client-gen-config.yaml api/{ps-id}/openapi.yaml
```
