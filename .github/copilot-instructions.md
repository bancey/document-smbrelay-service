# Document SMB Relay Service

**Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.**

A minimal FastAPI service that accepts file uploads via HTTP and writes them directly to SMB shares without mounting them on the host filesystem. The service uses pysmb for SMB connectivity and requires specific environment variables for configuration.

## Working Effectively

### Bootstrap and Build (2-3 minutes total)
- Install Python 3.10+ (Python 3.12+ recommended)
- `pip install -r requirements.txt` -- takes 1-2 minutes for fresh install. NEVER CANCEL. Set timeout to 5+ minutes.
- `python3 -m py_compile app/main.py` -- validate syntax (instant)
- `python3 -c "import app.main; print('Import successful')"` -- validate imports (instant)

### Running the Application
ALWAYS run the application with proper SMB environment variables for testing:
```bash
SMB_SERVER_NAME=testserver \
SMB_SERVER_IP=127.0.0.1 \
SMB_SHARE_NAME=testshare \
SMB_USERNAME=testuser \
SMB_PASSWORD=testpass \
uvicorn app.main:app --host 0.0.0.0 --port 8080
```

Server starts in ~3 seconds. The application will start successfully but SMB upload operations will fail with "Connection refused" when using test values - this is expected behavior.

### Docker Build and Run
- `docker build -t document-smb-relay:latest .` -- takes 3-5 minutes. NEVER CANCEL. Set timeout to 10+ minutes.
- Note: Docker build may fail in sandboxed environments due to SSL certificate issues when downloading Python packages. This is an environment limitation, not a code issue.

## Validation and Testing

### CRITICAL: Always run these validation steps after making changes

**1. Basic validation (30 seconds)**:
```bash
python3 -m py_compile app/main.py
python3 -c "import app.main; print('Import successful')"
```

**2. Full functional validation (30 seconds)**:
Start the server with test environment variables (see "Running the Application" above), then:
```bash
# Test docs endpoint
curl -s http://localhost:8080/docs | grep -q "Document SMB Relay Service"

# Test OpenAPI spec  
curl -s http://localhost:8080/openapi.json | python3 -m json.tool > /dev/null

# Test health endpoint with valid env vars but unreachable SMB server
curl -s http://localhost:8080/health | python3 -m json.tool
# Should return 503 with: {"status": "unhealthy", "smb_connection": "failed", ...}

# Test upload endpoint with missing env vars (start server without SMB_* variables)
curl -X POST http://localhost:8080/upload -F file=@README.md -F remote_path=test.txt
# Should return: {"detail":"Missing SMB configuration environment variables: ..."}

# Test upload endpoint with valid env vars but unreachable SMB server
curl -X POST http://localhost:8080/upload -F file=@README.md -F remote_path=test.txt  
# Should return: {"detail":"[Errno 111] Connection refused"}
```

**3. Complete end-to-end test scenario**:
ALWAYS test the complete workflow when making changes:
- Start server with SMB environment variables 
- Verify `/docs` endpoint returns Swagger UI with correct title
- Verify `/openapi.json` returns valid JSON specification
- Verify `/health` endpoint returns SMB connection status
- Test upload endpoint validates required environment variables (start server without SMB_* vars)
- Test upload endpoint attempts SMB connection and fails gracefully with connection errors
- Stop server and verify it shuts down cleanly

**4. Run automated test suite**:
```bash
# Run unit tests (fast, ~44 tests)
./run_tests.sh unit

# Run all tests if Docker is available
./run_tests.sh
```

### Automated Test Suite
This repository has a comprehensive test suite with unit and integration tests:

**Quick test commands**:
```bash
# Run all tests
./run_tests.sh

# Run only unit tests (fast, no external dependencies)
./run_tests.sh unit

# Run only integration tests (requires Docker)
./run_tests.sh integration
```

**Manual validation** is still recommended for end-to-end verification in addition to the automated tests.

## Required Environment Variables

The service requires these environment variables to operate:
- `SMB_SERVER_NAME`: NetBIOS name of the SMB server
- `SMB_SERVER_IP`: IP address or hostname of the SMB server  
- `SMB_SHARE_NAME`: Name of the SMB share (e.g., "Documents")
- `SMB_USERNAME`: SMB username for authentication
- `SMB_PASSWORD`: SMB password for authentication

Optional environment variables:
- `SMB_DOMAIN`: SMB domain/workgroup (default: empty)
- `SMB_PORT`: SMB port (default: 445)
- `SMB_USE_NTLM_V2`: Enable NTLMv2 authentication (default: true)

## API Endpoints and Usage

### GET /health
Health check endpoint that verifies application responsiveness and SMB connectivity.

Returns JSON with status information:
- `200`: Application and SMB server are healthy and accessible
- `503`: Application is unhealthy, SMB configuration is missing, or SMB server/share is inaccessible

Example response:
```json
{
  "status": "healthy",
  "app_status": "ok",
  "smb_connection": "ok", 
  "smb_share_accessible": true,
  "server": "testserver (127.0.0.1:445)",
  "share": "testshare"
}
```

