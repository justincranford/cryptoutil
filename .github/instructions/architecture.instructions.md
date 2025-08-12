---
description: "Instructions for configuration and application architecture"
applyTo: "**"
---
# Configuration and Architecture Instructions

- Use hierarchical application architecture: main -> application -> business logic -> repositories
- Support configuration via YAML files and command-line parameters (no environment variables)
- Implement proper dependency injection with context propagation
- Use structured configuration with validation and default values
- Implement proper lifecycle management with graceful startup and shutdown
- Use service layer pattern with clear separation of concerns
- Implement proper error propagation through application layers
- Use factory patterns for service initialization with proper error handling
- Support multiple deployment configurations (local dev, Docker, production)
- Implement proper resource cleanup in shutdown handlers
- Use atomic operations for critical state changes
- Implement proper timeout and retry mechanisms for external dependencies
- Implement proper configuration validation before application startup
- Support hot-reloading of configuration where appropriate
