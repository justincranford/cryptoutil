---
description: "Instructions for Docker and Docker Compose configuration"
applyTo: "**/*.yml"
---
# Docker Configuration Instructions

- Prefer using command directives in Docker Compose over separate script files
- Use Docker Compose standard practices for container configuration
- Set environment variables and use command substitution in Docker Compose commands
- Use shell form of commands for complex operations in Docker Compose
- Always use container networking with appropriate network settings
- Properly handle secrets and configuration using Docker's native mechanisms
- For port publishing, use explicit host:container port mappings when fixed ports are needed
