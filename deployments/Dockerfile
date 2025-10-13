#############################################################################################
# Use BuildKit syntax to enable cache mounts (fast rebuilds)
# To build with cache mounts set DOCKER_BUILDKIT=1
# Example: DOCKER_BUILDKIT=1 docker build -t cryptoutil -f deployments/Dockerfile .
#
#############################################################################################
ARG APP_VERSION=UNSET
ARG VCS_REF=UNSET
ARG BUILD_DATE=UNSET
#############################################################################################
ARG GO_VERSION=1.25.1
ARG ALPINE_VERSION=3.19
ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
ARG LDFLAGS="-s -w"
#############################################################################################

# Validate required build arguments
#############################################################################################
FROM alpine:${ALPINE_VERSION} AS validation
ARG APP_VERSION=UNSET
ARG VCS_REF=UNSET
ARG BUILD_DATE=UNSET

# Check all mandatory parameters and collect errors
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
        echo "Usage: DOCKER_BUILDKIT=1 docker build --build-arg APP_VERSION=<version> --build-arg VCS_REF=\$(git rev-parse HEAD) --build-arg BUILD_DATE=\$(date -u +\"%Y-%m-%dT%H:%M:%SZ\") -t cryptoutil -f deployments/Dockerfile ." >&2; \
        exit 1; \
    fi

# Write build parameters to file for runtime inspection
RUN echo "APP_VERSION=${APP_VERSION}" > /build-params && \
    echo "VCS_REF=${VCS_REF}" >> /build-params && \
    echo "BUILD_DATE=${BUILD_DATE}" >> /build-params && \
    cat /build-params

#############################################################################################
FROM golang:${GO_VERSION} AS builder
WORKDIR /src

# Redeclare build args for use in this stage
ARG APP_VERSION
ARG VCS_REF
ARG BUILD_DATE

# Copy dependency manifests first to leverage layer caching
COPY go.mod go.sum ./

# Download modules using BuildKit cache mounts when available
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	go mod download

# Copy the remainder of the source
COPY . .

# Build the static, trimmed binary. Uses BuildKit cache mounts for faster incremental builds.
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} \
	go build -trimpath -ldflags="${LDFLAGS}" -o /app/cryptoutil ./cmd/cryptoutil

# Create /app directory with proper ownership for the application user
RUN mkdir -p /app && chown -R 65532:65532 /app

# Create runtime directory for application data (certificates, etc.) with proper ownership
RUN mkdir -p /app/run && chown -R 65532:65532 /app/run

# Generate build metadata files for runtime inspection
RUN git rev-parse HEAD > /app/.vcs-ref && \
    date -u +"%Y-%m-%dT%H:%M:%SZ" > /app/.build-date

#############################################################################################
FROM alpine:${ALPINE_VERSION} AS runtime-deps
# hadolint ignore=DL3018 # Intentionally unpinned for automatic security updates
RUN apk --no-cache add ca-certificates tzdata tini
WORKDIR /root/ssl
RUN update-ca-certificates

#############################################################################################
# If shell needed for debugging final image, use alpine instead of scratch
FROM alpine:${ALPINE_VERSION}
#FROM scratch

# Copy validation marker to ensure required args were provided
COPY --from=validation /build-params /app/.build-params

# Copy build metadata files for runtime inspection
COPY --from=builder /app/.vcs-ref /app/.vcs-ref
COPY --from=builder /app/.build-date /app/.build-date

# Copy runtime dependencies (tini for process management)
COPY --from=runtime-deps /sbin/tini /sbin/tini

# Copy the application binary
COPY --from=builder /app/cryptoutil /app/cryptoutil

# Re-declare build args in this stage so they are available for labels
ARG APP_VERSION
ARG VCS_REF
ARG BUILD_DATE

# Image metadata labels (set at final stage so they are present in the produced image)
LABEL org.opencontainers.image.title="cryptoutil"
LABEL org.opencontainers.image.description="A small utility for cryptographic key and certificate operations"
LABEL org.opencontainers.image.source="https://github.com/justincranford/cryptoutil"
LABEL org.opencontainers.image.authors="Justin Cranford <justincranford@example.com>"
LABEL org.opencontainers.image.version="${APP_VERSION}"
LABEL org.opencontainers.image.revision="${VCS_REF}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"

# Use non-privileged ports
EXPOSE 8080 9090

# Healthcheck available with Alpine base image
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider --no-check-certificate https://localhost:9090/readyz || exit 1

# For Kubernetes: Use liveness/readiness probes like:
#   livenessProbe:
#     httpGet:
#       path: /readyz
#       port: 9090
# For Docker Compose: Use external healthcheck or depends_on with condition

# Use tini for proper signal handling and zombie process reaping
ENTRYPOINT ["/sbin/tini", "--", "/app/cryptoutil"]
