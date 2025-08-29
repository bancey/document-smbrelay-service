import pytest
import os
import tempfile
from unittest.mock import AsyncMock


@pytest.fixture
def sample_file_content():
    """Sample file content for testing."""
    return b"This is a test file content for SMB upload testing."


@pytest.fixture
def temp_file(sample_file_content):
    """Create a temporary file for testing."""
    with tempfile.NamedTemporaryFile(delete=False, suffix=".txt") as tmp:
        tmp.write(sample_file_content)
        tmp.flush()
        yield tmp.name
    # Cleanup
    try:
        os.unlink(tmp.name)
    except OSError:
        pass


@pytest.fixture
def mock_upload_file(sample_file_content):
    """Mock UploadFile for testing."""
    mock_file = AsyncMock()
    mock_file.filename = "test_file.txt"
    mock_file.read = AsyncMock(side_effect=[sample_file_content, b""])
    mock_file.close = AsyncMock()
    return mock_file


@pytest.fixture
def smb_env_vars():
    """Environment variables for SMB configuration."""
    return {
        "SMB_SERVER_NAME": "testserver",
        "SMB_SERVER_IP": "127.0.0.1", 
        "SMB_SHARE_NAME": "testshare",
        "SMB_USERNAME": "testuser",
        "SMB_PASSWORD": "testpass",
        "SMB_DOMAIN": "",
        "SMB_PORT": "445",
        "SMB_USE_NTLM_V2": "true"
    }


@pytest.fixture
def clear_env():
    """Clear SMB environment variables for testing missing config."""
    smb_vars = [
        "SMB_SERVER_NAME", "SMB_SERVER_IP", "SMB_SHARE_NAME",
        "SMB_USERNAME", "SMB_PASSWORD", "SMB_DOMAIN", "SMB_PORT", "SMB_USE_NTLM_V2"
    ]
    original_values = {}
    for var in smb_vars:
        if var in os.environ:
            original_values[var] = os.environ[var]
            del os.environ[var]
    
    yield
    
    # Restore original values
    for var, value in original_values.items():
        os.environ[var] = value