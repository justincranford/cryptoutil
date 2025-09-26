---
description: "Instructions for security implementation patterns"
applyTo: "**"
---
# Security Implementation Instructions

- Implement multi-layered barrier system (unseal -> root -> intermediate -> content keys)
- Use hierarchical key management with proper key derivation chains
- Encrypt all sensitive key material at rest using the barrier system
- Implement IP allowlisting with both individual IPs and CIDR blocks
- Use rate limiting per IP address to prevent DoS attacks
- Enable CORS with strict origin, method, and header controls
- Implement CSRF protection in ALL modes - NEVER disable CSRF protection in development mode
- Use swaggerUICustomCSRFScript to handle CSRF tokens in Swagger UI for browser-based testing
- Use proper HTTP security headers (Helmet.js equivalent)
- Support multiple unseal modes: simple keys, shared secrets (M-of-N), and system fingerprinting
- Implement proper key versioning and rotation capabilities
cryptographic operations (crypto/rand, not math/rand)
- Implement comprehensive audit logging for all security operations
- Use proper secret management for Docker deployments
- Ensure graceful degradation and secure failure modes
- Always use full certificate chain validation - NEVER set InsecureSkipVerify: true
- Always set MinVersion: tls.VersionTLS12 (or higher) for all TLS configurations
- Use proper root CA pools and certificate validation in all TLS connections
- Use full certificate chain validation with TLS 1.2+ minimum
