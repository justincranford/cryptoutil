---
description: "Instructions for documentation organization and structure"
applyTo: "**"
---
# Documentation Instructions

## Documentation Organization

- Keep all documentation consolidated in **two main files only**:
  1. **README.md** - Main project documentation (quick start, configuration, deployment)
  2. **docs/README.md** - Comprehensive architectural documentation (detailed technical deep dive)

- **DO NOT** create separate documentation files like:
  - `docs/API-ARCHITECTURE.md`
  - `docs/SECURITY.md` 
  - `docs/DEPLOYMENT.md`
  - `docs/CONFIGURATION.md`

## Content Distribution

### README.md should contain:
- Project overview and key features
- Quick start guide
- API architecture with context paths hierarchy diagram
- Configuration examples and best practices
- Deployment instructions (Docker Compose, Kubernetes)
- Testing procedures
- Development workflow

### docs/README.md should contain:
- Comprehensive architectural deep dive
- Detailed security implementation
- Cryptographic architecture details
- Performance and scalability information
- Recent enhancements and improvements
- Architectural strengths and use cases

## Documentation Principles

- Consolidate related content into logical sections within the two main files
- Use clear headings and subheadings for organization
- Include diagrams and code examples inline
- Maintain consistency between the two documentation files
- Avoid duplication of content between files
- Keep documentation focused and user-friendly
