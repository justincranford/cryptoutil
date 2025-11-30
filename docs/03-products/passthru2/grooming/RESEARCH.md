# Deployments Directory Structure Analysis

## Current Structure Overview

```
deployments/
├── Dockerfile (shared build)
├── identity/
│   ├── compose.yml (identity services + postgres)
│   ├── demo.yml
│   ├── identity/ (empty)
│   └── postgres/ (empty)
├── kms/
│   ├── compose.yml (kms services + postgres + telemetry)
│   ├── kms/ (configs + secrets)
│   └── postgres/ (secrets)
└── telemetry/
    ├── grafana-otel-lgtm/ (grafana config)
    └── otel/ (otel collector config)
```

## Pros

### 1. Product-Based Organization
- Clear separation by product domain (identity, kms)
- Each product has its own compose.yml for isolated development/testing
- Logical grouping of related services and configurations

### 2. Config Management
- Product-specific configs stored under product directories (kms/kms/)
- Secrets properly isolated per product
- Shared build Dockerfile at root level

### 3. Telemetry Integration
- KMS compose includes full telemetry stack (otel-collector + grafana)
- Identity services reference telemetry endpoints
- Enables observability for development environments

### 4. Scalability Foundation
- Easy to add new products following the same pattern
- Directory structure supports multiple database backends per product
- Compose files can be extended with additional services

## Cons

### 1. Telemetry Coupling
- Telemetry services only defined in KMS compose.yml
- Identity compose references `opentelemetry-collector-contrib:4317` but doesn't define the service
- Cannot run identity stack independently without telemetry (breaks isolation)

### 2. Inconsistent Config Locations
- KMS configs: `./kms/kms/*.yml` (relative to compose.yml)
- Identity configs: `../../configs/identity/*.yml` (outside deployments/)
- Mixed approaches for config file organization

### 3. Empty Directories
- `identity/identity/` and `identity/postgres/` are empty
- Creates confusion about intended structure
- May indicate incomplete refactoring

### 4. Service Dependencies
- Identity services depend on external telemetry without defining it
- No clear way to run product stacks with/without telemetry
- Hard-coded service names across compose files

### 5. Secret Management
- KMS has dedicated postgres/ directory for secrets
- Identity uses inline environment variables instead of secrets
- Inconsistent security practices across products

## Improvements

### 1. Extract Shared Telemetry
- Create `deployments/telemetry/compose.yml` with otel-collector and grafana services
- Use Docker Compose file merging: `docker compose -f telemetry/compose.yml -f kms/compose.yml`
- Allows products to opt-in to telemetry independently

### 2. Standardize Config Locations
- Move all product configs under `deployments/<product>/config/`
- Remove external config references (`../../configs/`)
- Consistent relative paths in compose files

### 3. Clean Up Structure
- Remove empty directories (`identity/identity/`, `identity/postgres/`)
- Add documentation for expected directory contents
- Ensure all directories serve a clear purpose

### 4. Improve Service Isolation
- Use Docker Compose profiles for optional services (telemetry, databases)
- Allow running products with different backend combinations
- Make service names configurable via environment variables

### 5. Enhance Secret Management
- Convert identity postgres credentials to Docker secrets
- Follow same pattern as KMS for all products
- Centralize shared secrets (unseal keys) in separate directory

## Scalability for New Products (cert-authority, jose-authority)

### Current Structure Scalability: ✅ Good

The current product-based directory structure scales well for adding new products:

```
deployments/
├── cert-authority/
│   ├── compose.yml
│   ├── config/ (product configs)
│   ├── secrets/ (product secrets)
│   └── postgres/ (if needed)
└── jose-authority/
    ├── compose.yml
    ├── config/
    ├── secrets/
    └── postgres/
```

### Developer Experience: ✅ Maintainable

- **Clear patterns**: Developers can copy existing product structure
- **Independent development**: Each product can be developed/tested in isolation
- **Shared infrastructure**: Common Dockerfile, telemetry patterns
- **CI/CD integration**: Easy to add product-specific workflows

### Operations: ⚠️ Needs Improvement

- **Shared telemetry**: Currently only in KMS, needs extraction for sharing
- **Service discovery**: Hard-coded service names need parameterization
- **Resource management**: No shared resource limits across products

## Shared Telemetry Compose Feasibility

### Yes, Possible with Docker Compose File Merging

**Current Limitation**: Telemetry services only in `kms/compose.yml`

**Solution**: Extract to `deployments/telemetry/compose.yml`

```bash
# Run KMS with telemetry
docker compose -f deployments/telemetry/compose.yml -f deployments/kms/compose.yml up

# Run Identity with telemetry
docker compose -f deployments/telemetry/compose.yml -f deployments/identity/compose.yml up

# Run product without telemetry
docker compose -f deployments/kms/compose.yml up
```

### Implementation Steps

1. **Create `deployments/telemetry/compose.yml`** with:
   - `opentelemetry-collector-contrib` service
   - `grafana-otel-lgtm` service
   - Shared network: `telemetry-network`
   - Shared volumes for grafana data

2. **Update product compose files**:
   - Remove telemetry services from `kms/compose.yml`
   - Ensure all products use `telemetry-network` for service discovery
   - Update depends_on conditions appropriately

3. **Service Discovery**:
   - Use service names consistent across all compose files
   - Environment variables for OTLP endpoints if needed

### Benefits
- **Flexibility**: Products can run with or without telemetry
- **DRY Principle**: Single definition of telemetry stack
- **Consistency**: Same telemetry setup across all products
- **Maintenance**: Updates to telemetry affect all products automatically

### Challenges
- **Network connectivity**: Ensure services can reach telemetry across merged networks
- **Volume sharing**: Grafana data persistence across different compose runs
- **Health checks**: Dependencies between telemetry and product services

## Recommendations

### Immediate Actions
1. Extract telemetry services to shared compose file
2. Standardize config locations under `deployments/<product>/config/`
3. Remove empty directories and document structure expectations
4. Convert identity postgres to use Docker secrets

### For New Products
1. Follow established directory structure
2. Use shared telemetry compose for observability
3. Implement consistent secret management patterns
4. Include both SQLite and PostgreSQL variants where applicable

### Long-term Architecture
1. Consider Docker Compose profiles for optional components
2. Implement service mesh (Istio/Linkerd) for advanced service discovery
3. Add shared CI/CD workflows for product-agnostic testing
4. Create product template repository for consistent onboarding</content>
<parameter name="filePath">c:\Dev\Projects\cryptoutil\docs\03-products\passthru2\grooming\RESEARCH.md
