---
description: "Instructions for Docker and Docker Compose configuration"
applyTo: "**/*.yml"
---
# Docker Configuration Instructions

- Follow [Docker Compose docs](https://docs.docker.com/compose/) for standard practices
- Prefer command directives over scripts; use container networking, secrets, and explicit port mappings as needed
- Use `docker compose` (not `docker-compose`)

## Hadolint Best Practices

- **Prefer inline ignore comments** over pre-commit config parameters
- Use `# hadolint ignore=DLXXXX` comments directly above the offending line
- Document the reason for ignoring rules when the ignore provides security/maintainability benefits
- **Inline explanations supported**: You can add explanations on the same line after another `#`
- Examples:
  ```dockerfile
  # hadolint ignore=DL3018 # Intentionally unpinned for automatic security updates
  RUN apk --no-cache add ca-certificates tzdata tini

  # Or use multi-line format:
  # Intentionally unpinned for automatic security updates
  # hadolint ignore=DL3018
  RUN apk --no-cache add ca-certificates tzdata tini
  ```
