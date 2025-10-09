# Build stage - install dependencies that require compilation
FROM python:3.13-slim AS builder

# Install build deps for pysmb
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libkrb5-dev \
    libsasl2-dev \
    python-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy requirements and install Python packages
COPY requirements.txt /app/requirements.txt
RUN pip install --no-cache-dir --user -r /app/requirements.txt

# Runtime stage - minimal image without build dependencies
FROM python:3.13-slim

WORKDIR /app

# Copy installed packages from builder stage
COPY --from=builder /root/.local /root/.local

# Copy application code
COPY app /app/app

ENV PYTHONUNBUFFERED=1

# Configure default log level (can be overridden at runtime)
ENV LOG_LEVEL=INFO

# Make sure scripts in .local are usable
ENV PATH=/root/.local/bin:$PATH

EXPOSE 8080

# Health check configuration  
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8080/health', timeout=5)" || exit 1

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8080"]
