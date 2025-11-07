# ============================================================================
# SECURITY WARNING - READ BEFORE DEPLOYING
# ============================================================================
#
# EngineerDNA V1 is designed for LOCALHOST-ONLY use and has:
#   - NO built-in authentication
#   - NO authorization
#   - NO TLS/HTTPS
#   - NO multi-user security
#
# This Docker container sets ENGINEERDNA_HOST=0.0.0.0 which exposes the
# application to your LOCAL NETWORK. This means:
#   - Anyone on your local network can access the application
#   - All data is transmitted UNENCRYPTED
#   - No user authentication is required
#   - All engineering metrics are visible to anyone with network access
#
# ONLY use this deployment if:
#   - You trust everyone on your local network
#   - You have firewall rules restricting access
#   - You understand the security implications
#
# For production use, wait for V2 which includes:
#   - JWT authentication
#   - Role-based authorization
#   - TLS/HTTPS support
#   - API key management
#
# ============================================================================

# Build stage - Frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend dependencies
COPY frontend/package*.json ./
RUN npm ci --prefer-offline --no-audit

# Copy frontend source and build
COPY frontend/ ./
RUN npm run build

# Build stage - Backend
FROM golang:1.25-alpine AS backend-builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make

# Copy Go modules and plugin-sdk (needed for replace directive)
COPY go.mod go.sum ./
COPY plugins/plugin-sdk/ ./plugins/plugin-sdk/
RUN go mod download

# Copy source
COPY . .

# Copy built frontend from previous stage
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Build with frontend embedded
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o engineerdna .

# Runtime stage
FROM alpine:3.19

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=backend-builder /app/engineerdna .

# Create data directory
RUN mkdir -p /app/data

# Environment variables
# WARNING: Setting ENGINEERDNA_HOST=0.0.0.0 exposes to local network
# Only use in trusted environments or with proper firewall rules
ENV ENGINEERDNA_HOST=0.0.0.0
ENV ENGINEERDNA_PORT=3847
ENV ENGINEERDNA_DATA_DIR=/app/data

# Expose port
EXPOSE 3847

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3847/api/health || exit 1

# Run as non-root user
RUN addgroup -g 1000 engineerdna && \
    adduser -D -u 1000 -G engineerdna engineerdna && \
    chown -R engineerdna:engineerdna /app

USER engineerdna

# Print security warning on startup
CMD echo "========================================" && \
    echo "SECURITY WARNING" && \
    echo "========================================" && \
    echo "EngineerDNA is running on 0.0.0.0:3847" && \
    echo "This exposes the application to your local network" && \
    echo "Anyone on your network can access WITHOUT authentication" && \
    echo "========================================" && \
    ./engineerdna init && \
    ./engineerdna serve
