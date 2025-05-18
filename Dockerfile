# Stage 1 — build
FROM golang:1.24 AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=arm64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go controller.go ./
RUN go build -ldflags="-s -w" -o nomad-controller .

# Stage 2 — минимальный образ (альтернатива distroless)
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -g '' appuser

WORKDIR /app
COPY --from=builder --chown=appuser:appuser /app/nomad-controller .

USER appuser

ENTRYPOINT ["/app/nomad-controller"]