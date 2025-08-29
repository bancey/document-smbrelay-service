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

- `SMB_DOMAIN` : SMB domain/workgroup (default: empty)
- `SMB_PORT` : SMB port (default: `445`)
- `SMB_USE_NTLM_V2` : `true|false` (default: `true`)

Docker

Build image:

```bash
docker build -t document-smb-relay:latest .
```

Run container (example):

```bash
docker run --rm -p 8080:8080 \
	-e SMB_SERVER_NAME=MYSMBSERVER \
	-e SMB_SERVER_IP=192.0.2.10 \
	-e SMB_SHARE_NAME=Documents \
	-e SMB_USERNAME=smbuser \
	-e SMB_PASSWORD='smb-password' \
	document-smb-relay:latest
```

API

POST /upload (multipart/form-data)

Fields:

- `file` : the file to upload
- `remote_path` : path within the share to write to (e.g., `inbox/report.pdf`). Leading `/` is stripped.
- `overwrite` : optional boolean (form) — when `true` the server will overwrite an existing file; default `false`.

Responses:

- 200: success — JSON `{ "status": "ok", "remote_path": "inbox/report.pdf" }`
- 409: conflict — the remote file already exists and `overwrite=false`
- 500: server error — connection or write failure

Examples

Upload and fail if remote exists (default):

```bash
curl -v -X POST http://localhost:8080/upload \
	-F file=@/path/to/local-document.pdf \
	-F remote_path=inbox/report.pdf
```

Upload and overwrite if exists:

```bash
curl -v -X POST http://localhost:8080/upload \
	-F file=@/path/to/local-document.pdf \
	-F remote_path=inbox/report.pdf \
	-F overwrite=true
```

Docker example (using env vars above):

```bash
curl -v -X POST http://localhost:8080/upload \
	-F file=@/path/to/local-document.pdf \
	-F remote_path=inbox/report.pdf
```

Notes and recommendations

- The service connects to the SMB share at runtime using `pysmb` (`SMBConnection`). It attempts to create missing directories on the share.
- For production, run behind HTTPS and protect the endpoint (authentication, API keys, mTLS). Avoid sending SMB credentials over plain HTTP.
- Use a secrets manager or Docker secrets to inject `SMB_USERNAME`/`SMB_PASSWORD` securely.
- If you require atomic replaces to avoid partial writes during overwrite, consider enabling the "safe rename" behavior (write to a temp filename on the share then atomically rename). I can implement that if desired.

If you'd like, I can add a `docker-compose.yml`, a health/readiness endpoint, or API key protection next.
