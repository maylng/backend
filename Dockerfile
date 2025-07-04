# Build stage - Use specific version instead of latest
FROM golang:1.21.5-alpine3.19 AS builder

# Create non-root user for build
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Install build dependencies and security updates
RUN apk add --no-cache git && \
    apk upgrade --no-cache

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with security flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o main ./cmd/api

# Final stage - Use specific distroless image for better security
FROM gcr.io/distroless/static-debian11:nonroot

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder --chown=nonroot:nonroot /app/main .

# Copy migrations
COPY --from=builder --chown=nonroot:nonroot /app/migrations ./migrations

# Expose port
EXPOSE 8080

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/main", "--health-check"] || exit 1

# Run the binary
ENTRYPOINT ["/app/main"]