### POST /upload
Accepts multipart/form-data with:
- `file`: The file to upload
- `remote_path`: Path within the SMB share (e.g., "inbox/report.pdf")
- `overwrite`: Optional boolean, defaults to false

### Testing Examples
```bash
# Upload with overwrite protection (default)
curl -X POST http://localhost:8080/upload \
  -F file=@document.pdf \
  -F remote_path=inbox/report.pdf

# Upload with overwrite enabled
curl -X POST http://localhost:8080/upload \
  -F file=@document.pdf \
  -F remote_path=inbox/report.pdf \
  -F overwrite=true
```

## Common Development Tasks

### Making Code Changes
1. Always validate syntax first: `python3 -m py_compile app/main.py`
2. Test imports: `python3 -c "import app.main"`
3. Start server with test environment variables
4. Run full validation tests (see Validation section above)
5. Check that the service properly handles both missing environment variables and SMB connection failures

### Debugging Connection Issues
- The service will return HTTP 500 with "Missing SMB configuration environment variables" if required env vars are not set
- The service will return HTTP 500 with connection error details if it cannot connect to the SMB server
- Check SMB server accessibility and credentials when debugging real deployment issues

### Code Quality
- The codebase uses standard Python formatting
- No specific linting tools are configured
- Keep the code simple and focused - this is intentionally a minimal service
- Always ensure proper error handling for SMB connection failures

## Important File Locations

### Core Application
- `app/main.py` - Main FastAPI application (~115 lines)
- `app/config/smb_config.py` - SMB configuration management
- `app/smb/connection.py` - SMB connection handling and health checks
- `app/smb/operations.py` - SMB file operations
- `app/utils/file_utils.py` - File utility functions
- `requirements.txt` - Python dependencies (5 packages)
- `requirements-test.txt` - Test dependencies
- `Dockerfile` - Container build configuration

### Testing Infrastructure
- `run_tests.sh` - Test runner script (unit/integration tests)
- `tests/unit/` - Unit tests with mocked dependencies
- `tests/integration/` - End-to-end tests with real SMB server
- `pytest.ini` - Pytest configuration

### Configuration and Documentation  
- `README.md` - Usage documentation and API examples
- `tests/README.md` - Test suite documentation
- `.github/workflows/docker-publish.yml` - Docker image CI/CD pipeline
- `renovate.json` - Dependency update automation

### Key Dependencies
- `fastapi==0.116.1` - Web framework
- `uvicorn[standard]==0.35.0` - ASGI server  
- `pysmb==1.2.11` - SMB client library
- `python-multipart==0.0.20` - Multipart form handling
- `aiofiles==24.1.0` - Async file operations

## Troubleshooting

### "SSL certificate verify failed" during Docker build
- This occurs in sandboxed environments and is not a code issue
- The Docker build works correctly in normal environments with internet access

### "Connection refused" during upload testing
- This is expected when using test SMB server values (127.0.0.1)
- Indicates the service is working correctly but cannot reach the specified SMB server

### "Missing SMB configuration environment variables"
- Ensure all required SMB_* environment variables are set before starting the server
- See "Required Environment Variables" section above for the complete list

## CI/CD Notes

- GitHub Actions workflow builds and publishes Docker images
- Comprehensive automated test suite with unit and integration tests
- Docker images are published to GitHub Container Registry (ghcr.io)  
- The workflow triggers on pushes to main branch and pull requests

## Common Repository Information

Save time by referencing these frequently-used outputs instead of running bash commands:

### Repository Structure
```
.
├── .github/
│   ├── copilot-instructions.md    # Development guidelines
│   └── workflows/
│       └── docker-publish.yml
├── app/
│   ├── __init__.py
│   ├── main.py                    # Main FastAPI application (~115 lines)
│   ├── config/
│   │   ├── __init__.py
│   │   └── smb_config.py         # SMB configuration management
│   ├── smb/
│   │   ├── __init__.py
│   │   ├── connection.py         # SMB connection and health checks
│   │   └── operations.py         # SMB file operations
│   └── utils/
│       ├── __init__.py
│       └── file_utils.py         # File utility functions
├── tests/
│   ├── conftest.py               # Pytest configuration and fixtures
│   ├── README.md                 # Test documentation
│   ├── integration/              # End-to-end tests
│   └── unit/                     # Unit tests with mocks
├── Dockerfile                    # Container build configuration
├── README.md                     # Usage documentation
├── pytest.ini                   # Pytest configuration
├── requirements.txt              # Runtime dependencies (5 packages)
├── requirements-test.txt         # Test dependencies
├── renovate.json                # Dependency automation
├── run_tests.sh                 # Test runner script
└── .gitignore                   # Git ignore rules
```

### Core Dependencies from requirements.txt
```
fastapi==0.116.1
uvicorn[standard]==0.35.0  
pysmb==1.2.11
python-multipart==0.0.20
aiofiles==24.1.0
```

### Key imports from app/main.py
```python
from fastapi import FastAPI, UploadFile, File, Form, HTTPException
from fastapi.responses import JSONResponse
import asyncio
import os

from .config.smb_config import load_smb_config_from_env
from .smb.connection import check_smb_health
from .smb.operations import smb_upload_file
from .utils.file_utils import save_upload_to_temp
```