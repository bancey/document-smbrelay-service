# Testing Documentation

This document describes the comprehensive test suite for the Document SMB Relay Service.

## Test Structure

The test suite is organized into two main categories:

### Unit Tests (`tests/unit/`)
Fast tests that use mocks and don't require external dependencies:

- **`test_file_utils.py`** - Tests for the `save_upload_to_temp()` function
  - File saving with various extensions
  - Large file handling in chunks
  - Empty file handling
  - Extension preservation

- **`test_smb_upload.py`** - Tests for the `smb_upload_file()` function  
  - SMB connection and upload logic
  - Directory creation behavior
  - File existence checking
  - Overwrite protection
  - Error handling scenarios

- **`test_endpoints.py`** - Tests for the FastAPI `/upload` endpoint
  - Environment variable validation
  - Path normalization
  - Error response handling
  - Temporary file cleanup

### Integration Tests (`tests/integration/`)
End-to-end tests using a real SMB server via Docker:
# Testing Documentation

This document describes how to run and reason about the project's test suite.

## Test layout

- `tests/unit/` — unit tests that mock external systems and run fast
- `tests/integration/` — end-to-end tests using a real SMB server (Docker)

Key test files:
- `tests/unit/test_file_utils.py` — temporary-file and upload-to-temp behavior
- `tests/unit/test_smb_upload.py` — SMB upload logic, directory creation, overwrite
- `tests/unit/test_endpoints.py` — FastAPI `/upload` endpoint behavior
- `tests/integration/test_smb_integration.py` — end-to-end SMB server tests

## Prerequisites

- Python 3.10+ (3.12+ recommended)
- Install runtime deps: `pip install -r requirements.txt`
- Install test deps: `pip install -r requirements-test.txt`

Integration tests additionally require:
- Docker (and Docker Compose if preferred)
- Ports available for SMB (typically `445`, and optionally `137-139` for NetBIOS)

If Docker is not available, integration tests can be skipped — unit tests cover core logic.

## Useful commands

- Run all tests:

```bash
./run_tests.sh
```

- Run only unit tests (fast):

```bash
./run_tests.sh unit
# or
python -m pytest -m unit -q
```

- Run only integration tests (requires Docker):

```bash
./run_tests.sh integration
# or
python -m pytest -m integration -q
```

- Run a single test file:

```bash
python -m pytest tests/unit/test_endpoints.py -q
```

- Run with verbose output:

```bash
python -m pytest -v
```

## Integration test details

The integration suite automatically starts a disposable SMB server inside Docker using test credentials and a temporary data volume. Tests verify:
- real SMB uploads and downloads
- directory creation on the share
- overwrite behavior and conflict handling

Integration server defaults used by the tests:
- Username: `testuser`
- Password: `testpass`
- Share name: `testshare`

The test harness cleans up any temporary containers and volumes after the tests complete.

## Test coverage and focus

The suite focuses on:
- File upload and temporary-file handling
- SMB connection/authentication and write logic
- Directory creation and path normalization
- Overwrite protection and conflict handling
- Error propagation from SMB layer to HTTP responses

Unit tests mock `pysmb` interactions so they run quickly and deterministically.

## Adding tests

Guidance when adding tests:
1. Add unit tests first and mock external dependencies.
2. Add integration tests only when the behavior requires a real SMB server.
3. Use markers `@pytest.mark.unit` and `@pytest.mark.integration`.
4. Keep tests isolated and ensure resources are cleaned up.

## Troubleshooting

- If integration tests fail with network/port errors, ensure Docker can bind the SMB ports on your host.
- If tests can't import the app, confirm `python3 -m py_compile app/main.py` and `python3 -c "import app.main"` succeed.
