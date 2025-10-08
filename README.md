# Document SMB Relay Service

A minimal FastAPI service that accepts file uploads and writes them directly to an SMB share without mounting it on the host.

Requirements

- Python 3.10+
- `pip install -r requirements.txt`

Running locally

Start the server with uvicorn:

```bash
uvicorn app.main:app --host 0.0.0.0 --port 8080
```

Configuration (environment variables)

The service reads SMB connection settings from environment variables. The following are required:

- `SMB_SERVER_NAME` : NetBIOS name (or identifier) of the SMB server
- `SMB_SERVER_IP` : IP or hostname to connect to
- `SMB_SHARE_NAME` : SMB share name (e.g. `Documents`)
- `SMB_USERNAME` : SMB username
- `SMB_PASSWORD` : SMB password

Optional environment variables:

# Document SMB Relay Service

A minimal FastAPI service that accepts multipart file uploads and writes them directly to an SMB share using `pysmb` without mounting the share on the host.

**What this repository contains**
- **Core app**: `app/main.py` — the FastAPI application and upload endpoint.
- **Tests**: `tests/` — unit and integration tests, plus helpers.
- **Container**: `Dockerfile` to build a runnable image.

**Requirements**
- **Python**: 3.10+ (3.12+ recommended)
- **Install deps**: `pip install -r requirements.txt`

**Quick Start (local)**
- **Run server**: set SMB environment variables (see below) and start with:

```bash
SMB_SERVER_NAME=testserver \
SMB_SERVER_IP=127.0.0.1 \
SMB_SHARE_NAME=testshare \
SMB_USERNAME=testuser \
SMB_PASSWORD=testpass \
uvicorn app.main:app --host 0.0.0.0 --port 8080
```

**Required Environment Variables**
- `SMB_SERVER_NAME`: NetBIOS name of the SMB server
- `SMB_SERVER_IP`: IP address or hostname of the SMB server
- `SMB_SHARE_NAME`: Name of the SMB share (for example `Documents`)
- `SMB_USERNAME`: SMB username (optional when using Kerberos authentication)
- `SMB_PASSWORD`: SMB password (optional when using Kerberos authentication)

**Optional Environment Variables**
- `SMB_DOMAIN`: SMB domain/workgroup (default: empty)
- `SMB_PORT`: SMB port (default: `445`)
- `SMB_USE_NTLM_V2`: `true|false` (default: `true`, deprecated - use `SMB_AUTH_PROTOCOL` instead)
- `SMB_AUTH_PROTOCOL`: Authentication protocol - `negotiate|ntlm|kerberos` (default: derived from `SMB_USE_NTLM_V2`)
- `LOG_LEVEL`: Application log level - `DEBUG|INFO|WARNING|ERROR|CRITICAL` (default: `INFO`)

**Authentication Methods**

This service supports three authentication protocols:

1. **NTLM** (default): Username/password authentication using NTLM protocol
   - Set `SMB_AUTH_PROTOCOL=ntlm` or `SMB_USE_NTLM_V2=true`
   - Requires `SMB_USERNAME` and `SMB_PASSWORD`

2. **Negotiate**: Automatic protocol negotiation (NTLM or Kerberos)
   - Set `SMB_AUTH_PROTOCOL=negotiate` or `SMB_USE_NTLM_V2=false`
   - Requires `SMB_USERNAME` and `SMB_PASSWORD`

3. **Kerberos**: Kerberos authentication for Active Directory environments
   - Set `SMB_AUTH_PROTOCOL=kerberos`
   - `SMB_USERNAME` and `SMB_PASSWORD` are optional (can use system Kerberos ticket cache)
   - Ideal for Windows DFS shares with domain authentication

**Windows DFS Support**

This service **fully supports Windows Distributed File System (DFS)** shares. The underlying `smbprotocol` library automatically handles DFS referrals and path resolution.

To use with Windows DFS:
- Set `SMB_SERVER_NAME` to your DFS namespace server (e.g., `dfs.example.com`)
- Set `SMB_SHARE_NAME` to the DFS share name
- Use Kerberos authentication for best results with domain-joined environments
- The service will automatically follow DFS referrals to the actual file server

**Example: Windows DFS with Kerberos**

```bash
# Using system Kerberos ticket cache (e.g., after kinit)
SMB_SERVER_NAME=dfs.corp.example.com \
SMB_SERVER_IP=dfs.corp.example.com \
SMB_SHARE_NAME=documents \
SMB_AUTH_PROTOCOL=kerberos \
uvicorn app.main:app --host 0.0.0.0 --port 8080

# Or with explicit credentials
SMB_SERVER_NAME=dfs.corp.example.com \
SMB_SERVER_IP=dfs.corp.example.com \
SMB_SHARE_NAME=documents \
SMB_USERNAME=myuser \
SMB_PASSWORD=mypassword \
SMB_DOMAIN=CORP \
SMB_AUTH_PROTOCOL=kerberos \
uvicorn app.main:app --host 0.0.0.0 --port 8080
```

**Notes on running**
- The app will start without the SMB env vars, but upload attempts will return a 500 explaining which variables are missing.
- Using the example `127.0.0.1` test values will usually produce a connection error (expected in local dev if no SMB server is running).
- For Kerberos authentication, ensure proper Kerberos configuration (`/etc/krb5.conf`) and valid tickets if using ticket cache.

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
docker build -t document-smb-relay:latest .
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
	document-smb-relay:latest
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

See [DOCKER_TESTING.md](./DOCKER_TESTING.md) for comprehensive Docker testing documentation, including:
- Development setup with hot-reload
- Multiple testing scenarios
- Troubleshooting guide
- Performance testing examples

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
