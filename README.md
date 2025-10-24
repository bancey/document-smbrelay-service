# Document SMB Relay Service

A high-performance HTTP service built with Go that accepts file uploads and writes them directly to SMB shares without mounting them on the host filesystem.

## Features

- **⚡ Fast & Lightweight**: Written in Go for high performance (~100ms startup, ~10MB memory)
- **🚀 Simple HTTP API**: POST files via multipart/form-data
- **📁 Direct SMB Access**: Uses the `smbclient` binary for robust SMB operations
- **🌐 Full DFS Support**: Native DFS referral handling via smbclient
- **❤️ Health Checks**: Built-in endpoint for monitoring
- **📚 OpenAPI Documentation**: Interactive Swagger UI at `/docs`
- **🐳 Docker Support**: Multi-stage builds for minimal image size (~30MB)
- **🔐 Flexible Authentication**: Supports NTLM, Negotiate, and Kerberos protocols
- **📊 OpenTelemetry Integration**: Full observability with traces, metrics, and Azure Application Insights support (via OpenTelemetry Collector)

## Quick Start

### Prerequisites
- Go 1.21 or higher
- `smbclient` binary installed (from Samba package)
- SMB server accessible on your network

### Install smbclient

**macOS:**
```bash
brew install samba
```

**Ubuntu/Debian:**
```bash
sudo apt-get install smbclient
```

**Alpine (Docker):**
```bash
apk add samba-client
```

### Build and Run

```bash
# Download dependencies
go mod download

# Build the application
go build -o server ./cmd/server

# Run with environment variables
export SMB_SERVER_NAME=your-server
export SMB_SERVER_IP=192.168.1.10
export SMB_SHARE_NAME=Documents
export SMB_USERNAME=your-username
export SMB_PASSWORD=your-password

./server
```

Or use Make:
```bash
make build
make run
```

See [QUICKSTART.md](QUICKSTART.md) for more detailed getting started guide.

## Configuration

### Required Environment Variables

- `SMB_SERVER_NAME`: NetBIOS name of the SMB server
- `SMB_SERVER_IP`: IP address or hostname of the SMB server
- `SMB_SHARE_NAME`: Name of the SMB share (e.g., `Documents`)
- `SMB_USERNAME`: SMB username for authentication (optional for Kerberos)
- `SMB_PASSWORD`: SMB password for authentication (optional for Kerberos)
  - **Supports all special characters** including Base64 (`+`, `/`, `=`), symbols, quotes, spaces, unicode, etc.
  - Passed securely via environment variables (not visible in process list)
  - See [BASE64_PASSWORD_SUPPORT.md](BASE64_PASSWORD_SUPPORT.md) for details

### Optional Environment Variables

