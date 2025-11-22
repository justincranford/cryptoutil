# Cryptoutil Infrastructure & Deployment TODOs

**IMPORTANT**: Delete completed tasks immediately after completion to maintain a clean, actionable TODO list.

**Last Updated**: November 2, 2025
**Status**: Release automation, Kubernetes deployment planning, configuration priority review, and workflow-level summary actions underway

---

## 🟡 MEDIUM - Infrastructure & Deployment

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

### Task INF6: IPv6 vs IPv4 Loopback Networking Investigation
- **Description**: Investigate and resolve IPv6/IPv4 loopback address inconsistencies in containerized deployments
- **Current State**: PARTIALLY COMPLETE - Docker health checks use 127.0.0.1 but application listener only binds to IPv4
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

---

## 🔵 HIGH - Artifact Consolidation Refactoring

### Task INF10: Consolidate All Temporary Artifacts to `.build/` Directory

- **Description**: Refactor entire project to consolidate all temporary build, test, and scan artifacts under single `.build/` directory for easier management, cleanup, and gitignore maintenance
- **Current State**: Artifacts scattered across 10+ locations (see docs/README.md for architecture details)
- **Proposed Structure**:

  ```text
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
     - Update docs/README.md with refactored structure
     - Remove old scattered artifact directories from repository
- **Files Modified**:
  - `.gitignore` (simplified to single `.build/` exclusion)
  - `cmd/workflow`
  - `internal/test/e2e/*.go` (E2E test utilities)
  - `.github/workflows/*.yml` (all 5 workflows)
  - `test/load/pom.xml` (Gatling output directory)
  - `README.md` and `docs/README.md`
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
