---
description: "Instructions for configuration and application architecture"
applyTo: "**"
---
# Configuration & Architecture

- Go project structure: See project-layout.instructions.md
- Use layered arch: main → app → business logic → repositories
- Config: YAML files & CLI only (no env vars for configuration; use Docker/Kubernetes secrets for sensitive data)
- Dependency injection with context propagation
- Structured config with validation/defaults
- Lifecycle: graceful startup/shutdown, resource cleanup
- Service layer: clear separation of concerns
- Error propagation through layers
- Factory pattern for service init with error handling
- Support local/dev, Docker, prod configs
- Atomic ops for critical state
- Timeout/retry for external deps
- Validate config before startup
- Hot-reload config if needed
