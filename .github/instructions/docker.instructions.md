---
description: "Instructions for Docker and Docker Compose configuration"
applyTo: "**/*.yml"
---
# Docker Configuration Instructions

- Follow [Docker Compose docs](https://docs.docker.com/compose/) for standard practices
- Prefer command directives over scripts; use container networking, secrets, and explicit port mappings as needed
- Use `docker compose` (not `docker-compose`)

## Multi-Stage Build Best Practices

### ARG Scoping Rules
- **Global ARGs**: Declare all build parameters at the top of Dockerfile for visibility and overrideability
- **Stage ARGs**: Redeclare ARGs in stages where they're used in LABEL instructions (Docker requirement)
- **LABEL Placement**: Put ALL LABELs on final published image, not intermediate stages
- **Build ARGs**: Move build-specific ARGs (CGO_ENABLED, GOOS, etc.) to global section for consistency

### LABEL Instructions
- **Final Image Only**: LABELs belong on the published artifact, not intermediate build stages
- **Comprehensive Metadata**: Include source, version, revision, title, description, created, authors
- **ARG Redeclaration**: Always redeclare ARGs in final stage before using in LABEL instructions

### Example Structure
```dockerfile
# Global ARGs - All build parameters visible at top
ARG GO_VERSION=1.25.1
ARG ALPINE_VERSION=3.19
ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
ARG LDFLAGS="-s -w"
ARG APP_VERSION=dev
ARG VCS_REF=unspecified
ARG BUILD_DATE=1970-01-01T00:00:00Z

FROM golang:${GO_VERSION} AS builder
# Clean intermediate stage - no LABELs, minimal ARGs

FROM scratch
# Stage ARGs required for LABEL instructions
ARG APP_VERSION=dev
ARG VCS_REF=unspecified
ARG BUILD_DATE=1970-01-01T00:00:00Z

# All metadata LABELs on final published image
LABEL org.opencontainers.image.source="https://github.com/justincranford/cryptoutil"
LABEL org.opencontainers.image.version="${APP_VERSION}"
LABEL org.opencontainers.image.revision="${VCS_REF}"
LABEL org.opencontainers.image.title="cryptoutil"
LABEL org.opencontainers.image.description="A small utility for cryptographic key and certificate operations"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.authors="Justin Cranford <justin@example.com>"
```

## Hadolint Best Practices

- **Prefer inline ignore comments** over pre-commit config parameters
- Use `# hadolint ignore=DLXXXX` comments directly above the offending line
- **Append explanations on the same line** after another hash symbol for conciseness
- Document the reason for ignoring rules when the ignore provides security/maintainability benefits
- Examples:
  ```dockerfile
  # Preferred: Same-line explanation
  # hadolint ignore=DL3018 # Intentionally unpinned for automatic security updates
  RUN apk --no-cache add ca-certificates tzdata tini

  # Alternative: Multi-line for complex explanations
  # Intentionally unpinned for automatic security updates and complex reasoning
  # hadolint ignore=DL3018
  RUN apk --no-cache add ca-certificates tzdata tini
  ```
