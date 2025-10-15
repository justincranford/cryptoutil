---
description: "Instructions for security implementation patterns"
applyTo: "**"
---
# Security Implementation Instructions

- Multi-layer barrier: unseal → root → intermediate → content keys
- Hierarchical key management, encrypted at rest
- IP allowlisting (IPs & CIDR), per-IP rate limiting
- CORS, CSRF, strict HTTP headers, audit logging
- Multiple unseal modes, key versioning/rotation, secure failure modes
- Always use crypto/rand, never math/rand
- Full cert chain validation, MinVersion: TLS 1.2+, never InsecureSkipVerify
- Use proper secret management for Docker, Kubernetes, CI/CD
- **ALWAYS prefer Docker secrets or Kubernetes secrets over environment variables for sensitive data**
- **NEVER use environment variables for secrets in production deployments**
- Use Docker secrets mounted to `/run/secrets/` with file:// URLs in command arguments
- Use Kubernetes secrets mounted as files or referenced directly, not via environment variables
- Scanning: run scripts/security-scan.{ps1,sh} before commits/high-risk changes; see README for scan options and workflow
