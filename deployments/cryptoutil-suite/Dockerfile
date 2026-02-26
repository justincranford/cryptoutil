#############################################################################################
# Cryptoutil Suite Dockerfile
# Use BuildKit syntax to enable cache mounts (fast rebuilds)
# To build with cache mounts set DOCKER_BUILDKIT=1
# Example: DOCKER_BUILDKIT=1 docker build -t cryptoutil -f deployments/cryptoutil-suite/Dockerfile .
#
#############################################################################################
ARG APP_VERSION=UNSET
ARG VCS_REF=UNSET
ARG BUILD_DATE=UNSET
#############################################################################################
ARG GO_VERSION=1.25.5
ARG ALPINE_VERSION=3.19
ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
# Build flags for production container: static linking with debug symbols retained.
ARG LDFLAGS="-s -extldflags '-static'"
#############################################################################################

# Validate required build arguments.
#############################################################################################
FROM alpine:${ALPINE_VERSION} AS validation
ARG APP_VERSION=UNSET
ARG VCS_REF=UNSET
ARG BUILD_DATE=UNSET

# Check all mandatory parameters and collect errors.
RUN set -e; \
    errors=""; \
    if [ "$APP_VERSION" = "UNSET" ]; then \
        errors="${errors}ERROR: APP_VERSION build argument is required\n"; \
    fi; \
    if [ "$VCS_REF" = "UNSET" ]; then \
        errors="${errors}ERROR: VCS_REF build argument is required\n"; \
    fi; \
    if [ "$BUILD_DATE" = "UNSET" ]; then \
        errors="${errors}ERROR: BUILD_DATE build argument is required\n"; \
    fi; \
    if [ -n "$errors" ]; then \
        printf "%b" "$errors" >&2; \
        echo "Usage: DOCKER_BUILDKIT=1 docker build --build-arg APP_VERSION=<version> --build-arg VCS_REF=\$(git rev-parse HEAD) --build-arg BUILD_DATE=\$(date -u +\"%Y-%m-%dT%H:%M:%SZ\") -t cryptoutil -f deployments/cryptoutil-suite/Dockerfile ." >&2; \
        exit 1; \
    fi

# Write build parameters to file for runtime inspection.
RUN mkdir -p /app && \
    echo "APP_VERSION=${APP_VERSION}" > /app/.build-params && \
    echo "VCS_REF=${VCS_REF}" >> /app/.build-params && \
    echo "BUILD_DATE=${BUILD_DATE}" >> /app/.build-params && \
    cat /app/.build-params

#############################################################################################
FROM golang:${GO_VERSION} AS builder
WORKDIR /src

# Redeclare build args for use in this stage.
ARG APP_VERSION
ARG VCS_REF
ARG BUILD_DATE
ARG GO_VERSION
ARG CGO_ENABLED
ARG GOOS
ARG GOARCH
ARG LDFLAGS

# Copy dependency manifests first to leverage layer caching.
COPY go.mod go.sum ./

# Download modules using BuildKit cache mounts when available.
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	go mod download

# Copy the remainder of the source.
COPY . .

# Build the static, trimmed binary.
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} \
	go build -a -tags netgo -trimpath -ldflags="${LDFLAGS}" -o /app/cryptoutil ./cmd/cryptoutil

# Validate that the binary is statically linked.
SHELL ["/bin/bash", "-c"]
RUN if ldd /app/cryptoutil 2>/dev/null; then \
        echo "✗ Binary is dynamically linked - failing build"; \
        ldd /app/cryptoutil; \
        exit 1; \
    else \
        echo "✓ Binary is statically linked"; \
    fi

# Create /app directory with proper ownership for the application user.
RUN mkdir -p /app && chmod 555 /app && chown -R 65532:65532 /app

# Create runtime directory for application data with proper ownership.
RUN mkdir -p /app/run && chmod 775 /app/run && chown -R 65532:65532 /app/run

#############################################################################################
FROM alpine:${ALPINE_VERSION} AS runtime

# Redeclare build args for use in labels.
ARG APP_VERSION
ARG VCS_REF
ARG BUILD_DATE

# OCI Image spec labels.
LABEL org.opencontainers.image.title="CA Server" \
      org.opencontainers.image.description="Certificate Authority REST API Server" \
      org.opencontainers.image.version="${APP_VERSION}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.source="https://github.com/justincranford/cryptoutil" \
      org.opencontainers.image.licenses="MIT"

# Install ca-certificates for TLS and busybox for wget (healthchecks).
# Create non-root user for security.
# hadolint ignore=DL3018
RUN apk --no-cache add ca-certificates && \
    adduser -D -H -u 65532 -s /sbin/nologin ca

# Copy build artifacts.
COPY --from=validation /app/.build-params /app/.build-params
COPY --from=builder --chown=65532:65532 /app/cryptoutil /app/cryptoutil
COPY --from=builder --chown=65532:65532 /app/run /app/run

# Create directories for config, data, and secrets.
RUN mkdir -p /etc/ca /app/data /run/secrets && \
    chown -R 65532:65532 /etc/ca /app/data /run/secrets

# Set working directory and default user.
WORKDIR /app
USER ca

# Expose ports.
EXPOSE 9093 9093

# Healthcheck using wget (available in Alpine).
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget --no-check-certificate -q -O /dev/null https://127.0.0.1:9090/admin/api/v1/livez || exit 1

# Default entrypoint.
ENTRYPOINT ["/app/cryptoutil"]
CMD ["server", "--config=/etc/cryptoutil/ca.yml"]
