---
description: "Instructions for Go project layout"
applyTo: "**"
---
# Go Project Layout Instructions

Follow the [Standard Go Project Layout](https://github.com/golang-standards/project-layout) for directory structure:

- `/cmd` - Main applications with minimal code that just imports and calls code from other packages
  - Application names should match executable names (e.g., `/cmd/cryptoutil`)
  - Keep main functions small, primarily importing from `/internal` and `/pkg`
  
- `/internal` - Private application and library code
  - `/internal/app` - Application-specific private code
  - `/internal/pkg` - Shared private library code

- `/pkg` - Library code that's safe for external applications to import
  - Only place code here if it's intended to be used by external projects

- `/api` - API definitions, OpenAPI/Swagger specs, JSON schema files, protocol definitions

- `/configs` - Configuration file templates and default configurations
  - Store YAML configuration files here

- `/scripts` - Scripts for build, install, analysis, and other operations

- `/docs` - Design and user documentation

- `/docker` - Container-related files (equivalent to `/deployments` in standard layout)

Use appropriate directory structures for:
- Clear separation of concerns between components
- Proper dependency management
- Testable and maintainable code organization

Avoid:
- Putting application logic directly in `/cmd` packages
- Using a `/src` directory at the project root level
- Deep nesting of packages that creates complex import paths
