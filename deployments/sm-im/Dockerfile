# Copyright (c) 2025 Justin Cranford
#
#

# Multi-stage Dockerfile for sm-im encrypted instant messaging service

# Build arguments
ARG GO_VERSION=1.25.7
ARG ALPINE_VERSION=3.19

# Stage 1: Builder - Compile the Go application
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /build

# Install build dependencies
# hadolint ignore=DL3018
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Copy go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . ./

# Build the application with static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static'" \
    -o /build/sm-im \
    ./cmd/sm-im

# Stage 2: Runtime - Minimal runtime image
FROM alpine:${ALPINE_VERSION} AS runtime

WORKDIR /app

# Install runtime dependencies
# hadolint ignore=DL3018
RUN apk add --no-cache \
        ca-certificates \
        wget \
        tzdata && \
    addgroup -S smim -g 1000 && \
    adduser -S smim -u 1000 -G smim

# Copy binary from builder
COPY --from=builder --chown=smim:smim /build/sm-im /app/sm-im

# Switch to non-root user
USER smim:smim

# Expose ports
EXPOSE 8888 9090

# Health check (admin server livez endpoint)
HEALTHCHECK --interval=10s --timeout=5s --start-period=30s --retries=3 \
    CMD wget --no-check-certificate --quiet --tries=1 --spider https://127.0.0.1:9090/admin/api/v1/livez || exit 1

# Run the application
ENTRYPOINT ["/app/sm-im"]
CMD []
