# Stage 1: Build 
FROM golang:1.23-alpine AS builder 
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make protobuf protobuf-dev

# Install Go protoc plugins (use specific versions compatible with Go 1.23)
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.33.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Clone proto submodule (if not exists or empty)
RUN if [ ! -f "proto/services/order-service.proto" ]; then \
    echo "Proto submodule not found, cloning..." && \
    rm -rf proto && \
    git clone https://github.com/astrotify/astok-proto.git proto; \
    fi

# Verify proto files exist
RUN ls -la proto/services/ && ls -la proto/modules/

# Generate Go code from proto files
RUN mkdir -p go-proto && \
    protoc --go_out=. --go_opt=paths=import \
    --go-grpc_out=. --go-grpc_opt=paths=import \
    --proto_path=proto \
    proto/modules/*.proto proto/services/*.proto && \
    if [ -d "order-service/go-proto" ]; then \
    mv order-service/go-proto/* go-proto/ 2>/dev/null || true; \
    rm -rf order-service; \
    fi && \
    find go-proto -name "*.pb.go" -type f -exec sed -i 's|"go-proto/|"order-service/go-proto/|g' {} \; && \
    echo "✅ Proto files generated!"

# Verify generated files
RUN ls -la go-proto/

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/order-service ./cmd/server

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies (ca-certificates + netcat for health check)
RUN apk --no-cache add ca-certificates netcat-openbsd

# Create non-root user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/order-service .

# Change ownership and switch to non-root user
RUN chown -R appuser:appuser /app
USER appuser

EXPOSE 5001

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD nc -z localhost 5001 || exit 1

CMD ["./order-service"]
