# Grooming Session 01: Docker Compose Strategy Refinement

## Overview

- **Focus Area**: Refining the Docker Compose strategy based on recent analysis and learnings
- **Related Spec Section**: DOCKER-COMPOSE-STRATEGY.doc
- **Prerequisites**: Understanding of current Docker Compose setup and identified inconsistencies

## Questions

### Q1: What should be the primary profile strategy for the root compose.yml file?

A) Separate profiles for each component (telemetry, kms, identity)
B) Combination profiles for common use cases (kms-only, identity-only, kms-identity)
C) Single profile with optional services based on environment variables
D) No profiles, use separate compose files for each combination

**Answer**: A and B simultaneously if possible, otherwise just A

**Notes**:

```
NEVER C
```

### Q2: How should service naming be standardized across product compose files?

A) product-service-instance (e.g., kms-sqlite-1, identity-authz-1)
B) service-product-instance (e.g., sqlite-kms-1, authz-identity-1)
C) product-instance-service (e.g., kms-1-sqlite, identity-1-authz)
D) Keep current inconsistent naming as it reflects different architectures

**Answer**: A

**Notes**:
```

```

### Q3: What is the best approach for handling federation dependencies between products?

A) Hard dependencies in compose files (services fail if peer product not running)
B) Soft dependencies with health checks (services start but warn if peer unavailable)
C) Configuration-based federation (services always start, federation enabled via config)
D) No federation support in compose, handle at application level only

**Answer**: A or C

**Notes**:
```
Note sure how to choose between A or C
```

### Q4: How should default data seeding be implemented for demo profiles?

A) SQL migration files executed on container startup
B) Pre-built database images with seeded data
C) Application-level seeding on first run with special flags
D) Manual seeding scripts that users must run separately

**Answer**: A

**Notes**:
```
See KMS Dockerfile, kms-postgres-1 runs first and applies migrations. Then kms-postgres-2 waits until kms-postgres-1 is healthy (i.e. it applies the SQL migration files).
Expanding that pattern to N, kms-postgres-1 runs first and applies migrations. Then all kms-postgres-2 through kms-postgres-N wait for kms-postgres-1 to be healthy (i.e. 2 through N can all start concurrently).
```

### Q5: What constitutes "sensible defaults" for user credentials and data?

A) Well-known test values (admin/admin, test@example.com)
B) Generated UUIDs and random values for each deployment
C) Environment-specific values based on profile names
D) No defaults, require explicit configuration for all values

**Answer**: A and B

**Notes**:
```
User can copy-n-paste, and remove the set of A or B that they don't want. Update/delete is always more intuitive than creating from scratch.
```

### Q6: How should the SPA RP service be integrated into the identity compose?

A) Same image as other identity services (unified binary)
B) Separate image built from Dockerfile.spa
C) External service not included in compose (run separately)
D) Replace with KMS Swagger UI for traditional backend pattern

**Answer**: B

**Notes**:
```

```

### Q7: What is the risk of using the same builder for multiple different binaries?

A) No risk, Docker caching handles it efficiently
B) Build conflicts and incorrect binaries in images
C) Increased build time but correct functionality
D) Security issues with shared build contexts

**Answer**:

**Notes**:
```
I don't know how to answer this the question in this form.
Identity services authz and idp can be in a single binary for sure, and the product start command can differentiate them.
Definitely spa should not be in same binary.
Not sure if rs should be in same binary as authz and idp, or separate.
In practice, I think authz and idp are releasable in future. I don't know if rs is releasable too. WDYT?
```

### Q8: How should profile dependencies be validated in CI/CD?

A) Automated tests that start each profile combination
B) Static analysis of compose files for dependency cycles
C) Manual testing with checklists for each profile
D) No validation needed, rely on user testing

**Answer**: A

**Notes**:
```

```

### Q9: What is the most efficient way to share secrets across product compose files?

A) Duplicate secret files in each product directory
B) Symlinks to shared secrets directory
C) Centralized secrets with product-specific subdirectories
D) Environment variables instead of file-based secrets

**Answer**: C

**Notes**:
```
NEVER D. NEVER ENVs!
```

### Q10: How should telemetry be included in demo profiles?

A) Always include telemetry even in demo (for observability learning)
B) Exclude telemetry from demo to reduce resource usage
C) Optional telemetry based on environment variable
D) Separate demo-telemetry profile combination

**Answer**: D

**Notes**:
```

```

