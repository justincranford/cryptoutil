#############################################################################################

FROM golang:latest AS builder1
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go mod vendor

#############################################################################################

FROM golang:latest AS builder2
WORKDIR /app
COPY --from=builder1 /go/pkg/mod /go/pkg/mod
COPY --from=builder1 /app        /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o cryptoutil ./cmd/cryptoutil

#############################################################################################

FROM alpine:latest AS runtime-deps
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

# Copy the static binary
COPY --from=builder2 /app/cryptoutil /cryptoutil

# Run as non-root user (65532:65532 industry standard)
# Note: scratch has no passwd file, so numeric IDs only
USER 65532:65532

# Use non-privileged ports
EXPOSE 8080 9090

# Note: No healthcheck possible with scratch (no shell/curl available)
# For Kubernetes: Use liveness/readiness probes like:
#   livenessProbe:
#     httpGet:
#       path: /healthz
#       port: 9090
# For Docker Compose: Use external healthcheck or depends_on with condition

# Use tini for proper signal handling and zombie process reaping
ENTRYPOINT ["/tini", "--", "/cryptoutil"]
