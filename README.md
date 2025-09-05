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
- `SMB_USERNAME`: SMB username
- `SMB_PASSWORD`: SMB password

**Optional Environment Variables**
- `SMB_DOMAIN`: SMB domain/workgroup (default: empty)
- `SMB_PORT`: SMB port (default: `445`)
- `SMB_USE_NTLM_V2`: `true|false` (default: `true`)

**Notes on running**
- The app will start without the SMB env vars, but upload attempts will return a 500 explaining which variables are missing.
- Using the example `127.0.0.1` test values will usually produce a connection error (expected in local dev if no SMB server is running).

**API**
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

- Run container (example):

```bash
docker run --rm -p 8080:8080 \
	-e SMB_SERVER_NAME=MYSMBSERVER \
	-e SMB_SERVER_IP=192.0.2.10 \
	-e SMB_SHARE_NAME=Documents \
	-e SMB_USERNAME=smbuser \
	-e SMB_PASSWORD='smb-password' \
	document-smb-relay:latest
```

Note: Docker builds in sandboxed/test environments may fail due to SSL/network restrictions; this is environment-specific.

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

**Troubleshooting**
- `Missing SMB configuration environment variables`: ensure required `SMB_*` env vars are set before starting the server.
- `Connection refused` on upload: verify SMB server IP/port and credentials; test connectivity from the host/container.
