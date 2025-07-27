---
description: "Instructions for error reporting"
applyTo: "**"
---
# Error Reporting Instructions

- Ensure all errors are logged using integrated telemetry/logging.
- Surface errors to API clients via OpenAPI responses.
- Avoid silent failures and fallback values in cryptographic operations.
- Review error handling in pool, keygen, and server modules.
- Always use GORM ORM for database operations, never use sql.DB directly.