- `SMB_DOMAIN`: SMB domain/workgroup (default: empty)
- `SMB_PORT`: SMB port (default: `445`)
- `SMB_BASE_PATH`: Base path within the SMB share to restrict operations to a specific subdirectory (default: empty - full share access)
  - Example: `apps/myapp` restricts all file operations to that subdirectory
  - All relative paths in API requests are resolved relative to this base path
  - See [Base Path Configuration](#base-path-configuration) section below
- `SMB_USE_NTLM_V2`: Enable NTLMv2 (default: `true`, deprecated - use `SMB_AUTH_PROTOCOL`)
- `SMB_AUTH_PROTOCOL`: Authentication protocol - `negotiate|ntlm|kerberos` (default: derived from `SMB_USE_NTLM_V2`)
- `LOG_LEVEL`: Application log level - `DEBUG|INFO|WARNING|ERROR` (default: `INFO`)
- `LOG_SMB_COMMANDS` or `SMB_LOG_COMMANDS`: Enable debug logging of smbclient commands - `true|false` (default: `false`)
  - Error output visible at INFO level
  - Success output visible at DEBUG level
  - See [LOGGING_OUTPUT_IMPROVEMENTS.md](LOGGING_OUTPUT_IMPROVEMENTS.md) for details
- `PORT`: HTTP server port (default: `8080`)
- `SMBCLIENT_PATH`: Path to smbclient binary (default: auto-detected from PATH or common locations)

#### Retry Configuration

The service automatically retries network-related SMB operations (connection failures, timeouts, etc.) with configurable exponential backoff:

- `SMB_MAX_RETRIES`: Maximum number of retry attempts for transient network errors (default: `3`)
  - Set to `0` to disable retries
  - Each SMB operation will attempt up to `1 + SMB_MAX_RETRIES` times (initial attempt + retries)
- `SMB_RETRY_INITIAL_DELAY`: Initial delay in seconds before first retry (default: `1.0`)
- `SMB_RETRY_MAX_DELAY`: Maximum delay in seconds between retries (default: `30.0`)
- `SMB_RETRY_BACKOFF`: Exponential backoff multiplier (default: `2.0`)
  - Delay calculation: `initial_delay * (backoff ^ attempt_number)`, capped at `max_delay`

**What gets retried:**
- Transient network errors: connection refused, timeouts, network unreachable, broken pipe
- SMB protocol timeouts: `NT_STATUS_IO_TIMEOUT`, `NT_STATUS_CONNECTION_REFUSED`

**What doesn't get retried:**
- Authentication failures: `NT_STATUS_LOGON_FAILURE`
- Permission errors: `NT_STATUS_ACCESS_DENIED`
- Invalid share: `NT_STATUS_BAD_NETWORK_NAME`
- File not found: `NT_STATUS_OBJECT_NAME_NOT_FOUND`
- File conflicts: `NT_STATUS_OBJECT_NAME_COLLISION`

**Example with custom retry settings:**
```bash
export SMB_MAX_RETRIES=5              # Allow up to 5 retries
export SMB_RETRY_INITIAL_DELAY=2.0    # Start with 2 second delay
export SMB_RETRY_MAX_DELAY=60.0       # Cap delays at 60 seconds
export SMB_RETRY_BACKOFF=2.0          # Double delay each retry (2s, 4s, 8s, 16s, 32s, 60s)
```

#### OpenTelemetry / Observability

The service includes comprehensive OpenTelemetry instrumentation for distributed tracing, metrics, and logging:

- `OTEL_ENABLED`: Enable OpenTelemetry instrumentation - `true|false` (default: `false`)
- `OTEL_SERVICE_NAME`: Service name for telemetry (default: `document-smbrelay-service`)
- `OTEL_SERVICE_VERSION`: Service version (default: `1.0.0`)
- `OTEL_TRACING_ENABLED`: Enable distributed tracing - `true|false` (default: `true` if `OTEL_ENABLED`)
- `OTEL_METRICS_ENABLED`: Enable metrics collection - `true|false` (default: `true` if `OTEL_ENABLED`)
- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTLP endpoint for traces and metrics (e.g., `localhost:4318`)
- `OTEL_EXPORTER_OTLP_HEADERS`: Additional headers for OTLP requests (format: `key1=value1,key2=value2`)

**Example with generic OTLP backend:**
```bash
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4318
export OTEL_SERVICE_NAME=my-smbrelay
```

**Azure Application Insights Integration:**

> **⚠️ Important**: Azure Application Insights does NOT support direct OTLP ingestion for Go applications. You must deploy an OpenTelemetry Collector as a bridge. See the complete setup guide in [docs/OPENTELEMETRY.md](docs/OPENTELEMETRY.md#azure-application-insights).

Quick setup:
```bash
# 1. Deploy OpenTelemetry Collector (see docker-compose.azure-insights.yml)
# 2. Configure your application to use the collector:
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=collector-host:4318
```

See [docs/OPENTELEMETRY.md](docs/OPENTELEMETRY.md) for complete OpenTelemetry documentation including:
- Detailed configuration options
- Azure Application Insights setup with OpenTelemetry Collector
- Supported OTLP backends (Jaeger, Grafana, etc.)
- Metrics and traces collected
- Usage examples and troubleshooting

## Authentication Methods

### 1. NTLM (Default)
Username/password authentication using NTLM protocol.

```bash
export SMB_AUTH_PROTOCOL=ntlm
export SMB_USERNAME=myuser
export SMB_PASSWORD=mypassword
```

### 2. Negotiate
Automatic protocol negotiation.

```bash
export SMB_AUTH_PROTOCOL=negotiate
export SMB_USERNAME=myuser
export SMB_PASSWORD=mypassword
```

### 3. Kerberos
Kerberos authentication for Active Directory environments. Username/password are optional - can use system ticket cache.

```bash
export SMB_AUTH_PROTOCOL=kerberos
export SMB_USERNAME=myuser  # Optional
export SMB_PASSWORD=mypassword  # Optional
```

## Windows DFS Support

This service **fully supports Windows Distributed File System (DFS)** shares. The `smbclient` binary handles DFS referrals and path resolution natively and automatically.

**Example DFS Configuration:**

```bash
export SMB_SERVER_NAME=dfs.corp.example.com
export SMB_SERVER_IP=dfs.corp.example.com
export SMB_SHARE_NAME=documents
export SMB_AUTH_PROTOCOL=kerberos  # Or negotiate
export SMB_USERNAME=myuser
export SMB_PASSWORD=mypassword
```

See [DFS_TESTING.md](DFS_TESTING.md) for more details.

## Base Path Configuration

You can configure the service to operate within a specific subdirectory of an SMB share using the `SMB_BASE_PATH` environment variable. This is useful when you want to restrict file operations to a specific folder.

**Example Scenario:**

Your SMB share structure:
```
\\example.com\data\
  ├── apps/
  │   ├── myapp/          ← You want operations here
  │   │   ├── inbox/
  │   │   └── archive/
  │   └── otherapp/
  └── shared/
```

**Configuration:**

```bash
export SMB_SERVER_NAME=example.com
export SMB_SERVER_IP=192.168.1.10
export SMB_SHARE_NAME=data
export SMB_BASE_PATH=apps/myapp    # Restrict to this subdirectory
export SMB_USERNAME=myuser
export SMB_PASSWORD=mypassword
```

**How It Works:**

With `SMB_BASE_PATH=apps/myapp`, all API operations are relative to that directory:

- Uploading to `inbox/file.pdf` → Stored at `\\example.com\data\apps\myapp\inbox\file.pdf`
- Listing `archive` → Lists `\\example.com\data\apps\myapp\archive\`
- Deleting `old.txt` → Deletes `\\example.com\data\apps\myapp\old.txt`

**Benefits:**

- ✅ **Security**: Prevents access to files outside the base path
- ✅ **Simplicity**: API clients don't need to know the full share structure
- ✅ **Multi-tenancy**: Deploy multiple instances with different base paths
- ✅ **Validation**: Health check verifies the base path exists and is accessible

**Notes:**

- Base path uses forward slashes (`/`) for consistency
- Leading and trailing slashes are automatically normalized
- Backslashes (`\`) are automatically converted to forward slashes
- Leave empty (default) for full share access

## API Endpoints

### GET /health

Health check endpoint that verifies application and SMB connectivity.

**Note:** If `SMB_BASE_PATH` is configured, the health check also validates that the base path exists and is accessible.

**Response (200 OK)**:
```json
{
  "status": "healthy",
  "app_status": "ok",
  "smb_connection": "ok",
  "smb_share_accessible": true,
  "server": "testserver (192.168.1.10:445)",
  "share": "Documents"
}
```

**Response (503 Service Unavailable)**:
```json
{
  "status": "unhealthy",
  "smb_connection": "failed",
  "error": "connection error details"
}
```

**Response (503 with invalid base path)**:
```json
{
  "status": "unhealthy",
  "smb_connection": "ok",
  "smb_share_accessible": false,
  "error": "base path validation failed: base path does not exist: apps/myapp"
}
```

### GET /list

List files and folders at a given path on the SMB share.

**Query Parameters**:
- `path`: Optional path within the SMB share (defaults to root)

**Response (200 OK)**:
```json
{
  "path": "subfolder",
  "files": [
    {
      "name": "document.pdf",
      "size": 1024,
      "is_dir": false,
      "timestamp": "Mon Jan 1 12:34:56 2024"
    },
    {
      "name": "reports",
      "size": 0,
      "is_dir": true,
      "timestamp": "Mon Jan 1 10:00:00 2024"
    }
  ]
}
```

**Response (404 Not Found)** - path does not exist:
```json
{
  "detail": "path not found: nonexistent"
}
```

**Response (403 Forbidden)** - access denied:
```json
{
  "detail": "access denied to path: protected"
}
```

**Response (500 Internal Server Error)**:
```json
{
  "detail": "error message"
}
```

### POST /upload

Upload a file to the SMB share.

**Request** (multipart/form-data):
- `file`: The file to upload
- `remote_path`: Path within the SMB share (e.g., `inbox/report.pdf`)
- `overwrite`: Optional boolean, defaults to `false`

**Response (200 OK)**:
```json
{
  "status": "ok",
  "remote_path": "inbox/report.pdf"
}
```

**Response (409 Conflict)** - file exists and overwrite is false:
```json
{
  "detail": "remote file already exists: inbox/report.pdf"
}
```

### DELETE /delete

Delete a file from the SMB share.

**Query Parameters**:
- `path`: Path to the file within the SMB share (required)

**Response (200 OK)**:
```json
{
  "status": "ok",
  "path": "folder/file.txt"
}
```

**Response (400 Bad Request)** - invalid path or attempting to delete directory:
```json
{
  "detail": "invalid remote path: cannot delete root directory"
}
```

**Response (403 Forbidden)** - access denied:
```json
{
  "detail": "access denied: cannot delete folder/file.txt"
}
```

**Response (404 Not Found)** - file not found:
```json
{
  "detail": "file not found: folder/file.txt"
}
```

**Response (500 Internal Server Error)**:
```json
{
  "detail": "error message"
}
```

### GET /docs

Interactive Swagger UI documentation interface.

### GET /openapi.json

OpenAPI 3.0 specification in JSON format.

## Usage Examples

### Upload a file (fail if exists)
```bash
curl -X POST http://localhost:8080/upload \
  -F file=@document.pdf \
  -F remote_path=inbox/report.pdf
```

### Upload with overwrite enabled
```bash
curl -X POST http://localhost:8080/upload \
  -F file=@document.pdf \
  -F remote_path=inbox/report.pdf \
  -F overwrite=true
```

### Delete a file
```bash
curl -X DELETE "http://localhost:8080/delete?path=inbox/report.pdf"
```

### Delete a file with special characters
```bash
curl -X DELETE "http://localhost:8080/delete?path=folder/my%20file.txt"
```

### Check service health
```bash
curl http://localhost:8080/health | jq
```

### List files in root directory
```bash
curl http://localhost:8080/list | jq
```

### List files in a subdirectory
```bash
curl "http://localhost:8080/list?path=subfolder" | jq
```

## Usage Examples with Base Path

When using `SMB_BASE_PATH`, all operations are relative to that directory:

### Configuration
```bash
export SMB_SERVER_NAME=example.com
export SMB_SERVER_IP=192.168.1.10
export SMB_SHARE_NAME=data
export SMB_BASE_PATH=apps/myapp    # Restrict to this subdirectory
export SMB_USERNAME=myuser
export SMB_PASSWORD=mypassword
```

### Upload to base path + relative path
```bash
# Uploads to: \\example.com\data\apps\myapp\inbox\report.pdf
curl -X POST http://localhost:8080/upload \
  -F file=@report.pdf \
  -F remote_path=inbox/report.pdf
```

### List files in base path + subdirectory
```bash
# Lists: \\example.com\data\apps\myapp\archive\
curl "http://localhost:8080/list?path=archive" | jq
```

### Delete from base path
```bash
# Deletes: \\example.com\data\apps\myapp\old.txt
curl -X DELETE "http://localhost:8080/delete?path=old.txt"
```

### Health check validates base path
```bash
# Verifies that \\example.com\data\apps\myapp exists and is accessible
curl http://localhost:8080/health | jq
```

## Docker

The Dockerfile uses multi-stage builds with two targets: `production` (default) and `debug`.

### Build Image

**Standard Production Image (recommended):**
```bash
docker build -t document-smbrelay:latest .
# or explicitly:
docker build --target production -t document-smbrelay:latest .
```

Or use Make:
```bash
make docker-build
```

**Debug Image (with SSH for troubleshooting):**
```bash
docker build --target debug -t document-smbrelay:debug .
```

Or use Make:
```bash
make docker-build-debug
```

The debug image includes:
- SSH server on port 2222
- Root access with password `Docker!`
- Useful for troubleshooting container issues
- **Not recommended for production use**

### Run Container

**Standard image:**
```bash
docker run --rm -p 8080:8080 \
  -e SMB_SERVER_NAME=MYSMBSERVER \
  -e SMB_SERVER_IP=192.168.1.10 \
  -e SMB_SHARE_NAME=Documents \
  -e SMB_USERNAME=smbuser \
  -e SMB_PASSWORD='smb-password' \
  -e LOG_LEVEL=INFO \
  document-smbrelay:latest
```

**Debug image (with SSH access):**
```bash
docker run --rm -p 8080:8080 -p 2222:2222 \
  -e SMB_SERVER_NAME=MYSMBSERVER \
  -e SMB_SERVER_IP=192.168.1.10 \
  -e SMB_SHARE_NAME=Documents \
  -e SMB_USERNAME=smbuser \
  -e SMB_PASSWORD='smb-password' \
  -e LOG_LEVEL=INFO \
  document-smbrelay:debug

# SSH into debug container:
ssh root@localhost -p 2222
# Password: Docker!
```

### Multi-Architecture Builds

**Standard image:**
```bash
docker buildx build --platform linux/amd64,linux/arm64 \
  --target production \
  -t document-smbrelay:latest .
```

**Debug image:**
```bash
docker buildx build --platform linux/amd64,linux/arm64 \
  --target debug \
  -t document-smbrelay:debug .
```

### Docker Compose

See [DOCKER_TESTING.md](DOCKER_TESTING.md) for docker-compose examples with test SMB servers.

For detailed information about the standard and debug Docker images, see [DOCKER_IMAGES.md](DOCKER_IMAGES.md).

## Deployment

### Azure App Service

This service can be deployed to Azure App Service with OpenTelemetry Collector integration for Application Insights.

**Quick Deploy:**
```bash
./deploy-azure.sh
```

The automated script will:
- ✅ Create all required Azure resources (Resource Group, Application Insights, Container Registry, App Service)
- ✅ Build and push Docker images
- ✅ Configure environment variables and health checks
- ✅ Set up OpenTelemetry Collector integration

**Deployment Options:**
1. **Multi-Container (Recommended):** Separate app and collector containers using Docker Compose
2. **Single-Container:** Combined app + collector using supervisord

**Manual Deployment:**

For complete step-by-step instructions, architecture diagrams, security best practices, and troubleshooting, see:

📖 **[Azure App Service Deployment Guide](docs/AZURE_APP_SERVICE_DEPLOYMENT.md)**

The guide covers:
- Complete Azure CLI setup instructions
- Both multi-container and single-container approaches
- Application Insights integration with OpenTelemetry
- Security configuration and secrets management
- Cost optimization strategies
- Comprehensive troubleshooting section

## Development

### Running Tests

```bash
go test ./...
# or
make test
```

### Code Formatting

```bash
go fmt ./...
# or
make fmt
```

### Linting

```bash
go vet ./...
# or
make vet
```

### Clean Build Artifacts

```bash
make clean
```

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration management
│   ├── handlers/
│   │   └── handlers.go       # HTTP handlers
│   ├── logger/
│   │   └── logger.go         # Logging utilities
│   └── smb/
│       ├── connection.go     # SMB connection handling
│       └── operations.go     # SMB file operations
├── Dockerfile.go             # Docker build configuration
├── Makefile                  # Build and development tasks
├── go.mod                    # Go module definition
├── go.sum                    # Dependency checksums
└── README.md                 # This file
```

## Dependencies

- **[gofiber/fiber](https://github.com/gofiber/fiber)** v2 - Express-inspired web framework
- **[hirochachacha/go-smb2](https://github.com/hirochachacha/go-smb2)** v1.1.0 - SMB2/3 client library

## Performance

- **Startup Time**: ~100ms
- **Memory Usage**: ~10MB idle, ~30MB during uploads
- **Docker Image Size**: ~20MB (Alpine-based multi-stage build)
- **Request Latency**: <100ms for small files on local network

## Troubleshooting

### "Missing SMB configuration environment variables"
**Solution**: Ensure all required `SMB_*` environment variables are set before starting the server.

### "Connection refused"
**Solution**: 
- Verify SMB server IP/port and credentials
- Test connectivity: `telnet <SMB_SERVER_IP> 445`
- Check firewall rules

### Build Errors
**Solution**:
- Ensure Go 1.21+ is installed: `go version`
- Clean and rebuild: `make clean && make build`
- Update dependencies: `go mod download && go mod tidy`

### "Authentication failed"
**Solution**:
- Verify `SMB_USERNAME` and `SMB_PASSWORD`
- Check domain if using domain accounts: `DOMAIN\username`
- Try different `SMB_AUTH_PROTOCOL` values

## Documentation

- **[QUICKSTART.md](QUICKSTART.md)** - Quick start guide
- **[README-GO.md](README-GO.md)** - Detailed Go implementation guide  
- **[VERIFICATION.md](VERIFICATION.md)** - Testing and verification checklist
- **[BASE64_PASSWORD_SUPPORT.md](BASE64_PASSWORD_SUPPORT.md)** - Base64 and special character password support
- **[LOGGING_OUTPUT_IMPROVEMENTS.md](LOGGING_OUTPUT_IMPROVEMENTS.md)** - Enhanced error logging and output visibility
- **[SMB_DEBUG_LOGGING.md](SMB_DEBUG_LOGGING.md)** - Debug logging configuration
- **[SMB_DEBUG_EXAMPLES.md](SMB_DEBUG_EXAMPLES.md)** - Debug logging examples
- **[DFS_TESTING.md](DFS_TESTING.md)** - DFS testing guide
- **[DOCKER_TESTING.md](DOCKER_TESTING.md)** - Docker testing guide
- **[DOCKER_IMAGES.md](DOCKER_IMAGES.md)** - Docker image variants (standard vs debug)

## Contributing

We welcome contributions! Please see our [CONTRIBUTING.md](CONTRIBUTING.md) guide for details on:

- Setting up your development environment
- Code standards and best practices
- Testing guidelines
- Pull request process
- Issue reporting guidelines

Quick checklist:
1. Code is formatted with `go fmt`
2. All tests pass: `go test ./...`
3. Update documentation as needed

## License

BSD-2-Clause (same as go-smb2 library)

## Roadmap

- [ ] Comprehensive integration test suite
- [ ] Performance benchmarks
- [ ] Metrics and monitoring endpoints (Prometheus)
- [ ] Rate limiting middleware
- [ ] Request logging middleware
- [ ] Multiple file upload support
- [ ] Batch operations

## Support

For issues, questions, or contributions, please open an issue on GitHub.

## Configuration (Environment Variables)

### Required Environment Variables
- `SMB_SERVER_NAME`: NetBIOS name of the SMB server
- `SMB_SERVER_IP`: IP address or hostname of the SMB server
- `SMB_SHARE_NAME`: Name of the SMB share (for example `Documents`)
- `SMB_USERNAME`: SMB username for authentication
- `SMB_PASSWORD`: SMB password for authentication

**Optional Environment Variables**
- `SMB_DOMAIN`: SMB domain/workgroup (default: empty)
- `SMB_PORT`: SMB port (default: `445`)
- `SMB_USE_NTLM_V2`: `true|false` (default: `true`, deprecated - use `SMB_AUTH_PROTOCOL` instead)
- `SMB_AUTH_PROTOCOL`: Authentication protocol - `negotiate|ntlm` (default: derived from `SMB_USE_NTLM_V2`)
- `LOG_LEVEL`: Application log level - `DEBUG|INFO|WARNING|ERROR|CRITICAL` (default: `INFO`)

**Retry Configuration** (for unreliable networks)
- `SMB_MAX_RETRIES`: Maximum retry attempts for network errors (default: `3`)
- `SMB_RETRY_INITIAL_DELAY`: Initial retry delay in seconds (default: `1.0`)
- `SMB_RETRY_MAX_DELAY`: Maximum retry delay in seconds (default: `30.0`)
- `SMB_RETRY_BACKOFF`: Exponential backoff multiplier (default: `2.0`)

**Authentication Methods**

This service supports two authentication protocols:

1. **NTLM** (default): Username/password authentication using NTLM protocol
   - Set `SMB_AUTH_PROTOCOL=ntlm` or `SMB_USE_NTLM_V2=true`
   - Requires `SMB_USERNAME` and `SMB_PASSWORD`

2. **Negotiate**: Automatic protocol negotiation
   - Set `SMB_AUTH_PROTOCOL=negotiate` or `SMB_USE_NTLM_V2=false`
   - Requires `SMB_USERNAME` and `SMB_PASSWORD`

**Windows DFS Support**

This service **fully supports Windows Distributed File System (DFS)** shares. The underlying `go-smb2` library automatically handles DFS referrals and path resolution.

To use with Windows DFS:
- Set `SMB_SERVER_NAME` to your DFS namespace server (e.g., `dfs.example.com`)
- Set `SMB_SHARE_NAME` to the DFS share name
- Use Negotiate authentication for best compatibility with domain-joined environments
- The service will automatically follow DFS referrals to the actual file server

**Example: Windows DFS with Negotiate Authentication**

```bash
SMB_SERVER_NAME=dfs.corp.example.com \
SMB_SERVER_IP=dfs.corp.example.com \
SMB_SHARE_NAME=documents \
SMB_USERNAME=myuser \
SMB_PASSWORD=mypassword \
SMB_DOMAIN=CORP \
SMB_AUTH_PROTOCOL=negotiate \
./server
```

**Notes on running**
- The app will start without the SMB env vars, but upload attempts will return a 500 explaining which variables are missing.
- Using the example `127.0.0.1` test values will usually produce a connection error (expected in local dev if no SMB server is running).

**API**
- **GET** `/health` — health check endpoint
	- Returns `200` if application and SMB server are healthy and accessible
	- Returns `503` if application is unhealthy, SMB configuration is missing, or SMB server/share is inaccessible
	- JSON response includes `status`, `app_status`, `smb_connection`, `smb_share_accessible`, `server`, and `share` fields

- **POST** `/upload` (multipart/form-data)
	- `file`: the uploaded file
	- `remote_path`: path inside the share to write to (e.g., `inbox/report.pdf`). Leading `/` is stripped.
	- `overwrite`: optional boolean form field — when `true` overwrite existing file; default `false`.

**Responses**
- `200`: success — JSON like `{ "status": "ok", "remote_path": "inbox/report.pdf" }`
- `409`: conflict — remote file exists and `overwrite=false`
- `500`: server error — missing configuration or SMB connection/write failure

**Examples**
- Upload and fail if remote exists (default):

```bash
curl -v -X POST http://localhost:8080/upload \
	-F file=@/path/to/local-document.pdf \
	-F remote_path=inbox/report.pdf
```

- Upload and overwrite if exists:

```bash
curl -v -X POST http://localhost:8080/upload \
	-F file=@/path/to/local-document.pdf \
	-F remote_path=inbox/report.pdf \
	-F overwrite=true
```

**Docker**
- Build image:

```bash
docker build -t document-smbrelay:latest .
```

- Multi-architecture builds are automatically published to GitHub Container Registry for `linux/amd64` and `linux/arm64` platforms when changes are pushed to the main branch

- Run container (example):

```bash
docker run --rm -p 8080:8080 \
	-e SMB_SERVER_NAME=MYSMBSERVER \
	-e SMB_SERVER_IP=192.0.2.10 \
	-e SMB_SHARE_NAME=Documents \
	-e SMB_USERNAME=smbuser \
	-e SMB_PASSWORD='smb-password' \
	-e LOG_LEVEL=DEBUG \
	document-smbrelay:latest
```

Note: Docker builds in sandboxed/test environments may fail due to SSL/network restrictions; this is environment-specific.

**Docker Compose (Local Development & Testing)**

For local development and testing with a containerized SMB server:

```bash
# Start both SMB server and relay service
docker-compose up -d

# Upload a test file
echo "Hello World" > test.txt
curl -X POST http://localhost:8080/upload \
  -F file=@test.txt \
  -F remote_path=test.txt

# Check service health
curl http://localhost:8080/health | jq

# Stop services
docker-compose down
```

**Testing DFS Connectivity**

For testing with a simulated DFS environment (multiple file servers):

```bash
# Start DFS environment with namespace server and file servers
docker-compose -f docker-compose.dfs.yml up -d

# Test upload through DFS namespace
curl -X POST http://localhost:8080/upload \
  -F file=@test.txt \
  -F remote_path=test.txt

# Stop DFS environment
docker-compose -f docker-compose.dfs.yml down
```

See documentation:
- [DOCKER_TESTING.md](./DOCKER_TESTING.md) - Basic Docker testing (single SMB server)
- [DFS_TESTING.md](./DFS_TESTING.md) - DFS testing with multiple servers
- [DFS_KERBEROS.md](./DFS_KERBEROS.md) - Production DFS and Kerberos setup

**Validation & Manual Tests (recommended after changes)**
- **Basic syntax & import checks**:

```bash
python3 -m py_compile app/main.py
python3 -c "import app.main; print('Import successful')"
```

- **Smoke test of docs and OpenAPI** (start server, then):

```bash
curl -s http://localhost:8080/docs | grep -q "Document SMB Relay Service"
curl -s http://localhost:8080/openapi.json | python3 -m json.tool > /dev/null
```

- **Upload endpoint behavior checks**
	- Start server without SMB env vars and POST `/upload` — server should return a 500 with a clear message about missing variables.
	- Start server with test SMB env vars but no SMB server reachable — POST `/upload` should return a connection error (e.g., `Connection refused`).

See `app/main.py` for the exact error messages and handling logic.

**Testing**
- Run all tests:

```bash
./run_tests.sh
```

- Run only unit tests:

```bash
./run_tests.sh unit
```

- Run integration tests (requires Docker / SMB server):

```bash
./run_tests.sh integration
```

See `tests/README.md` for more test details.

**Project Structure (important files)**
- `app/main.py`: main FastAPI app and endpoint
- `requirements.txt`: runtime dependencies
- `requirements-test.txt`: test dependencies
- `Dockerfile`: container build
- `run_tests.sh`: helper script to run unit/integration tests
- `tests/`: unit and integration test suites

**Key Dependencies**
- `fastapi` — web framework
- `uvicorn[standard]` — ASGI server
- `pysmb` — SMB client used to write files to shares
- `python-multipart` — multipart/form handling
- `aiofiles` — async file operations

**Troubleshooting**
- `Missing SMB configuration environment variables`: ensure required `SMB_*` env vars are set before starting the server.
- `Connection refused` on upload: verify SMB server IP/port and credentials; test connectivity from the host/container.
