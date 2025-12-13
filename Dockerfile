# Multi-stage Dockerfile for ISOMan
# Stage 1: Build frontend with Bun
# Stage 2: Build backend with Go
# Stage 3: Create minimal runtime image

# ============================================
# Stage 1: Build Frontend
# ============================================
FROM oven/bun:1-alpine AS frontend-builder

WORKDIR /app/ui

# Copy package files
COPY ui/package.json ui/bun.lock ./

# Install dependencies with cache mount
RUN --mount=type=cache,target=/root/.bun/install/cache \
    bun install --frozen-lockfile

# Copy frontend source
COPY ui/ ./

# Build frontend
RUN bun run build

# ============================================
# Stage 2: Build Backend
# ============================================
FROM golang:1.24-alpine AS backend-builder

# Build argument for version
ARG VERSION=dev

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy backend source
COPY backend/ ./

# Build backend binary (CGO_ENABLED=0 for static binary)
# Embed version in binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${VERSION}" \
    -o server .

# ============================================
# Stage 3: Runtime Image
# ============================================
FROM alpine:3.19

# Install ca-certificates for HTTPS downloads and su-exec for privilege dropping
RUN apk --no-cache add ca-certificates tzdata su-exec

# Create non-root user
RUN addgroup -g 1000 isoman && \
    adduser -D -u 1000 -G isoman isoman

WORKDIR /app

# Copy backend binary from builder
COPY --from=backend-builder /app/server ./server

# Copy frontend dist from builder
COPY --from=frontend-builder /app/ui/dist ./ui/dist

# Copy migrations directory from builder
COPY --from=backend-builder /app/migrations ./migrations

# Copy entrypoint script
COPY backend/docker-entrypoint.sh /entrypoint.sh

# Create data directory with proper permissions
RUN mkdir -p /data/isos /data/db && \
    chown -R isoman:isoman /app /data && \
    chmod +x /entrypoint.sh

# Set entrypoint (runs as root, then drops to isoman user)
ENTRYPOINT ["/entrypoint.sh"]

# Expose port
EXPOSE 8080

# Environment variables
ENV PORT=8080
ENV DATA_DIR=/data
ENV WORKER_COUNT=2
ENV GIN_MODE=release

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the server (passed to entrypoint)
CMD ["./server"]
