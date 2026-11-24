# Multi-stage build for Bkam Go API Gateway
FROM golang:alpine AS builder

WORKDIR /app

# Enable dynamic toolchain downloading if needed
ENV GOTOOLCHAIN=auto

# Install git, ca-certificates, and build tools
RUN apk add --no-cache git ca-certificates tzdata

# Copy module files and pre-fetch dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Compile statically linked Linux binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# Lean runtime container (~20MB)
FROM alpine:latest

# Install ca-certificates and tzdata for Africa/Cairo timezone operations
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy compiled executable from builder
COPY --from=builder /app/server /app/server

# Expose API Gateway port
EXPOSE 3001

# Start the Go API Gateway
CMD ["/app/server"]
