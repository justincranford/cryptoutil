# Cryptoutil Infrastructure & Deployment TODOs

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
- **Priority**: High - Configuration security and reliability</content>
<parameter name="filePath">c:\Dev\Projects\cryptoutil\docs\todos-infrastructure.md
