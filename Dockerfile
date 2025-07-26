
FROM golang:latest AS builder1
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go mod vendor

FROM golang:latest AS builder2
WORKDIR /app
COPY --from=builder1 /go/pkg/mod /go/pkg/mod
COPY --from=builder1 /app        /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o cryptoutil .

FROM scratch
WORKDIR /app
ENV USER=cryptoutil
COPY --from=builder2 /app/cryptoutil /app/cryptoutil
ENTRYPOINT ["/app/cryptoutil", "--dev"]
