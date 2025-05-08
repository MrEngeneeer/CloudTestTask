# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o balancer ./cmd

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/balancer ./balancer

RUN mkdir configs
COPY configs/ ./configs/

EXPOSE 8080

ENTRYPOINT ["./balancer", "-config", "configs/config.yml"]
