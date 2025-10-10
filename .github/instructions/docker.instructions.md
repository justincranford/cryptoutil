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
- **Append explanations on the same line** after another `#` for conciseness
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
