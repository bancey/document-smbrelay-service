# OpenTelemetry Integration

The Document SMB Relay Service includes comprehensive OpenTelemetry instrumentation for observability with support for Azure Application Insights and other OTLP-compatible backends.

## Features

- **📊 Distributed Tracing**: Automatic trace generation for all HTTP requests and SMB operations
- **📈 Metrics Collection**: HTTP request metrics, SMB operation metrics, and custom business metrics
- **🔍 Structured Logging**: Context-aware logging integrated with OpenTelemetry
- **☁️ Azure Application Insights**: Native support via OTLP protocol
- **🔌 OTLP Protocol**: Works with any OTLP-compatible backend (Jaeger, Prometheus, Grafana, etc.)
- **🛠️ Development Mode**: stdout exporters for local testing

## Quick Start

### Enable OpenTelemetry

Set the `OTEL_ENABLED` environment variable to enable telemetry:

```bash
export OTEL_ENABLED=true
export OTEL_SERVICE_NAME=document-smbrelay-service
export OTEL_SERVICE_VERSION=1.0.0
```

### Development Mode (stdout exporters)

When no OTLP endpoint is configured, telemetry is exported to stdout for development and testing:

```bash
export OTEL_ENABLED=true
./server
```

You'll see trace and metric data in the console output.

### Azure Application Insights

The easiest way to use Azure Application Insights is with the connection string:

```bash
export APPLICATIONINSIGHTS_CONNECTION_STRING="InstrumentationKey=xxx;IngestionEndpoint=https://xxx.applicationinsights.azure.com"
```

This automatically:
- Enables OpenTelemetry
- Configures the OTLP endpoint
- Sets up authentication headers

### Generic OTLP Backend

For other OTLP backends (Jaeger, Grafana Cloud, etc.):

```bash
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=your-otlp-endpoint:4318
export OTEL_EXPORTER_OTLP_HEADERS="x-api-key=your-api-key"
```

## Configuration Reference

### Core Configuration

| Environment Variable | Description | Default | Required |
|---------------------|-------------|---------|----------|
| `OTEL_ENABLED` | Enable OpenTelemetry instrumentation | `false` | No |
| `OTEL_SERVICE_NAME` | Service name for telemetry | `document-smbrelay-service` | No |
| `OTEL_SERVICE_VERSION` | Service version | `1.0.0` | No |
| `OTEL_TRACING_ENABLED` | Enable distributed tracing | `true` (if OTEL_ENABLED) | No |
| `OTEL_METRICS_ENABLED` | Enable metrics collection | `true` (if OTEL_ENABLED) | No |

### OTLP Exporter Configuration

| Environment Variable | Description | Default | Required |
|---------------------|-------------|---------|----------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP endpoint (without /v1/traces) | stdout exporter | No |
| `OTEL_EXPORTER_OTLP_HEADERS` | Additional headers (format: `key1=value1,key2=value2`) | none | No |

### Azure Application Insights

| Environment Variable | Description | Default | Required |
|---------------------|-------------|---------|----------|
| `APPLICATIONINSIGHTS_CONNECTION_STRING` | Azure App Insights connection string | none | No |

When this is set, it automatically enables OpenTelemetry and configures the OTLP endpoint.

## Instrumentation Details

### HTTP Request Tracing

Every HTTP request is automatically traced with:

- **Span Attributes**:
  - HTTP method, route, URL, scheme
  - Status code
  - Request and response body sizes
  - Client IP address
  - User agent

- **Span Status**: 
  - `Ok` for 2xx and 3xx responses
  - `Error` for 4xx and 5xx responses

Example trace:
```json
{
  "Name": "GET /health",
  "SpanKind": "Server",
  "Attributes": {
    "http.request.method": "GET",
    "http.route": "/health",
    "http.response.status_code": 200,
    "server.address": "localhost:8080",
    "client.address": "192.168.1.10"
  }
}
```

### SMB Operation Tracing

All SMB operations are traced with detailed context:

- **List Operations**:
  - Path being listed
  - Server and share names
  - Number of files returned
  - Duration

- **Upload Operations**:
  - Local and remote paths
  - File size
  - Overwrite flag
  - Duration

- **Delete Operations**:
  - Path being deleted
  - Server and share names
  - Duration

