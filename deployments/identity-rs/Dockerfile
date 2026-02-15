#############################################################################################
# Identity RS (Resource Server) Dockerfile
# Use BuildKit syntax to enable cache mounts (fast rebuilds)
# To build with cache mounts set DOCKER_BUILDKIT=1
# Example: DOCKER_BUILDKIT=1 docker build -t cryptoutil-identity-rs -f deployments/identity/Dockerfile.rs .
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

# Validate required build arguments
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
        echo "Usage: DOCKER_BUILDKIT=1 docker build --build-arg APP_VERSION=<version> --build-arg VCS_REF=\$(git rev-parse HEAD) --build-arg BUILD_DATE=\$(date -u +\"%Y-%m-%dT%H:%M:%SZ\") -t cryptoutil-identity-rs -f deployments/identity/Dockerfile.rs ." >&2; \
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

# Build the static, trimmed binary. Uses BuildKit cache mounts for faster incremental builds.
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

# Create runtime directory for application data (certificates, etc.) with proper ownership.
RUN mkdir -p /app/run && chmod 775 /app/run && chown -R 65532:65532 /app/run

# Create cache directory for any caching needs with proper ownership.
RUN mkdir -p /app/run/.cache && chmod 777 /app/run/.cache && chown -R 65532:65532 /app/run/.cache

# Generate build metadata files for runtime inspection.
RUN git rev-parse HEAD > /app/.vcs-ref && \
    date -u +"%Y-%m-%dT%H:%M:%SZ" > /app/.build-date

#############################################################################################
FROM alpine:${ALPINE_VERSION} AS runtime-deps
WORKDIR /root/ssl
# hadolint ignore=DL3018 # Intentionally unpinned for automatic security updates.
RUN --mount=type=cache,target=/var/cache/apk \
    apk --no-cache add ca-certificates tzdata tini && \
    update-ca-certificates

#############################################################################################
# If shell needed for debugging final image, use alpine instead of scratch.
FROM alpine:${ALPINE_VERSION} AS final

# Copy apk installed dependencies from runtime-deps.
COPY --from=runtime-deps /sbin/tini /sbin/tini
COPY --from=runtime-deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy builder artifacts.
COPY --from=builder /app /app

# Copy validation artifacts.
COPY --from=validation /app/.build-params /app/.build-params

# Set working directory for the application (generated certs are written to working dir).
WORKDIR /app/run

# Set Go cache environment variables (Empty directory is copied from builder).
ENV GOMODCACHE=/app/run/.cache
ENV GOCACHE=/app/run/.cache

# Re-declare build args in this stage so they are available for labels.
ARG APP_VERSION
ARG VCS_REF
ARG BUILD_DATE

# Image metadata labels (set at final stage so they are present in the produced image).
LABEL org.opencontainers.image.title="cryptoutil-identity-rs"
LABEL org.opencontainers.image.description="AGPL-3.0 Identity Resource Server"
LABEL org.opencontainers.image.source="https://github.com/justincranford/cryptoutil"
LABEL org.opencontainers.image.authors="Justin Cranford <justincranford@example.com>"
LABEL org.opencontainers.image.version="${APP_VERSION}"
LABEL org.opencontainers.image.revision="${VCS_REF}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"

# Use non-privileged ports.
EXPOSE 18200

# Healthcheck available with Alpine base image.
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
  CMD wget --no-check-certificate -q -O /dev/null https://127.0.0.1:9090/admin/api/v1/livez || exit 1

# Use tini for proper signal handling and zombie process reaping.
ENTRYPOINT ["/sbin/tini", "--", "/app/cryptoutil", "identity-rs", "start"]

# Switch to non-root user for security.
# USER 65532:65532
