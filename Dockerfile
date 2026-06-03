# Compilation Environment ---
FROM golang:1.24.4-alpine AS builder
WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source trees and compile
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o rewards-api ./cmd/api/main.go

# Runtime Minimal Image Environment
FROM alpine:3.20
WORKDIR /app

# Pull built binary output from previous pipeline compilation stage
COPY --from=builder /app/rewards-api .

EXPOSE 8080
CMD ["./rewards-api"]
