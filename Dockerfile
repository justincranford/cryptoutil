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
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o cryptoutil .

#############################################################################################

FROM alpine:latest
WORKDIR /app
COPY --from=builder2 /app/cryptoutil /app/cryptoutil
RUN adduser -D -H -h /app cryptoutil               && \
    chown -R cryptoutil:cryptoutil /app            && \
    chmod +x /app/cryptoutil                       && \
    apk --no-cache add ca-certificates tzdata curl && \
    update-ca-certificates
EXPOSE 8080
USER cryptoutil

HEALTHCHECK --start-period=5s --interval=60s --timeout=3s --retries=3 \
  CMD curl -f http://localhost:8080/readyz || exit 1

ENTRYPOINT ["/app/cryptoutil", "--dev", "--migrations", "--log-level=INFO", "--bind-address=0.0.0.0"]
