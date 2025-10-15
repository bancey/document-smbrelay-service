# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Runtime stage - minimal image
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies including smbclient for SMB operations
# ca-certificates: for HTTPS
# krb5-libs: for Kerberos authentication support
# samba-client: provides smbclient binary for SMB operations with DFS support
RUN apk --no-cache add ca-certificates krb5-libs samba-client

# Copy the binary from builder
COPY --from=builder /app/server /app/server

# Copy startup script if it exists
COPY startup.sh /app/startup.sh 2>/dev/null || true
RUN chmod +x /app/startup.sh 2>/dev/null || true

ENV PORT=8080
ENV LOG_LEVEL=INFO

EXPOSE 8080

# Health check configuration
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Use startup script if it exists, otherwise run the binary directly
CMD ["/app/server"]
