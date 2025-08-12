---
description: "Instructions for OpenAPI and code generation patterns"
applyTo: "**"
---
# OpenAPI and Code Generation Instructions

- Use OpenAPI 3.0.3 specification for all API definitions
- Split OpenAPI specs into components.yaml (schemas) and paths.yaml (endpoints) for maintainability
- Generate all models, servers, and clients using oapi-codegen with specific version pins
- Use strict server interface patterns with proper request/response validation
- Implement comprehensive HTTP status code responses (400, 401, 403, 404, 429, 500, 502, 503, 504)
- Include detailed schema descriptions and examples in OpenAPI components
- Use proper parameter definitions for query parameters with validation
- Implement pagination patterns with page/size parameters
- Use OpenAPI validation middleware for request/response validation
- Generate Swagger UI with CSRF token handling for interactive testing
- Follow REST conventions for resource naming and HTTP methods
- Use JSON content types for all API requests/responses except JWE/JWS operations (text/plain)
- Include comprehensive error response schemas with status, error, and message fields
