#############################################################################################
# Use BuildKit syntax to enable cache mounts (fast rebuilds)
# To build with cache mounts set DOCKER_BUILDKIT=1
# Example: DOCKER_BUILDKIT=1 docker build -t cryptoutil -f deployments/Dockerfile .
#
# Single pinned Go builder stage (preferred over duplicating builder stages)
ARG GO_VERSION=1.25.1
ARG ALPINE_VERSION=3.19
ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
ARG LDFLAGS="-s -w"
ARG APP_VERSION=dev
ARG VCS_REF=UNSET
ARG BUILD_DATE=UNSET

#############################################################################################
FROM alpine:${ALPINE_VERSION} AS validation
# Validate required build arguments
ARG VCS_REF=UNSET
ARG BUILD_DATE=UNSET
RUN if [ "$VCS_REF" = "UNSET" ]; then \
        echo "ERROR: VCS_REF build argument is required" >&2 && \
        echo "Usage: docker build --build-arg VCS_REF=\$(git rev-parse HEAD) ..." >&2 && \
        exit 1; \
    fi
RUN if [ "$BUILD_DATE" = "UNSET" ]; then \
        echo "ERROR: BUILD_DATE build argument is required" >&2 && \
        echo "Usage: docker build --build-arg BUILD_DATE=\$(date -u +\"%%Y-%%m-%%dT%%H:%%M:%%SZ\") ..." >&2 && \
        exit 1; \
    fi
# Create validation marker
RUN echo "Build arguments validated successfully" > /validation-marker

#############################################################################################
FROM golang:${GO_VERSION} AS builder
WORKDIR /src

# Redeclare build args for use in this stage
ARG VCS_REF
ARG BUILD_DATE
ARG APP_VERSION

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
    date -u +"%Y-%m-%dT%H:%M:%SZ" > /app/.build-date && \
    echo "${APP_VERSION}" > /app/.app-version

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
COPY --from=validation /validation-marker /validation-marker

# Copy build metadata files for runtime inspection
COPY --from=builder /app/.vcs-ref /app/.vcs-ref
COPY --from=builder /app/.build-date /app/.build-date
COPY --from=builder /app/.app-version /app/.app-version

# Copy runtime dependencies (tini for process management)
COPY --from=runtime-deps /sbin/tini /app/tini

# Copy the application binary
COPY --from=builder /app/cryptoutil /app/cryptoutil

# Re-declare build args in this stage so they are available for labels
ARG APP_VERSION=dev
ARG VCS_REF
ARG BUILD_DATE

# Image metadata labels (set at final stage so they are present in the produced image)
LABEL org.opencontainers.image.source="https://github.com/justincranford/cryptoutil"
LABEL org.opencontainers.image.version="${APP_VERSION}"
LABEL org.opencontainers.image.revision="${VCS_REF}"
LABEL org.opencontainers.image.title="cryptoutil"
LABEL org.opencontainers.image.description="A small utility for cryptographic key and certificate operations"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.authors="Justin Cranford <justin@example.com>"

# Use non-privileged ports
EXPOSE 8080 9090

# Healthcheck available with Alpine base image
# HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
#   CMD wget --no-verbose --tries=1 --spider http://localhost:9090/readyz || exit 1

# For Kubernetes: Use liveness/readiness probes like:
#   livenessProbe:
#     httpGet:
#       path: /readyz
#       port: 9090
# For Docker Compose: Use external healthcheck or depends_on with condition

# Use tini for proper signal handling and zombie process reaping
ENTRYPOINT ["/app/tini", "--", "/app/cryptoutil"]

#
# Example CI build with buildx and cache export/import (GitHub Actions / build servers)
#
# docker buildx build \
#   --builder mybuilder \
#   --cache-to=type=registry,ref=ghcr.io/ORG/cryptoutil-cache:buildcache,mode=max \
#   --cache-from=type=registry,ref=ghcr.io/ORG/cryptoutil-cache:buildcache \
#   --build-arg APP_VERSION=${{ github.ref_name }} \
#   --build-arg VCS_REF=${{ github.sha }} \
#   --build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
#   --tag ghcr.io/ORG/cryptoutil:${{ github.ref_name }} \
#   --push \
#   -f deployments/Dockerfile .
