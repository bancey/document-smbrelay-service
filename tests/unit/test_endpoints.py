import pytest
import os
from unittest.mock import AsyncMock, Mock, patch
from fastapi.testclient import TestClient
from httpx import AsyncClient
from app.main import app
import tempfile
import io


@pytest.mark.unit
class TestUploadEndpoint:
    """Test cases for the /upload endpoint."""

    @pytest.fixture
    def client(self):
        """Create test client."""
        return TestClient(app)

    @pytest.fixture
    def test_file_data(self):
        """Test file data."""
        return io.BytesIO(b"Test file content")

    def test_upload_missing_env_vars(self, client, clear_env):
        """Test upload with missing environment variables."""
        test_file = io.BytesIO(b"test content")
        
        response = client.post(
            "/upload",
            files={"file": ("test.txt", test_file, "text/plain")},
            data={"remote_path": "test.txt"}
        )
        
        assert response.status_code == 500
        assert "Missing SMB configuration environment variables" in response.json()["detail"]
        assert "SMB_SERVER_NAME" in response.json()["detail"]

    def test_upload_partial_env_vars(self, client, clear_env):
        """Test upload with only some environment variables set."""
        # Set only some environment variables
        os.environ["SMB_SERVER_NAME"] = "testserver"
        os.environ["SMB_SERVER_IP"] = "127.0.0.1"
        # Missing SMB_SHARE_NAME, SMB_USERNAME, SMB_PASSWORD
        
        test_file = io.BytesIO(b"test content")
        
        response = client.post(
            "/upload",
            files={"file": ("test.txt", test_file, "text/plain")},
            data={"remote_path": "test.txt"}
        )
        
        assert response.status_code == 500
        detail = response.json()["detail"]
        assert "Missing SMB configuration environment variables" in detail
        assert "SMB_SHARE_NAME" in detail
        assert "SMB_USERNAME" in detail
        assert "SMB_PASSWORD" in detail

    def test_upload_path_normalization(self, client, smb_env_vars):
        """Test that leading slashes are stripped from remote_path."""
        for var, value in smb_env_vars.items():
            os.environ[var] = value
        
        test_file = io.BytesIO(b"test content")
        
        with patch('app.main.smb_upload_file') as mock_upload:
            response = client.post(
                "/upload",
                files={"file": ("test.txt", test_file, "text/plain")},
                data={"remote_path": "/leading/slash/file.txt"}
            )
        
        # The upload function should be called with normalized path
        mock_upload.assert_called_once()
        args = mock_upload.call_args[0]  # Positional arguments
        # Arguments: tmp_path, server_name, server_ip, share_name, remote_path, username, password, domain, port, use_ntlm_v2, overwrite
        assert args[4] == "leading/slash/file.txt"  # remote_path is 5th argument (index 4)

    def test_upload_success(self, client, smb_env_vars):
        """Test successful file upload."""
        for var, value in smb_env_vars.items():
            os.environ[var] = value
        
        test_file = io.BytesIO(b"test content")
        
        with patch('app.main.smb_upload_file') as mock_upload:
            response = client.post(
                "/upload",
                files={"file": ("test.txt", test_file, "text/plain")},
                data={"remote_path": "folder/test.txt"}
            )
        
        assert response.status_code == 200
        assert response.json() == {"status": "ok", "remote_path": "folder/test.txt"}
        
        # Verify upload function was called with correct parameters
        mock_upload.assert_called_once()
        args = mock_upload.call_args[0]  # Positional arguments
        # Arguments: tmp_path, server_name, server_ip, share_name, remote_path, username, password, domain, port, use_ntlm_v2, overwrite
        assert args[1] == "testserver"        # server_name
        assert args[2] == "127.0.0.1"        # server_ip
        assert args[3] == "testshare"         # share_name
        assert args[4] == "folder/test.txt"   # remote_path
        assert args[5] == "testuser"          # username
        assert args[6] == "testpass"          # password
        assert args[10] is False              # overwrite

    def test_upload_with_overwrite(self, client, smb_env_vars):
        """Test file upload with overwrite enabled."""
        for var, value in smb_env_vars.items():
            os.environ[var] = value
        
        test_file = io.BytesIO(b"test content")
        
        with patch('app.main.smb_upload_file') as mock_upload:
            response = client.post(
                "/upload",
                files={"file": ("test.txt", test_file, "text/plain")},
                data={"remote_path": "test.txt", "overwrite": "true"}
            )
        
        assert response.status_code == 200
        
        # Verify overwrite parameter was passed correctly
        args = mock_upload.call_args[0]  # Positional arguments
        assert args[10] is True  # overwrite is 11th argument (index 10)

    def test_upload_file_exists_error(self, client, smb_env_vars):
        """Test handling of FileExistsError."""
        for var, value in smb_env_vars.items():
            os.environ[var] = value
        
        test_file = io.BytesIO(b"test content")
        
        with patch('app.main.smb_upload_file') as mock_upload:
            mock_upload.side_effect = FileExistsError("Remote file already exists: test.txt")
            
            response = client.post(
                "/upload",
                files={"file": ("test.txt", test_file, "text/plain")},
                data={"remote_path": "test.txt"}
            )
        
        assert response.status_code == 409
        assert "Remote file already exists: test.txt" in response.json()["detail"]

    def test_upload_connection_error(self, client, smb_env_vars):
        """Test handling of connection errors."""
        for var, value in smb_env_vars.items():
            os.environ[var] = value
        
        test_file = io.BytesIO(b"test content")
        
        with patch('app.main.smb_upload_file') as mock_upload:
            mock_upload.side_effect = ConnectionError("Could not connect to SMB server")
            
            response = client.post(
                "/upload",
                files={"file": ("test.txt", test_file, "text/plain")},
                data={"remote_path": "test.txt"}
            )
        
        assert response.status_code == 500
        assert "Could not connect to SMB server" in response.json()["detail"]

    def test_upload_generic_error(self, client, smb_env_vars):
        """Test handling of generic errors."""
        for var, value in smb_env_vars.items():
            os.environ[var] = value
        
        test_file = io.BytesIO(b"test content")
        
        with patch('app.main.smb_upload_file') as mock_upload:
            mock_upload.side_effect = Exception("Unexpected error")
            
            response = client.post(
                "/upload",
                files={"file": ("test.txt", test_file, "text/plain")},
                data={"remote_path": "test.txt"}
            )
        
        assert response.status_code == 500
        assert "Unexpected error" in response.json()["detail"]

    def test_upload_environment_parsing(self, client, clear_env):
        """Test parsing of environment variables."""
        # Set environment variables with different formats
        os.environ["SMB_SERVER_NAME"] = "testserver"
        os.environ["SMB_SERVER_IP"] = "127.0.0.1"
        os.environ["SMB_SHARE_NAME"] = "testshare"
        os.environ["SMB_USERNAME"] = "testuser"
        os.environ["SMB_PASSWORD"] = "testpass"
        os.environ["SMB_DOMAIN"] = "TESTDOMAIN"
        os.environ["SMB_PORT"] = "139"
        os.environ["SMB_USE_NTLM_V2"] = "false"
        
        test_file = io.BytesIO(b"test content")
        
        with patch('app.main.smb_upload_file') as mock_upload:
            response = client.post(
                "/upload",
                files={"file": ("test.txt", test_file, "text/plain")},
                data={"remote_path": "test.txt"}
            )
        
        # Verify parsed values are passed correctly
        args = mock_upload.call_args[0]  # Positional arguments
        # Arguments: tmp_path, server_name, server_ip, share_name, remote_path, username, password, domain, port, use_ntlm_v2, overwrite
        assert args[7] == "TESTDOMAIN"  # domain
        assert args[8] == 139           # port
        assert args[9] is False         # use_ntlm_v2

    def test_upload_ntlm_v2_parsing_variations(self, client, smb_env_vars):
        """Test different ways of specifying SMB_USE_NTLM_V2."""
        test_values = [
            ("1", True),
            ("true", True),
            ("TRUE", True),
            ("yes", True),
            ("YES", True),
            ("0", False),
            ("false", False),
            ("FALSE", False),
            ("no", False),
            ("anything_else", False)
        ]
        
        for env_value, expected in test_values:
            for var, value in smb_env_vars.items():
                os.environ[var] = value
            os.environ["SMB_USE_NTLM_V2"] = env_value
            
            test_file = io.BytesIO(b"test content")
            
            with patch('app.main.smb_upload_file') as mock_upload:
                response = client.post(
                    "/upload",
                    files={"file": ("test.txt", test_file, "text/plain")},
                    data={"remote_path": "test.txt"}
                )
            
            args = mock_upload.call_args[0]  # Positional arguments
            assert args[9] is expected, f"Failed for env_value: {env_value}"  # use_ntlm_v2 is 10th argument (index 9)

    def test_upload_missing_file(self, client, smb_env_vars):
        """Test upload request without file."""
        for var, value in smb_env_vars.items():
            os.environ[var] = value
        
        response = client.post(
            "/upload",
            data={"remote_path": "test.txt"}
        )
        
        assert response.status_code == 422  # Validation error

    def test_upload_missing_remote_path(self, client, smb_env_vars):
        """Test upload request without remote_path."""
        for var, value in smb_env_vars.items():
            os.environ[var] = value
        
        test_file = io.BytesIO(b"test content")
        
        response = client.post(
            "/upload",
            files={"file": ("test.txt", test_file, "text/plain")}
        )
        
        assert response.status_code == 422  # Validation error

    @patch('app.main.os.remove')
    def test_temp_file_cleanup_success(self, mock_remove, client, smb_env_vars):
        """Test that temporary files are cleaned up after successful upload."""
        for var, value in smb_env_vars.items():
            os.environ[var] = value
        
        test_file = io.BytesIO(b"test content")
        
        with patch('app.main.smb_upload_file'):
            response = client.post(
                "/upload",
                files={"file": ("test.txt", test_file, "text/plain")},
                data={"remote_path": "test.txt"}
            )
        
        assert response.status_code == 200
        # Temporary file should be removed
        mock_remove.assert_called_once()

    @patch('app.main.os.remove')
    def test_temp_file_cleanup_on_error(self, mock_remove, client, smb_env_vars):
        """Test that temporary files are cleaned up even when upload fails."""
        for var, value in smb_env_vars.items():
            os.environ[var] = value
        
        test_file = io.BytesIO(b"test content")
        
        with patch('app.main.smb_upload_file') as mock_upload:
            mock_upload.side_effect = Exception("Upload failed")
            
            response = client.post(
                "/upload",
                files={"file": ("test.txt", test_file, "text/plain")},
                data={"remote_path": "test.txt"}
            )
        
        assert response.status_code == 500
        # Temporary file should still be removed
        mock_remove.assert_called_once()

    def test_temp_file_cleanup_remove_raises_is_swallowed(self, client, smb_env_vars):
        """If os.remove raises during cleanup it should be swallowed and not affect response."""
        for var, value in smb_env_vars.items():
            os.environ[var] = value

        test_file = io.BytesIO(b"test content")

        # Patch smb_upload_file to succeed, and os.remove to raise
        with patch('app.main.smb_upload_file'):
            with patch('app.main.os.remove', side_effect=Exception('remove failed')) as mock_remove:
                response = client.post(
                    "/upload",
                    files={"file": ("test.txt", test_file, "text/plain")},
                    data={"remote_path": "test.txt"}
                )

        # Endpoint should still return success even if remove failed
        assert response.status_code == 200
        mock_remove.assert_called_once()