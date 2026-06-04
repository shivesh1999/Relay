# Multi-stage build: compile in builder stage, run in minimal stage
# This keeps the final image small and secure

# Stage 1: Builder - compile the application
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/

# Build the application
# - Disable CGO for static linking
# - Strip debug symbols for smaller binary
# - Set version at build time (can be passed as ARG)
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version={$VERSION}" \
    -o /build/api \
    ./cmd/api/main.go

# Stage 2: Runtime - minimal image
FROM alpine:latest

# Install runtime dependencies
# - ca-certificates: for HTTPS connections
# - tzdata: for timezone support
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/api /app/api

# Create config directory
RUN mkdir -p configs && chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
# Kubernetes will use this to determine if container is healthy
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["/app/api"]
