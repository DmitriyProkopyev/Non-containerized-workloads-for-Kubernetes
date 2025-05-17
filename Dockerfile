# Stage 1 — build
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go controller.go ./
RUN go build -o nomad-controller .

# Stage 2 — минимальный образ
FROM gcr.io/distroless/base-debian11

WORKDIR /app
COPY --from=builder /app/nomad-controller .

ENTRYPOINT ["/app/nomad-controller"]
