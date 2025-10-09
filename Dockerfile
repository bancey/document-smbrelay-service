# Build stage - install dependencies that require compilation
FROM python:3.14-slim AS builder

# Install build deps for pysmb
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libkrb5-dev \
    libsasl2-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy requirements and install Python packages
COPY requirements.txt /app/requirements.txt
RUN pip install --no-cache-dir --user -r /app/requirements.txt

# Runtime stage - minimal image without build dependencies
FROM python:3.14-slim

WORKDIR /app

# Install runtime libraries for Kerberos/GSSAPI support
RUN apt-get update && apt-get install -y --no-install-recommends \
    libkrb5-3 \
    libgssapi-krb5-2 \
    libsasl2-2 \
    krb5-user \
    && rm -rf /var/lib/apt/lists/*

# Copy installed packages from builder stage
COPY --from=builder /root/.local /root/.local

# Copy application code
COPY app /app/app

# Copy startup script
COPY startup.sh /app/startup.sh
RUN chmod +x /app/startup.sh

ENV PYTHONUNBUFFERED=1

# Configure default log level (can be overridden at runtime)
ENV LOG_LEVEL=INFO

# Make sure scripts in .local are usable
ENV PATH=/root/.local/bin:$PATH

EXPOSE 8080

# Health check configuration  
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8080/health', timeout=5)" || exit 1

CMD ["/app/startup.sh"]