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
- Implement CSRF protection in production mode (disabled in dev mode for testing)
- Use proper HTTP security headers (Helmet.js equivalent)
- Support multiple unseal modes: simple keys, shared secrets (M-of-N), and system fingerprinting
- Use UUIDv7-based key identifiers for all cryptographic material
- Implement proper key versioning and rotation capabilities
- Use secure random number generation for all cryptographic operations
- Validate all cryptographic parameters against FIPS 140-3 requirements
- Implement comprehensive audit logging for all security operations
- Use proper secret management for Docker deployments
- Ensure graceful degradation and secure failure modes
