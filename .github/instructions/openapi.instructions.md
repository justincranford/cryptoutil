---
description: "Instructions for OpenAPI and code generation patterns"
applyTo: "**"
---
# OpenAPI and Code Generation Instructions

- Use [OpenAPI 3.0.3](https://spec.openapis.org/oas/v3.0.3) for all API specs
- Split specs into components.yaml/paths.yaml; generate code with [oapi-codegen](https://github.com/deepmap/oapi-codegen)
- Use oapi-codegen's strict server pattern
- Use request/response validation, status codes, REST conventions, JSON content types, error schemas, pagination with page/size parameters
