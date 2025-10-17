
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

# Base runtime stage - shared between production and debug
FROM alpine:latest AS base

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

# Set smbclient path - Alpine installs it to /usr/bin/smbclient
ENV SMBCLIENT_PATH=/usr/bin/smbclient

RUN adduser -D -u 1000 appuser && \
    chown -R appuser:appuser /app && \
    mkdir -p /tmp && \
    chown appuser:appuser /tmp

# Production stage - minimal image (default)
FROM base AS production

# Switch to non-root user for security
USER appuser

EXPOSE 8080

# Health check configuration
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the server binary directly
CMD ["/app/server"]

# Debug stage - includes SSH for troubleshooting
FROM base AS debug

# Install OpenSSH and set up SSH access for debugging
RUN apk add --no-cache openssh su-exec \
    && echo "root:Docker!" | chpasswd \
    && mkdir -p /var/run/sshd \
    && ssh-keygen -A

# Create sshd_config inline
RUN cat > /etc/ssh/sshd_config << 'EOF'
Port 2222
ListenAddress 0.0.0.0
LoginGraceTime 180
X11Forwarding yes
Ciphers aes128-cbc,3des-cbc,aes256-cbc,aes128-ctr,aes192-ctr,aes256-ctr
MACs hmac-sha1,hmac-sha1-96
StrictModes yes
SyslogFacility DAEMON
PasswordAuthentication yes
PermitEmptyPasswords no
PermitRootLogin yes
Subsystem sftp internal-sftp
EOF

COPY entrypoint.debug.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

# Note: We don't switch to appuser here because sshd needs to run as root
# The entrypoint script will handle running sshd as root and then su to appuser for the app

EXPOSE 8080 2222

# Health check configuration
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the server binary via entrypoint that starts sshd first
ENTRYPOINT ["./entrypoint.sh"]
