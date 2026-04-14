#############################################################################################
# sm-kms Dockerfile — Key Management
#
# Build: DOCKER_BUILDKIT=1 docker build \
#          --build-arg APP_VERSION=<ver> \
#          --build-arg VCS_REF=$(git rev-parse HEAD) \
#          --build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
#          -t cryptoutil-sm-kms:<tag> \
#          -f deployments/sm-kms/Dockerfile .
#############################################################################################
ARG APP_VERSION=UNSET
ARG VCS_REF=UNSET
ARG BUILD_DATE=UNSET
#############################################################################################
ARG GO_VERSION=1.26.1
# hadolint ignore=DL3007
ARG ALPINE_VERSION=latest
# Design intent is to NEVER set CGO_ENABLED=1
ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64
ARG LDFLAGS="-s -extldflags '-static'"
#############################################################################################
# Non-root container user/group IDs. Default 65532:65532 is the well-known nonroot UID/GID
# convention. Running as non-root limits blast radius of container escapes.
ARG CONTAINER_UID=65532
ARG CONTAINER_GID=65532
#############################################################################################

#############################################################################################
# Stage 1: Validate build arguments
#############################################################################################
FROM alpine:${ALPINE_VERSION} AS validation
ARG APP_VERSION=UNSET
ARG VCS_REF=UNSET
ARG BUILD_DATE=UNSET

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
        exit 1; \
    fi

RUN mkdir -p /app && \
    echo "APP_VERSION=${APP_VERSION}" > /app/.build-params && \
    echo "VCS_REF=${VCS_REF}" >> /app/.build-params && \
    echo "BUILD_DATE=${BUILD_DATE}" >> /app/.build-params

#############################################################################################
# Stage 2: Build Go binary
#############################################################################################
FROM golang:${GO_VERSION} AS builder
WORKDIR /src

ARG APP_VERSION
ARG VCS_REF
ARG BUILD_DATE
ARG GO_VERSION
ARG CGO_ENABLED
ARG GOOS
ARG GOARCH
ARG LDFLAGS

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARCH=${GOARCH} \
    go build -a -tags netgo -trimpath -ldflags="${LDFLAGS}" \
    -o /app/sm-kms ./cmd/sm-kms

# Validate static linking
SHELL ["/bin/bash", "-c"]
RUN if ldd /app/sm-kms 2>/dev/null; then \
        echo "Binary is dynamically linked - failing build"; \
        exit 1; \
    fi

ARG CONTAINER_UID
ARG CONTAINER_GID
RUN mkdir -p /app && chmod 555 /app && chown -R ${CONTAINER_UID}:${CONTAINER_GID} /app

#############################################################################################
# Stage 3: Runtime dependencies
#############################################################################################
# hadolint ignore=DL3006
FROM alpine:${ALPINE_VERSION} AS runtime-deps
# hadolint ignore=DL3018
RUN apk --no-cache add ca-certificates tzdata tini && \
    update-ca-certificates

#############################################################################################
# Stage 4: Final image
#############################################################################################
# hadolint ignore=DL3006
FROM alpine:${ALPINE_VERSION} AS final

# Copy runtime dependencies
COPY --from=runtime-deps /sbin/tini /sbin/tini
COPY --from=runtime-deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy application binary and build metadata
COPY --from=builder /app /app
COPY --from=validation /app/.build-params /app/.build-params

# Create non-root user
ARG CONTAINER_UID
ARG CONTAINER_GID
RUN addgroup -g ${CONTAINER_GID} -S appgroup && \
    adduser -u ${CONTAINER_UID} -S appuser -G appgroup -h /app -s /sbin/nologin

WORKDIR /app

# OCI image labels
ARG APP_VERSION
ARG VCS_REF
ARG BUILD_DATE
LABEL org.opencontainers.image.title="cryptoutil-sm-kms" \
    org.opencontainers.image.description="Key Management" \
    org.opencontainers.image.source="https://github.com/justincranford/cryptoutil" \
    org.opencontainers.image.authors="Justin Cranford" \
    org.opencontainers.image.version="${APP_VERSION}" \
    org.opencontainers.image.revision="${VCS_REF}" \
    org.opencontainers.image.created="${BUILD_DATE}"

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD /app/sm-kms livez || exit 1

ENTRYPOINT ["/sbin/tini", "--"]

USER ${CONTAINER_UID}:${CONTAINER_GID}
