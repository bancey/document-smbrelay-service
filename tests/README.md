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

- **`test_smb_integration.py`** - Complete workflow tests
  - Real SMB server upload/download
  - Directory creation verification  
  - Overwrite behavior testing
  - Large file transfers
  - FastAPI endpoint integration

## Running Tests

### Prerequisites

**For Unit Tests:**
- Python 3.10+
- Test dependencies: `pip install -r requirements-test.txt`

**For Integration Tests:**
- Docker and Docker Compose
- Available ports 445, 137-139 for SMB server

### Test Commands

**Run all tests:**
```bash
./run_tests.sh
```

**Run only unit tests (fast):**
```bash
./run_tests.sh unit
# or
python -m pytest -m unit
```

**Run only integration tests:**
```bash
./run_tests.sh integration  
# or
python -m pytest -m integration
```

**Run with verbose output:**
```bash
python -m pytest -v
```

**Run specific test file:**
```bash
python -m pytest tests/unit/test_endpoints.py -v
```

### Test Coverage

The test suite provides comprehensive coverage of:

- ✅ File upload functionality
- ✅ SMB connection and authentication  
- ✅ Directory creation logic
- ✅ Overwrite protection
- ✅ Error handling and edge cases
- ✅ Environment variable parsing
- ✅ Temporary file management
- ✅ Path normalization
- ✅ Large file handling
- ✅ End-to-end API workflow

## Integration Test Details

### SMB Server Setup

Integration tests automatically start a Docker-based SMB server with:
- **Username:** `testuser`
- **Password:** `testpass`  
- **Share name:** `testshare`
- **Ports:** 445, 137-139

The server uses a temporary Docker volume that's cleaned up after tests.

### Test Data

- Test files range from empty to 1MB in size
- Various file extensions are tested (.txt, .pdf, .bin, etc.)
- Nested directory structures up to 3 levels deep
- Unicode and special characters in filenames

### Error Scenarios

Integration tests verify proper handling of:
- Connection failures
- Authentication errors
- Permission denied scenarios
- Disk space limitations
- Network timeouts

## Test Configuration

Configuration is managed through:
- **`pytest.ini`** - Pytest settings and markers
- **`tests/conftest.py`** - Shared fixtures and test utilities
- **`docker-compose.test.yml`** - SMB server for integration tests

## Continuous Integration

The test suite is designed to work in CI environments:
- Unit tests run without external dependencies
- Integration tests can be skipped if Docker is unavailable  
- All tests use temporary files and clean up after themselves
- Test results are formatted for CI parsing

## Adding New Tests

When adding new functionality:

1. **Add unit tests first** - Fast feedback during development
2. **Mock external dependencies** - Keep unit tests isolated
3. **Add integration tests** - Verify real-world scenarios
4. **Use appropriate markers** - `@pytest.mark.unit` or `@pytest.mark.integration`
5. **Clean up resources** - Remove test files and connections

### Example Test Structure

```python
@pytest.mark.unit
class TestNewFeature:
    def test_basic_functionality(self):
        # Test with mocks
        pass
    
    def test_error_handling(self):
        # Test error scenarios
        pass

@pytest.mark.integration  
class TestNewFeatureIntegration:
    def test_end_to_end(self, smb_server):
        # Test with real SMB server
        pass
```