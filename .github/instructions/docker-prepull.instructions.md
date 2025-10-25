---
description: "Instructions for optimizing Docker image pulls in CI/CD workflows"
applyTo: ".github/workflows/*.yml"
---
# Docker Image Pre-Pull Optimization

## Container Images Used Across Workflows

### Direct Docker Images (docker run/services)
- `postgres:18` - PostgreSQL 18 specific version (dast.yml, e2e.yml, e2e.yml services)
- `alpine:latest` - Health check utilities (compose.yml)
- `otel/opentelemetry-collector-contrib:latest` - OpenTelemetry Collector (compose.yml)
- `grafana/otel-lgtm:latest` - Grafana OTEL LGTM stack (compose.yml)

### Docker Actions (run in containers)
- `ghcr.io/zaproxy/zaproxy:stable` - OWASP ZAP DAST scanner (dast.yml)
- `golang:1.25.1` - Go builder image (Dockerfile multi-stage)
- `alpine:3.19` - Alpine runtime base (Dockerfile multi-stage)

### Docker Compose Builds
- Custom `cryptoutil` images built from local Dockerfile (e2e.yml, compose.yml)

## Pre-Pull Optimization Pattern

**CRITICAL**: When workflows use `docker` or `docker compose`, add a dedicated pre-pull step that runs ALL image pulls concurrently before they are needed.

### Benefits
- **Parallel downloads**: All images download simultaneously instead of sequentially
- **Faster workflows**: Reduces total pull time by 50-80%
- **Better diagnostics**: Clear separation of pull failures vs runtime failures
- **Cached layers**: Docker BuildKit cache benefits from having base images present

### Implementation Pattern

```yaml
- name: Pre-pull Docker images (parallel)
  run: |
    echo "🐳 Pre-pulling all Docker images in parallel..."

    # Array of all images used in this workflow
    IMAGES=(
      "postgres:18"
      "ghcr.io/zaproxy/zaproxy:stable"
      "alpine:latest"
    )

    # Pull all images concurrently
    for image in "${IMAGES[@]}"; do
      echo "Pulling $image..."
      docker pull "$image" &
    done

    # Wait for all pulls to complete
    wait

    echo "✅ All images pre-pulled successfully"
```

### Workflow-Specific Image Lists

**DAST Workflow (dast.yml)**:
```yaml
IMAGES=(
  "postgres:18"
  "ghcr.io/zaproxy/zaproxy:stable"
)
```

**E2E Workflow (e2e.yml)**:
```yaml
IMAGES=(
  "postgres:18"
  "alpine:3.19"         # For buildx/compose base
  "golang:1.25.1"       # For custom image builds
)
```

**Quality Workflow (quality.yml)**:
```yaml
IMAGES=(
  "alpine:3.19"
  "golang:1.25.1"
)
```

### Docker Compose Pre-Pull Pattern

For workflows using `docker compose`, pull compose-managed images separately:

```yaml
- name: Pre-pull Docker Compose images (parallel)
  run: |
    echo "🐳 Pre-pulling Docker Compose images..."

    # Pull external images concurrently
    docker pull postgres:18 &
    docker pull otel/opentelemetry-collector-contrib:latest &
    docker pull grafana/otel-lgtm:latest &
    docker pull alpine:latest &

    wait

    echo "✅ Compose images pre-pulled"

- name: Build Docker images (after pre-pull)
  run: |
    echo "🏗️ Building Docker images for end-to-end testing..."
    docker compose -f ${{ env.COMPOSE_FILE }} build
```

### Placement Guidelines

1. **Add pre-pull step BEFORE first Docker usage** in workflow
2. **After checkout/setup steps** to have Dockerfile available for inspection
3. **Before build/compose steps** to ensure base images are cached
4. **Use descriptive step names** indicating parallel pulling

### Error Handling

```yaml
- name: Pre-pull Docker images with error handling
  run: |
    echo "🐳 Pre-pulling Docker images in parallel..."

    FAILED_IMAGES=()

    for image in "${IMAGES[@]}"; do
      (docker pull "$image" || echo "$image" >> /tmp/failed_images.txt) &
    done

    wait

    if [ -f /tmp/failed_images.txt ]; then
      echo "❌ Failed to pull images:"
      cat /tmp/failed_images.txt
      exit 1
    fi

    echo "✅ All images pre-pulled successfully"
```

## When NOT to Pre-Pull

Skip pre-pull optimization when:
- Workflow uses only GitHub Actions (no direct Docker commands)
- Single image used once (minimal benefit)
- Images already cached by previous workflow steps
- Workflow timeout budget is very tight (pre-pull adds upfront cost)

## Maintenance

When adding new Docker images to workflows:
1. **Identify all image references** (services, docker run, compose files)
2. **Update pre-pull step** to include new images
3. **Test workflow** to verify parallel pulling works
4. **Update this instruction file** with new image in the lists above
