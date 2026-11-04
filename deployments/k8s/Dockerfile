# Current-source local Kubernetes suite image. Build the repository with
# --platform=linux/amd64 and its adjacent source-only Docker ignore file.
# The base is the inspected golang:1.27.3 linux/amd64 digest, not an invented pin.
FROM golang:1.27.3@sha256:2920bb124f1a2e12a1d39f0cab756ae0068441ee1da7655dec7f5e374944095b AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY api/ ./api/
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
# Preserve the image's native CGO/SQLite semantics, including its matching libc.
# Do not force static glibc (SQLite dlopen requires the matching runtime library).
# Use native Linux image defaults. Do not override Go/tool environment settings.
RUN go build -trimpath -o /build/cryptoutil ./cmd/cryptoutil && ldd /build/cryptoutil

FROM scratch
LABEL org.opencontainers.image.title="cryptoutil" \
    org.opencontainers.image.version="0.0.1" \
    org.opencontainers.image.source="https://github.com/justincranford/cryptoutil" \
    org.opencontainers.image.authors="Justin Cranford" \
    org.opencontainers.image.description="Current-source local Kubernetes suite runtime"
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /lib/x86_64-linux-gnu/libc.so.6 /lib/x86_64-linux-gnu/libc.so.6
COPY --from=builder /lib64/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2
COPY --from=builder --chown=65532:65532 /build/cryptoutil /app/cryptoutil
USER 65532:65532
WORKDIR /app
# Select the existing suite service/product route or lifecycle operation in Helm.
ENTRYPOINT ["/app/cryptoutil"]
