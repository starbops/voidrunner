# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies including swag for docs generation
RUN apk add --no-cache git ca-certificates tzdata
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Set working directory
WORKDIR /app

# Copy go mod and sum files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate OpenAPI documentation
RUN swag init -g cmd/main.go -o docs

# Build the application with size optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-s -w" \
    -o voidrunner ./cmd/main.go

# Final stage - using distroless for minimal size and security
FROM gcr.io/distroless/static:nonroot

# Copy ca-certificates from builder (needed for HTTPS calls if any)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data from builder (needed for time operations)
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the binary from builder stage
COPY --from=builder /app/voidrunner /voidrunner

# Expose port
EXPOSE 8080

# Health check removed for minimal image size
# For health checks in production, use external monitoring or k8s probes

# Run the application
# distroless/static:nonroot already runs as non-root user (65532:65532)
CMD ["/voidrunner"]