### Q11: What is the best way to handle port conflicts between products?

A) Fixed port assignments per product (KMS: 8080-8099, Identity: 8100-8199)
B) Dynamic port allocation with published ports
C) User-configurable ports via environment variables
D) Single port per service type across all products

**Answer**: A

**Notes**:
```

```

### Q12: How should database migrations be handled in compose?

A) Automatic migration on container startup
B) Manual migration scripts in documentation
C) Pre-migrated database images
D) Migration as separate service in compose

**Answer**: None of the above

**Notes**:
```
See KMS Dockerfile, kms-postgres-1 runs first and applies migrations. Then kms-postgres-2 waits until kms-postgres-1 is healthy (i.e. it applies the SQL migration files).
Expanding that pattern to N, kms-postgres-1 runs first and applies migrations. Then all kms-postgres-2 through kms-postgres-N wait for kms-postgres-1 to be healthy (i.e. 2 through N can all start concurrently).
```

### Q13: What level of health checking is appropriate for demo profiles?

A) Full health checks with dependencies
B) Basic health checks without external dependencies
C) No health checks for faster startup
D) Health checks only for critical services

**Answer**: A

**Notes**:
```
Demo is for demonstrating working functionality, so the subset of included functionality must be fully working, like if it was
enabled in full integration or future production mode.
```

### Q14: How should configuration files be organized for different profiles?

A) Profile-specific config files (config/demo.yml, config/full.yml)
B) Single config with profile-based overrides
C) Environment variables for profile differences
D) No profile-specific configs, use defaults

**Answer**:

**Notes**:
```
Configs are service specific, plus there is a common one. See KMS deployment configs.
Within each service, it would be OK to extend the service specific ones to be per profile.
```

### Q15: What is the risk of making configurations "too intuitive"?

A) Users won't read documentation and miss important details
B) Increased maintenance burden on developers
C) Security issues from exposed defaults
D) No significant risk, intuition is always better

**Answer**: D, C

**Notes**:
```
During startup, maybe there is a way for services to detect and log if they are using well known defaults that pose a security risk.
For example, a could a job as the last compose service/job to start and finish, and it can log whether services are using well known
defaults for any settings that are a security risk?
```

### Q16: How should the transition from old to new compose structure be managed?

A) Immediate replacement with migration guide
B) Gradual migration with both structures supported
C) Keep old structure for backward compatibility
D) No transition needed, new structure is additive

**Answer**: None of the above

**Notes**:
```
IMMEDIATE REPLACEMENT.
NO MIGRATION NEEDED.
FIX AND VERIFY ALL REFERENCES IN WORKFLOWS, DOCS, EVERYTHING REFERENCE THE NEW DOCKER COMPOSE STRATEGY.
```

### Q17: What metrics should be used to validate "self-documenting" configurations?

A) Time for new team member to start services without help
B) Number of support tickets about compose usage
C) Lines of documentation needed for compose files
D) Automated tests passing without manual intervention

**Answer**: D

**Notes**:
```
Everything must work out-of-the-box without any modifications.
Zero documentation needed, everything in deployments MUST be self-evident and intuitive.
```

### Q18: How should multi-architecture builds be handled?

A) Build all architectures in single Dockerfile
B) Separate Dockerfiles per architecture
C) Use buildx for multi-platform builds
D) Single architecture only (amd64)

**Answer**: D

**Notes**:
```
D for now, C if it is ever requested
```

### Q19: What is the most important success criterion for the new structure?

A) Reduced file count and complexity
B) Improved developer experience and productivity
C) Better resource utilization in containers
D) Enhanced security through isolation

**Answer**: B

**Notes**:
```
When I evaluate if you completed a plan, feature, etc I will use `docker compose up -d` and `docker compose down -v`.
If those commands work, and I can intuitively find the UIs and APIs, and successfully use then without error, then that
is the MOST IMPORTANT SUCCESS CRITERION. If any of the above have any errors or warnings that block me, you get an
instant F grade.
```

### Q20: How should the strategy handle future products beyond KMS and Identity?

A) Fixed structure that all products must follow
B) Flexible template that products can adapt
C) Separate compose structure for each new product
D) No planning needed, handle when products are added

**Answer**: A

**Notes**:
```
Definitely A for ease of scaling to potentially dozens of products, without having to learn the ins and outs of each
individual one. I can't think of any outliers that might justify a one-off B or C, so assume A is CRITICAL.
```
