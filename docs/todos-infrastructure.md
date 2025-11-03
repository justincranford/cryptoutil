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
w- **Current State**: PARTIALLY COMPLETE - Docker health checks use 127.0.0.1 but application listener only binds to IPv4
- **Action Items**:
  - ✅ Audit entire project for use of "localhost" hostname vs explicit IP addresses (DONE - compose.yml uses 127.0.0.1)
  - ❌ Improve `application_listener.go` to listen on both IPv6 and IPv4 loopback addresses (NEEDS WORK)
  - ✅ Update Docker health checks to use explicit IP addresses instead of localhost (DONE - compose.yml uses 127.0.0.1)
  - ❌ Review container networking configuration for proper IPv4/IPv6 support (NEEDS WORK)
  - ❌ Test health checks and connectivity in various container environments (NEEDS WORK)
  - ❌ Update Docker instructions with IPv6/IPv4 loopback best practices (NEEDS WORK)
- **Files**: `internal/server/application/application_listener.go`, `deployments/compose/compose.yml`, Docker health checks, networking code
- **Expected Outcome**: Reliable networking in containerized environments with proper IPv4/IPv6 loopback support
- **Priority**: Medium - Container networking reliability

### Task INF8: Use HTTPS 127.0.0.1:9090 for Admin APIs
- **Description**: Ensure admin APIs (shutdown, livez, readyz) are accessed via private server HTTPS 127.0.0.1:9090, not public server
- **Current State**: INCOMPLETE - E2E tests check public service endpoints instead of private admin endpoints
- **Action Items**:
  - ❌ Update e2e tests to check readiness on private server URLs (9090) (NEEDS WORK)
  - ✅ Update documentation to show correct admin API endpoints (DONE - README shows https://localhost:9090)
  - ✅ Ensure health checks use private server endpoints (DONE - compose.yml uses 127.0.0.1:9090)
  - ✅ Remove admin API routes from public server if accidentally added (DONE - admin routes only on private server)
- **Files**: `internal/test/e2e/e2e_test.go`, documentation, health check scripts
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
- **Files**: `internal/common/config/config.go`, `internal/server/application/application_listener.go`, `internal/test/e2e/e2e_test.go`, documentation
- **Expected Outcome**: Properly prefixed admin APIs with configurable context path
- **Priority**: Medium - API organization and consistency

### Task INF11: Regular GitHub Actions Version Updates
- **Description**: Keep workflow actions updated (e.g., v4→v5 transitions) to maintain compatibility
- **Current State**: Actions are periodically updated but not systematically tracked
- **Action Items**:
  - Monitor GitHub Actions releases for new versions
  - Update action versions in workflows when new versions are stable
  - Test workflow compatibility after updates
  - Update documentation examples with latest action versions
- **Files**: `.github/workflows/*.yml`, workflow documentation
- **Expected Outcome**: Workflows use latest stable action versions
- **Priority**: LOW - Maintenance and compatibility
- **Timeline**: Ongoing maintenance

---

## 🔵 HIGH - Artifact Consolidation Refactoring

### Task INF10: Consolidate All Temporary Artifacts to `.build/` Directory
- **Description**: Refactor entire project to consolidate all temporary build, test, and scan artifacts under single `.build/` directory for easier management, cleanup, and gitignore maintenance
- **Current State**: Artifacts scattered across 10+ locations (see DEEP-ANALYSIS.md for full inventory)
- **Proposed Structure**:
  ```
  .build/                         # Single consolidated directory for ALL temporary artifacts
  ├── bin/                        # Compiled binaries
  │   ├── cryptoutil              # Main application binary
  │   └── *.test                  # Test binaries (per-package)
  ├── coverage/                   # Test coverage reports
  │   ├── coverage.out            # Coverage profile
  │   ├── coverage.html           # HTML report
  │   └── coverage-{timestamp}.* # Timestamped archives
  ├── dast/                       # DAST security scan outputs
  │   ├── nuclei/                 # Nuclei scan results
  │   │   ├── nuclei.log
  │   │   ├── nuclei.sarif
  │   │   └── nuclei-templates.version
  │   ├── zap/                    # OWASP ZAP scan results
  │   │   ├── full/               # Full scan reports
  │   │   └── api/                # API scan reports
  │   ├── headers/                # Security header captures
  │   ├── app-logs/               # Application logs during DAST
  │   │   ├── cryptoutil.stdout
  │   │   └── cryptoutil.stderr
  │   ├── container-logs/         # Docker container logs
  │   └── diagnostics/            # System diagnostics
  ├── e2e/                        # End-to-end test artifacts
  │   ├── service-logs/           # Combined service logs
  │   ├── container-logs/         # Individual container logs
  │   └── reports/                # E2E test reports
  ├── workflows/                  # GitHub workflow execution logs
  │   ├── {workflow}-{timestamp}.log
  │   ├── {workflow}-analysis-{timestamp}.md
  │   └── combined-{timestamp}.log
  ├── mutation/                   # Mutation testing results
  │   └── mutation-{package}.json
  ├── sarif/                      # SARIF security scan results
  │   ├── trivy-image.sarif
  │   ├── docker-scout-cves.sarif
  │   └── nuclei.sarif (symlink to dast/nuclei/)
  ├── sbom/                       # Software Bill of Materials
  │   └── sbom.spdx.json
  ├── load/                       # Gatling load test results
  │   └── gatling/                # Gatling simulation results
  └── tmp/                        # Ephemeral temporary files
      └── nohup.out
  ```
- **Action Items**:
  1. **Phase 1 - Create Directory Structure & Update .gitignore**:
     - Create `.build/` directory and subdirectories
     - Update `.gitignore` to ignore entire `.build/` directory (single line)
     - Remove existing scattered artifact patterns from `.gitignore`

  2. **Phase 2 - Update Go Code**:
     - Update `cmd/workflow` to use `.build/workflows/`
     - Update E2E test code in `internal/test/e2e/` to output to `.build/e2e/`
     - Update test helpers to write coverage to `.build/coverage/`

  3. **Phase 3 - Update Workflow Files**:
     - Update `ci-dast.yml` to use `.build/dast/` for all DAST artifacts
     - Update `ci-e2e.yml` to use `.build/e2e/` for E2E artifacts
     - Update `ci-quality.yml` to use `.build/{coverage,sarif,sbom}/`

  4. **Phase 4 - Update Scripts**:
     - Update `cmd/workflow` to use `.build/workflow/`

  5. **Phase 5 - Update Build Configuration**:
     - Update Makefile (if exists) to use `.build/bin/`
     - Update `go build` commands in documentation to output to `.build/bin/`
     - Update Gatling POM to output to `.build/load/` instead of `test/load/target/`

  6. **Phase 6 - Cleanup & Documentation**:
     - Add cleanup script: `scripts/clean-build.{ps1,sh}` to remove `.build/`
     - Update README.md with new artifact locations
     - Update DEEP-ANALYSIS.md with refactored structure
     - Remove old scattered artifact directories from repository
- **Files Modified**:
  - `.gitignore` (simplified to single `.build/` exclusion)
  - `cmd/workflow`
  - `internal/test/e2e/*.go` (E2E test utilities)
  - `.github/workflows/*.yml` (all 5 workflows)
  - `test/load/pom.xml` (Gatling output directory)
  - `README.md` and `docs/DEEP-ANALYSIS.md`
- **Expected Outcome**:
  - Single directory for all temporary artifacts
  - Simplified `.gitignore` (one line instead of 15+)
  - Easy cleanup: `rm -rf .build/` or `scripts/clean-build.sh`
  - Better artifact discovery and organization
  - Consistent artifact paths across all tools
- **Benefits**:
  - **Developer Experience**: Easier to find all artifacts in one place
  - **Cleanup**: Single command removes all temporary files
  - **Gitignore Maintenance**: One pattern vs 15+ scattered patterns
  - **CI/CD**: Consistent artifact paths across workflows
  - **Documentation**: Clear, predictable artifact locations
- **Priority**: High - Developer productivity and project organization
- **Estimated Effort**: 4-6 hours across 6 phases
- **Dependencies**: None (can be done incrementally per phase)
