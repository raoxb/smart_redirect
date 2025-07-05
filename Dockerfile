# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o smart_redirect cmd/server/main.go

# Development stage
FROM builder AS development
RUN go install github.com/cosmtrek/air@latest
WORKDIR /app
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# Production stage
FROM alpine:latest AS production

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/smart_redirect .

# Copy configuration files and static assets
COPY --from=builder /app/config ./config
COPY --from=builder /app/geoip ./geoip

# Create directories for logs and data
RUN mkdir -p logs data

# Create a non-root user
RUN adduser -D -s /bin/sh appuser && \
    chown -R appuser:appuser /root
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Command to run
CMD ["./smart_redirect"]