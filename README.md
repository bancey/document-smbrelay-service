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

### Optional Environment Variables

- `SMB_DOMAIN`: SMB domain/workgroup (default: empty)
- `SMB_PORT`: SMB port (default: `445`)
- `SMB_USE_NTLM_V2`: Enable NTLMv2 (default: `true`, deprecated - use `SMB_AUTH_PROTOCOL`)
- `SMB_AUTH_PROTOCOL`: Authentication protocol - `negotiate|ntlm|kerberos` (default: derived from `SMB_USE_NTLM_V2`)
- `LOG_LEVEL`: Application log level - `DEBUG|INFO|WARNING|ERROR` (default: `INFO`)
- `PORT`: HTTP server port (default: `8080`)

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

## API Endpoints

### GET /health

Health check endpoint that verifies application and SMB connectivity.

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

## Docker

### Build Image

```bash
docker build -f Dockerfile.go -t document-smbrelay:latest .
```

Or use Make:
```bash
make docker-build
```

### Run Container

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

### Multi-Architecture Builds

```bash
docker buildx build --platform linux/amd64,linux/arm64 \
  -f Dockerfile.go \
  -t document-smbrelay:latest .
```

### Docker Compose

See [DOCKER_TESTING.md](DOCKER_TESTING.md) for docker-compose examples with test SMB servers.

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
- **[DFS_TESTING.md](DFS_TESTING.md)** - DFS testing guide
- **[DOCKER_TESTING.md](DOCKER_TESTING.md)** - Docker testing guide

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
