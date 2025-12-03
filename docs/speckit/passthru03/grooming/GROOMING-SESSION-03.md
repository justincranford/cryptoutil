# Grooming Session 03: Docker Compose Implementation Details

## Overview

- **Focus Area**: Technical implementation details for the refined Docker Compose strategy
- **Related Spec Section**: DOCKER-COMPOSE-STRATEGY.doc
- **Prerequisites**: Understanding of grooming sessions 01-02 decisions and refined strategy

## Questions

### Q41: How should shared containers between profiles be implemented technically?

A) Docker Compose extends with profile-specific overrides
B) YAML anchors and aliases for service reuse
C) Separate compose files with shared service definitions
D) Runtime container sharing via Docker networks

**Answer**:

**Notes**:

```text
No idea how to choose A or B.
NOT C for sure.

```

### Q42: What is the best way to implement the migration pattern across products?

A) Health check dependencies in compose (depends_on with condition)
B) Init containers that run migrations before main containers
C) Application-level migration coordination
D) Pre-migrated database images with version tags

**Answer**: A

**Notes**:

```text
NEVER B
NEVER D
C MIGHT BE NICE, BUT HOW? I TRIED USING MIGRATIONS IN KMS, BUT ALWAYS RAN INTO ISSUES. A IS PROBABLY SAFEST.
```

### Q43: How should profile validation be implemented in automated tests?

A) Shell scripts that start/stop each profile combination
B) Go integration tests using testcontainers
C) Docker Compose health check validation only
D) Manual testing checklists converted to automated assertions

**Answer**: None of the above

**Notes**:

```text
Go e2e tests orchestrating `docker compose`
```

### Q44: What logging mechanism should detect default credential usage?

A) Application startup logs with security warnings
B) Separate validation service that scans configurations
C) Environment variable detection in entrypoint scripts
D) Runtime metrics and alerts for default usage

**Answer**: B

**Notes**:

```text

```

### Q45: How should the centralized secrets structure be organized?

A) `secrets/{product}/{service}/` hierarchy
B) `secrets/{product}/shared/` for cross-service secrets
C) Flat structure with product prefixes in filenames
D) Git-ignored local secrets with documented templates

**Answer**: A and B

**Notes**:

```text

```

### Q46: What build optimization provides the best efficiency for multiple products?

A) Multi-stage Dockerfiles with shared base layers
B) BuildKit with caching and parallel builds
C) Pre-built base images stored in registry
D) Monorepo build with selective rebuilds

**Answer**: Probably A

**Notes**:

```text
Not sure about B
```

### Q47: How should service discovery work between federated products?

A) Hardcoded URLs in configuration files
B) Docker Compose service names as hostnames
C) Environment variables populated by compose
D) Service registry with dynamic discovery

**Answer**: B

**Notes**:

```text

```

### Q48: What is the most maintainable way to handle profile-specific configurations?

A) Profile-specific config files
B) Single config with profile-based conditionals
C) Environment variable overrides
D) Separate config directories per profile

**Answer**:

**Notes**:

```text

```

### Q49: How should the fixed port ranges be documented and enforced?

A) Comments in compose files with reserved ranges
B) Validation scripts that check for conflicts
C) Centralized port allocation registry
D) Runtime port conflict detection

**Answer**: A

**Notes**:

```text

```

### Q50: What constitutes a "self-evident" configuration for new developers?

A) Obvious service names and clear purposes
B) Inline documentation in compose files
C) Consistent patterns across all products
D) All of the above

**Answer**: D

**Notes**:

```text

```
