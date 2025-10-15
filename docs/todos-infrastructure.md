# Cryptoutil Infrastructure & Deployment TODOs

**Last Updated**: October 14, 2025
**Status**: Release automation and Kubernetes deployment planning underway

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

### Task INF4: Pin Docker Image Versions
- **Description**: Pin all Docker image versions in compose configuration for reproducible builds
- **Current State**: Some images may use latest tags
- **Action Items**:
  - Audit all Docker images in compose files
  - Pin versions to specific tags (avoid :latest)
  - Set up automated dependency updates for security patches
- **Files**: `deployments/compose/*.yml`
- **Expected Outcome**: Reproducible and secure container deployments
- **Priority**: Medium - Infrastructure stability</content>
<parameter name="filePath">c:\Dev\Projects\cryptoutil\docs\todos-infrastructure.md
