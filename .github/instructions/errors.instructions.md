---
description: "Instructions for error reporting"
applyTo: "**"
---
# Error Reporting Instructions

- Ensure all errors are logged using integrated telemetry/logging
- Surface errors to API clients via OpenAPI responses
- Avoid silent failures and fallback values in cryptographic operations
- Review error handling in pool, keygen, and server modules
- All errors must be handled (no errcheck violations)
- Follow established patterns for error handling and validation
