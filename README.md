# Document SMB Relay Service

A minimal FastAPI service that accepts file uploads and writes them directly to an SMB share without mounting it on the host.

Requirements

- Python 3.10+
- `pip install -r requirements.txt`

Running

Start the server with uvicorn:

```bash
uvicorn app.main:app --host 0.0.0.0 --port 8080
```

API

POST /upload (multipart/form-data)

Fields:
- `file`: the file to upload
- `server_name`: NetBIOS name of the SMB server (arbitrary identifier)
- `server_ip`: IP or hostname to connect to
- `share_name`: name of the SMB share (e.g., `Documents`)
- `remote_path`: path within the share to write to (e.g., `inbox/report.pdf`)
- `username`: SMB username
- `password`: SMB password
- `domain`: Optional SMB domain/workgroup

Response:
- 200: {"status":"ok","remote_path":"inbox/report.pdf"}
- 500: {"detail":"error message"}

Notes

- The service writes files by connecting at runtime using `pysmb`. It attempts to create missing directories.
- For production, run behind a process manager and secure transport (TLS) and credentials storage.
