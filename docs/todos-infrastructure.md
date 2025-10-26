# Cryptoutil Infrastructure & Deployment TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: October 15, 2025
**Status**: Release automation, Kubernetes deployment planning, and configuration priority review underway

---

## 🟡 MEDIUM - Infrastructure & Deployment

### Task INF1: Automated Release Pipeline
- **Description**: Implement automated release pipeline with semantic versioning
- **Current State**: Manual releases only
- **Action Items**:
  - Create `.github/workflows/release.yml` with automated changelog generation
  - Implement semantic versioning and automated releases
  - Set up container registry publishing
  - Configure multi-environment deployment strategy (dev → staging → production)
- **Files**: `.github/workflows/release.yml`, release scripts
- **Expected Outcome**: Automated, reliable release process
- **Priority**: High - Production deployment

### Task INF2: Kubernetes Deployment Manifests
- **Description**: Create production-ready Kubernetes deployment configurations
- **Current State**: Docker Compose only
- **Action Items**:
  - Create Kubernetes deployment, service, and ingress manifests
  - Implement ConfigMaps and Secrets management
  - Set up health checks and readiness probes
  - Configure resource limits and requests
- **Files**: `deployments/kubernetes/` directory with YAML manifests
- **Expected Outcome**: Production Kubernetes deployment capability
- **Priority**: Medium - Production infrastructure

### Task INF3: Helm Charts for Flexible Deployment
- **Description**: Create Helm charts for flexible, templated deployments
- **Current State**: No Helm support
- **Action Items**:
  - Create Helm chart with configurable values
  - Implement chart templating for different environments
  - Add chart testing and validation
  - Document Helm deployment procedures
- **Files**: `deployments/helm/cryptoutil/` directory
- **Expected Outcome**: Flexible deployment across environments
- **Priority**: Medium - Production infrastructure

### Task INF5: Configuration Priority Order Review
- **Description**: Review whole project to ensure correct configuration priority order across all deployment scenarios
- **Current State**: Configuration sources may not follow proper precedence
- **Required Priority Order**:
  1. **Docker/Kubernetes secrets** (credentials and sensitive settings)
  2. **Configuration YAML files** (non-sensitive settings); may be single config or split based on different groupings
  3. **Command parameters**, as first fallback to override 1 or 2
  4. **Environment variables**, as last fallback to override 1, 2, or 3
- **Action Items**:
  - Audit all configuration loading code for proper precedence
  - Confirm Viper configuration conforms to this priority order
  - Update configuration loading logic if needed
  - Document configuration precedence in architecture instructions
  - Test configuration override behavior across all deployment methods
- **Files**: `internal/common/config/`, Viper setup code, deployment configs
- **Expected Outcome**: Consistent, secure configuration management across all environments
- **Priority**: High - Configuration security and reliability

### Task INF6: IPv6 vs IPv4 Loopback Networking Investigation
- **Description**: Investigate and resolve IPv6/IPv4 loopback address inconsistencies in containerized deployments
- **Current State**: Health checks failing due to localhost resolving to IPv6 (::1) while servers only listen on IPv4 (127.0.0.1)
- **Action Items**:
  - Audit entire project for use of "localhost" hostname vs explicit IP addresses
  - Improve `application_listener.go` to listen on both IPv6 and IPv4 loopback addresses
  - Update Docker health checks to use explicit IP addresses instead of localhost
  - Review container networking configuration for proper IPv4/IPv6 support
  - Test health checks and connectivity in various container environments
  - Update Docker instructions with IPv6/IPv4 loopback best practices
- **Files**: `internal/server/application/application_listener.go`, `deployments/compose/compose.yml`, Docker health checks, networking code
- **Expected Outcome**: Reliable networking in containerized environments with proper IPv4/IPv6 loopback support
- **Priority**: Medium - Container networking reliability

### Task INF7: Revert Docker Compose to Use Mounted Secrets
- **Description**: Revert `deployments/compose/compose.yml` postgres service to use mounted Docker secrets instead of hard-coded credential values
- **Current State**: Postgres service uses hard-coded environment variables (USR, PWD, DB)
- **Required Changes**:
  - Change `POSTGRES_USER: USR` to `POSTGRES_USER_FILE: /run/secrets/postgres_username.secret`
  - Change `POSTGRES_PASSWORD: PWD` to `POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password.secret`
  - Change `POSTGRES_DB: DB` to `POSTGRES_DB_FILE: /run/secrets/postgres_database.secret`
- **Action Items**:
  - Update postgres service environment variables in `deployments/compose/compose.yml`
  - Ensure secrets are properly mounted (already configured)
  - Test database connectivity after changes
  - Verify health checks still work with secret-based configuration
- **Files**: `deployments/compose/compose.yml`
- **Expected Outcome**: Secure credential management using Docker secrets instead of hard-coded values
- **Priority**: High - Security and deployment best practices
- **Examples of Changes**:
  ```yaml
  # BEFORE (hard-coded values):
  postgres:
    environment:
      POSTGRES_USER: USR
      POSTGRES_PASSWORD: PWD
      POSTGRES_DB: DB

  # AFTER (mounted secrets):
  postgres:
    environment:
      POSTGRES_USER_FILE: /run/secrets/postgres_username.secret
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password.secret
      POSTGRES_DB_FILE: /run/secrets/postgres_database.secret
  ```

### Task INF8: Use HTTPS 127.0.0.1:9090 for Admin APIs
- **Description**: Ensure admin APIs (shutdown, livez, readyz) are accessed via private server HTTPS 127.0.0.1:9090, not public server
- **Current State**: Admin APIs incorrectly accessed on public ports (8080)
- **Action Items**:
  - Update e2e tests to check readiness on private server URLs (9090)
  - Update documentation to show correct admin API endpoints
  - Ensure health checks use private server endpoints
  - Remove admin API routes from public server if accidentally added
- **Files**: `internal/e2e/e2e_test.go`, documentation, health check scripts
- **Expected Outcome**: Admin APIs properly isolated to private server
- **Priority**: High - API security and correct architecture

### Task INF9: Add /admin/v1 Prefix to Private Admin APIs
- **Description**: Add configurable /admin/v1 prefix to private admin APIs (shutdown, livez, readyz) on HTTPS 127.0.0.1:9090
- **Current State**: Admin APIs use root paths (/shutdown, /livez, /readyz)
- **Action Items**:
  - Add `privateAdminAPIContextPath` setting to config.go with default "/admin/v1"
  - Update application_listener.go to use prefixed paths for admin endpoints
  - Update health check functions to use prefixed endpoints
  - Update e2e tests to use prefixed admin API endpoints
  - Update documentation with new admin API paths
- **Files**: `internal/common/config/config.go`, `internal/server/application/application_listener.go`, `internal/e2e/e2e_test.go`, documentation
- **Expected Outcome**: Properly prefixed admin APIs with configurable context path
- **Priority**: Medium - API organization and consistency