Example SMB trace:
```json
{
  "Name": "SMB upload",
  "SpanKind": "Client",
  "Attributes": {
    "smb.operation": "upload",
    "smb.path": "documents/report.pdf",
    "smb.server": "fileserver",
    "smb.share": "Documents",
    "smb.overwrite": false
  }
}
```

### Metrics Collection

The service collects the following metrics:

#### HTTP Metrics

| Metric Name | Type | Description | Labels |
|------------|------|-------------|--------|
| `http.server.request.duration` | Histogram | Duration of HTTP requests (ms) | method, route, status_code |
| `http.server.requests.total` | Counter | Total HTTP requests | method, route, status_code |
| `http.server.request.size` | Histogram | Request body size (bytes) | method, route, status_code |
| `http.server.response.size` | Histogram | Response body size (bytes) | method, route, status_code |

#### SMB Metrics

| Metric Name | Type | Description | Labels |
|------------|------|-------------|--------|
| `smb.operation.duration` | Histogram | Duration of SMB operations (ms) | operation |
| `smb.operations.total` | Counter | Total SMB operations | operation, success |
| `smb.errors.total` | Counter | Total SMB errors | operation, error |
| `smb.file.size` | Histogram | File sizes in operations (bytes) | operation |

## Usage Examples

### Example 1: Local Development with stdout

```bash
# Enable telemetry with stdout exporter
export OTEL_ENABLED=true
export OTEL_SERVICE_NAME=smb-relay-dev
export LOG_LEVEL=DEBUG

# Configure SMB
export SMB_SERVER_NAME=testserver
export SMB_SERVER_IP=192.168.1.10
export SMB_SHARE_NAME=testshare
export SMB_USERNAME=testuser
export SMB_PASSWORD=testpass

./server
```

Traces and metrics will be printed to stdout in JSON format.

### Example 2: Azure Application Insights

```bash
# Get your connection string from Azure Portal
export APPLICATIONINSIGHTS_CONNECTION_STRING="InstrumentationKey=your-key;IngestionEndpoint=https://region.applicationinsights.azure.com"

# Configure SMB
export SMB_SERVER_NAME=fileserver
export SMB_SERVER_IP=192.168.1.100
export SMB_SHARE_NAME=Documents
export SMB_USERNAME=user
export SMB_PASSWORD=pass

./server
```

Telemetry will be sent to Azure Application Insights. View traces and metrics in the Azure Portal.

### Example 3: Grafana Cloud

```bash
# Enable telemetry
export OTEL_ENABLED=true
export OTEL_SERVICE_NAME=document-smbrelay

# Configure Grafana Cloud OTLP endpoint
export OTEL_EXPORTER_OTLP_ENDPOINT=otlp-gateway-prod-us-east-0.grafana.net:443
export OTEL_EXPORTER_OTLP_HEADERS="Authorization=Basic base64-encoded-credentials"

# Configure SMB
export SMB_SERVER_NAME=fileserver
# ... other SMB config ...

./server
```

### Example 4: Jaeger

```bash
# Enable telemetry
export OTEL_ENABLED=true

# Configure Jaeger OTLP endpoint (running locally)
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318

# Configure SMB
export SMB_SERVER_NAME=fileserver
# ... other SMB config ...

./server
```

### Example 5: Disable Metrics but Keep Tracing

```bash
export OTEL_ENABLED=true
export OTEL_METRICS_ENABLED=false
export OTEL_EXPORTER_OTLP_ENDPOINT=your-trace-backend:4318

./server
```

### Example 6: Custom Headers for Authentication

```bash
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=your-backend:4318
export OTEL_EXPORTER_OTLP_HEADERS="x-api-key=secret-key,x-tenant-id=tenant123"

./server
```

## Docker Deployment

### Example docker-compose.yml with Azure Application Insights

```yaml
version: '3.8'
services:
  smbrelay:
    image: document-smbrelay:latest
    ports:
      - "8080:8080"
    environment:
      # SMB Configuration
      SMB_SERVER_NAME: fileserver
      SMB_SERVER_IP: 192.168.1.100
      SMB_SHARE_NAME: Documents
      SMB_USERNAME: user
      SMB_PASSWORD: pass
      
      # OpenTelemetry with Azure Application Insights
      APPLICATIONINSIGHTS_CONNECTION_STRING: "InstrumentationKey=xxx;IngestionEndpoint=https://xxx.applicationinsights.azure.com"
      OTEL_SERVICE_NAME: document-smbrelay
      OTEL_SERVICE_VERSION: 1.0.0
      
      # Logging
      LOG_LEVEL: INFO
```

