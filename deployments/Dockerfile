#############################################################################################
# Use BuildKit syntax to enable cache mounts (fast rebuilds)
# To build with cache mounts set DOCKER_BUILDKIT=1
# Example: DOCKER_BUILDKIT=1 docker build -t cryptoutil -f deployments/Dockerfile .
#
ARG ALPINE_VERSION=3.19
# Single pinned Go builder stage (preferred over duplicating builder stages)
ARG GO_VERSION=1.25.1
ARG APP_VERSION=dev
ARG VCS_REF=unspecified
ARG BUILD_DATE=1970-01-01T00:00:00Z
FROM golang:${GO_VERSION} AS builder
WORKDIR /src

ARG APP_VERSION=dev
ARG VCS_REF=unspecified
ARG BUILD_DATE=1970-01-01T00:00:00Z

LABEL org.opencontainers.image.source="https://github.com/justincranford/cryptoutil"
LABEL org.opencontainers.image.version="${APP_VERSION}"
LABEL org.opencontainers.image.revision="${VCS_REF}"
LABEL org.opencontainers.image.title="cryptoutil"
LABEL org.opencontainers.image.description="A small utility for cryptographic key and certificate operations"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.authors="Justin Cranford <justin@example.com>"

# Copy dependency manifests first to leverage layer caching
COPY go.mod go.sum ./

# Download modules using BuildKit cache mounts when available
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	go mod download

# Copy the remainder of the source
COPY . .

# Build arguments for portability and reproducible builds
ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
ARG LDFLAGS="-s -w"

# Build the static, trimmed binary. Uses BuildKit cache mounts for faster incremental builds.
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} \
	go build -trimpath -ldflags="${LDFLAGS}" -o /cryptoutil ./cmd/cryptoutil

#############################################################################################

ARG ALPINE_VERSION=3.19
FROM alpine:${ALPINE_VERSION} AS runtime-deps
# Intentionally unpinned for automatic security updates
# hadolint ignore=DL3018
RUN apk --no-cache add ca-certificates tzdata tini
WORKDIR /root/ssl
RUN update-ca-certificates

FROM scratch

# Copy tini for proper signal handling and PID 1 management
COPY --from=runtime-deps /sbin/tini /tini

# Copy CA certificates for HTTPS requests
COPY --from=runtime-deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data for proper time handling
COPY --from=runtime-deps /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the static binary produced by the single builder stage
COPY --from=builder /cryptoutil /cryptoutil

# Run as non-root user (65532:65532 industry standard)
# Note: scratch has no passwd file, so numeric IDs only
USER 65532:65532

# Re-declare build args in this stage so they are available for labels
ARG APP_VERSION=dev
ARG VCS_REF=unspecified

# Image metadata labels (set at final stage so they are present in the produced image)
LABEL org.opencontainers.image.version="${APP_VERSION}"
LABEL org.opencontainers.image.revision="${VCS_REF}"

# Use non-privileged ports
EXPOSE 8080 9090

# Note: No healthcheck possible with scratch (no shell/curl available)
# For Kubernetes: Use liveness/readiness probes like:
#   livenessProbe:
#     httpGet:
#       path: /readyz
#       port: 9090
# For Docker Compose: Use external healthcheck or depends_on with condition

# Use tini for proper signal handling and zombie process reaping
ENTRYPOINT ["/tini", "--", "/cryptoutil"]

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
