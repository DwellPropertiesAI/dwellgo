# Multi-stage build for Dwell Property Management API
FROM golang:1.25-alpine AS builder

RUN go version

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Verify go.mod consistency
RUN go mod tidy

# Build the application
RUN CGO_ENABLED=0 go build -o dwell .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1001 -S dwell && \
    adduser -u 1001 -S dwell -G dwell

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/dwell .

# Copy database schema
COPY --from=builder /app/internal/database/schema.sql ./schema.sql

# Change ownership to non-root user
RUN chown -R dwell:dwell /app

# Switch to non-root user
USER dwell

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -fsS http://localhost:8080/api/v1/health || exit 1

# Run the application
CMD ["./dwell"]