### Example with Generic OTLP Backend

```yaml
version: '3.8'
services:
  smbrelay:
    image: document-smbrelay:latest
    ports:
      - "8080:8080"
    environment:
      # SMB Configuration
      SMB_SERVER_NAME: fileserver
      SMB_SERVER_IP: 192.168.1.100
      SMB_SHARE_NAME: Documents
      SMB_USERNAME: user
      SMB_PASSWORD: pass
      
      # OpenTelemetry
      OTEL_ENABLED: "true"
      OTEL_SERVICE_NAME: document-smbrelay
      OTEL_SERVICE_VERSION: 1.0.0
      OTEL_EXPORTER_OTLP_ENDPOINT: jaeger:4318
      
      # Logging
      LOG_LEVEL: INFO
      
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"  # Jaeger UI
      - "4318:4318"    # OTLP HTTP receiver
```

## Viewing Telemetry Data

### Azure Application Insights

1. Go to the Azure Portal
2. Navigate to your Application Insights resource
3. View telemetry in:
   - **Transaction search**: Individual traces
   - **Performance**: Request durations and dependencies
   - **Failures**: Error traces and exceptions
   - **Metrics**: Custom metrics and counters
   - **Application map**: Service topology

### Jaeger

1. Open Jaeger UI (http://localhost:16686)
2. Select service name from dropdown
3. Click "Find Traces"
4. Explore trace details with span hierarchy

### Grafana

1. Configure OpenTelemetry data source
2. Create dashboards with:
   - Request rate and duration panels
   - Error rate panels
   - SMB operation metrics
   - Custom business metrics

## Troubleshooting

### Telemetry not appearing

1. **Check if enabled**: Verify `OTEL_ENABLED=true`
2. **Check endpoint**: Ensure `OTEL_EXPORTER_OTLP_ENDPOINT` is correct
3. **Check logs**: Look for initialization messages:
   ```
   [INFO] Initializing OpenTelemetry instrumentation
   [INFO] Service: your-service-name, Version: 1.0.0
   [INFO] Tracing initialized successfully
   [INFO] Metrics initialized successfully
   ```
4. **Test with stdout**: Remove OTLP endpoint to use stdout exporter
5. **Network connectivity**: Ensure the service can reach the OTLP endpoint

### Azure Application Insights connection issues

1. Verify connection string format:
   ```
   InstrumentationKey=xxx;IngestionEndpoint=https://xxx.applicationinsights.azure.com
   ```
2. Check that ingestion endpoint is accessible
3. Verify API key/instrumentation key is valid
4. Check Azure Portal for any service issues

### High cardinality warnings

If you see warnings about high cardinality metrics:
- Consider reducing label dimensions
- Use trace sampling for high-volume endpoints
- Aggregate similar operations

## Performance Considerations

- **Minimal Overhead**: OpenTelemetry adds ~1-2ms per request
- **Async Export**: Traces and metrics are exported asynchronously
- **Batch Processing**: Data is batched before export (5s for traces, 60s for metrics)
- **Memory Usage**: Adds ~5-10MB depending on traffic volume
- **Sampling**: Currently using `AlwaysSample` - consider adjusting for high-volume production

## Best Practices

1. **Use meaningful service names**: Helps identify services in distributed traces
2. **Version your service**: Track behavior changes across versions
3. **Monitor metric cardinality**: Avoid high-cardinality labels
4. **Set appropriate log levels**: Use `INFO` in production, `DEBUG` for troubleshooting
5. **Configure alerts**: Set up alerts on key metrics (error rates, latency, etc.)
6. **Use dashboards**: Create dashboards for at-a-glance monitoring
7. **Test locally first**: Use stdout exporter to verify instrumentation

## Additional Resources

- [OpenTelemetry Documentation](https://opentelemetry.io/docs/)
- [Azure Monitor OpenTelemetry](https://learn.microsoft.com/en-us/azure/azure-monitor/app/opentelemetry-enable?tabs=go)
- [OTLP Specification](https://opentelemetry.io/docs/specs/otlp/)
- [Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/)
