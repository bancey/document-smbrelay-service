
# Build stage
FROM golang:1.25-alpine AS builder

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

# Install smbclient (which has native DFS support) and ca-certificates, tzdata
RUN apk --no-cache add \
    samba-client \
    ca-certificates \
    tzdata \
    bind-tools \
    netcat-openbsd

# Copy the binary from builder
COPY --from=builder /app/server /app/server

ENV PORT=8080
ENV LOG_LEVEL=INFO

RUN adduser -D -u 1000 appuser && \
    chown -R appuser:appuser /app && \
    mkdir -p /tmp && \
    chown appuser:appuser /tmp

USER appuser

EXPOSE 8080

# Health check configuration
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the server binary directly
CMD ["/app/server"]
